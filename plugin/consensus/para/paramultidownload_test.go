// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"fmt"
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

const (
	count = 3
)

func getTestInv(height int64, isDone bool) *inventory {

	headDetail := &types.ParaTxDetail{
		Type: types.AddBlock,
		Header: &types.Header{
			ParentHash: []byte(fmt.Sprint(height - 1)),
			Hash:       []byte(fmt.Sprint(height))}}
	endDetail := &types.ParaTxDetail{
		Type: types.AddBlock,
		Header: &types.Header{
			ParentHash: []byte(fmt.Sprint(height + count - 2)),
			Hash:       []byte(fmt.Sprint(height + count - 1))}}
	txs1 := &types.ParaTxDetails{Items: []*types.ParaTxDetail{headDetail, endDetail, endDetail}}
	return &inventory{
		start:  height,
		end:    height + count - 1,
		txs:    txs1,
		isDone: isDone,
	}
}

func TestVerifyInvs(t *testing.T) {
	var invs []*inventory
	start := int64(1)
	inv1 := getTestInv(start, false)
	invs = append(invs, inv1)

	end := start + count
	inv2 := getTestInv(end, true)
	invs = append(invs, inv2)

	end += count
	inv3 := getTestInv(end, false)
	invs = append(invs, inv3)

	end += count
	inv4 := getTestInv(end, true)
	invs = append(invs, inv4)

	end += count
	inv5 := getTestInv(end, true)
	invs = append(invs, inv5)

	end += count
	inv6 := getTestInv(end, false)
	invs = append(invs, inv6)

	end += count
	inv7 := getTestInv(end, true)
	invs = append(invs, inv7)

	end += count
	inv8 := getTestInv(end, true)
	invs = append(invs, inv8)

	end += count
	inv9 := getTestInv(end, false)
	invs = append(invs, inv9)

	wrongItems := []*inventory{inv1, inv3, inv6, inv9}

	preBlock := &types.ParaTxDetail{
		Type:   types.AddBlock,
		Header: &types.Header{Hash: []byte(fmt.Sprint(start - 1))},
	}
	d := &downloadJob{
		parentBlock: preBlock,
		invs:        invs,
	}

	retry := d.verifyInvs()
	assert.Equal(t, wrongItems, retry)
}

func TestCheckDownLoadRate(t *testing.T) {
	conn1 := &connectCli{downTimes: downTimesFastThreshold + 1}
	conn2 := &connectCli{downTimes: downTimesFastThreshold - 20}
	conn3 := &connectCli{downTimes: downTimesSlowThreshold - 10}
	conn4 := &connectCli{downTimes: downTimesSlowThreshold - 20}
	conn5 := &connectCli{downTimes: downTimesSlowThreshold + 5}
	mCli := &multiDldClient{conns: []*connectCli{conn1, conn2, conn3, conn4, conn5}}
	d := &downloadJob{
		mDldCli: mCli,
	}

	expectConn := []*connectCli{conn1, conn2, conn5}

	d.checkDownLoadRate()
	assert.Equal(t, true, d.mDldCli.connsCheckDone)
	assert.Equal(t, expectConn, d.mDldCli.conns)
}

func TestGetInvs(t *testing.T) {
	subCfg := &subConfig{BatchFetchBlockCount: 1000}
	cli := &client{subCfg: subCfg}
	mCli := &multiDldClient{paraClient: cli}

	var start, end int64
	start = 345800
	end = start + maxRollbackHeight + 3
	invs := mCli.getInvs(start, end)
	assert.Equal(t, 11, len(invs))
	assert.Equal(t, invs[0].end+1, invs[1].start)
	assert.Equal(t, invs[3].end+1, invs[4].start)
	assert.Equal(t, invs[9].end+1, invs[10].start)

}
