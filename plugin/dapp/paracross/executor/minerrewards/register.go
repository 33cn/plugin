package minerrewards

import (
	"fmt"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

type RewardPolicy interface {
	GetConfigReward(cfg *types.Chain33Config, height int64) (int64, int64, int64)
	RewardMiners(coinReward int64, miners []string, height int64) ([]*pt.ParaMinerReward, int64)
}

const (
	normalMiner = iota
	halveMiner
	customMiner
)

var MinerRewards = make(map[int]RewardPolicy)

func register(ty int, policy RewardPolicy) {
	if _, ok := MinerRewards[ty]; ok {
		panic(fmt.Sprintf("paracross minerreward ty=%d registered", ty))
	}
	MinerRewards[ty] = policy
}
