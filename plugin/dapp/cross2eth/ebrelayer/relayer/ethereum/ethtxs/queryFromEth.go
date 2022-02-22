package ethtxs

import (
	"context"
	"errors"
	"fmt"
	chain33Abi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"

	bep20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/bep20/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

//GetOperator ...
func GetOperator(client ethinterface.EthClientSpec, sender, bridgeBank common.Address) (common.Address, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		txslog.Error("GetOperator", "Failed to get HeaderByNumber due to:", err.Error())
		return common.Address{}, err
	}

	// Set up CallOpts auth
	auth := bind.CallOpts{
		Pending:     true,
		From:        sender,
		BlockNumber: header.Number,
		Context:     context.Background(),
	}

	// Initialize BridgeRegistry instance
	bridgeBankInstance, err := generated.NewBridgeBank(bridgeBank, client)
	if err != nil {
		txslog.Error("GetOperator", "Failed to NewBridgeBank to:", err.Error())
		return common.Address{}, err
	}

	return bridgeBankInstance.Operator(&auth)
}

//IsActiveValidator ...
func IsActiveValidator(validator common.Address, valset *generated.Valset) (bool, error) {
	opts := &bind.CallOpts{
		Pending: true,
		From:    validator,
		Context: context.Background(),
	}

	// Initialize BridgeRegistry instance
	isActiveValidator, err := valset.IsActiveValidator(opts, validator)
	if err != nil {
		txslog.Error("IsActiveValidator", "Failed to query IsActiveValidator due to:", err.Error())
		return false, err
	}

	return isActiveValidator, nil
}

//IsProphecyPending ...
func IsProphecyPending(claimID [32]byte, validator common.Address, chain33Bridge *generated.Chain33Bridge) (bool, error) {
	opts := &bind.CallOpts{
		Pending: true,
		From:    validator,
		Context: context.Background(),
	}

	// Initialize BridgeRegistry instance
	active, err := chain33Bridge.IsProphecyClaimActive(opts, claimID)
	if err != nil {
		txslog.Error("IsActiveValidatorFromChain33Bridge", "Failed to query IsActiveValidator due to:", err.Error())
		return false, err
	}
	return active, nil
}

//GetBalance ...
func GetBalance(client ethinterface.EthClientSpec, tokenAddr, owner string) (string, error) {
	//查询ERC20余额
	if tokenAddr != "" {
		bep20Token, err := bep20.NewBEP20Token(common.HexToAddress(tokenAddr), client)
		if nil != err {
			txslog.Error("GetBalance", "generated.NewBridgeToken err:", err.Error(), "tokenAddr", tokenAddr, "owner", owner)
			return "", err
		}
		ownerAddr := common.HexToAddress(owner)
		opts := &bind.CallOpts{
			Pending: true,
			From:    ownerAddr,
			Context: context.Background(),
		}
		balance, err := bep20Token.BalanceOf(opts, ownerAddr)
		if nil != err {
			txslog.Error("GetBalance", "bridgeToken.BalanceOf err:", err.Error(), "tokenAddr", tokenAddr, "owner", owner)
			return "", err
		}
		return balance.String(), nil
	}

	//查询ETH余额
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(owner), nil)
	if nil != err {
		return "", err
	}
	return balance.String(), nil
}

//GetLockedFunds ...
func GetLockedFunds(bridgeBank *generated.BridgeBank, tokenAddrStr string) (string, error) {
	var tokenAddr common.Address
	if tokenAddrStr != "" {
		tokenAddr = common.HexToAddress(tokenAddrStr)
	}
	opts := &bind.CallOpts{
		Pending: true,
		From:    tokenAddr,
		Context: context.Background(),
	}
	balance, err := bridgeBank.LockedFunds(opts, tokenAddr)
	if nil != err {
		return "", err
	}
	return balance.String(), nil
}

//GetDepositFunds ...
func GetDepositFunds(client bind.ContractBackend, tokenAddrStr string) (string, error) {
	if tokenAddrStr == "" {
		return "", errors.New("nil token address")
	}

	tokenAddr := common.HexToAddress(tokenAddrStr)
	bridgeToken, err := generated.NewBridgeToken(tokenAddr, client)
	if nil != err {
		return "", err
	}

	opts := &bind.CallOpts{
		Pending: true,
		From:    tokenAddr,
		Context: context.Background(),
	}
	supply, err := bridgeToken.TotalSupply(opts)
	if nil != err {
		return "", err
	}
	return supply.String(), nil
}

//GetToken2address ...
func GetToken2address(bridgeBank *generated.BridgeBank, tokenSymbol string) (string, error) {
	opts := &bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	txslog.Info("GetToken2address", "Name", tokenSymbol)
	tokenAddr, err := bridgeBank.GetToken2address(opts, tokenSymbol)
	if nil != err {
		return "", err
	}
	txslog.Info("GetToken2address", "Name", tokenSymbol, "Address", tokenAddr.String())
	return tokenAddr.String(), nil
}

//GetLockedTokenAddress ...
func GetLockedTokenAddress(bridgeBank *generated.BridgeBank, tokenSymbol string) (string, error) {
	opts := &bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	tokenAddr, err := bridgeBank.GetLockedTokenAddress(opts, tokenSymbol)
	if nil != err {
		return "", err
	}
	txslog.Info("GetLockedTokenAddress", "Name", tokenSymbol, "Address", tokenAddr.String())
	return tokenAddr.String(), nil
}

func QueryResult(param, abiData, contract, owner string, client ethinterface.EthClientSpec) (string, error) {
	txslog.Info("QueryResult", "param", param, "contract", contract, "owner", owner)
	// 首先解析参数字符串，分析出方法名以及个参数取值
	methodName, params, err := chain33Abi.ProcFuncCall(param)
	if err != nil {
		return methodName + " ProcFuncCall fail", err
	}

	// 解析ABI数据结构，获取本次调用的方法对象
	abi_, err := chain33Abi.JSON(strings.NewReader(abiData))
	if err != nil {
		txslog.Info("QueryResult", "JSON fail", err)
		return methodName + " JSON fail", err
	}

	var method chain33Abi.Method
	var ok bool
	txslog.Info("QueryResult Methods")
	if method, ok = abi_.Methods[methodName]; !ok {
		err = fmt.Errorf("function %v not exists", methodName)
		txslog.Info("QueryResult", "Methods fail", err)
		return methodName, err
	}

	if !method.IsConstant() {
		return methodName, errors.New("method is not readonly")
	}
	if len(params) != method.Inputs.LengthNonIndexed() {
		err = fmt.Errorf("function params error:%v", params)
		return methodName, err
	}

	// 获取方法参数对象，遍历解析各参数，获得参数的Go取值
	paramVals := []interface{}{}
	if len(params) != 0 {
		// 首先检查参数个数和ABI中定义的是否一致
		if method.Inputs.LengthNonIndexed() != len(params) {
			err = fmt.Errorf("function Params count error: %v", param)
			return methodName, err
		}

		for i, v := range method.Inputs.NonIndexed() {
			paramVal, err := chain33Abi.Str2GoValue(v.Type, params[i])
			txslog.Info("QueryResult Str2GoValue")
			if err != nil {
				txslog.Info("QueryResult", "Str2GoValue fail", err)
				return methodName + " Str2GoValue fail", err
			}
			paramVals = append(paramVals, paramVal)
		}
	}

	ownerAddr := common.HexToAddress(owner)
	opts := &bind.CallOpts{
		Pending: true,
		From:    ownerAddr,
		Context: context.Background(),
	}
	var out []interface{}
	txslog.Info("QueryResult LoadABI", "abiData", abiData)
	//contactAbi := LoadABI(abiData)
	// Convert the raw abi into a usable format
	contractABI, err := abi.JSON(strings.NewReader(abiData))
	if err != nil {
		panic(err)
	}
	txslog.Info("QueryResult LoadABI", "abiData", abiData)
	boundContract := bind.NewBoundContract(common.HexToAddress(contract), contractABI, client, nil, nil)
	txslog.Info("QueryResult Call", "methodName", methodName, "paramVals", paramVals)
	err = boundContract.Call(opts, &out, methodName, paramVals...)
	if err != nil {
		txslog.Info("QueryResult", "call fail", err)
		return "call err", err
	}
	fmt.Println("QueryBridgeBankResult ret=", out[0])
	return fmt.Sprint(out[0]), err
}
