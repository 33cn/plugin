package chain33

import (
	"fmt"
	"os"

	"github.com/33cn/chain33/common/address"
	chain33Types "github.com/33cn/chain33/types"

	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/masterChef"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/syrupBar"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/spf13/cobra"
)

func farmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "farm",
		Short: "deploy farm and set lp, transfer ownership",
	}
	cmd.AddCommand(
		deployFarmContractCmd(),
		TransferOwnerShipCmd(),
		AddPoolCmd(),
		UpdateAllocPointCmd(),
	)
	return cmd
}

func deployFarmContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy Farm contract",
		Run:   deployFarmContract,
	}
	addDeployFarmContractFlags(cmd)
	return cmd
}

func addDeployFarmContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("caller", "c", "", "the caller address")
	cmd.MarkFlagRequired("caller")

	cmd.Flags().StringP("expire", "", "120s", "transaction expire time (optional)")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func deployFarmContract(cmd *cobra.Command, args []string) {
	err := DeployFarm(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func AddPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add pool to farm, should transfer ownership first",
		Run:   AddPool2Farm,
	}

	addAddPoolCmdFlags(cmd)

	return cmd
}

func addAddPoolCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("masterchef", "m", "", "master Chef Addr ")
	_ = cmd.MarkFlagRequired("masterchef")

	cmd.Flags().StringP("lptoken", "l", "", "lp Addr ")
	_ = cmd.MarkFlagRequired("lptoken")

	cmd.Flags().Int64P("alloc", "p", 0, "allocation point ")
	_ = cmd.MarkFlagRequired("alloc")

	cmd.Flags().BoolP("update", "u", true, "with update")
	_ = cmd.MarkFlagRequired("update")

	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
	cmd.MarkFlagRequired("fee")

	cmd.Flags().StringP("caller", "c", "", "caller address")
	_ = cmd.MarkFlagRequired("caller")
	cmd.Flags().StringP("expire", "e", "120s", "transaction expire time (optional)")
}

func AddPool2Farm(cmd *cobra.Command, args []string) {
	masterChefAddrStr, _ := cmd.Flags().GetString("masterchef")
	allocPoint, _ := cmd.Flags().GetInt64("alloc")
	lpToken, _ := cmd.Flags().GetString("lptoken")
	update, _ := cmd.Flags().GetBool("update")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	caller, _ := cmd.Flags().GetString("caller")
	paraName, _ := cmd.Flags().GetString("paraName")
	expire, _ := cmd.Flags().GetString("expire")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	fee, _ := cmd.Flags().GetFloat64("fee")
	feeInt64 := uint64(fee*1e4) * 1e4

	//function add(uint256 _allocPoint, IBEP20 _lpToken, bool _withUpdate) public onlyOwner
	parameter := fmt.Sprintf("add(%d, %s, %v)", allocPoint, lpToken, update)
	_, packData, err := evmAbi.Pack(parameter, masterChef.MasterChefABI, false)
	if nil != err {
		fmt.Println("AddPool2FarmHandle", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	exector := chain33Types.GetExecName("evm", paraName)
	action := evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: parameter, Para: packData, ContractAddr: address.ExecAddress(exector)}

	data, err := createEvmTx(chainID, &action, exector, caller, masterChefAddrStr, expire, rpcLaddr, feeInt64)
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

func UpdateAllocPointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the given pool's CAKE allocation point, should transfer ownership first",
		Run:   UpdateAllocPoint,
	}

	updateAllocPointCmdFlags(cmd)

	return cmd
}

func updateAllocPointCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("masterchef", "m", "", "master Chef Addr ")
	_ = cmd.MarkFlagRequired("masterchef")

	cmd.Flags().Int64P("pid", "d", 0, "id of pool")
	_ = cmd.MarkFlagRequired("pid")

	cmd.Flags().Int64P("alloc", "p", 0, "allocation point ")
	_ = cmd.MarkFlagRequired("alloc")

	cmd.Flags().BoolP("update", "u", true, "with update")
	_ = cmd.MarkFlagRequired("update")

	cmd.Flags().StringP("caller", "c", "", "caller address")
	_ = cmd.MarkFlagRequired("caller")
}

func UpdateAllocPoint(cmd *cobra.Command, args []string) {
	masterChefAddrStr, _ := cmd.Flags().GetString("masterchef")
	pid, _ := cmd.Flags().GetInt64("pid")
	allocPoint, _ := cmd.Flags().GetInt64("alloc")
	update, _ := cmd.Flags().GetBool("update")

	chainID, _ := cmd.Flags().GetInt32("chainID")
	caller, _ := cmd.Flags().GetString("caller")
	paraName, _ := cmd.Flags().GetString("paraName")
	expire, _ := cmd.Flags().GetString("expire")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	fee, _ := cmd.Flags().GetFloat64("fee")
	feeInt64 := uint64(fee*1e4) * 1e4

	//function set(uint256 _pid, uint256 _allocPoint, bool _withUpdate) public onlyOwner
	parameter := fmt.Sprintf("set(%d, %d, %v)", pid, allocPoint, update)
	_, packData, err := evmAbi.Pack(parameter, masterChef.MasterChefABI, false)
	if nil != err {
		fmt.Println("UpdateAllocPoint", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	exector := chain33Types.GetExecName("evm", paraName)
	action := evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: parameter, Para: packData, ContractAddr: address.ExecAddress(exector)}

	data, err := createEvmTx(chainID, &action, exector, caller, masterChefAddrStr, expire, rpcLaddr, feeInt64)
	if err != nil {
		fmt.Println("UpdateAllocPoint", "Failed to do createEvmTx due to:", err.Error())
		return
	}

	txhex, err := sendTransactionRpc(data, rpcLaddr)
	if err != nil {
		fmt.Println("UpdateAllocPoint", "Failed to do sendTransactionRpc due to:", err.Error())
		return
	}
	fmt.Println(txhex)
}

func TransferOwnerShipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer OwnerShip, should transfer both cakeToken and syrupbar's ownership to masterChef",
		Run:   TransferOwnerShip,
	}

	TransferOwnerShipFlags(cmd)

	return cmd
}

func TransferOwnerShipFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("new", "n", "", "new owner")
	_ = cmd.MarkFlagRequired("new")

	cmd.Flags().StringP("contract", "c", "", "contract address")
	_ = cmd.MarkFlagRequired("contract")

	cmd.Flags().StringP("old", "o", "", "old owner")
	_ = cmd.MarkFlagRequired("old")

	cmd.Flags().Int32P("chainID", "i", 0, "chain id, default to 0(optional)")
}

func TransferOwnerShip(cmd *cobra.Command, args []string) {
	newOwner, _ := cmd.Flags().GetString("new")
	contract, _ := cmd.Flags().GetString("contract")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	caller, _ := cmd.Flags().GetString("old")
	paraName, _ := cmd.Flags().GetString("paraName")
	expire, _ := cmd.Flags().GetString("expire")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	fee, _ := cmd.Flags().GetFloat64("fee")
	feeInt64 := uint64(fee*1e4) * 1e4

	//function transferOwnership(address newOwner) public onlyOwner
	parameter := fmt.Sprintf("transferOwnership(%s)", newOwner)
	_, packData, err := evmAbi.Pack(parameter, syrupBar.OwnableABI, false)
	if nil != err {
		fmt.Println("TransferOwnerShip", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	exector := chain33Types.GetExecName("evm", paraName)
	action := evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: parameter, Para: packData, ContractAddr: address.ExecAddress(exector)}

	data, err := createEvmTx(chainID, &action, exector, caller, contract, expire, rpcLaddr, feeInt64)
	if err != nil {
		fmt.Println("TransferOwnerShip", "Failed to do createEvmTx due to:", err.Error())
		return
	}

	txhex, err := sendTransactionRpc(data, rpcLaddr)
	if err != nil {
		fmt.Println("TransferOwnerShip", "Failed to do sendTransactionRpc due to:", err.Error())
		return
	}
	fmt.Println(txhex)
}
