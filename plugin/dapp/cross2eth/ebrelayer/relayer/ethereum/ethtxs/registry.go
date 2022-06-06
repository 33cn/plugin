package ethtxs

import (
	"context"
	"log"

	bridgeRegistry "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// ContractRegistry :
type ContractRegistry byte

const (
	// Valset : valset contract
	Valset ContractRegistry = iota + 1
	// Oracle : oracle contract
	Oracle
	// BridgeBank : bridgeBank contract
	BridgeBank
	// Chain33Bridge : chain33Bridge contract
	Chain33Bridge
)

// String : returns the event type as a string
func (d ContractRegistry) String() string {
	return [...]string{"valset", "oracle", "bridgebank", "chain33bridge", "notsupport"}[d-1]
}

// GetAddressFromBridgeRegistry : utility method which queries the requested contract address from the BridgeRegistry
func GetAddressFromBridgeRegistry(client ethinterface.EthClientSpec, sender, registry common.Address, target ContractRegistry) (address *common.Address, err error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		txslog.Error("GetAddressFromBridgeRegistry", "Failed to get HeaderByNumber due to:", err.Error())
		return nil, err
	}

	// Set up CallOpts auth
	auth := bind.CallOpts{
		Pending:     true,
		From:        sender,
		BlockNumber: header.Number,
		Context:     context.Background(),
	}

	// Initialize BridgeRegistry instance
	registryInstance, err := bridgeRegistry.NewBridgeRegistry(registry, client)
	if err != nil {
		txslog.Error("GetAddressFromBridgeRegistry", "Failed to NewBridgeRegistry to:", err.Error())
		return nil, err
	}

	switch target {
	case Valset:
		valsetAddress, err := registryInstance.Valset(&auth)
		if err != nil {
			log.Fatal(err)
		}
		return &valsetAddress, nil
	case Oracle:
		oracleAddress, err := registryInstance.Oracle(&auth)
		if err != nil {
			txslog.Error("GetAddressFromBridgeRegistry", "Failed to get oracle contract:", err)
			return nil, err
		}
		return &oracleAddress, nil
	case BridgeBank:
		bridgeBankAddress, err := registryInstance.BridgeBank(&auth)
		if err != nil {
			log.Fatal(err)
		}
		return &bridgeBankAddress, nil
	case Chain33Bridge:
		chain33BridgeAddress, err := registryInstance.Chain33Bridge(&auth)
		if err != nil {
			log.Fatal(err)
		}
		return &chain33BridgeAddress, nil
	default:
		txslog.Error("GetAddressFromBridgeRegistry", "invalid target contract type:", target)
		return nil, ebrelayerTypes.ErrInvalidContractAddress
	}
}

// GetDeployHeight : 获取合约部署高度
func GetDeployHeight(client ethinterface.EthClientSpec, sender, registry common.Address) (height int64, err error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		txslog.Error("GetAddressFromBridgeRegistry", "Failed to get HeaderByNumber due to:", err.Error())
		return 0, err
	}

	// Set up CallOpts auth
	callOpts := &bind.CallOpts{
		Pending:     true,
		From:        sender,
		BlockNumber: header.Number,
		Context:     context.Background(),
	}

	// Initialize BridgeRegistry instance
	registryInstance, err := bridgeRegistry.NewBridgeRegistry(registry, client)
	if err != nil {
		txslog.Error("GetAddressFromBridgeRegistry", "Failed to NewBridgeRegistry to:", err.Error())
		return 0, err
	}
	bgInt, err := registryInstance.DeployHeight(callOpts)
	if nil != err {
		return 0, err
	}
	height = bgInt.Int64()
	txslog.Info("GetDeployHeight", "deploy height:", height)

	return
}
