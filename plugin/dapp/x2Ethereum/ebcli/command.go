// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/types"
	"github.com/spf13/cobra"
)

// RelayerCmd command func
func RelayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relayer",
		Short: "relayer of Chain33 and Ethereum ",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		SetPwdCmd(),
		LockCmd(),
		UnlockCmd(),
		Chain33RelayerCmd(),
		EthereumRelayerCmd(),
	)

	return cmd
}

// SetPwdCmd set password
func SetPwdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_pwd",
		Short: "Set password",
		Run:   setPwd,
	}
	addSetPwdFlags(cmd)
	return cmd
}

func addSetPwdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("old", "o", "", "old password")
	cmd.MarkFlagRequired("old")

	cmd.Flags().StringP("new", "n", "", "new password,[8-30]letter and digit")
	cmd.MarkFlagRequired("new")
}

func setPwd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	oldPwd, _ := cmd.Flags().GetString("old")
	newPwd, _ := cmd.Flags().GetString("new")
	params := relayerTypes.ReqSetPasswd{
		OldPassphase: oldPwd,
		NewPassphase: newPwd,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.SetPassphase", params, &res)
	ctx.Run()
}

// LockCmd lock the relayer manager
func LockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Lock relayer manager",
		Run:   lock,
	}
	return cmd
}

func lock(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.Lock", nil, &res)
	ctx.Run()
}

// UnlockCmd unlock the wallet
func UnlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock relayer manager",
		Run:   unLock,
	}
	addUnlockFlags(cmd)
	return cmd
}

func addUnlockFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pwd", "p", "", "password needed to unlock")
	cmd.MarkFlagRequired("pwd")
}

func unLock(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pwd, _ := cmd.Flags().GetString("pwd")

	params := pwd
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.Unlock", params, &res)
	ctx.Run()
}
