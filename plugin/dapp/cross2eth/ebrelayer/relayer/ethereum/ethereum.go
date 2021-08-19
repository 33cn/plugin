package ethereum

// -----------------------------------------------------
//      Ethereum relayer
//
//      Initializes the relayer service, which parses,
//      encodes, and packages named events on an Ethereum
//      Smart Contract for validator's to sign and send
//      to the Chain33 bridge.
// -----------------------------------------------------

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"

	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

//Relayer4Ethereum ...
type Relayer4Ethereum struct {
	provider           string
	clientChainID      *big.Int
	bridgeRegistryAddr common.Address
	db                 dbm.DB
	rwLock             sync.RWMutex

	privateKey4Ethereum *ecdsa.PrivateKey
	ethSender           common.Address

	ethValidator           common.Address
	unlockchan             chan int
	maturityDegree         int32
	fetchHeightPeriodMs    int32
	eventLogIndex          ebTypes.EventLogIndex
	clientSpec             ethinterface.EthClientSpec
	bridgeBankAddr         common.Address
	bridgeBankSub          ethereum.Subscription
	bridgeBankLog          chan types.Log
	bridgeBankEventLockSig string
	bridgeBankEventBurnSig string
	bridgeBankAbi          abi.ABI
	deployInfo             *ebTypes.Deploy
	x2EthDeployInfo        *ethtxs.X2EthDeployInfo
	deployPara             *ethtxs.DeployPara
	operatorInfo           *ethtxs.OperatorInfo
	x2EthContracts         *ethtxs.X2EthContracts
	ethBridgeClaimChan     chan<- *ebTypes.EthBridgeClaim
	chain33MsgChan         <-chan *events.Chain33Msg
	totalTx4Eth2Chain33    int64
	symbol2Addr            map[string]common.Address
	symbol2LockAddr        map[string]common.Address
	mulSignAddr            string
}

var (
	relayerLog = log.New("module", "ethereum_relayer")
)

const (
	DefaultBlockPeriod = 5000
)

type EthereumStartPara struct {
	DbHandle           dbm.DB
	EthProvider        string
	BridgeRegistryAddr string
	DeployInfo         *ebTypes.Deploy
	Degree             int32
	BlockInterval      int32
	EthBridgeClaimChan chan<- *ebTypes.EthBridgeClaim
	Chain33MsgChan     <-chan *events.Chain33Msg
}

//StartEthereumRelayer ///
func StartEthereumRelayer(startPara *EthereumStartPara) *Relayer4Ethereum {
	if 0 == startPara.BlockInterval {
		startPara.BlockInterval = DefaultBlockPeriod
	}
	ethRelayer := &Relayer4Ethereum{
		provider:            startPara.EthProvider,
		db:                  startPara.DbHandle,
		unlockchan:          make(chan int, 2),
		bridgeRegistryAddr:  common.HexToAddress(startPara.BridgeRegistryAddr),
		deployInfo:          startPara.DeployInfo,
		maturityDegree:      startPara.Degree,
		fetchHeightPeriodMs: startPara.BlockInterval,
		ethBridgeClaimChan:  startPara.EthBridgeClaimChan,
		chain33MsgChan:      startPara.Chain33MsgChan,
		totalTx4Eth2Chain33: 0,
		symbol2Addr:         make(map[string]common.Address),
		symbol2LockAddr:     make(map[string]common.Address),
	}

	registrAddrInDB, err := ethRelayer.getBridgeRegistryAddr()
	//如果输入的registry地址非空，且和数据库保存地址不一致，则直接使用输入注册地址
	if startPara.BridgeRegistryAddr != "" && nil == err && registrAddrInDB != startPara.BridgeRegistryAddr {
		relayerLog.Error("StartEthereumRelayer", "BridgeRegistry is setted already with value", registrAddrInDB, "but now setting to", startPara.BridgeRegistryAddr)
		_ = ethRelayer.setBridgeRegistryAddr(startPara.BridgeRegistryAddr)
	} else if startPara.BridgeRegistryAddr == "" && registrAddrInDB != "" {
		//输入地址为空，且数据库中保存地址不为空，则直接使用数据库中的地址
		ethRelayer.bridgeRegistryAddr = common.HexToAddress(registrAddrInDB)
	}
	ethRelayer.eventLogIndex = ethRelayer.getLastBridgeBankProcessedHeight()
	ethRelayer.initBridgeBankTx()
	ethRelayer.mulSignAddr = ethRelayer.getMultiSignAddress()

	// Start clientSpec with infura ropsten provider
	relayerLog.Info("Relayer4Ethereum proc", "Started Ethereum websocket with provider:", ethRelayer.provider)
	client, err := ethtxs.SetupWebsocketEthClient(ethRelayer.provider)
	if err != nil {
		panic(err)
	}
	ethRelayer.clientSpec = client

	ctx := context.Background()
	clientChainID, err := client.NetworkID(ctx)
	if err != nil {
		errinfo := fmt.Sprintf("Failed to get NetworkID due to:%s", err.Error())
		panic(errinfo)
	}
	ethRelayer.clientChainID = clientChainID

	go ethRelayer.proc()
	return ethRelayer
}

func (ethRelayer *Relayer4Ethereum) recoverDeployPara() (err error) {
	if nil == ethRelayer.deployInfo {
		return nil
	}
	deployPrivateKey, err := crypto.ToECDSA(common.FromHex(ethRelayer.deployInfo.DeployerPrivateKey))
	if nil != err {
		return err
	}
	deployerAddr := crypto.PubkeyToAddress(deployPrivateKey.PublicKey)
	ethRelayer.rwLock.Lock()
	ethRelayer.operatorInfo = &ethtxs.OperatorInfo{
		PrivateKey: deployPrivateKey,
		Address:    deployerAddr,
	}
	ethRelayer.rwLock.Unlock()

	return nil
}

//DeployContrcts 部署以太坊合约
func (ethRelayer *Relayer4Ethereum) DeployContrcts() (bridgeRegistry string, err error) {
	bridgeRegistry = ""
	if nil == ethRelayer.deployInfo {
		return bridgeRegistry, errors.New("no deploy info configured yet")
	}
	deployPrivateKey, err := crypto.ToECDSA(common.FromHex(ethRelayer.deployInfo.DeployerPrivateKey))
	if nil != err {
		return bridgeRegistry, err
	}
	if len(ethRelayer.deployInfo.ValidatorsAddr) != len(ethRelayer.deployInfo.InitPowers) {
		return bridgeRegistry, errors.New("not same number for validator address and power")
	}
	if len(ethRelayer.deployInfo.ValidatorsAddr) < 3 {
		return bridgeRegistry, errors.New("the number of validator must be not less than 3")
	}

	nilAddr := common.Address{}

	//已经设置了注册合约地址，说明已经部署了相关的合约，不再重复部署
	if ethRelayer.bridgeRegistryAddr != nilAddr {
		return bridgeRegistry, errors.New("contract deployed already")
	}

	var validators []common.Address
	var initPowers []*big.Int

	for i, addr := range ethRelayer.deployInfo.ValidatorsAddr {
		validators = append(validators, common.HexToAddress(addr))
		initPowers = append(initPowers, big.NewInt(ethRelayer.deployInfo.InitPowers[i]))
	}
	deployerAddr := crypto.PubkeyToAddress(deployPrivateKey.PublicKey)
	para := &ethtxs.DeployPara{
		DeployPrivateKey: deployPrivateKey,
		Deployer:         deployerAddr,
		Operator:         deployerAddr,
		InitValidators:   validators,
		ValidatorPriKey:  []*ecdsa.PrivateKey{deployPrivateKey},
		InitPowers:       initPowers,
	}

	for i, power := range para.InitPowers {
		relayerLog.Info("deploy", "the validator address ", para.InitValidators[i].String(),
			"power", power.String())
	}

	x2EthContracts, x2EthDeployInfo, err := ethtxs.DeployAndInit(ethRelayer.clientSpec, para)
	if err != nil {
		return bridgeRegistry, err
	}
	ethRelayer.rwLock.Lock()
	ethRelayer.operatorInfo = &ethtxs.OperatorInfo{
		PrivateKey: deployPrivateKey,
		Address:    deployerAddr,
	}
	ethRelayer.deployPara = para
	ethRelayer.x2EthDeployInfo = x2EthDeployInfo
	ethRelayer.x2EthContracts = x2EthContracts
	bridgeRegistry = x2EthDeployInfo.BridgeRegistry.Address.String()
	_ = ethRelayer.setBridgeRegistryAddr(bridgeRegistry)
	//设置注册合约地址，同时设置启动中继服务的信号
	ethRelayer.bridgeRegistryAddr = x2EthDeployInfo.BridgeRegistry.Address
	ethRelayer.rwLock.Unlock()
	ethRelayer.unlockchan <- start
	relayerLog.Info("deploy", "the BridgeRegistry address is", bridgeRegistry)

	return bridgeRegistry, nil
}

//GetBalance ：获取某一个币种的余额
func (ethRelayer *Relayer4Ethereum) GetBalance(tokenAddr, owner string) (string, error) {
	return ethtxs.GetBalance(ethRelayer.clientSpec, tokenAddr, owner)
}

func (ethRelayer *Relayer4Ethereum) ShowMultiBalance(tokenAddr, owner string) (string, error) {
	relayerLog.Info("ShowMultiBalance", "tokenAddr", tokenAddr, "owner", owner)
	opts := &bind.CallOpts{
		From:    ethRelayer.ethValidator,
		Context: context.Background(),
	}

	gnosisSafeAddr := common.HexToAddress(ethRelayer.mulSignAddr)
	gnosisSafeInt, err := gnosis.NewGnosisSafe(gnosisSafeAddr, ethRelayer.clientSpec)
	if nil != err {
		return "", err
	}

	balance, err := gnosisSafeInt.GetSelfBalance(opts)
	if nil != err {
		return "", err
	}
	return balance.String(), nil
}

//ShowBridgeBankAddr ...
func (ethRelayer *Relayer4Ethereum) ShowBridgeBankAddr() (string, error) {
	if nil == ethRelayer.x2EthDeployInfo {
		return "", errors.New("the relayer is not started yes")
	}

	return ethRelayer.x2EthDeployInfo.BridgeBank.Address.String(), nil
}

//ShowBridgeRegistryAddr ...
func (ethRelayer *Relayer4Ethereum) ShowBridgeRegistryAddr() (string, error) {
	if nil == ethRelayer.x2EthDeployInfo {
		return "", errors.New("the relayer is not started yes")
	}

	return ethRelayer.x2EthDeployInfo.BridgeRegistry.Address.String(), nil
}

//ShowLockStatics ...
func (ethRelayer *Relayer4Ethereum) ShowLockStatics(tokenAddr string) (string, error) {
	return ethtxs.GetLockedFunds(ethRelayer.x2EthContracts.BridgeBank, tokenAddr)
}

//ShowDepositStatics ...
func (ethRelayer *Relayer4Ethereum) ShowDepositStatics(tokenAddr string) (string, error) {
	return ethtxs.GetDepositFunds(ethRelayer.clientSpec, tokenAddr)
}

//ShowTokenAddrBySymbol ...
func (ethRelayer *Relayer4Ethereum) ShowTokenAddrBySymbol(tokenSymbol string) (string, error) {
	return ethtxs.GetToken2address(ethRelayer.x2EthContracts.BridgeBank, tokenSymbol)
}

func (ethRelayer *Relayer4Ethereum) ShowLockedTokenAddress(tokenSymbol string) (string, error) {
	return ethtxs.GetLockedTokenAddress(ethRelayer.x2EthContracts.BridgeBank, tokenSymbol)
}

//IsProphecyPending ...
func (ethRelayer *Relayer4Ethereum) IsProphecyPending(claimID [32]byte) (bool, error) {
	return ethtxs.IsProphecyPending(claimID, ethRelayer.ethValidator, ethRelayer.x2EthContracts.Chain33Bridge)
}

//CreateBridgeToken ...
func (ethRelayer *Relayer4Ethereum) CreateBridgeToken(symbol string) (string, error) {
	ethRelayer.rwLock.RLock()
	tokenAddr, err := ethtxs.CreateBridgeToken(symbol, ethRelayer.clientSpec, ethRelayer.operatorInfo, ethRelayer.x2EthDeployInfo, ethRelayer.x2EthContracts)
	ethRelayer.rwLock.RUnlock()
	if nil == err {
		token2set := ebTypes.TokenAddress{
			Address:   tokenAddr,
			Symbol:    symbol,
			ChainName: ebTypes.EthereumBlockChainName,
		}
		_ = ethRelayer.SetTokenAddress(token2set)
	}
	return tokenAddr, err
}

// AddToken2LockList ...
func (ethRelayer *Relayer4Ethereum) AddToken2LockList(symbol, token string) (string, error) {
	txhash, err := ethtxs.AddToken2LockList(symbol, token, ethRelayer.clientSpec, ethRelayer.operatorInfo, ethRelayer.x2EthContracts)
	return txhash, err
}

//DeployERC20 ...
func (ethRelayer *Relayer4Ethereum) DeployERC20(ownerAddr, name, symbol, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	ethRelayer.rwLock.RLock()
	defer ethRelayer.rwLock.RUnlock()
	return ethtxs.DeployERC20(ownerAddr, name, symbol, bn, ethRelayer.clientSpec, ethRelayer.operatorInfo)
}

//ApproveAllowance ...
func (ethRelayer *Relayer4Ethereum) ApproveAllowance(ownerPrivateKey, tokenAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.ApproveAllowance(ownerPrivateKey, tokenAddr, ethRelayer.x2EthDeployInfo.BridgeBank.Address, bn, ethRelayer.clientSpec)
}

//Burn ...
func (ethRelayer *Relayer4Ethereum) Burn(ownerPrivateKey, tokenAddr, chain33Receiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.Burn(ownerPrivateKey, tokenAddr, chain33Receiver, ethRelayer.x2EthDeployInfo.BridgeBank.Address, bn, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.clientSpec)
}

//BurnAsync ...
func (ethRelayer *Relayer4Ethereum) BurnAsync(ownerPrivateKey, tokenAddr, chain33Receiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.BurnAsync(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.clientSpec)
}

//TransferToken ...
func (ethRelayer *Relayer4Ethereum) TransferToken(tokenAddr, fromKey, toAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.TransferToken(tokenAddr, fromKey, toAddr, bn, ethRelayer.clientSpec)
}

//TransferEth ...
func (ethRelayer *Relayer4Ethereum) TransferEth(fromKey, toAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.TransferEth(fromKey, toAddr, bn, ethRelayer.clientSpec)
}

//GetDecimals ...
func (ethRelayer *Relayer4Ethereum) GetDecimals(tokenAddr string) (uint8, error) {
	opts := &bind.CallOpts{
		Pending: true,
		From:    common.HexToAddress(tokenAddr),
		Context: context.Background(),
	}
	bridgeToken, _ := generated.NewBridgeToken(common.HexToAddress(tokenAddr), ethRelayer.clientSpec)
	return bridgeToken.Decimals(opts)
}

//LockEthErc20Asset ...
func (ethRelayer *Relayer4Ethereum) LockEthErc20Asset(ownerPrivateKey, tokenAddr, amount string, chain33Receiver string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.LockEthErc20Asset(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.clientSpec, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.x2EthDeployInfo.BridgeBank.Address)
}

//LockEthErc20AssetAsync ...
func (ethRelayer *Relayer4Ethereum) LockEthErc20AssetAsync(ownerPrivateKey, tokenAddr, amount string, chain33Receiver string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.LockEthErc20AssetAsync(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.clientSpec, ethRelayer.x2EthContracts.BridgeBank)
}

//ShowTxReceipt ...
func (ethRelayer *Relayer4Ethereum) ShowTxReceipt(hash string) (*types.Receipt, error) {
	txhash := common.HexToHash(hash)
	return ethRelayer.clientSpec.TransactionReceipt(context.Background(), txhash)
}

func (ethRelayer *Relayer4Ethereum) proc() {
	//等待用户导入
	relayerLog.Info("Please unlock or import private key for Ethereum relayer")
	if err := ethRelayer.RestoreTokenAddress(); nil != err {
		relayerLog.Info("Failed to RestoreTokenAddress")
		return
	}

	nilAddr := common.Address{}
	if nilAddr != ethRelayer.bridgeRegistryAddr {
		relayerLog.Info("proc", "Going to recover corresponding solidity contract handler with bridgeRegistryAddr", ethRelayer.bridgeRegistryAddr.String())
		var err error
		ethRelayer.rwLock.Lock()
		ethRelayer.x2EthContracts, ethRelayer.x2EthDeployInfo, err = ethtxs.RecoverContractHandler(ethRelayer.clientSpec, ethRelayer.bridgeRegistryAddr, ethRelayer.bridgeRegistryAddr)
		if nil != err {
			panic("Failed to recover corresponding solidity contract handler due to:" + err.Error())
		}
		ethRelayer.rwLock.Unlock()
		relayerLog.Info("^-^ ^-^ Succeed to recover corresponding solidity contract handler")
		//if nil != ethRelayer.recoverDeployPara() {
		//	panic("Failed to recoverDeployPara")
		//}
		ethRelayer.unlockchan <- start
	}

	var timer *time.Ticker
	ctx := context.Background()
	continueFailCount := int32(0)
	for range ethRelayer.unlockchan {
		relayerLog.Info("Received ethRelayer.unlockchan")
		ethRelayer.rwLock.RLock()
		privateKey4Ethereum := ethRelayer.privateKey4Ethereum
		ethRelayer.rwLock.RUnlock()
		if nil != privateKey4Ethereum && nilAddr != ethRelayer.bridgeRegistryAddr {
			relayerLog.Info("Ethereum relayer starts to run...")
			ethRelayer.prePareSubscribeEvent()
			//向bridgeBank订阅事件
			ethRelayer.subscribeEvent()
			ethRelayer.filterLogEvents()
			relayerLog.Info("Ethereum relayer starts to process online log event...")
			timer = time.NewTicker(time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
			goto latter
		}
	}

latter:
	for {
		select {
		case <-timer.C:
			ethRelayer.procNewHeight(ctx, &continueFailCount)
		case err := <-ethRelayer.bridgeBankSub.Err():
			panic("bridgeBankSub" + err.Error())
		case vLog := <-ethRelayer.bridgeBankLog:
			ethRelayer.storeBridgeBankLogs(vLog, true)
		case chain33Msg := <-ethRelayer.chain33MsgChan:
			ethRelayer.handleChain33Msg(chain33Msg)
		}
	}
}

func (ethRelayer *Relayer4Ethereum) handleChain33Msg(chain33Msg *events.Chain33Msg) {
	relayerLog.Info("handleChain33Msg", "Received chain33Msg", chain33Msg, "tx hash string", common.Bytes2Hex(chain33Msg.TxHash))

	// Parse the Chain33Msg into a ProphecyClaim for relay to Ethereum
	prophecyClaim := ethtxs.Chain33MsgToProphecyClaim(*chain33Msg)
	var tokenAddr common.Address
	exist := false
	if chain33Msg.ClaimType == events.ClaimTypeLock {
		tokenAddr, exist = ethRelayer.symbol2Addr[prophecyClaim.Symbol]
		if !exist {
			relayerLog.Info("handleChain33Msg", "Query address from ethereum for symbol", prophecyClaim.Symbol)
			//因为是lock操作，则需要从创建的bridgeToken中进行查询
			addr, err := ethRelayer.ShowTokenAddrBySymbol(prophecyClaim.Symbol)
			if err != nil {
				panic(fmt.Sprintf("Pls create bridge token in advance for token:%s", prophecyClaim.Symbol))
			}
			token2set := ebTypes.TokenAddress{
				Address:   addr,
				Symbol:    prophecyClaim.Symbol,
				ChainName: ebTypes.EthereumBlockChainName,
			}
			err = ethRelayer.SetTokenAddress(token2set)
			if nil != err {
				// 尽管设置数据失败，但是不影响运行，只是relayer启动时，每次从节点远程获取bridge token地址而已
				relayerLog.Error("handleChain33Msg", "Failed to SetTokenAddress due to", err.Error())
			}
			tokenAddr = common.HexToAddress(addr)
		}
	} else {
		tokenAddr, exist = ethRelayer.symbol2LockAddr[prophecyClaim.Symbol]
		if !exist {
			//因为是burn操作，必须从允许lock的token地址中进行查询
			addr, err := ethRelayer.ShowLockedTokenAddress(prophecyClaim.Symbol)
			if err != nil {
				panic(fmt.Sprintf("Pls create lock token in advance for token:%s", prophecyClaim.Symbol))
			}
			token2set := ebTypes.TokenAddress{
				Address:   addr,
				Symbol:    prophecyClaim.Symbol,
				ChainName: ebTypes.EthereumBlockChainName,
			}
			err = ethRelayer.SetLockedTokenAddress(token2set)
			if nil != err {
				relayerLog.Error("handleChain33Msg", "Failed to SetLockedTokenAddress due to", err.Error())
			}
			tokenAddr = common.HexToAddress(addr)
		}
	}

	// Relay the Chain33Msg to the Ethereum network
	txhash, err := ethtxs.RelayOracleClaimToEthereum(ethRelayer.x2EthContracts.Oracle, ethRelayer.clientSpec, ethRelayer.ethSender, tokenAddr, prophecyClaim, ethRelayer.privateKey4Ethereum)
	if nil != err {
		panic("RelayOracleClaimToEthereum failed due to" + err.Error())
	}
	relayerLog.Info("handleChain33Msg", "RelayOracleClaimToEthereum with tx hash", txhash)

	//保存交易hash，方便查询
	atomic.AddInt64(&ethRelayer.totalTx4Eth2Chain33, 1)
	txIndex := atomic.LoadInt64(&ethRelayer.totalTx4Eth2Chain33)
	if err = ethRelayer.updateTotalTxAmount2chain33(txIndex); nil != err {
		relayerLog.Error("handleChain33Msg", "Failed to RelayLockToChain33 due to:", err.Error())
		return
	}
	statics := &ebTypes.Chain33ToEthereumStatics{
		EthTxstatus:      ebTypes.Tx_Status_Pending,
		Chain33Txhash:    common.Bytes2Hex(chain33Msg.TxHash),
		EthereumTxhash:   txhash,
		BurnLock:         int32(chain33Msg.ClaimType),
		Chain33Sender:    chain33Msg.Chain33Sender.String(),
		EthereumReceiver: chain33Msg.EthereumReceiver.String(),
		Symbol:           chain33Msg.Symbol,
		Amount:           chain33Msg.Amount.String(),
		Nonce:            chain33Msg.Nonce,
		TxIndex:          txIndex,
	}
	data := chain33Types.Encode(statics)
	if err = ethRelayer.setLastestStatics(int32(chain33Msg.ClaimType), txIndex, data); nil != err {
		relayerLog.Error("handleChain33Msg", "Failed to RelayLockToChain33 due to:", err.Error())
		return
	}
}

func (ethRelayer *Relayer4Ethereum) procNewHeight(ctx context.Context, continueFailCount *int32) {
	head, err := ethRelayer.clientSpec.HeaderByNumber(ctx, nil)
	if nil != err {
		*continueFailCount++
		if *continueFailCount >= (12 * 5) {
			panic(err.Error())
		}
		relayerLog.Error("Failed to get ethereum height", "provider", ethRelayer.provider,
			"continueFailCount", continueFailCount)
		return
	}
	ethRelayer.updateTxStatus()
	*continueFailCount = 0
	currentHeight := head.Number.Uint64()
	relayerLog.Info("procNewHeight", "currentHeight", currentHeight, "ethRelayer.eventLogIndex.Height", ethRelayer.eventLogIndex.Height, "uint64(ethRelayer.maturityDegree)", uint64(ethRelayer.maturityDegree))

	//一次最大只获取10个logEvent进行处理
	fetchCnt := int32(10)
	for ethRelayer.eventLogIndex.Height+uint64(ethRelayer.maturityDegree)+1 <= currentHeight {
		logs, err := ethRelayer.getNextValidEthTxEventLogs(ethRelayer.eventLogIndex.Height, ethRelayer.eventLogIndex.Index, fetchCnt)
		if nil != err {
			relayerLog.Error("Failed to get ethereum height", "getNextValidEthTxEventLogs err", err.Error())
			return
		}

		for i, vLog := range logs {
			if vLog.BlockNumber+uint64(ethRelayer.maturityDegree)+1 > currentHeight {
				logs = logs[:i]
				break
			}
			ethRelayer.procBridgeBankLogs(*vLog)
		}

		cnt := int32(len(logs))
		if len(logs) > 0 {
			//firstHeight := logs[0].BlockNumber
			lastHeight := logs[cnt-1].BlockNumber
			index := logs[cnt-1].TxIndex
			//获取的数量小于批量获取数量，则认为直接
			ethRelayer.setBridgeBankProcessedHeight(lastHeight, uint32(index))
			ethRelayer.eventLogIndex.Height = lastHeight
			ethRelayer.eventLogIndex.Index = uint32(index)
		}

		//当前需要处理的event数量已经少于10个，直接返回
		if cnt < fetchCnt {
			return
		}
	}
}

func (ethRelayer *Relayer4Ethereum) storeBridgeBankLogs(vLog types.Log, setBlockNumber bool) {
	//lock,用于捕捉 (ETH/ERC20----->chain33) 跨链转移
	//burn,用于捕捉 (chain33 token----->chain33) 实现chain33资产withdraw操作，之后在chain33上实现unlock操作
	if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventLockSig {
		//先进行数据的持久化，等到一定的高度成熟度之后再进行处理
		relayerLog.Info("Relayer4Ethereum storeBridgeBankLogs", "^_^ ^_^ Received bridgeBankLog for event", "lock",
			"Block number:", vLog.BlockNumber, "tx Index", vLog.TxIndex, "log Index", vLog.Index, "Tx hash:", vLog.TxHash.Hex())
		if err := ethRelayer.setEthTxEvent(vLog); nil != err {
			panic(err.Error())
		}
	} else if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventBurnSig {
		relayerLog.Info("Relayer4Ethereum storeBridgeBankLogs", "^_^ ^_^ Received bridgeBankLog for event", "burn",
			"Block number:", vLog.BlockNumber, "tx Index", vLog.TxIndex, "log Index", vLog.Index, "Tx hash:", vLog.TxHash.Hex())
		if err := ethRelayer.setEthTxEvent(vLog); nil != err {
			panic(err.Error())
		}
	}

	//确定是否需要更新保存同步日志高度
	if setBlockNumber {
		if err := ethRelayer.setHeight4BridgeBankLogAt(vLog.BlockNumber); nil != err {
			panic(err.Error())
		}
	}
}

func (ethRelayer *Relayer4Ethereum) procBridgeBankLogs(vLog types.Log) {
	if ethRelayer.checkTxProcessed(vLog.TxHash.Bytes()) {
		relayerLog.Info("procBridgeBankLogs", "Tx has been already Processed with hash:", vLog.TxHash.Hex(),
			"height", vLog.BlockNumber, "index", vLog.Index)
		return
	}

	defer func() {
		if err := ethRelayer.setTxProcessed(vLog.TxHash.Bytes()); nil != err {
			panic(err.Error())
		}
	}()

	//检查当前交易是否因为区块回退而导致交易丢失
	receipt, err := ethRelayer.clientSpec.TransactionReceipt(context.Background(), vLog.TxHash)
	if nil != err {
		relayerLog.Error("procBridgeBankLogs", "Failed to get tx receipt with hash", vLog.TxHash.String())
		return
	}

	//检查当前的交易是否成功执行
	if receipt.Status != types.ReceiptStatusSuccessful {
		relayerLog.Error("procBridgeBankLogs", "tx not successful with status", receipt.Status)
		return
	}

	//lock,用于捕捉 (ETH/ERC20----->chain33) 跨链转移
	if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventLockSig {
		eventName := events.LogLock.String()
		relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
			"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
		err := ethRelayer.handleLogLockEvent(ethRelayer.clientChainID, ethRelayer.bridgeBankAbi, eventName, vLog)
		if err != nil {
			errinfo := fmt.Sprintf("Failed to handleLogLockEvent due to:%s", err.Error())
			relayerLog.Info("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
			panic(errinfo)
		}
	} else if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventBurnSig {
		//burn,用于捕捉 (chain33 token----->chain33) 实现chain33资产withdraw操作，之后在chain33上实现unlock操作
		eventName := events.LogChain33TokenBurn.String()
		relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
			"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
		err := ethRelayer.handleLogBurnEvent(ethRelayer.clientChainID, ethRelayer.bridgeBankAbi, eventName, vLog)
		if err != nil {
			errinfo := fmt.Sprintf("Failed to handleLogBurnEvent due to:%s", err.Error())
			relayerLog.Info("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
			panic(errinfo)
		}
	}
}

func (ethRelayer *Relayer4Ethereum) filterLogEvents() {
	deployHeight, _ := ethtxs.GetDeployHeight(ethRelayer.clientSpec, ethRelayer.x2EthDeployInfo.BridgeRegistry.Address, ethRelayer.x2EthDeployInfo.BridgeRegistry.Address)
	height4BridgeBankLogAt := int64(ethRelayer.getHeight4BridgeBankLogAt())

	if height4BridgeBankLogAt < deployHeight {
		height4BridgeBankLogAt = deployHeight
	}

	header, err := ethRelayer.clientSpec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		errinfo := fmt.Sprintf("Failed to get HeaderByNumbers due to:%s", err.Error())
		panic(errinfo)
	}
	curHeight := int64(header.Number.Uint64())
	relayerLog.Info("filterLogEvents", "curHeight:", curHeight)

	bridgeBankSig := make(map[string]bool)
	bridgeBankSig[ethRelayer.bridgeBankEventLockSig] = true
	bridgeBankSig[ethRelayer.bridgeBankEventBurnSig] = true
	bridgeBankLog := make(chan types.Log)
	done := make(chan int)

	go ethRelayer.filterLogEventsProc(bridgeBankLog, done, "bridgeBank", curHeight, height4BridgeBankLogAt, ethRelayer.bridgeBankAddr, bridgeBankSig)

	for {
		select {
		case vLog := <-bridgeBankLog:
			ethRelayer.storeBridgeBankLogs(vLog, true)
		case vLog := <-ethRelayer.bridgeBankLog:
			//因为此处是同步保存信息，防止未同步完成出现panic时，直接将其设置为最新高度，中间出现部分信息不同步的情况
			ethRelayer.storeBridgeBankLogs(vLog, false)
		case <-done:
			relayerLog.Info("Finshed offline logs processed")
			return
		}
	}
}

func (ethRelayer *Relayer4Ethereum) filterLogEventsProc(logchan chan<- types.Log, done chan<- int, title string, curHeight, heightLogProcAt int64, contractAddr common.Address, eventSig map[string]bool) {
	relayerLog.Info(title, "eventSig", eventSig, "heightLogProcAt", heightLogProcAt, "curHeight", curHeight)

	startHeight := heightLogProcAt
	batchCount := int64(10)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
	}

	for {
		if batchCount < (curHeight - startHeight + 1) {
			stopHeight := startHeight + batchCount - 1
			query.FromBlock = big.NewInt(startHeight)
			query.ToBlock = big.NewInt(stopHeight)
		} else {
			query.FromBlock = big.NewInt(startHeight)
			query.ToBlock = big.NewInt(curHeight)
		}

		// Filter by contract and event, write results to logs
		logs, err := ethRelayer.clientSpec.FilterLogs(context.Background(), query)
		if err != nil {
			errinfo := fmt.Sprintf("Failed to filterLogEvents due to:%s", err.Error())
			panic(errinfo)
		}

		relayerLog.Info(title, "received logs with number", len(logs),
			"start height", query.FromBlock.String(), "stop height", query.ToBlock.String())
		for _, logv := range logs {
			relayerLog.Info(title, "received log with topics", logv.Topics[0].Hex(), "BlockNumber", logv.BlockNumber)
			if _, exist := eventSig[logv.Topics[0].Hex()]; !exist {
				continue
			}
			logchan <- logv
			relayerLog.Info(title, "get unprocessed log with topic:", logv.Topics[0].String(),
				"BlockNumber", logv.BlockNumber)
		}
		//更新
		if err := ethRelayer.setHeight4BridgeBankLogAt(query.ToBlock.Uint64()); nil != err {
			panic(err.Error())
		}

		if query.ToBlock.Int64() == curHeight {
			relayerLog.Info(title, "Finished FilterLogs to height", curHeight)
			done <- 1
			break
		}
		startHeight = query.ToBlock.Int64() + 1
	}
}

func (ethRelayer *Relayer4Ethereum) prePareSubscribeEvent() {
	var eventName string
	//bridgeBank处理
	contactAbi := ethtxs.LoadABI(ethtxs.BridgeBankABI)
	ethRelayer.bridgeBankAbi = contactAbi
	eventName = events.LogLock.String()
	ethRelayer.bridgeBankEventLockSig = contactAbi.Events[eventName].ID.Hex()
	eventName = events.LogChain33TokenBurn.String()
	ethRelayer.bridgeBankEventBurnSig = contactAbi.Events[eventName].ID.Hex()
	ethRelayer.bridgeBankAddr = ethRelayer.x2EthDeployInfo.BridgeBank.Address
}

func (ethRelayer *Relayer4Ethereum) subscribeEvent() {
	targetAddress := ethRelayer.bridgeBankAddr

	// We need the target address in bytes[] for the query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{targetAddress},
		FromBlock: big.NewInt(int64(1)),
	}
	// We will check logs for new events
	logs := make(chan types.Log, 10)
	// Filter by contract and event, write results to logs
	sub, err := ethRelayer.clientSpec.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		errinfo := fmt.Sprintf("Failed to SubscribeFilterLogs due to:%s", err.Error())
		panic(errinfo)
	}
	relayerLog.Info("subscribeEvent", "Subscribed to contract at address:", targetAddress.Hex())
	ethRelayer.bridgeBankLog = logs
	ethRelayer.bridgeBankSub = sub
}

//IsValidatorActive ...
func (ethRelayer *Relayer4Ethereum) IsValidatorActive(addr string) (bool, error) {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	if !re.MatchString(addr) {
		return false, errors.New("this address is not an ethereum address")
	}
	ethRelayer.rwLock.RLock()
	active, err := ethtxs.IsActiveValidator(common.HexToAddress(addr), ethRelayer.x2EthContracts.Valset)
	ethRelayer.rwLock.RUnlock()
	return active, err
}

//ShowOperator ...
func (ethRelayer *Relayer4Ethereum) ShowOperator() (string, error) {
	operator, err := ethtxs.GetOperator(ethRelayer.clientSpec, ethRelayer.ethValidator, ethRelayer.bridgeBankAddr)
	if nil != err {
		return "", err
	}
	return operator.String(), nil
}

//QueryTxhashRelay2Chain33 ...
func (ethRelayer *Relayer4Ethereum) QueryTxhashRelay2Chain33() ebTypes.Txhashes {
	txhashs := ethRelayer.queryTxhashes([]byte(chain33ToEthStaticsPrefix))
	return ebTypes.Txhashes{Txhash: txhashs}
}

// handleLogLockEvent : unpacks a LogLock event, converts it to a ProphecyClaim, and relays a tx to chain33
func (ethRelayer *Relayer4Ethereum) handleLogLockEvent(clientChainID *big.Int, contractABI abi.ABI, eventName string, log types.Log) error {
	// Unpack the LogLock event using its unique event signature from the contract's ABI
	event, err := events.UnpackLogLock(contractABI, eventName, log.Data)
	if nil != err {
		return err
	}

	var decimal uint8
	if event.Token.String() == "" || event.Token.String() == "0x0000000000000000000000000000000000000000" {
		decimal = 18
	} else {
		opts := &bind.CallOpts{
			Pending: true,
			From:    common.HexToAddress(event.Token.String()),
			Context: context.Background(),
		}
		bridgeToken, _ := generated.NewBridgeToken(common.HexToAddress(event.Token.String()), ethRelayer.clientSpec)

		decimal, err = bridgeToken.Decimals(opts)
		if err != nil {
			return err
		}
	}

	// Parse the LogLock event's payload into a struct
	prophecyClaim, err := ethtxs.LogLockToEthBridgeClaim(event, clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), log.TxHash.String(), int64(decimal))
	if err != nil {
		return err
	}

	ethRelayer.ethBridgeClaimChan <- prophecyClaim

	return nil
}

// handleLogBurnEvent : unpacks a burn event, converts it to a ProphecyClaim, and relays a tx to chain33
func (ethRelayer *Relayer4Ethereum) handleLogBurnEvent(clientChainID *big.Int, contractABI abi.ABI, eventName string, log types.Log) error {
	event, err := events.UnpackLogBurn(contractABI, eventName, log.Data)
	if nil != err {
		return err
	}

	var decimal uint8
	if event.Token.String() == "" || event.Token.String() == "0x0000000000000000000000000000000000000000" {
		decimal = 18
	} else {
		opts := &bind.CallOpts{
			Pending: true,
			From:    common.HexToAddress(event.Token.String()),
			Context: context.Background(),
		}
		bridgeToken, _ := generated.NewBridgeToken(common.HexToAddress(event.Token.String()), ethRelayer.clientSpec)

		decimal, err = bridgeToken.Decimals(opts)
		if err != nil {
			return err
		}
	}

	// Parse the LogLock event's payload into a struct
	prophecyClaim, err := ethtxs.LogBurnToEthBridgeClaim(event, clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), log.TxHash.String(), int64(decimal))
	if err != nil {
		return err
	}

	ethRelayer.ethBridgeClaimChan <- prophecyClaim

	return nil
}

func (ethRelayer *Relayer4Ethereum) ShowStatics(request ebTypes.TokenStaticsRequest) (*ebTypes.TokenStaticsResponse, error) {
	res := &ebTypes.TokenStaticsResponse{}

	datas, err := ethRelayer.getStatics(request.Operation, request.TxIndex)
	if nil != err {
		return nil, err
	}

	for _, data := range datas {
		var statics ebTypes.Chain33ToEthereumStatics
		_ = chain33Types.Decode(data, &statics)
		if request.Status != 0 {
			if ebTypes.Tx_Status_Map[request.Status] != statics.EthTxstatus {
				continue
			}
		}
		res.C2Estatics = append(res.C2Estatics, &statics)
	}
	return res, nil
}

func (ethRelayer *Relayer4Ethereum) updateTxStatus() {
	ethRelayer.updateSingleTxStatus(events.ClaimTypeBurn)
	ethRelayer.updateSingleTxStatus(events.ClaimTypeLock)
}

func (ethRelayer *Relayer4Ethereum) updateSingleTxStatus(claimType events.ClaimType) {
	txIndex := ethRelayer.getEthLockTxUpdateTxIndex(claimType)
	if ebTypes.Invalid_Tx_Index == txIndex {
		return
	}
	datas, _ := ethRelayer.getStatics(int32(claimType), txIndex)
	if nil == datas {
		return
	}
	for _, data := range datas {
		var statics ebTypes.Chain33ToEthereumStatics
		_ = chain33Types.Decode(data, &statics)
		receipt, _ := ethRelayer.clientSpec.TransactionReceipt(context.Background(), common.HexToHash(statics.EthereumTxhash))
		//当前处理机制比较简单，如果发现该笔交易未执行，就不再产寻后续交易的回执
		if nil == receipt {
			break
		}
		status := ebTypes.Tx_Status_Success
		if receipt.Status != types.ReceiptStatusSuccessful {
			status = ebTypes.Tx_Status_Failed
		}
		statics.EthTxstatus = status
		dataNew := chain33Types.Encode(&statics)
		_ = ethRelayer.setLastestStatics(int32(claimType), statics.TxIndex, dataNew)
		_ = ethRelayer.setEthLockTxUpdateTxIndex(statics.TxIndex, claimType)
		relayerLog.Info("updateSingleTxStatus", "txHash", statics.EthereumTxhash, "updated status", status)
	}
}

func (ethRelayer *Relayer4Ethereum) DeployMulsign() (mulsign string, err error) {
	mulsign, err = ethtxs.DeployMulSign2Eth(ethRelayer.clientSpec, ethRelayer.operatorInfo)
	if err != nil {
		return "", err
	}
	ethRelayer.rwLock.Lock()
	ethRelayer.mulSignAddr = mulsign
	ethRelayer.rwLock.Unlock()

	ethRelayer.setMultiSignAddress(mulsign)

	return mulsign, nil
}

func (ethRelayer *Relayer4Ethereum) SetupMulSign(setupMulSign ebTypes.SetupMulSign) (string, error) {
	if "" == ethRelayer.mulSignAddr {
		return "", ebTypes.ErrMulSignNotDeployed
	}

	return ethtxs.SetupMultiSign(setupMulSign.OperatorPrivateKey, ethRelayer.mulSignAddr, setupMulSign.Owners, ethRelayer.clientSpec)
}

func (ethRelayer *Relayer4Ethereum) SafeTransfer(para ebTypes.SafeTransfer) (string, error) {
	if "" == ethRelayer.mulSignAddr {
		return "", ebTypes.ErrMulSignNotDeployed
	}

	return ethtxs.SafeTransfer(para.OperatorPrivateKey, ethRelayer.mulSignAddr, para.To, para.Token, para.OwnerPrivateKeys, para.Amount, ethRelayer.clientSpec)
}

func (ethRelayer *Relayer4Ethereum) ConfigOfflineSaveAccount(addr string) (string, error) {
	txhash, err := ethtxs.ConfigOfflineSaveAccount(addr, ethRelayer.clientSpec, ethRelayer.operatorInfo, ethRelayer.x2EthContracts)
	return txhash, err
}

func (ethRelayer *Relayer4Ethereum) ConfigLockedTokenOfflineSave(addr, symbol, threshold string, percents uint32) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(threshold), 10)
	txhash, err := ethtxs.ConfigLockedTokenOfflineSave(addr, symbol, bn, uint8(percents), ethRelayer.clientSpec, ethRelayer.operatorInfo, ethRelayer.x2EthContracts)
	return txhash, err
}

func (ethRelayer *Relayer4Ethereum) SetMultiSignAddr(address string) {
	ethRelayer.rwLock.Lock()
	ethRelayer.mulSignAddr = address
	ethRelayer.rwLock.Unlock()

	ethRelayer.setMultiSignAddress(address)
}
