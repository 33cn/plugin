// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/common/address"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"

	coinTy "github.com/33cn/plugin/plugin/dapp/coinsx/types"
)

// Exec_Transfer transfer of exec
func (c *Coinsx) Exec_Transfer(transfer *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	from := tx.From()
	to := tx.GetRealToAddr()
	if !checkTransferEnable(c.GetAPI().GetConfig(), c.GetStateDB(), from, to) {
		return nil, errors.Wrapf(types.ErrNotAllow, "transfer is limited from=%s to=%s", from, to)
	}
	//to 是 execs 合约地址
	if drivers.IsDriverAddress(to, c.GetHeight()) {
		return c.GetCoinsAccount().TransferToExec(from, to, transfer.Amount)
	}
	return c.GetCoinsAccount().Transfer(from, to, transfer.Amount)
}

// Exec_TransferToExec the transfer to exec address
func (c *Coinsx) Exec_TransferToExec(transfer *types.AssetsTransferToExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	types.AssertConfig(c.GetAPI())
	cfg := c.GetAPI().GetConfig()
	if !cfg.IsFork(c.GetHeight(), "ForkTransferExec") {
		return nil, types.ErrActionNotSupport
	}
	from := tx.From()
	to := tx.GetRealToAddr()
	//to 是 execs 合约地址
	if !isExecAddrMatch(transfer.ExecName, to) {
		return nil, types.ErrToAddrNotSameToExecAddr
	}
	if !checkTransferEnable(cfg, c.GetStateDB(), from, to) {
		return nil, errors.Wrapf(types.ErrNotAllow, "transfer is limited from=%s to=%s", from, to)
	}
	return c.GetCoinsAccount().TransferToExec(from, to, transfer.Amount)
}

// Exec_Withdraw withdraw exec
func (c *Coinsx) Exec_Withdraw(withdraw *types.AssetsWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	types.AssertConfig(c.GetAPI())
	cfg := c.GetAPI().GetConfig()
	if !cfg.IsFork(c.GetHeight(), "ForkWithdraw") {
		withdraw.ExecName = ""
	}
	from := tx.From()
	to := tx.GetRealToAddr()
	if !checkTransferEnable(cfg, c.GetStateDB(), from, to) {
		return nil, errors.Wrapf(types.ErrNotAllow, "withdraw is limited from=%s to=%s", from, to)
	}
	//to 是 execs 合约地址
	if drivers.IsDriverAddress(tx.GetRealToAddr(), c.GetHeight()) || isExecAddrMatch(withdraw.ExecName, tx.GetRealToAddr()) {
		return c.GetCoinsAccount().TransferWithdraw(from, tx.GetRealToAddr(), withdraw.Amount)
	}
	return nil, types.ErrActionNotSupport
}

// Exec_Genesis genesis of exec
func (c *Coinsx) Exec_Genesis(genesis *types.AssetsGenesis, tx *types.Transaction, index int) (*types.Receipt, error) {
	if c.GetHeight() == 0 {
		if drivers.IsDriverAddress(tx.GetRealToAddr(), c.GetHeight()) {
			return c.GetCoinsAccount().GenesisInitExec(genesis.ReturnAddress, genesis.Amount, tx.GetRealToAddr())
		}
		return c.GetCoinsAccount().GenesisInit(tx.GetRealToAddr(), genesis.Amount)
	}
	return nil, types.ErrReRunGenesis
}

func isExecAddrMatch(name string, to string) bool {
	toaddr := address.ExecAddress(name)
	return toaddr == to
}

// Exec_Config genesis of exec
func (c *Coinsx) Exec_Config(config *coinTy.CoinsConfig, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(c, tx)
	receipt, err := a.config(config, tx, index)
	if err != nil {
		clog.Error("Coins config failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}
