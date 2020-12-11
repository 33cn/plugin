// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

const (
	startN    uint32 = 11
	minerUnit int64  = 10000

	//addr:quota的配置
	// 18.75/100份额，由于浮点数的原因，都扩大100倍即1875/10000
	addrStaffA = "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4:1875" //18.75quota 1875/10000
	addrStaffB = "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR:1875" //18.75 quota
	addrBoss   = "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k:6250" //62.50 quota

)

var addrs = []string{addrStaffA, addrStaffB, addrBoss}
var addrsMap = make(map[string]int64)

func checkQuota() {
	var sum int64
	for _, a := range addrs {
		val := strings.Split(a, ":")
		v, err := strconv.Atoi(val[1])
		if err != nil || v == 0 {
			panic(fmt.Sprintf("minerCustom checkQuota err=%s,addr=%s", err, a))
		}
		addrsMap[val[0]] = int64(v)
		sum += int64(v)
	}
	if sum > minerUnit {
		panic(fmt.Sprintf("minerCustom checkQuota sum=%d beyond %d", sum, minerUnit))
	}
}

func init() {
	getConfigRewards[customMiner] = getCustomReward
	rewardMiner[customMiner] = customRewardMiner

	checkQuota()
}

func getCustomReward(cfg *types.Chain33Config, height int64) (int64, int64, int64) {
	n := getCurrentN(height)
	return calcCoins(n), 0, 0

}

func getNHeight(n uint32) int64 {
	v := 1 << n
	return 40960 * (int64(v) - 1)
}

//高度 4096*(2^n -1)+1 开始减半， n=1:1~4096, n=2:4096+1~12288
func getCurrentN(height int64) uint32 {
	if height <= 0 {
		panic("height should bigger than 0")
	}

	var totalCycle uint32 = 1 << 30
	offsetHeight := getNHeight(6)
	for n := uint32(0); n < totalCycle; n++ {
		leftVal := getNHeight(n)
		rightVal := getNHeight(n + 1)
		if offsetHeight+height > leftVal && offsetHeight+height <= rightVal {
			return n + 1
		}
	}
	panic("not enought total cycle")

}

//客户原有链大约50s出一个块，一个块是32/16/8..coins, 我们平均5s一个块，需要/10
func calcCoins(n uint32) int64 {
	if n <= startN {
		return 1e7 * (1 << (startN - n))
	}

	v := 1 << (n - startN)
	vf := 1e7 / v

	return int64(math.Trunc(float64(vf)))
}

func customRewardMiner(coinReward int64, miners []string, height int64) ([]*pt.ParaMinerReward, int64) {
	//找零
	var change int64
	var rewards []*pt.ParaMinerReward

	coins, _, _ := getCustomReward(nil, height)
	var sum int64
	//get quto to miner
	for _, m := range miners {
		if v, ok := addrsMap[m]; ok {
			amount := (v * coins) / minerUnit
			sum += amount
			r := &pt.ParaMinerReward{Addr: m, Amount: amount}
			rewards = append(rewards, r)
		}
	}

	//所有找零给addrC
	change = coins - sum
	if change > 0 {
		val := strings.Split(addrBoss, ":")
		r := &pt.ParaMinerReward{Addr: val[0], Amount: change}
		rewards = append(rewards, r)
	}
	return rewards, 0
}
