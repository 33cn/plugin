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
	"golang.org/x/net/context"
	"strings"
)

func bindMiner(cfg *types.Chain33Config, param *ty.ReqBindMiner) (*ty.ReplyBindMiner, error) {
	tBind := &ty.TicketBind{
		MinerAddress:  param.BindAddr,
		ReturnAddress: param.OriginAddr,
	}
	data, err := types.CallCreateTx(cfg, cfg.ExecName(ty.TicketX), "Tbind", tBind)
	if err != nil {
		return nil, err
	}
	hex := common.ToHex(data)
	return &ty.ReplyBindMiner{TxHex: hex}, nil
}

// CreateBindMiner 创建绑定挖矿
func (g *channelClient) CreateBindMiner(ctx context.Context, in *ty.ReqBindMiner) (*ty.ReplyBindMiner, error) {
	//调整十六进制地址大小写转换,如果不进行大小写转换，在TicketBind 进行action.fromaddr != tbind.ReturnAddress 地址校验的时候会返回ErrFromAddr
	if common.IsHex(in.OriginAddr) {
		in.OriginAddr = strings.ToLower(in.OriginAddr)
	}

	if in.BindAddr != "" {
		err := address.CheckAddress(in.BindAddr, -1)
		if err != nil {
			return nil, err
		}
	}
	err := address.CheckAddress(in.OriginAddr, -1)
	if err != nil {
		return nil, err
	}

	cfg := g.GetConfig()
	if in.CheckBalance {
		header, err := g.GetLastHeader()
		if err != nil {
			return nil, err
		}
		price := ty.GetTicketMinerParam(cfg, header.Height).TicketPrice
		if price == 0 {
			return nil, types.ErrInvalidParam
		}
		if in.Amount%ty.GetTicketMinerParam(cfg, header.Height).TicketPrice != 0 || in.Amount < 0 {
			return nil, types.ErrAmount
		}

		getBalance := &types.ReqBalance{Addresses: []string{in.OriginAddr}, Execer: cfg.GetCoinExec(), AssetSymbol: "bty", AssetExec: cfg.GetCoinExec()}
		balances, err := g.GetCoinsAccountDB().GetBalance(g, getBalance)
		if err != nil {
			return nil, err
		}
		if len(balances) == 0 {
			return nil, types.ErrInvalidParam
		}
		if balances[0].Balance < in.Amount+2*cfg.GetCoinPrecision() {
			return nil, types.ErrNoBalance
		}
	}
	return bindMiner(cfg, in)
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

// GetTicketList get ticket list info
func (g *channelClient) GetTicketList(ctx context.Context, in *types.ReqNil) ([]*ty.Ticket, error) {
	inn := *in
	data, err := g.ExecWalletFunc(ty.TicketX, "WalletGetTickets", &inn)
	if err != nil {
		return nil, err
	}

	return data.(*ty.ReplyWalletTickets).Tickets, nil
}

// GetTicketList get ticket list info
func (c *Jrpc) GetTicketList(in *types.ReqNil, result *interface{}) error {
	resp, err := c.cli.GetTicketList(context.Background(), &types.ReqNil{})
	if err != nil {
		return err
	}
	*result = resp
	return nil

}
