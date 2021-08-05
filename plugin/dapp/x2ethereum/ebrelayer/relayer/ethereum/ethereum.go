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

	chain33Crypto "github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	x2ethTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

//Relayer4Ethereum ...
type Relayer4Ethereum struct {
	provider               string
	clientChainID          *big.Int
	bridgeRegistryAddr     common.Address
	db                     dbm.DB
	rwLock                 sync.RWMutex
	privateKey4Chain33     chain33Crypto.PrivKey
	ethValidator           common.Address
	totalTx4Eth2Chain33    int64
	rpcURL2Chain33         string
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
}

var (
	relayerLog = log.New("module", "ethereum_relayer")
)

//const ...
const (
	DefaultBlockPeriod = 5000
)

//StartEthereumRelayer ///
func StartEthereumRelayer(rpcURL2Chain33 string, db dbm.DB, provider, registryAddress string, deploy *ebTypes.Deploy, degree, blockInterval int32) *Relayer4Ethereum {
	if 0 == blockInterval {
		blockInterval = DefaultBlockPeriod
	}
	ethRelayer := &Relayer4Ethereum{
		provider:            provider,
		db:                  db,
		unlockchan:          make(chan int, 2),
		rpcURL2Chain33:      rpcURL2Chain33,
		bridgeRegistryAddr:  common.HexToAddress(registryAddress),
		deployInfo:          deploy,
		maturityDegree:      degree,
		fetchHeightPeriodMs: blockInterval,
	}

	registrAddrInDB, err := ethRelayer.getBridgeRegistryAddr()
	//如果输入的registry地址非空，且和数据库保存地址不一致，则直接使用输入注册地址
	if registryAddress != "" && nil == err && registrAddrInDB != registryAddress {
		relayerLog.Error("StartEthereumRelayer", "BridgeRegistry is setted already with value", registrAddrInDB,
			"but now setting to", registryAddress)
		_ = ethRelayer.setBridgeRegistryAddr(registryAddress)
	} else if registryAddress == "" && registrAddrInDB != "" {
		//输入地址为空，且数据库中保存地址不为空，则直接使用数据库中的地址
		ethRelayer.bridgeRegistryAddr = common.HexToAddress(registrAddrInDB)
	}
	ethRelayer.eventLogIndex = ethRelayer.getLastBridgeBankProcessedHeight()
	ethRelayer.initBridgeBankTx()

	// Start clientSpec with infura ropsten provider
	relayerLog.Info("Relayer4Ethereum proc", "Started Ethereum websocket with provider:", ethRelayer.provider, "rpcURL2Chain33:", ethRelayer.rpcURL2Chain33)
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

//IsProphecyPending ...
func (ethRelayer *Relayer4Ethereum) IsProphecyPending(claimID [32]byte) (bool, error) {
	return ethtxs.IsProphecyPending(claimID, ethRelayer.ethValidator, ethRelayer.x2EthContracts.Chain33Bridge)
}

//CreateBridgeToken ...
func (ethRelayer *Relayer4Ethereum) CreateBridgeToken(symbol string) (string, error) {
	ethRelayer.rwLock.RLock()
	defer ethRelayer.rwLock.RUnlock()
	return ethtxs.CreateBridgeToken(symbol, ethRelayer.clientSpec, ethRelayer.operatorInfo, ethRelayer.x2EthDeployInfo, ethRelayer.x2EthContracts)
}

//CreateERC20Token ...
func (ethRelayer *Relayer4Ethereum) CreateERC20Token(symbol string) (string, error) {
	ethRelayer.rwLock.RLock()
	defer ethRelayer.rwLock.RUnlock()
	return ethtxs.CreateERC20Token(symbol, ethRelayer.clientSpec, ethRelayer.operatorInfo)
}

//MintERC20Token ...
func (ethRelayer *Relayer4Ethereum) MintERC20Token(tokenAddr, ownerAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(x2ethTypes.TrimZeroAndDot(amount), 10)
	ethRelayer.rwLock.RLock()
	defer ethRelayer.rwLock.RUnlock()
	return ethtxs.MintERC20Token(tokenAddr, ownerAddr, bn, ethRelayer.clientSpec, ethRelayer.operatorInfo)
}

//ApproveAllowance ...
func (ethRelayer *Relayer4Ethereum) ApproveAllowance(ownerPrivateKey, tokenAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(x2ethTypes.TrimZeroAndDot(amount), 10)
	return ethtxs.ApproveAllowance(ownerPrivateKey, tokenAddr, ethRelayer.x2EthDeployInfo.BridgeBank.Address, bn, ethRelayer.clientSpec)
}

//Burn ...
func (ethRelayer *Relayer4Ethereum) Burn(ownerPrivateKey, tokenAddr, chain33Receiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(x2ethTypes.TrimZeroAndDot(amount), 10)
	return ethtxs.Burn(ownerPrivateKey, tokenAddr, chain33Receiver, ethRelayer.x2EthDeployInfo.BridgeBank.Address, bn, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.clientSpec)
}

//BurnAsync ...
func (ethRelayer *Relayer4Ethereum) BurnAsync(ownerPrivateKey, tokenAddr, chain33Receiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(x2ethTypes.TrimZeroAndDot(amount), 10)
	return ethtxs.BurnAsync(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.clientSpec)
}

//TransferToken ...
func (ethRelayer *Relayer4Ethereum) TransferToken(tokenAddr, fromKey, toAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(x2ethTypes.TrimZeroAndDot(amount), 10)
	return ethtxs.TransferToken(tokenAddr, fromKey, toAddr, bn, ethRelayer.clientSpec)
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
	bn, _ = bn.SetString(x2ethTypes.TrimZeroAndDot(amount), 10)
	return ethtxs.LockEthErc20Asset(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.clientSpec, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.x2EthDeployInfo.BridgeBank.Address)
}

//LockEthErc20AssetAsync ...
func (ethRelayer *Relayer4Ethereum) LockEthErc20AssetAsync(ownerPrivateKey, tokenAddr, amount string, chain33Receiver string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(x2ethTypes.TrimZeroAndDot(amount), 10)
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
		if nil != ethRelayer.recoverDeployPara() {
			panic("Failed to recoverDeployPara")
		}
		ethRelayer.unlockchan <- start
	}

	var timer *time.Ticker
	ctx := context.Background()
	continueFailCount := int32(0)
	for range ethRelayer.unlockchan {
		relayerLog.Info("Received ethRelayer.unlockchan")
		ethRelayer.rwLock.RLock()
		privateKey4Chain33 := ethRelayer.privateKey4Chain33
		ethRelayer.rwLock.RUnlock()
		if nil != privateKey4Chain33 && nilAddr != ethRelayer.bridgeRegistryAddr {
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
		}
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
	*continueFailCount = 0

	currentHeight := head.Number.Uint64()
	relayerLog.Info("procNewHeight", "currentHeight", currentHeight)
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
	relayerLog.Info(title, "eventSig", eventSig)

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

//QueryTxhashRelay2Eth ...
func (ethRelayer *Relayer4Ethereum) QueryTxhashRelay2Eth() ebTypes.Txhashes {
	txhashs := ethRelayer.queryTxhashes([]byte(chain33ToEthTxHashPrefix))
	return ebTypes.Txhashes{Txhash: txhashs}
}

//QueryTxhashRelay2Chain33 ...
func (ethRelayer *Relayer4Ethereum) QueryTxhashRelay2Chain33() ebTypes.Txhashes {
	txhashs := ethRelayer.queryTxhashes([]byte(eth2chain33TxHashPrefix))
	return ebTypes.Txhashes{Txhash: txhashs}
}

// handleLogLockEvent : unpacks a LogLock event, converts it to a ProphecyClaim, and relays a tx to chain33
func (ethRelayer *Relayer4Ethereum) handleLogLockEvent(clientChainID *big.Int, contractABI abi.ABI, eventName string, log types.Log) error {
	rpcURL := ethRelayer.rpcURL2Chain33

	// Unpack the LogLock event using its unique event signature from the contract's ABI
	event, err := events.UnpackLogLock(contractABI, eventName, log.Data)
	if nil != err {
		return err
	}
	// Add the event to the record
	events.NewEventWrite(log.TxHash.Hex(), *event)

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
	prophecyClaim, err := ethtxs.LogLockToEthBridgeClaim(event, clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), int64(decimal))
	if err != nil {
		return err
	}

	// Initiate the relay
	ethRelayer.rwLock.RLock()
	privateKey4Chain33 := ethRelayer.privateKey4Chain33
	ethRelayer.rwLock.RUnlock()
	txhash, err := ethtxs.RelayLockToChain33(privateKey4Chain33, prophecyClaim, rpcURL)
	if err != nil {
		relayerLog.Error("handleLogLockEvent", "Failed to RelayLockToChain33 due to:", err.Error())
		return err
	}
	relayerLog.Info("handleLogLockEvent", "RelayLockToChain33 with hash", txhash)

	//保存交易hash，方便查询
	atomic.AddInt64(&ethRelayer.totalTx4Eth2Chain33, 1)
	txIndex := atomic.LoadInt64(&ethRelayer.totalTx4Eth2Chain33)
	if err = ethRelayer.updateTotalTxAmount2chain33(txIndex); nil != err {
		relayerLog.Error("handleLogLockEvent", "Failed to RelayLockToChain33 due to:", err.Error())
		return err
	}
	if err = ethRelayer.setLastestRelay2Chain33Txhash(txhash, txIndex); nil != err {
		relayerLog.Error("handleLogLockEvent", "Failed to RelayLockToChain33 due to:", err.Error())
		return err
	}

	return nil
}

// handleLogBurnEvent : unpacks a burn event, converts it to a ProphecyClaim, and relays a tx to chain33
func (ethRelayer *Relayer4Ethereum) handleLogBurnEvent(clientChainID *big.Int, contractABI abi.ABI, eventName string, log types.Log) error {
	rpcURL := ethRelayer.rpcURL2Chain33

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
	prophecyClaim, err := ethtxs.LogBurnToEthBridgeClaim(event, clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), int64(decimal))
	if err != nil {
		return err
	}

	// Initiate the relay
	ethRelayer.rwLock.RLock()
	privateKey4Chain33 := ethRelayer.privateKey4Chain33
	ethRelayer.rwLock.RUnlock()
	txhash, err := ethtxs.RelayBurnToChain33(privateKey4Chain33, prophecyClaim, rpcURL)
	if err != nil {
		relayerLog.Error("handleLogLockEvent", "Failed to RelayLockToChain33 due to:", err.Error())
		return err
	}
	relayerLog.Info("handleLogLockEvent", "RelayBurnToChain33 with hash", txhash)

	//保存交易hash，方便查询
	atomic.AddInt64(&ethRelayer.totalTx4Eth2Chain33, 1)
	txIndex := atomic.LoadInt64(&ethRelayer.totalTx4Eth2Chain33)
	if err = ethRelayer.updateTotalTxAmount2chain33(txIndex); nil != err {
		relayerLog.Error("handleLogBurnEvent", "Failed to RelayLockToChain33 due to:", err.Error())
		return err
	}
	if err = ethRelayer.setLastestRelay2Chain33Txhash(txhash, txIndex); nil != err {
		relayerLog.Error("handleLogBurnEvent", "Failed to RelayLockToChain33 due to:", err.Error())
		return err
	}

	return nil
}
