// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/33cn/chain33/common/address"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/qbft/types"
	vt "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
	"github.com/spf13/cobra"
)

var (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
	genFile  = "genesis_file.json"
	pvFile   = "priv_validator_"
	//AuthBLS ...
	AuthBLS = 259
)

// ValCmd qbftNode cmd register
func ValCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "qbft",
		Short: "Construct qbft transactions",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		IsSyncCmd(),
		GetBlockInfoCmd(),
		GetNodeInfoCmd(),
		GetCurrentStateCmd(),
		GetPerfStatCmd(),
		AddNodeCmd(),
		CreateCmd(),
	)
	return cmd
}

// IsSyncCmd query qbft is sync
func IsSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is_sync",
		Short: "Query qbft consensus is sync",
		Run:   isSync,
	}
	return cmd
}

func isSync(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res bool
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "qbftNode.IsSync", nil, &res)
	ctx.Run()
}

// GetNodeInfoCmd get validator nodes
func GetNodeInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Get qbft validator nodes",
		Run:   getNodeInfo,
	}
	return cmd
}

func getNodeInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res *vt.QbftNodeInfoSet
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "qbftNode.GetNodeInfo", nil, &res)
	ctx.Run()
}

// GetBlockInfoCmd get block info
func GetBlockInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get qbft consensus info",
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
	req := &vt.ReqQbftBlockInfo{
		Height: height,
	}
	params := rpctypes.Query4Jrpc{
		Execer:   vt.QbftNodeX,
		FuncName: "GetBlockInfoByHeight",
		Payload:  types.MustPBToJSON(req),
	}

	var res vt.QbftBlockInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCb(jsonOutput)
	result, err := ctx.RunResult()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(result)
}

func jsonOutput(arg interface{}) (interface{}, error) {
	res := arg.(types.Message)
	data, err := types.PBToJSON(res)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, data, "", "    ")
	if err != nil {
		return nil, err
	}
	return buf.String(), nil
}

// GetCurrentStateCmd get current consensus state
func GetCurrentStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Get qbft current consensus state",
		Run:   getCurrentState,
	}
	return cmd
}

func getCurrentState(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	params := rpctypes.Query4Jrpc{
		Execer:   vt.QbftNodeX,
		FuncName: "GetCurrentState",
	}

	var res vt.QbftState
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCb(jsonOutput)
	result, err := ctx.RunResult()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(result)
}

// GetPerfStatCmd get block info
func GetPerfStatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perf_stat",
		Short: "Get qbft performance statistics",
		Run:   getPerfStat,
	}
	addGetPerfStatFlags(cmd)
	return cmd
}

func addGetPerfStatFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("start", "s", 0, "start block height")
	cmd.Flags().Int64P("end", "e", 0, "end block height")
}

func getPerfStat(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	start, _ := cmd.Flags().GetInt64("start")
	end, _ := cmd.Flags().GetInt64("end")
	req := &vt.ReqQbftPerfStat{
		Start: start,
		End:   end,
	}
	params := rpctypes.Query4Jrpc{
		Execer:   vt.QbftNodeX,
		FuncName: "GetPerfStat",
		Payload:  types.MustPBToJSON(req),
	}

	var res vt.QbftPerfStat
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// AddNodeCmd add validator node
func AddNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add qbft validator node",
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
	pubkey, _ := cmd.Flags().GetString("pubkey")
	power, _ := cmd.Flags().GetInt64("power")

	value := &vt.QbftNodeAction_Node{Node: &vt.QbftNode{PubKey: pubkey, Power: power}}
	action := &vt.QbftNodeAction{Value: value, Ty: vt.QbftNodeActionUpdate}
	tx := &types.Transaction{
		Payload: types.Encode(action),
		Nonce:   rand.Int63(),
		Execer:  []byte(vt.QbftNodeX),
	}
	tx.To = address.ExecAddress(string(tx.Execer))

	txHex := types.Encode(tx)
	fmt.Println(hex.EncodeToString(txHex))
}

//CreateCmd to create keyfiles
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen_file",
		Short: "Generate qbft genesis and priv file",
		Run:   createFiles,
	}
	addCreateCmdFlags(cmd)
	return cmd
}

func addCreateCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("num", "n", "", "num of the keyfile to create")
	cmd.MarkFlagRequired("num")
	cmd.Flags().StringP("type", "t", "ed25519", "sign type of the keyfile (secp256k1, ed25519, sm2, bls)")
}

// RandStr ...
func RandStr(length int) string {
	chars := []byte{}
MAIN_LOOP:
	for {
		val := rand.Int63()
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 {         // only 62 characters in strChars
				val >>= 6
				continue
			} else {
				chars = append(chars, strChars[v])
				if len(chars) == length {
					break MAIN_LOOP
				}
				val >>= 6
			}
		}
	}

	return string(chars)
}

func initCryptoImpl(signType int) error {
	ttypes.CryptoName = types.GetSignName("", signType)
	cr, err := crypto.Load(ttypes.CryptoName, -1)
	if err != nil {
		fmt.Printf("init crypto fail: %v", err)
		return err
	}
	ttypes.ConsensusCrypto = cr
	return nil
}

func createFiles(cmd *cobra.Command, args []string) {
	// init crypto instance
	ty, _ := cmd.Flags().GetString("type")
	signType, ok := ttypes.SignMap[ty]
	if !ok {
		fmt.Println("type parameter is not valid")
		return
	}
	err := initCryptoImpl(signType)
	if err != nil {
		return
	}

	// genesis file
	genDoc := ttypes.GenesisDoc{
		ChainID:     fmt.Sprintf("chain33-%v", RandStr(6)),
		GenesisTime: time.Now(),
	}

	num, _ := cmd.Flags().GetString("num")
	n, err := strconv.Atoi(num)
	if err != nil {
		fmt.Println("num parameter is not valid digit")
		return
	}
	for i := 0; i < n; i++ {
		// create private validator file
		pvFileName := pvFile + strconv.Itoa(i) + ".json"
		privValidator := ttypes.LoadOrGenPrivValidatorFS(pvFileName)
		if privValidator == nil {
			fmt.Println("create priv_validator file fail")
			break
		}

		// create genesis validator by the pubkey of private validator
		gv := ttypes.GenesisValidator{
			PubKey: ttypes.KeyText{Kind: ttypes.CryptoName, Data: privValidator.GetPubKey().KeyString()},
			Power:  10,
		}
		genDoc.Validators = append(genDoc.Validators, gv)
	}

	if err := genDoc.SaveAs(genFile); err != nil {
		fmt.Println("generate genesis file fail")
		return
	}
	fmt.Printf("generate genesis file path %v\n", genFile)
}
