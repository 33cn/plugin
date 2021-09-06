package chain33

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-swap-periphery/src/pancakeFactory"

	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"

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
		setFeeToCmd(),
		farmCmd(),
	)
	return cmd
}

func setFeeToCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setFeeTo",
		Short: "set FeeTo address to factory",
		Run:   setFeeTo,
	}
	addSetFeeToFlags(cmd)
	return cmd
}

func addSetFeeToFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("factory", "f", "", "factory Addr ")
	_ = cmd.MarkFlagRequired("factory")

	cmd.Flags().StringP("feeTo", "t", "", "feeTo Addr ")
	_ = cmd.MarkFlagRequired("feeTo")
	cmd.Flags().StringP("key", "k", "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944", "private key of feetoSetter")
}

func setFeeTo(cmd *cobra.Command, args []string) {
	factroy, _ := cmd.Flags().GetString("factory")
	feeTo, _ := cmd.Flags().GetString("feeTo")

	chainID, _ := cmd.Flags().GetInt32("chainID")
	caller, _ := cmd.Flags().GetString("caller")
	paraName, _ := cmd.Flags().GetString("paraName")
	expire, _ := cmd.Flags().GetString("expire")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	fee, _ := cmd.Flags().GetFloat64("fee")
	feeInt64 := uint64(fee*1e4) * 1e4

	//function add(uint256 _allocPoint, IBEP20 _lpToken, bool _withUpdate) public onlyOwner
	parameter := fmt.Sprintf("setFeeTo(%s)", feeTo)
	_, packData, err := evmAbi.Pack(parameter, pancakeFactory.PancakeFactoryABI, true)
	if nil != err {
		fmt.Println("AddPool2FarmHandle", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	action := evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: parameter, Para: packData, ContractAddr: factroy}

	data, err := createEvmTx(chainID, &action, paraName+"evm", caller, factroy, expire, rpcLaddr, feeInt64)
	if err != nil {
		fmt.Println("AddPool2FarmHandle", "Failed to do createEvmTx due to:", err.Error())
		return
	}

	txhex, err := sendTransactionRpc(data, rpcLaddr)
	if err != nil {
		fmt.Println("AddPool2FarmHandle", "Failed to do sendTransactionRpc due to:", err.Error())
		return
	}
	fmt.Println(txhex)
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
