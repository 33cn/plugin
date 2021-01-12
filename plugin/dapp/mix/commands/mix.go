// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
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
		CreateDepositCmd(),
		CreateTransferCmd(),
		CreateWithdrawCmd(),
		CreateConfigCmd(),
		CreateAuthorizeCmd(),
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
func CreateDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "Create a asset deposit to mix coin contract",
		Run:   createDeposit,
	}
	addCreateDepositFlags(cmd)
	return cmd
}

func addCreateDepositFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proofs", "p", "", "'proof-pubinput' pair, multi pairs allowed with ','")
	cmd.MarkFlagRequired("proofs")

	cmd.Flags().Uint64P("amount", "a", 0, "deposit amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("secretPayment", "s", "", "secret for payment addr")
	cmd.MarkFlagRequired("secretPayment")

	cmd.Flags().StringP("pubX", "x", "", "receiving pub key X")
	cmd.MarkFlagRequired("pubX")

	cmd.Flags().StringP("pubY", "y", "", "receiving pub key Y")
	cmd.MarkFlagRequired("pubY")

	cmd.Flags().StringP("secretAuth", "u", "", "secret for authorize addr")

	cmd.Flags().StringP("secretReturn", "r", "", "secret for return addr")

}

func parseProofPara(input string) ([]*mixTy.ZkProofInfo, error) {
	var proofInputs []*mixTy.ZkProofInfo
	inputParas := strings.Split(input, ",")
	for _, i := range inputParas {
		inputs := strings.Split(i, "-")
		if len(inputs) != 2 {
			fmt.Println("proofs parameters not correct:", i)
			return nil, types.ErrInvalidParam
		}
		var proofInfo mixTy.ZkProofInfo
		proofInfo.Proof = inputs[0]
		proofInfo.PublicInput = inputs[1]
		proofInputs = append(proofInputs, &proofInfo)
	}
	return proofInputs, nil
}

func createDeposit(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	amount, _ := cmd.Flags().GetUint64("amount")
	proofsPara, _ := cmd.Flags().GetString("proofs")
	secretPayment, _ := cmd.Flags().GetString("secretPayment")
	pubX, _ := cmd.Flags().GetString("pubX")
	pubY, _ := cmd.Flags().GetString("pubY")

	proofInputs, err := parseProofPara(proofsPara)
	if err != nil {
		return
	}

	var pubkey mixTy.PubKey
	pubkey.X = pubX
	pubkey.Y = pubY
	var paySecret mixTy.DHSecret
	paySecret.Secret = secretPayment
	paySecret.Epk = &pubkey

	var group mixTy.DHSecretGroup
	group.Spender = &paySecret

	payload := &mixTy.MixDepositAction{}
	payload.Amount = amount
	payload.NewCommits = append(payload.NewCommits, proofInputs...)
	payload.NewCommits[0].Group = &group

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, mixTy.MixX),
		ActionName: "Deposit",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

// CreateWithdrawCmd create raw asset transfer tx
func CreateWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "Create a asset withdraw from mix coin contract",
		Run:   createWithdraw,
	}
	addCreateWithdrawFlags(cmd)
	return cmd
}

func addCreateWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proofs", "p", "", "spend 'proof-pubinput' pair, multi pairs allowed with ','")
	cmd.MarkFlagRequired("proofs")

	cmd.Flags().Uint64P("amount", "a", 0, "withdraw amount")
	cmd.MarkFlagRequired("amount")

}

func createWithdraw(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	amount, _ := cmd.Flags().GetUint64("amount")
	proofsPara, _ := cmd.Flags().GetString("proofs")

	proofInputs, err := parseProofPara(proofsPara)
	if err != nil {
		return
	}

	payload := &mixTy.MixWithdrawAction{}
	payload.Amount = amount
	payload.SpendCommits = append(payload.SpendCommits, proofInputs...)
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, mixTy.MixX),
		ActionName: "Withdraw",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

// CreateTransferCmd create raw asset transfer tx
func CreateTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Create a asset transfer in mix coin contract",
		Run:   createTransfer,
	}
	addCreateTransferFlags(cmd)
	return cmd
}

func addCreateTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("input", "i", "", "input 'proof-pubinput' pair, multi pairs allowed with ','")
	cmd.MarkFlagRequired("input")

	cmd.Flags().StringP("output", "o", "", "output 'proof-pubinput' pair, multi pairs allowed with ','")
	cmd.MarkFlagRequired("output")

}

func createTransfer(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	proofsInput, _ := cmd.Flags().GetString("input")
	proofsOutput, _ := cmd.Flags().GetString("output")

	proofInputs, err := parseProofPara(proofsInput)
	if err != nil {
		fmt.Println("proofsInput error")
		return
	}
	proofOutputs, err := parseProofPara(proofsOutput)
	if err != nil {
		fmt.Println("proofsOutput error")
		return
	}

	payload := &mixTy.MixTransferAction{}
	payload.Input = append(payload.Input, proofInputs...)
	payload.Output = append(payload.Output, proofOutputs...)
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, mixTy.MixX),
		ActionName: "Transfer",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

// CreateAuthorizeCmd create raw asset transfer tx
func CreateAuthorizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorize",
		Short: "Create a asset authorize in mix coin contract",
		Run:   createAuthorize,
	}
	addCreateAuthorizeFlags(cmd)
	return cmd
}

func addCreateAuthorizeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proofs", "p", "", "authorize 'proof-pubinput' pair, multi pairs allowed with ','")
	cmd.MarkFlagRequired("proofs")

}

func createAuthorize(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	proofsPara, _ := cmd.Flags().GetString("proofs")

	proofInputs, err := parseProofPara(proofsPara)
	if err != nil {
		return
	}

	payload := &mixTy.MixAuthorizeAction{}
	payload.AuthCommits = append(payload.AuthCommits, proofInputs...)
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, mixTy.MixX),
		ActionName: "Authorize",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

// CreateDepositCmd create raw asset transfer tx
func CreateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Proof parameters config to mix coin contract",
	}
	cmd.AddCommand(mixConfigVerifyKeyParaCmd())
	cmd.AddCommand(mixConfigAuthPubKeyParaCmd())

	return cmd
}

func mixConfigVerifyKeyParaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "zk proof verify key config cmd",
		Run:   createConfigVerify,
	}
	addVkConfigFlags(cmd)

	return cmd
}

func addVkConfigFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("action", "a", 0, "0:add,1:delete")

	cmd.Flags().Uint32P("curveid", "i", 3, "zk curve id,1:bls377,2:bls381,3:bn256")
	cmd.MarkFlagRequired("curveid")

	cmd.Flags().Uint32P("circuit", "c", 0, "mix circuit type,0:deposit,1:withdraw,2:spendinput,3:spendout,4:authorize")
	cmd.MarkFlagRequired("circuit")

	cmd.Flags().StringP("zkey", "z", "", "zk proof verify key")
	cmd.MarkFlagRequired("zkey")

}

func createConfigVerify(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	action, _ := cmd.Flags().GetUint32("action")
	curveid, _ := cmd.Flags().GetUint32("curveid")
	circuit, _ := cmd.Flags().GetUint32("circuit")
	key, _ := cmd.Flags().GetString("zkey")

	var zkVk mixTy.ZkVerifyKey
	zkVk.Value = key
	zkVk.Type = mixTy.VerifyType(circuit)
	zkVk.CurveId = mixTy.ZkCurveId(curveid)

	payload := &mixTy.MixConfigAction{}
	payload.Ty = mixTy.MixConfigType_VerifyKey
	payload.Action = mixTy.MixConfigAct(action)
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
		Use:   "pubkey",
		Short: "mix authorize pub key config cmd",
		Run:   createConfigPubKey,
	}
	addPubKeyConfigFlags(cmd)

	return cmd
}

func addPubKeyConfigFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("action", "a", 0, "0:add,1:delete")

	cmd.Flags().StringP("key", "k", "", "authorize pub key")
	cmd.MarkFlagRequired("key")

}

func createConfigPubKey(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	action, _ := cmd.Flags().GetUint32("action")
	key, _ := cmd.Flags().GetString("key")

	//var pubkey mixTy.AuthorizePubKey
	//pubkey.Value = key

	payload := &mixTy.MixConfigAction{}
	payload.Ty = mixTy.MixConfigType_AuthPubKey
	payload.Action = mixTy.MixConfigAct(action)
	payload.Value = &mixTy.MixConfigAction_AuthPk{AuthPk: key}

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
	cmd.AddCommand(EncodeSecretDataCmd())
	cmd.AddCommand(EncryptSecretDataCmd())
	cmd.AddCommand(DecryptSecretDataCmd())

	return cmd
}

// ShowAccountPrivacyInfo get para chain status by height
func ShowAccountPrivacyInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "show account privacy keys",
		Run:   accountPrivacy,
	}
	accountPrivacyCmdFlags(cmd)
	return cmd
}

func accountPrivacyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("account", "a", "", "accounts")
	cmd.MarkFlagRequired("account")

}

func accountPrivacy(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	account, _ := cmd.Flags().GetString("account")

	var res mixTy.WalletAddrPrivacy
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.ShowAccountPrivacyInfo", &types.ReqString{Data: account}, &res)
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
	cmd.Flags().StringP("accounts", "a", "", "accounts")
	cmd.MarkFlagRequired("accounts")

}

func accountNote(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	accounts, _ := cmd.Flags().GetString("accounts")

	l := strings.Split(accounts, ",")

	var params types.ReqAddrs
	params.Addrs = append(params.Addrs, l...)

	var res mixTy.WalletIndexResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.ShowAccountNoteInfo", params, &res)
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

// ShowAccountPrivacyInfo get para chain status by height
func EncodeSecretDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode",
		Short: "encode secret data",
		Run:   encodeSecret,
	}
	encodeSecretCmdFlags(cmd)
	return cmd
}

func encodeSecretCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("payment", "p", "", "payment key")
	cmd.MarkFlagRequired("payment")

	cmd.Flags().StringP("return", "r", "", "return key")

	cmd.Flags().StringP("authorize", "a", "", "authorize key")

	cmd.Flags().StringP("amount", "m", "", "amount with 1e8")
	cmd.MarkFlagRequired("amount")

}

func encodeSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payment, _ := cmd.Flags().GetString("payment")
	returnKey, _ := cmd.Flags().GetString("return")
	authorize, _ := cmd.Flags().GetString("authorize")
	amount, _ := cmd.Flags().GetString("amount")

	req := mixTy.SecretData{
		PaymentPubKey:   payment,
		ReturnPubKey:    returnKey,
		AuthorizePubKey: authorize,
		Amount:          amount,
	}

	var res mixTy.EncodedSecretData
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.EncodeSecretData", req, &res)
	ctx.Run()
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

	cmd.Flags().StringP("pubX", "x", "", "receiving pub key X")
	cmd.MarkFlagRequired("pubX")

	cmd.Flags().StringP("pubY", "y", "", "receiving pub key Y")
	cmd.MarkFlagRequired("pubY")

}

func encryptSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	secret, _ := cmd.Flags().GetString("secret")
	x, _ := cmd.Flags().GetString("pubX")
	y, _ := cmd.Flags().GetString("pubY")

	pubkey := &mixTy.PubKey{X: x, Y: y}

	req := mixTy.EncryptSecretData{
		Secret:      secret,
		ReceivingPk: pubkey,
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

	cmd.Flags().StringP("epkX", "x", "", "ephemeral pub key X")
	cmd.MarkFlagRequired("epkX")

	cmd.Flags().StringP("epkY", "y", "", "ephemeral pub key Y")
	cmd.MarkFlagRequired("epkY")
}

func decryptSecret(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	secret, _ := cmd.Flags().GetString("secret")
	p, _ := cmd.Flags().GetString("pri")
	x, _ := cmd.Flags().GetString("epkX")
	y, _ := cmd.Flags().GetString("epkY")

	pubkey := &mixTy.PubKey{X: x, Y: y}

	prikey := &mixTy.PrivKey{Data: p}

	req := mixTy.DecryptSecretData{
		Secret:          secret,
		Epk:             pubkey,
		ReceivingPriKey: prikey,
	}

	var res mixTy.SecretData
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "mix.DecryptSecretData", req, &res)
	ctx.Run()
}
