// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package merkletree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

func TestLeafHash(t *testing.T) {
	leaves := []string{
		"16308793397024662832064523892418908145900866571524124093537199035808550255649",
	}

	h := mimc.NewMiMC("seed")
	s := leafSum(h, mixTy.Str2Byte(leaves[0]))
	assert.Equal(t, "4010939160279929375357088561050093294975728828994381439611270589357856115894", mixTy.Byte2Str(s))

	leaves = []string{
		"4062509694129705639089082212179777344962624935939361647381392834235969534831",
		"3656010751855274388516368747583374746848682779395325737100877017850943546836",
	}
	h.Reset()
	s = nodeSum(h, mixTy.Str2Byte(leaves[0]), mixTy.Str2Byte(leaves[1]))
	assert.Equal(t, "19203142125680456902129919467753534834062383756119332074615431762320316227830", mixTy.Byte2Str(s))

	proves := []string{
		"21467822781369104390668289189963532506973289112396605437823854946060028754354",
		"4010939160279929375357088561050093294975728828994381439611270589357856115894",
		"19203142125680456902129919467753534834062383756119332074615431762320316227830",
		"5949485921924623528448830799540295699445001672535082168235329176256394356669",
	}

	var sum []byte

	for i, l := range proves {
		if len(sum) == 0 {
			sum = leafSum(h, mixTy.Str2Byte(l))
			continue
		}
		//第0个leaf在第一个leaf的右侧，需要调整顺序
		if i < 2 {
			sum = nodeSum(h, mixTy.Str2Byte(l), sum)
			continue
		}
		sum = nodeSum(h, sum, mixTy.Str2Byte(l))
	}
	fmt.Println("sum", mixTy.Byte2Str(sum))
}
