package ethbridge

import (
	"strconv"

	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/executor/common"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
	gethCommon "github.com/ethereum/go-ethereum/common"
)

// MsgLock defines a message for locking coins and triggering a related event
type MsgLock struct {
	EthereumChainID  int               `json:"ethereum_chain_id" yaml:"ethereum_chain_id"`
	TokenContract    common.EthAddress `json:"token_contract_address" yaml:"token_contract_address"`
	Chain33Sender    string            `json:"chain33_sender" yaml:"chain33_sender"`
	EthereumReceiver common.EthAddress `json:"ethereum_receiver" yaml:"ethereum_receiver"`
	Amount           uint64            `json:"amount" yaml:"amount"`
}

// NewMsgLock is a constructor function for MsgLock
func NewMsgLock(ethereumChainID int, tokenContract string, cosmosSender string, ethereumReceiver string, amount uint64) MsgLock {
	return MsgLock{
		EthereumChainID:  ethereumChainID,
		TokenContract:    common.NewEthereumAddress(tokenContract),
		Chain33Sender:    cosmosSender,
		EthereumReceiver: common.NewEthereumAddress(ethereumReceiver),
		Amount:           amount,
	}
}

// Route should return the name of the module
func (msg MsgLock) Route() string { return types.ModuleName }

// Type should return the action
func (msg MsgLock) Type() string { return "lock" }

// ValidateBasic runs stateless checks on the message
func (msg MsgLock) ValidateBasic() error {
	if strconv.Itoa(msg.EthereumChainID) == "" {
		return types.ErrInvalidChainID
	}

	if msg.TokenContract.String() == "" {
		return types.ErrInvalidEthAddress
	}

	if !gethCommon.IsHexAddress(msg.TokenContract.String()) {
		return types.ErrInvalidEthAddress
	}

	if types.AddressIsEmpty(msg.Chain33Sender) {
		return types.ErrInvalidAddress
	}

	if msg.EthereumReceiver.String() == "" {
		return types.ErrInvalidEthAddress
	}

	if !gethCommon.IsHexAddress(msg.EthereumReceiver.String()) {
		return types.ErrInvalidEthAddress
	}

	return nil
}
