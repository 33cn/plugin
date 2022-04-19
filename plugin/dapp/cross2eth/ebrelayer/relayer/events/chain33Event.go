package events

import (
	"errors"
	"math/big"

	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	chain33EvmCommon "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Chain33EvmEvent int

const (
	UnsupportedEvent Chain33EvmEvent = iota
	//在chain33的evm合约中产生了lock事件
	Chain33EventLogLock
	//在chain33的evm合约中产生了burn事件
	Chain33EventLogBurn
	//在chain33的evm合约中产生了withdraw事件
	Chain33EventLogWithdraw
)

// String : returns the event type as a string
func (d Chain33EvmEvent) String() string {
	return [...]string{"unknown-event", "LogLock", "LogEthereumTokenBurn", "LogEthereumTokenWithdraw"}[d]
}

// Chain33Msg : contains data from MsgBurn and MsgLock events
type Chain33Msg struct {
	ClaimType            ClaimType
	Chain33Sender        chain33EvmCommon.Address
	EthereumReceiver     common.Address
	TokenContractAddress chain33EvmCommon.Address
	Symbol               string
	Amount               *big.Int
	TxHash               []byte
	Nonce                int64
	ForwardTimes         int32
	ForwardIndex         int64
}

// 发生在chain33evm上的lock事件，当bty跨链转移到eth时会发生该种事件
type LockEventOnChain33 struct {
	From   chain33EvmCommon.Hash160Address
	To     []byte
	Token  chain33EvmCommon.Hash160Address
	Symbol string
	Value  *big.Int
	Nonce  *big.Int
}

// 发生在chain33 evm上的withdraw事件，当用户发起通过代理人提币交易时，则弹射出该事件信息
type WithdrawEventOnChain33 struct {
	BridgeToken      chain33EvmCommon.Hash160Address
	Symbol           string
	Amount           *big.Int
	OwnerFrom        chain33EvmCommon.Hash160Address
	EthereumReceiver []byte
	ProxyReceiver    chain33EvmCommon.Hash160Address
	Nonce            *big.Int
}

// 发生在chain33evm上的burn事件，当eth/erc20资产需要提币回到以太坊链上时，会发生该种事件
type BurnEventOnChain33 struct {
	Token            chain33EvmCommon.Hash160Address
	Symbol           string
	Amount           *big.Int
	OwnerFrom        chain33EvmCommon.Hash160Address
	EthereumReceiver []byte
	Nonce            *big.Int
}

func UnpackChain33LogLock(contractAbi abi.ABI, eventName string, eventData []byte) (lockEvent *LockEventOnChain33, err error) {
	lockEvent = &LockEventOnChain33{}
	// Parse the event's attributes as Ethereum network variables
	err = contractAbi.UnpackIntoInterface(lockEvent, eventName, eventData)
	if err != nil {
		eventsLog.Error("UnpackLogLock", "Failed to unpack abi due to:", err.Error())
		return nil, ebrelayerTypes.ErrUnpack
	}

	eventsLog.Info("UnpackLogLock", "value", lockEvent.Value.String(),
		"symbol", lockEvent.Symbol,
		"token addr on chain33 evm", lockEvent.Token.ToAddress().String(),
		"chain33 sender", lockEvent.From.ToAddress().String(),
		"ethereum recipient", common.BytesToAddress(lockEvent.To).String(),
		"nonce", lockEvent.Nonce.String())

	return lockEvent, nil
}

func UnpackChain33LogBurn(contractAbi abi.ABI, eventName string, eventData []byte) (burnEvent *BurnEventOnChain33, err error) {
	burnEvent = &BurnEventOnChain33{}
	// Parse the event's attributes as Ethereum network variables
	err = contractAbi.UnpackIntoInterface(burnEvent, eventName, eventData)
	if err != nil {
		eventsLog.Error("UnpackLogBurn", "Failed to unpack abi due to:", err.Error())
		return nil, ebrelayerTypes.ErrUnpack
	}

	eventsLog.Info("UnpackLogBurn", "token addr on chain33 evm", burnEvent.Token.ToAddress().String(),
		"symbol", burnEvent.Symbol,
		"Amount", burnEvent.Amount.String(),
		"Owner address from chain33", burnEvent.OwnerFrom.ToAddress().String(),
		"EthereumReceiver", common.BytesToAddress(burnEvent.EthereumReceiver).String(),
		"nonce", burnEvent.Nonce.String())
	return burnEvent, nil
}

func UnpackLogWithdraw(contractAbi abi.ABI, eventName string, eventData []byte) (withdrawEvent *WithdrawEventOnChain33, err error) {
	withdrawEvent = &WithdrawEventOnChain33{}
	err = contractAbi.UnpackIntoInterface(withdrawEvent, eventName, eventData)
	if err != nil {
		eventsLog.Error("UnpackLogWithdraw", "Failed to unpack abi due to:", err.Error())
		return nil, err
	}

	eventsLog.Info("UnpackLogWithdraw", "bridge token addr on chain33 evm", withdrawEvent.BridgeToken.ToAddress().String(),
		"symbol", withdrawEvent.Symbol,
		"Amount", withdrawEvent.Amount.String(),
		"Owner address from chain33", withdrawEvent.OwnerFrom.ToAddress().String(),
		"EthereumReceiver", common.BytesToAddress(withdrawEvent.EthereumReceiver).String(),
		"ProxyReceiver", withdrawEvent.ProxyReceiver.ToAddress().String(),
		"nonce", withdrawEvent.Nonce.String())
	return withdrawEvent, nil
}

// ParseBurnLock4chain33 ParseBurnLockTxReceipt : parses data from a Burn/Lock/Withdraw event witnessed on chain33 into a Chain33Msg struct
func ParseBurnLock4chain33(evmEventType Chain33EvmEvent, data []byte, bridgeBankAbi abi.ABI, chain33TxHash []byte) (*Chain33Msg, error) {
	if Chain33EventLogLock == evmEventType {
		lockEvent, err := UnpackChain33LogLock(bridgeBankAbi, evmEventType.String(), data)
		if nil != err {
			return nil, err
		}

		chain33Msg := &Chain33Msg{
			ClaimType:            ClaimTypeLock,
			Chain33Sender:        lockEvent.From.ToAddress(),
			EthereumReceiver:     common.BytesToAddress(lockEvent.To),
			TokenContractAddress: lockEvent.Token.ToAddress(),
			Symbol:               lockEvent.Symbol,
			Amount:               lockEvent.Value,
			TxHash:               chain33TxHash,
			Nonce:                lockEvent.Nonce.Int64(),
		}
		return chain33Msg, nil

	} else if Chain33EventLogBurn == evmEventType {
		burnEvent, err := UnpackChain33LogBurn(bridgeBankAbi, evmEventType.String(), data)
		if nil != err {
			return nil, err
		}

		chain33Msg := &Chain33Msg{
			ClaimType:            ClaimTypeBurn,
			Chain33Sender:        burnEvent.OwnerFrom.ToAddress(),
			EthereumReceiver:     common.BytesToAddress(burnEvent.EthereumReceiver),
			TokenContractAddress: burnEvent.Token.ToAddress(),
			Symbol:               burnEvent.Symbol,
			Amount:               burnEvent.Amount,
			TxHash:               chain33TxHash,
			Nonce:                burnEvent.Nonce.Int64(),
		}
		return chain33Msg, nil
	} else if Chain33EventLogWithdraw == evmEventType {
		burnEvent, err := UnpackLogWithdraw(bridgeBankAbi, evmEventType.String(), data)
		if nil != err {
			return nil, err
		}

		chain33Msg := &Chain33Msg{
			ClaimType:            ClaimTypeWithdraw,
			Chain33Sender:        burnEvent.OwnerFrom.ToAddress(),
			EthereumReceiver:     common.BytesToAddress(burnEvent.EthereumReceiver),
			TokenContractAddress: burnEvent.BridgeToken.ToAddress(),
			Symbol:               burnEvent.Symbol,
			Amount:               burnEvent.Amount,
			TxHash:               chain33TxHash,
			Nonce:                burnEvent.Nonce.Int64(),
		}
		return chain33Msg, nil
	}

	return nil, errors.New("unknown-event")
}
