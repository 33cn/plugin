package events

// -----------------------------------------------------
//    ethereumEvent : Creates LockEvents from new events on the
//			  Ethereum blockchain.
// -----------------------------------------------------

import (
	"math/big"

	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// LockEvent : struct which represents a LogLock event
type LockEvent struct {
	From   common.Address
	To     []byte
	Token  common.Address
	Symbol string
	Value  *big.Int
	Nonce  *big.Int
}

// BurnEvent : struct which represents a BurnEvent event
type BurnEvent struct {
	Token           common.Address
	Symbol          string
	Amount          *big.Int
	OwnerFrom       common.Address
	Chain33Receiver []byte
	Nonce           *big.Int
}

// NewProphecyClaimEvent : struct which represents a LogNewProphecyClaim event
type NewProphecyClaimEvent struct {
	ProphecyID       *big.Int
	ClaimType        uint8
	Chain33Sender    []byte
	EthereumReceiver common.Address
	ValidatorAddress common.Address
	TokenAddress     common.Address
	Symbol           string
	Amount           *big.Int
}

//LogNewBridgeToken ...
type LogNewBridgeToken struct {
	Token  common.Address
	Symbol string
}

// UnpackLogLock : Handles new LogLock events
func UnpackLogLock(contractAbi abi.ABI, eventName string, eventData []byte) (lockEvent *LockEvent, err error) {
	event := &LockEvent{}
	// Parse the event's attributes as Ethereum network variables
	err = contractAbi.UnpackIntoInterface(event, eventName, eventData)
	if err != nil {
		eventsLog.Error("UnpackLogLock", "Failed to unpack abi due to:", err.Error())
		return nil, ebrelayerTypes.ErrUnpack
	}

	eventsLog.Info("UnpackLogLock", "value", event.Value.String(), "symbol", event.Symbol,
		"token addr", event.Token.Hex(), "sender", event.From.Hex(),
		"recipient", string(event.To), "nonce", event.Nonce.String())

	return event, nil
}

//UnpackLogBurn ...
func UnpackLogBurn(contractAbi abi.ABI, eventName string, eventData []byte) (burnEvent *BurnEvent, err error) {
	event := &BurnEvent{}
	// Parse the event's attributes as Ethereum network variables
	err = contractAbi.UnpackIntoInterface(event, eventName, eventData)
	if err != nil {
		eventsLog.Error("UnpackLogBurn", "Failed to unpack abi due to:", err.Error())
		return nil, ebrelayerTypes.ErrUnpack
	}

	eventsLog.Info("UnpackLogBurn", "token addr", event.Token.Hex(), "symbol", event.Symbol,
		"Amount", event.Amount.String(), "OwnerFrom", event.OwnerFrom.String(),
		"Chain33Receiver", string(event.Chain33Receiver), "nonce", event.Nonce.String())
	return event, nil
}
