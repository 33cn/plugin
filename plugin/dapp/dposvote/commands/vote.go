// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/33cn/chain33/common/crypto"

	vrf "github.com/33cn/chain33/common/vrf/secp256k1"
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/spf13/cobra"
)

var (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
	genFile  = "genesis_file.json"
	pvFile   = "priv_validator_"
)

//DPosCmd DPosVote合约命令行
func DPosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dpos",
		Short: "dpos vote management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		DPosRegistCmd(),
		DPosCancelRegistCmd(),
		DPosVoteCmd(),
		DPosReRegistCmd(),
		DPosVoteCancelCmd(),
		DPosCandidatorQueryCmd(),
		DPosVoteQueryCmd(),
		DPosVrfMRegistCmd(),
		DPosVrfRPRegistCmd(),
		DPosVrfQueryCmd(),
		DPosCreateCmd(),
		DPosVrfVerifyCmd(),
		DPosVrfEvaluateCmd(),
		DPosCBRecordCmd(),
		DPosCBQueryCmd(),
		DPosTopNQueryCmd(),
	)

	return cmd
}

//DPosRegistCmd 构造候选节点注册的命令行
func DPosRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regist",
		Short: "regist a new candidator",
		Run:   regist,
	}
	addRegistFlags(cmd)
	return cmd
}

func addRegistFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().StringP("address", "a", "", "address")
	cmd.MarkFlagRequired("address")

	cmd.Flags().StringP("ip", "i", "", "ip")
	cmd.MarkFlagRequired("address")
}

func regist(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	address, _ := cmd.Flags().GetString("address")
	ip, _ := cmd.Flags().GetString("ip")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"address\":\"%s\", \"IP\":\"%s\"}", pubkey, address, ip)
	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(dty.DPosX, paraName),
		ActionName: dty.CreateRegistTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()

}

//DPosCancelRegistCmd 构造候选节点去注册的命令行
func DPosCancelRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancelRegist",
		Short: "cancel regist for a candidator",
		Run:   cancelRegist,
	}
	addCancelRegistFlags(cmd)
	return cmd
}

func addCancelRegistFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().StringP("address", "a", "", "address")
	cmd.MarkFlagRequired("address")
}

func cancelRegist(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	address, _ := cmd.Flags().GetString("address")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"address\":\"%s\"}", pubkey, address)
	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(dty.DPosX, paraName),
		ActionName: dty.CreateCancelRegistTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosVoteCmd 构造为候选节点投票的命令行
func DPosVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "vote for one candidator",
		Run:   vote,
	}
	addVoteFlags(cmd)
	return cmd
}

func addVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey of a candidator")
	cmd.MarkFlagRequired("pubkey")
	cmd.Flags().Int64P("votes", "v", 0, "votes")
	cmd.MarkFlagRequired("votes")
	cmd.Flags().StringP("addr", "a", "", "address of voter")
	cmd.MarkFlagRequired("addr")
}

func vote(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	votes, _ := cmd.Flags().GetInt64("votes")
	addr, _ := cmd.Flags().GetString("addr")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"votes\":\"%d\", \"fromAddr\":\"%s\"}", pubkey, votes, addr)
	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(dty.DPosX, paraName),
		ActionName: dty.CreateVoteTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosVoteCancelCmd 构造撤销对候选节点投票的命令行
func DPosVoteCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancelVote",
		Short: "cancel votes to a candidator",
		Run:   cancelVote,
	}
	addCancelVoteFlags(cmd)
	return cmd
}

func addCancelVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey of a candidator")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().Int64P("index", "i", 0, "index")
	cmd.MarkFlagRequired("index")
}

func cancelVote(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	index, _ := cmd.Flags().GetInt64("index")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"index\":\"%d\"}", pubkey, index)
	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(dty.DPosX, paraName),
		ActionName: dty.CreateCancelVoteTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosReRegistCmd 构造重新注册候选节点的命令行
func DPosReRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reRegist",
		Short: "re regist a canceled candidator",
		Run:   reRegist,
	}
	addReRegistFlags(cmd)
	return cmd
}

func addReRegistFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().StringP("address", "a", "", "address")
	cmd.MarkFlagRequired("address")

	cmd.Flags().StringP("ip", "i", "", "ip")
	cmd.MarkFlagRequired("address")
}

func reRegist(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	address, _ := cmd.Flags().GetString("address")
	ip, _ := cmd.Flags().GetString("ip")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"address\":\"%s\", \"IP\":\"%s\"}", pubkey, address, ip)
	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(dty.DPosX, paraName),
		ActionName: dty.CreateReRegistTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()

}

//DPosCandidatorQueryCmd 构造查询候选节点信息的命令行
func DPosCandidatorQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidatorQuery",
		Short: "query candidator info",
		Run:   candidatorQuery,
	}
	addCandidatorQueryFlags(cmd)
	return cmd
}

func addCandidatorQueryFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("type", "t", "", "topN/pubkeys")

	cmd.Flags().Int64P("top", "n", 0, "top N by votes")

	cmd.Flags().StringP("pubkeys", "k", "", "pubkeys")
}

func candidatorQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ty, _ := cmd.Flags().GetString("type")
	pubkeys, _ := cmd.Flags().GetString("pubkeys")
	topN, _ := cmd.Flags().GetInt64("top")

	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	switch ty {
	case "topN":
		req := &dty.CandidatorQuery{
			TopN: int32(topN),
		}
		params.FuncName = dty.FuncNameQueryCandidatorByTopN
		params.Payload = types.MustPBToJSON(req)
		var res dty.CandidatorReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "pubkeys":
		keys := strings.Split(pubkeys, ";")
		req := &dty.CandidatorQuery{}
		req.Pubkeys = append(req.Pubkeys, keys...)
		params.FuncName = dty.FuncNameQueryCandidatorByPubkeys
		params.Payload = types.MustPBToJSON(req)
		var res dty.CandidatorReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	}
}

//DPosVoteQueryCmd 构造投票信息查询的命令行
func DPosVoteQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voteQuery",
		Short: "query vote info",
		Run:   voteQuery,
	}
	addVoteQueryFlags(cmd)
	return cmd
}

func addVoteQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkeys", "k", "", "pubkeys")
	cmd.Flags().StringP("address", "a", "", "address")
}

func voteQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkeys, _ := cmd.Flags().GetString("pubkeys")
	addr, _ := cmd.Flags().GetString("address")

	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	req := &dty.DposVoteQuery{
		Addr: addr,
	}

	keys := strings.Split(pubkeys, ";")
	req.Pubkeys = append(req.Pubkeys, keys...)

	params.FuncName = dty.FuncNameQueryVote
	params.Payload = types.MustPBToJSON(req)
	var res dty.DposVoteReply
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()

}

//DPosVrfMRegistCmd 构造注册VRF M信息（输入信息）的命令行
func DPosVrfMRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrfMRegist",
		Short: "regist m of vrf",
		Run:   vrfM,
	}
	addVrfMFlags(cmd)
	return cmd
}

func addVrfMFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().Int64P("cycle", "c", 0, "cycle no. of dpos consensus")
	cmd.MarkFlagRequired("cycle")

	cmd.Flags().StringP("m", "m", "", "input of vrf")
	cmd.MarkFlagRequired("m")
}

func vrfM(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	cycle, _ := cmd.Flags().GetInt64("cycle")
	m, _ := cmd.Flags().GetString("m")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"cycle\":\"%d\", \"m\":\"%s\"}", pubkey, cycle, m)
	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(dty.DPosX, paraName),
		ActionName: dty.CreateRegistVrfMTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosVrfRPRegistCmd 构造VRF R/P(hash及proof)注册的命令行
func DPosVrfRPRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrfRPRegist",
		Short: "regist r,p of vrf",
		Run:   vrfRP,
	}
	addVrfRPRegistFlags(cmd)
	return cmd
}

func addVrfRPRegistFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().Int64P("cycle", "c", 0, "cycle no. of dpos consensus")
	cmd.MarkFlagRequired("cycle")

	cmd.Flags().StringP("hash", "r", "", "hash of vrf")
	cmd.MarkFlagRequired("hash")

	cmd.Flags().StringP("proof", "p", "", "proof of vrf")
	cmd.MarkFlagRequired("proof")
}

func vrfRP(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	cycle, _ := cmd.Flags().GetInt64("cycle")
	hash, _ := cmd.Flags().GetString("hash")
	proof, _ := cmd.Flags().GetString("proof")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"cycle\":\"%d\", \"r\":\"%s\", \"p\":\"%s\"}", pubkey, cycle, hash, proof)
	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(dty.DPosX, paraName),
		ActionName: dty.CreateRegistVrfRPTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosVrfQueryCmd 构造VRF相关信息查询的命令行
func DPosVrfQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrfQuery",
		Short: "query vrf info",
		Run:   vrfQuery,
	}
	addVrfQueryFlags(cmd)
	return cmd
}

func addVrfQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("type", "t", "", "query type:dtime/timestamp/cycle/topN/pubkeys")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringP("time", "d", "", "time like 2019-06-18")
	cmd.Flags().Int64P("timestamp", "s", 0, "time stamp from 1970-1-1")
	cmd.Flags().Int64P("cycle", "c", 0, "cycle,one time point belongs to a cycle")

	cmd.Flags().StringP("pubkeys", "k", "", "pubkeys")
}

func vrfQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ty, _ := cmd.Flags().GetString("type")
	dtime, _ := cmd.Flags().GetString("time")
	timestamp, _ := cmd.Flags().GetInt64("timestamp")
	cycle, _ := cmd.Flags().GetInt64("cycle")
	pubkeys, _ := cmd.Flags().GetString("pubkeys")

	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	switch ty {
	case "dtime":
		t, err := time.Parse("2006-01-02 15:04:05", dtime)
		if err != nil {
			fmt.Println("err time format:", dtime)
			return
		}

		req := &dty.DposVrfQuery{
			Ty:        dty.QueryVrfByTime,
			Timestamp: t.Unix(),
		}

		params.FuncName = dty.FuncNameQueryVrfByTime
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposVrfReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "timestamp":
		if timestamp <= 0 {
			fmt.Println("err timestamp:", timestamp)
			return
		}

		req := &dty.DposVrfQuery{
			Ty:        dty.QueryVrfByTime,
			Timestamp: timestamp,
		}

		params.FuncName = dty.FuncNameQueryVrfByTime
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposVrfReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "cycle":
		if cycle <= 0 {
			fmt.Println("err cycle:", cycle)
			return
		}

		req := &dty.DposVrfQuery{
			Ty:    dty.QueryVrfByCycle,
			Cycle: cycle,
		}

		params.FuncName = dty.FuncNameQueryVrfByCycle
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposVrfReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "topN":
		if cycle <= 0 {
			fmt.Println("err cycle:", cycle)
			return
		}

		req := &dty.DposVrfQuery{
			Ty:    dty.QueryVrfByCycleForTopN,
			Cycle: cycle,
		}

		params.FuncName = dty.FuncNameQueryVrfByCycleForTopN
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposVrfReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "pubkeys":
		if cycle <= 0 {
			fmt.Println("err cycle:", cycle)
			return
		}

		req := &dty.DposVrfQuery{
			Ty:    dty.QueryVrfByCycleForPubkeys,
			Cycle: cycle,
		}

		keys := strings.Split(pubkeys, ";")
		req.Pubkeys = append(req.Pubkeys, keys...)

		params.FuncName = dty.FuncNameQueryVrfByCycleForPubkeys
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposVrfReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	}

}

//DPosCreateCmd to create keyfiles
func DPosCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init_keyfile",
		Short: "Initialize dpos Keyfile",
		Run:   createFiles,
	}
	addCreateCmdFlags(cmd)
	return cmd
}

func addCreateCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("num", "n", "", "Num of the keyfile to create")
	cmd.MarkFlagRequired("num")
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

func initCryptoImpl() error {
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		fmt.Printf("New crypto impl failed err: %v", err)
		return err
	}
	ttypes.ConsensusCrypto = cr
	return nil
}

func createFiles(cmd *cobra.Command, args []string) {
	// init crypto instance
	err := initCryptoImpl()
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
			fmt.Println("Create priv_validator file failed.")
			break
		}

		// create genesis validator by the pubkey of private validator
		gv := ttypes.GenesisValidator{
			PubKey: ttypes.KeyText{Kind: "secp256k1", Data: privValidator.GetPubKey().KeyString()},
			Name:   "",
		}
		genDoc.Validators = append(genDoc.Validators, gv)
	}

	if err := genDoc.SaveAs(genFile); err != nil {
		fmt.Println("Generated genesis file failed.")
		return
	}
	fmt.Printf("Generated genesis file path %v\n", genFile)
}

//DPosVrfVerifyCmd to create keyfiles
func DPosVrfVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrfVerify",
		Short: "vrf verify",
		Run:   verify,
	}
	addVrfVerifyCmdFlags(cmd)
	return cmd
}

func addVrfVerifyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "key")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().StringP("m", "m", "", "input of vrf")
	cmd.MarkFlagRequired("m")

	cmd.Flags().StringP("hash", "r", "", "hash of vrf")
	cmd.MarkFlagRequired("hash")

	cmd.Flags().StringP("proof", "p", "", "proof of vrf")
	cmd.MarkFlagRequired("proof")
}

func verify(cmd *cobra.Command, args []string) {
	// init crypto instance
	err := initCryptoImpl()
	if err != nil {
		return
	}

	key, _ := cmd.Flags().GetString("pubkey")
	data, _ := cmd.Flags().GetString("m")
	hash, _ := cmd.Flags().GetString("hash")
	proof, _ := cmd.Flags().GetString("proof")

	m := []byte(data)

	r, err := hex.DecodeString(hash)
	if err != nil {
		fmt.Println("Error DecodeString vrf hash data failed: ", err)
		return
	}

	p, err := hex.DecodeString(proof)
	if err != nil {
		fmt.Println("Error DecodeString vrf proof data failed: ", err)
		return
	}

	bKey, err := hex.DecodeString(key)
	if err != nil {
		fmt.Println("Error DecodeString bKey data failed: ", err)
		return
	}

	pubKey, err := secp256k1.ParsePubKey(bKey, secp256k1.S256())
	if err != nil {
		fmt.Println("vrf Verify failed: ", err)
		return
	}
	vrfPub := &vrf.PublicKey{PublicKey: (*ecdsa.PublicKey)(pubKey)}
	vrfHash, err := vrfPub.ProofToHash(m, p)
	if err != nil {
		fmt.Println("vrf Verify failed: ", err)
		return
	}

	fmt.Println("vrf m:", data)
	fmt.Println("vrf proof:", proof)
	fmt.Println("vrf hash:", hex.EncodeToString(vrfHash[:]))
	fmt.Println("input hash:", hash)

	if !bytes.Equal(vrfHash[:], r) {
		fmt.Println("vrfVerify failed: invalid VRF hash")
		return
	}

	fmt.Println("vrf hash is same with input hash, vrf Verify succeed")
}

//DPosVrfEvaluateCmd to create keyfiles
func DPosVrfEvaluateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrfEvaluate",
		Short: "vrf evaluate",
		Run:   evaluate,
	}
	addVrfEvaluateCmdFlags(cmd)
	return cmd
}

func addVrfEvaluateCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("privKey", "p", "", "privKey")
	cmd.MarkFlagRequired("privKey")

	cmd.Flags().StringP("m", "m", "", "input of vrf")
	cmd.MarkFlagRequired("m")
}

func evaluate(cmd *cobra.Command, args []string) {
	// init crypto instance
	err := initCryptoImpl()
	if err != nil {
		return
	}

	key, _ := cmd.Flags().GetString("privKey")
	data, _ := cmd.Flags().GetString("m")

	m := []byte(data)

	bKey, err := hex.DecodeString(key)
	if err != nil {
		fmt.Println("Error DecodeString bKey data failed: ", err)
		return
	}

	privKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), bKey)
	vrfPriv := &vrf.PrivateKey{PrivateKey: (*ecdsa.PrivateKey)(privKey)}
	vrfHash, vrfProof := vrfPriv.Evaluate(m)
	fmt.Println("vrf evaluate:")
	fmt.Println("input:", data)
	fmt.Println(fmt.Sprintf("hash:%x", vrfHash))
	fmt.Println(fmt.Sprintf("proof:%x", vrfProof))
}

//DPosCBRecordCmd to create keyfiles
func DPosCBRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cbRecord",
		Short: "record cycle boundary info",
		Run:   recordCB,
	}
	addCBRecordCmdFlags(cmd)
	return cmd
}

func addCBRecordCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("cycle", "c", 0, "cycle")
	cmd.MarkFlagRequired("cycle")
	cmd.Flags().Int64P("height", "m", 0, "height")
	cmd.MarkFlagRequired("height")
	cmd.Flags().StringP("hash", "s", "", "block hash")
	cmd.MarkFlagRequired("hash")
	cmd.Flags().StringP("privKey", "k", "", "private key")
	cmd.MarkFlagRequired("privKey")
}

func recordCB(cmd *cobra.Command, args []string) {
	// init crypto instance
	err := initCryptoImpl()
	if err != nil {
		return
	}
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("privKey")
	cycle, _ := cmd.Flags().GetInt64("cycle")
	height, _ := cmd.Flags().GetInt64("height")
	hash, _ := cmd.Flags().GetString("hash")

	bKey, err := hex.DecodeString(key)
	if err != nil {
		fmt.Println("Error DecodeString bKey data failed: ", err)
		return
	}

	privKey, err := ttypes.ConsensusCrypto.PrivKeyFromBytes(bKey)
	if err != nil {
		fmt.Println("Error PrivKeyFromBytes failed: ", err)
		return
	}

	buf := new(bytes.Buffer)

	canonical := dty.CanonicalOnceCBInfo{
		Cycle:      cycle,
		StopHeight: height,
		StopHash:   hash,
		Pubkey:     hex.EncodeToString(privKey.PubKey().Bytes()),
	}

	byteCB, err := json.Marshal(&canonical)
	if err != nil {
		fmt.Println("Error Marshal failed: ", err)
		return
	}

	_, err = buf.Write(byteCB)
	if err != nil {
		fmt.Println("Error buf.Write failed: ", err)
		return
	}

	signature := privKey.Sign(buf.Bytes())
	sig := hex.EncodeToString(signature.Bytes())

	payload := fmt.Sprintf("{\"cycle\":\"%d\", \"stopHeight\":\"%d\", \"stopHash\":\"%s\", \"pubkey\":\"%s\", \"signature\":\"%s\"}",
		cycle, height, hash, hex.EncodeToString(privKey.PubKey().Bytes()), sig)

	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(dty.DPosX, paraName),
		ActionName: dty.CreateRecordCBTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosCBQueryCmd 查询Cycle Boundary info的命令
func DPosCBQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cbQuery",
		Short: "query cycle boundary info",
		Run:   cbQuery,
	}
	addCBQueryFlags(cmd)
	return cmd
}

func addCBQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("type", "t", "", "query type:cycle/height/hash")
	cmd.MarkFlagRequired("type")

	cmd.Flags().Int64P("cycle", "c", 0, "cycle")
	cmd.Flags().Int64P("height", "m", 0, "height")
	cmd.Flags().StringP("hash", "s", "", "block hash")
}

func cbQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ty, _ := cmd.Flags().GetString("type")
	cycle, _ := cmd.Flags().GetInt64("cycle")
	height, _ := cmd.Flags().GetInt64("height")
	hash, _ := cmd.Flags().GetString("hash")

	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	switch ty {
	case "cycle":
		req := &dty.DposCBQuery{
			Ty:    dty.QueryCBInfoByCycle,
			Cycle: cycle,
		}

		params.FuncName = dty.FuncNameQueryCBInfoByCycle
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposCBReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "height":
		req := &dty.DposCBQuery{
			Ty:         dty.QueryCBInfoByHeight,
			StopHeight: height,
		}

		params.FuncName = dty.FuncNameQueryCBInfoByHeight
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposCBReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "hash":
		req := &dty.DposCBQuery{
			Ty:       dty.QueryCBInfoByHash,
			StopHash: hash,
		}

		params.FuncName = dty.FuncNameQueryCBInfoByHash
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposCBReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	}
}

//DPosTopNQueryCmd 构造TopN相关信息查询的命令行
func DPosTopNQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topNQuery",
		Short: "query topN info",
		Run:   topNQuery,
	}
	addTopNQueryFlags(cmd)
	return cmd
}

func addTopNQueryFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("version", "v", 0, "version")
	cmd.MarkFlagRequired("version")
}

func topNQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	version, _ := cmd.Flags().GetInt64("version")

	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	req := &dty.TopNCandidatorsQuery{
		Version: version,
	}

	params.FuncName = dty.FuncNameQueryTopNByVersion
	params.Payload = types.MustPBToJSON(req)
	var res dty.TopNCandidatorsReply
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}
