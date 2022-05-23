/*Package commands implement dapp client commands*/
package commands

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

// ZksyncCmd zksync client command
func ZksyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zksync",
		Short: "zksync command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		depositCmd(),
		withdrawCmd(),
		contractToTreeCmd(),
		treeToContractCmd(),
		transferCmd(),
		transferToNewCmd(),
		forceExitCmd(),
		setPubKeyCmd(),
		fullExitCmd(),
		setVerifyKeyCmd(),
		setOperatorCmd(),
		getChain33AddrCmd(),
		setTokenFeeCmd(),
		queryCmd(),
		//NFT
		nftCmd(),
	)
	return cmd
}

func depositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "get deposit tx",
		Run:   deposit,
	}
	depositFlag(cmd)
	return cmd
}

func depositFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 1, "deposit tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "a", "0", "deposit amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "", "deposit ethaddress")
	cmd.MarkFlagRequired("ethAddress")
	cmd.Flags().StringP("chain33Addr", "c", "", "deposit chain33Addr")
	cmd.MarkFlagRequired("chain33Addr")
	cmd.Flags().Uint64P("queueId", "i", 0, "eth queue id")
	cmd.MarkFlagRequired("queueId")

}

func deposit(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")
	queueId, _ := cmd.Flags().GetUint64("queueId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	deposit := &zt.ZkDeposit{
		TokenId:            tokenId,
		Amount:             amount,
		EthAddress:         ethAddress,
		Chain33Addr:        chain33Addr,
		EthPriorityQueueId: int64(queueId),
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "Deposit",
		Payload:    types.MustPBToJSON(deposit),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func withdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "get withdraw tx",
		Run:   withdraw,
	}
	withdrawFlag(cmd)
	return cmd
}

func withdrawFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 1, "withdraw tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "a", "0", "withdraw amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint64P("accountId", "i", 0, "withdraw accountId")
	cmd.MarkFlagRequired("accountId")

}

func withdraw(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyWithdrawAction, tokenId, amount, "", "", "", accountId, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "Withdraw",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func treeToContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree2contract",
		Short: "get treeToContract tx",
		Run:   treeToContract,
	}
	treeToContractFlag(cmd)
	return cmd
}

func treeToContractFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 1, "treeToContract tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "a", "0", "treeToContract amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint64P("accountId", "", 0, "treeToContract accountId")
	cmd.MarkFlagRequired("accountId")

}

func treeToContract(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyTreeToContractAction, tokenId, amount, "", "", "", accountId, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "TreeToContract",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func contractToTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract2tree",
		Short: "get contractToTree tx",
		Run:   contractToTree,
	}
	contractToTreeFlag(cmd)
	return cmd
}

func contractToTreeFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 1, "contractToTree tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "a", "0", "contractToTree amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint64P("accountId", "", 0, "contractToTree accountId")
	cmd.MarkFlagRequired("accountId")

}

func contractToTree(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyContractToTreeAction, tokenId, amount, "", "", "", accountId, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "ContractToTree",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func transferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "get transfer tx",
		Run:   transfer,
	}
	transferFlag(cmd)
	return cmd
}

func transferFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 1, "transfer tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "a", "0", "transfer amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint64P("accountId", "", 0, "transfer fromAccountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().Uint64P("toAccountId", "", 0, "transfer toAccountId")
	cmd.MarkFlagRequired("toAccountId")

}

func transfer(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	toAccountId, _ := cmd.Flags().GetUint64("toAccountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyTransferAction, tokenId, amount, "", "", "", accountId, toAccountId)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "Transfer",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func transferToNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer2new",
		Short: "get transferToNew tx",
		Run:   transferToNew,
	}
	transferToNewFlag(cmd)
	return cmd
}

func transferToNewFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 1, "transferToNew tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "a", "0", "transferToNew amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint64P("accountId", "", 0, "transferToNew fromAccountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().StringP("ethAddress", "e", "", "transferToNew toEthAddress")
	cmd.MarkFlagRequired("ethAddress")
	cmd.Flags().StringP("chain33Addr", "c", "", "transferToNew toChain33Addr")
	cmd.MarkFlagRequired("chain33Addr")
}

func transferToNew(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	toEthAddress, _ := cmd.Flags().GetString("ethAddress")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyTransferToNewAction, tokenId, amount, "", toEthAddress, chain33Addr, accountId, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "TransferToNew",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func forceExitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forceexit",
		Short: "withdraw by other addr",
		Run:   forceExit,
	}
	forceExitFlag(cmd)
	return cmd
}

func forceExitFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 1, "target tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("accountId", "a", 0, "target accountId")
	cmd.MarkFlagRequired("accountId")

}

func forceExit(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	accountId, _ := cmd.Flags().GetUint64("accountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyForceExitAction, tokenId, "0", "", "", "", accountId, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "ForceExit",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func setPubKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubkey",
		Short: "set pubkey",
		Run:   setPubKey,
	}
	setPubKeyFlag(cmd)
	return cmd
}

func setPubKeyFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("accountId", "a", 0, "setPubKeyFlag accountId")
	cmd.MarkFlagRequired("accountId")

	cmd.Flags().Uint64P("pubkeyT", "t", 0, "self default:0, proxy pubkey ty: 1: normal,2:system,3:super")

	cmd.Flags().StringP("pubkeyX", "x", "", "proxy pubkey x value")
	cmd.Flags().StringP("pubkeyY", "y", "", "proxy pubkey y value")

}

func setPubKey(cmd *cobra.Command, args []string) {
	accountId, _ := cmd.Flags().GetUint64("accountId")
	pubkeyT, _ := cmd.Flags().GetUint64("pubkeyT")
	pubkeyX, _ := cmd.Flags().GetString("pubkeyX")
	pubkeyY, _ := cmd.Flags().GetString("pubkeyY")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	if pubkeyT > 0 && (len(pubkeyX) == 0 || len(pubkeyY) == 0) {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("set proxy pubkey, need set pubkeyX pubkeyY"))
		return
	}

	pubkey := &zt.ZkSetPubKey{
		AccountId: accountId,
		PubKeyTy:  pubkeyT,
		PubKey: &zt.ZkPubKey{
			X: pubkeyX,
			Y: pubkeyY,
		},
	}
	payload := types.MustPBToJSON(pubkey)
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "SetPubKey",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func fullExitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fullexit",
		Short: "get fullExit tx",
		Run:   fullExit,
	}
	fullExitFlag(cmd)
	return cmd
}

func fullExitFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 1, "fullExit tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("accountId", "a", 0, "fullExit accountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().Uint64P("queueId", "i", 0, "eth queue id")
	cmd.MarkFlagRequired("queueId")
}

func fullExit(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	queueId, _ := cmd.Flags().GetUint64("queueId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	fullExit := &zt.ZkFullExit{
		TokenId:            tokenId,
		AccountId:          accountId,
		EthPriorityQueueId: int64(queueId),
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "FullExit",
		Payload:    types.MustPBToJSON(fullExit),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func setVerifyKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vkey",
		Short: "set verify key for zk-proof",
		Run:   verifyKey,
	}
	addVerifyKeyCmdFlags(cmd)
	return cmd
}

func addVerifyKeyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("vkey", "v", "", "verify key")
	_ = cmd.MarkFlagRequired("vkey")

}

func verifyKey(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	vkey, _ := cmd.Flags().GetString("vkey")

	payload := &zt.ZkVerifyKey{
		Key: vkey,
	}
	exec := zt.Zksync
	if strings.HasPrefix(paraName, pt.ParaPrefix) {
		exec = paraName + zt.Zksync
	}
	params := &rpctypes.CreateTxIn{
		Execer:     exec,
		ActionName: "SetVerifyKey",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func setOperatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "set operators for commit zk-proof",
		Run:   setOperator,
	}
	addOperatorCmdFlags(cmd)
	return cmd
}

func addOperatorCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("operator", "o", "", "operators, separate with '-'")
	_ = cmd.MarkFlagRequired("operator")

}

func setOperator(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	operator, _ := cmd.Flags().GetString("operator")

	payload := &zt.ZkVerifier{
		Verifiers: strings.Split(operator, "-"),
	}
	exec := zt.Zksync
	if strings.HasPrefix(paraName, pt.ParaPrefix) {
		exec = paraName + zt.Zksync
	}
	params := &rpctypes.CreateTxIn{
		Execer:     exec,
		ActionName: "SetVerifier",
		Payload:    types.MustPBToJSON(payload),
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func commitProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "commit proof test",
		Run:   commitProof,
	}
	addCommitProofCmdFlags(cmd)
	return cmd
}

func addCommitProofCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("start", "s", 0, "block start ")
	_ = cmd.MarkFlagRequired("start")
	cmd.Flags().Uint64P("end", "e", 0, "block end ")
	_ = cmd.MarkFlagRequired("end")
	cmd.Flags().StringP("old", "o", "0", "old tree hash")
	_ = cmd.MarkFlagRequired("old")
	cmd.Flags().StringP("new", "n", "0", "new tree hash")
	_ = cmd.MarkFlagRequired("new")
	cmd.Flags().StringP("pubdata", "d", "0", "pub datas, separate with '-'")
	_ = cmd.MarkFlagRequired("pubdata")
	cmd.Flags().StringP("public", "i", "0", "public input")
	_ = cmd.MarkFlagRequired("public")
	cmd.Flags().StringP("proof", "p", "0", "proof")
	_ = cmd.MarkFlagRequired("proof")

}

func commitProof(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	start, _ := cmd.Flags().GetUint64("start")
	end, _ := cmd.Flags().GetUint64("end")
	old, _ := cmd.Flags().GetString("old")
	new, _ := cmd.Flags().GetString("new")
	pubdata, _ := cmd.Flags().GetString("pubdata")
	public, _ := cmd.Flags().GetString("public")
	proof, _ := cmd.Flags().GetString("proof")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	payload := &zt.ZkCommitProof{
		BlockStart:  start,
		BlockEnd:    end,
		OldTreeRoot: old,
		NewTreeRoot: new,
		PublicInput: public,
		Proof:       proof,
		PubDatas:    strings.Split(pubdata, "-"),
	}
	exec := zt.Zksync
	if strings.HasPrefix(paraName, pt.ParaPrefix) {
		exec = paraName + zt.Zksync
	}
	params := &rpctypes.CreateTxIn{
		Execer:     exec,
		ActionName: "CommitProof",
		Payload:    types.MustPBToJSON(payload),
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func getChain33AddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "l2addr",
		Short: "get chain33 l2 address by privateKey",
		Run:   getChain33Addr,
	}
	getChain33AddrFlag(cmd)
	return cmd
}

func getChain33AddrFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("private", "k", "", "private key")
	_ = cmd.MarkFlagRequired("private")

	cmd.Flags().BoolP("pubkey", "p", false, "print pubkey")
}

func getChain33Addr(cmd *cobra.Command, args []string) {
	privateKeyString, _ := cmd.Flags().GetString("private")
	pubkey, _ := cmd.Flags().GetBool("pubkey")
	privateKeyBytes, err := common.FromHex(privateKeyString)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "hex.DecodeString"))
		return
	}
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(privateKeyBytes))

	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "eddsa.GenerateKey"))
		return
	}

	hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	hash.Write(zt.Str2Byte(privateKey.PublicKey.A.X.String()))
	hash.Write(zt.Str2Byte(privateKey.PublicKey.A.Y.String()))
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))
	if pubkey {
		fmt.Println("pubKey.X:", privateKey.PublicKey.A.X.String())
		fmt.Println("pubKey.Y:", privateKey.PublicKey.A.Y.String())
	}

}

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query related cmd",
	}
	cmd.AddCommand(queryAccountCmd())
	cmd.AddCommand(queryProofCmd())

	return cmd
}

func queryAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "query account related cmd",
	}
	cmd.AddCommand(getAccountTreeCmd())
	cmd.AddCommand(getAccountByIdCmd())
	cmd.AddCommand(getAccountByEthCmd())
	cmd.AddCommand(getAccountByChain33Cmd())
	cmd.AddCommand(getContractAccountCmd())
	cmd.AddCommand(getTokenBalanceCmd())

	return cmd
}

func getAccountTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree",
		Short: "get current accountTree",
		Run:   getAccountTree,
	}
	getAccountTreeFlag(cmd)
	return cmd
}

func getAccountTreeFlag(cmd *cobra.Command) {
}

func getAccountTree(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := new(types.ReqNil)

	params.FuncName = "GetAccountTree"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.AccountTree
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getAccountByIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "get zksync account by id",
		Run:   getAccountById,
	}
	getAccountByIdFlag(cmd)
	return cmd
}

func getAccountByIdFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("accountId", "a", 0, "zksync accountId")
	cmd.MarkFlagRequired("accountId")
}

func getAccountById(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	accountId, _ := cmd.Flags().GetUint64("accountId")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		AccountId: accountId,
	}

	params.FuncName = "GetAccountById"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.Leaf
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getAccountByEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accountE",
		Short: "get zksync account by ethAddress",
		Run:   getAccountByEth,
	}
	getAccountByEthFlag(cmd)
	return cmd
}

func getAccountByEthFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("ethAddress", "e", " ", "zksync account ethAddress")
	cmd.MarkFlagRequired("ethAddress")
}

func getAccountByEth(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		EthAddress: ethAddress,
	}

	params.FuncName = "GetAccountByEth"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getAccountByChain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accountC",
		Short: "get zksync account by chain33Addr",
		Run:   getAccountByChain33,
	}
	getAccountByChain33Flag(cmd)
	return cmd
}

func getAccountByChain33Flag(cmd *cobra.Command) {
	cmd.Flags().StringP("chain33Addr", "c", "", "zksync account chain33Addr")
	cmd.MarkFlagRequired("chain33Addr")
}

func getAccountByChain33(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		Chain33Addr: chain33Addr,
	}

	params.FuncName = "GetAccountByChain33"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getContractAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contractAccount",
		Short: "get zksync contractAccount by chain33WalletAddr and token symbol",
		Run:   getContractAccount,
	}
	getContractAccountFlag(cmd)
	return cmd
}

func getContractAccountFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("chain33Addr", "c", "", "chain33 wallet address")
	cmd.MarkFlagRequired("chain33Addr")
	cmd.Flags().StringP("token", "t", "", "token symbol")
	cmd.MarkFlagRequired("token")
}

func getContractAccount(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")
	token, _ := cmd.Flags().GetString("token")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenSymbol:       token,
		Chain33WalletAddr: chain33Addr,
	}

	params.FuncName = "GetZkContractAccount"
	params.Payload = types.MustPBToJSON(req)

	var resp types.Account
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getTokenBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "get zksync tokenBalance by accountId and tokenId",
		Run:   getTokenBalance,
	}
	getTokenBalanceFlag(cmd)
	return cmd
}

func getTokenBalanceFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("accountId", "a", 0, "zksync account id")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().Uint64P("token", "t", 1, "zksync token id")
	cmd.MarkFlagRequired("token")
}

func getTokenBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	token, _ := cmd.Flags().GetUint64("token")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenId:   token,
		AccountId: accountId,
	}

	params.FuncName = "GetTokenBalance"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func queryProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proof",
		Short: "query proof related cmd",
	}
	cmd.AddCommand(getTxProofCmd())
	cmd.AddCommand(getTxProofByHeightCmd())
	cmd.AddCommand(getProofByHeightsCmd())
	cmd.AddCommand(getLastCommitProofCmd())
	cmd.AddCommand(getZkCommitProofCmd())
	cmd.AddCommand(getFirstRootHashCmd())
	cmd.AddCommand(getZkCommitProofListCmd())

	//cmd.AddCommand(commitProofCmd())

	return cmd
}

func getTxProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "get tx proof",
		Run:   getTxProof,
	}
	getTxProofFlag(cmd)
	return cmd
}

func getTxProofFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("height", "g", 0, "zksync proof height")
	cmd.MarkFlagRequired("height")
	cmd.Flags().Uint32P("index", "i", 0, "tx index")
	cmd.MarkFlagRequired("index")
}

func getTxProof(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	height, _ := cmd.Flags().GetUint64("height")
	index, _ := cmd.Flags().GetUint32("index")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		BlockHeight: height,
		TxIndex:     index,
	}

	params.FuncName = "GetTxProof"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.OperationInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getTxProofByHeightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "get block proofs by height",
		Run:   getTxProofByHeight,
	}
	getTxProofByHeightFlag(cmd)
	return cmd
}

func getTxProofByHeightFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("height", "g", 0, "zksync proof height")
	cmd.MarkFlagRequired("height")
}

func getTxProofByHeight(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	height, _ := cmd.Flags().GetUint64("height")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		BlockHeight: height,
	}

	params.FuncName = "GetTxProofByHeight"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getProofByHeightsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocks",
		Short: "get proofs by height range",
		Run:   getProofByHeights,
	}
	getProofByHeightsFlag(cmd)
	return cmd
}

func getProofByHeightsFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("start", "s", 0, "start height")
	cmd.MarkFlagRequired("start")

	cmd.Flags().Uint64P("end", "e", 0, "end height")
	cmd.MarkFlagRequired("end")

	cmd.Flags().Uint64P("index", "i", 0, "start index of block")
	cmd.MarkFlagRequired("start")

	cmd.Flags().Uint32P("op", "o", 0, "op index of block")
	cmd.MarkFlagRequired("op")

	cmd.Flags().BoolP("detail", "d", false, "if need detail")
	cmd.MarkFlagRequired("detail")
}

func getProofByHeights(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	start, _ := cmd.Flags().GetUint64("start")
	end, _ := cmd.Flags().GetUint64("end")
	index, _ := cmd.Flags().GetUint64("index")
	op, _ := cmd.Flags().GetUint32("op")
	detail, _ := cmd.Flags().GetBool("detail")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryProofReq{
		StartBlockHeight: start,
		EndBlockHeight:   end,
		StartIndex:       index,
		OpIndex:          op,
		NeedDetail:       detail,
	}

	params.FuncName = "GetTxProofByHeights"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryProofResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getLastCommitProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last",
		Short: "get last committed proof",
		Run:   getLastCommitProof,
	}

	return cmd
}

func getLastCommitProof(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync

	params.FuncName = "GetLastCommitProof"
	params.Payload = types.MustPBToJSON(&types.ReqNil{})

	var resp zt.CommitProofState
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getZkCommitProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "get zkcommit proof by proofId",
		Run:   getZkCommitProof,
	}
	getZkCommitProofFlag(cmd)
	return cmd
}

func getZkCommitProofFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("proofId", "i", 0, "commit proof id")
	cmd.MarkFlagRequired("proofId")
}

func getZkCommitProof(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proofId, _ := cmd.Flags().GetUint64("proofId")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		ProofId: proofId,
	}

	params.FuncName = "GetCommitProodByProofId"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkCommitProof
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func setTokenFeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee",
		Short: "set zkoption fee",
		Run:   setTokenFee,
	}
	setTokenFeeFlag(cmd)
	return cmd
}

func setTokenFeeFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "token id")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("fee", "f", "10000", "fee")
	cmd.MarkFlagRequired("fee")
	cmd.Flags().Int32P("action", "a", 0, "action ty")
	cmd.MarkFlagRequired("action")
}

func setTokenFee(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	fee, _ := cmd.Flags().GetString("fee")
	action, _ := cmd.Flags().GetInt32("action")

	payload := &zt.ZkSetFee{
		TokenId:  tokenId,
		Amount:   fee,
		ActionTy: action,
	}

	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "SetFee",
		Payload:    types.MustPBToJSON(payload),
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func getFirstRootHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initroot",
		Short: "get merkel tree init root, default from cfg fee",
		Run:   getFirstRootHash,
	}
	getFirstRootHashFlag(cmd)
	return cmd
}

func getFirstRootHashFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("ethAddr", "e", "", "optional eth fee addr, hex format default from config")
	cmd.Flags().StringP("chain33Addr", "c", "", "optional chain33 fee addr, hex format,default from config")
}

func getFirstRootHash(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	eth, _ := cmd.Flags().GetString("ethAddr")
	chain33, _ := cmd.Flags().GetString("chain33Addr")

	var params rpctypes.Query4Jrpc
	params.Execer = zt.Zksync
	req := &types.ReqAddrs{Addrs: []string{eth, chain33}}

	params.FuncName = "GetTreeInitRoot"
	params.Payload = types.MustPBToJSON(req)

	var resp types.ReplyString
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getZkCommitProofListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plist",
		Short: "get committed proof list",
		Run:   getZkCommitProofList,
	}
	getZkCommitProofListFlag(cmd)
	return cmd
}

func getZkCommitProofListFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("proofId", "i", 0, "commit proof id")
	cmd.MarkFlagRequired("proofId")
	cmd.Flags().Uint64P("onChainProofId", "s", 0, "commit on chain proof id")

	cmd.Flags().BoolP("onChain", "o", true, "if req onChain proof by sub id")
	cmd.Flags().BoolP("latestProof", "l", false, "if req latest proof")
	cmd.Flags().Uint64P("endHeight", "e", 0, "latest proof pre endHeight")

}

func getZkCommitProofList(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proofId, _ := cmd.Flags().GetUint64("proofId")
	onChainProofId, _ := cmd.Flags().GetUint64("onChainProofId")
	onChain, _ := cmd.Flags().GetBool("onChain")
	latestProof, _ := cmd.Flags().GetBool("latestProof")
	end, _ := cmd.Flags().GetUint64("endHeight")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkFetchProofList{
		ProofId:         proofId,
		OnChainProofId:  onChainProofId,
		ReqOnChainProof: onChain,
		ReqLatestProof:  latestProof,
		EndHeight:       end,
	}

	params.FuncName = "GetProofList"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkCommitProof
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func nftCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nft",
		Short: "nft related cmd",
	}
	cmd.AddCommand(mintNFTCmd())
	cmd.AddCommand(transferNFTCmd())
	cmd.AddCommand(withdrawNFTCmd())
	cmd.AddCommand(getNftByIdCmd())
	cmd.AddCommand(getNftByHashCmd())

	return cmd
}

func mintNFTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "mint nft command",
		Run:   setMintNFT,
	}
	mintNFTFlag(cmd)
	return cmd
}

func mintNFTFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("creatorId", "f", 0, "NFT creator id")
	cmd.MarkFlagRequired("creatorId")

	cmd.Flags().Uint64P("recipientId", "t", 0, "NFT recipient id")
	cmd.MarkFlagRequired("recipientId")

	cmd.Flags().StringP("contentHash", "e", "", "NFT content hash,must 64 hex char")
	cmd.MarkFlagRequired("contentHash")

	cmd.Flags().Uint64P("protocol", "p", 1, "NFT protocol, 1:ERC1155, 2: ERC721")
	cmd.MarkFlagRequired("protocol")

	cmd.Flags().Uint64P("amount", "n", 1, "mint amount, only for ERC1155 case")

}

func setMintNFT(cmd *cobra.Command, args []string) {
	accountId, _ := cmd.Flags().GetUint64("creatorId")
	toId, _ := cmd.Flags().GetUint64("recipientId")
	contentHash, _ := cmd.Flags().GetString("contentHash")
	protocol, _ := cmd.Flags().GetUint64("protocol")
	amount, _ := cmd.Flags().GetUint64("amount")

	if protocol == zt.ZKERC721 && amount > 1 {
		fmt.Fprintln(os.Stderr, errors.Wrapf(types.ErrInvalidParam, "NFT erc721 only allow 1 amount"))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nft := &zt.ZkMintNFT{
		FromAccountId: accountId,
		RecipientId:   toId,
		ContentHash:   contentHash,
		ErcProtocol:   protocol,
		Amount:        amount,
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "MintNFT",
		Payload:    types.MustPBToJSON(nft),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func transferNFTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "transfer nft command",
		Run:   transferNFT,
	}
	transferNFTFlag(cmd)
	return cmd
}

func transferNFTFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("fromId", "a", 0, "NFT from id")
	cmd.MarkFlagRequired("fromId")

	cmd.Flags().Uint64P("toId", "t", 0, "NFT to id")
	cmd.MarkFlagRequired("toId")

	cmd.Flags().Uint64P("tokenId", "i", 0, "NFT token id")
	cmd.MarkFlagRequired("tokenId")

	cmd.Flags().Uint64P("amount", "n", 1, "NFT token id")
	cmd.MarkFlagRequired("amount")
}

func transferNFT(cmd *cobra.Command, args []string) {
	accountId, _ := cmd.Flags().GetUint64("fromId")
	toId, _ := cmd.Flags().GetUint64("toId")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nft := &zt.ZkTransferNFT{
		FromAccountId: accountId,
		RecipientId:   toId,
		NFTTokenId:    tokenId,
		Amount:        amount,
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "TransferNFT",
		Payload:    types.MustPBToJSON(nft),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func withdrawNFTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "withdraw to L1",
		Run:   withdrawNFT,
	}
	withdrawNFTFlag(cmd)
	return cmd
}

func withdrawNFTFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("fromId", "a", 0, "NFT from id")
	cmd.MarkFlagRequired("fromId")

	cmd.Flags().Uint64P("tokenId", "i", 0, "NFT token id")
	cmd.MarkFlagRequired("tokenId")

	cmd.Flags().Uint64P("amount", "n", 0, "amount")
	cmd.MarkFlagRequired("amount")
}

func withdrawNFT(cmd *cobra.Command, args []string) {
	accountId, _ := cmd.Flags().GetUint64("fromId")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nft := &zt.ZkWithdrawNFT{
		FromAccountId: accountId,
		NFTTokenId:    tokenId,
		Amount:        amount,
	}
	params := &rpctypes.CreateTxIn{
		Execer:     zt.Zksync,
		ActionName: "WithdrawNFT",
		Payload:    types.MustPBToJSON(nft),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func getNftByIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "get nft by id",
		Run:   getNftId,
	}
	getNftByIdFlag(cmd)
	return cmd
}

func getNftByIdFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("id", "i", 0, "nft token Id")
	cmd.MarkFlagRequired("id")
}

func getNftId(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	id, _ := cmd.Flags().GetUint64("id")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenId: id,
	}

	params.FuncName = "GetNFTStatus"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkNFTTokenStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getNftByHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash",
		Short: "get nft by hash",
		Run:   getNftHash,
	}
	getNftByHashFlag(cmd)
	return cmd
}

func getNftByHashFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("hash", "s", "", "nft content hash")
	cmd.MarkFlagRequired("hash")
}

func getNftHash(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	hash, _ := cmd.Flags().GetString("hash")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.ReqString{
		Data: hash,
	}

	params.FuncName = "GetNFTId"
	params.Payload = types.MustPBToJSON(req)

	var id types.Int64
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &id)
	ctx.Run()
}
