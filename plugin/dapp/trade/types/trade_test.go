// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTradeType_GetName(t *testing.T) {
	tp := newType()
	assert.Equal(t, TradeX, tp.GetName())
}

func TestTradeType_GetTypeMap(t *testing.T) {
	tp := newType()
	actoins := tp.GetTypeMap()
	assert.NotNil(t, actoins)
	assert.NotEqual(t, 0, len(actoins))
}

func TestTradeType_GetLogMap(t *testing.T) {
	tp := newType()
	l := tp.GetLogMap()
	assert.NotNil(t, l)
	assert.NotEqual(t, 0, len(l))
}
