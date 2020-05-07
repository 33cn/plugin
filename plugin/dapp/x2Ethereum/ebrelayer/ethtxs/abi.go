package ethtxs

import (
	"strings"

	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/ethcontract/generated"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	BridgeBankABI    = "BridgeBankABI"
	Chain33BankABI   = "Chain33BankABI"
	Chain33BridgeABI = "Chain33BridgeABI"
	EthereumBankABI  = "EthereumBankABI"
)

func LoadABI(contractName string) abi.ABI {
	var abiJson string
	switch contractName {
	case BridgeBankABI:
		abiJson = generated.BridgeBankABI
	case Chain33BankABI:
		abiJson = generated.Chain33BankABI
	case Chain33BridgeABI:
		abiJson = generated.Chain33BridgeABI
	case EthereumBankABI:
		abiJson = generated.EthereumBankABI
	default:
		panic("No abi matched")
	}

	// Convert the raw abi into a usable format
	contractABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		panic(err)
	}

	return contractABI
}
