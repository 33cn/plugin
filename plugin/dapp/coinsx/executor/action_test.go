// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestFilterByAddrs(t *testing.T) {
	curr := []string{"aa", "bb", "cc"}
	del := []string{"bb"}
	expect := []string{"aa", "cc"}
	ret := filterByAddrs(curr, del)
	assert.Equal(t, ret, expect)

	del = []string{"aa", "bb", "cc"}
	ret = filterByAddrs(curr, del)
	assert.Equal(t, []string(nil), ret)

}

func TestFilterAddrs(t *testing.T) {
	curr := []string{"aa", "bb", "cc", "bb"}
	exp := []string{"aa", "bb", "cc"}
	ret := filterAddrs(curr)
	assert.Equal(t, ret, exp)

	curr = []string{"bb", "bb"}
	ret = filterAddrs(curr)
	assert.Equal(t, []string{"bb"}, ret)

}
