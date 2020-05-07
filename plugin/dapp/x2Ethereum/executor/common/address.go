package common

import (
	"fmt"
	"reflect"

	gethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// EthereumAddress defines a standard ethereum address
type EthAddress gethCommon.Address

// NewEthereumAddress is a constructor function for EthereumAddress
func NewEthereumAddress(address string) EthAddress {
	return EthAddress(gethCommon.HexToAddress(address))
}

// Route should return the name of the module
func (ethAddr EthAddress) String() string {
	return gethCommon.Address(ethAddr).String()
}

// MarshalJSON marshals the etherum address to JSON
func (ethAddr EthAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", ethAddr.String())), nil
}

// UnmarshalJSON unmarshals an ethereum address
func (ethAddr *EthAddress) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(reflect.TypeOf(gethCommon.Address{}), input, ethAddr[:])
}
