// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/33cn/plugin/plugin/dapp/pokerbull/types"
)

// PokerCardNum 牌数，4 * 13 不带大小王
var PokerCardNum = 52

// ColorOffset 牌花色偏移
var ColorOffset uint32 = 8

// ColorBitMask 牌花色bit掩码
var ColorBitMask = 0xFF

// CardNumPerColor 每种花色的牌数
var CardNumPerColor = 13

// CardNumPerGame 一手牌的牌数
var CardNumPerGame = 5

const (
	// PokerbullResultX1 赌注倍数1倍
	PokerbullResultX1 = 1
	// PokerbullResultX2 赌注倍数2倍
	PokerbullResultX2 = 2
	// PokerbullResultX3 赌注倍数3倍
	PokerbullResultX3 = 3
	// PokerbullResultX4 赌注倍数4倍
	PokerbullResultX4 = 4
	// PokerbullResultX5 赌注倍数5倍
	PokerbullResultX5 = 5
	// PokerbullLeverageMax 赌注倍数最大倍数
	PokerbullLeverageMax = PokerbullResultX1
)

// NewPoker 创建一副牌
func NewPoker() *types.PBPoker {
	poker := new(types.PBPoker)
	poker.Cards = make([]int32, PokerCardNum)
	poker.Pointer = int32(PokerCardNum - 1)

	for i := 0; i < PokerCardNum; i++ {
		color := i / CardNumPerColor
		num := i%CardNumPerColor + 1
		poker.Cards[i] = int32(color<<ColorOffset + num)
	}
	return poker
}

// Shuffle 洗牌
func Shuffle(poker *types.PBPoker, rng int64) {
	rndn := rand.New(rand.NewSource(rng))

	for i := 0; i < PokerCardNum; i++ {
		idx := rndn.Intn(PokerCardNum - 1)
		tmpV := poker.Cards[idx]
		poker.Cards[idx] = poker.Cards[PokerCardNum-i-1]
		poker.Cards[PokerCardNum-i-1] = tmpV
	}
	poker.Pointer = int32(PokerCardNum - 1)
}

// Deal 发牌
func Deal(poker *types.PBPoker, rng int64) []int32 {
	if poker.Pointer < int32(CardNumPerGame) {
		logger.Error(fmt.Sprintf("Wait to be shuffled: deal cards [%d], left [%d]", CardNumPerGame, poker.Pointer+1))
		Shuffle(poker, rng+int64(poker.Cards[0]))
	}

	rndn := rand.New(rand.NewSource(rng))
	res := make([]int32, CardNumPerGame)
	for i := 0; i < CardNumPerGame; i++ {
		idx := rndn.Intn(int(poker.Pointer))
		tmpV := poker.Cards[poker.Pointer]
		res[i] = poker.Cards[idx]
		poker.Cards[idx] = tmpV
		poker.Cards[poker.Pointer] = res[i]
		poker.Pointer--
	}

	return res
}

// Result 计算斗牛结果
func Result(cards []int32) int32 {
	temp := 0
	r := -1 //是否有牛标志

	pk := newcolorCard(cards)

	//花牌等于10
	cardsC := make([]int, len(cards))
	for i := 0; i < len(pk); i++ {
		if pk[i].num > 10 {
			cardsC[i] = 10
		} else {
			cardsC[i] = pk[i].num
		}
	}

	//斗牛算法
	result := make([]int, 10)
	var offset = 0
	for x := 0; x < 3; x++ {
		for y := x + 1; y < 4; y++ {
			for z := y + 1; z < 5; z++ {
				if (cardsC[x]+cardsC[y]+cardsC[z])%10 == 0 {
					for j := 0; j < len(cardsC); j++ {
						if j != x && j != y && j != z {
							temp += cardsC[j]
						}
					}

					if temp%10 == 0 {
						r = 10 //若有牛，且剩下的两个数也是牛十
					} else {
						r = temp % 10 //若有牛，剩下的不是牛十
					}
					result[offset] = r
					offset++
				}
			}
		}
	}

	//没有牛
	if r == -1 {
		return -1
	}

	return int32(result[0])
}

// Leverage 计算结果倍数
func Leverage(hand *types.PBHand) int32 {
	result := hand.Result

	// 小牛 [1, 6]
	if result < 7 {
		return PokerbullResultX1
	}

	// 大牛 [7, 9]
	if result >= 7 && result < 10 {
		return PokerbullResultX2
	}

	var flowers = 0
	if result == 10 {
		for _, card := range hand.Cards {
			if (int(card) & ColorBitMask) > 10 {
				flowers++
			}
		}

		// 牛牛
		if flowers < 4 {
			return PokerbullResultX3
		}

		// 四花
		if flowers == 4 {
			return PokerbullResultX4
		}

		// 五花
		if flowers == 5 {
			return PokerbullResultX5
		}
	}

	return PokerbullResultX1
}

type pokerCard struct {
	num   int
	color int
}

type colorCardSlice []*pokerCard

func (p colorCardSlice) Len() int {
	return len(p)
}
func (p colorCardSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p colorCardSlice) Less(i, j int) bool {
	if i >= p.Len() || j >= p.Len() {
		logger.Error("length error. slice length:", p.Len(), " compare lenth: ", i, " ", j)
	}

	if p[i] == nil || p[j] == nil {
		logger.Error("nil pointer at ", i, " ", j)
	}
	return p[i].num < p[j].num
}

func newcolorCard(a []int32) colorCardSlice {
	var cardS []*pokerCard
	for i := 0; i < len(a); i++ {
		num := int(a[i]) & ColorBitMask
		color := int(a[i]) >> ColorOffset
		cardS = append(cardS, &pokerCard{num, color})
	}

	return cardS
}

// CompareResult 两手牌比较结果
func CompareResult(i, j *types.PBHand) bool {
	if i.Result < j.Result {
		return true
	}

	if i.Result == j.Result {
		return Compare(i.Cards, j.Cards)
	}

	return false
}

// Compare 比较两手牌的斗牛结果
func Compare(a []int32, b []int32) bool {
	cardA := newcolorCard(a)
	cardB := newcolorCard(b)

	if !sort.IsSorted(cardA) {
		sort.Sort(cardA)
	}
	if !sort.IsSorted(cardB) {
		sort.Sort(cardB)
	}

	maxA := cardA[len(a)-1]
	maxB := cardB[len(b)-1]
	if maxA.num != maxB.num {
		return maxA.num < maxB.num
	}

	return maxA.color < maxB.color
}
