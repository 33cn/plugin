// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"github.com/33cn/chain33/types"
	typesmocks "github.com/33cn/chain33/types/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	hh := splitHeights2Rows(heights, 3)
	h11 := []*types.BlockInfo{h0, h1, h2}
	h12 := []*types.BlockInfo{h3, h4, h5}
	h13 := []*types.BlockInfo{h6, h7, h8}
	h14 := []*types.BlockInfo{h9}
	expect := [][]*types.BlockInfo{h11, h12, h13, h14}
	assert.Equal(t, expect, hh)

	s, e := getHeaderStartEndRange(0, 100, hh, 0)
	assert.Equal(t, int64(0), s)
	assert.Equal(t, h2.Height, e)

	s, e = getHeaderStartEndRange(0, 100, hh, 1)
	assert.Equal(t, h2.Height+1, s)
	assert.Equal(t, h5.Height, e)

	s, e = getHeaderStartEndRange(0, 100, hh, 2)
	assert.Equal(t, h5.Height+1, s)
	assert.Equal(t, h8.Height, e)

	s, e = getHeaderStartEndRange(0, 100, hh, 3)
	assert.Equal(t, h8.Height+1, s)
	assert.Equal(t, int64(100), e)
}

func TestFetchHeightListBlocks(t *testing.T) {
	para := &client{}
	grpcClient := &typesmocks.Chain33Client{}
	para.grpcClient = grpcClient
	jump := &jumpDldClient{paraClient: para}

	b1 := &types.ParaTxDetail{Header: &types.Header{Height: 1}}
	b2 := &types.ParaTxDetail{Header: &types.Header{Height: 2}}
	b3 := &types.ParaTxDetail{Header: &types.Header{Height: 3}}
	b4 := &types.ParaTxDetail{Header: &types.Header{Height: 4}}
	b5 := &types.ParaTxDetail{Header: &types.Header{Height: 5}}
	b6 := &types.ParaTxDetail{Header: &types.Header{Height: 6}}
	b7 := &types.ParaTxDetail{Header: &types.Header{Height: 7}}
	b8 := &types.ParaTxDetail{Header: &types.Header{Height: 8}}
	b9 := &types.ParaTxDetail{Header: &types.Header{Height: 9}}
	blocks1 := &types.ParaTxDetails{Items: []*types.ParaTxDetail{b1, b2, b3}}
	blocks2 := &types.ParaTxDetails{Items: []*types.ParaTxDetail{b4, b5, b6, b7}}
	blocks3 := &types.ParaTxDetails{Items: []*types.ParaTxDetail{b8, b9}}
	grpcClient.On("GetParaTxByHeight", mock.Anything, mock.Anything).Return(blocks1, nil).Once()
	grpcClient.On("GetParaTxByHeight", mock.Anything, mock.Anything).Return(blocks2, nil).Once()
	grpcClient.On("GetParaTxByHeight", mock.Anything, mock.Anything).Return(blocks3, nil).Once()

	allBlocks := &types.ParaTxDetails{}
	allBlocks.Items = append(allBlocks.Items, blocks1.Items...)
	allBlocks.Items = append(allBlocks.Items, blocks2.Items...)
	allBlocks.Items = append(allBlocks.Items, blocks3.Items...)
	hlist := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	blocks, err := jump.fetchHeightListBlocks(hlist, "title")
	assert.NoError(t, err)
	assert.Equal(t, allBlocks.Items, blocks.Items)
}
