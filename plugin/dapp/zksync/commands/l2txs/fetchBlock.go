package l2txs

import (
	"context"
	"fmt"
	"strings"

	"github.com/33cn/chain33/rpc/grpcclient"

	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"

	//"gitlab.33.cn/zkrelayer/relayer/chain33/calcwitness"
	//"gitlab.33.cn/zkrelayer/relayer/common"
	chain33Ty "github.com/33cn/chain33/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io/ioutil"
	"time"
)

var (
	L2ActionType2nameMap = map[int]string{
		zt.TyNoopAction:           zt.NameNoopAction,
		zt.TyDepositAction:        zt.NameDepositAction,
		zt.TyWithdrawAction:       zt.NameWithdrawAction,
		zt.TyContractToTreeAction: zt.NameContractToTreeAction,
		zt.TyTreeToContractAction: zt.NameTreeToContractAction,
		zt.TyTransferAction:       zt.NameTransferAction,
		zt.TyTransferToNewAction:  zt.NameTransferToNewAction,
		zt.TyProxyExitAction:      zt.NameForceExitAction,
		zt.TySetPubKeyAction:      zt.NameSetPubKeyAction,
		zt.TyFullExitAction:       zt.NameFullExitAction,
		zt.TySwapAction:           zt.NameSwapAction,
		zt.TySetVerifyKeyAction:   zt.NameSetVerifyKeyAction,
		zt.TyCommitProofAction:    zt.NameCommitProofAction,
		zt.TySetVerifierAction:    zt.NameSetVerifierAction,
		zt.TySetFeeAction:         zt.NameSetFeeAction,
		zt.TySetTokenSymbolAction: zt.NameSetTokenSymbolAction,
		zt.TyMintNFTAction:        zt.NameMintNFTAction,
		zt.TyWithdrawNFTAction:    zt.NameWithdrawNFTACTION,
		zt.TyTransferNFTAction:    zt.NameTransferNFTAction,
	}
)

func NewMainChainClient(paraRemoteGrpcClient string) chain33Ty.Chain33Client {
	// paraChainGrpcRecSize 平行链receive最大100M
	const paraChainGrpcRecSize = 100 * 1024 * 1024
	if paraRemoteGrpcClient == "" {
		paraRemoteGrpcClient = "127.0.0.1:8802"
	}
	kp := keepalive.ClientParameters{
		Time:                time.Second * 5,
		Timeout:             time.Second * 20,
		PermitWithoutStream: true,
	}
	conn, err := grpc.Dial(grpcclient.NewMultipleURL(paraRemoteGrpcClient), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(paraChainGrpcRecSize)),
		grpc.WithKeepaliveParams(kp))
	if err != nil {
		return nil
	}
	grpcClient := chain33Ty.NewChain33Client(conn)
	return grpcClient
}

//该命令用于从区块链获取包含l2交易的区块
//./chain33-cli zksync sendl2 fetch -s 800 -e 6015 -p . --rpc_laddr 172.18.0.6:8902
func fetchL2BlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "fetch a L2 block specified by height ",
		Run:   fetchL2Block,
	}
	fetchL2BlockFlags(cmd)
	return cmd
}

func fetchL2BlockFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("start", "s", 0, "start block height")
	_ = cmd.MarkFlagRequired("start")

	cmd.Flags().Int64P("end", "e", 0, "end block height")
	_ = cmd.MarkFlagRequired("end")

	cmd.Flags().StringP("path", "p", "", "path to store the block")
	_ = cmd.MarkFlagRequired("path")
}

func fetchL2Block(cmd *cobra.Command, args []string) {
	grpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	start, _ := cmd.Flags().GetInt64("start")
	end, _ := cmd.Flags().GetInt64("end")
	path, _ := cmd.Flags().GetString("path")

	if end < start || end < 0 || start < 0 {
		fmt.Println("Wrong start or end input")
		return
	}

	grpClient := NewMainChainClient(grpcLaddr)

	descriptionFileName := path + fmt.Sprintf("/l2_blocks_%d_to_%d_description.txt", start, end)
	description := ""
	blocks := "L2 txs existed on blocks: "
	for height := start; height <= end; height++ {
		blockSeq, err := grpClient.GetBlockBySeq(context.Background(), &types.Int64{Data: height})
		if nil != err {
			return
		}
		txCntMap := make(map[int]int)
		for i := 0; i < zt.TyTransferNFTAction; i++ {
			txCntMap[i] = 0
		}

		for _, tx := range blockSeq.Detail.Block.Txs {
			if !strings.HasSuffix(string(tx.Execer), "zksync") {
				continue
			}
			var action zt.ZksyncAction
			if err := types.Decode(tx.Payload, &action); nil != err {
				return

			}

			switch action.Ty {
			case zt.TyDepositAction:
				txCntMap[zt.TyDepositAction] = txCntMap[zt.TyDepositAction] + 1
			case zt.TyWithdrawAction:
				txCntMap[zt.TyWithdrawAction] = txCntMap[zt.TyWithdrawAction] + 1
			case zt.TyTransferAction:
				txCntMap[zt.TyTransferAction] = txCntMap[zt.TyTransferAction] + 1
			case zt.TyTransferToNewAction:
				txCntMap[zt.TyTransferToNewAction] = txCntMap[zt.TyTransferToNewAction] + 1
			case zt.TyProxyExitAction:
				txCntMap[zt.TyProxyExitAction] = txCntMap[zt.TyProxyExitAction] + 1
			case zt.TySetPubKeyAction:
				txCntMap[zt.TySetPubKeyAction] = txCntMap[zt.TySetPubKeyAction] + 1
			case zt.TyFullExitAction:
				txCntMap[zt.TyFullExitAction] = txCntMap[zt.TyFullExitAction] + 1
			case zt.TySwapAction:
				txCntMap[zt.TySwapAction] = txCntMap[zt.TySwapAction] + 1
			case zt.TyContractToTreeAction:
				txCntMap[zt.TyContractToTreeAction] = txCntMap[zt.TyContractToTreeAction] + 1
			case zt.TyTreeToContractAction:
				txCntMap[zt.TyTreeToContractAction] = txCntMap[zt.TyTreeToContractAction] + 1
			case zt.TyMintNFTAction:
				txCntMap[zt.TyMintNFTAction] = txCntMap[zt.TyMintNFTAction] + 1
			case zt.TyWithdrawNFTAction:
				txCntMap[zt.TyWithdrawNFTAction] = txCntMap[zt.TyWithdrawNFTAction] + 1
			case zt.TyTransferNFTAction:
				txCntMap[zt.TyTransferNFTAction] = txCntMap[zt.TyTransferNFTAction] + 1
			case zt.TySetFeeAction:
				txCntMap[zt.TySetFeeAction] = txCntMap[zt.TySetFeeAction] + 1
			case zt.TySetTokenSymbolAction:
				txCntMap[zt.TySetTokenSymbolAction] = txCntMap[zt.TySetTokenSymbolAction] + 1
			}
		}

		existL2Tx := false
		for i := 0; i <= zt.TyAssetWithdrawAction; i++ {
			if txCntMap[i] != 0 {
				name := L2ActionType2nameMap[i]
				//fileName += fmt.Sprintf("%s_%d", name, txCntMap[i])
				description += fmt.Sprintf("%s_%d", name, txCntMap[i])
				existL2Tx = true
			}
		}
		if !existL2Tx {
			fmt.Println("No L2 txs in block with height", height)
			continue
		}
		description += fmt.Sprintf("_on_block_%d", blockSeq.Detail.Block.Height)
		blocks += fmt.Sprintf("%d,", blockSeq.Detail.Block.Height)
		fileName := path + fmt.Sprintf("/block_%d.data", height)
		data := types.Encode(blockSeq)
		writeToFile(fileName, data)
		description += "\n"
	}
	description += "\n" + blocks
	writeToFile(descriptionFileName, []byte(description))
}

func writeToFile(fileName string, data []byte) {
	err := ioutil.WriteFile(fileName, data, 0666)
	if err != nil {
		fmt.Println("Failed to write to file:", fileName)
	}
	fmt.Println("L2 block is written to file: ", fileName)
}
