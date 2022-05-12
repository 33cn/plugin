package ethtxs

// --------------------------------------------------------
//      Parser
//
//      Parses structs containing event information into
//      unsigned transactions for validators to sign, then
//      relays the data packets as transactions on the
//      chain33 Bridge.
// --------------------------------------------------------

import (
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
)

// LogLockToEthBridgeClaim : parses and packages a LockEvent struct with a validator address in an EthBridgeClaim msg
func LogLockToEthBridgeClaim(event *events.LockEvent, ethereumChainID int64, bridgeBrankAddr, ethTxHash string, decimal int64) (*ebrelayerTypes.EthBridgeClaim, error) {
	recipient := event.To
	if 0 == len(recipient) {
		return nil, ebrelayerTypes.ErrEmptyAddress
	}

	chain33Receiver := new(address.Address)
	chain33Receiver.SetBytes(recipient)

	witnessClaim := &ebrelayerTypes.EthBridgeClaim{}
	witnessClaim.EthereumChainID = ethereumChainID
	witnessClaim.BridgeBrankAddr = bridgeBrankAddr
	witnessClaim.Nonce = event.Nonce.Int64()
	witnessClaim.TokenAddr = event.Token.String()
	witnessClaim.Symbol = event.Symbol
	witnessClaim.EthereumSender = event.From.String()
	witnessClaim.Chain33Receiver = chain33Receiver.String()
	witnessClaim.Amount = event.Value.String()

	witnessClaim.ClaimType = int32(events.ClaimTypeLock)
	witnessClaim.ChainName = ""
	witnessClaim.Decimal = decimal
	witnessClaim.EthTxHash = ethTxHash

	return witnessClaim, nil
}

//LogBurnToEthBridgeClaim ...
func LogBurnToEthBridgeClaim(event *events.BurnEvent, ethereumChainID int64, bridgeBrankAddr, ethTxHash string, decimal int64) (*ebrelayerTypes.EthBridgeClaim, error) {
	recipient := event.Chain33Receiver
	if 0 == len(recipient) {
		return nil, ebrelayerTypes.ErrEmptyAddress
	}

	chain33Receiver := new(address.Address)
	chain33Receiver.SetBytes(recipient)

	witnessClaim := &ebrelayerTypes.EthBridgeClaim{}
	witnessClaim.EthereumChainID = ethereumChainID
	witnessClaim.BridgeBrankAddr = bridgeBrankAddr
	witnessClaim.Nonce = event.Nonce.Int64()
	witnessClaim.TokenAddr = event.Token.String()
	witnessClaim.Symbol = event.Symbol
	witnessClaim.EthereumSender = event.OwnerFrom.String()
	witnessClaim.Chain33Receiver = chain33Receiver.String()
	witnessClaim.Amount = event.Amount.String()
	witnessClaim.ClaimType = int32(events.ClaimTypeBurn)
	witnessClaim.ChainName = ""
	witnessClaim.Decimal = decimal
	witnessClaim.EthTxHash = ethTxHash

	return witnessClaim, nil
}

// Chain33MsgToProphecyClaim : parses event data from a Chain33Msg, packaging it as a ProphecyClaim
func Chain33MsgToProphecyClaim(msg events.Chain33Msg) ProphecyClaim {
	claimType := msg.ClaimType
	chain33Sender := msg.Chain33Sender
	ethereumReceiver := msg.EthereumReceiver
	symbol := msg.Symbol
	amount := msg.Amount

	prophecyClaim := ProphecyClaim{
		ClaimType:        claimType,
		Chain33Sender:    chain33Sender.Bytes(),
		EthereumReceiver: ethereumReceiver,
		//TokenContractAddress: tokenContractAddress,
		Symbol:        symbol,
		Amount:        amount,
		Chain33TxHash: msg.TxHash,
	}

	return prophecyClaim
}
