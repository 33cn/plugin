// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	vt "github.com/33cn/plugin/plugin/dapp/valnode/types"
	"github.com/spf13/cobra"
)

// ValCmd valnode cmd register
func ValCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "valnode",
		Short: "Construct valnode transactions",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		IsSyncCmd(),
		GetBlockInfoCmd(),
		GetNodeInfoCmd(),
		AddNodeCmd(),
	)
	return cmd
}

// IsSyncCmd query tendermint is sync
func IsSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is_sync",
		Short: "Query tendermint consensus is sync",
		Run:   isSync,
	}
	return cmd
}

func isSync(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res bool
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "valnode.IsSync", nil, &res)
	ctx.Run()
}

// GetNodeInfoCmd get validator nodes
func GetNodeInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Get tendermint validator nodes",
		Run:   getNodeInfo,
	}
	return cmd
}

func getNodeInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "valnode.GetNodeInfo", nil, &res)
	ctx.SetResultCb(parseNodeInfo)
	ctx.Run()
}

func parseNodeInfo(arg interface{}) (interface{}, error) {
	var result vt.ValidatorSet
	res := arg.(*string)
	data, err := hex.DecodeString(*res)
	if err != nil {
		return nil, err
	}
	err = types.Decode(data, &result)
	if err != nil {
		return nil, err
	}
	return result.Validators, nil
}

// GetBlockInfoCmd get block info
func GetBlockInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get tendermint consensus info",
		Run:   getBlockInfo,
	}
	addGetBlockInfoFlags(cmd)
	return cmd
}

func addGetBlockInfoFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("height", "t", 0, "block height (larger than 0)")
	cmd.MarkFlagRequired("height")
}

func getBlockInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	height, _ := cmd.Flags().GetInt64("height")
	req := &vt.ReqBlockInfo{
		Height: height,
	}
	params := rpctypes.Query4Jrpc{
		Execer:   vt.ValNodeX,
		FuncName: "GetBlockInfoByHeight",
		Payload:  types.MustPBToJSON(req),
	}

	var res vt.TendermintBlockInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// AddNodeCmd add validator node
func AddNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add tendermint validator node",
		Run:   addNode,
	}
	addNodeFlags(cmd)
	return cmd
}

func addNodeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "p", "", "public key")
	cmd.MarkFlagRequired("pubkey")
	cmd.Flags().Int64P("power", "w", 0, "voting power")
	cmd.MarkFlagRequired("power")
}

func addNode(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	power, _ := cmd.Flags().GetInt64("power")

	pubkeybyte, err := hex.DecodeString(pubkey)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	privkey, err := getprivkey()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	value := &vt.ValNodeAction_Node{Node: &vt.ValNode{PubKey: pubkeybyte, Power: power}}
	action := &vt.ValNodeAction{Value: value, Ty: vt.ValNodeActionUpdate}
	tx := &types.Transaction{Execer: []byte(vt.ValNodeX), Payload: types.Encode(action), Fee: 0}
	err = tx.SetRealFee(types.GInt("MinFee"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	tx.To = address.ExecAddress(vt.ValNodeX)
	tx.Sign(types.SECP256K1, privkey)

	txHex := types.Encode(tx)
	data := hex.EncodeToString(txHex)
	params := rpctypes.RawParm{
		Data: data,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.SendTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func getprivkey() (crypto.PrivKey, error) {
	key := "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		return nil, err
	}
	bkey, err := common.FromHex(key)
	if err != nil {
		return nil, err
	}
	priv, err := cr.PrivKeyFromBytes(bkey)
	if err != nil {
		return nil, err
	}
	return priv, nil
}
