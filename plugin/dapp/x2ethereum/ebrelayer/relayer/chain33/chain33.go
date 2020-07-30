package chain33

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethinterface"
	relayerTx "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
	syncTx "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/relayer/chain33/transceiver/sync"
	ebTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/utils"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
	ethCommon "github.com/ethereum/go-ethereum/common"
)

var relayerLog = log.New("module", "chain33_relayer")

//Relayer4Chain33 ...
type Relayer4Chain33 struct {
	syncTxReceipts      *syncTx.TxReceipts
	ethClient           ethinterface.EthClientSpec
	rpcLaddr            string //用户向指定的blockchain节点进行rpc调用
	fetchHeightPeriodMs int64
	db                  dbm.DB
	lastHeight4Tx       int64 //等待被处理的具有相应的交易回执的高度
	matDegree           int32 //成熟度         heightSync2App    matDegress   height
	//passphase            string
	privateKey4Ethereum  *ecdsa.PrivateKey
	ethSender            ethCommon.Address
	bridgeRegistryAddr   ethCommon.Address
	oracleInstance       *generated.Oracle
	totalTx4Chain33ToEth int64
	statusCheckedIndex   int64
	ctx                  context.Context
	rwLock               sync.RWMutex
	unlock               chan int
}

// StartChain33Relayer : initializes a relayer which witnesses events on the chain33 network and relays them to Ethereum
func StartChain33Relayer(ctx context.Context, syncTxConfig *ebTypes.SyncTxConfig, registryAddr, provider string, db dbm.DB) *Relayer4Chain33 {
	chian33Relayer := &Relayer4Chain33{
		rpcLaddr:            syncTxConfig.Chain33Host,
		fetchHeightPeriodMs: syncTxConfig.FetchHeightPeriodMs,
		unlock:              make(chan int),
		db:                  db,
		ctx:                 ctx,
		bridgeRegistryAddr:  ethCommon.HexToAddress(registryAddr),
	}

	syncCfg := &ebTypes.SyncTxReceiptConfig{
		Chain33Host:       syncTxConfig.Chain33Host,
		PushHost:          syncTxConfig.PushHost,
		PushName:          syncTxConfig.PushName,
		PushBind:          syncTxConfig.PushBind,
		StartSyncHeight:   syncTxConfig.StartSyncHeight,
		StartSyncSequence: syncTxConfig.StartSyncSequence,
		StartSyncHash:     syncTxConfig.StartSyncHash,
	}

	client, err := relayerTx.SetupWebsocketEthClient(provider)
	if err != nil {
		panic(err)
	}
	chian33Relayer.ethClient = client
	chian33Relayer.totalTx4Chain33ToEth = chian33Relayer.getTotalTxAmount2Eth()
	chian33Relayer.statusCheckedIndex = chian33Relayer.getStatusCheckedIndex()

	go chian33Relayer.syncProc(syncCfg)
	return chian33Relayer
}

//QueryTxhashRelay2Eth ...
func (chain33Relayer *Relayer4Chain33) QueryTxhashRelay2Eth() ebTypes.Txhashes {
	txhashs := utils.QueryTxhashes([]byte(chain33ToEthBurnLockTxHashPrefix), chain33Relayer.db)
	return ebTypes.Txhashes{Txhash: txhashs}
}

func (chain33Relayer *Relayer4Chain33) syncProc(syncCfg *ebTypes.SyncTxReceiptConfig) {
	_, _ = fmt.Fprintln(os.Stdout, "Pls unlock or import private key for Chain33 relayer")
	<-chain33Relayer.unlock
	_, _ = fmt.Fprintln(os.Stdout, "Chain33 relayer starts to run...")

	chain33Relayer.syncTxReceipts = syncTx.StartSyncTxReceipt(syncCfg, chain33Relayer.db)
	chain33Relayer.lastHeight4Tx = chain33Relayer.loadLastSyncHeight()

	oracleInstance, err := relayerTx.RecoverOracleInstance(chain33Relayer.ethClient, chain33Relayer.bridgeRegistryAddr, chain33Relayer.bridgeRegistryAddr)
	if err != nil {
		panic(err.Error())
	}
	chain33Relayer.oracleInstance = oracleInstance

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
	chain33Relayer.rwLock.Lock()
	for chain33Relayer.statusCheckedIndex < chain33Relayer.totalTx4Chain33ToEth {
		index := chain33Relayer.statusCheckedIndex + 1
		txhash, err := chain33Relayer.getEthTxhash(index)
		if nil != err {
			relayerLog.Error("onNewHeightProc", "getEthTxhash for index ", index, "error", err.Error())
			break
		}
		status := relayerTx.GetEthTxStatus(chain33Relayer.ethClient, txhash)
		//按照提交交易的先后顺序检查交易，只要出现当前交易还在pending状态，就不再检查后续交易，等到下个区块再从该交易进行检查
		//TODO:可能会由于网络和打包挖矿的原因，使得交易执行顺序和提交顺序有差别，后续完善该检查逻辑
		if status == relayerTx.EthTxPending.String() {
			break
		}
		_ = chain33Relayer.setLastestRelay2EthTxhash(status, txhash.Hex(), index)
		atomic.AddInt64(&chain33Relayer.statusCheckedIndex, 1)
		_ = chain33Relayer.setStatusCheckedIndex(chain33Relayer.statusCheckedIndex)
	}
	chain33Relayer.rwLock.Unlock()
	//未达到足够的成熟度，不进行处理
	//  +++++++++||++++++++++++||++++++++++||
	//           ^             ^           ^
	// lastHeight4Tx    matDegress   currentHeight
	for chain33Relayer.lastHeight4Tx+int64(chain33Relayer.matDegree)+1 <= currentHeight {
		relayerLog.Info("onNewHeightProc", "currHeight", currentHeight, "lastHeight4Tx", chain33Relayer.lastHeight4Tx)

		lastHeight4Tx := chain33Relayer.lastHeight4Tx
		TxReceipts, err := chain33Relayer.syncTxReceipts.GetNextValidTxReceipts(lastHeight4Tx)
		if nil == TxReceipts || nil != err {
			if err != nil {
				relayerLog.Error("onNewHeightProc", "Failed to GetNextValidTxReceipts due to:", err.Error())
			}
			break
		}
		relayerLog.Debug("onNewHeightProc", "currHeight", currentHeight, "valid tx receipt with height:", TxReceipts.Height)

		txs := TxReceipts.Tx
		for i, tx := range txs {
			//检查是否为lns的交易(包括平行链：user.p.xxx.lns)，将闪电网络交易进行收集
			if 0 != bytes.Compare(tx.Execer, []byte(relayerTx.X2Eth)) &&
				(len(tx.Execer) > 4 && string(tx.Execer[(len(tx.Execer)-4):]) != "."+relayerTx.X2Eth) {
				relayerLog.Debug("onNewHeightProc, the tx is not x2ethereum", "Execer", string(tx.Execer), "height:", TxReceipts.Height)
				continue
			}
			var ss types.X2EthereumAction
			_ = chain33Types.Decode(tx.Payload, &ss)
			actionName := ss.GetActionName()
			if relayerTx.BurnAction == actionName || relayerTx.LockAction == actionName {
				relayerLog.Debug("^_^ ^_^ Processing chain33 tx receipt", "ActionName", actionName, "fromAddr", tx.From(), "exec", string(tx.Execer))
				actionEvent := getOracleClaimType(actionName)
				if err := chain33Relayer.handleBurnLockMsg(actionEvent, TxReceipts.ReceiptData[i], tx.Hash()); nil != err {
					errInfo := fmt.Sprintf("Failed to handleBurnLockMsg due to:%s", err.Error())
					panic(errInfo)
				}
			}
		}
		chain33Relayer.lastHeight4Tx = TxReceipts.Height
		chain33Relayer.setLastSyncHeight(chain33Relayer.lastHeight4Tx)
	}
}

// getOracleClaimType : sets the OracleClaim's claim type based upon the witnessed event type
func getOracleClaimType(eventType string) events.Event {
	var claimType events.Event

	switch eventType {
	case events.MsgBurn.String():
		claimType = events.Event(events.ClaimTypeBurn)
	case events.MsgLock.String():
		claimType = events.Event(events.ClaimTypeLock)
	default:
		panic(errors.New("eventType invalid"))
	}

	return claimType
}

// handleBurnLockMsg : parse event data as a Chain33Msg, package it into a ProphecyClaim, then relay tx to the Ethereum Network
func (chain33Relayer *Relayer4Chain33) handleBurnLockMsg(claimEvent events.Event, receipt *chain33Types.ReceiptData, chain33TxHash []byte) error {
	relayerLog.Info("handleBurnLockMsg", "Received tx with hash", ethCommon.Bytes2Hex(chain33TxHash))

	// Parse the witnessed event's data into a new Chain33Msg
	chain33Msg := relayerTx.ParseBurnLockTxReceipt(claimEvent, receipt)
	if nil == chain33Msg {
		//收到执行失败的交易，直接跳过
		relayerLog.Error("handleBurnLockMsg", "Received failed tx with hash", ethCommon.Bytes2Hex(chain33TxHash))
		return nil
	}

	// Parse the Chain33Msg into a ProphecyClaim for relay to Ethereum
	prophecyClaim := relayerTx.Chain33MsgToProphecyClaim(*chain33Msg)

	// Relay the Chain33Msg to the Ethereum network
	txhash, err := relayerTx.RelayOracleClaimToEthereum(chain33Relayer.oracleInstance, chain33Relayer.ethClient, chain33Relayer.ethSender, claimEvent, prophecyClaim, chain33Relayer.privateKey4Ethereum, chain33TxHash)
	if nil != err {
		return err
	}

	//保存交易hash，方便查询
	atomic.AddInt64(&chain33Relayer.totalTx4Chain33ToEth, 1)
	txIndex := atomic.LoadInt64(&chain33Relayer.totalTx4Chain33ToEth)
	if err = chain33Relayer.updateTotalTxAmount2Eth(txIndex); nil != err {
		relayerLog.Error("handleLogNewProphecyClaimEvent", "Failed to RelayLockToChain33 due to:", err.Error())
		return err
	}
	if err = chain33Relayer.setLastestRelay2EthTxhash(relayerTx.EthTxPending.String(), txhash, txIndex); nil != err {
		relayerLog.Error("handleLogNewProphecyClaimEvent", "Failed to RelayLockToChain33 due to:", err.Error())
		return err
	}
	return nil
}
