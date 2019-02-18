// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/spf13/cobra"
)

// TicketCmd ticket command type
func TicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket",
		Short: "Ticket management",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		BindMinerCmd(),
		CountTicketCmd(),
		CloseTicketCmd(),
		GetColdAddrByMinerCmd(),
	)

	return cmd
}

// BindMinerCmd bind miner
func BindMinerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind_miner",
		Short: "Bind private key to miner address",
		Run:   bindMiner,
	}
	addBindMinerFlags(cmd)
	return cmd
}

func addBindMinerFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("bind_addr", "b", "", "miner address")
	cmd.MarkFlagRequired("bind_addr")

	cmd.Flags().StringP("origin_addr", "o", "", "origin address")
	cmd.MarkFlagRequired("origin_addr")
}

func bindMiner(cmd *cobra.Command, args []string) {
	bindAddr, _ := cmd.Flags().GetString("bind_addr")
	originAddr, _ := cmd.Flags().GetString("origin_addr")
	//c, _ := crypto.New(types.GetSignName(wallet.SignType))
	//a, _ := common.FromHex(key)
	//privKey, _ := c.PrivKeyFromBytes(a)
	//originAddr := account.PubKeyToAddress(privKey.PubKey().Bytes()).String()
	ta := &ty.TicketAction{}
	tBind := &ty.TicketBind{
		MinerAddress:  bindAddr,
		ReturnAddress: originAddr,
	}
	ta.Value = &ty.TicketAction_Tbind{Tbind: tBind}
	ta.Ty = ty.TicketActionBind

	tx, err := types.CreateFormatTx("ticket", types.Encode(ta))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	txHex := types.Encode(tx)
	fmt.Println(hex.EncodeToString(txHex))
}

// CountTicketCmd get ticket count
func CountTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Get ticket count",
		Run:   countTicket,
	}
	return cmd
}

func countTicket(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res int64
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "ticket.GetTicketCount", nil, &res)
	ctx.Run()
}

// CloseTicketCmd close all accessible tickets
func CloseTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close tickets",
		Run:   closeTicket,
	}
	addCloseBindAddr(cmd)
	return cmd
}

func addCloseBindAddr(cmd *cobra.Command) {
	cmd.Flags().StringP("miner_addr", "m", "", "miner address (optional)")
}

func closeTicket(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	bindAddr, _ := cmd.Flags().GetString("miner_addr")
	status, err := getWalletStatus(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	isAutoMining := status.(rpctypes.WalletStatus).IsAutoMining
	if isAutoMining {
		fmt.Fprintln(os.Stderr, types.ErrMinerNotClosed)
		return
	}

	tClose := &ty.TicketClose{
		MinerAddress: bindAddr,
	}

	var res rpctypes.ReplyHashes
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = rpc.Call("ticket.CloseTickets", tClose, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if len(res.Hashes) == 0 {
		fmt.Println("no ticket to be close")
		return
	}

	data, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(string(data))
}

func getWalletStatus(rpcAddr string) (interface{}, error) {
	rpc, err := jsonclient.NewJSONClient(rpcAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	var res rpctypes.WalletStatus
	err = rpc.Call("Chain33.GetWalletStatus", nil, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	return res, nil
}

// GetColdAddrByMinerCmd get cold address by miner
func GetColdAddrByMinerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cold",
		Short: "Get cold wallet address of miner",
		Run:   coldAddressOfMiner,
	}
	addColdAddressOfMinerFlags(cmd)
	return cmd
}

func addColdAddressOfMinerFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("miner", "m", "", "miner address")
	cmd.MarkFlagRequired("miner")
}

func coldAddressOfMiner(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("miner")
	reqaddr := &types.ReqString{
		Data: addr,
	}
	var params rpctypes.Query4Jrpc
	params.Execer = "ticket"
	params.FuncName = "MinerSourceList"
	params.Payload = types.MustPBToJSON(reqaddr)

	var res types.ReplyStrings
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}
