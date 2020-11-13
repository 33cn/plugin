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
func ParcCmd() *cobra.Command {
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

	proofInputs, err := parseProofPara(proofsPara)
	if err != nil {
		return
	}

	payload := &mixTy.MixDepositAction{}
	payload.Amount = amount
	payload.NewCommits = append(payload.NewCommits, proofInputs...)
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

	cmd.Flags().Uint32P("curveid", "i", 0, "zk curve id,1:bls377,2:bls381,3:bn256")
	cmd.MarkFlagRequired("curveid")

	cmd.Flags().Uint32P("circuit", "c", 0, "mix circuit type,0:deposit,1:withdraw,2:spendinput,3:spendout,4:authorize")
	cmd.MarkFlagRequired("circuit")

	cmd.Flags().StringP("key", "k", "", "zk proof verify key")
	cmd.MarkFlagRequired("key")

}

func createConfigVerify(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	action, _ := cmd.Flags().GetUint32("action")
	curveid, _ := cmd.Flags().GetUint32("curveid")
	circuit, _ := cmd.Flags().GetUint32("circuit")
	key, _ := cmd.Flags().GetString("key")

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

	cmd.Flags().StringP("key", "k", "", "zk proof verify key")
	cmd.MarkFlagRequired("key")

}

func createConfigPubKey(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	action, _ := cmd.Flags().GetUint32("action")
	key, _ := cmd.Flags().GetString("key")

	var pubkey mixTy.AuthorizePubKey
	pubkey.Value = key

	payload := &mixTy.MixConfigAction{}
	payload.Ty = mixTy.MixConfigType_AuthPubKey
	payload.Action = mixTy.MixConfigAct(action)
	payload.Value = &mixTy.MixConfigAction_AuthKey{AuthKey: &pubkey}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, mixTy.MixX),
		ActionName: "Config",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}
