package ethbridge

import (
	"strconv"

	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/executor/common"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
	gethCommon "github.com/ethereum/go-ethereum/common"
)

type Msg_Burn struct {
	EthereumChainID  int64             `json:"ethereum_chain_id" yaml:"ethereum_chain_id"`
	TokenContract    common.EthAddress `json:"token_contract_address" yaml:"token_contract_address"`
	Chain33Sender    string            `json:"chain33_sender" yaml:"chain33_sender"`
	EthereumReceiver common.EthAddress `json:"ethereum_receiver" yaml:"ethereum_receiver"`
	Amount           uint64            `json:"amount" yaml:"amount"`
}

func NewMsgBurn(ethereumChainID int64, tokenContract string, chain33Sender string, ethereumReceiver string, amount uint64) Msg_Burn {
	return Msg_Burn{
		EthereumChainID:  ethereumChainID,
		TokenContract:    common.NewEthereumAddress(tokenContract),
		Chain33Sender:    chain33Sender,
		EthereumReceiver: common.NewEthereumAddress(ethereumReceiver),
		Amount:           amount,
	}
}

// Route should return the name of the module
func (msg Msg_Burn) Route() string { return types.ModuleName }

// Type should return the action
func (msg Msg_Burn) Type() string { return "burn" }

// ValidateBasic runs stateless checks on the message
func (msg Msg_Burn) ValidateBasic() error {
	if strconv.Itoa(int(msg.EthereumChainID)) == "" {
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
