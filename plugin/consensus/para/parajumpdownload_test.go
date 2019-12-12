// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

func TestGetHeightsArry(t *testing.T) {
	h0 := &types.BlockInfo{Height: 1}
	h1 := &types.BlockInfo{Height: 3}
	h2 := &types.BlockInfo{Height: 5}
	h3 := &types.BlockInfo{Height: 6}
	h4 := &types.BlockInfo{Height: 9}
	h5 := &types.BlockInfo{Height: 15}
	h6 := &types.BlockInfo{Height: 21}
	h7 := &types.BlockInfo{Height: 25}
	h8 := &types.BlockInfo{Height: 31}
	h9 := &types.BlockInfo{Height: 41}

	heights := []*types.BlockInfo{h0, h1, h2, h3, h4, h5, h6, h7, h8, h9}

	hh := getHeightsArry(heights, 3)
	h11 := []*types.BlockInfo{h0, h1, h2}
	h12 := []*types.BlockInfo{h3, h4, h5}
	h13 := []*types.BlockInfo{h6, h7, h8}
	h14 := []*types.BlockInfo{h9}
	expect := [][]*types.BlockInfo{h11, h12, h13, h14}
	assert.Equal(t, expect, hh)

	s, e := getStartEndHeight(0, 100, hh, 0)
	assert.Equal(t, int64(0), s)
	assert.Equal(t, h2.Height, e)

	s, e = getStartEndHeight(0, 100, hh, 1)
	assert.Equal(t, h2.Height+1, s)
	assert.Equal(t, h5.Height, e)

	s, e = getStartEndHeight(0, 100, hh, 2)
	assert.Equal(t, h5.Height+1, s)
	assert.Equal(t, h8.Height, e)

	s, e = getStartEndHeight(0, 100, hh, 3)
	assert.Equal(t, h8.Height+1, s)
	assert.Equal(t, int64(100), e)
}
