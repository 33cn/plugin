package ethtxs

import (
	"strings"

	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

//const
const (
	BridgeBankABI    = "BridgeBankABI"
	Chain33BankABI   = "Chain33BankABI"
	Chain33BridgeABI = "Chain33BridgeABI"
	EthereumBankABI  = "EthereumBankABI"
)

//LoadABI ...
func LoadABI(contractName string) abi.ABI {
	var abiJSON string
	switch contractName {
	case BridgeBankABI:
		abiJSON = generated.BridgeBankABI
	case Chain33BankABI:
		abiJSON = generated.Chain33BankABI
	case Chain33BridgeABI:
		abiJSON = generated.Chain33BridgeABI
	case EthereumBankABI:
		abiJSON = generated.EthereumBankABI
	default:
		panic("No abi matched")
	}

	// Convert the raw abi into a usable format
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		panic(err)
	}

	return contractABI
}
