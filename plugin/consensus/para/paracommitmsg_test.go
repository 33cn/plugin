// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	_ "github.com/33cn/chain33/system"
	"github.com/stretchr/testify/assert"
)

func TestIsSelfConsEnable(t *testing.T) {
	commitCli := new(commitMsgClient)
	enable := commitCli.isSelfConsEnable(0)
	assert.Equal(t, false, enable)

	s1 := &paraSelfConsEnable{startHeight: 10, endHeight: 20}
	s2 := &paraSelfConsEnable{startHeight: 30, endHeight: 40}

	commitCli.selfConsEnableList = append(commitCli.selfConsEnableList, s1)
	commitCli.selfConsEnableList = append(commitCli.selfConsEnableList, s2)

	enable = commitCli.isSelfConsEnable(10)
	assert.Equal(t, true, enable)
	enable = commitCli.isSelfConsEnable(21)
	assert.Equal(t, false, enable)
	enable = commitCli.isSelfConsEnable(30)
	assert.Equal(t, true, enable)
}
