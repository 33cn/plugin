// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/33cn/chain33/common"
	"math/rand"
	"os"
	"strings"
	"time"

	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/pkg/errors"

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
		AutoMineCmd(),
		BindMinerCmd(),
		CountTicketCmd(),
		CloseTicketCmd(),
		GetColdAddrByMinerCmd(),
		listTicketCmd(),
		CreateCloseTicketCmd(),
	)

	return cmd
}

// AutoMineCmd  set auto mining
func AutoMineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auto_mine",
		Short: "Set auto mine on/off",
		Run:   autoMine,
	}
	addAutoMineFlags(cmd)
	return cmd
}

func addAutoMineFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("flag", "f", 0, `auto mine(0: off, 1: on)`)
	cmd.MarkFlagRequired("flag")
}

func autoMine(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	flag, _ := cmd.Flags().GetInt32("flag")
	if flag != 0 && flag != 1 {
		err := cmd.UsageFunc()(cmd)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	params := struct {
		Flag int32
	}{
		Flag: flag,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "ticket.SetAutoMining", params, &res)
	ctx.Run()
}

// BindMinerCmd bind miner
func BindMinerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind",
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
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	bindAddr, _ := cmd.Flags().GetString("bind_addr")
	originAddr, _ := cmd.Flags().GetString("origin_addr")
	//c, _ := crypto.Load(types.GetSignName(wallet.SignType))
	//a, _ := common.FromHex(key)
	//privKey, _ := c.PrivKeyFromBytes(a)
	//originAddr := account.PubKeyToAddress(privKey.PubKey().Bytes()).String()
	if common.IsHex(originAddr) {
		originAddr = strings.ToLower(originAddr)
	}
	ta := &ty.TicketAction{}
	tBind := &ty.TicketBind{
		MinerAddress:  bindAddr,
		ReturnAddress: originAddr,
	}
	ta.Value = &ty.TicketAction_Tbind{Tbind: tBind}
	ta.Ty = ty.TicketActionBind

	rawTx := &types.Transaction{Payload: types.Encode(ta)}
	tx, err := types.FormatTxExt(cfg.ChainID, len(paraName) > 0, cfg.MinTxFeeRate, ty.TicketX, rawTx)
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

// listTicketCmd get ticket count
func listTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get ticket id list",
		Run:   listTicket,
	}
	cmd.Flags().StringP("miner_acct", "m", "", "miner address (optional)")
	cmd.Flags().Int32P("status", "s", 1, "ticket status (default 1:opened tickets)")
	cmd.Flags().StringP("return_addr", "r", "", "return address")
	return cmd
}

func listTicket(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	minerAddr, _ := cmd.Flags().GetString("miner_acct")
	status, _ := cmd.Flags().GetInt32("status")
	returnAddr, _ := cmd.Flags().GetString("return_addr")
	if minerAddr != "" {
		var params rpctypes.Query4Jrpc

		params.Execer = ty.TicketX
		params.FuncName = "TicketList"
		req := ty.TicketList{Addr: minerAddr, Status: status}
		params.Payload = types.MustPBToJSON(&req)
		if returnAddr == "" {
			var res ty.ReplyTicketList
			ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
			return
		}

		var res ty.ReplyTicketList
		rpc, err := jsonclient.NewJSONClient(rpcLaddr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		err = rpc.Call("Chain33.Query", params, &res)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		returnTickets := make([]*ty.Ticket, 0)
		for _, v := range res.Tickets {
			if v.ReturnAddress == returnAddr {
				returnTickets = append(returnTickets, v)
			}
		}
		res.Tickets = returnTickets
		data, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		fmt.Println(string(data))
		return
	}

	var res []ty.Ticket
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = rpc.Call("ticket.GetTicketList", nil, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if returnAddr != "" {
		returnTickets := make([]ty.Ticket, 0)
		for _, v := range res {
			if v.ReturnAddress == returnAddr {
				returnTickets = append(returnTickets, v)
			}
		}
		res = returnTickets
	}

	data, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(data))
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
		fmt.Println("no ticket to be close or close fail,to check log")
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

// CreateCloseTicketCmd create close all tickets in status (2,3)
func CreateCloseTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_close",
		Short: "Create Close tickets transaction",
		Run:   createCloseTicket,
	}
	addCreateCloseTicket(cmd)
	return cmd
}

func addCreateCloseTicket(cmd *cobra.Command) {
	cmd.Flags().StringP("miner_addr", "m", "", "miner address")
	cmd.MarkFlagRequired("miner_addr")
	cmd.Flags().StringP("return_addr", "r", "", "return address (optional)")
	cmd.Flags().Int32P("status", "s", 1, "ticket status (default 1:opened tickets 2: mined tickets)")
	cmd.Flags().Int64P("withdraw_time", "w", 172800, "ticketWithdrawTime")
	cmd.Flags().Int64P("miner_wait_time", "i", 7200, "ticketMinerWaitTime")

}

func createCloseTicket(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	returnAddr, _ := cmd.Flags().GetString("return_addr")
	minerAddr, _ := cmd.Flags().GetString("miner_addr")
	withdrawTime, _ := cmd.Flags().GetInt64("withdraw_time")
	minerWaitTime, _ := cmd.Flags().GetInt64("miner_wait_time")
	status, _ := cmd.Flags().GetInt32("status")

	now := time.Now().Unix()

	if status != ty.TicketOpened && status != ty.TicketMined {
		fmt.Fprintln(os.Stderr, errors.New("status must be set 1 (TicketOpened) or 2 (TicketMined)"))
		return
	}

	var params rpctypes.Query4Jrpc

	params.Execer = ty.TicketX
	params.FuncName = "TicketList"
	req := ty.TicketList{Addr: minerAddr, Status: status}
	params.Payload = types.MustPBToJSON(&req)

	var res ty.ReplyTicketList
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = rpc.Call("Chain33.Query", params, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	signList := make([]string, 0)
	for _, v := range res.Tickets {
		if len(signList) == 200 {
			break
		}
		if returnAddr != "" && returnAddr != v.ReturnAddress {
			continue
		}
		if v.Status == ty.TicketOpened {
			if v.CreateTime+withdrawTime < now {
				signList = append(signList, v.TicketId)
			}
		} else if v.Status == ty.TicketMined {
			if v.CreateTime+withdrawTime < now && v.MinerTime+minerWaitTime < now {
				signList = append(signList, v.TicketId)
			}
		}
	}
	if len(signList) == 0 {
		fmt.Println("no tickerIds to close")
		return
	}
	tx := createTicketCloseTx(signList)
	fee, err := tx.GetRealFee(100000)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	tx.Fee = fee
	fmt.Println(hex.EncodeToString(types.Encode(tx)))
}

func createTicketCloseTx(ids []string) *types.Transaction {
	var transaction types.Transaction
	nonce := rand.Int63()
	tclose := &ty.TicketClose{TicketId: ids}
	action := &ty.TicketAction{
		Ty:    ty.TicketActionClose,
		Value: &ty.TicketAction_Tclose{Tclose: tclose},
	}
	transaction.Payload = types.Encode(action)
	transaction.Execer = []byte("ticket")
	transaction.Fee = 0
	transaction.Nonce = nonce
	transaction.Expire = time.Now().Unix() + 300
	transaction.To = "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	return &transaction
}
