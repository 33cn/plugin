/*Package commands implement dapp client commands*/
package commands

import (
	"fmt"
	commandtypes "github.com/33cn/chain33/system/dapp/commands/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
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
	cmd.Flags().StringP("ethAddress", "e", "", "ethaddress")
	cmd.MarkFlagRequired("ethAddress")

}

func deposit(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyDepositAction, tokenId, amount, ethAddress, 0, 0, "", cfg)
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
	cmd.Flags().Int32P("accountId", "ac", 0, "withdraw accountId")
	cmd.MarkFlagRequired("accountId")

}

func withdraw(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	accountId, _ := cmd.Flags().GetInt32("accountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyWithdrawAction, tokenId, amount, "", accountId, 0, "", cfg)
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
	cmd.Flags().Int32P("accountId", "ac", 0, "leafToContract accountId")
	cmd.MarkFlagRequired("accountId")

}

func leafToContract(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	accountId, _ := cmd.Flags().GetInt32("accountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyLeafToContractAction, tokenId, amount, "", accountId, 0, "", cfg)
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
	cmd.Flags().Int32P("accountId", "ac", 0, "contractToLeaf accountId")
	cmd.MarkFlagRequired("accountId")

}

func contractToLeaf(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	accountId, _ := cmd.Flags().GetInt32("accountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyContractToLeafAction, tokenId, amount, "", accountId, 0, "", cfg)
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
	cmd.Flags().Int32P("accountId", "ac", 0, "transfer fromAccountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().Int32P("toAccountId", "ta", 0, "transfer toAccountId")
	cmd.MarkFlagRequired("toAccountId")

}

func transfer(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	accountId, _ := cmd.Flags().GetInt32("accountId")
	toAccountId, _ := cmd.Flags().GetInt32("toAccountId")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyTransferAction, tokenId, amount, "", accountId, toAccountId, "", cfg)
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
	cmd.Flags().Int32P("accountId", "ac", 0, "transferToNew fromAccountId")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().StringP("toEthAddress", "te", "", "transferToNew toEthAddress")
	cmd.MarkFlagRequired("toEthAddress")

}

func transferToNew(cmd *cobra.Command, args []string) {
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")
	accountId, _ := cmd.Flags().GetInt32("accountId")
	toEthAddress, _ := cmd.Flags().GetString("toEthAddress")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	txHex, err := wallet.CreateRawTx(zt.TyTransferToNewAction, tokenId, amount, "", accountId, 0, toEthAddress, cfg)
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
	txHex, err := wallet.CreateRawTx(zt.TyForceExitAction, tokenId, 0, ethAddress, 0, 0, "", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "createRawTx"))
		return
	}
	fmt.Println(txHex)
}


