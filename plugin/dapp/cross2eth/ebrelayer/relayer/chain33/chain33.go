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

	chain33EvmCommon "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"

	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"

	"github.com/33cn/chain33/common"
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
	"github.com/ethereum/go-ethereum/crypto"
)

var relayerLog = log.New("module", "chain33_relayer")

//Relayer4Chain33 ...
type Relayer4Chain33 struct {
	syncEvmTxLogs       *syncTx.EVMTxLogs
	rpcLaddr            string //用户向指定的blockchain节点进行rpc调用
	chain33RpcUrls      []string
	chainName           string //用来区别主链中继还是平行链，主链为空，平行链则是user.p.xxx.
	chainID             int32
	fetchHeightPeriodMs int64
	db                  dbm.DB
	lastHeight4Tx       int64 //等待被处理的具有相应的交易回执的高度
	matDegree           int32 //成熟度         heightSync2App    matDegress   height

	privateKey4Chain33         chain33Crypto.PrivKey
	privateKey4Chain33_ecdsa   *ecdsa.PrivateKey
	ctx                        context.Context
	rwLock                     sync.RWMutex
	unlockChan                 chan int
	bridgeBankEventLockSig     string
	bridgeBankEventBurnSig     string
	bridgeBankEventWithdrawSig string
	bridgeBankAbi              abi.ABI
	totalTx4RelayEth2chai33    int64
	//新增//
	ethBridgeClaimChan        <-chan *ebTypes.EthBridgeClaim
	txRelayAckRecvChan        <-chan *ebTypes.TxRelayAck
	txRelayAckSendChan        map[string]chan<- *ebTypes.TxRelayAck
	chain33MsgChan            map[string]chan<- *events.Chain33Msg
	bridgeRegistryAddr        string
	oracleAddr                string
	bridgeBankAddr            string
	mulSignAddr               string
	deployResult              *X2EthDeployResult
	symbol2Addr               map[string]string
	bridgeSymbol2EthChainName map[string]string //在chain33上发行的跨链token的名称到以太坊链的名称映射
	processWithDraw           bool
	delayedSend               bool
	delayedSendTime           int64
}

type Chain33StartPara struct {
	ChainName          string
	Ctx                context.Context
	SyncTxConfig       *ebTypes.SyncTxConfig
	BridgeRegistryAddr string
	DBHandle           dbm.DB
	EthBridgeClaimChan <-chan *ebTypes.EthBridgeClaim
	TxRelayAckRecvChan <-chan *ebTypes.TxRelayAck
	TxRelayAckSendChan map[string]chan<- *ebTypes.TxRelayAck
	Chain33MsgChan     map[string]chan<- *events.Chain33Msg
	ChainID            int32
	ProcessWithDraw    bool
	DelayedSend        bool
	DelayedSendTime    int64
}

// StartChain33Relayer : initializes a relayer which witnesses events on the chain33 network and relays them to Ethereum
func StartChain33Relayer(startPara *Chain33StartPara) *Relayer4Chain33 {
	chain33Relayer := &Relayer4Chain33{
		rpcLaddr:                startPara.SyncTxConfig.Chain33Host,
		chain33RpcUrls:          startPara.SyncTxConfig.Chain33RpcUrls,
		chainName:               startPara.ChainName,
		chainID:                 startPara.ChainID,
		fetchHeightPeriodMs:     startPara.SyncTxConfig.FetchHeightPeriodMs,
		unlockChan:              make(chan int),
		db:                      startPara.DBHandle,
		ctx:                     startPara.Ctx,
		bridgeRegistryAddr:      startPara.BridgeRegistryAddr,
		ethBridgeClaimChan:      startPara.EthBridgeClaimChan,
		txRelayAckRecvChan:      startPara.TxRelayAckRecvChan,
		txRelayAckSendChan:      startPara.TxRelayAckSendChan,
		chain33MsgChan:          startPara.Chain33MsgChan,
		totalTx4RelayEth2chai33: 0,
		symbol2Addr:             make(map[string]string),
		processWithDraw:         startPara.ProcessWithDraw,
		delayedSend:             startPara.DelayedSend,
		delayedSendTime:         startPara.DelayedSendTime,
	}

	syncCfg := &ebTypes.SyncTxReceiptConfig{
		Chain33Host:       startPara.SyncTxConfig.Chain33Host,
		PushHost:          startPara.SyncTxConfig.PushHost,
		PushName:          startPara.SyncTxConfig.PushName,
		PushBind:          startPara.SyncTxConfig.PushBind,
		StartSyncHeight:   startPara.SyncTxConfig.StartSyncHeight,
		StartSyncSequence: startPara.SyncTxConfig.StartSyncSequence,
		StartSyncHash:     startPara.SyncTxConfig.StartSyncHash,
		KeepAliveDuration: startPara.SyncTxConfig.KeepAliveDuration,
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
		err := chain33Relayer.setLastestRelay2Chain33TxStatics(0, int32(events.ClaimTypeLock), data)
		if err != nil {
			relayerLog.Error("StartChain33Relayer", "setLastestRelay2Chain33TxStatics ClaimTypeLock error", err.Error())
		}
		err = chain33Relayer.setLastestRelay2Chain33TxStatics(0, int32(events.ClaimTypeBurn), data)
		if err != nil {
			relayerLog.Error("StartChain33Relayer", "setLastestRelay2Chain33TxStatics ClaimTypeBurn error", err.Error())
		}
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
		chain33txLog.Debug("bridgeRegistryAddr empty")
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
	chain33Relayer.bridgeSymbol2EthChainName = chain33Relayer.restoreSymbol2chainName()
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

		case txRelayAck := <-chain33Relayer.txRelayAckRecvChan:
			chain33Relayer.procTxRelayAck(txRelayAck)
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
	chain33Relayer.checkTxRelay2Ethereum()

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
				} else if chain33Relayer.bridgeBankEventWithdrawSig == common.ToHex(evmlog.Topic[0]) {
					evmEventType = events.Chain33EventLogWithdraw
				} else {
					continue
				}

				if evmEventType == events.Chain33EventLogWithdraw && !chain33Relayer.processWithDraw {
					//代理提币消息只由代理提币节点处理
					continue
				}
				if evmEventType != events.Chain33EventLogWithdraw && chain33Relayer.processWithDraw {
					//lock和burn消息消息只由普通中继节点处理
					continue
				}

				if err := chain33Relayer.handleBurnLockWithdrawEvent(evmEventType, evmlog.Data, tx.Hash()); nil != err {
					relayerLog.Error("onNewHeightProc", "Failed to handleBurnLockWithdrawEvent due to:%s", err.Error())
				}

			}
		}
		chain33Relayer.lastHeight4Tx = txLogs.Height
		chain33Relayer.setLastSyncHeight(chain33Relayer.lastHeight4Tx)
	}
}

// handleBurnLockMsg : parse event data as a Chain33Msg, package it into a ProphecyClaim, then relay tx to the Ethereum Network
func (chain33Relayer *Relayer4Chain33) handleBurnLockWithdrawEvent(evmEventType events.Chain33EvmEvent, data []byte, chain33TxHash []byte) error {
	txHashStr := common.ToHex(chain33TxHash)
	relayerLog.Info("handleBurnLockWithdrawEvent", "Received tx with hash", txHashStr)

	// 删除已发送校验, 如果ethereum端发生交易后没有打包, 可重新再发生
	//if chain33Relayer.checkTxProcessed(txHashStr) {
	//	relayerLog.Info("handleBurnLockWithdrawEvent", "Tx has been already Processed with hash:", txHashStr)
	//	return nil
	//}

	// Parse the witnessed event's data into a new Chain33Msg
	chain33Msg, err := events.ParseBurnLock4chain33(evmEventType, data, chain33Relayer.bridgeBankAbi, chain33TxHash)
	if nil != err {
		return err
	}
	fdIndex := chain33Relayer.getFdTx2EthTotalAmount() + 1
	chain33Msg.ForwardTimes = 1
	chain33Msg.ForwardIndex = fdIndex

	relayerLog.Info("handleBurnLockWithdrawEvent", "Going to send chain33Msg.ClaimType", chain33Msg.ClaimType.String())

	var chainName string
	//specical process: withdraw YCC　only to bsc
	if events.Chain33EventLogWithdraw == evmEventType && "YCC" == chain33Msg.Symbol {
		chainName = ebTypes.BinanceChainName
	} else {
		ok := false
		chainName, ok = chain33Relayer.bridgeSymbol2EthChainName[chain33Msg.Symbol]
		if !ok {
			relayerLog.Error("handleBurnLockWithdrawEvent", "No bridgeSymbol2EthChainName", chain33Msg.Symbol)
			return errors.New("ErrNoEthChainName4BridgeSymbol")
		}
	}

	channel, ok := chain33Relayer.chain33MsgChan[chainName]
	if !ok {
		relayerLog.Error("handleBurnLockWithdrawEvent", "No bridgeSymbol2EthChainName", chainName)
		return errors.New("ErrNoChain33MsgChan4EthChainName")
	}

	_ = chain33Relayer.updateFdTx2EthTotalAmount(fdIndex)
	txRelayConfirm4Chain33 := &ebTypes.TxRelayConfirm4Chain33{
		EventType:   int32(evmEventType),
		Data:        data,
		FdTimes:     1,
		FdIndex:     fdIndex,
		ToChainName: chainName,
		TxHash:      chain33TxHash,
		Resend:      false,
	}

	if chain33Relayer.delayedSend {
		go chain33Relayer.delayedSendTxs(chainName, chain33Msg, chain33TxHash, txRelayConfirm4Chain33)
	} else {
		channel <- chain33Msg
		//relaychain33ToEthereumCheckPonit 1:send chain33Msg to ethereum relay service
		relayerLog.Info("handleBurnLockWithdrawEvent::relaychain33ToEthereumCheckPonit_1", "chain33TxHash", txHashStr, "ForwardIndex", chain33Msg.ForwardIndex, "FdTimes", 1)
		err = chain33Relayer.setChain33TxIsRelayedUnconfirm(txHashStr, fdIndex, txRelayConfirm4Chain33)
	}

	return err
}

func (chain33Relayer *Relayer4Chain33) delayedSendTxs(chainName string, chain33Msg *events.Chain33Msg, chain33TxHash []byte, txRelayConfirm4Chain33 *ebTypes.TxRelayConfirm4Chain33) {
	delayedSendTime := time.Duration(chain33Relayer.delayedSendTime) * time.Millisecond
	relayerLog.Debug("delayedSendTxs", "setEthTxWaitingForSend chain33TxHash", common.ToHex(chain33TxHash))
	time.Sleep(delayedSendTime)
	channel, ok := chain33Relayer.chain33MsgChan[chainName]
	if !ok {
		relayerLog.Error("handleBurnLockWithdrawEvent", "No bridgeSymbol2EthChainName", chainName)
		return
	}

	channel <- chain33Msg

	//relaychain33ToEthereumCheckPonit 1:send chain33Msg to ethereum relay service
	relayerLog.Info("handleBurnLockWithdrawEvent::relaychain33ToEthereumCheckPonit_1", "chain33TxHash", common.ToHex(chain33TxHash), "ForwardIndex", chain33Msg.ForwardIndex, "FdTimes", 1)
	_ = chain33Relayer.setChain33TxIsRelayedUnconfirm(common.ToHex(chain33TxHash), txRelayConfirm4Chain33.FdIndex, txRelayConfirm4Chain33)
}

func (chain33Relayer *Relayer4Chain33) ResendChain33Event(height int64) (err error) {
	txLogs, err := chain33Relayer.syncEvmTxLogs.GetNextValidEvmTxLogs(height)
	if nil == txLogs || nil != err {
		if err != nil {
			relayerLog.Error("ResendChain33Event", "Failed to GetNextValidTxReceipts due to:", err.Error())
			return err
		}
		return nil
	}
	relayerLog.Debug("ResendChain33Event", "lastHeight4Tx", chain33Relayer.lastHeight4Tx, "valid tx receipt with height:", txLogs.Height)

	txAndLogs := txLogs.TxAndLogs
	for _, txAndLog := range txAndLogs {
		tx := txAndLog.Tx

		//确认订阅的evm交易类型和合约地址
		if !strings.Contains(string(tx.Execer), "evm") {
			relayerLog.Error("ResendChain33Event received logs not from evm tx", "tx.Execer", string(tx.Execer))
			continue
		}

		var evmAction evmtypes.EVMContractAction
		err := chain33Types.Decode(tx.Payload, &evmAction)
		if nil != err {
			relayerLog.Error("ResendChain33Event", "Failed to decode action for tx with hash", common.ToHex(tx.Hash()))
			continue
		}

		//确认监听的合约地址
		if evmAction.ContractAddr != chain33Relayer.bridgeBankAddr {
			relayerLog.Error("ResendChain33Event received logs not from bridgeBank", "evmAction.ContractAddr", evmAction.ContractAddr)
			continue
		}

		for _, evmlog := range txAndLog.LogsPerTx.Logs {
			var evmEventType events.Chain33EvmEvent
			if chain33Relayer.bridgeBankEventBurnSig == common.ToHex(evmlog.Topic[0]) {
				evmEventType = events.Chain33EventLogBurn
			} else if chain33Relayer.bridgeBankEventLockSig == common.ToHex(evmlog.Topic[0]) {
				evmEventType = events.Chain33EventLogLock
			} else if chain33Relayer.bridgeBankEventWithdrawSig == common.ToHex(evmlog.Topic[0]) {
				evmEventType = events.Chain33EventLogWithdraw
			} else {
				continue
			}

			if evmEventType == events.Chain33EventLogWithdraw && !chain33Relayer.processWithDraw {
				//代理提币消息只由代理提币节点处理
				continue
			}
			if evmEventType != events.Chain33EventLogWithdraw && chain33Relayer.processWithDraw {
				//lock和burn消息消息只由普通中继节点处理
				continue
			}

			if err := chain33Relayer.handleBurnLockWithdrawEvent(evmEventType, evmlog.Data, tx.Hash()); nil != err {
				return err
			}
		}
	}

	return nil
}

func (chain33Relayer *Relayer4Chain33) checkIsResendEthClaim(claim *ebTypes.EthBridgeClaim) bool {
	if claim.ForwardTimes <= 1 {
		return false
	}
	ethTxHash := claim.EthTxHash
	relayerLog.Info("checkIsResendEthClaim", "Received the same EthBridgeClaim more than once with times", claim.ForwardTimes, "tx hash string", ethTxHash)
	relayTxDetail, _ := chain33Relayer.getEthTxRelayAlreadyInfo(ethTxHash)
	if nil == relayTxDetail {
		relayerLog.Info("checkIsResendEthClaim::haven't relay yet")
		return false
	}

	//if relay already, just ack it
	chain33Relayer.txRelayAckSendChan[claim.ChainName] <- &ebTypes.TxRelayAck{
		TxHash:  ethTxHash,
		FdIndex: claim.ForwardIndex,
	}
	relayerLog.Info("checkIsResendEthClaim", "have relay already with tx hash:", relayTxDetail.Txhash)
	return true
}

func (chain33Relayer *Relayer4Chain33) relayLockBurnToChain33(claim *ebTypes.EthBridgeClaim) {
	relayerLog.Debug("relayLockBurnToChain33", "new EthBridgeClaim received", claim)
	if chain33Relayer.checkIsResendEthClaim(claim) {
		return
	}

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
		if _, ok := chain33Relayer.bridgeSymbol2EthChainName[claim.Symbol]; !ok {
			chain33Relayer.bridgeSymbol2EthChainName[claim.Symbol] = claim.ChainName
			chain33Relayer.storeSymbol2chainName(chain33Relayer.bridgeSymbol2EthChainName)
		}
		//如果是代理打币节点，则只收集symbol和chain name相关信息
		if chain33Relayer.processWithDraw {
			return
		}

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

			token2set := &ebTypes.TokenAddress{
				Address:   tokenAddr,
				Symbol:    claim.Symbol,
				ChainName: ebTypes.Chain33BlockChainName,
			}
			if err := chain33Relayer.SetTokenAddress(token2set); nil != err {
				relayerLog.Info("relayLockBurnToChain33", "Failed to SetTokenAddress due to", err.Error())
			}
		}
	}

	//因为发行的合约的精度为8，所以需要进行相应的缩放
	if 8 != claim.Decimal {
		if claim.Decimal > 8 {
			dist := claim.Decimal - 8
			value, exist := utils.Decimal2value[int(dist)]
			if !exist {
				panic(fmt.Sprintf("does support for decimal, %d", claim.Decimal))
			}
			bigAmount.Div(bigAmount, big.NewInt(value))
			claim.Amount = bigAmount.String()
		} else {
			dist := 8 - claim.Decimal
			value, exist := utils.Decimal2value[int(dist)]
			if !exist {
				panic(fmt.Sprintf("does support for decimal, %d", claim.Decimal))
			}
			bigAmount.Mul(bigAmount, big.NewInt(value))
			claim.Amount = bigAmount.String()
		}
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

	txhash, err := relayEvmTx2Chain33(chain33Relayer.privateKey4Chain33, claim, parameter, chain33Relayer.oracleAddr, chain33Relayer.chainName, chain33Relayer.chain33RpcUrls)
	if err != nil {
		relayerLog.Error("relayLockBurnToChain33", "Failed to RelayEvmTx2Chain33 due to:", err.Error(), "EthereumTxhash", claim.EthTxHash)
		return
	}

	chain33Relayer.txRelayAckSendChan[claim.ChainName] <- &ebTypes.TxRelayAck{
		TxHash:  claim.EthTxHash,
		FdIndex: claim.ForwardIndex,
	}
	//relayEthereum2chain33CheckPonit 2:send ack
	relayerLog.Info("relayLockBurnToChain33::relayEthereum2chain33CheckPonit_2::sendAck", "ethTxhash", claim.EthTxHash, "ForwardIndex", claim.ForwardIndex, "FdTimes", claim.ForwardTimes)

	relayTxDetail := &ebTypes.RelayTxDetail{
		ClaimType:      claim.ClaimType,
		TxIndexRelayed: claim.ForwardIndex,
		Txhash:         txhash,
	}

	//set flag to indicate that the eth tx has been relayed to chain33
	if err = chain33Relayer.setEthTxRelayAlreadyInfo(claim.EthTxHash, relayTxDetail); nil != err {
		relayerLog.Error("relayLockBurnToChain33", "Failed to setTxRelayAlreadyInfo due to:", err.Error())
		return
	}
	//relayEthereum2chain33CheckPonit 3:setFalgRelayFinish
	relayerLog.Info("relayLockBurnToChain33::relayEthereum2chain33CheckPonit_3::setFalgRelayFinish", "ethTxhash", claim.EthTxHash, "ForwardIndex", claim.ForwardIndex, "FdTimes", claim.ForwardTimes)

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

// 该函数用于定期检查是否有需要重新发送给以太坊协成的chain33事件信息,用于产生relay event
func (chain33Relayer *Relayer4Chain33) checkTxRelay2Ethereum() {
	txInfos, err := chain33Relayer.getAllTxsUnconfirm()
	if err != nil {
		relayerLog.Error("chain33Relayer::checkTxRelay2Ethereum", "Failed to getAllTxsUnconfirm due to", err.Error())
		return
	}
	if 0 == len(txInfos) {
		return
	}
	for _, txInfo := range txInfos {
		txHashStr := chain33EvmCommon.Bytes2Hex(txInfo.TxHash)

		if !txInfo.Resend {
			//为了防止转发出去的消息之后，下一个区块时间马上到来，首次转发的消息需要至少等一个区块间隔之后才会进行转发
			txInfo.Resend = true
			err = chain33Relayer.setChain33TxIsRelayedUnconfirm(txHashStr, txInfo.FdIndex, txInfo)
			if nil != err {
				relayerLog.Error("chain33Relayer::checkTxRelay2Ethereum", "Failed to SetTxIsRelayedconfirm due to", err.Error())
				return
			}
			continue
		}

		chain33Msg, err := events.ParseBurnLock4chain33(events.Chain33EvmEvent(txInfo.EventType), txInfo.Data, chain33Relayer.bridgeBankAbi, txInfo.TxHash)
		if nil != err {
			relayerLog.Error("chain33Relayer::checkTxRelay2Ethereum", "Failed to ParseBurnLock4chain33 due to", err.Error())
			return
		}
		txInfo.FdTimes = txInfo.FdTimes + 1
		chain33Msg.ForwardTimes = txInfo.FdTimes
		chain33Msg.ForwardIndex = txInfo.FdIndex

		channel, ok := chain33Relayer.chain33MsgChan[txInfo.ToChainName]
		if !ok {
			relayerLog.Error("chain33Relayer::checkTxRelay2Ethereum", "No chain33MsgChan for ethereum chain with name", txInfo.ToChainName)
			return
		}
		channel <- chain33Msg

		//relaychain33ToEthereumCheckPonit 5: checkTxRelay2Ethereum
		relayerLog.Info("chain33Relayer::relaychain33ToEthereumCheckPonit_5::checkTxRelay2Ethereum", "chain33TxHash", txHashStr, "ForwardIndex", chain33Msg.ForwardIndex, "FdTimes", chain33Msg.ForwardTimes)
		err = chain33Relayer.setChain33TxIsRelayedUnconfirm(txHashStr, txInfo.FdIndex, txInfo)
		if nil != err {
			relayerLog.Error("chain33Relayer::checkTxRelay2Ethereum", "Failed to SetTxIsRelayedconfirm due to", err.Error())
			return
		}
	}
}

//用于chain33的事件信息被中继之后的ack信息，重置标志位
func (chain33Relayer *Relayer4Chain33) procTxRelayAck(ack *ebTypes.TxRelayAck) {
	//reset with another key to exclude from the check list to resend the same message
	if err := chain33Relayer.resetKeyChain33TxRelayedAlready(ack.TxHash); nil != err {
		relayerLog.Error("chain33Relayer::procTxRelayAck", "Failed to resetKeyTxRelayedAlready due to:", err.Error())
		return
	}
	//relaychain33ToEthereumCheckPonit 4: recv ack from ethereum relay service
	relayerLog.Info("chain33Relayer::procTxRelayAck::relaychain33ToEthereumCheckPonit_4", "chain33TxHash", ack.TxHash, "ForwardIndex", ack.FdIndex)
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
		result := GetTxStatusByHashesRpc(statics.Chain33Txhash, chain33Relayer.rpcLaddr)
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

func (chain33Relayer *Relayer4Chain33) SetupMulSign(setupMulSign *ebTypes.SetupMulSign) (string, error) {
	if "" == chain33Relayer.mulSignAddr {
		return "", ebTypes.ErrMulSignNotDeployed
	}

	return setupMultiSign(setupMulSign.OperatorPrivateKey, chain33Relayer.mulSignAddr, chain33Relayer.chainName, chain33Relayer.rpcLaddr, setupMulSign.Owners)
}

func (chain33Relayer *Relayer4Chain33) SafeTransfer(para *ebTypes.SafeTransfer) (string, error) {
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

func (chain33Relayer *Relayer4Chain33) GetMultiSignAddr() string {
	return chain33Relayer.getMultiSignAddress()
}

func (chain33Relayer *Relayer4Chain33) WithdrawFromChain33(ownerPrivateKey, tokenAddr, ethereumReceiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return withdrawAsync(ownerPrivateKey, tokenAddr, ethereumReceiver, bn.Int64(), chain33Relayer.bridgeBankAddr, chain33Relayer.chainName, chain33Relayer.rpcLaddr)
}

func (chain33Relayer *Relayer4Chain33) BurnWithIncreaseAsyncFromChain33(ownerPrivateKey, tokenAddr, ethereumReceiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return burnWithIncreaseAsync(ownerPrivateKey, tokenAddr, ethereumReceiver, bn.Int64(), chain33Relayer.bridgeBankAddr, chain33Relayer.chainName, chain33Relayer.rpcLaddr)
}
