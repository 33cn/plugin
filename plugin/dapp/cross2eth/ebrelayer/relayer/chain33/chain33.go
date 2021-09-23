package chain33

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	chain33Crypto "github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	chain33Types "github.com/33cn/chain33/types"
	syncTx "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/chain33/transceiver/sync"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var relayerLog = log.New("module", "chain33_relayer")

//Relayer4Chain33 ...
type Relayer4Chain33 struct {
	syncEvmTxLogs       *syncTx.EVMTxLogs
	rpcLaddr            string //用户向指定的blockchain节点进行rpc调用
	chainName           string //用来区别主链中继还是平行链，主链为空，平行链则是user.p.xxx.
	chainID             int32
	fetchHeightPeriodMs int64
	db                  dbm.DB
	lastHeight4Tx       int64 //等待被处理的具有相应的交易回执的高度
	matDegree           int32 //成熟度         heightSync2App    matDegress   height

	privateKey4Chain33       chain33Crypto.PrivKey
	privateKey4Chain33_ecdsa *ecdsa.PrivateKey
	ctx                      context.Context
	rwLock                   sync.RWMutex
	unlockChan               chan int
	bridgeBankEventLockSig   string
	bridgeBankEventBurnSig   string
	bridgeBankAbi            abi.ABI
	deployInfo               *ebTypes.Deploy
	totalTx4RelayEth2chai33  int64
	//新增//
	ethBridgeClaimChan <-chan *ebTypes.EthBridgeClaim
	chain33MsgChan     chan<- *events.Chain33Msg
	bridgeRegistryAddr string
	oracleAddr         string
	bridgeBankAddr     string
	mulSignAddr        string
	deployResult       *X2EthDeployResult
	symbol2Addr        map[string]string
}

type Chain33StartPara struct {
	ChainName          string
	Ctx                context.Context
	SyncTxConfig       *ebTypes.SyncTxConfig
	BridgeRegistryAddr string
	DeployInfo         *ebTypes.Deploy
	DBHandle           dbm.DB
	EthBridgeClaimChan <-chan *ebTypes.EthBridgeClaim
	Chain33MsgChan     chan<- *events.Chain33Msg
	ChainID            int32
}

// StartChain33Relayer : initializes a relayer which witnesses events on the chain33 network and relays them to Ethereum
func StartChain33Relayer(startPara *Chain33StartPara) *Relayer4Chain33 {
	chain33Relayer := &Relayer4Chain33{
		rpcLaddr:                startPara.SyncTxConfig.Chain33Host,
		chainName:               startPara.ChainName,
		chainID:                 startPara.ChainID,
		fetchHeightPeriodMs:     startPara.SyncTxConfig.FetchHeightPeriodMs,
		unlockChan:              make(chan int),
		db:                      startPara.DBHandle,
		ctx:                     startPara.Ctx,
		deployInfo:              startPara.DeployInfo,
		bridgeRegistryAddr:      startPara.BridgeRegistryAddr,
		ethBridgeClaimChan:      startPara.EthBridgeClaimChan,
		chain33MsgChan:          startPara.Chain33MsgChan,
		totalTx4RelayEth2chai33: 0,
		symbol2Addr:             make(map[string]string),
	}

	syncCfg := &ebTypes.SyncTxReceiptConfig{
		Chain33Host:       startPara.SyncTxConfig.Chain33Host,
		PushHost:          startPara.SyncTxConfig.PushHost,
		PushName:          startPara.SyncTxConfig.PushName,
		PushBind:          startPara.SyncTxConfig.PushBind,
		StartSyncHeight:   startPara.SyncTxConfig.StartSyncHeight,
		StartSyncSequence: startPara.SyncTxConfig.StartSyncSequence,
		StartSyncHash:     startPara.SyncTxConfig.StartSyncHash,
	}

	registrAddrInDB, err := chain33Relayer.getBridgeRegistryAddr()
	//如果输入的registry地址非空，且和数据库保存地址不一致，则直接使用输入注册地址
	if chain33Relayer.bridgeRegistryAddr != "" && nil == err && registrAddrInDB != chain33Relayer.bridgeRegistryAddr {
		relayerLog.Error("StartChain33Relayer", "BridgeRegistry is setted already with value", registrAddrInDB,
			"but now setting to", startPara.BridgeRegistryAddr)
		_ = chain33Relayer.setBridgeRegistryAddr(startPara.BridgeRegistryAddr)
	} else if startPara.BridgeRegistryAddr == "" && registrAddrInDB != "" {
		//输入地址为空，且数据库中保存地址不为空，则直接使用数据库中的地址
		chain33Relayer.bridgeRegistryAddr = registrAddrInDB
	}
	chain33Relayer.totalTx4RelayEth2chai33 = chain33Relayer.getTotalTxAmount()
	if 0 == chain33Relayer.totalTx4RelayEth2chai33 {
		statics := &ebTypes.Ethereum2Chain33Statics{}
		data := chain33Types.Encode(statics)
		chain33Relayer.setLastestRelay2Chain33TxStatics(0, int32(events.ClaimTypeLock), data)
		chain33Relayer.setLastestRelay2Chain33TxStatics(0, int32(events.ClaimTypeBurn), data)
	}

	go chain33Relayer.syncProc(syncCfg)
	return chain33Relayer
}

func (chain33Relayer *Relayer4Chain33) syncProc(syncCfg *ebTypes.SyncTxReceiptConfig) {
	_, _ = fmt.Fprintln(os.Stdout, "Pls unlock or import private key for Chain33 relayer")
	<-chain33Relayer.unlockChan
	_, _ = fmt.Fprintln(os.Stdout, "Chain33 relayer starts to run...")
	if err := chain33Relayer.RestoreTokenAddress(); nil != err {
		relayerLog.Info("Failed to RestoreTokenAddress")
		return
	}
	setChainID(chain33Relayer.chainID)
	//如果该中继器的bridgeRegistryAddr为空，就说明合约未部署，需要等待部署成功之后再继续
	if "" == chain33Relayer.bridgeRegistryAddr {
		<-chain33Relayer.unlockChan
	}
	//如果oracleAddr为空，则通过bridgeRegistry合约进行查询
	if "" != chain33Relayer.bridgeRegistryAddr && "" == chain33Relayer.oracleAddr {
		oracleAddr, bridgeBankAddr := recoverContractAddrFromRegistry(chain33Relayer.bridgeRegistryAddr, chain33Relayer.rpcLaddr)
		if "" == oracleAddr || "" == bridgeBankAddr {
			panic("Failed to recoverContractAddrFromRegistry")
		}
		chain33Relayer.oracleAddr = oracleAddr
		chain33Relayer.bridgeBankAddr = bridgeBankAddr
		chain33txLog.Debug("recoverContractAddrFromRegistry", "bridgeRegistryAddr", chain33Relayer.bridgeRegistryAddr,
			"oracleAddr", chain33Relayer.oracleAddr, "bridgeBankAddr", chain33Relayer.bridgeBankAddr)
	}

	syncCfg.Contracts = append(syncCfg.Contracts, chain33Relayer.bridgeBankAddr)
	chain33Relayer.syncEvmTxLogs = syncTx.StartSyncEvmTxLogs(syncCfg, chain33Relayer.db)
	chain33Relayer.lastHeight4Tx = chain33Relayer.loadLastSyncHeight()
	chain33Relayer.mulSignAddr = chain33Relayer.getMultiSignAddress()
	chain33Relayer.prePareSubscribeEvent()
	timer := time.NewTicker(time.Duration(chain33Relayer.fetchHeightPeriodMs) * time.Millisecond)
	for {
		select {
		case <-timer.C:
			height := chain33Relayer.getCurrentHeight()
			relayerLog.Debug("syncProc", "getCurrentHeight", height)
			chain33Relayer.onNewHeightProc(height)

		case <-chain33Relayer.ctx.Done():
			timer.Stop()
			return

		case ethBridgeClaim := <-chain33Relayer.ethBridgeClaimChan:
			chain33Relayer.relayLockBurnToChain33(ethBridgeClaim)
		}
	}
}

func (chain33Relayer *Relayer4Chain33) getCurrentHeight() int64 {
	var res rpctypes.Header
	ctx := jsonclient.NewRPCCtx(chain33Relayer.rpcLaddr, "Chain33.GetLastHeader", nil, &res)
	_, err := ctx.RunResult()
	if nil != err {
		relayerLog.Error("getCurrentHeight", "Failede due to:", err.Error())
	}
	return res.Height
}

func (chain33Relayer *Relayer4Chain33) onNewHeightProc(currentHeight int64) {
	//检查已经提交的交易结果
	chain33Relayer.updateTxStatus()

	//未达到足够的成熟度，不进行处理
	//  +++++++++||++++++++++++||++++++++++||
	//           ^             ^           ^
	// lastHeight4Tx    matDegress   currentHeight
	for chain33Relayer.lastHeight4Tx+int64(chain33Relayer.matDegree)+1 <= currentHeight {
		relayerLog.Info("onNewHeightProc", "currHeight", currentHeight, "lastHeight4Tx", chain33Relayer.lastHeight4Tx)

		lastHeight4Tx := chain33Relayer.lastHeight4Tx
		txLogs, err := chain33Relayer.syncEvmTxLogs.GetNextValidEvmTxLogs(lastHeight4Tx)
		if nil == txLogs || nil != err {
			if err != nil {
				relayerLog.Error("onNewHeightProc", "Failed to GetNextValidTxReceipts due to:", err.Error())
			}
			break
		}
		relayerLog.Debug("onNewHeightProc", "currHeight", currentHeight, "valid tx receipt with height:", txLogs.Height)

		txAndLogs := txLogs.TxAndLogs
		for _, txAndLog := range txAndLogs {
			tx := txAndLog.Tx

			//确认订阅的evm交易类型和合约地址
			if !strings.Contains(string(tx.Execer), "evm") {
				relayerLog.Error("onNewHeightProc received logs not from evm tx", "tx.Execer", string(tx.Execer))
				continue
			}

			var evmAction evmtypes.EVMContractAction
			err := chain33Types.Decode(tx.Payload, &evmAction)
			if nil != err {
				relayerLog.Error("onNewHeightProc", "Failed to decode action for tx with hash", common.ToHex(tx.Hash()))
				continue
			}

			//确认监听的合约地址
			if evmAction.ContractAddr != chain33Relayer.bridgeBankAddr {
				relayerLog.Error("onNewHeightProc received logs not from bridgeBank", "evmAction.ContractAddr", evmAction.ContractAddr)
				continue
			}

			for _, evmlog := range txAndLog.LogsPerTx.Logs {
				var evmEventType events.Chain33EvmEvent
				if chain33Relayer.bridgeBankEventBurnSig == common.ToHex(evmlog.Topic[0]) {
					evmEventType = events.Chain33EventLogBurn
				} else if chain33Relayer.bridgeBankEventLockSig == common.ToHex(evmlog.Topic[0]) {
					evmEventType = events.Chain33EventLogLock
				} else {
					continue
				}

				if err := chain33Relayer.handleBurnLockEvent(evmEventType, evmlog.Data, tx.Hash()); nil != err {
					errInfo := fmt.Sprintf("Failed to handleBurnLockEvent due to:%s", err.Error())
					panic(errInfo)
				}
			}
		}
		chain33Relayer.lastHeight4Tx = txLogs.Height
		chain33Relayer.setLastSyncHeight(chain33Relayer.lastHeight4Tx)
	}
}

// handleBurnLockMsg : parse event data as a Chain33Msg, package it into a ProphecyClaim, then relay tx to the Ethereum Network
func (chain33Relayer *Relayer4Chain33) handleBurnLockEvent(evmEventType events.Chain33EvmEvent, data []byte, chain33TxHash []byte) error {
	relayerLog.Info("handleBurnLockEvent", "Received tx with hash", ethCommon.Bytes2Hex(chain33TxHash))

	// Parse the witnessed event's data into a new Chain33Msg
	chain33Msg, err := events.ParseBurnLock4chain33(evmEventType, data, chain33Relayer.bridgeBankAbi, chain33TxHash)
	if nil != err {
		return err
	}

	chain33Relayer.chain33MsgChan <- chain33Msg

	return nil
}

//DeployContrcts 部署以太坊合约
func (chain33Relayer *Relayer4Chain33) DeployContracts() (bridgeRegistry string, err error) {
	bridgeRegistry = ""
	if nil == chain33Relayer.deployInfo {
		return bridgeRegistry, errors.New("no deploy info configured yet")
	}
	if len(chain33Relayer.deployInfo.ValidatorsAddr) != len(chain33Relayer.deployInfo.InitPowers) {
		return bridgeRegistry, errors.New("not same number for validator address and power")
	}
	if len(chain33Relayer.deployInfo.ValidatorsAddr) < 3 {
		return bridgeRegistry, errors.New("the number of validator must be not less than 3")
	}

	//已经设置了注册合约地址，说明已经部署了相关的合约，不再重复部署
	if chain33Relayer.bridgeRegistryAddr != "" {
		return bridgeRegistry, errors.New("contract deployed already")
	}

	var validators []address.Address
	var initPowers []*big.Int

	for i, addrStr := range chain33Relayer.deployInfo.ValidatorsAddr {
		addr, err := address.NewAddrFromString(addrStr)
		if nil != err {
			panic(fmt.Sprintf("Failed to NewAddrFromString for:%s", addrStr))
		}
		validators = append(validators, *addr)
		initPowers = append(initPowers, big.NewInt(chain33Relayer.deployInfo.InitPowers[i]))
	}
	deployerAddr, err := address.NewAddrFromString(chain33Relayer.deployInfo.OperatorAddr)
	if nil != err {
		panic(fmt.Sprintf("Failed to NewAddrFromString for:%s", chain33Relayer.deployInfo.OperatorAddr))
	}
	para4deploy := &DeployPara4Chain33{
		Deployer:       *deployerAddr,
		Operator:       *deployerAddr,
		InitValidators: validators,
		InitPowers:     initPowers,
	}

	for i, power := range para4deploy.InitPowers {
		relayerLog.Info("deploy", "the validator address ", para4deploy.InitValidators[i].String(),
			"power", power.String())
	}

	x2EthDeployInfo, err := deployAndInit2Chain33(chain33Relayer.rpcLaddr, chain33Relayer.chainName, para4deploy)
	if err != nil {
		return bridgeRegistry, err
	}
	chain33Relayer.rwLock.Lock()

	chain33Relayer.deployResult = x2EthDeployInfo
	bridgeRegistry = x2EthDeployInfo.BridgeRegistry.Address.String()
	_ = chain33Relayer.setBridgeRegistryAddr(bridgeRegistry)
	//设置注册合约地址，同时设置启动中继服务的信号
	chain33Relayer.bridgeRegistryAddr = bridgeRegistry
	chain33Relayer.oracleAddr = x2EthDeployInfo.Oracle.Address.String()
	chain33Relayer.bridgeBankAddr = x2EthDeployInfo.BridgeBank.Address.String()
	chain33Relayer.rwLock.Unlock()
	chain33Relayer.unlockChan <- start
	relayerLog.Info("deploy", "the BridgeRegistry address is", bridgeRegistry)

	return bridgeRegistry, nil
}

//DeployContrcts 部署以太坊合约
func (chain33Relayer *Relayer4Chain33) DeployMulsign() (mulsign string, err error) {
	mulsign, err = deployMulSign2Chain33(chain33Relayer.rpcLaddr, chain33Relayer.chainName, chain33Relayer.deployInfo.OperatorAddr)
	if err != nil {
		return "", err
	}
	chain33Relayer.rwLock.Lock()
	chain33Relayer.mulSignAddr = mulsign
	chain33Relayer.rwLock.Unlock()

	chain33Relayer.setMultiSignAddress(mulsign)

	return mulsign, nil
}

func (chain33Relayer *Relayer4Chain33) CreateERC20ToChain33(param ebTypes.ERC20Token) (erc20 string, err error) {
	erc20, err = deployERC20ToChain33(chain33Relayer.rpcLaddr, chain33Relayer.chainName, chain33Relayer.deployInfo.OperatorAddr, param)
	if err != nil {
		return "", err
	}

	return erc20, nil
}

func (chain33Relayer *Relayer4Chain33) relayLockBurnToChain33(claim *ebTypes.EthBridgeClaim) {
	relayerLog.Debug("relayLockBurnToChain33", "new EthBridgeClaim received", claim)

	nonceBytes := big.NewInt(claim.Nonce).Bytes()
	bigAmount := big.NewInt(0)
	bigAmount.SetString(claim.Amount, 10)
	amountBytes := bigAmount.Bytes()
	claimID := crypto.Keccak256Hash(nonceBytes, []byte(claim.EthereumSender), []byte(claim.Chain33Receiver), []byte(claim.Symbol), amountBytes)

	// Sign the hash using the active validator's private key
	signature, err := utils.SignClaim4Evm(claimID, chain33Relayer.privateKey4Chain33_ecdsa)
	if nil != err {
		panic("SignClaim4Evm due to" + err.Error())
	}

	var tokenAddr string
	operationType := events.ClaimType(claim.ClaimType).String()
	if int32(events.ClaimTypeBurn) == claim.ClaimType {
		//burn 分支
		if ebTypes.SYMBOL_BTY == claim.Symbol {
			tokenAddr = ebTypes.BTYAddrChain33
		} else {
			tokenAddr = getLockedTokenAddress(chain33Relayer.bridgeBankAddr, claim.Symbol, chain33Relayer.rpcLaddr)
			if "" == tokenAddr {
				relayerLog.Error("relayLockBurnToChain33", "No locked token address created for symbol", claim.Symbol)
				return
			}
		}
	} else {
		//lock 分支
		var exist bool
		tokenAddr, exist = chain33Relayer.symbol2Addr[claim.Symbol]
		if !exist {
			tokenAddr = getBridgeToken2address(chain33Relayer.bridgeBankAddr, claim.Symbol, chain33Relayer.rpcLaddr)
			if "" == tokenAddr {
				relayerLog.Error("relayLockBurnToChain33", "No bridge token address created for symbol", claim.Symbol)
				return
			}
			relayerLog.Info("relayLockBurnToChain33", "Succeed to get bridge token address for symbol", claim.Symbol,
				"address", tokenAddr)

			token2set := ebTypes.TokenAddress{
				Address:   tokenAddr,
				Symbol:    claim.Symbol,
				ChainName: ebTypes.Chain33BlockChainName,
			}
			if err := chain33Relayer.SetTokenAddress(token2set); nil != err {
				relayerLog.Info("relayLockBurnToChain33", "Failed to SetTokenAddress due to", err.Error())
			}
		}
	}

	//因为发行的合约的精度为8，所以需要缩小，在进行burn的时候，再进行倍乘,在函数ParseBurnLock4chain33进行
	if ebTypes.SYMBOL_ETH == claim.Symbol {
		bigAmount.Div(bigAmount, big.NewInt(int64(1e10)))
		claim.Amount = bigAmount.String()
	}

	parameter := fmt.Sprintf("newOracleClaim(%d, %s, %s, %s, %s, %s, %s, %s)",
		claim.ClaimType,
		claim.EthereumSender,
		claim.Chain33Receiver,
		tokenAddr,
		claim.Symbol,
		claim.Amount,
		claimID.String(),
		common.ToHex(signature))
	relayerLog.Info("relayLockBurnToChain33", "parameter", parameter)

	claim.ChainName = chain33Relayer.chainName
	txhash, err := relayEvmTx2Chain33(chain33Relayer.privateKey4Chain33, claim, parameter, chain33Relayer.rpcLaddr, chain33Relayer.oracleAddr)
	if err != nil {
		relayerLog.Error("relayLockBurnToChain33", "Failed to RelayEvmTx2Chain33 due to:", err.Error(), "EthereumTxhash", claim.EthTxHash)
		return
	}

	//第一个有效的index从１开始，方便list
	txIndex := atomic.AddInt64(&chain33Relayer.totalTx4RelayEth2chai33, 1)
	if err = chain33Relayer.updateTotalTxAmount2Eth(txIndex); nil != err {
		relayerLog.Error("relayLockBurnToChain33", "Failed to updateTotalTxAmount2Eth due to:", err.Error())
		return
	}

	statics := &ebTypes.Ethereum2Chain33Statics{
		Chain33Txstatus: ebTypes.Tx_Status_Pending,
		Chain33Txhash:   txhash,
		EthereumTxhash:  claim.EthTxHash,
		BurnLock:        claim.ClaimType,
		EthereumSender:  claim.EthereumSender,
		Chain33Receiver: claim.Chain33Receiver,
		Symbol:          claim.Symbol,
		Amount:          claim.Amount,
		Nonce:           claim.Nonce,
		TxIndex:         txIndex,
		OperationType:   operationType,
	}
	data := chain33Types.Encode(statics)
	if err = chain33Relayer.setLastestRelay2Chain33TxStatics(txIndex, claim.ClaimType, data); nil != err {
		relayerLog.Error("relayLockBurnToChain33", "Failed to setLastestRelay2Chain33TxStatics due to:", err.Error())
		return
	}
	relayerLog.Info("relayLockBurnToChain33::successful",
		"txIndex", txIndex,
		"Chain33Txhash", txhash,
		"EthereumTxhash", claim.EthTxHash,
		"type", operationType,
		"Symbol", claim.Symbol,
		"Amount", claim.Amount,
		"EthereumSender", claim.EthereumSender,
		"Chain33Receiver", claim.Chain33Receiver)
}

func (chain33Relayer *Relayer4Chain33) BurnAsyncFromChain33(ownerPrivateKey, tokenAddr, ethereumReceiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return burnAsync(ownerPrivateKey, tokenAddr, ethereumReceiver, bn.Int64(), chain33Relayer.bridgeBankAddr, chain33Relayer.chainName, chain33Relayer.rpcLaddr)
}

func (chain33Relayer *Relayer4Chain33) LockBTYAssetAsync(ownerPrivateKey, ethereumReceiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return lockAsync(ownerPrivateKey, ethereumReceiver, bn.Int64(), chain33Relayer.bridgeBankAddr, chain33Relayer.chainName, chain33Relayer.rpcLaddr)
}

//ShowBridgeRegistryAddr ...
func (chain33Relayer *Relayer4Chain33) ShowBridgeRegistryAddr() (string, error) {
	if "" == chain33Relayer.bridgeRegistryAddr {
		return "", errors.New("the relayer is not started yet")
	}

	return chain33Relayer.bridgeRegistryAddr, nil
}

func (chain33Relayer *Relayer4Chain33) ShowStatics(request *ebTypes.TokenStaticsRequest) (*ebTypes.TokenStaticsResponse, error) {
	res := &ebTypes.TokenStaticsResponse{}

	datas, err := chain33Relayer.getStatics(request.Operation, request.TxIndex, request.Count)
	if nil != err {
		return nil, err
	}
	//todo:完善分页显示功能
	for _, data := range datas {
		var statics ebTypes.Ethereum2Chain33Statics
		_ = chain33Types.Decode(data, &statics)
		if request.Status != 0 && ebTypes.Tx_Status_Map[request.Status] != statics.Chain33Txstatus {
			continue
		}
		if len(request.Symbol) > 0 && request.Symbol != statics.Symbol {
			continue
		}
		res.E2Cstatics = append(res.E2Cstatics, &statics)
	}
	return res, nil
}

func (chain33Relayer *Relayer4Chain33) updateTxStatus() {
	chain33Relayer.updateSingleTxStatus(events.ClaimTypeBurn)
	chain33Relayer.updateSingleTxStatus(events.ClaimTypeLock)
}

func (chain33Relayer *Relayer4Chain33) updateSingleTxStatus(claimType events.ClaimType) {
	txIndex := chain33Relayer.getChain33UpdateTxIndex(claimType)
	datas, _ := chain33Relayer.getStatics(int32(claimType), txIndex, 0)
	if nil == datas {
		return
	}
	for _, data := range datas {
		var statics ebTypes.Ethereum2Chain33Statics
		_ = chain33Types.Decode(data, &statics)
		result := getTxStatusByHashesRpc(statics.Chain33Txhash, chain33Relayer.rpcLaddr)
		//当前处理机制比较简单，如果发现该笔交易未执行，就不再产寻后续交易的回执
		if ebTypes.Invalid_Chain33Tx_Status == result {
			relayerLog.Debug("chain33Relayer::updateSingleTxStatus", "no receipt for tx index", statics.TxIndex)
			break
		}
		status := ebTypes.Tx_Status_Success
		if result != chain33Types.ExecOk {
			status = ebTypes.Tx_Status_Failed
		}
		statics.Chain33Txstatus = status
		dataNew := chain33Types.Encode(&statics)
		_ = chain33Relayer.setLastestRelay2Chain33TxStatics(statics.TxIndex, int32(claimType), dataNew)
		_ = chain33Relayer.setChain33UpdateTxIndex(statics.TxIndex, claimType)
		relayerLog.Debug("updateSingleTxStatus", "TxIndex", statics.TxIndex, "operationType", statics.OperationType, "txHash", statics.Chain33Txhash, "updated status", status)
	}
}

func (chain33Relayer *Relayer4Chain33) SetupMulSign(setupMulSign ebTypes.SetupMulSign) (string, error) {
	if "" == chain33Relayer.mulSignAddr {
		return "", ebTypes.ErrMulSignNotDeployed
	}

	return setupMultiSign(setupMulSign.OperatorPrivateKey, chain33Relayer.mulSignAddr, chain33Relayer.chainName, chain33Relayer.rpcLaddr, setupMulSign.Owners)
}

func (chain33Relayer *Relayer4Chain33) SafeTransfer(para ebTypes.SafeTransfer) (string, error) {
	if "" == chain33Relayer.mulSignAddr {
		return "", ebTypes.ErrMulSignNotDeployed
	}

	return safeTransfer(para.OwnerPrivateKeys[0], chain33Relayer.mulSignAddr, chain33Relayer.chainName,
		chain33Relayer.rpcLaddr, para.To, para.Token, para.OwnerPrivateKeys, para.Amount)
}

func (chain33Relayer *Relayer4Chain33) SetMultiSignAddr(address string) {
	chain33Relayer.rwLock.Lock()
	chain33Relayer.mulSignAddr = address
	chain33Relayer.rwLock.Unlock()

	chain33Relayer.setMultiSignAddress(address)
}
