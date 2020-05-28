package events

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Chain33Msg : contains data from MsgBurn and MsgLock events
type Chain33Msg struct {
	ClaimType            Event
	Chain33Sender        []byte
	EthereumReceiver     common.Address
	TokenContractAddress common.Address
	Symbol               string
	Amount               *big.Int
}

// NewChain33Msg : creates a new Chain33Msg
func NewChain33Msg(
	claimType Event,
	chain33Sender []byte,
	ethereumReceiver common.Address,
	symbol string,
	amount *big.Int,
	tokenContractAddress common.Address,
) Chain33Msg {
	// Package data into a Chain33Msg
	chain33Msg := Chain33Msg{
		ClaimType:            claimType,
		Chain33Sender:        chain33Sender,
		EthereumReceiver:     ethereumReceiver,
		Symbol:               symbol,
		Amount:               amount,
		TokenContractAddress: tokenContractAddress,
	}

	return chain33Msg
}
