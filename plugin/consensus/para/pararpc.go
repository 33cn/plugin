// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"bytes"
	"context"
	"encoding/hex"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
)

func (client *client) GetBlockedSeq(hash []byte) (int64, error) {
	//from blockchain db
	blockedSeq, err := client.GetAPI().GetMainSequenceByHash(&types.ReqHash{Hash: hash})
	if err != nil {
		return -2, err
	}
	return blockedSeq.Data, nil

}

func (client *client) GetBlockByHeight(height int64) (*types.Block, error) {
	//from blockchain db
	blockDetails, err := client.GetAPI().GetBlocks(&types.ReqBlocks{Start: height, End: height})
	if err != nil {
		plog.Error("paracommitmsg get node status block count fail")
		return nil, err
	}
	if 1 != int64(len(blockDetails.Items)) {
		plog.Error("paracommitmsg get node status block count fail")
		return nil, types.ErrInvalidParam
	}
	return blockDetails.Items[0].Block, nil
}

// 获取当前平行链block对应主链seq，hash信息
// 对于云端主链节点，创世区块记录seq在不同主链节点上差异很大，通过记录的主链hash获取真实seq使用
func (client *client) getLastBlockMainInfo() (int64, *types.Block, error) {
	lastBlock, err := client.getLastBlockInfo()
	if err != nil {
		return -2, nil, err
	}
	//如果在云端节点获取不到对应MainHash，切换到switchLocalHashMatchedBlock 去循环查找
	mainSeq, err := client.GetSeqByHashOnMainChain(lastBlock.MainHash)
	if err != nil {
		return 0, lastBlock, nil
	}
	return mainSeq, lastBlock, nil
}

func (client *client) getLastBlockInfo() (*types.Block, error) {
	lastBlock, err := client.RequestLastBlock()
	if err != nil {
		plog.Error("Parachain RequestLastBlock fail", "err", err)
		return nil, err
	}

	return lastBlock, nil
}

func (client *client) GetForkHeightOnMainChain(key string) (int64, error) {
	ret, err := client.grpcClient.GetFork(context.Background(), &types.ReqKey{Key: []byte(key)})
	if err != nil {
		plog.Error("para get rpc ForkHeight fail", "key", key, "err", err.Error())
		return types.MaxHeight, err
	}

	return ret.Data, nil
}

func (client *client) GetLastHeightOnMainChain() (int64, error) {
	header, err := client.grpcClient.GetLastHeader(context.Background(), &types.ReqNil{})
	if err != nil {
		plog.Error("GetLastHeightOnMainChain", "Error", err.Error())
		return -1, err
	}
	return header.Height, nil
}

func (client *client) GetLastSeqOnMainChain() (int64, error) {
	seq, err := client.grpcClient.GetLastBlockSequence(context.Background(), &types.ReqNil{})
	if err != nil {
		plog.Error("GetLastSeqOnMainChain", "Error", err.Error())
		return -1, err
	}
	//the reflect checked in grpcHandle
	return seq.Data, nil
}

func (client *client) GetSeqByHeightOnMainChain(height int64) (int64, []byte, error) {
	hash, err := client.GetHashByHeightOnMainChain(height)
	if err != nil {
		return -1, nil, err
	}
	seq, err := client.GetSeqByHashOnMainChain(hash)
	return seq, hash, err
}

func (client *client) GetHashByHeightOnMainChain(height int64) ([]byte, error) {
	reply, err := client.grpcClient.GetBlockHash(context.Background(), &types.ReqInt{Height: height})
	if err != nil {
		plog.Error("GetHashByHeightOnMainChain", "Error", err.Error())
		return nil, err
	}
	return reply.Hash, nil
}

func (client *client) GetSeqByHashOnMainChain(hash []byte) (int64, error) {
	seq, err := client.grpcClient.GetSequenceByHash(context.Background(), &types.ReqHash{Hash: hash})
	if err != nil {
		plog.Error("GetSeqByHashOnMainChain", "Error", err.Error(), "hash", hex.EncodeToString(hash))
		return -1, err
	}
	//the reflect checked in grpcHandle
	return seq.Data, nil
}

func (client *client) GetBlockOnMainBySeq(seq int64) (*types.BlockSeq, error) {
	blockSeq, err := client.grpcClient.GetBlockBySeq(context.Background(), &types.Int64{Data: seq})
	if err != nil {
		plog.Error("Not found block on main", "seq", seq)
		return nil, err
	}

	hash := blockSeq.Detail.Block.HashByForkHeight(mainBlockHashForkHeight)
	if !bytes.Equal(blockSeq.Seq.Hash, hash) {
		plog.Error("para compare ForkBlockHash fail", "forkHeight", mainBlockHashForkHeight,
			"seqHash", hex.EncodeToString(blockSeq.Seq.Hash), "calcHash", hex.EncodeToString(hash))
		return nil, types.ErrBlockHashNoMatch
	}

	return blockSeq, nil
}

func (client *client) GetBlockOnMainByHash(hash []byte) (*types.Block, error) {
	blocks, err := client.grpcClient.GetBlockByHashes(context.Background(), &types.ReqHashes{Hashes: [][]byte{hash}})
	if err != nil || blocks.Items[0] == nil {
		plog.Error("GetBlockOnMainByHash Not found", "blockhash", common.ToHex(hash))
		return nil, err
	}

	return blocks.Items[0].Block, nil
}

func (client *client) QueryTxOnMainByHash(hash []byte) (*types.TransactionDetail, error) {
	detail, err := client.grpcClient.QueryTransaction(context.Background(), &types.ReqHash{Hash: hash})
	if err != nil {
		plog.Error("QueryTxOnMainByHash Not found", "txhash", common.ToHex(hash))
		return nil, err
	}

	return detail, nil
}
