// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// CreateEvmCallTx 创建未签名的部署合约交易
func (c *channelClient) CreateDeployTx(ctx context.Context, in evmtypes.EvmContractCreateReq) (*types.UnsignTx, error) {
	amountInt64 := in.Amount
	exec := in.ParaName + evmtypes.ExecutorName
	toAddr := address.ExecAddress(exec)

	bCode, err := common.FromHex(in.Code)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse evm code error", err)
		return nil, err
	}
	if "" != in.Parameter {
		packData, err := evmAbi.PackContructorPara(in.Parameter, in.Abi)
		if err != nil {
			return nil, err
		}

		bCode = append(bCode, packData...)
	}

	action := evmtypes.EVMContractAction{
		Amount:       uint64(amountInt64),
		GasLimit:     0,
		GasPrice:     0,
		Code:         bCode,
		Para:         nil,
		Alias:        in.Alias,
		Note:         in.Note,
		ContractAddr: toAddr,
	}

	cfg := c.GetConfig()
	tx := &types.Transaction{Execer: []byte(exec), Payload: types.Encode(&action), Fee: 0, To: toAddr}

	tx.Fee, _ = tx.GetRealFee(cfg.GetMinTxFeeRate())
	if tx.Fee < in.Fee {
		tx.Fee += in.Fee
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	tx.ChainID = cfg.GetChainID()
	txHex := types.Encode(tx)

	return &types.UnsignTx{Data: txHex}, nil
}

func (c *channelClient) CreateCallTx(ctx context.Context, in evmtypes.EvmContractCallReq) (*types.UnsignTx, error) {
	amountInt64 := in.Amount
	feeInt64 := in.Fee
	exec := in.ParaName + evmtypes.ExecutorName
	toAddr := address.ExecAddress(exec)

	_, packedParameter, err := evmAbi.Pack(in.Parameter, in.Abi, false)
	if err != nil {
		return nil, err
	}

	action := evmtypes.EVMContractAction{
		Amount:       uint64(amountInt64),
		GasLimit:     0,
		GasPrice:     0,
		Code:         nil,
		Para:         packedParameter,
		Alias:        "",
		Note:         in.Note,
		ContractAddr: in.ContractAddr,
	}

	tx := &types.Transaction{Execer: []byte(exec), Payload: types.Encode(&action), Fee: 0, To: toAddr}

	cfg := c.GetConfig()
	tx.Fee, _ = tx.GetRealFee(cfg.GetMinTxFeeRate())
	if tx.Fee < feeInt64 {
		tx.Fee += feeInt64
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	tx.ChainID = cfg.GetChainID()
	txHex := types.Encode(tx)

	return &types.UnsignTx{Data: txHex}, nil
}

func (c *channelClient) CreateTransferOnlyTx(ctx context.Context, in evmtypes.EvmTransferOnlyReq) (*types.UnsignTx, error) {
	exec := in.ParaName + evmtypes.ExecutorName
	toAddr := address.ExecAddress(exec)

	r_addr, err := address.NewBtcAddress(in.To)
	if nil != err {
		return nil, err
	}

	action := evmtypes.EVMContractAction{
		Amount:       uint64(in.Amount),
		GasLimit:     0,
		GasPrice:     0,
		Code:         nil,
		Para:         r_addr.Hash160[:],
		Alias:        "",
		Note:         in.Note,
		ContractAddr: toAddr,
	}
	tx := &types.Transaction{Execer: []byte(exec), Payload: types.Encode(&action), Fee: 0, To: toAddr}

	cfg := c.GetConfig()
	tx.Fee, _ = tx.GetRealFee(cfg.GetMinTxFeeRate())
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	tx.ChainID = cfg.GetChainID()
	txHex := types.Encode(tx)

	return &types.UnsignTx{Data: txHex}, nil
}
