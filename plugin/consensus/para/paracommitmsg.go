// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/33cn/chain33/common/address"

	"strings"

	"sync/atomic"
	"unsafe"

	"sync"

	"strconv"

	"bytes"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	paracross "github.com/33cn/plugin/plugin/dapp/paracross/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

const (
	consensusInterval = 10 //about 1 new block interval
	minerInterval     = 10 //5s的主块间隔后分叉概率增加，10s可以消除一些分叉回退

	waitBlocks4CommitMsg int32  = 5 //commit msg共识发送后等待几个块没确认则重发
	waitConsensStopTimes uint32 = 3 //3*10s
)

type paraSelfConsEnable struct {
	startHeight int64
	endHeight   int64
}

type commitMsgClient struct {
	paraClient           *client
	waitMainBlocks       int32  //等待平行链共识消息在主链上链并成功的块数，超出会重发共识消息
	waitConsensStopTimes uint32 //共识高度低于完成高度， reset高度重发等待的次数
	resetCh              chan interface{}
	sendMsgCh            chan *types.Transaction
	currentTx            unsafe.Pointer
	chainHeight          int64
	sendingHeight        int64
	consensHeight        int64
	consensDoneHeight    int64
	txFeeRate            int64
	minerSwitch          int32
	selfConsensError     int32 //自共识比主链共识更高的异常场景，需要等待自共识<=主链共识再发送
	authAccount          string
	authAccountIn        bool
	isRollBack           int32
	checkTxCommitTimes   int32
	selfConsEnableList   []*paraSelfConsEnable //适配在自共识合约配置前有自共识的平行链项目，fork之后，采用合约配置
	privateKey           crypto.PrivKey
	addressId            int32
	quit                 chan struct{}
	mutex                sync.Mutex
}

type commitCheckParams struct {
	consensStopTimes uint32
}

func newCommitMsgCli(para *client, cfg *subConfig) *commitMsgClient {
	cli := &commitMsgClient{
		paraClient:           para,
		authAccount:          cfg.AuthAccount,
		waitMainBlocks:       waitBlocks4CommitMsg,
		waitConsensStopTimes: waitConsensStopTimes,
		consensHeight:        -2,
		sendingHeight:        -1,
		consensDoneHeight:    -1,
		resetCh:              make(chan interface{}, 1),
		quit:                 make(chan struct{}),
	}
	if cfg.WaitBlocks4CommitMsg > 0 {
		cli.waitMainBlocks = cfg.WaitBlocks4CommitMsg
	}

	if cfg.WaitConsensStopTimes > 0 {
		cli.waitConsensStopTimes = cfg.WaitConsensStopTimes
	}

	// 设置平行链共识起始高度，在共识高度为-1也就是从未共识过的环境中允许从设置的非0起始高度开始共识
	//note：只有在主链LoopCheckCommitTxDoneForkHeight之后才支持设置ParaConsensStartHeight
	if cfg.ParaConsensStartHeight > 0 {
		cli.consensDoneHeight = cfg.ParaConsensStartHeight - 1
	}
	return cli
}

// 1. 链高度回滚，低于当前发送高度，需要重新计算当前发送高度,不然不会重新发送回滚的高度
// 2. 定时轮询是在比如锁定解锁钱包这类外部条件变化时候，其他输入条件不会触发时候及时响应，不然任何一个外部条件变化都触发一下发送，可能条件比较多
func (client *commitMsgClient) handler() {
	var readTick <-chan time.Time
	checkParams := &commitCheckParams{}

	client.paraClient.wg.Add(1)
	go client.getMainConsensusInfo()

	if client.authAccount != "" {
		client.paraClient.wg.Add(1)
		client.sendMsgCh = make(chan *types.Transaction, 1)
		go client.sendCommitMsg()

		ticker := time.NewTicker(time.Second * time.Duration(minerInterval))
		readTick = ticker.C
		defer ticker.Stop()
	}

out:
	for {
		select {
		//出错场景入口，需要reset 重发
		case <-client.resetCh:
			client.resetSend()
			client.createCommitTx()
		//例行检查发送入口,及时触发未发送共识
		case <-readTick:
			client.procChecks(checkParams)
			client.createCommitTx()

		case <-client.quit:
			break out
		}
	}

	client.paraClient.wg.Done()
}

//chain height更新时候入口
func (client *commitMsgClient) updateChainHeightNotify(height int64, isDel bool) {
	if isDel {
		atomic.StoreInt32(&client.isRollBack, 1)
	} else {
		atomic.StoreInt32(&client.isRollBack, 0)
	}

	atomic.StoreInt64(&client.chainHeight, height)

	client.checkRollback(height)
	client.createCommitTx()
}

func (client *commitMsgClient) setInitChainHeight(height int64) {
	atomic.StoreInt64(&client.chainHeight, height)
}

// reset notify 提供重设发送参数，发送tx的入口
func (client *commitMsgClient) resetNotify() {
	client.resetCh <- 1
}

//新的区块产生，检查是否有commitTx正在发送入口
func (client *commitMsgClient) commitTxCheckNotify(block *types.ParaTxDetail) {
	if client.checkCommitTxSuccess(block) {
		client.createCommitTx()
	}
}

func (client *commitMsgClient) resetSendEnv() {
	client.sendingHeight = -1
	client.setCurrentTx(nil)
}
func (client *commitMsgClient) resetSend() {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.resetSendEnv()
}

//自共识后直接从本地获取最新共识高度，没有自共识，获取主链的共识高度
func (client *commitMsgClient) getConsensusHeight() int64 {
	status, err := client.getSelfConsensus()
	if err != nil {
		return atomic.LoadInt64(&client.consensHeight)
	}

	return status.Height
}

func (client *commitMsgClient) createCommitTx() {
	tx := client.getCommitTx()
	if tx == nil {
		return
	}
	//如果配置了blsSign 则发送到p2p的leader节点来聚合发送，否则发送到主链
	if client.paraClient.blsSignCli.blsSignOn {
		client.pushCommitTx2P2P(tx)
		return
	}
	client.pushCommitTx(tx)
}

//四个触发：1,新增区块 2,10s tick例行检查 3,发送交易成功上链 4,异常重发
//1&2　只要共识高度追赶上了sendingHeight，就可以继续发送，即便当前节点发送交易仍未上链也直接取消发送新交易
func (client *commitMsgClient) getCommitTx() *types.Transaction {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	consensHeight := client.getConsensusHeight()
	//只有从未共识过，才可以设置从初始起始高度跳跃
	if consensHeight == -1 && consensHeight < client.consensDoneHeight {
		consensHeight = client.consensDoneHeight
	}

	chainHeight := atomic.LoadInt64(&client.chainHeight)
	sendingHeight := client.sendingHeight
	if sendingHeight < consensHeight {
		sendingHeight = consensHeight
	}

	isSync := client.isSync()
	plog.Info("para commitMsg---status", "chainHeight", chainHeight, "sendingHeight", sendingHeight,
		"consensHeight", consensHeight, "isSendingTx", client.isSendingCommitMsg(), "sync", isSync)

	if !isSync {
		return nil
	}

	//1.如果是在主链共识场景，共识高度可能大于平行链的链高度
	//2.已发送，未共识场景
	if sendingHeight > consensHeight || consensHeight > chainHeight || sendingHeight >= chainHeight {
		return nil
	}

	//满足　sendingHeight <= consensHeight <= chainHeight && sendingHeight < chainHeight
	signTx, count := client.getSendingTx(sendingHeight, chainHeight)
	if signTx == nil {
		return nil
	}
	client.sendingHeight = sendingHeight + count
	return signTx

}

//client.checkTxCommitTimes和client.sendingHeight锁的场景可以区分
//发送commitTx，可能跟checkCommitTxSuccess获取全局变量冲突，加锁，　如果有仍未成功上链的交易，直接覆盖重置
func (client *commitMsgClient) pushCommitTx(signTx *types.Transaction) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.checkTxCommitTimes = 0
	client.setCurrentTx(signTx)
	client.sendMsgCh <- signTx
}

//仍旧setCurrentTx， 这样在几个块之后仍旧会触发重发，重发只是广播，不然发送p2p之后，如果共识没增加，也没有其他触发的条件了
func (client *commitMsgClient) pushCommitTx2P2P(signTx *types.Transaction) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.checkTxCommitTimes = 0
	client.setCurrentTx(signTx)

	plog.Debug("bls.event.para bls commitMs send to p2p", "hash", common.ToHex(signTx.Hash()))
	act := &pt.ParaP2PSubMsg{Ty: P2pSubCommitTx, Value: &pt.ParaP2PSubMsg_CommitTx{CommitTx: signTx}}
	client.paraClient.SendPubP2PMsg(paraBlsSignTopic, types.Encode(act))
}

//根据收集的commit action，签名发送, 比如BLS签名后的commit msg
func (client *commitMsgClient) sendCommitActions(acts []*pt.ParacrossCommitAction) {
	//如果当前正在发送交易，则取消此次发送，待发送被确认或取消后再触发. 考虑到已经聚合共识成功，又收到某节点消息场景，会多发送交易
	curTx := client.getCurrentTx()
	if curTx != nil {
		plog.Info("bls.event.paracommitmsg isSendingCommitMsg, cancel this operation", "sending.tx", common.ToHex(curTx.Hash()))
		return
	}

	txs, _, err := client.createCommitMsgTxs(acts)
	if err != nil {
		return
	}
	plog.Info("bls.event.paracommitmsg sendCommitActions", "txhash", common.ToHex(txs.Hash()))
	for i, msg := range acts {
		plog.Debug("paracommitmsg sendCommitActions", "idx", i, "height", msg.Status.Height, "mainheight", msg.Status.MainBlockHeight,
			"blockhash", common.HashHex(msg.Status.BlockHash), "mainHash", common.HashHex(msg.Status.MainBlockHash),
			"addrsmap", hex.EncodeToString(msg.Bls.AddrsMap), "sign", common.ToHex(msg.Bls.Sign))
	}
	client.pushCommitTx(txs)
}

func (client *commitMsgClient) checkTxIn(block *types.ParaTxDetail, tx *types.Transaction) bool {
	//committx是平行链交易
	if types.IsParaExecName(string(tx.Execer)) {
		for _, tx := range block.TxDetails {
			if bytes.HasSuffix(tx.Tx.Execer, []byte(pt.ParaX)) && tx.Receipt.Ty == types.ExecOk {
				return true
			}
		}
		return false
	}

	//主链交易，向主链查询,平行链获取到的只是过滤了的平行链交易
	receipt, _ := client.paraClient.QueryTxOnMainByHash(tx.Hash())
	if receipt != nil && receipt.Receipt.Ty == types.ExecOk {
		return true
	}
	return false
}

func (client *commitMsgClient) checkCommitTxSuccess(block *types.ParaTxDetail) bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	curTx := client.getCurrentTx()
	if curTx == nil {
		return false
	}

	//当前addType是回滚，则不计数，如果有累计则撤销上次累计次数，重新计数
	if block.Type != types.AddBlock {
		if client.checkTxCommitTimes > 0 {
			client.checkTxCommitTimes--
		}
		return false
	}

	if client.checkTxIn(block, curTx) {
		client.setCurrentTx(nil)
		return true
	}

	return client.reSendCommitTx(curTx)
}

func (client *commitMsgClient) reSendCommitTx(tx *types.Transaction) bool {
	client.checkTxCommitTimes++
	if client.checkTxCommitTimes < client.waitMainBlocks {
		return false
	}
	client.checkTxCommitTimes = 0
	client.resetSendEnv()
	return true
}

//如果共识高度一直没有追上发送高度，超出等待时间后，说明共识一直没达成，安全起见，超过停止次数后，重发
func (client *commitMsgClient) checkConsensusStop(checks *commitCheckParams) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	consensHeight := client.getConsensusHeight()
	if client.sendingHeight > consensHeight {
		checks.consensStopTimes++
		if checks.consensStopTimes > client.waitConsensStopTimes {
			plog.Debug("para checkConsensusStop", "times", checks.consensStopTimes, "consens", consensHeight, "send", client.sendingHeight)
			checks.consensStopTimes = 0
			client.resetSendEnv()
		}
	}
}

func (client *commitMsgClient) checkAuthAccountIn() {
	nodeStr, err := client.getNodeGroupAddrs()
	nodeSupervisionStr, errSupervision := client.getSupervisionNodeGroupAddrs() // 判断是否是监督节点
	if err != nil && errSupervision != nil {
		return
	}

	authExist1 := strings.Contains(nodeStr, client.authAccount)
	authExist2 := strings.Contains(nodeSupervisionStr, client.authAccount)
	authExist := authExist1 || authExist2

	//如果授权节点重新加入，需要从当前共识高度重新发送
	if !client.authAccountIn && authExist {
		client.resetSend()
	}

	client.authAccountIn = authExist
}

func (client *commitMsgClient) procChecks(checks *commitCheckParams) {
	client.checkConsensusStop(checks)
	client.checkAuthAccountIn()
}

func (client *commitMsgClient) isSync() bool {
	height := atomic.LoadInt64(&client.chainHeight)
	if height <= 0 {
		plog.Info("para is not Sync", "chainHeight", height)
		return false
	}

	height = atomic.LoadInt64(&client.consensHeight)
	if height == -2 {
		plog.Info("para is not Sync", "consensHeight", height)
		return false
	}

	if atomic.LoadInt32(&client.selfConsensError) != 0 {
		plog.Info("para is not Sync", "selfConsensError", atomic.LoadInt32(&client.selfConsensError))
		return false
	}

	if !client.authAccountIn {
		plog.Info("para is not Sync", "authAccountIn", client.authAccountIn)
		return false
	}

	if atomic.LoadInt32(&client.minerSwitch) != 1 {
		plog.Info("para is not Sync", "isMiner", atomic.LoadInt32(&client.minerSwitch))
		return false
	}

	if atomic.LoadInt32(&client.isRollBack) == 1 {
		plog.Info("para is not Sync", "isRollBack", atomic.LoadInt32(&client.isRollBack))
		return false
	}

	if !client.paraClient.isCaughtUp() {
		plog.Info("para is not Sync", "caughtUp", client.paraClient.isCaughtUp())
		return false
	}

	if !client.paraClient.blockSyncClient.syncHasCaughtUp() {
		plog.Info("para is not Sync", "syncCaughtUp", client.paraClient.blockSyncClient.syncHasCaughtUp())
		return false
	}

	return true
}

func (client *commitMsgClient) getSendingTx(startHeight, endHeight int64) (*types.Transaction, int64) {
	count := endHeight - startHeight
	if count > int64(types.MaxTxGroupSize) {
		count = int64(types.MaxTxGroupSize)
	}
	status, err := client.getNodeStatus(startHeight+1, startHeight+count)
	if err != nil {
		plog.Error("para commit msg read tick", "err", err.Error())
		return nil, 0
	}
	if len(status) == 0 {
		return nil, 0
	}

	var commits []*pt.ParacrossCommitAction
	for _, stat := range status {
		commits = append(commits, &pt.ParacrossCommitAction{Status: stat})
	}

	if client.paraClient.blsSignCli.blsSignOn {
		err = client.paraClient.blsSignCli.blsSign(commits)
		if err != nil {
			plog.Error("paracommitmsg bls sign", "err", err)
			return nil, 0
		}
	}

	signTx, count, err := client.createCommitMsgTxs(commits)
	if err != nil || signTx == nil {
		return nil, 0
	}

	sendingMsgs := status[:count]
	plog.Debug("paracommitmsg sending", "txhash", common.ToHex(signTx.Hash()), "exec", string(signTx.Execer))
	for i, msg := range sendingMsgs {
		plog.Debug("paracommitmsg sending", "idx", i, "height", msg.Height, "mainheight", msg.MainBlockHeight,
			"blockhash", common.HashHex(msg.BlockHash), "mainHash", common.HashHex(msg.MainBlockHash),
			"from", client.authAccount)
	}

	return signTx, count
}

func (client *commitMsgClient) createCommitMsgTxs(notifications []*pt.ParacrossCommitAction) (*types.Transaction, int64, error) {
	txs, count, err := client.batchCalcTxGroup(notifications, atomic.LoadInt64(&client.txFeeRate))
	if err != nil {
		txs, err = client.singleCalcTx((notifications)[0], atomic.LoadInt64(&client.txFeeRate))
		if err != nil {
			plog.Error("single calc tx", "height", notifications[0].Status.Height)

			return nil, 0, err
		}
		return txs, 1, nil
	}
	return txs, int64(count), nil
}

func (client *commitMsgClient) getTxsGroup(txsArr *types.Transactions) (*types.Transaction, error) {
	if len(txsArr.Txs) < 2 {
		tx := txsArr.Txs[0]
		tx.Sign(types.EncodeSignID(types.SECP256K1, client.addressId), client.privateKey)
		return tx, nil
	}
	cfg := client.paraClient.GetAPI().GetConfig()
	group, err := types.CreateTxGroup(txsArr.Txs, cfg.GetMinTxFeeRate())
	if err != nil {
		plog.Error("para CreateTxGroup", "err", err.Error())
		return nil, err
	}
	err = group.Check(cfg, 0, cfg.GetMinTxFeeRate(), cfg.GetMaxTxFee())
	if err != nil {
		plog.Error("para CheckTxGroup", "err", err.Error())
		return nil, err
	}
	for i := range group.Txs {
		group.SignN(i, types.EncodeSignID(types.SECP256K1, client.addressId), client.privateKey)
	}

	newtx := group.Tx()
	return newtx, nil
}

func (client *commitMsgClient) getExecName(commitHeight int64) string {
	cfg := client.paraClient.GetAPI().GetConfig()
	if cfg.IsDappFork(commitHeight, pt.ParaX, pt.ForkParaFullMinerHeight) {
		return paracross.GetExecName(cfg)
	}

	if cfg.IsDappFork(commitHeight, pt.ParaX, pt.ForkParaSelfConsStages) {
		return paracross.GetExecName(cfg)
	}

	execName := pt.ParaX
	if client.isSelfConsEnable(commitHeight) {
		execName = paracross.GetExecName(cfg)
	}
	return execName

}

func (client *commitMsgClient) batchCalcTxGroup(notifications []*pt.ParacrossCommitAction, feeRate int64) (*types.Transaction, int, error) {
	var rawTxs types.Transactions
	cfg := client.paraClient.GetAPI().GetConfig()
	for i, notify := range notifications {
		if i >= int(types.MaxTxGroupSize) {
			break
		}
		execName := client.getExecName(notify.Status.Height)
		tx, err := paracross.CreateRawCommitTx4MainChain(cfg, notify, execName, feeRate)
		if err != nil {
			plog.Error("para get commit tx", "block height", notify.Status.Height)
			return nil, 0, err
		}
		rawTxs.Txs = append(rawTxs.Txs, tx)
	}

	txs, err := client.getTxsGroup(&rawTxs)
	if err != nil {
		return nil, 0, err
	}
	return txs, len(notifications), nil
}

func (client *commitMsgClient) singleCalcTx(notify *pt.ParacrossCommitAction, feeRate int64) (*types.Transaction, error) {
	cfg := client.paraClient.GetAPI().GetConfig()
	execName := client.getExecName(notify.Status.Height)
	tx, err := paracross.CreateRawCommitTx4MainChain(cfg, notify, execName, feeRate)
	if err != nil {
		plog.Error("para get commit tx", "block height", notify.Status.Height)
		return nil, err
	}
	tx.Sign(types.EncodeSignID(types.SECP256K1, client.addressId), client.privateKey)
	return tx, nil

}

func (client *commitMsgClient) setCurrentTx(tx *types.Transaction) {
	atomic.StorePointer(&client.currentTx, unsafe.Pointer(tx))
}

func (client *commitMsgClient) getCurrentTx() *types.Transaction {
	return (*types.Transaction)(atomic.LoadPointer(&client.currentTx))
}

func (client *commitMsgClient) isSendingCommitMsg() bool {
	return client.getCurrentTx() != nil
}

func (client *commitMsgClient) checkRollback(height int64) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if height < client.sendingHeight {
		client.resetSendEnv()
	}
}

func (client *commitMsgClient) sendCommitTxOut(tx *types.Transaction) error {
	if tx == nil {
		return nil
	}
	resp, err := client.paraClient.grpcClient.SendTransaction(context.Background(), tx)
	if err != nil {
		plog.Error("sendCommitTxOut send tx", "tx", common.ToHex(tx.Hash()), "err", err.Error())
		return err
	}

	if !resp.GetIsOk() {
		plog.Error("sendCommitTxOut send tx Nok", "tx", common.ToHex(tx.Hash()), "err", string(resp.GetMsg()))
		return errors.New(string(resp.GetMsg()))
	}

	return nil

}

func needResentErr(err error) bool {
	switch err {
	case nil, types.ErrBalanceLessThanTenTimesFee, types.ErrNoBalance, types.ErrDupTx, types.ErrTxExist, types.ErrTxExpire:
		return false
	default:
		return true
	}
}

func (client *commitMsgClient) sendCommitMsg() {
	var err error
	var tx *types.Transaction
	var resendTimer <-chan time.Time

out:
	for {
		select {
		case tx = <-client.sendMsgCh:
			err = client.sendCommitTxOut(tx)
			if err != nil && err == types.ErrTxFeeTooLow {
				err := client.GetProperFeeRate()
				if err == nil {
					client.resetNotify()
				}
				continue
			}
			if needResentErr(err) {
				resendTimer = time.After(time.Second * 2)
			}
		case <-resendTimer:
			if err != nil && tx != nil {
				client.sendCommitTxOut(tx)
			}
		case <-client.quit:
			break out
		}
	}

	client.paraClient.wg.Done()
}

//当前未考虑获取key非常多失败的场景， 如果获取height非常多，block模块会比较大，但是使用完了就释放了
//如果有必要也可以考虑每次最多取20个一个txgroup，发送共识部分循环获取发送也没问题
func (client *commitMsgClient) getNodeStatus(start, end int64) ([]*pt.ParacrossNodeStatus, error) {
	var ret []*pt.ParacrossNodeStatus
	if start == 0 {
		geneStatus, err := client.getGenesisNodeStatus()
		if err != nil {
			return nil, err
		}
		ret = append(ret, geneStatus)
		start++
	}
	if end < start {
		return ret, nil
	}

	req := &types.ReqBlocks{Start: start, End: end}
	count := req.End - req.Start + 1
	nodeList := make(map[int64]*pt.ParacrossNodeStatus, count+1)
	keys := &types.LocalDBGet{}
	cfg := client.paraClient.GetAPI().GetConfig()
	for i := 0; i < int(count); i++ {
		key := paracross.CalcMinerHeightKey(cfg.GetTitle(), req.Start+int64(i))
		keys.Keys = append(keys.Keys, key)
	}

	r, err := client.paraClient.GetAPI().LocalGet(keys)
	if err != nil {
		return nil, err
	}
	if count != int64(len(r.Values)) {
		plog.Error("paracommitmsg get node status key", "expect count", count, "actual count", len(r.Values))
		return nil, err
	}
	for _, val := range r.Values {
		status := &pt.ParacrossNodeStatus{}
		err = types.Decode(val, status)
		if err != nil {
			return nil, err
		}
		if !(status.Height >= req.Start && status.Height <= req.End) {
			plog.Error("paracommitmsg decode node status", "height", status.Height, "expect start", req.Start,
				"end", req.End, "status", status)
			return nil, errors.New("paracommitmsg wrong key result")
		}
		nodeList[status.Height] = status

	}
	for i := 0; i < int(count); i++ {
		if nodeList[req.Start+int64(i)] == nil {
			plog.Error("paracommitmsg get node status key nil", "height", req.Start+int64(i))
			return nil, errors.New("paracommitmsg wrong key status result")
		}
	}

	v, err := client.paraClient.GetAPI().GetBlocks(req)
	if err != nil {
		return nil, err
	}
	if count != int64(len(v.Items)) {
		plog.Error("paracommitmsg get node status block", "expect count", count, "actual count", len(v.Items))
		return nil, err
	}
	for _, block := range v.Items {
		if !(block.Block.Height >= req.Start && block.Block.Height <= req.End) {
			plog.Error("paracommitmsg get node status block", "height", block.Block.Height, "expect start", req.Start, "end", req.End)
			return nil, errors.New("paracommitmsg wrong block result")
		}
		nodeList[block.Block.Height].BlockHash = block.Block.Hash(cfg)
		if !paracross.IsParaForkHeight(cfg, nodeList[block.Block.Height].MainBlockHeight, paracross.ForkLoopCheckCommitTxDone) {
			nodeList[block.Block.Height].StateHash = block.Block.StateHash
		}
	}

	var needSentTxs uint32
	for i := 0; i < int(count); i++ {
		ret = append(ret, nodeList[req.Start+int64(i)])
		needSentTxs += nodeList[req.Start+int64(i)].NonCommitTxCounts
	}
	//1.如果是只有commit tx的空块，推迟发送，直到等到一个完全没有commit tx的空块或者其他tx的块
	//2,如果20个块都是 commit tx的空块，20个块打包一次发送，尽量减少commit tx造成的空块
	//3,如果形如xxoxx的块排列，x代表commit空块，o代表实际的块，即只要不全部是commit块，也要全部打包一起发出去
	//如果=0 意味着全部是paracross commit tx，延迟发送
	if needSentTxs == 0 && len(ret) < int(types.MaxTxGroupSize) {
		plog.Info("para commit tx are all self-consensus tx,postpone send ", "start", start, "end", end)
		return nil, nil
	}

	//clear flag
	for _, v := range ret {
		v.NonCommitTxCounts = 0
	}

	return ret, nil

}

func (client *commitMsgClient) getGenesisNodeStatus() (*pt.ParacrossNodeStatus, error) {
	var status pt.ParacrossNodeStatus
	req := &types.ReqBlocks{Start: 0, End: 0}
	v, err := client.paraClient.GetAPI().GetBlocks(req)
	if err != nil {
		return nil, err
	}
	block := v.Items[0].Block
	if block.Height != 0 {
		return nil, errors.New("block chain not return 0 height block")
	}
	cfg := client.paraClient.GetAPI().GetConfig()
	status.Title = cfg.GetTitle()
	status.Height = block.Height
	status.BlockHash = block.Hash(cfg)

	return &status, nil
}

//only sync once, as main usually sync, here just need the first sync status after start up
func (client *commitMsgClient) mainSync() error {
	req := &types.ReqNil{}
	reply, err := client.paraClient.grpcClient.IsSync(context.Background(), req)
	if err != nil {
		plog.Error("Paracross main is syncing", "err", err.Error())
		return err
	}
	if !reply.IsOk {
		plog.Error("Paracross main reply not ok")
		return err
	}

	plog.Info("Paracross main sync succ")
	return nil

}

func (client *commitMsgClient) getMainConsensusInfo() {
	ticker := time.NewTicker(time.Second * time.Duration(consensusInterval))
	isSync := false
	defer ticker.Stop()

out:
	for {
		select {
		case <-client.quit:
			break out
		case <-ticker.C:
			if !isSync {
				err := client.mainSync()
				if err != nil {
					continue
				}
				isSync = true
			}

			if client.authAccount != "" {
				client.GetProperFeeRate()
			}

			selfHeight := int64(-2)
			selfStatus, _ := client.getSelfConsensusStatus()
			if selfStatus != nil {
				selfHeight = selfStatus.Height
			}

			mainStatus, err := client.getMainConsensusStatus()
			if err != nil {
				continue
			}

			//如果主链的共识高度小于自共识高度，需要等待自共识回滚
			if mainStatus.Height < selfHeight {
				atomic.StoreInt32(&client.selfConsensError, 1)
			} else {
				atomic.StoreInt32(&client.selfConsensError, 0)
			}

			preHeight := atomic.LoadInt64(&client.consensHeight)
			atomic.StoreInt64(&client.consensHeight, mainStatus.Height)
			//如果主链的共识高度产生了回滚，本地链也需要重新检查共识高度,不然可能会一直等待共识追赶上来
			if mainStatus.Height < preHeight {
				client.resetNotify()
			}

			plog.Info("para consensusHeight", "mainHeight", mainStatus.Height, "selfHeight", selfHeight)
		}
	}

	client.paraClient.wg.Done()
}

func (client *commitMsgClient) GetProperFeeRate() error {
	feeRate, err := client.paraClient.grpcClient.GetProperFee(context.Background(), &types.ReqProperFee{})
	if err != nil {
		plog.Error("para commit.GetProperFee", "err", err.Error())
		return err
	}
	if feeRate == nil {
		plog.Error("para commit.GetProperFee return nil")
		return types.ErrInvalidParam
	}

	atomic.StoreInt64(&client.txFeeRate, feeRate.ProperFee)
	return nil
}

//在自共识阶段获取共识高度
func (client *commitMsgClient) getSelfConsensus() (*pt.ParacrossStatus, error) {
	block, err := client.paraClient.getLastBlockInfo()
	if err != nil {
		return nil, err
	}
	ret, err := client.paraClient.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "paracross",
		FuncName: "GetSelfConsOneStage",
		Param:    types.Encode(&types.Int64{Data: block.Height}),
	})
	if err != nil {
		plog.Debug("getSelfConsensusStatus.GetSelfConsOneStage ", "err", err.Error())
		return nil, err
	}
	stage, ok := ret.(*pt.SelfConsensStage)
	if !ok {
		plog.Error("getSelfConsensusStatus nok")
		return nil, types.ErrInvalidParam
	}
	if stage.Enable == pt.ParaConfigYes {
		resp, err := client.getSelfConsensusStatus()
		if err != nil {
			return nil, err
		}
		//开启自共识后也要等到自共识真正切换之后再使用，如果本地区块已经过了自共识高度，但自共识的高度还没达成，就会导致共识机制出错
		if resp.Height > stage.StartHeight {
			return resp, nil
		}
	}
	return nil, types.ErrNotFound
}

//从本地查询共识高度
func (client *commitMsgClient) getSelfConsensusStatus() (*pt.ParacrossStatus, error) {
	cfg := client.paraClient.GetAPI().GetConfig()
	ret, err := client.paraClient.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "paracross",
		FuncName: "GetTitle",
		Param:    types.Encode(&types.ReqString{Data: cfg.GetTitle()}),
	})
	if err != nil {
		plog.Error("getSelfConsensusStatus ", "err", err)
		return nil, err
	}
	resp, ok := ret.(*pt.ParacrossStatus)
	if !ok {
		plog.Error("getSelfConsensusStatus ParacrossStatus nok")
		return nil, types.ErrNotFound
	}
	return resp, nil

}

//通过grpc获取主链状态可能耗时，放在定时器里面处理
func (client *commitMsgClient) getMainConsensusStatus() (*pt.ParacrossStatus, error) {
	block, err := client.paraClient.getLastBlockInfo()
	if err != nil {
		return nil, err
	}
	cfg := client.paraClient.GetAPI().GetConfig()
	//去主链获取共识高度
	reply, err := client.paraClient.grpcClient.QueryChain(context.Background(), &types.ChainExecutor{
		Driver:   "paracross",
		FuncName: "GetTitleByHash",
		Param:    types.Encode(&pt.ReqParacrossTitleHash{Title: cfg.GetTitle(), BlockHash: block.MainHash}),
	})
	if err != nil {
		plog.Error("getMainConsensusStatus", "err", err.Error())
		return nil, err
	}
	if !reply.GetIsOk() {
		plog.Info("getMainConsensusStatus nok", "error", reply.GetMsg())
		return nil, types.ErrNotFound
	}
	var result pt.ParacrossStatus
	err = types.Decode(reply.Msg, &result)
	if err != nil {
		plog.Error("getMainConsensusStatus decode", "err", err.Error())
		return nil, err
	}
	return &result, nil

}

//node group会在主链和平行链都同时配置,只本地查询就可以
func (client *commitMsgClient) getNodeGroupAddrs() (string, error) {
	cfg := client.paraClient.GetAPI().GetConfig()
	ret, err := client.paraClient.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "paracross",
		FuncName: "GetNodeGroupAddrs",
		Param:    types.Encode(&pt.ReqParacrossNodeInfo{Title: cfg.GetTitle()}),
	})
	if err != nil {
		plog.Error("commitmsg.getNodeGroupAddrs ", "err", err.Error())
		return "", err
	}
	resp, ok := ret.(*types.ReplyConfig)
	if !ok {
		plog.Error("commitmsg.getNodeGroupAddrs rsp nok")
		return "", err
	}

	return resp.Value, nil
}

//Supervision node group会在主链和平行链都同时配置,只本地查询就可以
func (client *commitMsgClient) getSupervisionNodeGroupAddrs() (string, error) {
	cfg := client.paraClient.GetAPI().GetConfig()
	ret, err := client.paraClient.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "paracross",
		FuncName: "GetSupervisionNodeGroupAddrs",
		Param:    types.Encode(&pt.ReqParacrossNodeInfo{Title: cfg.GetTitle()}),
	})
	if err != nil {
		plog.Error("commitmsg.getSupervisionNodeGroupAddrs ", "err", err.Error())
		return "", err
	}
	resp, ok := ret.(*types.ReplyConfig)
	if !ok {
		plog.Error("commitmsg.getSupervisionNodeGroupAddrs rsp nok")
		return "", err
	}

	return resp.Value, nil
}

func (client *commitMsgClient) onWalletStatus(status *types.WalletStatus) {
	if status == nil || client.authAccount == "" {
		plog.Info("para onWalletStatus", "status", status == nil, "auth", client.authAccount == "")
		return
	}
	if !status.IsWalletLock && client.privateKey == nil {
		plog.Info("para commit fetchPriKey try")
		client.fetchPriKey()
		plog.Info("para commit fetchPriKey ok")
	}

	if client.privateKey == nil {
		plog.Info("para commit wallet status prikey null", "status", status.IsWalletLock)
		return
	}

	if status.IsWalletLock {
		atomic.StoreInt32(&client.minerSwitch, 0)
	} else {
		atomic.StoreInt32(&client.minerSwitch, 1)
	}

}

func (client *commitMsgClient) onWalletAccount(acc *types.Account) {
	if acc == nil || client.authAccount == "" || client.authAccount != acc.Addr || client.privateKey != nil {
		return
	}
	plog.Error("para onWalletAccount try fetch prikey")
	err := client.fetchPriKey()
	if err != nil {
		plog.Error("para onWalletAccount", "err", err.Error())
		return
	}

	atomic.StoreInt32(&client.minerSwitch, 1)

}

func getSecpPriKey(key string) (crypto.PrivKey, error) {
	pk, err := common.FromHex(key)
	if err != nil && pk == nil {
		return nil, errors.Wrapf(err, "fromhex=%s", key)
	}

	secp, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		return nil, errors.Wrapf(err, "crypto=%s", key)
	}

	priKey, err := secp.PrivKeyFromBytes(pk)
	if err != nil {
		return nil, errors.Wrapf(err, "fromBytes=%s", key)
	}

	return priKey, nil
}

func (client *commitMsgClient) fetchPriKey() error {
	req := &types.ReqString{Data: client.authAccount}

	resp, err := client.paraClient.GetAPI().ExecWalletFunc("wallet", "DumpPrivkey", req)
	if err != nil {
		plog.Error("para fetchPriKey dump priKey", "err", err)
		return err
	}
	str := resp.(*types.ReplyString).Data
	priKey, err := getSecpPriKey(str)
	if err != nil {
		plog.Error("para fetchPriKey get priKey", "err", err)
		return err
	}

	client.privateKey = priKey
	client.paraClient.blsSignCli.setBlsPriKey(priKey.Bytes())

  addressId, err := address.GetAddressType(client.authAccount)
  if err != nil {
    client.addressId = address.GetDefaultAddressID()
  } else {
    client.addressId = addressId
  }

	return nil
}

func parseSelfConsEnableStr(selfEnables []string) ([]*paraSelfConsEnable, error) {
	var list []*paraSelfConsEnable
	for _, e := range selfEnables {
		ret, err := divideStr2Int64s(e, "-")
		if err != nil {
			return nil, err
		}
		list = append(list, &paraSelfConsEnable{ret[0], ret[1]})
	}
	return list, nil
}

//only for "0:50" or "0-50" with one sep
func divideStr2Int64s(s, sep string) ([]int64, error) {
	var r []int64
	a := strings.Split(s, sep)
	if len(a) != 2 {
		plog.Error("error format for config to separate", "s", s)
		return nil, types.ErrInvalidParam
	}

	for _, v := range a {
		val, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			plog.Error("error format for config to parse to int", "s", s)
			return nil, err
		}
		r = append(r, val)
	}
	return r, nil
}

func (client *commitMsgClient) setSelfConsEnable() error {
	cfg := client.paraClient.GetAPI().GetConfig()
	selfEnables := types.Conf(cfg, pt.ParaPrefixConsSubConf).GStrList(pt.ParaSelfConsConfPreContract)
	list, err := parseSelfConsEnableStr(selfEnables)
	if err != nil {
		return err
	}
	client.selfConsEnableList = append(client.selfConsEnableList, list...)
	return nil
}

//适配在自共识合约配置前有自共识的平行链项目，fork之后，采用合约配置
func (client *commitMsgClient) isSelfConsEnable(height int64) bool {
	for _, v := range client.selfConsEnableList {
		if height >= v.startHeight && height <= v.endHeight {
			return true
		}
	}
	return false
}
