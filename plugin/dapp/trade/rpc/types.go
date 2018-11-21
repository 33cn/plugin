// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/trade/types"
)

type channelClient struct {
	rpctypes.ChannelClient
}

//Jrpc : Jrpc struct definition
type Jrpc struct {
	cli *channelClient
}

//Grpc : Grpc struct definition
type Grpc struct {
	*channelClient
}

//Init : do the init operation
func Init(name string, s rpctypes.RPCServer) {
	cli := &channelClient{}
	grpc := &Grpc{channelClient: cli}
	cli.Init(name, s, &Jrpc{cli: cli}, grpc)
	ptypes.RegisterTradeServer(s.GRPC(), grpc)
}

//GetLastMemPool : get the last memory pool
func (jrpc *Jrpc) GetLastMemPool(in types.ReqNil, result *interface{}) error {
	reply, err := jrpc.cli.GetLastMempool()
	if err != nil {
		return err
	}

	{
		var txlist rpctypes.ReplyTxList
		txs := reply.GetTxs()
		for _, tx := range txs {
			tran, err := rpctypes.DecodeTx(tx)
			if err != nil {
				continue
			}
			txlist.Txs = append(txlist.Txs, tran)
		}
		*result = &txlist
	}
	return nil
}
