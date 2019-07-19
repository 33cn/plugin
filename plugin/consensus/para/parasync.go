// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"fmt"
	"errors"
	"sync/atomic"

	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/common"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"time"
)


//NextActionType 定义每一轮可执行状态
type NextActionType int8
const (
	_ NextActionType = iota
	//NextActionRollback 回滚到前一区块
	NextActionRollback
	//NextActionKeep 保持
	NextActionKeep
	//NextActionAdd 增加一个新的区块
	NextActionAdd
)

//获取同步状态，供发送层调用
func (client *client) SyncHasCaughtUp() bool {
	return  atomic.LoadInt32(&client.syncCaughtUpAtom) == 1
}

//下载状态通知，供下载层调用
func (client *client) NotifyLocalChange() {
	atomic.StoreInt32(&client.localChangeAtom,1)
}

//创建创世区块
func (client *client) CreateGenesisBlock(newblock *types.Block) error {
	return client.writeBlock(zeroHash[:], newblock)
}

//区块执行线程
//循环执行
func (client *client) SyncBlocks() {

	client.syncInit()
	isSyncCaughtUp := false
	for {
		//获取同步状态,在需要同步的情况下执行同步
		curSyncCaughtState, err := client.syncBlocksIfNeed()
		if err != nil {
			client.printError(err)
		}

		//同步状态改变，发出通知并保存新状态
		if curSyncCaughtState != isSyncCaughtUp {
			isSyncCaughtUp = curSyncCaughtState
			client.setSyncCaughtUp(curSyncCaughtState)
		}

		//没有需要同步的块,清理本地数据库中localCacheCount前的块
		canCleanLocalBlocks := isSyncCaughtUp &&
			!client.getAndFlipLocalChangeStateIfNeed()
		if canCleanLocalBlocks {
			cleanUpSomeBlocks, err := client.clearLocalOldBlocks()
			if err != nil {
				client.printError(err)
			}
			if !cleanUpSomeBlocks {
				time.Sleep(time.Second)
			}
		}

	}
}

//获取每一轮可执行状态
func (client *client) getNextAction() (NextActionType,*types.Block,*pt.ParaLocalDbBlock,int64,error) {
	lastBlock, err := client.getLastBlockInfo()
	if  err != nil {
        //取已执行最新区块发生错误，不做任何操作
		return  NextActionKeep,nil,nil,-1,err
	}

	lastLocalHeight, err := client.getLastLocalHeight()
	if  err != nil {
		//取db中最新高度区块发生错误，不做任何操作
		return  NextActionKeep,nil,nil,lastLocalHeight,err
	}

	if lastLocalHeight <= 0  {
		//db中最新高度为0,不做任何操作（创世区块）
		return  NextActionKeep,nil,nil,lastLocalHeight,err
	} else if lastLocalHeight < lastBlock.Height {
		//db中最新区块高度小于已执行最新区块高度,回滚
		return NextActionRollback,lastBlock,nil,lastLocalHeight,err
	} else if lastLocalHeight == lastBlock.Height {
		localBlock, err := client.getLocalBlockByHeight(lastBlock.Height)
		if  err != nil {
			//取db中指定高度区块发生错误，不做任何操作
			return  NextActionKeep,nil,nil,lastLocalHeight,err
		}
		if common.ToHex(localBlock.MainHash) == common.ToHex(lastBlock.MainHash)   {
			//db中最新区块高度等于已执行最新区块高度并且hash相同,不做任何操作(已保持同步状态)
			return  NextActionKeep,nil,nil,lastLocalHeight,err
		}
		//db中最新区块高度等于已执行最新区块高度并且hash不同,回滚
		return NextActionRollback,lastBlock,nil,lastLocalHeight,err
	}

	// lastLocalHeight > lastBlock.Height
	localBlock, err := client.getLocalBlockByHeight(lastBlock.Height+1)
	if  err != nil {
		//取db中后一高度区块发生错误，不做任何操作
		return  NextActionKeep,nil,nil,lastLocalHeight,err
	}
	if common.ToHex(localBlock.ParentMainHash) != common.ToHex(lastBlock.MainHash)  {
		//db中后一高度区块的父hash不等于已执行最新区块的hash,回滚
		return NextActionRollback,lastBlock,nil,lastLocalHeight,err
	}
	//db中后一高度区块的父hash等于已执行最新区块的hash,执行区块创建
	return  NextActionAdd,lastBlock,localBlock,lastLocalHeight,err
}

//根据当前可执行状态执行区块操作
//返回参数
//bool 是否已完成同步
func (client *client) syncBlocksIfNeed() (bool,error) {
	nextAction, lastBlock, localBlock,lastLocalHeight, err := client.getNextAction()
	if err != nil {
		return  false,err
	}

	switch nextAction {
	case NextActionAdd:
		//1 db中后一高度区块的父hash等于已执行最新区块的hash
		plog.Info("Para sync add block",
			"lastBlock.Height",lastBlock.Height,"lastLocalHeight",lastLocalHeight)
		return false,client.addBlock(lastBlock, localBlock)
	case NextActionRollback:
		//1 db中最新区块高度小于已执行最新区块高度
		//2 db中最新区块高度等于已执行最新区块高度并且hash不同
		//3 db中后一高度区块的父hash不等于已执行最新区块的hash
		plog.Info("Para sync rollback block",
			"lastBlock.Height",lastBlock.Height,"lastLocalHeight",lastLocalHeight)
		return false,client.rollbackBlock(lastBlock)
	default: //NextActionKeep
	    //1 已完成同步，没有需要同步的块
		return  true,err
	}

}

//批量删除下载层缓冲数据
func (client *client) delLocalBlocks(startHeight int64,endHeight int64) error {
	if startHeight > endHeight {
		return  errors.New("startHeight > endHeight,can't clear local blocks")
	}

	index := startHeight
	set := &types.LocalDBSet{}
	for {
		if index > endHeight {
			break
		}

		key := calcTitleHeightKey(types.GetTitle(), index)
		kv := &types.KeyValue{Key: key, Value: nil}
		set.KV = append(set.KV, kv)

		index++
	}

	key := calcTitleFirstHeightKey(types.GetTitle())
	kv := &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: endHeight+1})}
	set.KV = append(set.KV, kv)

	plog.Info("Para sync clear local blocks", "startHeight:",startHeight,"endHeight:",endHeight)

	return client.setLocalDb(set)
}

//最低高度没有设置的时候设置一下最低高度
func (client *client) initFirstLocalHeightIfNeed() error {
	height,err := client.getFirstLocalHeight()

	if err != nil || height < 0 {
		set := &types.LocalDBSet{}
		key := calcTitleFirstHeightKey(types.GetTitle())
		kv := &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: 0})}
		set.KV = append(set.KV, kv)

		return client.setLocalDb(set)
	}

	return err
}

//获取下载层缓冲数据的区块最低高度
func (client *client) getFirstLocalHeight() (int64,error) {
	key := calcTitleFirstHeightKey(types.GetTitle())
	set := &types.LocalDBGet{Keys: [][]byte{key}}
	value, err := client.getLocalDb(set, len(set.Keys))
	if err != nil {
		return -1, err
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
func (client *client) clearLocalOldBlocks() (bool,error) {
	lastLocalHeight, err := client.getLastLocalHeight()
	if err != nil {
		return  false,err
	}

	firstLocalHeight,err := client.getFirstLocalHeight()
	if err != nil {
		return  false,err
	}

	canDelCount := lastLocalHeight - firstLocalHeight - localCacheCount + 1
	if canDelCount <= 0 {
		return  false,nil
	}

	return  true,client.delLocalBlocks(firstLocalHeight,firstLocalHeight +canDelCount- 1)
}

// miner tx need all para node create, but not all node has auth account, here just not sign to keep align
func (client *client) addMinerTx(preStateHash []byte, block *types.Block,localBlock *pt.ParaLocalDbBlock) error {
	status := &pt.ParacrossNodeStatus{
		Title:           types.GetTitle(),
		Height:          block.Height,
		PreBlockHash:    block.ParentHash,
		PreStateHash:    preStateHash,
		MainBlockHash:   localBlock.MainHash,
		MainBlockHeight: localBlock.MainHeight,
	}

	tx, err := pt.CreateRawMinerTx(&pt.ParacrossMinerAction{
		Status:          status,
		IsSelfConsensus: isParaSelfConsensusForked(status.MainBlockHeight),
	})
	if err != nil {
		return err
	}

	tx.Sign(types.SECP256K1, client.privateKey)
	block.Txs = append([]*types.Transaction{tx}, block.Txs...)

	return nil
}

//添加一个区块
func (client *client) addBlock(lastBlock *types.Block,localBlock *pt.ParaLocalDbBlock ) error {
	var newBlock types.Block
	plog.Debug(fmt.Sprintf("the len txs is: %v", len(localBlock.Txs)))

	newBlock.ParentHash = lastBlock.Hash()
	newBlock.Height = lastBlock.Height + 1
	newBlock.Txs = localBlock.Txs
	err := client.addMinerTx(lastBlock.StateHash, &newBlock, localBlock)
	if err != nil {
		return err
	}
	//挖矿固定难度
	newBlock.Difficulty = types.GetP(0).PowLimitBits
	newBlock.TxHash = merkle.CalcMerkleRoot(newBlock.Txs)
	newBlock.BlockTime = localBlock.BlockTime
	newBlock.MainHash = localBlock.MainHash
	newBlock.MainHeight = localBlock.MainHeight

	err = client.writeBlock(lastBlock.StateHash, &newBlock)

	plog.Debug("para create new Block", "newblock.ParentHash", common.ToHex(newBlock.ParentHash),
		"newblock.Height", newBlock.Height, "newblock.TxHash", common.ToHex(newBlock.TxHash),
		"newblock.BlockTime", newBlock.BlockTime)

	return err
}

// 向blockchain删区块
func (client *client) rollbackBlock(block *types.Block) error {
	plog.Debug("delete block in parachain")

	start := block.Height
	if start == 0 {
		panic("Parachain attempt to Delete GenesisBlock !")
	}

	msg := client.GetQueueClient().NewMessage("blockchain", types.EventGetBlocks, &types.ReqBlocks{Start: start, End: start, IsDetail: true, Pid: []string{""}})
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}
	blocks := resp.GetData().(*types.BlockDetails)

	parablockDetail := &types.ParaChainBlockDetail{Blockdetail: blocks.Items[0]}
	msg = client.GetQueueClient().NewMessage("blockchain", types.EventDelParaChainBlockDetail, parablockDetail)
	err = client.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}
	resp, err = client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}

	if resp.GetData().(*types.Reply).IsOk {
		if client.authAccount != "" {
			client.commitMsgClient.updateChainHeight(blocks.Items[0].Block.Height, true)

		}
	} else {
		reply := resp.GetData().(*types.Reply)
		return errors.New(string(reply.GetMsg()))
	}
	return nil
}

// 向blockchain写区块
func (client *client) writeBlock(prev []byte, paraBlock *types.Block) error {
	//共识模块不执行block，统一由blockchain模块执行block并做去重的处理，返回执行后的blockdetail
	blockDetail := &types.BlockDetail{Block: paraBlock}

	parablockDetail := &types.ParaChainBlockDetail{Blockdetail: blockDetail}
	msg := client.GetQueueClient().NewMessage("blockchain", types.EventAddParaChainBlockDetail, parablockDetail)
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}
	blkdetail := resp.GetData().(*types.BlockDetail)
	if blkdetail == nil {
		return errors.New("block detail is nil")
	}

	client.SetCurrentBlock(blkdetail.Block)

	if client.authAccount != "" {
		client.commitMsgClient.updateChainHeight(blockDetail.Block.Height, false)
	}

	return nil
}

//设置同步状态，原子操作，线程访问安全,原则上只限于此线程单元使用
func (client *client) setSyncCaughtUp(isSyncCaughtUp bool) {
	if isSyncCaughtUp {
		atomic.StoreInt32(&client.syncCaughtUpAtom,1)
	} else {
		atomic.StoreInt32(&client.syncCaughtUpAtom,0)
	}
}

//初始化下载状态,原则上只限于此线程单元使用
func (client *client) initLocalChangeState() {
	atomic.StoreInt32(&client.localChangeAtom,0)
}

//获取当前是否有新的下载到来,获取一次，并马上把状态设置为没有新通知
//此函数原则上只限于此线程单元使用
func (client *client) getAndFlipLocalChangeStateIfNeed() bool {
	hasLocalChange := atomic.LoadInt32(&client.localChangeAtom) == 1
	if hasLocalChange {
		atomic.StoreInt32(&client.localChangeAtom,0)
	}
	return hasLocalChange
}


//打印错误日志
func (client *client) printError(err error) {
	plog.Error(fmt.Sprintf("----------------->Para Sync Block Error:%v", err.Error()))
}


//初始化
func (client *client) syncInit() {
	client.setSyncCaughtUp(false)
	client.initLocalChangeState()
	//false
	err := client.initFirstLocalHeightIfNeed()
	if err != nil {
		client.printError(err)
	}
}








