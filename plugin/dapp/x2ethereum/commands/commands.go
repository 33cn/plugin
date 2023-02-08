/*Package commands implement dapp client commands*/
package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/33cn/chain33/rpc/jsonclient"
	types2 "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/common/commands"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebcli/buildflags"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/utils"
	types3 "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

// Cmd x2ethereum client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "x2ethereum",
		Short: "x2ethereum command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		CreateRawWithdrawChain33TxCmd(),
		CreateRawChain33ToEthTxCmd(),
		CreateRawAddValidatorTxCmd(),
		CreateRawRemoveValidatorTxCmd(),
		CreateRawModifyValidatorTxCmd(),
		CreateRawSetConsensusTxCmd(),
		CreateTransferCmd(),
		CreateTokenTransferExecCmd(),
		CreateTokenWithdrawCmd(),
		queryCmd(),
		queryRelayerBalanceCmd(),
	)

	if buildflags.NodeAddr == "" {
		buildflags.NodeAddr = "http://127.0.0.1:7545"
	}
	cmd.PersistentFlags().String("node_addr", buildflags.NodeAddr, "eth node url")
	return cmd
}

//CreateRawWithdrawChain33TxCmd Burn
func CreateRawWithdrawChain33TxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn",
		Short: "Create a burn tx in chain33,withdraw chain33ToEth",
		Run:   burn,
	}

	addChain33ToEthFlags(cmd)

	return cmd
}

func addChain33ToEthFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("contract", "q", "", "token contract address,nil for ETH")

	cmd.Flags().StringP("symbol", "t", "", "token symbol in chain33,coins.bty etc.")
	_ = cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("receiver", "r", "", "ethereum receiver address")
	_ = cmd.MarkFlagRequired("cExec")

	cmd.Flags().Float64P("amount", "a", float64(0), "the amount of this contract want to lock")
	_ = cmd.MarkFlagRequired("amount")
}

func burn(cmd *cobra.Command, args []string) {
	contract, _ := cmd.Flags().GetString("contract")
	csymbol, _ := cmd.Flags().GetString("symbol")
	receiver, _ := cmd.Flags().GetString("receiver")
	amount, _ := cmd.Flags().GetFloat64("amount")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	if contract == "" {
		contract = "0x0000000000000000000000000000000000000000"
	}

	decimal, err := utils.GetDecimalsFromNode(contract, nodeAddr)
	if err != nil {
		fmt.Println("get decimal error")
		return
	}

	params := &types3.Chain33ToEth{
		TokenContract:    contract,
		EthereumReceiver: receiver,
		Amount:           types3.TrimZeroAndDot(strconv.FormatFloat(amount*1e8, 'f', 4, 64)),
		IssuerDotSymbol:  csymbol,
		Decimals:         decimal,
	}

	payLoad := types.MustPBToJSON(params)

	createTx(cmd, payLoad, types3.NameWithdrawChain33Action)
}

//CreateRawChain33ToEthTxCmd Lock
func CreateRawChain33ToEthTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Create a lock tx in chain33,create a chain33ToEth tx",
		Run:   lock,
	}

	addChain33ToEthFlags(cmd)

	return cmd
}

func lock(cmd *cobra.Command, args []string) {
	contract, _ := cmd.Flags().GetString("contract")
	csymbol, _ := cmd.Flags().GetString("symbol")
	receiver, _ := cmd.Flags().GetString("receiver")
	amount, _ := cmd.Flags().GetFloat64("amount")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	decimal, err := utils.GetDecimalsFromNode(contract, nodeAddr)
	if err != nil {
		fmt.Println("get decimal error")
		return
	}

	if contract == "" {
		fmt.Println("get token address error")
		return
	}

	params := &types3.Chain33ToEth{
		TokenContract:    contract,
		EthereumReceiver: receiver,
		Amount:           strconv.FormatFloat(amount*1e8, 'f', 4, 64),
		IssuerDotSymbol:  csymbol,
		Decimals:         decimal,
	}

	payLoad := types.MustPBToJSON(params)

	createTx(cmd, payLoad, types3.NameChain33ToEthAction)
}

//CreateTransferCmd Transfer
func CreateTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Create a transfer tx in chain33",
		Run:   transfer,
	}

	addTransferFlags(cmd)
	return cmd
}

func transfer(cmd *cobra.Command, args []string) {
	commands.CreateAssetTransfer(cmd, args, types3.X2ethereumX)
}

func addTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("to", "t", "", "receiver account address")
	_ = cmd.MarkFlagRequired("to")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info,optional")

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")

}

// CreateTokenTransferExecCmd create raw transfer tx
func CreateTokenTransferExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send_exec",
		Short: "Create a token send to executor transaction",
		Run:   createTokenSendToExec,
	}
	addCreateTokenSendToExecFlags(cmd)
	return cmd
}

func addCreateTokenSendToExecFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "receiver executor address")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")
}

func createTokenSendToExec(cmd *cobra.Command, args []string) {
	commands.CreateAssetSendToExec(cmd, args, types3.X2ethereumX)
}

// CreateTokenWithdrawCmd create raw withdraw tx
func CreateTokenWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdrawfromexec",
		Short: "Create a token withdraw transaction",
		Run:   createTokenWithdraw,
	}
	addCreateTokenWithdrawFlags(cmd)
	return cmd
}

func addCreateTokenWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "execer withdrawn from")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().Float64P("amount", "a", 0, "withdraw amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")
}

func createTokenWithdraw(cmd *cobra.Command, args []string) {
	commands.CreateAssetWithdraw(cmd, args, types3.X2ethereumX)
}

//CreateRawAddValidatorTxCmd AddValidator
func CreateRawAddValidatorTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a add validator tx in chain33",
		Run:   addValidator,
	}

	addValidatorFlags(cmd)
	cmd.Flags().Int64P("power", "p", 0, "validator power set,must be 1-100")
	_ = cmd.MarkFlagRequired("power")
	return cmd
}

func addValidatorFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "the address you want to add/remove/modify as a validator ")
	_ = cmd.MarkFlagRequired("address")
}

func addValidator(cmd *cobra.Command, args []string) {
	address, _ := cmd.Flags().GetString("address")
	power, _ := cmd.Flags().GetInt64("power")

	params := &types3.MsgValidator{
		Address: address,
		Power:   power,
	}

	payLoad := types.MustPBToJSON(params)

	createTx(cmd, payLoad, types3.NameAddValidatorAction)
}

//CreateRawRemoveValidatorTxCmd RemoveValidator
func CreateRawRemoveValidatorTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Create a remove validator tx in chain33",
		Run:   removeValidator,
	}

	addValidatorFlags(cmd)
	return cmd
}

func removeValidator(cmd *cobra.Command, args []string) {
	address, _ := cmd.Flags().GetString("address")

	params := &types3.MsgValidator{
		Address: address,
	}

	payLoad := types.MustPBToJSON(params)

	createTx(cmd, payLoad, types3.NameRemoveValidatorAction)
}

//CreateRawModifyValidatorTxCmd ModifyValidator
func CreateRawModifyValidatorTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Create a modify validator tx in chain33",
		Run:   modify,
	}

	addValidatorFlags(cmd)

	cmd.Flags().Int64P("power", "p", 0, "validator power set,must be 1-100")
	_ = cmd.MarkFlagRequired("power")
	return cmd
}

func modify(cmd *cobra.Command, args []string) {
	address, _ := cmd.Flags().GetString("address")
	power, _ := cmd.Flags().GetInt64("power")

	params := &types3.MsgValidator{
		Address: address,
		Power:   power,
	}

	payLoad := types.MustPBToJSON(params)

	createTx(cmd, payLoad, types3.NameModifyPowerAction)
}

//CreateRawSetConsensusTxCmd MsgSetConsensusNeeded
func CreateRawSetConsensusTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setconsensus",
		Short: "Create a set consensus threshold tx in chain33",
		Run:   setConsensus,
	}

	addSetConsensusFlags(cmd)
	return cmd
}

func addSetConsensusFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("power", "p", 0, "the power you want to set consensus need,must be 1-100")
	_ = cmd.MarkFlagRequired("power")
}

func setConsensus(cmd *cobra.Command, args []string) {
	power, _ := cmd.Flags().GetInt64("power")

	params := &types3.MsgConsensusThreshold{
		ConsensusThreshold: power,
	}

	payLoad := types.MustPBToJSON(params)

	createTx(cmd, payLoad, types3.NameSetConsensusThresholdAction)
}

func queryRelayerBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "query balance of x2ethereum",
		Run:   queryRelayerBalance,
	}

	cmd.Flags().StringP("token", "t", "", "token symbol")
	_ = cmd.MarkFlagRequired("token")

	cmd.Flags().StringP("address", "s", "", "the address you want to query")
	_ = cmd.MarkFlagRequired("address")

	cmd.Flags().StringP("tokenaddress", "a", "", "token address,nil for all this token symbol")
	return cmd
}

func queryRelayerBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	token, _ := cmd.Flags().GetString("token")
	address, _ := cmd.Flags().GetString("address")
	contract, _ := cmd.Flags().GetString("tokenaddress")

	get := &types3.QueryRelayerBalance{
		TokenSymbol: token,
		Address:     address,
		TokenAddr:   contract,
	}

	payLoad, err := types.PBToJSON(get)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}

	query := types2.Query4Jrpc{
		Execer:   types3.X2ethereumX,
		FuncName: types3.FuncQueryRelayerBalance,
		Payload:  payLoad,
	}

	channel := &types3.ReceiptQueryRelayerBalance{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", query, channel)
	ctx.Run()
}

func createTx(cmd *cobra.Command, payLoad []byte, action string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	pm := &types2.CreateTxIn{
		Execer:     types.GetExecName(types3.X2ethereumX, paraName),
		ActionName: action,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}
