// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	tokenty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/spf13/cobra"
)

// CreateRawTx 创建交易
// 主链交易使用 Transaction 的 To
// 平行链使用的 payload 的 To， Transaction 的 To 作为执行器的合约地址
// 在执行合约时， 有GetRealTo 用来读取这个项
// 在命令行中， 参数需要加前缀才是对的
// 在平行链上 执行器地址注册不带前缀， withdraw To地址有前缀， 算出来不一致
func CreateRawTx(cmd *cobra.Command, to string, amount float64, note string, isWithdraw bool, tokenSymbol, execName string) (string, error) {
	if amount < 0 {
		return "", types.ErrAmount
	}
	paraName, _ := cmd.Flags().GetString("paraName")
	amountInt64 := int64(math.Trunc((amount+0.0000001)*1e4)) * 1e4
	var tx *types.Transaction
	transfer := &tokenty.TokenAction{}
	if isWithdraw {
		v := &tokenty.TokenAction_Withdraw{Withdraw: &types.AssetsWithdraw{Cointoken: tokenSymbol, Amount: amountInt64,
			Note: []byte(note), To: to, ExecName: getRealExecName(paraName, execName)}}
		transfer.Value = v
		transfer.Ty = tokenty.ActionWithdraw
	} else if execName != "" {
		execName = getRealExecName(paraName, execName)
		if execName != "" && !types.IsAllowExecName([]byte(execName), []byte(execName)) {
			return "", types.ErrExecNameNotMatch
		}
		to, err := GetExecAddr(execName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return "", err
		}
		v := &tokenty.TokenAction_TransferToExec{TransferToExec: &types.AssetsTransferToExec{Cointoken: tokenSymbol,
			Amount: amountInt64, Note: []byte(note), ExecName: execName, To: to}}
		transfer.Value = v
		transfer.Ty = tokenty.TokenActionTransferToExec
	} else {
		v := &tokenty.TokenAction_Transfer{Transfer: &types.AssetsTransfer{Cointoken: tokenSymbol, Amount: amountInt64, Note: []byte(note), To: to}}
		transfer.Value = v
		transfer.Ty = tokenty.ActionTransfer
	}
	execer := []byte(getRealExecName(paraName, "token"))
	if paraName == "" {
		tx = &types.Transaction{Execer: execer, Payload: types.Encode(transfer), To: to}
	} else {
		tx = &types.Transaction{Execer: execer, Payload: types.Encode(transfer), To: address.ExecAddress(string(execer))}
	}
	tx, err := types.FormatTx(string(execer), tx)
	if err != nil {
		return "", err
	}
	txHex := types.Encode(tx)
	return hex.EncodeToString(txHex), nil
}

// GetExecAddr 获取执行器地址
func GetExecAddr(exec string) (string, error) {
	if ok := types.IsAllowExecName([]byte(exec), []byte(exec)); !ok {
		return "", types.ErrExecNameNotAllow
	}

	addrResult := address.ExecAddress(exec)
	result := addrResult
	return result, nil
}

func getRealExecName(paraName string, name string) string {
	if strings.HasPrefix(name, "user.p.") {
		return name
	}
	return paraName + name
}
