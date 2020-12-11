// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"testing"

	"github.com/bmizerany/assert"

	_ "github.com/33cn/plugin/plugin/crypto/bls"
)

func TestGetNHeight(t *testing.T) {
	h := getNHeight(1)
	assert.Equal(t, int64(40960), h)

	h = getNHeight(2)
	assert.Equal(t, int64(122880), h)

	h = getNHeight(6)
	assert.Equal(t, int64(2580480), h)

	h = getNHeight(7)
	assert.Equal(t, int64(5201920), h)

}

func TestGetN(t *testing.T) {
	n := getCurrentN(1)
	assert.Equal(t, uint32(7), n)

	n = getCurrentN(2621440)
	assert.Equal(t, uint32(7), n)

	n = getCurrentN(2621441)
	assert.Equal(t, uint32(8), n)

	n = getCurrentN(39321600)
	assert.Equal(t, uint32(10), n)

	n = getCurrentN(39321601)
	assert.Equal(t, uint32(11), n)
}

func TestGetBlockNum(t *testing.T) {
	offset := getNHeight(6)
	for n := uint32(7); n < 50; n++ {
		blocks := getNHeight(n) - offset
		secs := blocks * 5
		secsOfYear := int64(60 * 60 * 24 * 365)
		year := secs / secsOfYear
		fmt.Println("n=", n, "coins=", calcCoins(n), "height=", blocks, "year=", year)
	}

}

func TestGetCoins(t *testing.T) {
	c := calcCoins(7)
	assert.Equal(t, int64(1.6*1e8), c)

	c = calcCoins(6)
	assert.Equal(t, int64(3.2*1e8), c)

	//for i:=uint32(0);i<50;i++{
	//	coin := calcCoins(i)
	//	fmt.Println("n",i,"coins",coin)
	//}

}

func getCustomRewardMinerRst(miners []string, height int64) (map[string]int64, int64) {
	res, change := customRewardMiner(0, miners, height)
	check := make(map[string]int64)
	for _, r := range res {
		//fmt.Println("addr",r.Addr,"amount",r.Amount)
		check[r.Addr] += r.Amount
	}
	return check, change
}

func TestCustomRewardMiner(t *testing.T) {
	staffA := "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
	staffB := "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	staffBoss := "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
	miners := []string{staffA, staffB, staffBoss}
	offsetHeight := getNHeight(6)

	//height=1
	height := int64(1)
	check, change := getCustomRewardMinerRst(miners, height)
	assert.Equal(t, int64(3*1e7), check[staffA])
	assert.Equal(t, int64(3*1e7), check[staffB])
	assert.Equal(t, int64(10*1e7), check[staffBoss])
	assert.Equal(t, int64(0), change)

	//height=262144
	var n uint32
	n = 8
	height = getNHeight(n)
	coins := calcCoins(n)
	fmt.Println("n=", n, "coins=", coins)
	check, change = getCustomRewardMinerRst(miners, height-offsetHeight)
	assert.Equal(t, int64(1.5*1e7), check[staffA])
	assert.Equal(t, int64(1.5*1e7), check[staffB])
	assert.Equal(t, int64(5*1e7), check[staffBoss])
	assert.Equal(t, int64(0), change)

	//
	n = 17
	height = getNHeight(n)
	coins = calcCoins(n)
	fmt.Println("n=", n, "coins=", coins)
	check, change = getCustomRewardMinerRst(miners, height-offsetHeight)
	assert.Equal(t, int64(29296), check[staffA])
	assert.Equal(t, int64(29296), check[staffB])
	assert.Equal(t, int64(97658), check[staffBoss])
	assert.Equal(t, int64(0), change)

	//
	n = 18
	height = getNHeight(n)
	coins = calcCoins(n)
	fmt.Println("n=", n, "coins=", coins)
	check, change = getCustomRewardMinerRst(miners, height-offsetHeight)
	assert.Equal(t, int64(14648), check[staffA])
	assert.Equal(t, int64(14648), check[staffB])
	assert.Equal(t, int64(48829), check[staffBoss])
	assert.Equal(t, int64(0), change)

}
