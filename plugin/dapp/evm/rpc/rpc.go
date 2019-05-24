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
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// CreateEvmCallTx 创建未签名的调用evm交易
func (c *channelClient) Create(ctx context.Context, in evmtypes.EvmContractCreateReq) (*types.UnsignTx, error) {
	bCode, err := common.FromHex(in.Code)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse evm code error", err)
		return nil, err
	}

	action := evmtypes.EVMContractAction{Amount: 0, Code: bCode, GasLimit: 0, GasPrice: 0, Note: in.Note, Abi: in.Abi}

	execer := types.ExecName(in.ParaName + "evm")
	addr := address.ExecAddress(types.ExecName(in.ParaName + "evm"))
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(&action), Fee: 0, To: addr}

	tx.Fee, _ = tx.GetRealFee(types.GInt("MinFee"))
	if tx.Fee < in.Fee {
		tx.Fee += in.Fee
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()

	txHex := types.Encode(tx)

	return &types.UnsignTx{Data: txHex}, nil
}

func (c *channelClient) Call(ctx context.Context, in evmtypes.EvmContractCallReq) (*types.UnsignTx, error) {
	amountInt64 := in.Amount * 1e4 * 1e4
	feeInt64 := in.Fee * 1e4 * 1e4
	toAddr := address.ExecAddress(in.Exec)

	bCode, err := common.FromHex(in.Code)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse evm code error", err)
		return nil, err
	}

	action := evmtypes.EVMContractAction{Amount: amountInt64, Code: bCode, GasLimit: 0, GasPrice: 0, Note: in.Note, Abi: in.Abi}

	tx := &types.Transaction{Execer: []byte(in.Exec), Payload: types.Encode(&action), Fee: 0, To: toAddr}

	tx.Fee, _ = tx.GetRealFee(types.GInt("MinFee"))
	if tx.Fee < feeInt64 {
		tx.Fee += feeInt64
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()

	txHex := types.Encode(tx)

	return &types.UnsignTx{Data: txHex}, nil
}

func (c *channelClient) Transfer(ctx context.Context, in evmtypes.EvmContractTransferReq, isWithdraw bool) (*types.UnsignTx, error) {
	var tx *types.Transaction
	transfer := &cty.CoinsAction{}
	amountInt64 := int64(in.Amount*1e4) * 1e4
	execName := in.Exec

	if isWithdraw {
		transfer.Value = &cty.CoinsAction_Withdraw{Withdraw: &types.AssetsWithdraw{Amount: amountInt64, ExecName: execName, To: address.ExecAddress(execName)}}
		transfer.Ty = cty.CoinsActionWithdraw
	} else {
		transfer.Value = &cty.CoinsAction_TransferToExec{TransferToExec: &types.AssetsTransferToExec{Amount: amountInt64, ExecName: execName, To: address.ExecAddress(execName)}}
		transfer.Ty = cty.CoinsActionTransferToExec
	}
	if in.ParaName == "" {
		tx = &types.Transaction{Execer: []byte(types.ExecName(in.ParaName + "coins")), Payload: types.Encode(transfer), To: address.ExecAddress(execName)}
	} else {
		tx = &types.Transaction{Execer: []byte(types.ExecName(in.ParaName + "coins")), Payload: types.Encode(transfer), To: address.ExecAddress(types.ExecName(in.ParaName + "coins"))}
	}

	var err error
	tx.Fee, err = tx.GetRealFee(types.GInt("MinFee"))
	if err != nil {
		return nil, err
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()

	txHex := types.Encode(tx)

	return &types.UnsignTx{Data: txHex}, nil
}
