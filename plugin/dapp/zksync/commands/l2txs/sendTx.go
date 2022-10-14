package l2txs

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/spf13/cobra"
)

func SendChain33L2TxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sendl2",
		Short: "send l2 tx to chain33 ",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		sendDepositTxCmd(),
		batchSendDepositTxCmd(),
		sendWithdrawTxCmd(),
		BatchSendTransferTxCmd(),
		SendTransferTxCmd(),
		sendManyDepositTxCmd(),
		sendManyWithdrawTxCmd(),
		treeManyToContractCmd(),
		contractManyToTreeCmd(),
		SendManyTransferTxCmd(),
		SendManyTransferTxFromOneCmd(),
		transferManyToNewCmd(),
		transferToNewManyCmd(),
		proxyManyExitCmd(),
		nftManyCmd(),
		setManyPubKeyCmd(),
		fetchL2BlockCmd(),
	)

	return cmd
}

func sendTx(rpcLaddr string, tx *types.Transaction) {
	txData := types.Encode(tx)
	dataStr := common.ToHex(txData)

	//fmt.Println("sendTx", "dataStr", dataStr)
	params := rpctypes.RawParm{
		Token: "BTY",
		Data:  dataStr,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.SendTransaction", params, nil)
	ctx.RunWithoutMarshal()
}
