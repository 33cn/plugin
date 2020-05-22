// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"github.com/33cn/chain33/queue"
	_ "github.com/33cn/chain33/system"
	drivers "github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/paracross/testnode"
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

func TestParseSelfConsEnableStr(t *testing.T) {
	t1 := []string{"1-100", "200-300"}
	e1 := &paraSelfConsEnable{startHeight: 1, endHeight: 100}
	e2 := &paraSelfConsEnable{startHeight: 200, endHeight: 300}
	ep1 := []*paraSelfConsEnable{e1, e2}
	t2 := []string{"1-100", "200-"}

	l1, err := parseSelfConsEnableStr(t1)
	assert.Nil(t, err)
	assert.Equal(t, ep1, l1)

	l2, err := parseSelfConsEnableStr(t2)
	assert.NotNil(t, err)
	assert.Nil(t, l2)
}

func TestSetSelfConsEnable(t *testing.T) {
	cfg := types.NewChain33Config(testnode.DefaultConfig)
	q := queue.New("channel")
	q.SetConfig(cfg)
	para := new(client)
	para.subCfg = new(subConfig)

	baseCli := drivers.NewBaseClient(&types.Consensus{Name: "name"})
	para.BaseClient = baseCli

	para.InitClient(q.Client(), initTestSyncBlock)
	para.commitMsgClient = &commitMsgClient{
		paraClient: para,
	}
	err := para.commitMsgClient.setSelfConsEnable()
	assert.Nil(t, err)
	e1 := &paraSelfConsEnable{startHeight: 0, endHeight: 1000}
	ep1 := []*paraSelfConsEnable{e1}
	assert.Equal(t, ep1, para.commitMsgClient.selfConsEnableList)

}
