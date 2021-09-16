// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/33cn/chain33/rpc/jsonclient"
	"github.com/33cn/chain33/types"

	rpctypes "github.com/33cn/chain33/rpc/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/spf13/cobra"
)

//ParcCmd paracross cmd register
func MixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mix",
		Short: "Construct mix coin transactions",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		CreateDepositRawTxCmd(),
		CreateTransferRawTxCmd(),
		CreateWithdrawRawTxCmd(),
		CreateAuthRawTxCmd(),

		CreateConfigCmd(),

		QueryCmd(),
		WalletCmd(),
	)
	return cmd
}

func getRealExecName(paraName string, name string) string {
	if strings.HasPrefix(name, "user.p.") {
		return name
	}
	return paraName + name
}

// CreateDepositCmd create raw asset transfer tx
func CreateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Proof parameters config to mix coin contract",
	}
	cmd.AddCommand(mixConfigVerifyKeyParaCmd())
	//cmd.AddCommand(mixConfigAuthPubKeyParaCmd())
	cmd.AddCommand(mixConfigPaymentPubKeyParaCmd())

	return cmd
}

func mixConfigVerifyKeyParaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vk",
		Short: "zk proof verify key config cmd",
		Run:   createConfigVerify,
	}
	addVkConfigFlags(cmd)

	return cmd
}

func addVkConfigFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("circuit", "c", 0, "mix circuit type,0:deposit,1:withdraw,2:tansferinput,3:transferoutput,4:authorize")
	cmd.MarkFlagRequired("circuit")

	cmd.Flags().StringP("zkey", "z", "", "zk proof verify key")
	cmd.MarkFlagRequired("zkey")

}

func createConfigVerify(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	circuit, _ := cmd.Flags().GetUint32("circuit")
	key, _ := cmd.Flags().GetString("zkey")

	var zkVk mixTy.ZkVerifyKey
	zkVk.Value = key
	zkVk.Type = mixTy.VerifyType(circuit)

	payload := &mixTy.MixConfigAction{}
	payload.Ty = mixTy.MixConfigType_Verify
	payload.Action = mixTy.MixConfigAct(0)
	payload.Value = &mixTy.MixConfigAction_VerifyKey{VerifyKey: &zkVk}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, mixTy.MixX),
		ActionName: "Config",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func mixConfigAuthPubKeyParaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "mix authorize pub key config cmd",
		Run:   createConfigPubKey,
	}
	addPubKeyConfigFlags(cmd)

	return cmd
}

func addPubKeyConfigFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("action", "t", 0, "0:add,1:delete")

	cmd.Flags().StringP("key", "a", "", "authorize pub key")
	cmd.MarkFlagRequired("key")

}

func createConfigPubKey(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	action, _ := cmd.Flags().GetUint32("action")
	key, _ := cmd.Flags().GetString("key")

	//var pubkey mixTy.AuthorizePubKey
	//pubkey.Value = key

	payload := &mixTy.MixConfigAction{}
	payload.Ty = mixTy.MixConfigType_Auth
	payload.Action = mixTy.MixConfigAct(action)
	payload.Value = &mixTy.MixConfigAction_AuthKey{AuthKey: key}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, mixTy.MixX),
		ActionName: "Config",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func mixConfigPaymentPubKeyParaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "receiver key register cmd",
		Run:   createConfigPayPubKey,
	}
	addPayPubKeyConfigFlags(cmd)

	return cmd
}

func addPayPubKeyConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "register for addr ")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().StringP("receiver", "r", "", "receiver key")
	cmd.MarkFlagRequired("receiver")

	cmd.Flags().StringP("encryptKey", "e", "", "encrypt key for secret")
	cmd.MarkFlagRequired("encryptKey")

}

func createConfigPayPubKey(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	addr, _ := cmd.Flags().GetString("addr")
	receiver, _ := cmd.Flags().GetString("receiver")
	encryptKey, _ := cmd.Flags().GetString("encryptKey")

	payload := &mixTy.MixConfigAction{}
	payload.Ty = mixTy.MixConfigType_Payment

	payload.Value = &mixTy.MixConfigAction_PaymentKey{PaymentKey: &mixTy.PaymentKey{Addr: addr, ReceiverKey: receiver, EncryptKey: encryptKey}}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, mixTy.MixX),
		ActionName: "Config",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query cmd",
	}
	cmd.AddCommand(GetTreePathCmd())
	cmd.AddCommand(GetTreeLeavesCmd())
	cmd.AddCommand(GetTreeRootsCmd())
	cmd.AddCommand(ShowMixTxsCmd())
	cmd.AddCommand(ShowPaymentPubKeyCmd())
	return cmd
}

// GetParaInfoCmd get para chain status by height
func GetTreePathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Get leaf tree path",
		Run:   treePath,
	}
	addGetPathCmdFlags(cmd)
	return cmd
}

func addGetPathCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("root", "r", "", "tree root hash, null allowed")

	cmd.Flags().StringP("leaf", "l", "", "leaf hash")
	cmd.MarkFlagRequired("leaf")

}

func treePath(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	root, _ := cmd.Flags().GetString("root")
	leaf, _ := cmd.Flags().GetString("leaf")

	var params rpctypes.Query4Jrpc
	params.Execer = mixTy.MixX
	params.FuncName = "GetTreePath"
	req := mixTy.TreeInfoReq{
		RootHash: root,
		LeafHash: leaf,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res mixTy.CommitTreeProve
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetParaInfoCmd get para chain status by height
func GetTreeLeavesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leaves",
		Short: "Get tree leaves",
		Run:   treeLeaves,
	}
	addGetLeavesCmdFlags(cmd)
	return cmd
}

func addGetLeavesCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("root", "r", "", "tree root hash, null means current leaves")

}

func treeLeaves(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	root, _ := cmd.Flags().GetString("root")

	var params rpctypes.Query4Jrpc
	params.Execer = mixTy.MixX
	params.FuncName = "Query_GetLeavesList"
	req := mixTy.TreeInfoReq{
		RootHash: root,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res mixTy.TreeListResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetParaInfoCmd get para chain status by height
func GetTreeRootsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roots",
		Short: "Get archive roots",
		Run:   treeRoot,
	}
	//addGetPathCmdFlags(cmd)
	return cmd
}

func treeRoot(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc
	params.Execer = mixTy.MixX
	params.FuncName = "GetRootList"

	params.Payload = types.MustPBToJSON(&types.ReqNil{})

	var res mixTy.TreeListResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// ShowProposalBoardCmd 显示提案查询信息
func ShowMixTxsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "show mix txs info",
		Run:   showMixTxs,
	}
	addShowMixTxsflags(cmd)
	return cmd
}

func addShowMixTxsflags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("type", "y", 0, "type(0:query by hash; 1:list)")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringP("hash", "s", "", "mix tx hash")

	cmd.Flags().Int64P("height", "t", -1, "height, default is -1")
	cmd.Flags().Int64P("index", "i", -1, "index, default is -1")

	cmd.Flags().Int32P("count", "c", 1, "count, default is 1")
	cmd.Flags().Int32P("direction", "d", 0, "direction, default is reserve")

}

func showMixTxs(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	typ, _ := cmd.Flags().GetUint32("type")
	hash, _ := cmd.Flags().GetString("hash")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	height, _ := cmd.Flags().GetInt64("height")
	index, _ := cmd.Flags().GetInt64("index")

	var params rpctypes.Query4Jrpc

	params.Execer = mixTy.MixX
	var req *mixTy.MixTxListReq
	if typ < 1 {
		req = &mixTy.MixTxListReq{
			Count:     count,
			Direction: direction,
			Hash:      hash,
		}
	} else {
		req = &mixTy.MixTxListReq{
			Count:     count,
			Direction: direction,
			Height:    height,
			Index:     index,
		}
	}

	params.FuncName = "ListMixTxs"
	params.Payload = types.MustPBToJSON(req)

	var resp mixTy.MixTxListResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

// ShowProposalBoardCmd 显示提案查询信息
func ShowPaymentPubKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peer",
		Short: "get peer addr receive key info",
		Run:   showPayment,
	}
	addShowPaymentflags(cmd)
	return cmd
}

func addShowPaymentflags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account addr")
	cmd.MarkFlagRequired("addr")

}

func showPayment(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	var params rpctypes.Query4Jrpc

	params.Execer = mixTy.MixX

	params.FuncName = "PaymentPubKey"
	params.Payload = types.MustPBToJSON(&types.ReqString{Data: addr})

	var resp mixTy.PaymentKey
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func WalletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "wallet query cmd",
	}
	cmd.AddCommand(ShowAccountPrivacyInfo())
	cmd.AddCommand(ShowAccountNoteInfo())
	cmd.AddCommand(RescanCmd())
	cmd.AddCommand(RescanStatusCmd())
	cmd.AddCommand(EnableCmd())
	cmd.AddCommand(SecretCmd())

	return cmd
}

// ShowAccountPrivacyInfo get para chain status by height
func ShowAccountPrivacyInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "get account privacy keys for mix note",
		Run:   accountPrivacy,
	}
	accountPrivacyCmdFlags(cmd)
	return cmd
}

func accountPrivacyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "user wallet addr")

	cmd.Flags().StringP("priv", "p", "", "user wallet addr's privacy key,option")

	cmd.Flags().Uint32P("detail", "d", 0, "if get keys' privacy keys,option")

}

func accountPrivacy(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	priv, _ := cmd.Flags().GetString("priv")
	addr, _ := cmd.Flags().GetString("addr")
	detail, _ := cmd.Flags().GetUint32("detail")

	if len(priv) == 0 && len(addr) == 0 {
		fmt.Println("err: one of addr or priv should be fill")
		return
	}

	var res mixTy.WalletAddrPrivacy
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.ShowAccountPrivacyInfo", &mixTy.PaymentKeysReq{PrivKey: priv, Addr: addr, Detail: int32(detail)}, &res)
	ctx.Run()
}

// ShowAccountPrivacyInfo get para chain status by height
func ShowAccountNoteInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notes",
		Short: "show account notes",
		Run:   accountNote,
	}
	accountNoteCmdFlags(cmd)
	return cmd
}

func accountNoteCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("account", "a", "", "account")
	cmd.Flags().StringP("hash", "n", "", "notehash")
	cmd.Flags().Int32P("status", "s", 0, "note status:1:valid,2:used,3:frozen,4:unfrozen")

}

func accountNote(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	account, _ := cmd.Flags().GetString("account")
	hash, _ := cmd.Flags().GetString("hash")
	status, _ := cmd.Flags().GetInt32("status")

	if len(account) == 0 && len(hash) == 0 && status == 0 {
		fmt.Println("neet set parameters, check by --help")
		return
	}

	var params mixTy.WalletMixIndexReq
	params.Account = account
	params.NoteHash = hash
	params.Status = status

	var res mixTy.WalletNoteResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.ShowAccountNoteInfo", &params, &res)
	ctx.Run()
}

// ShowAccountPrivacyInfo get para chain status by height
func RescanStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "rescan status",
		Run:   rescanStatus,
	}
	rescanStatusCmdFlags(cmd)
	return cmd
}

func rescanStatusCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("accounts", "a", "", "accounts")
	//cmd.MarkFlagRequired("accounts")

}

func rescanStatus(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	//accounts, _ := cmd.Flags().GetString("accounts")

	//l := strings.Split(accounts,",")

	//var params types.ReqAddrs
	//params.Addrs = append(params.Addrs,l...)

	var res types.ReqString
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.GetRescanStatus", &types.ReqNil{}, &res)
	ctx.Run()
}

// ShowAccountPrivacyInfo get para chain status by height
func RescanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rescan",
		Short: "rescan notes",
		Run:   rescanNote,
	}
	rescanNoteCmdFlags(cmd)
	return cmd
}

func rescanNoteCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("accounts", "a", "", "accounts")
	//cmd.MarkFlagRequired("accounts")

}

func rescanNote(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	//accounts, _ := cmd.Flags().GetString("accounts")

	//l := strings.Split(accounts,",")

	//var params types.ReqAddrs
	//params.Addrs = append(params.Addrs,l...)

	var res types.ReqString
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.RescanNotes", &types.ReqNil{}, &res)
	ctx.Run()
}

// ShowAccountPrivacyInfo get para chain status by height
func EnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "enable privacy",
		Run:   enablePrivacy,
	}
	enablePrivacyCmdFlags(cmd)
	return cmd
}

func enablePrivacyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("accounts", "a", "", "accounts")
	//cmd.MarkFlagRequired("accounts")

}

func enablePrivacy(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	accounts, _ := cmd.Flags().GetString("accounts")

	var params types.ReqAddrs
	if len(accounts) > 0 {
		l := strings.Split(accounts, ",")
		params.Addrs = append(params.Addrs, l...)
	}

	var res mixTy.ReqEnablePrivacyRst
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.EnablePrivacy", &params, &res)
	ctx.Run()
}

func SecretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "note secret cmd",
	}
	//cmd.AddCommand(EncodeSecretDataCmd())
	cmd.AddCommand(DecodeSecretDataCmd())
	cmd.AddCommand(EncryptSecretDataCmd())
	cmd.AddCommand(DecryptSecretDataCmd())
	cmd.AddCommand(DecodePubInputDataCmd())

	return cmd
}

// DecodePublicInputDataCmd decode zk public data
func DecodePubInputDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse",
		Short: "parse zk public input data",
		Run:   decodePubInput,
	}
	decodePubInputCmdFlags(cmd)
	return cmd
}

func decodePubInputCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("data", "d", "", "public input data")
	cmd.MarkFlagRequired("data")

	cmd.Flags().Int32P("type", "t", 0, "type 0:deposit,1:withdraw,2:transIn,3:transOut,4:auth")
	cmd.MarkFlagRequired("type")

}

func decodePubInput(cmd *cobra.Command, args []string) {
	//rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	data, _ := cmd.Flags().GetString("data")
	ty, _ := cmd.Flags().GetInt32("type")

	v, err := mixTy.DecodePubInput(mixTy.VerifyType(ty), data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	rst, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(rst))
}

// EncodeSecretDataCmd get para chain status by height
func DecodeSecretDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode",
		Short: "decode secret data",
		Run:   decodeSecret,
	}
	decodeSecretCmdFlags(cmd)
	return cmd
}

func decodeSecretCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("data", "d", "", "receiver data")
	cmd.MarkFlagRequired("data")

}

func decodeSecret(cmd *cobra.Command, args []string) {
	//rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	data, _ := cmd.Flags().GetString("data")

	var secret mixTy.DHSecret
	d, err := hex.DecodeString(data)
	if err != nil {
		fmt.Println("decode string fail")
		return
	}
	err = types.Decode(d, &secret)
	if err != nil {
		fmt.Println("decode data fail")
		return
	}

	rst, err := json.MarshalIndent(secret, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(rst))
}

// ShowAccountPrivacyInfo get para chain status by height
func EncryptSecretDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "encrypt secret data",
		Run:   encryptSecret,
	}
	encryptSecrettCmdFlags(cmd)
	return cmd
}

func encryptSecrettCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("secret", "s", "", "raw secret data")
	cmd.MarkFlagRequired("secret")

	cmd.Flags().StringP("peerKey", "a", "", "peer pub key ")
	cmd.MarkFlagRequired("peerKey")

}

func encryptSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	secret, _ := cmd.Flags().GetString("secret")
	peerKey, _ := cmd.Flags().GetString("peerKey")

	req := mixTy.EncryptSecretData{
		Secret:  secret,
		PeerKey: peerKey,
	}

	var res mixTy.DHSecret
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.EncryptSecretData", req, &res)
	ctx.Run()
}

// ShowAccountPrivacyInfo get para chain status by height
func DecryptSecretDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "decrypt secret data by receiving privacy key",
		Run:   decryptSecret,
	}
	decryptSecrettCmdFlags(cmd)
	return cmd
}

func decryptSecrettCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("secret", "s", "", "raw secret data")
	cmd.MarkFlagRequired("secret")

	cmd.Flags().StringP("pri", "p", "", "receiving pri key")
	cmd.MarkFlagRequired("pri")

	cmd.Flags().StringP("peerKey", "a", "", "ephemeral pub key X")
	cmd.MarkFlagRequired("peerKey")

}

func decryptSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	secret, _ := cmd.Flags().GetString("secret")
	pri, _ := cmd.Flags().GetString("pri")
	peerKey, _ := cmd.Flags().GetString("peerKey")

	req := mixTy.DecryptSecretData{
		Secret:  secret,
		PeerKey: peerKey,
		PriKey:  pri,
	}

	var res mixTy.SecretData
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.DecryptSecretData", req, &res)
	ctx.Run()
}

func readFile(file string) string {
	// open file
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("err", err)
	}
	defer f.Close()

	var buff bytes.Buffer
	buff.ReadFrom(f)
	ret := hex.EncodeToString(buff.Bytes())
	fmt.Println("file", ret)
	return ret
}

// CreateDepositRawTxCmd get para chain status by height
func CreateDepositRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "one key get deposit input data",
		Run:   depositSecret,
	}
	depositSecretCmdFlags(cmd)
	return cmd
}

func depositSecretCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("targets", "t", "", "target addrs,seperated by ','")
	cmd.MarkFlagRequired("targets")

	cmd.Flags().StringP("return", "r", "", "return addr,optional")

	cmd.Flags().StringP("authorize", "a", "", "authorize addr,optional")

	cmd.Flags().StringP("amount", "m", "", "amounts,seperated by ','")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("token", "s", "BTY", "asset token, default BTY")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")

	cmd.Flags().StringP("path", "p", "", "deposit circuit path ")
	cmd.Flags().Int32P("privacy", "w", 0, "get zk privacy data print, 1:print, default not")
	cmd.Flags().Int32P("verify", "v", 0, "verify on chain:0 on local:1, default 0 ")

}

func depositSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	targets, _ := cmd.Flags().GetString("targets")
	returnAddr, _ := cmd.Flags().GetString("return")
	authorize, _ := cmd.Flags().GetString("authorize")
	amount, _ := cmd.Flags().GetString("amount")
	assetExec, _ := cmd.Flags().GetString("exec")
	token, _ := cmd.Flags().GetString("token")

	path, _ := cmd.Flags().GetString("path")
	privacy, _ := cmd.Flags().GetInt32("privacy")
	verify, _ := cmd.Flags().GetInt32("verify")

	deposit := &mixTy.DepositInfo{
		ReceiverAddrs: targets,
		ReturnAddr:    returnAddr,
		AuthorizeAddr: authorize,
		Amounts:       amount,
	}

	tx := &mixTy.DepositTxReq{
		Deposit: deposit,
		ZkPath:  path,
	}

	params := &mixTy.CreateRawTxReq{
		ActionTy:   mixTy.MixActionDeposit,
		Data:       types.Encode(tx),
		AssetExec:  assetExec,
		AssetToken: token,
		Title:      paraName,
		Privacy:    privacy,
		Verify:     verify,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.CreateRawTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateTransferRawTxCmd get para chain status by height
func CreateTransferRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "one key get transfer input output data",
		Run:   transferSecret,
	}
	transferSecretCmdFlags(cmd)
	return cmd
}

func transferSecretCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("noteHash", "n", "", "note hash to spend, seperate by ',' ")
	cmd.MarkFlagRequired("noteHash")

	cmd.Flags().StringP("toAddr", "t", "", "transfer to addr, only one addr")
	cmd.MarkFlagRequired("toAddr")

	cmd.Flags().StringP("auth", "a", "", "transfer to auth addr,optional")

	cmd.Flags().StringP("returner", "r", "", "transfer to returner addr,optional")

	cmd.Flags().StringP("amount", "m", "", "transfer amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("token", "s", "BTY", "asset token, default BTY")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")

	cmd.Flags().StringP("inpath", "i", "", "input path ")
	cmd.MarkFlagRequired("inpath")

	cmd.Flags().StringP("outpath", "o", "", "output pk file ")
	cmd.MarkFlagRequired("outpath")
	cmd.Flags().Int32P("privacy", "w", 0, "get zk privacy data print, 1:print, default not")
	cmd.Flags().Int32P("verify", "v", 0, "verify on chain:0 on local:1, default 0 ")

}

func transferSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	noteHash, _ := cmd.Flags().GetString("noteHash")
	toAddr, _ := cmd.Flags().GetString("toAddr")
	auth, _ := cmd.Flags().GetString("auth")
	returner, _ := cmd.Flags().GetString("returner")
	amount, _ := cmd.Flags().GetString("amount")

	inpath, _ := cmd.Flags().GetString("inpath")
	outpath, _ := cmd.Flags().GetString("outpath")

	assetExec, _ := cmd.Flags().GetString("exec")
	token, _ := cmd.Flags().GetString("token")

	privacy, _ := cmd.Flags().GetInt32("privacy")
	verify, _ := cmd.Flags().GetInt32("verify")

	input := &mixTy.TransferInputTxReq{
		NoteHashs: noteHash,
		ZkPath:    inpath,
	}

	deposit := &mixTy.DepositInfo{
		ReceiverAddrs: toAddr,
		ReturnAddr:    returner,
		AuthorizeAddr: auth,
		Amounts:       amount,
	}

	output := &mixTy.TransferOutputTxReq{
		Deposit: deposit,
		ZkPath:  outpath,
	}

	req := &mixTy.TransferTxReq{
		Input:  input,
		Output: output,
	}

	params := &mixTy.CreateRawTxReq{
		ActionTy:   mixTy.MixActionTransfer,
		Data:       types.Encode(req),
		AssetExec:  assetExec,
		AssetToken: token,
		Title:      paraName,
		Privacy:    privacy,
		Verify:     verify,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.CreateRawTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateWithdrawRawTxCmd get para chain status by height
func CreateWithdrawRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "one key get withdraw proof input data",
		Run:   withdrawSecret,
	}
	withdrawSecretCmdFlags(cmd)
	return cmd
}

func withdrawSecretCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("noteHashs", "n", "", "note hashs to spend,separate by ','")
	cmd.MarkFlagRequired("noteHashs")

	cmd.Flags().Uint64P("amount", "m", 0, "total amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("token", "s", "BTY", "asset token, default BTY")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")

	cmd.Flags().StringP("path", "p", "", "withdraw pk file ")
	cmd.MarkFlagRequired("path")

	cmd.Flags().Int32P("privacy", "w", 0, "get zk privacy data print, 1:print, default not")
	cmd.Flags().Int32P("verify", "v", 0, "verify on chain:0 on local:1, default 0 ")

}

func withdrawSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	noteHashs, _ := cmd.Flags().GetString("noteHashs")
	amount, _ := cmd.Flags().GetUint64("amount")

	assetExec, _ := cmd.Flags().GetString("exec")
	token, _ := cmd.Flags().GetString("token")

	path, _ := cmd.Flags().GetString("path")
	privacy, _ := cmd.Flags().GetInt32("privacy")
	verify, _ := cmd.Flags().GetInt32("verify")

	req := &mixTy.WithdrawTxReq{
		TotalAmount: amount,
		NoteHashs:   noteHashs,
		ZkPath:      path,
	}

	params := &mixTy.CreateRawTxReq{
		ActionTy:   mixTy.MixActionWithdraw,
		Data:       types.Encode(req),
		AssetExec:  assetExec,
		AssetToken: token,
		Title:      paraName,
		Privacy:    privacy,
		Verify:     verify,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.CreateRawTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateAuthRawTxCmd get para chain status by height
func CreateAuthRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "one key get authorize input data",
		Run:   authSecret,
	}
	authSecretCmdFlags(cmd)
	return cmd
}

func authSecretCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("noteHash", "n", "", "note hash to authorize")
	cmd.MarkFlagRequired("noteHash")

	cmd.Flags().StringP("toKey", "a", "", "authorize to key")
	cmd.MarkFlagRequired("toKey")

	cmd.Flags().StringP("token", "s", "BTY", "asset token, default BTY")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")

	cmd.Flags().StringP("path", "p", "", "auth path file ")
	cmd.MarkFlagRequired("path")

	cmd.Flags().Int32P("privacy", "w", 0, "get zk privacy data print, 1:print, default not")
	cmd.Flags().Int32P("verify", "v", 0, "verify on chain:0 on local:1, default 0 ")

}

func authSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	noteHash, _ := cmd.Flags().GetString("noteHash")
	toKey, _ := cmd.Flags().GetString("toKey")

	assetExec, _ := cmd.Flags().GetString("exec")
	token, _ := cmd.Flags().GetString("token")

	path, _ := cmd.Flags().GetString("path")

	privacy, _ := cmd.Flags().GetInt32("privacy")
	verify, _ := cmd.Flags().GetInt32("verify")

	req := &mixTy.AuthTxReq{
		AuthorizeToAddr: toKey,
		NoteHash:        noteHash,
		ZkPath:          path,
	}

	params := &mixTy.CreateRawTxReq{
		ActionTy:   mixTy.MixActionAuth,
		Data:       types.Encode(req),
		AssetExec:  assetExec,
		AssetToken: token,
		Title:      paraName,
		Privacy:    privacy,
		Verify:     verify,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.CreateRawTransaction", params, nil)
	ctx.RunWithoutMarshal()
}
