// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

func newLeftProportionUnfreeze(start, period, total, tenThousandth int64) *pty.Unfreeze {
	return &pty.Unfreeze{
		TotalCount: total,
		Means:      "LeftProportion",
		StartTime:  start,
		MeansOpt: &pty.Unfreeze_LeftProportion{LeftProportion: &pty.LeftProportion{
			Period:        period,
			TenThousandth: tenThousandth,
		}},
	}
}

// A user-supplied StartTime far in the past makes unfreezeTimes ~1e15. The decay loop must be
// bounded so such a transaction cannot spin inside Exec and halt the node.
func TestCalcFrozenHugeTimesDoesNotSpin(t *testing.T) {
	const now = int64(1_700_000_000)
	for _, m := range []Means{&leftProportion{}, &leftProportionV2{}} {
		u := newLeftProportionUnfreeze(-1e15, 1, 1e18, 1)
		begin := time.Now()
		frozen, err := m.calcFrozen(u, now)
		require.NoError(t, err)
		require.Zero(t, frozen)
		require.Less(t, time.Since(begin), time.Second, "calcFrozen must not iterate unfreezeTimes times")
	}
}

// The bound must not change results for ordinary unfreezes.
func TestCalcFrozenBoundKeepsSmallResults(t *testing.T) {
	u := newLeftProportionUnfreeze(10000, 10, 10000, 2)
	frozen, err := (&leftProportionV2{}).calcFrozen(u, 10011)
	require.NoError(t, err)
	require.Equal(t, int64(9996), frozen)
}
