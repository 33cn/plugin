// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/stretchr/testify/assert"
)

func TestSortStages(t *testing.T) {
	stages := &pt.SelfConsensStages{}
	n1 := &pt.SelfConsensStage{StartHeight: 200, Enable: pt.ParaConfigYes}
	e1 := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{n1}}

	n2 := &pt.SelfConsensStage{StartHeight: 100, Enable: pt.ParaConfigNo}
	e2 := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{n2, n1}}

	n3 := &pt.SelfConsensStage{StartHeight: 700, Enable: pt.ParaConfigYes}
	e3 := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{n2, n1, n3}}

	n4 := &pt.SelfConsensStage{StartHeight: 500, Enable: pt.ParaConfigNo}
	e4 := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{n2, n1, n4, n3}}

	sortStages(stages, n1)
	assert.Equal(t, e1, stages)
	sortStages(stages, n2)
	assert.Equal(t, e2, stages)
	sortStages(stages, n3)
	assert.Equal(t, e3, stages)
	sortStages(stages, n4)
	assert.Equal(t, e4, stages)
}
