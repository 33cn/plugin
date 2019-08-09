// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"time"

	"encoding/hex"

	"bytes"

	"sync/atomic"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	paraexec "github.com/33cn/plugin/plugin/dapp/paracross/executor"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

var (
	fetchFilterParaTxsEnable bool
)

func (client *client) addLocalBlock(height int64, block *pt.ParaLocalDbBlock) error {
	set := &types.LocalDBSet{}

	key := calcTitleHeightKey(types.GetTitle(), height)
	kv := &types.KeyValue{Key: key, Value: types.Encode(block)}
	set.KV = append(set.KV, kv)

	//两个key原子操作
	key = calcTitleLastHeightKey(types.GetTitle())
	kv = &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: height})}
	set.KV = append(set.KV, kv)

	return client.setLocalDb(set)
}

func (client *client) createLocalBlock(lastBlock *pt.ParaLocalDbBlock, txs []*types.Transaction, mainBlock *types.ParaTxDetail) error {
	var newblock pt.ParaLocalDbBlock

	newblock.Height = lastBlock.Height + 1
	newblock.MainHash = mainBlock.Header.Hash
	newblock.MainHeight = mainBlock.Header.Height
	newblock.ParentMainHash = lastBlock.MainHash
	newblock.BlockTime = mainBlock.Header.BlockTime

	newblock.Txs = txs

	err := client.addLocalBlock(newblock.Height, &newblock)
	if err != nil {
		return err
	}
	client.commitMsgClient.commitTxCheckNotify(mainBlock.TxDetails)
	return err
}

func (client *client) createLocalGenesisBlock(genesis *types.Block) error {
	return client.alignLocalBlock2ChainBlock(genesis)
}

func (client *client) delLocalBlock(height int64) error {
	set := &types.LocalDBSet{}
	key := calcTitleHeightKey(types.GetTitle(), height)
	kv := &types.KeyValue{Key: key, Value: nil}
	set.KV = append(set.KV, kv)

	//两个key原子操作
	key = calcTitleLastHeightKey(types.GetTitle())
	kv = &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: height - 1})}
	set.KV = append(set.KV, kv)

	return client.setLocalDb(set)
}

// localblock 设置到当前高度，当前高度后面block会被新的区块覆盖
func (client *client) removeLocalBlocks(curHeight int64) error {
	set := &types.LocalDBSet{}

	key := calcTitleLastHeightKey(types.GetTitle())
	kv := &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: curHeight})}
	set.KV = append(set.KV, kv)

	return client.setLocalDb(set)
}

func (client *client) getLastLocalHeight() (int64, error) {
	key := calcTitleLastHeightKey(types.GetTitle())
	set := &types.LocalDBGet{Keys: [][]byte{key}}
	value, err := client.getLocalDb(set, len(set.Keys))
	if err != nil {
		return -1, err
	}
	if len(value) == 0 {
		return -1, types.ErrNotFound
	}

	height := &types.Int64{}
	err = types.Decode(value[0], height)
	if err != nil {
		return -1, err
	}
	return height.Data, nil

}

func (client *client) getLocalBlockByHeight(height int64) (*pt.ParaLocalDbBlock, error) {
	key := calcTitleHeightKey(types.GetTitle(), height)
	set := &types.LocalDBGet{Keys: [][]byte{key}}

	value, err := client.getLocalDb(set, len(set.Keys))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, types.ErrNotFound
	}

	var block pt.ParaLocalDbBlock
	err = types.Decode(value[0], &block)
	if err != nil {
		return nil, err
	}
	return &block, nil

}

func (client *client) getLocalBlockSeq(height int64) (int64, []byte, error) {
	lastBlock, err := client.getLocalBlockByHeight(height)
	if err != nil {
		return -2, nil, err
	}

	//如果当前mainHash对应seq获取不到，返回0 seq，和当前hash，去switchLocalHashMatchedBlock里面回溯查找
	mainSeq, err := client.GetSeqByHashOnMainChain(lastBlock.MainHash)
	if err != nil {
		return 0, lastBlock.MainHash, nil
	}
	return mainSeq, lastBlock.MainHash, nil

}

//根据匹配上的chainblock，设置当前localdb block
func (client *client) alignLocalBlock2ChainBlock(chainBlock *types.Block) error {
	localBlock := &pt.ParaLocalDbBlock{
		Height:     chainBlock.Height,
		MainHeight: chainBlock.MainHeight,
		MainHash:   chainBlock.MainHash,
		BlockTime:  chainBlock.BlockTime,
	}

	return client.addLocalBlock(localBlock.Height, localBlock)

}

//如果localdb里面没有信息，就从chain block返回，至少有创世区块，然后进入循环匹配切换场景
func (client *client) getLastLocalBlockSeq() (int64, []byte, error) {
	height, err := client.getLastLocalHeight()
	if err == nil {
		mainSeq, mainHash, err := client.getLocalBlockSeq(height)
		if err == nil {
			return mainSeq, mainHash, nil
		}
	}

	plog.Info("Parachain getLastLocalBlockSeq from block")
	//说明localDb获取存在错误，从chain获取
	mainSeq, chainBlock, err := client.getLastBlockMainInfo()
	if err != nil {
		return -2, nil, err
	}

	//chain block中获取成功，设置last local block和找到的chainBlock main高度和mainhash对齐
	err = client.alignLocalBlock2ChainBlock(chainBlock)
	if err != nil {
		return -2, nil, err
	}
	return mainSeq, chainBlock.MainHash, nil

}

func (client *client) getLastLocalBlock() (*pt.ParaLocalDbBlock, error) {
	height, err := client.getLastLocalHeight()
	if err != nil {
		return nil, err
	}

	return client.getLocalBlockByHeight(height)
}

//genesis block scenario
func (client *client) syncFromGenesisBlock() (int64, *types.Block, error) {
	lastSeq, lastBlock, err := client.getLastBlockMainInfo()
	if err != nil {
		plog.Error("Parachain getLastBlockInfo fail", "err", err)
		return -2, nil, err
	}
	plog.Info("syncFromGenesisBlock sync from height 0")
	return lastSeq, lastBlock, nil
}

func (client *client) getMatchedBlockOnChain(startHeight int64) (int64, *types.Block, error) {
	lastBlock, err := client.RequestLastBlock()
	if err != nil {
		plog.Error("Parachain RequestLastBlock fail", "err", err)
		return -2, nil, err
	}

	if lastBlock.Height == 0 {
		return client.syncFromGenesisBlock()
	}

	if startHeight == 0 || startHeight > lastBlock.Height {
		startHeight = lastBlock.Height
	}

	depth := client.subCfg.SearchHashMatchedBlockDepth
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
					"search depth", client.subCfg.SearchHashMatchedBlockDepth)
				panic("search HashMatchedBlock overflow, re-setting search depth and restart to try")
			}
			if height == 1 {
				plog.Error("switchHashMatchedBlock search to height=1 not found", "lastBlockHeight", lastBlock.Height,
					"height1 mainHash", hex.EncodeToString(block.MainHash))
				return client.syncFromGenesisBlock()

			}
			continue
		}

		plog.Info("getMatchedBlockOnChain succ", "currHeight", height, "initHeight", lastBlock.Height,
			"new currSeq", mainSeq, "new preMainBlockHash", hex.EncodeToString(block.MainHash))
		return mainSeq, block, nil
	}
	return -2, nil, pt.ErrParaCurHashNotMatch
}

func (client *client) switchMatchedBlockOnChain(startHeight int64) (int64, []byte, error) {
	mainSeq, chainBlock, err := client.getMatchedBlockOnChain(startHeight)
	if err != nil {
		return -2, nil, err
	}
	//chain block中获取成功，设置last local block和找到的chainBlock main高度和mainhash对齐
	err = client.alignLocalBlock2ChainBlock(chainBlock)
	if err != nil {
		return -2, nil, err
	}
	return mainSeq, chainBlock.MainHash, nil
}

func (client *client) switchHashMatchedBlock() (int64, []byte, error) {
	mainSeq, mainHash, err := client.switchLocalHashMatchedBlock()
	if err != nil {
		return client.switchMatchedBlockOnChain(0)
	}
	return mainSeq, mainHash, nil
}

//
func (client *client) switchLocalHashMatchedBlock() (int64, []byte, error) {
	lastBlock, err := client.getLastLocalBlock()
	if err != nil {
		plog.Error("Parachain RequestLastBlock fail", "err", err)
		return -2, nil, err
	}

	for height := lastBlock.Height; height >= 0; height-- {
		block, err := client.getLocalBlockByHeight(height)
		if err != nil {
			return -2, nil, err
		}
		//当前block结构已经有mainHash和MainHeight但是从blockchain获取的block还没有写入，以后如果获取到，可以替换从minerTx获取
		plog.Info("switchLocalHashMatchedBlock", "height", height, "mainHeight", block.MainHeight, "mainHash", hex.EncodeToString(block.MainHash))
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
	return -2, nil, pt.ErrParaCurHashNotMatch
}

func (client *client) getBatchSeqCount(currSeq int64) (int64, error) {
	lastSeq, err := client.GetLastSeqOnMainChain()
	if err != nil {
		return 0, err
	}

	if lastSeq > currSeq {
		if lastSeq-currSeq > client.subCfg.EmptyBlockInterval {
			atomic.StoreInt32(&client.caughtUp, 0)
		} else {
			atomic.StoreInt32(&client.caughtUp, 1)
		}
		if fetchFilterParaTxsEnable && lastSeq-currSeq > client.subCfg.BatchFetchBlockCount {
			return client.subCfg.BatchFetchBlockCount, nil
		}
		return 0, nil
	}

	if lastSeq == currSeq {
		return 0, nil
	}

	// lastSeq = currSeq -1
	if lastSeq+1 == currSeq {
		plog.Debug("Waiting new sequence from main chain")
		return 0, pt.ErrParaWaitingNewSeq
	}

	// lastSeq < currSeq-1
	return 0, pt.ErrParaCurHashNotMatch

}

func verifyMainBlockHash(preMainBlockHash []byte, mainBlock *types.ParaTxDetail) error {
	if (bytes.Equal(preMainBlockHash, mainBlock.Header.ParentHash) && mainBlock.Type == addAct) ||
		(bytes.Equal(preMainBlockHash, mainBlock.Header.Hash) && mainBlock.Type == delAct) {
		return nil
	}
	plog.Error("verifyMainBlockHash", "preMainBlockHash", hex.EncodeToString(preMainBlockHash),
		"mainParentHash", hex.EncodeToString(mainBlock.Header.ParentHash), "mainHash", hex.EncodeToString(mainBlock.Header.Hash),
		"type", mainBlock.Type, "height", mainBlock.Header.Height)
	return pt.ErrParaCurHashNotMatch
}

func verifyMainBlocks(preMainBlockHash []byte, mainBlocks *types.ParaTxDetails) error {
	pre := preMainBlockHash
	for _, block := range mainBlocks.Items {
		err := verifyMainBlockHash(pre, block)
		if err != nil {
			return err
		}
		if block.Type == addAct {
			pre = block.Header.Hash
		} else {
			pre = block.Header.ParentHash
		}

	}
	return nil
}

func (client *client) requestTxsFromBlock(currSeq int64, preMainBlockHash []byte) (*types.ParaTxDetails, error) {
	blockSeq, err := client.GetBlockOnMainBySeq(currSeq)
	if err != nil {
		return nil, err
	}

	txDetail := blockSeq.Detail.FilterParaTxsByTitle(types.GetTitle())
	txDetail.Type = blockSeq.Seq.Type

	err = verifyMainBlockHash(preMainBlockHash, txDetail)
	if err != nil {
		plog.Error("requestTxsFromBlock", "curr seq", currSeq, "preMainBlockHash", hex.EncodeToString(preMainBlockHash))
		return nil, err
	}
	return &types.ParaTxDetails{Items: []*types.ParaTxDetail{txDetail}}, nil
}

func (client *client) requestFilterParaTxs(currSeq int64, count int64, preMainBlockHash []byte) (*types.ParaTxDetails, error) {
	req := &types.ReqParaTxByTitle{Start: currSeq, End: currSeq + count, Title: types.GetTitle()}
	details, err := client.GetParaTxByTitle(req)
	if err != nil {
		return nil, err
	}

	err = verifyMainBlocks(preMainBlockHash, details)
	if err != nil {
		plog.Error("requestTxsOnlyPara", "curSeq", currSeq, "count", count, "preMainBlockHash", hex.EncodeToString(preMainBlockHash))
		return nil, err
	}
	return details, nil
}

func (client *client) RequestTx(currSeq int64, count int64, preMainBlockHash []byte) (*types.ParaTxDetails, error) {
	if fetchFilterParaTxsEnable {
		return client.requestFilterParaTxs(currSeq, count, preMainBlockHash)
	}

	return client.requestTxsFromBlock(currSeq, preMainBlockHash)
}

func (client *client) processHashNotMatchError(currSeq int64, lastSeqMainHash []byte, err error) (int64, []byte, error) {
	if err == pt.ErrParaCurHashNotMatch {
		preSeq, preSeqMainHash, err := client.switchHashMatchedBlock()
		if err == nil {
			return preSeq + 1, preSeqMainHash, nil
		}
	}
	return currSeq, lastSeqMainHash, err
}

func (client *client) procLocalBlock(mainBlock *types.ParaTxDetail) (bool, error) {
	lastSeqMainHeight := mainBlock.Header.Height

	lastBlock, err := client.getLastLocalBlock()
	if err != nil {
		plog.Error("Parachain getLastLocalBlock", "err", err)
		return false, err
	}

	txs := paraexec.FilterTxsForPara(mainBlock)

	plog.Info("Parachain process block", "lastBlockHeight", lastBlock.Height, "lastBlockMainHeight", lastBlock.MainHeight,
		"lastBlockMainHash", common.ToHex(lastBlock.MainHash), "currMainHeight", lastSeqMainHeight,
		"curMainHash", common.ToHex(mainBlock.Header.Hash), "seqTy", mainBlock.Type)

	if mainBlock.Type == delAct {
		if len(txs) == 0 {
			if lastSeqMainHeight > lastBlock.MainHeight {
				return false, nil
			}
			plog.Info("Delete empty block")
		}
		return true, client.delLocalBlock(lastBlock.Height)

	} else if mainBlock.Type == addAct {
		if len(txs) == 0 {
			if lastSeqMainHeight-lastBlock.MainHeight < client.subCfg.EmptyBlockInterval {
				return false, nil
			}
			plog.Info("Create empty block")
		}
		return true, client.createLocalBlock(lastBlock, txs, mainBlock)

	}
	return false, types.ErrInvalidParam

}

func (client *client) procLocalBlocks(mainBlocks *types.ParaTxDetails) error {
	var notify bool
	for _, main := range mainBlocks.Items {
		changed, err := client.procLocalBlock(main)
		if err != nil {
			return err
		}
		if changed {
			notify = true
		}
	}
	if notify {
		client.blockSyncClient.notifyLocalChange()
	}

	return nil
}

func (client *client) CreateBlock() {
	lastSeq, lastSeqMainHash, err := client.getLastLocalBlockSeq()
	if err != nil {
		plog.Error("Parachain CreateBlock getLastLocalBlockSeq fail", "err", err.Error())
		return
	}
	currSeq := lastSeq + 1

out:
	for {
		select {
		case <-client.quitCreate:
			break out
		default:
			count, err := client.getBatchSeqCount(currSeq)
			if err != nil {
				currSeq, lastSeqMainHash, err = client.processHashNotMatchError(currSeq, lastSeqMainHash, err)
				if err == nil {
					continue
				}
				time.Sleep(time.Second * time.Duration(client.subCfg.WriteBlockSeconds))
				continue
			}

			plog.Debug("Parachain CreateBlock", "curSeq", currSeq, "count", count, "lastSeqMainHash", common.ToHex(lastSeqMainHash))
			paraTxs, err := client.RequestTx(currSeq, count, lastSeqMainHash)
			if err != nil {
				currSeq, lastSeqMainHash, err = client.processHashNotMatchError(currSeq, lastSeqMainHash, err)
				continue
			}

			if count+1 != int64(len(paraTxs.Items)) {
				plog.Error("para CreateBlock count not match", "count", count+1, "items", len(paraTxs.Items))
				continue
			}

			err = client.procLocalBlocks(paraTxs)
			if err != nil {
				//根据localblock，重新搜索匹配
				lastSeqMainHash = nil
				plog.Error("para CreateBlock.procLocalBlocks", "err", err.Error())
				continue
			}

			//重新设定seq和lastSeqMainHash
			lastSeqMainHash = paraTxs.Items[count].Header.Hash
			if paraTxs.Items[count].Type == delAct {
				lastSeqMainHash = paraTxs.Items[count].Header.ParentHash
			}
			currSeq = currSeq + count + 1

		}
	}

	plog.Info("para CreateBlock quit")
	client.wg.Done()
}
