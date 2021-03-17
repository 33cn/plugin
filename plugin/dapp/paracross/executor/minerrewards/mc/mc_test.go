// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mc

import (
	"fmt"
	"testing"

	"github.com/bmizerany/assert"
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

	for i := 1; i < 10; i++ {
		fmt.Println("n=", i, "height=", getNHeight(uint32(i)))
	}

	//n= 7 height= 5201920
	//n= 8 height= 10444800
	//n= 9 height= 20930560
}

func TestOffsetNHeight(t *testing.T) {
	h := getOffsetNHeight(6)
	assert.Equal(t, int64(0), h)

	h = getOffsetNHeight(7)
	assert.Equal(t, int64(2621440), h)

	h = getOffsetNHeight(8)
	assert.Equal(t, int64(7864320), h)

	h = getOffsetNHeight(9)
	assert.Equal(t, int64(18350080), h)

	for i := 7; i < 13; i++ {
		fmt.Println("n=", i, "height=", getOffsetNHeight(uint32(i)))
	}
	//n= 7 height= 2621440
	//n= 8 height= 7864320
	//n= 9 height= 18350080
}

func TestGetN(t *testing.T) {
	n := getCurrentN(1)
	assert.Equal(t, uint32(7), n)

	n = getCurrentN(2621440)
	assert.Equal(t, uint32(7), n)

	n = getCurrentN(2621441)
	assert.Equal(t, uint32(8), n)

	n = getCurrentN(7864320)
	assert.Equal(t, uint32(8), n)

	n = getCurrentN(7864321)
	assert.Equal(t, uint32(9), n)

	n = getCurrentN(18350080)
	assert.Equal(t, uint32(9), n)

	n = getCurrentN(18350081)
	assert.Equal(t, uint32(10), n)

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
	assert.Equal(t, int64(16*1e7), c)

	c = calcCoins(6)
	assert.Equal(t, int64(32*1e7), c)

	c = calcCoins(13)
	assert.Equal(t, int64(0.25*1e7), c)

	for i := uint32(0); i < 50; i++ {
		coin := calcCoins(i)
		fmt.Println("n", i, "coins", coin)
	}

}

func getCustomRewardMinerRst(miners []string, height int64) (map[string]int64, int64) {
	c := &custom{}
	res, change := c.RewardMiners(nil, 0, miners, height)
	check := make(map[string]int64)
	for _, r := range res {
		//fmt.Println("addr",r.Addr,"amount",r.Amount)
		check[r.Addr] += r.Amount
	}
	return check, change
}

func TestCustomRewardMiner(t *testing.T) {
	staffA := "1CdECPZbnm6zaFF1CvkXn4nMKhwYbNBLN5"
	staffB := "1DV7mnSCXmt5dWyAX8nKqjSgkby1J6vZME"
	staffBoss := "1ARyjNWRY71mJvzQ5RuU3wgvKMaZSY8Qeb"
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
