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
	"math/big"
	"strings"

	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
	"github.com/ethereum/go-ethereum/common"
)

// LogLockToEthBridgeClaim : parses and packages a LockEvent struct with a validator address in an EthBridgeClaim msg
func LogLockToEthBridgeClaim(event *events.LockEvent, ethereumChainID int64, bridgeBrankAddr string, decimal int64) (*ebrelayerTypes.EthBridgeClaim, error) {
	recipient := event.To
	if 0 == len(recipient) {
		return nil, ebrelayerTypes.ErrEmptyAddress
	}
	// Symbol formatted to lowercase
	symbol := strings.ToLower(event.Symbol)
	if symbol == "eth" && event.Token != common.HexToAddress("0x0000000000000000000000000000000000000000") {
		return nil, ebrelayerTypes.ErrAddress4Eth
	}

	witnessClaim := &ebrelayerTypes.EthBridgeClaim{}
	witnessClaim.EthereumChainID = ethereumChainID
	witnessClaim.BridgeBrankAddr = bridgeBrankAddr
	witnessClaim.Nonce = event.Nonce.Int64()
	witnessClaim.TokenAddr = event.Token.String()
	witnessClaim.Symbol = event.Symbol
	witnessClaim.EthereumSender = event.From.String()
	witnessClaim.Chain33Receiver = string(recipient)

	if decimal > 8 {
		event.Value = event.Value.Quo(event.Value, big.NewInt(int64(types.MultiplySpecifyTimes(1, decimal-8))))
	} else {
		event.Value = event.Value.Mul(event.Value, big.NewInt(int64(types.MultiplySpecifyTimes(1, 8-decimal))))
	}
	witnessClaim.Amount = event.Value.String()

	witnessClaim.ClaimType = types.LockClaimType
	witnessClaim.ChainName = types.LockClaim
	witnessClaim.Decimal = decimal

	return witnessClaim, nil
}

//LogBurnToEthBridgeClaim ...
func LogBurnToEthBridgeClaim(event *events.BurnEvent, ethereumChainID int64, bridgeBrankAddr string, decimal int64) (*ebrelayerTypes.EthBridgeClaim, error) {
	recipient := event.Chain33Receiver
	if 0 == len(recipient) {
		return nil, ebrelayerTypes.ErrEmptyAddress
	}

	witnessClaim := &ebrelayerTypes.EthBridgeClaim{}
	witnessClaim.EthereumChainID = ethereumChainID
	witnessClaim.BridgeBrankAddr = bridgeBrankAddr
	witnessClaim.Nonce = event.Nonce.Int64()
	witnessClaim.TokenAddr = event.Token.String()
	witnessClaim.Symbol = event.Symbol
	witnessClaim.EthereumSender = event.OwnerFrom.String()
	witnessClaim.Chain33Receiver = string(recipient)
	witnessClaim.Amount = event.Amount.String()
	witnessClaim.ClaimType = types.BurnClaimType
	witnessClaim.ChainName = types.BurnClaim
	witnessClaim.Decimal = decimal

	return witnessClaim, nil
}

// ParseBurnLockTxReceipt : parses data from a Burn/Lock event witnessed on chain33 into a Chain33Msg struct
func ParseBurnLockTxReceipt(claimType events.Event, receipt *chain33Types.ReceiptData) *events.Chain33Msg {
	// Set up variables
	var chain33Sender []byte
	var ethereumReceiver, tokenContractAddress common.Address
	var symbol string
	var amount *big.Int

	// Iterate over attributes
	for _, log := range receipt.Logs {
		if log.Ty == types.TyChain33ToEthLog || log.Ty == types.TyWithdrawChain33Log {
			txslog.Debug("ParseBurnLockTxReceipt", "value", string(log.Log))
			var chain33ToEth types.ReceiptChain33ToEth
			err := chain33Types.Decode(log.Log, &chain33ToEth)
			if err != nil {
				return nil
			}
			chain33Sender = []byte(chain33ToEth.Chain33Sender)
			ethereumReceiver = common.HexToAddress(chain33ToEth.EthereumReceiver)
			tokenContractAddress = common.HexToAddress(chain33ToEth.TokenContract)
			symbol = chain33ToEth.IssuerDotSymbol
			chain33ToEth.Amount = types.TrimZeroAndDot(chain33ToEth.Amount)
			amount = big.NewInt(1)
			amount, _ = amount.SetString(chain33ToEth.Amount, 10)
			if chain33ToEth.Decimals > 8 {
				amount = amount.Mul(amount, big.NewInt(int64(types.MultiplySpecifyTimes(1, chain33ToEth.Decimals-8))))
			} else {
				amount = amount.Quo(amount, big.NewInt(int64(types.MultiplySpecifyTimes(1, 8-chain33ToEth.Decimals))))
			}

			txslog.Info("ParseBurnLockTxReceipt", "chain33Sender", chain33Sender, "ethereumReceiver", ethereumReceiver.String(), "tokenContractAddress", tokenContractAddress.String(), "symbol", symbol, "amount", amount.String())
			// Package the event data into a Chain33Msg
			chain33Msg := events.NewChain33Msg(claimType, chain33Sender, ethereumReceiver, symbol, amount, tokenContractAddress)
			return &chain33Msg
		}
	}
	return nil
}

// Chain33MsgToProphecyClaim : parses event data from a Chain33Msg, packaging it as a ProphecyClaim
func Chain33MsgToProphecyClaim(event events.Chain33Msg) ProphecyClaim {
	claimType := event.ClaimType
	chain33Sender := event.Chain33Sender
	ethereumReceiver := event.EthereumReceiver
	tokenContractAddress := event.TokenContractAddress
	symbol := strings.ToLower(event.Symbol)
	amount := event.Amount

	prophecyClaim := ProphecyClaim{
		ClaimType:            claimType,
		Chain33Sender:        chain33Sender,
		EthereumReceiver:     ethereumReceiver,
		TokenContractAddress: tokenContractAddress,
		Symbol:               symbol,
		Amount:               amount,
	}

	return prophecyClaim
}
