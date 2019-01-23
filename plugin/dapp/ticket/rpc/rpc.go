// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	context "golang.org/x/net/context"
)

func bindMiner(param *ty.ReqBindMiner) (*ty.ReplyBindMiner, error) {
	tBind := &ty.TicketBind{
		MinerAddress:  param.BindAddr,
		ReturnAddress: param.OriginAddr,
	}
	data, err := types.CallCreateTx(types.ExecName(ty.TicketX), "Tbind", tBind)
	if err != nil {
		return nil, err
	}
	hex := common.ToHex(data)
	return &ty.ReplyBindMiner{TxHex: hex}, nil
}

// CreateBindMiner 创建绑定挖矿
func (g *channelClient) CreateBindMiner(ctx context.Context, in *ty.ReqBindMiner) (*ty.ReplyBindMiner, error) {
	if in.Amount%(10000*types.Coin) != 0 || in.Amount < 0 {
		return nil, types.ErrAmount
	}
	err := address.CheckAddress(in.BindAddr)
	if err != nil {
		return nil, err
	}
	err = address.CheckAddress(in.OriginAddr)
	if err != nil {
		return nil, err
	}

	if in.CheckBalance {
		getBalance := &types.ReqBalance{Addresses: []string{in.OriginAddr}, Execer: "coins"}
		balances, err := g.GetCoinsAccountDB().GetBalance(g, getBalance)
		if err != nil {
			return nil, err
		}
		if len(balances) == 0 {
			return nil, types.ErrInvalidParam
		}
		if balances[0].Balance < in.Amount+2*types.Coin {
			return nil, types.ErrNoBalance
		}
	}
	return bindMiner(in)
}

// SetAutoMining set auto mining
func (g *channelClient) SetAutoMining(ctx context.Context, in *ty.MinerFlag) (*types.Reply, error) {
	data, err := g.ExecWalletFunc(ty.TicketX, "WalletAutoMiner", in)
	if err != nil {
		return nil, err
	}
	return data.(*types.Reply), nil
}

// GetTicketCount get count
func (g *channelClient) GetTicketCount(ctx context.Context, in *types.ReqNil) (*types.Int64, error) {
	data, err := g.QueryConsensusFunc(ty.TicketX, "GetTicketCount", &types.ReqNil{})
	if err != nil {
		return nil, err
	}
	return data.(*types.Int64), nil
}

// CloseTickets close ticket
func (g *channelClient) CloseTickets(ctx context.Context, in *ty.TicketClose) (*types.ReplyHashes, error) {
	inn := *in
	data, err := g.ExecWalletFunc(ty.TicketX, "CloseTickets", &inn)
	if err != nil {
		return nil, err
	}
	return data.(*types.ReplyHashes), nil
}

// CreateBindMiner create bind miner
func (c *Jrpc) CreateBindMiner(in *ty.ReqBindMiner, result *interface{}) error {
	reply, err := c.cli.CreateBindMiner(context.Background(), in)
	if err != nil {
		return err
	}
	*result = reply
	return nil
}

// GetTicketCount get ticket count
func (c *Jrpc) GetTicketCount(in *types.ReqNil, result *int64) error {
	resp, err := c.cli.GetTicketCount(context.Background(), &types.ReqNil{})
	if err != nil {
		return err
	}
	*result = resp.GetData()
	return nil

}

// CloseTickets close ticket
func (c *Jrpc) CloseTickets(in *ty.TicketClose, result *interface{}) error {
	resp, err := c.cli.CloseTickets(context.Background(), in)
	if err != nil {
		return err
	}
	var hashes rpctypes.ReplyHashes
	for _, has := range resp.Hashes {
		hashes.Hashes = append(hashes.Hashes, common.ToHex(has))
	}
	*result = &hashes
	return nil
}

// SetAutoMining set auto mining
func (c *Jrpc) SetAutoMining(in *ty.MinerFlag, result *rpctypes.Reply) error {
	resp, err := c.cli.SetAutoMining(context.Background(), in)
	if err != nil {
		return err
	}
	var reply rpctypes.Reply
	reply.IsOk = resp.GetIsOk()
	reply.Msg = string(resp.GetMsg())
	*result = reply
	return nil
}
