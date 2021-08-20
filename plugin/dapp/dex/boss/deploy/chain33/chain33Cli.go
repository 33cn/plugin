package chain33

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/33cn/chain33/rpc/jsonclient"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

func Chain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain33",
		Short: "deploy to chain33",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		deployMulticallCmd(),
		deployERC20ContractCmd(),
		deployPancakeContractCmd(),
		farmCmd(),
	)
	return cmd
}

func deployMulticallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployMulticall",
		Short: "deploy Multicall",
		Run:   deployMulticallContract,
	}
	addDeployMulticallContractFlags(cmd)
	return cmd
}

func addDeployMulticallContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("caller", "c", "", "the caller address")
	cmd.MarkFlagRequired("caller")

	cmd.Flags().StringP("expire", "", "120s", "transaction expire time (optional)")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func deployMulticallContract(cmd *cobra.Command, args []string) {
	err := DeployMulticall(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// 创建ERC20合约
func deployERC20ContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployERC20",
		Short: "deploy ERC20 contract",
		Run:   deployERC20Contract,
	}
	addDeployERC20ContractFlags(cmd)
	return cmd
}

func addDeployERC20ContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("caller", "c", "", "the caller address")
	cmd.MarkFlagRequired("caller")
	cmd.Flags().StringP("name", "a", "", "REC20 name")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringP("symbol", "s", "", "REC20 symbol")
	cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("supply", "m", "", "REC20 supply")
	cmd.MarkFlagRequired("supply")

	cmd.Flags().StringP("expire", "", "120s", "transaction expire time (optional)")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func deployERC20Contract(cmd *cobra.Command, args []string) {
	err := DeployERC20(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func deployPancakeContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployPancake",
		Short: "deploy Pancake contract",
		Run:   deployPancakeContract,
	}
	addDeployPancakeContractFlags(cmd)
	return cmd
}

func addDeployPancakeContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("caller", "c", "", "the caller address")
	cmd.MarkFlagRequired("caller")

	cmd.Flags().StringP("expire", "", "120s", "transaction expire time (optional)")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
	cmd.Flags().StringP("parameter", "p", "", "construction contract parameter")
}

func deployPancakeContract(cmd *cobra.Command, args []string) {
	err := DeployPancake(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func createEvmTx(chainID int32, action proto.Message, execer, caller, addr, expire, rpcLaddr string, fee uint64) (string, error) {
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(action), Fee: 0, To: addr}

	tx.Fee = int64(1e7)
	if tx.Fee < int64(fee) {
		tx.Fee += int64(fee)
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	tx.ChainID = chainID
	txHex := types.Encode(tx)
	rawTx := hex.EncodeToString(txHex)

	unsignedTx := &types.ReqSignRawTx{
		Addr:   caller,
		TxHex:  rawTx,
		Expire: expire,
		Fee:    tx.Fee,
	}

	var res string
	client, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Println("createEvmTx::", "jsonclient.NewJSONClient failed due to:", err)
		return "", err
	}

	err = client.Call("Chain33.SignRawTx", unsignedTx, &res)
	if err != nil {
		fmt.Println("createEvmTx::", "Chain33.SignRawTx failed due to:", err)
		return "", err
	}

	return res, nil
}
