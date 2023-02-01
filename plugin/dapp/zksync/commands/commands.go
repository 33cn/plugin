/*Package commands implement dapp client commands*/
package commands

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/33cn/plugin/plugin/dapp/zksync/commands/l2txs"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/common/commands"
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
		layer2Cmd(),
		contractCmd(),
		queryCmd(),
		//NFT
		nftCmd(),
		//batch send command
		l2txs.SendChain33L2TxCmd(),
	)
	return cmd
}

func layer2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "l2",
		Short: "layer2 related cmd",
	}
	cmd.AddCommand(
		depositCmd(),
		withdrawCmd(),
		contractToTreeCmd(),
		treeToContractCmd(),
		transferCmd(),
		transferToNewCmd(),
		proxyExitCmd(),
		setPubKeyCmd(),
		fullExitCmd(),
		setVerifyKeyCmd(),
		setOperatorCmd(),
		getChain33AddrCmd(),
		setTokenFeeCmd(),
		setTokenSymbolCmd(),
		setExodusModeCmd(),
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

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	deposit := &zt.ZkDeposit{
		TokenId:      tokenId,
		Amount:       amount,
		EthAddress:   ethAddress,
		Chain33Addr:  chain33Addr,
		L1PriorityId: int64(queueId),
	}
	params := &rpctypes.CreateTxIn{
		Execer:     commands.GetRealExecName(paraName, zt.Zksync),
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

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyWithdrawAction, tokenId, amount, "", "", accountId, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
		ActionName: "ZkWithdraw",
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
	cmd.Flags().Uint64P("tokenId", "t", 1, "token Id,eth=0")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "a", "0", "token self decimal amount, like 1 eth fill 1e18")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint64P("accountId", "i", 0, "treeToContract accountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().StringP("exec", "x", "", "to contract exec, default nil to zksync self")
}

func treeToContract(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	exec, _ := cmd.Flags().GetString("exec")

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	leafToContract := &zt.ZkTreeToContract{
		TokenId:   tokenId,
		Amount:    amount,
		AccountId: accountId,
		ToAcctId:  zt.SystemTree2ContractAcctId,
		ToExec:    exec,
	}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
		ActionName: "TreeToContract",
		Payload:    types.MustPBToJSON(leafToContract),
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
	cmd.Flags().StringP("tokenSymbol", "t", "", "token symbol asset")
	cmd.MarkFlagRequired("tokenSymbol")
	cmd.Flags().StringP("amount", "a", "0", "chain33 side decimal amount,default decimal 8, like 1eth fill 1e8")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint64P("accountId", "i", 0, "contractToTree to accountId")
	cmd.Flags().StringP("ethAddr", "e", "", "to eth addr")
	cmd.Flags().StringP("layer2Addr", "l", "", "to layer2 addr")
	cmd.Flags().StringP("exec", "x", "", "from contract exec")
}

func contractToTree(cmd *cobra.Command, args []string) {
	tokenSymbol, _ := cmd.Flags().GetString("tokenSymbol")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	ethAddr, _ := cmd.Flags().GetString("ethAddr")
	layer2Addr, _ := cmd.Flags().GetString("layer2Addr")
	exec, _ := cmd.Flags().GetString("exec")

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	contractToLeaf := &zt.ZkContractToTree{
		TokenSymbol:  tokenSymbol,
		Amount:       amount,
		ToAccountId:  accountId,
		ToEthAddr:    ethAddr,
		ToLayer2Addr: layer2Addr,
		FromExec:     exec,
	}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
		ActionName: "ContractToTree",
		Payload:    types.MustPBToJSON(contractToLeaf),
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
	cmd.Flags().Uint64P("tokenId", "i", 1, "transfer tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "a", "0", "transfer amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint64P("accountId", "f", 0, "transfer fromAccountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().Uint64P("toAccountId", "t", 0, "transfer toAccountId")
	cmd.MarkFlagRequired("toAccountId")
}

func transfer(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	toAccountId, _ := cmd.Flags().GetUint64("toAccountId")

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyTransferAction, tokenId, amount, "", "", accountId, toAccountId)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
		ActionName: "ZkTransfer",
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
	cmd.Flags().Uint64P("accountId", "f", 0, "transferToNew fromAccountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().StringP("ethAddress", "e", "", "transferToNew toEthAddress")
	cmd.MarkFlagRequired("ethAddress")
	cmd.Flags().StringP("chain33Addr", "c", "", "transferToNew toChain33Addr")
	cmd.MarkFlagRequired("chain33Addr")
}

func transferStr2Int(s string, base int) (*big.Int, error) {
	s = zt.FilterHexPrefix(s)
	v, ok := new(big.Int).SetString(s, base)
	if !ok {
		return nil, errors.New(fmt.Sprintf("transferStr2Int s=%s,base=%d", s, base))
	}
	return v, nil
}

func transferToNew(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	toEthAddress, _ := cmd.Flags().GetString("ethAddress")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")

	ethAddrBigInt, err := transferStr2Int(toEthAddress, 16)
	if nil != err {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	toEthAddress = ethAddrBigInt.Text(16) //没有前缀0x

	chain33AddrBigInt, err := transferStr2Int(chain33Addr, 16)
	if nil != err {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	chain33Addr = chain33AddrBigInt.Text(16)

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyTransferToNewAction, tokenId, amount, toEthAddress, chain33Addr, accountId, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
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

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyProxyExitAction, tokenId, "0", "", "", accountId, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
		ActionName: "ForceExit",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func proxyExitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxyexit",
		Short: "withdraw by other addr",
		Run:   proxyExit,
	}
	proxyExitFlag(cmd)
	return cmd
}

func proxyExitFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "i", 1, "target tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("accountId", "a", 0, "proxy accountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().Uint64P("toId", "t", 0, "target accountId")
	cmd.MarkFlagRequired("toId")
	cmd.Flags().StringP("maker", "p", "0", "from account fee")
	cmd.Flags().StringP("taker", "q", "0", "to account fee")

}

func getRealExecName(paraName string, name string) string {
	if strings.HasPrefix(name, pt.ParaPrefix) {
		return name
	}
	return paraName + name
}

func proxyExit(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	toId, _ := cmd.Flags().GetUint64("toId")

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	payload, err := wallet.CreateRawTx(zt.TyProxyExitAction, tokenId, "0", "", "", accountId, toId)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
		ActionName: "ProxyExit",
		Payload:    payload,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func setPubKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setpubkey",
		Short: "set layer2 account's pubkey",
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

	paraName, _ := cmd.Flags().GetString("paraName")
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
		Execer:     getRealExecName(paraName, zt.Zksync),
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

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	fullExit := &zt.ZkFullExit{
		TokenId:            tokenId,
		AccountId:          accountId,
		EthPriorityQueueId: int64(queueId),
	}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
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

	seed, err := wallet.GetLayer2PrivateKeySeed(privateKeyString, "", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "eddsa.GetLayer2PrivateKeySeed"))
		return
	}
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(seed))
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

func setTokenFeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee",
		Short: "set operation fee",
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
	cmd.Flags().Int32P("action", "a", 0, "action ty,2:withdraw,3:transfer,4:transfer2new,5:proxyExit,9:contract2tree,10:tree2contract")
	cmd.MarkFlagRequired("action")
}

func setTokenFee(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	fee, _ := cmd.Flags().GetString("fee")
	action, _ := cmd.Flags().GetInt32("action")
	paraName, _ := cmd.Flags().GetString("paraName")

	payload := &zt.ZkSetFee{
		TokenId:  tokenId,
		Amount:   fee,
		ActionTy: action,
	}

	params := &rpctypes.CreateTxIn{
		Execer:     commands.GetRealExecName(paraName, zt.Zksync),
		ActionName: "SetFee",
		Payload:    types.MustPBToJSON(payload),
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func contractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "zksync contract asset related cmd",
	}
	cmd.AddCommand(
		CreateRawTransferCmd(),
		CreateRawTransferToExecCmd(),
		CreateRawWithdrawCmd(),
	)

	return cmd
}

//CreateRawTransferCmd  create raw transfer tx
func CreateRawTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Create a transfer transaction",
		Run:   createTransfer,
	}
	addCreateTransferFlags(cmd)
	return cmd
}

func addCreateTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("to", "t", "", "receiver account address")
	_ = cmd.MarkFlagRequired("to")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "asset symbol in layer2")
	_ = cmd.MarkFlagRequired("symbol")
}

func createTransfer(cmd *cobra.Command, args []string) {
	commands.CreateAssetTransfer(cmd, args, zt.Zksync)
}

//CreateRawTransferToExecCmd create raw transfer to exec tx
func CreateRawTransferToExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer_exec",
		Short: "Create a transfer to exec transaction",
		Run:   createTransferToExec,
	}
	addCreateTransferToExecFlags(cmd)
	return cmd
}

func addCreateTransferToExecFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "asset symbol in layer2")
	_ = cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("exec", "e", "", "asset deposit exec")
	_ = cmd.MarkFlagRequired("exec")
}

func createTransferToExec(cmd *cobra.Command, args []string) {
	commands.CreateAssetSendToExec(cmd, args, zt.Zksync)
}

//CreateRawWithdrawCmd create raw withdraw tx
func CreateRawWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "Create a withdraw transaction",
		Run:   createWithdraw,
	}
	addCreateWithdrawFlags(cmd)
	return cmd
}

func addCreateWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("amount", "a", 0, "withdraw amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "asset symbol in layer2")
	_ = cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("exec", "e", "", "asset deposit exec")
	_ = cmd.MarkFlagRequired("exec")
}

func createWithdraw(cmd *cobra.Command, args []string) {
	commands.CreateAssetWithdraw(cmd, args, zt.Zksync)
}

func setTokenSymbolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "symbol",
		Short: "set token symbol",
		Run:   setTokenSymbol,
	}
	setTokenSymbolFlag(cmd)
	return cmd
}

func setTokenSymbolFlag(cmd *cobra.Command) {
	cmd.Flags().Uint32P("tokenId", "t", 0, "token id")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("symbol", "s", "", "symbol")
	cmd.MarkFlagRequired("symbol")
	cmd.Flags().Uint32P("decimal", "d", 18, "token decimal")
	cmd.MarkFlagRequired("decimal")
}

func setTokenSymbol(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint32("tokenId")
	symbol, _ := cmd.Flags().GetString("symbol")
	decimal, _ := cmd.Flags().GetUint32("decimal")
	paraName, _ := cmd.Flags().GetString("paraName")

	payload := &zt.ZkTokenSymbol{
		Id:      strconv.Itoa(int(tokenId)),
		Symbol:  symbol,
		Decimal: decimal,
	}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
		ActionName: "SetTokenSymbol",
		Payload:    types.MustPBToJSON(payload),
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func setExodusModeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_exodus",
		Short: "set exodus mode",
		Run:   setExodusMode,
	}
	setExodusModeFlag(cmd)
	return cmd
}

func setExodusModeFlag(cmd *cobra.Command) {
	cmd.Flags().Uint32P("mode", "m", 0, "0:invalid,1:normal,2:pause,3:exodus prepare,4:final")
	cmd.MarkFlagRequired("mode")

	cmd.Flags().Uint64P("proofId", "i", 0, "final mode, last success proofId on L1")
	cmd.Flags().Uint32P("knownGap", "s", 0, "final mode, manager known balance gap if any,1:known,0:default")

}

func setExodusMode(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	mode, _ := cmd.Flags().GetUint32("mode")
	proofId, _ := cmd.Flags().GetUint64("proofId")
	knownGap, _ := cmd.Flags().GetUint32("knownGap")

	payload := &zt.ZkExodusMode{
		Mode: mode,
	}
	if mode == zt.ExodusFinalMode {
		if proofId == 0 {
			fmt.Fprintln(os.Stderr, "final mode,proofId should > 0")
			return
		}
		payload.Value = &zt.ZkExodusMode_Rollback{Rollback: &zt.ZkExodusRollbackModeParm{
			LastSuccessProofId: proofId,
			KnownBalanceGap:    knownGap,
		}}
	}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, zt.Zksync),
		ActionName: "SetExodusMode",
		Payload:    types.MustPBToJSON(payload),
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}
