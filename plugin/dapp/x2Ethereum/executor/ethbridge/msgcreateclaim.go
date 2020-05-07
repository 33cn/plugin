package ethbridge

import (
	"errors"
	"fmt"
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
	gethCommon "github.com/ethereum/go-ethereum/common"
)

// MsgCreateEthBridgeClaim defines a message for creating claims on the ethereum bridge
type MsgCreateEthBridgeClaim types.Eth2Chain33

// NewMsgCreateEthBridgeClaim is a constructor function for MsgCreateBridgeClaim
func NewMsgCreateEthBridgeClaim(ethBridgeClaim types.Eth2Chain33) MsgCreateEthBridgeClaim {
	return MsgCreateEthBridgeClaim(ethBridgeClaim)
}

// Route should return the name of the module
func (msg MsgCreateEthBridgeClaim) Route() string { return types.ModuleName }

// Type should return the action
func (msg MsgCreateEthBridgeClaim) Type() string { return "create_bridge_claim" }

// ValidateBasic runs stateless checks on the message
func (msg MsgCreateEthBridgeClaim) ValidateBasic() error {
	if types.AddressIsEmpty(msg.Chain33Receiver) {
		return types.ErrInvalidAddress
	}

	if types.AddressIsEmpty(msg.ValidatorAddress) {
		return types.ErrInvalidAddress
	}

	if msg.Nonce < 0 {
		return types.ErrInvalidEthNonce
	}

	if !gethCommon.IsHexAddress(msg.EthereumSender) {
		return types.ErrInvalidEthAddress
	}
	if !gethCommon.IsHexAddress(msg.BridgeContractAddress) {
		return types.ErrInvalidEthAddress
	}
	if strings.ToLower(msg.LocalCoinSymbol) == "eth" && msg.TokenContractAddress != "0x0000000000000000000000000000000000000000" {
		return types.ErrInvalidEthSymbol
	}
	return nil
}

// MapOracleClaimsToEthBridgeClaims maps a set of generic oracle claim data into EthBridgeClaim objects
func MapOracleClaimsToEthBridgeClaims(ethereumChainID int, bridgeContract string, nonce int, symbol string, tokenContract string, ethereumSender string, oracleValidatorClaims map[string]string, f func(int, string, int, string, string, string, string, string) (types.Eth2Chain33, error)) ([]types.Eth2Chain33, error) {
	mappedClaims := make([]types.Eth2Chain33, len(oracleValidatorClaims))
	i := 0
	for validator, validatorClaim := range oracleValidatorClaims {
		parseErr := address.CheckAddress(validator)
		if parseErr != nil {
			return nil, errors.New(fmt.Sprintf("failed to parse claim: %s", parseErr))
		}
		mappedClaim, err := f(ethereumChainID, bridgeContract, nonce, symbol, tokenContract, ethereumSender, validator, validatorClaim)
		if err != nil {
			return nil, err
		}
		mappedClaims[i] = mappedClaim
		i++
	}
	return mappedClaims, nil
}
