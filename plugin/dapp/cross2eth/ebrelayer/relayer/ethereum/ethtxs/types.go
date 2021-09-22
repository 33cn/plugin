package ethtxs

import (
	"math/big"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	"github.com/ethereum/go-ethereum/common"
)

//const ...
const (
	X2Eth      = "x2ethereum"
	BurnAction = "Chain33ToEthBurn"
	LockAction = "Chain33ToEthLock"
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
	chain33TxHash        []byte
}
