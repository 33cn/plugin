// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

const (
	//defaultMaxCacheCount 默认local最大缓冲数
	defaultMaxCacheCount = int64(1000)
	//defaultMaxSyncErrCount 默认连续错误最大数量
	defaultMaxSyncErrCount = int32(100)
)

//blockSyncClient 区块同步控制和状态变量
type blockSyncClient struct {
	paraClient *client
	//notifyChan 下载通知通道
	notifyChan chan bool
	//quitChan 线程退出通知通道
	quitChan chan struct{}
	//syncState 同步状态
	syncState int32
	//syncErrMaxCount 同步错误最大数量
	maxSyncErrCount int32
	//maxCacheCount 本地缓冲的最大块数量
	maxCacheCount int64
	//isSyncCaughtUp 是否追赶上
	isSyncCaughtUpAtom int32
	//isDownloadCaughtUpAtom  下载是否已经追赶上
	isDownloadCaughtUpAtom int32
	//isSyncFirstCaughtUp 系统启动后download 层和sync层第一次都追赶上的设置，后面如果因为回滚或节点切换不同步，则不考虑
	isSyncFirstCaughtUp bool
}

//nextActionType 定义每一轮可执行操作
type nextActionType int8

const (
	//nextActionKeep 保持
	nextActionKeep nextActionType = iota
	//nextActionRollback 回滚到前一区块
	nextActionRollback
	//nextActionAdd 增加一个新的区块
	nextActionAdd
)

//blockSyncState 定义当前区块同步状态
type blockSyncState int32

const (
	//blockSyncStateNone 未同步状态
	blockSyncStateNone blockSyncState = iota
	//blockSyncStateSyncing 正在同步中
	blockSyncStateSyncing
	//blockSyncStateFinished 同步完成
	blockSyncStateFinished
)

func newBlockSyncCli(para *client, cfg *subConfig) *blockSyncClient {
	cli := &blockSyncClient{
		paraClient:      para,
		notifyChan:      make(chan bool, 1),
		quitChan:        make(chan struct{}),
		maxCacheCount:   defaultMaxCacheCount,
		maxSyncErrCount: defaultMaxSyncErrCount,
	}
	if cfg.MaxCacheCount > 0 {
		cli.maxCacheCount = cfg.MaxCacheCount
	}
	if cfg.MaxSyncErrCount > 0 {
		cli.maxSyncErrCount = cfg.MaxSyncErrCount
	}
	return cli
}

//syncHasCaughtUp 判断同步是否已追赶上，供发送层调用
func (client *blockSyncClient) syncHasCaughtUp() bool {
	return atomic.LoadInt32(&client.isSyncCaughtUpAtom) == 1
}

//handleLocalChangedMsg 处理下载通知消息，供下载层调用
func (client *blockSyncClient) handleLocalChangedMsg() {
	client.printDebugInfo("Para sync - notify change")
	if client.getBlockSyncState() == blockSyncStateSyncing || client.paraClient.isCancel() {
		return
	}
	client.printDebugInfo("Para sync - notified change")
	client.notifyChan <- true
}

//handleLocalCaughtUpMsg 处理下载已追赶上消息，供下载层调用
func (client *blockSyncClient) handleLocalCaughtUpMsg() {
	client.printDebugInfo("Para sync -notify download has caughtUp")
	if !client.downloadHasCaughtUp() {
		client.setDownloadHasCaughtUp(true)
	}
}

//createGenesisBlock 创建创世区块
func (client *blockSyncClient) createGenesisBlock(newblock *types.Block) error {
	return client.writeBlock(zeroHash[:], newblock)
}

//syncBlocks 区块执行线程
//循环执行
func (client *blockSyncClient) syncBlocks() {

	client.syncInit()
	//首次同步，不用等待通知
	client.batchSyncBlocks()
	//开始正常同步,需要等待通知信号触发
out:
	for {
		select {
		case <-client.notifyChan:

			client.batchSyncBlocks()

		case <-client.quitChan:
			break out
		}
	}

	plog.Info("Para sync - quit block sync goroutine")
	client.paraClient.wg.Done()
}

//批量执行同步区块
func (client *blockSyncClient) batchSyncBlocks() {
	client.setBlockSyncState(blockSyncStateSyncing)
	client.printDebugInfo("Para sync - syncing")

	errCount := int32(0)
	for {
		if client.paraClient.isCancel() {
			return
		}
		//获取同步状态,在需要同步的情况下执行同步
		curSyncCaughtState, err := client.syncBlocksIfNeed()

		//统计连续出错数量
		if err != nil {
			errCount++
			client.printError(err)
		} else {
			errCount = int32(0)
		}
		//连续出错达到一定数量,退出循环，等待下一次通知
		if errCount > client.maxSyncErrCount {
			client.printError(errors.New(
				"para sync - sync has some errors,please check"))
			client.setBlockSyncState(blockSyncStateNone)
			return
		}
		//没有需要同步的块,清理本地数据库中localCacheCount前的块
		if err == nil && curSyncCaughtState {
			err := client.clearLocalOldBlocks()
			if err != nil {
				client.printError(err)
			}

			client.setBlockSyncState(blockSyncStateFinished)
			client.printDebugInfo("Para sync - finished")
			return
		}
	}

}

//获取每一轮可执行状态
func (client *blockSyncClient) getNextAction() (nextActionType, *types.Block, *pt.ParaLocalDbBlock, int64, error) {
	lastBlock, err := client.paraClient.getLastBlockInfo()
	if err != nil {
		//取已执行最新区块发生错误，不做任何操作
		return nextActionKeep, nil, nil, -1, err
	}

	lastLocalHeight, err := client.paraClient.getLastLocalHeight()
	if err != nil {
		//取db中最新高度区块发生错误，不做任何操作
		return nextActionKeep, nil, nil, lastLocalHeight, err
	}

	if lastLocalHeight <= 0 {
		//db中最新高度为0,不做任何操作（创世区块）
		return nextActionKeep, nil, nil, lastLocalHeight, nil
	}

	switch {
	case lastLocalHeight < lastBlock.Height:
		//db中最新区块高度小于已执行最新区块高度,回滚
		return nextActionRollback, lastBlock, nil, lastLocalHeight, nil
	case lastLocalHeight == lastBlock.Height:
		localBlock, err := client.paraClient.getLocalBlockByHeight(lastBlock.Height)
		if err != nil {
			//取db中指定高度区块发生错误，不做任何操作
			return nextActionKeep, nil, nil, lastLocalHeight, err
		}
		if common.ToHex(localBlock.MainHash) == common.ToHex(lastBlock.MainHash) {
			//db中最新区块高度等于已执行最新区块高度并且hash相同,不做任何操作(已保持同步状态)
			return nextActionKeep, nil, nil, lastLocalHeight, nil
		}
		//db中最新区块高度等于已执行最新区块高度并且hash不同,回滚
		return nextActionRollback, lastBlock, nil, lastLocalHeight, nil
	default:
		// lastLocalHeight > lastBlock.Height
		localBlock, err := client.paraClient.getLocalBlockByHeight(lastBlock.Height + 1)
		if err != nil {
			//取db中后一高度区块发生错误，不做任何操作
			return nextActionKeep, nil, nil, lastLocalHeight, err
		}
		if common.ToHex(localBlock.ParentMainHash) != common.ToHex(lastBlock.MainHash) {
			//db中后一高度区块的父hash不等于已执行最新区块的hash,回滚
			return nextActionRollback, lastBlock, nil, lastLocalHeight, nil
		}
		//db中后一高度区块的父hash等于已执行最新区块的hash,执行区块创建
		return nextActionAdd, lastBlock, localBlock, lastLocalHeight, nil
	}
}

//根据当前可执行状态执行区块操作
//返回参数
//bool 是否已完成同步
func (client *blockSyncClient) syncBlocksIfNeed() (bool, error) {
	nextAction, lastBlock, localBlock, lastLocalHeight, err := client.getNextAction()
	if err != nil {
		return false, err
	}

	switch nextAction {
	case nextActionAdd:
		//1 db中后一高度区块的父hash等于已执行最新区块的hash
		plog.Info("Para sync -    add block",
			"lastBlock.Height", lastBlock.Height, "lastLocalHeight", lastLocalHeight)

		err := client.addBlock(lastBlock, localBlock)

		//通知发送层
		if err == nil {
			isSyncCaughtUp := lastBlock.Height+1 == lastLocalHeight
			client.setSyncCaughtUp(isSyncCaughtUp)
			if client.paraClient.commitMsgClient.authAccount != "" {
				client.printDebugInfo("Para sync - add block commit", "isSyncCaughtUp", isSyncCaughtUp)
				client.paraClient.commitMsgClient.updateChainHeightNotify(lastBlock.Height+1, false)
			}
		}

		return false, err
	case nextActionRollback:
		//1 db中最新区块高度小于已执行最新区块高度
		//2 db中最新区块高度等于已执行最新区块高度并且hash不同
		//3 db中后一高度区块的父hash不等于已执行最新区块的hash
		plog.Info("Para sync -    rollback block",
			"lastBlock.Height", lastBlock.Height, "lastLocalHeight", lastLocalHeight)

		err := client.rollbackBlock(lastBlock)

		//通知发送层
		if err == nil {
			client.setSyncCaughtUp(false)
			if client.paraClient.commitMsgClient.authAccount != "" {
				client.printDebugInfo("Para sync - rollback block commit", "isSyncCaughtUp", false)
				client.paraClient.commitMsgClient.updateChainHeightNotify(lastBlock.Height-1, true)
			}
		}

		return false, err
	default: //nextActionKeep
		//1 已完成同步，没有需要同步的块
		return true, nil
	}

}

//批量删除下载层缓冲数据
func (client *blockSyncClient) delLocalBlocks(startHeight int64, endHeight int64) error {
	if startHeight > endHeight {
		return errors.New("para sync - startHeight > endHeight,can't clear local blocks")
	}

	plog.Info("Para sync - clear local blocks", "startHeight:", startHeight, "endHeight:", endHeight)

	index := startHeight
	set := &types.LocalDBSet{}
	cfg := client.paraClient.GetAPI().GetConfig()
	for {
		if index > endHeight {
			break
		}

		key := calcTitleHeightKey(cfg.GetTitle(), index)
		kv := &types.KeyValue{Key: key, Value: nil}
		set.KV = append(set.KV, kv)

		index++
	}

	key := calcTitleFirstHeightKey(cfg.GetTitle())
	kv := &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: endHeight + 1})}
	set.KV = append(set.KV, kv)

	return client.paraClient.setLocalDb(set)
}

//最低高度没有设置的时候设置一下最低高度
func (client *blockSyncClient) initFirstLocalHeightIfNeed() error {
	height, err := client.getFirstLocalHeight()
	cfg := client.paraClient.GetAPI().GetConfig()
	if err != nil || height < 0 {
		set := &types.LocalDBSet{}
		key := calcTitleFirstHeightKey(cfg.GetTitle())
		kv := &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: 0})}
		set.KV = append(set.KV, kv)

		return client.paraClient.setLocalDb(set)
	}

	return err
}

//获取下载层缓冲数据的区块最低高度
func (client *blockSyncClient) getFirstLocalHeight() (int64, error) {
	cfg := client.paraClient.GetAPI().GetConfig()
	key := calcTitleFirstHeightKey(cfg.GetTitle())
	set := &types.LocalDBGet{Keys: [][]byte{key}}
	value, err := client.paraClient.getLocalDb(set, len(set.Keys))
	if err != nil {
		return -1, err
	}

	if len(value) == 0 {
		return -1, types.ErrNotFound
	}

	if value[0] == nil {
		return -1, types.ErrNotFound
	}

	height := &types.Int64{}
	err = types.Decode(value[0], height)
	if err != nil {
		return -1, err
	}
	return height.Data, nil
}

//清除指定数量(localCacheCount)以前的区块
func (client *blockSyncClient) clearLocalOldBlocks() error {
	lastLocalHeight, err := client.paraClient.getLastLocalHeight()
	if err != nil {
		return err
	}

	firstLocalHeight, err := client.getFirstLocalHeight()
	if err != nil {
		return err
	}

	canDelCount := lastLocalHeight - firstLocalHeight - client.maxCacheCount + 1
	count := canDelCount / client.maxCacheCount
	for i := int64(0); i < count; i++ {
		start := firstLocalHeight + i*client.maxCacheCount
		end := start + client.maxCacheCount - 1
		err = client.delLocalBlocks(start, end)
		if err != nil {
			return err
		}
	}
	return nil
}

// miner tx need all para node create, but not all node has auth account, here just not sign to keep align
func (client *blockSyncClient) addMinerTx(preStateHash []byte, block *types.Block, localBlock *pt.ParaLocalDbBlock) error {
	cfg := client.paraClient.GetAPI().GetConfig()
	status := &pt.ParacrossNodeStatus{
		Title:           cfg.GetTitle(),
		Height:          block.Height,
		MainBlockHash:   localBlock.MainHash,
		MainBlockHeight: localBlock.MainHeight,
	}

	maxHeight := pt.GetDappForkHeight(cfg, pt.ForkLoopCheckCommitTxDone)
	if maxHeight < client.paraClient.subCfg.RmCommitParamMainHeight {
		maxHeight = client.paraClient.subCfg.RmCommitParamMainHeight
	}
	if status.MainBlockHeight < maxHeight {
		status.PreBlockHash = block.ParentHash
		status.PreStateHash = preStateHash
	}

	//selfConsensEnablePreContract 是ForkParaSelfConsStages之前对是否开启共识的设置，fork之后采用合约控制，两个高度不能重叠
	tx, err := pt.CreateRawMinerTx(cfg, &pt.ParacrossMinerAction{
		Status:          status,
		IsSelfConsensus: client.paraClient.commitMsgClient.isSelfConsEnable(status.Height),
	})
	if err != nil {
		return err
	}

	tx.Sign(types.SECP256K1, client.paraClient.minerPrivateKey)
	block.Txs = append([]*types.Transaction{tx}, block.Txs...)

	return nil
}

//添加一个区块
func (client *blockSyncClient) addBlock(lastBlock *types.Block, localBlock *pt.ParaLocalDbBlock) error {
	cfg := client.paraClient.GetAPI().GetConfig()
	var newBlock types.Block
	newBlock.ParentHash = lastBlock.Hash(cfg)
	newBlock.Height = lastBlock.Height + 1
	newBlock.Txs = localBlock.Txs
	err := client.addMinerTx(lastBlock.StateHash, &newBlock, localBlock)
	if err != nil {
		return err
	}
	//挖矿固定难度
	newBlock.Difficulty = cfg.GetP(0).PowLimitBits
	newBlock.BlockTime = localBlock.BlockTime
	newBlock.MainHash = localBlock.MainHash
	newBlock.MainHeight = localBlock.MainHeight
	if newBlock.Height == 1 && newBlock.BlockTime < client.paraClient.cfg.GenesisBlockTime {
		panic("genesisBlockTime　bigger than the 1st block time, need rmv db and reset genesisBlockTime")
	}
	//需要首先对交易进行排序然后再计算TxHash
	if cfg.IsFork(newBlock.GetMainHeight(), "ForkRootHash") {
		newBlock.Txs = types.TransactionSort(newBlock.Txs)
	}
	//在之前版本中CalcMerkleRoot的height是未初始化的MainHeight，等于0，在这个平行链的分叉ForkParaRootHash高度后统一采用新高度
	if cfg.IsDappFork(newBlock.Height, pt.ParaX, pt.ForkParaRootHash) {
		newBlock.TxHash = merkle.CalcMerkleRoot(cfg, newBlock.GetMainHeight(), newBlock.Txs)
	} else {
		newBlock.TxHash = merkle.CalcMerkleRoot(cfg, 0, newBlock.Txs)
	}

	err = client.writeBlock(lastBlock.StateHash, &newBlock)

	client.printDebugInfo("Para sync - create new Block",
		"newblock.ParentHash", common.ToHex(newBlock.ParentHash),
		"newblock.Height", newBlock.Height, "newblock.TxHash", common.ToHex(newBlock.TxHash),
		"newblock.BlockTime", newBlock.BlockTime)

	return err
}

// 向blockchain删区块
func (client *blockSyncClient) rollbackBlock(block *types.Block) error {
	start := block.Height
	if start <= 0 {
		return errors.New("para sync - attempt to rollbackBlock genesisBlock")
	}

	reqBlocks := &types.ReqBlocks{Start: start, End: start, IsDetail: true, Pid: []string{""}}
	msg := client.paraClient.GetQueueClient().NewMessage("blockchain", types.EventGetBlocks, reqBlocks)
	err := client.paraClient.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}

	resp, err := client.paraClient.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}

	blocks := resp.GetData().(*types.BlockDetails)
	if len(blocks.Items) == 0 {
		return errors.New("para sync -blocks Items len = 0 ")
	}

	paraBlockDetail := &types.ParaChainBlockDetail{Blockdetail: blocks.Items[0]}
	msg = client.paraClient.GetQueueClient().NewMessage("blockchain", types.EventDelParaChainBlockDetail, paraBlockDetail)
	err = client.paraClient.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}

	resp, err = client.paraClient.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}

	if !resp.GetData().(*types.Reply).IsOk {
		reply := resp.GetData().(*types.Reply)
		return errors.New(string(reply.GetMsg()))
	}

	return nil
}

// 向blockchain写区块
func (client *blockSyncClient) writeBlock(prev []byte, paraBlock *types.Block) error {
	//共识模块不执行block，统一由blockchain模块执行block并做去重的处理，返回执行后的blockdetail
	blockDetail := &types.BlockDetail{Block: paraBlock}
	//database刷盘设置，默认不刷盘，提高执行速度，系统启动后download 层和sync层第一次都追赶上设置为刷盘，后面如果有回滚或节点切换不同步，则不再改变，减少数据库损坏风险
	if !client.isSyncFirstCaughtUp && client.downloadHasCaughtUp() && client.syncHasCaughtUp() {
		client.isSyncFirstCaughtUp = true
		plog.Info("Para sync - SyncFirstCaughtUp", "Height", paraBlock.Height)
	}

	paraBlockDetail := &types.ParaChainBlockDetail{Blockdetail: blockDetail, IsSync: client.isSyncFirstCaughtUp}
	msg := client.paraClient.GetQueueClient().NewMessage("blockchain", types.EventAddParaChainBlockDetail, paraBlockDetail)
	err := client.paraClient.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}

	resp, err := client.paraClient.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}

	respBlockDetail := resp.GetData().(*types.BlockDetail)
	if respBlockDetail == nil {
		return errors.New("para sync - block detail is nil")
	}

	client.paraClient.SetCurrentBlock(respBlockDetail.Block)

	return nil
}

//获取同步状态
func (client *blockSyncClient) getBlockSyncState() blockSyncState {
	return blockSyncState(atomic.LoadInt32(&client.syncState))
}

//设置同步状态
func (client *blockSyncClient) setBlockSyncState(state blockSyncState) {
	atomic.StoreInt32(&client.syncState, int32(state))
}

//设置是否追赶上
func (client *blockSyncClient) setSyncCaughtUp(isCaughtUp bool) {
	if isCaughtUp {
		atomic.StoreInt32(&client.isSyncCaughtUpAtom, 1)
	} else {
		atomic.StoreInt32(&client.isSyncCaughtUpAtom, 0)
	}
}

//下载是否已经追赶上
func (client *blockSyncClient) downloadHasCaughtUp() bool {
	return atomic.LoadInt32(&client.isDownloadCaughtUpAtom) == 1
}

//设置下载同步追赶状态
func (client *blockSyncClient) setDownloadHasCaughtUp(isCaughtUp bool) {
	if isCaughtUp {
		atomic.CompareAndSwapInt32(&client.isDownloadCaughtUpAtom, 0, 1)
	} else {
		atomic.CompareAndSwapInt32(&client.isDownloadCaughtUpAtom, 1, 0)
	}
}

//打印错误日志
func (client *blockSyncClient) printError(err error) {
	plog.Error(fmt.Sprintf("Para sync - sync block error:%v", err.Error()))
}

//打印调试信息
func (client *blockSyncClient) printDebugInfo(msg string, ctx ...interface{}) {
	plog.Debug(msg, ctx...)
}

//初始化
func (client *blockSyncClient) syncInit() {
	client.printDebugInfo("Para sync - init")
	client.setBlockSyncState(blockSyncStateNone)
	client.setSyncCaughtUp(false)
	client.setDownloadHasCaughtUp(false)
	err := client.initFirstLocalHeightIfNeed()
	if err != nil {
		client.printError(err)
	}

	//设置初始chainHeight,及时发送共识，不然需要等到再产生一个新块才发送
	lastBlock, err := client.paraClient.getLastBlockInfo()
	if err != nil {
		//取已执行最新区块发生错误，不做任何操作
		plog.Info("Para sync init", "err", err)
	} else {
		client.paraClient.commitMsgClient.setInitChainHeight(lastBlock.Height)
	}

}
