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

func (client *client) GetBlockByHeight(height int64) (*types.Block, error) {
	//from blockchain db
	blockDetails, err := client.GetAPI().GetBlocks(&types.ReqBlocks{Start: height, End: height})
	if err != nil {
		plog.Error("GetBlockByHeight fail", "err", err)
		return nil, err
	}
	if 1 != int64(len(blockDetails.Items)) {
		plog.Error("GetBlockByHeight count fail", "len", len(blockDetails.Items))
		return nil, types.ErrInvalidParam
	}
	return blockDetails.Items[0].Block, nil
}

func (client *client) GetBlockHeaders(req *types.ReqBlocks) (*types.Headers, error) {
	//from blockchain db
	headers, err := client.grpcClient.GetHeaders(context.Background(), req)
	if err != nil {
		plog.Error("GetBlockHeaders fail", "err", err)
		return nil, err
	}

	count := req.End - req.Start + 1
	if int64(len(headers.Items)) != count {
		plog.Error("GetBlockHeaders", "start", req.Start, "end", req.End, "reals", headers.Items[0].Height, "reale", headers.Items[len(headers.Items)-1].Height,
			"len", len(headers.Items), "count", count)
		return nil, types.ErrBlockHeightNoMatch
	}
	return headers, nil
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
		plog.Error("Parachain getLastBlockInfo fail", "err", err)
		return nil, err
	}

	return lastBlock, nil
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

	hash := blockSeq.Detail.Block.HashByForkHeight(client.subCfg.MainBlockHashForkHeight)
	if !bytes.Equal(blockSeq.Seq.Hash, hash) {
		plog.Error("para compare ForkBlockHash fail", "forkHeight", client.subCfg.MainBlockHashForkHeight,
			"seqHash", hex.EncodeToString(blockSeq.Seq.Hash), "calcHash", hex.EncodeToString(hash))
		return nil, types.ErrBlockHashNoMatch
	}

	return blockSeq, nil
}

func (client *client) GetParaTxByTitle(req *types.ReqParaTxByTitle) (*types.ParaTxDetails, error) {
	txDetails, err := client.grpcClient.GetParaTxByTitle(context.Background(), req)
	if err != nil {
		plog.Error("GetParaTxByTitle wrong", "err", err.Error(), "start", req.Start, "end", req.End)
		return nil, err
	}

	return txDetails, nil
}

func (client *client) QueryTxOnMainByHash(hash []byte) (*types.TransactionDetail, error) {
	detail, err := client.grpcClient.QueryTransaction(context.Background(), &types.ReqHash{Hash: hash})
	if err != nil {
		plog.Error("QueryTxOnMainByHash Not found", "txhash", common.ToHex(hash))
		return nil, err
	}

	return detail, nil
}

func (client *client) GetParaHeightsByTitle(req *types.ReqHeightByTitle) (*types.ReplyHeightByTitle, error) {
	//from blockchain db
	heights, err := client.grpcClient.LoadParaTxByTitle(context.Background(), req)
	if err != nil {
		plog.Error("GetParaHeightsByTitle fail", "err", err)
		return nil, err
	}

	return heights, nil
}

func (client *client) GetParaTxByHeight(req *types.ReqParaTxByHeight) (*types.ParaTxDetails, error) {
	//from blockchain db
	blocks, err := client.grpcClient.GetParaTxByHeight(context.Background(), req)
	if err != nil {
		plog.Error("GetParaTxByHeight get node status block count fail")
		return nil, err
	}

	//可以小于等于，不能大于
	if len(blocks.Items) > len(req.Items) {
		plog.Error("GetParaTxByHeight get blocks more than req")
		return nil, types.ErrInvalidParam
	}
	return blocks, nil
}
