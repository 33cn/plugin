// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"time"
	"errors"

	"github.com/33cn/chain33/common"
	paracross "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/33cn/chain33/types"
	"encoding/hex"
)





func (client *client) setLocalBlock(set *types.LocalDBSet) (error) {
	//如果追赶上主链了，则落盘
	if client.isCaughtUp{
		set.Txid = 1
	}

	msg := client.GetQueueClient().NewMessage("blockchain", types.EventSetValueByKey, set)
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}
	if resp.GetData().(*types.Reply).IsOk{
		return nil
	}
	return errors.New(string(resp.GetData().(*types.Reply).GetMsg()))
}


func (client *client) addLocalBlock(height int64, block *paracross.ParaLocalDbBlock) (error) {
	set := &types.LocalDBSet{}

	key := calcTitleHeightKey(types.GetTitle(),height)
	kv := &types.KeyValue{Key: key, Value: types.Encode(block)}
	set.KV = append(set.KV,kv)

	//两个key原子操作
	key = calcTitleLastHeightKey(types.GetTitle())
	kv = &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: height})}
	set.KV = append(set.KV,kv)

	return client.setLocalBlock(set)
}

func (client *client) createLocalBlock(lastBlock *paracross.ParaLocalDbBlock, txs []*types.Transaction, mainBlock *types.BlockSeq) error {
	var newblock paracross.ParaLocalDbBlock

	newblock.Height = lastBlock.Height + 1
	newblock.Txs = txs

	newblock.BlockTime = mainBlock.Detail.Block.BlockTime
	newblock.MainHash = mainBlock.Seq.Hash
	newblock.MainHeight = mainBlock.Detail.Block.Height

	return client.addLocalBlock(newblock.Height, &newblock)
}



func (client *client) delLocalBlock(height int64) (error) {
	set := &types.LocalDBSet{}
	key := calcTitleHeightKey(types.GetTitle(),height)
	kv := &types.KeyValue{Key: key, Value: nil}
	set.KV = append(set.KV,kv)

	//两个key原子操作
	key = calcTitleLastHeightKey(types.GetTitle())
	kv = &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: height-1})}
	set.KV = append(set.KV,kv)

	return client.setLocalBlock(set)
}

// localblock 最小高度不为0，如果minHeight=0，则把localblocks 清空，只设置lastHeight key
func (client *client) removeLocalBlocks(minHeight int64) error {
	set := &types.LocalDBSet{}

	key := calcTitleLastHeightKey(types.GetTitle())
	if minHeight >0{
		kv := &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: minHeight})}
		set.KV = append(set.KV,kv)
	}else {
		kv := &types.KeyValue{Key: key, Value: nil}
		set.KV = append(set.KV,kv)
	}

	return client.setLocalBlock(set)
}



func (client *client) getFromLocalDb(set *types.LocalDBGet, count int) ([][]byte, error) {
	msg := client.GetQueueClient().NewMessage("blockchain", types.EventGetValueByKey, set)
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return nil,err
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return nil,err
	}

	reply := resp.GetData().(*types.LocalReplyValue)
	if len(reply.Values) != count{
		plog.Error("Parachain getFromLocalDb count not match", "expert", count,"real",len(reply.Values))
		return nil, types.ErrInvalidParam
	}

	return reply.Values,nil

}



func (client *client) getLastLocalHeight() (int64, error) {
	key := calcTitleLastHeightKey(types.GetTitle())
	set := &types.LocalDBGet{Keys:[][]byte{key}}
	value,err := client.getFromLocalDb(set,len(set.Keys))
	if err != nil{
		return -1, err
	}
	if value[0] == nil{
		return -1, types.ErrNotFound
	}

	height := &types.Int64{}
	err = types.Decode(value[0],height)
	if err != nil{
		return -1,err
	}
	return height.Data,nil

}

func (client *client) getLocalBlockByHeight(height int64) (*paracross.ParaLocalDbBlock, error) {
	key := calcTitleHeightKey(types.GetTitle(),height)
	set := &types.LocalDBGet{Keys:[][]byte{key}}

	value,err := client.getFromLocalDb(set,len(set.Keys))
	if err != nil{
		return nil, err
	}
	if value[0] == nil{
		return nil, types.ErrNotFound
	}

	var block paracross.ParaLocalDbBlock
	err = types.Decode(value[0],&block)
	if err != nil{
		return nil,err
	}
	return &block,nil


}

//TODO 是否考虑mainHash获取不到，回溯查找？
func (client *client) getLocalBlockInfoByHeight(height int64) (int64, []byte, error) {
	lastBlock, err := client.getLocalBlockByHeight(height)
	if err != nil {
		return -2, nil, err
	}

	mainSeq, err := client.GetSeqByHashOnMainChain(lastBlock.MainHash)
	if err != nil {
		return -2, nil, err
	}
	return mainSeq, lastBlock.MainHash, nil

}

func (client *client) setLocalBlockByChainBlock(chainBlock *types.Block) (error) {
	//根据匹配上的chainblock，设置当前localdb block
	localBlock := &paracross.ParaLocalDbBlock{
		Height:chainBlock.Height,
		MainHeight:chainBlock.MainHeight,
		MainHash:chainBlock.MainHash,
		BlockTime:chainBlock.BlockTime,
	}

	return client.addLocalBlock(localBlock.Height,localBlock)

}


//如果localdb里面没有信息，就从chain block返回，至少有创世区块，然后进入循环匹配切换场景
func (client *client) getLastLocalBlockInfo() (int64, []byte, error) {
	height,err := client.getLastLocalHeight()
	if err == nil{
		mainSeq,mainHash,err := client.getLocalBlockInfoByHeight(height)
		if err == nil{
			return mainSeq, mainHash,nil
		}
	}

	mainSeq,chainBlock,err := client.getLastBlockMainInfo()
	if err != nil{
		return -2,nil,err
	}

	err = client.setLocalBlockByChainBlock(chainBlock)
	if err != nil{
		return -2,nil,err
	}
	return mainSeq,chainBlock.MainHash,nil

}

func (client *client) getLastDbBlock() (*paracross.ParaLocalDbBlock, error) {
	height,err := client.getLastLocalHeight()
	if err != nil {
		return nil, err
	}

	return client.getLocalBlockByHeight(height)
}


func (client *client) reqChainMatchedBlock(startHeight int64) (int64, *types.Block, error) {
	lastBlock, err := client.RequestLastBlock()
	if err != nil {
		plog.Error("Parachain RequestLastBlock fail", "err", err)
		return -2, nil, err
	}

	if lastBlock.Height == 0 {
		return client.syncFromGenesisBlock()
	}

	if startHeight == 0 || startHeight > lastBlock.Height{
		startHeight = lastBlock.Height
	}

	depth := searchHashMatchDepth
	for height := startHeight; height > 0 && depth > 0; height-- {
		block, err := client.GetBlockByHeight(height)
		if err != nil {
			return -2, nil, err
		}
		//当前block结构已经有mainHash和MainHeight但是从blockchain获取的block还没有写入，以后如果获取到，可以替换从minerTx获取
		plog.Info("switchHashMatchedBlock", "lastParaBlockHeight", height, "mainHeight",
			block.MainHeight, "mainHash", hex.EncodeToString(block.MainHash))
		mainSeq, err := client.GetSeqByHashOnMainChain(block.MainHash)
		if err != nil {
			depth--
			if depth == 0 {
				plog.Error("switchHashMatchedBlock depth overflow", "last info:mainHeight", block.MainHeight,
					"mainHash", hex.EncodeToString(block.MainHash), "search startHeight", lastBlock.Height, "curHeight", height,
					"search depth", searchHashMatchDepth)
				panic("search HashMatchedBlock overflow, re-setting search depth and restart to try")
			}
			if height == 1 {
				plog.Error("switchHashMatchedBlock search to height=1 not found", "lastBlockHeight", lastBlock.Height,
					"height1 mainHash", hex.EncodeToString(block.MainHash))
					return client.syncFromGenesisBlock()

			}
			continue
		}

		plog.Info("reqChainMatchedBlock succ", "currHeight", height, "initHeight", lastBlock.Height,
			"new currSeq", mainSeq, "new preMainBlockHash", hex.EncodeToString(block.MainHash))
		return mainSeq, block, nil
	}
	return -2, nil, paracross.ErrParaCurHashNotMatch
}

func (client *client) switchChainMatchedBlock(startHeight int64) (int64, []byte, error) {
	mainSeq, chainBlock,err := client.reqChainMatchedBlock(startHeight)
	if err != nil{
		return -2,nil,err
	}
	err = client.setLocalBlockByChainBlock(chainBlock)
	if err != nil{
		return -2,nil,err
	}
	return mainSeq,chainBlock.MainHash,nil
}

// search base on para block but not last MainBlockHash, last MainBlockHash can not back tracing
func (client *client) switchLocalHashMatchedBlock(currSeq int64) (int64, []byte, error) {
	lastBlock, err := client.getLastDbBlock()
	if err != nil {
		if err == types.ErrNotFound{
			//TODO 或者通知执行层去切换
			return client.switchChainMatchedBlock(0)
		}
		plog.Error("Parachain RequestLastBlock fail", "err", err)
		return -2, nil, err
	}


	for height := lastBlock.Height; height > 0 ; height-- {
		block, err := client.getLocalBlockByHeight(height)
		if err != nil {
			if err == types.ErrNotFound{
				plog.Error("switchLocalHashMatchedBlock search not found", "lastBlockHeight", height)
				err = client.removeLocalBlocks(height)
				if err != nil{
					return -2, nil, err
				}
				return client.switchChainMatchedBlock(height)
			}
			return -2, nil, err
		}
		//当前block结构已经有mainHash和MainHeight但是从blockchain获取的block还没有写入，以后如果获取到，可以替换从minerTx获取
		plog.Info("switchLocalHashMatchedBlock", "lastlocalBlockHeight", height, "mainHeight",
			block.MainHeight, "mainHash", hex.EncodeToString(block.MainHash))
		mainSeq, err := client.GetSeqByHashOnMainChain(block.MainHash)
		if err != nil {
			continue
		}

		//remove fail, the para chain may be remove part, set the preMainBlockHash to nil, to match nothing, force to search from last
		err = client.removeLocalBlocks(height)
		if err != nil {
			return -2, nil, err
		}

		plog.Info("switchLocalHashMatchedBlock succ", "currHeight", height, "initHeight", lastBlock.Height,
			"currSeq", mainSeq, "currMainBlockHash", hex.EncodeToString(block.MainHash))
		return mainSeq, block.MainHash, nil
	}
	return -2, nil, paracross.ErrParaCurHashNotMatch
}



func (client *client) downloadBlocks() {
	lastSeq, lastSeqMainHash, err := client.getLastLocalBlockInfo()
	if err != nil {
		plog.Error("Parachain CreateBlock getLastLocalBlockInfo fail", "err", err.Error())
		return
	}
	currSeq := lastSeq+1
	for {
		txs, mainBlock, err := client.RequestTx(currSeq, lastSeqMainHash)
		if err != nil {
			if err == paracross.ErrParaCurHashNotMatch {
				preSeq, preSeqMainHash, err := client.switchLocalHashMatchedBlock(currSeq)
				if err == nil {
					currSeq = preSeq+1
					lastSeqMainHash = preSeqMainHash
					continue
				}
			}
			time.Sleep(time.Second * time.Duration(blockSec))
			continue
		}

		lastSeqMainHeight := mainBlock.Detail.Block.Height
		lastSeqMainHash = mainBlock.Seq.Hash
		if mainBlock.Seq.Type == delAct {
			lastSeqMainHash = mainBlock.Detail.Block.ParentHash
		}

		lastBlock, err := client.getLastDbBlock()
		if err != nil && err != types.ErrNotFound{
			plog.Error("Parachain getLastDbBlock", "err", err)
			time.Sleep(time.Second)
			continue
		}

		plog.Info("Parachain process block", "curSeq", currSeq,"lastBlockHeight", lastBlock.Height,
			"currSeqMainHeight", lastSeqMainHeight, "currSeqMainHash", common.ToHex(lastSeqMainHash),
			"lastBlockMainHeight", lastBlock.MainHeight, "lastBlockMainHash", common.ToHex(lastBlock.MainHash), "seqTy", mainBlock.Seq.Type)

		if mainBlock.Seq.Type == delAct {
			if len(txs) == 0 {
				if lastSeqMainHeight > lastBlock.MainHeight {
					currSeq++
					continue
				}
				plog.Info("Delete empty block")
			}
			err = client.delLocalBlock(lastBlock.Height)
		} else if mainBlock.Seq.Type == addAct {
			if len(txs) == 0 {
				if lastSeqMainHeight-lastBlock.MainHeight < emptyBlockInterval {
					currSeq++
					continue
				}
				plog.Info("Create empty block")
			}
			err = client.createLocalBlock(lastBlock, txs, mainBlock)

		} else {
			err = types.ErrInvalidParam
		}

		if err != nil{
			plog.Error("para DownloadBlocks", "type",mainBlock.Seq.Type,"err",err.Error())
			time.Sleep(time.Second)
			continue
		}
		currSeq++
	}
}
