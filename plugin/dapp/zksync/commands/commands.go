/*Package commands implement dapp client commands*/
package commands

import (
	"fmt"
	"os"

	commandtypes "github.com/33cn/chain33/system/dapp/commands/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
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
		contractToLeafCmd(),
		leafToContractCmd(),
		transferCmd(),
		transferToNewCmd(),
		forceExitCmd(),
		setPubKeyCmd(),
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
	cmd.Flags().Int32P("tokenId", "t", 1, "deposit tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("amount", "a", 0, "deposit amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "", "deposit ethaddress")
	cmd.MarkFlagRequired("ethAddress")
	cmd.Flags().StringP("chain33Addr", "c", "", "deposit chain33Addr")
	cmd.MarkFlagRequired("chain33Addr")

}

func deposit(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyDepositAction, tokenId, amount, ethAddress, "", chain33Addr, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
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
	cmd.Flags().Int32P("tokenId", "t", 1, "withdraw tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("amount", "a", 0, "withdraw amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "", "withdraw ethAddress")
	cmd.MarkFlagRequired("ethAddress")

}

func withdraw(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyWithdrawAction, tokenId, amount, ethAddress, "", "", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
}

func leafToContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leafToContract",
		Short: "get leafToContract tx",
		Run:   leafToContract,
	}
	leafToContractFlag(cmd)
	return cmd
}

func leafToContractFlag(cmd *cobra.Command) {
	cmd.Flags().Int32P("tokenId", "t", 1, "leafToContract tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("amount", "a", 0, "leafToContract amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "0", "leafToContract ethAddress")
	cmd.MarkFlagRequired("ethAddress")

}

func leafToContract(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyLeafToContractAction, tokenId, amount, ethAddress, "", "", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
}

func contractToLeafCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contractToLeaf",
		Short: "get contractToLeaf tx",
		Run:   contractToLeaf,
	}
	contractToLeafFlag(cmd)
	return cmd
}

func contractToLeafFlag(cmd *cobra.Command) {
	cmd.Flags().Int32P("tokenId", "t", 1, "contractToLeaf tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("amount", "a", 0, "contractToLeaf amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "", "contractToLeaf ethAddress")
	cmd.MarkFlagRequired("ethAddress")

}

func contractToLeaf(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyContractToLeafAction, tokenId, amount, ethAddress, "", "", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
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
	cmd.Flags().Int32P("tokenId", "t", 1, "transfer tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("amount", "a", 0, "transfer amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "", "transfer fromEthAddress")
	cmd.MarkFlagRequired("ethAddress")
	cmd.Flags().StringP("toEthAddress", "te", "", "transfer toEthAddress")
	cmd.MarkFlagRequired("toEthAddress")

}

func transfer(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")
	toEthAddress, _ := cmd.Flags().GetString("toEthAddress")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyTransferAction, tokenId, amount, ethAddress, toEthAddress, "", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
}

func transferToNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transferToNew",
		Short: "get transferToNew tx",
		Run:   transferToNew,
	}
	transferToNewFlag(cmd)
	return cmd
}

func transferToNewFlag(cmd *cobra.Command) {
	cmd.Flags().Int32P("tokenId", "t", 1, "transferToNew tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("amount", "a", 0, "transferToNew amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "", "transferToNew fromEthAddress")
	cmd.MarkFlagRequired("ethAddress")
	cmd.Flags().StringP("toEthAddress", "te", "", "transferToNew toEthAddress")
	cmd.MarkFlagRequired("toEthAddress")
	cmd.Flags().StringP("chain33Addr", "c", "", "transferToNew chain33Addr")
	cmd.MarkFlagRequired("chain33Addr")
}

func transferToNew(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")
	toEthAddress, _ := cmd.Flags().GetString("toEthAddress")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyTransferToNewAction, tokenId, amount, ethAddress, toEthAddress, chain33Addr, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
}

func forceExitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forceExit",
		Short: "get forceExit tx",
		Run:   forceExit,
	}
	forceExitFlag(cmd)
	return cmd
}

func forceExitFlag(cmd *cobra.Command) {
	cmd.Flags().Int32P("tokenId", "t", 1, "forceExit tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("ethAddress", "e", "", "forceExit ethAddress")
	cmd.MarkFlagRequired("ethAddress")

}

func forceExit(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyForceExitAction, tokenId, 0, ethAddress, "", "", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
}

func setPubKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setPubKey",
		Short: "get setPubKey tx",
		Run:   setPubKey,
	}
	setPubKeyFlag(cmd)
	return cmd
}

func setPubKeyFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("ethAddress", "e", "", "setPubKeyFlag ethAddress")
	cmd.MarkFlagRequired("ethAddress")

}

func setPubKey(cmd *cobra.Command, args []string) {
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyForceExitAction, 0, 0, ethAddress, "", "", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
}
