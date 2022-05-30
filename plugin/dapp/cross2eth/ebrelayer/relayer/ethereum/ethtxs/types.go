package ethtxs

import (
	"math/big"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	"github.com/ethereum/go-ethereum/common"
)

//const ...
const (
	EthNullAddr = "0x0000000000000000000000000000000000000000"
)

// OracleClaim : contains data required to make an OracleClaim
type OracleClaim struct {
	ProphecyID *big.Int
	Message    [32]byte
	Signature  []byte
}

// ProphecyClaim : contains data required to make an ProphecyClaim
type ProphecyClaim struct {
	ClaimType            events.ClaimType
	Chain33Sender        []byte
	EthereumReceiver     common.Address
	TokenContractAddress common.Address
	Symbol               string
	Amount               *big.Int
	Chain33TxHash        []byte
}

type WithdrawStatus int32

const (
	WDError      = WithdrawStatus(1)
	WDPending    = WithdrawStatus(2)
	WDFailed     = WithdrawStatus(3)
	WDSuccess    = WithdrawStatus(4)
	BinanceChain = "Binance"
)

// 此处的名字命令不能随意改动，需要与合约event中的命名完全一致
func (d WithdrawStatus) String() string {
	return [...]string{"undefined", "Error,not submitted to ethereum", "Pending", "Submitted to ethereum, but Failed", "Success"}[d]
}
