package ethbridge

import (
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/executor/oracle"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

// OracleKeeper defines the expected oracle keeper
type OracleKeeper interface {
	ProcessClaim(claim types.OracleClaim) (oracle.Status, error)
	GetProphecy(id string) (oracle.Prophecy, error)
	GetValidatorArray() ([]types.MsgValidator, error)
	SetConsensusThreshold(ConsensusThreshold int64)
	GetConsensusThreshold() int64
}
