package chain33

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	chain33Types "github.com/33cn/chain33/types"
	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/multicall/multicall"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-swap-periphery/src/pancakeFactory"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-swap-periphery/src/pancakeRouter"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	ethereumcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func DeployMulticall(cmd *cobra.Command) error {
	caller, _ := cmd.Flags().GetString("caller")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	txMulticall, err := deployContract(cmd, multicall.MulticallBin, multicall.MulticallABI, "", "Multicall")
	if err != nil {
		return errors.New(err.Error())
	}

	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("Deploy ERC20 timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(txMulticall, rpcLaddr)
				if data == "" {
					fmt.Println("No receipt received yet for Deploy Multicall tx and continue to wait")
					continue
				} else if data != "2" {
					return errors.New("Deploy Multicall failed due to" + ", ty = " + data)
				}
				fmt.Println("Succeed to deploy Multicall with address =", getContractAddr(caller, txMulticall))
				return nil

			}
		}
	}
}

func DeployERC20(cmd *cobra.Command) error {
	caller, _ := cmd.Flags().GetString("caller")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	supply, _ := cmd.Flags().GetString("supply")

	txhexERC20, err := deployContract(cmd, erc20.ERC20Bin, erc20.ERC20ABI, name+","+symbol+","+supply+","+caller, "ERC20")
	if err != nil {
		return errors.New(err.Error())
	}

	timeout := time.NewTimer(300 * time.Second)
	oneSecondtimeout := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timeout.C:
			panic("Deploy ERC20 timeout")
		case <-oneSecondtimeout.C:
			data, _ := getTxByHashesRpc(txhexERC20, rpcLaddr)
			if data == "" {
				fmt.Println("No receipt received yet for Deploy ERC20 tx and continue to wait")
				continue
			} else if data != "2" {
				return errors.New("Deploy ERC20 failed due to" + ", ty = " + data)
			}
			fmt.Println("Succeed to deploy ERC20 with address =", getContractAddr(caller, txhexERC20), "\\n")
			return nil
		}
	}
}

func DeployPancake(cmd *cobra.Command) error {
	caller, _ := cmd.Flags().GetString("caller")
	parameter, _ := cmd.Flags().GetString("parameter")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	txhexERC20, err := deployContract(cmd, erc20.ERC20Bin, erc20.ERC20ABI, "ycc, ycc, 3300000000, "+caller, "ERC20")
	if err != nil {
		return errors.New(err.Error())
	}

	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("Deploy ERC20 timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(txhexERC20, rpcLaddr)
				if data == "" {
					fmt.Println("No receipt received yet for Deploy ERC20 tx and continue to wait")
					continue
				} else if data != "2" {
					return errors.New("Deploy ERC20 failed due to" + ", ty = " + data)
				}
				fmt.Println("Succeed to deploy ERC20 with address =", getContractAddr(caller, txhexERC20), "\\n")
				goto deployPancakeFactory
			}
		}
	}

deployPancakeFactory:
	txhexPancakeFactory, err := deployContract(cmd, pancakeFactory.PancakeFactoryBin, pancakeFactory.PancakeFactoryABI, parameter, "PancakeFactory")
	if err != nil {
		return errors.New(err.Error())
	}

	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("Deploy PancakeFactory timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(txhexPancakeFactory, rpcLaddr)
				if data == "" {
					fmt.Println("No receipt received yet for Deploy PancakeFactory tx and continue to wait")
					continue
				} else if data != "2" {
					return errors.New("Deploy PancakeFactory failed due to" + ", ty = " + data)
				}
				fmt.Println("Succeed to deploy pancakeFactory with address =", getContractAddr(caller, txhexPancakeFactory), "\\n")
				goto deployWeth9
			}
		}
	}

deployWeth9:
	txhexWeth9, err := deployContract(cmd, pancakeRouter.WETH9Bin, pancakeRouter.WETH9ABI, "", "Weth9")
	if err != nil {
		return errors.New(err.Error())
	}

	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("Deploy Weth9 timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(txhexWeth9, rpcLaddr)
				if data == "" {
					fmt.Println("No receipt received yet for Deploy Weth9 tx and continue to wait")
					continue
				} else if data != "2" {
					return errors.New("Deploy Weth9 failed due to" + ", ty = " + data)
				}
				fmt.Println("Succeed to deploy Weth9 with address =", getContractAddr(caller, txhexWeth9), "\\n")
				goto deployPancakeRouter
			}
		}
	}

deployPancakeRouter:
	param := getContractAddr(caller, txhexPancakeFactory) + "," + getContractAddr(caller, txhexWeth9)
	txhexPancakeRouter, err := deployContract(cmd, pancakeRouter.PancakeRouterBin, pancakeRouter.PancakeRouterABI, param, "PancakeRouter")
	if err != nil {
		return errors.New(err.Error())
	}

	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("Deploy PancakeRouter timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(txhexPancakeRouter, rpcLaddr)
				if data == "" {
					fmt.Println("No receipt received yet for Deploy PancakeRouter tx and continue to wait")
					continue
				} else if data != "2" {
					return errors.New("Deploy PancakeRouter failed due to" + ", ty = " + data)
				}
				fmt.Println("Succeed to deploy PancakeRouter with address =", getContractAddr(caller, txhexPancakeRouter), "\\n")
				return nil
			}
		}
	}
}

func getContractAddr(caller, txhex string) string {
	return address.ExecAddress(caller + ethereumcommon.Bytes2Hex(common.HexToHash(txhex).Bytes()))
}

func deployContract(cmd *cobra.Command, code, abi, parameter, contractName string) (string, error) {
	caller, _ := cmd.Flags().GetString("caller")
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := uint64(fee*1e4) * 1e4

	var action evmtypes.EVMContractAction
	bCode, err := common.FromHex(code)
	if err != nil {
		return "", errors.New(contractName + " parse evm code error " + err.Error())
	}
	exector := chain33Types.GetExecName("evm", paraName)
	action = evmtypes.EVMContractAction{Amount: 0, Code: bCode, GasLimit: 0, GasPrice: 0, Note: note, ContractAddr: address.ExecAddress(exector)}
	if parameter != "" {
		constructorPara := "constructor(" + parameter + ")"
		packData, err := evmAbi.PackContructorPara(constructorPara, abi)
		if err != nil {
			return "", errors.New(contractName + " " + constructorPara + " Pack Contructor Para error:" + err.Error())
		}
		action.Code = append(action.Code, packData...)
	}
	data, err := createEvmTx(chainID, &action, paraName+"evm", caller, address.ExecAddress(paraName+"evm"), expire, rpcLaddr, feeInt64)
	if err != nil {
		return "", errors.New(contractName + " create contract error:" + err.Error())
	}

	txhex, err := sendTransactionRpc(data, rpcLaddr)
	if err != nil {
		return "", errors.New(contractName + " send transaction error:" + err.Error())
	}
	fmt.Println("Deploy", contractName, "tx hash:", txhex)

	return txhex, nil
}

func getTxByHashesRpc(txhex, rpcLaddr string) (string, error) {
	hashesArr := strings.Split(txhex, " ")
	params2 := rpctypes.ReqHashes{
		Hashes: hashesArr,
	}

	var res rpctypes.TransactionDetails
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.GetTxByHashes", params2, &res)
	ctx.SetResultCb(queryTxsByHashesRes)
	result, err := ctx.RunResult()
	if err != nil || result == nil {
		return "", err
	}
	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func queryTxsByHashesRes(arg interface{}) (interface{}, error) {
	var receipt *rpctypes.ReceiptDataResult
	for _, v := range arg.(*rpctypes.TransactionDetails).Txs {
		if v == nil {
			continue
		}
		receipt = v.Receipt
		if nil != receipt {
			return receipt.Ty, nil
		}
	}
	return nil, nil
}

func sendTransactionRpc(data, rpcLaddr string) (string, error) {
	params := rpctypes.RawParm{
		Data: data,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.SendTransaction", params, nil)
	var txhex string
	rpc, err := jsonclient.NewJSONClient(ctx.Addr)
	if err != nil {
		return "", err
	}

	err = rpc.Call(ctx.Method, ctx.Params, &txhex)
	if err != nil {
		return "", err
	}

	return txhex, nil
}
