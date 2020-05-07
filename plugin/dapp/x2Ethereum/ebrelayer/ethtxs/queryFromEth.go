package ethtxs

import (
	"context"
	"errors"

	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/ethcontract/generated"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetOperator(client *ethclient.Client, sender, bridgeBank common.Address) (common.Address, error) {
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

func GetBalance(client *ethclient.Client, tokenAddr, owner string) (string, error) {
	//查询ERC20余额
	if tokenAddr != "" {
		bridgeToken, err := generated.NewBridgeToken(common.HexToAddress(tokenAddr), client)
		if nil != err {
			return "", err
		}
		ownerAddr := common.HexToAddress(owner)
		opts := &bind.CallOpts{
			Pending: true,
			From:    ownerAddr,
			Context: context.Background(),
		}
		balance, err := bridgeToken.BalanceOf(opts, ownerAddr)
		if nil != err {
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

func GetDepositFunds(client *ethclient.Client, tokenAddrStr string) (string, error) {
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
