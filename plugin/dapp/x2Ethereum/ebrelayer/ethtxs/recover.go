package ethtxs

import (
	"errors"

	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/ethcontract/generated"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func RecoverContractHandler(client *ethclient.Client, sender, registry common.Address) (*X2EthContracts, *X2EthDeployInfo, error) {
	bridgeBankAddr, err := GetAddressFromBridgeRegistry(client, sender, registry, BridgeBank)
	if nil != err {
		return nil, nil, errors.New("Failed to get addr for bridgeBank from registry")
	}
	bridgeBank, err := generated.NewBridgeBank(*bridgeBankAddr, client)
	if nil != err {
		return nil, nil, errors.New("Failed to NewBridgeBank")
	}

	chain33BridgeAddr, err := GetAddressFromBridgeRegistry(client, sender, registry, Chain33Bridge)
	if nil != err {
		return nil, nil, errors.New("Failed to get addr for chain33BridgeAddr from registry")
	}
	chain33Bridge, err := generated.NewChain33Bridge(*chain33BridgeAddr, client)
	if nil != err {
		return nil, nil, errors.New("Failed to NewChain33Bridge")
	}

	oracleAddr, err := GetAddressFromBridgeRegistry(client, sender, registry, Oracle)
	if nil != err {
		return nil, nil, errors.New("Failed to get addr for oracleBridgeAddr from registry")
	}
	oracle, err := generated.NewOracle(*oracleAddr, client)
	if nil != err {
		return nil, nil, errors.New("Failed to NewOracle")
	}

	registryInstance, _ := generated.NewBridgeRegistry(registry, client)
	x2EthContracts := &X2EthContracts{
		BridgeRegistry: registryInstance,
		BridgeBank:     bridgeBank,
		Chain33Bridge:  chain33Bridge,
		Oracle:         oracle,
	}

	x2EthDeployInfo := &X2EthDeployInfo{
		BridgeRegistry: &DeployResult{Address: registry},
		BridgeBank:     &DeployResult{Address: *bridgeBankAddr},
		Chain33Bridge:  &DeployResult{Address: *chain33BridgeAddr},
		Oracle:         &DeployResult{Address: *oracleAddr},
	}

	return x2EthContracts, x2EthDeployInfo, nil
}

func RecoverOracleInstance(client *ethclient.Client, sender, registry common.Address) (*generated.Oracle, error) {
	oracleAddr, err := GetAddressFromBridgeRegistry(client, sender, registry, Oracle)
	if nil != err {
		return nil, errors.New("Failed to get addr for oracleBridgeAddr from registry")
	}
	oracle, err := generated.NewOracle(*oracleAddr, client)
	if nil != err {
		return nil, errors.New("Failed to NewOracle")
	}

	return oracle, nil
}
