// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

// newForkTestCfg 构造一个 Title 非 local 的配置，避免 fork 被强制置 0。
func newForkTestCfg(terminate, decayBound int64) *types.Chain33Config {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), `Title="local"`, `Title="chain33"`, 1))
	cfg.SetDappFork(pty.UnfreezeX, pty.ForkTerminatePartX, terminate)
	cfg.SetDappFork(pty.UnfreezeX, pty.ForkUnfreezeDecayBoundX, decayBound)
	return cfg
}

func newLeftProportionUnfreeze(total, tenThousandth, period, start int64) *pty.Unfreeze {
	return &pty.Unfreeze{
		TotalCount: total,
		Means:      pty.LeftProportionX,
		StartTime:  start,
		MeansOpt: &pty.Unfreeze_LeftProportion{LeftProportion: &pty.LeftProportion{
			Period:        period,
			TenThousandth: tenThousandth,
		}},
	}
}

// fork 启用后，v1/v2 都要把"每期衰减 >= 100%"的退化仓位视为已完全解锁：
// 旧实现按轮数奇偶在 ±TotalCount 间跳变（此处奇数轮返回 -MaxInt64）。
func TestDecayBoundForkNormalizesDegenerateTenThousandth(t *testing.T) {
	const now = int64(1_700_000_000)
	const times = int64(1_000_001)
	start := now + 1 - times // period=1 时 times = now+1-start

	for _, tc := range []struct {
		name      string
		terminate int64
	}{
		{"v1", 1 << 40},
		{"v2", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newForkTestCfg(tc.terminate, 0)
			m, err := newMeans(cfg, pty.LeftProportionX, 1)
			require.NoError(t, err)

			got, err := m.calcFrozen(newLeftProportionUnfreeze(math.MaxInt64, 20000, 1, start), now)
			require.NoError(t, err)
			require.Zero(t, got)
		})
	}
}

// fork 前必须保持完全一致：同样的输入仍然按旧循环计算。
func TestDecayBoundNotActiveBeforeFork(t *testing.T) {
	const now = int64(1_700_000_000)
	const times = int64(1_000_001)
	start := now + 1 - times

	cfg := newForkTestCfg(1<<40, 1<<50) // fork 未启用
	m, err := newMeans(cfg, pty.LeftProportionX, 1)
	require.NoError(t, err)
	v1, ok := m.(*leftProportion)
	require.True(t, ok)
	require.False(t, v1.bounded)

	got, err := m.calcFrozen(newLeftProportionUnfreeze(1000, 20000, 1, start), now)
	require.NoError(t, err)
	require.Equal(t, int64(-1000), got) // 奇数轮 * -1
}

// fork 后超长 StartTime 不能再产生超长循环。
func TestDecayBoundForkCapsHugeTimes(t *testing.T) {
	const now = int64(1_700_000_000)
	for _, tc := range []struct {
		name      string
		terminate int64
	}{
		{"v1", 1 << 40},
		{"v2", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newForkTestCfg(tc.terminate, 0)
			m, err := newMeans(cfg, pty.LeftProportionX, 1)
			require.NoError(t, err)

			begin := time.Now()
			got, err := m.calcFrozen(newLeftProportionUnfreeze(1e18, 1, 1, -1e15), now)
			require.NoError(t, err)
			require.Zero(t, got)
			require.Less(t, time.Since(begin), time.Second)
		})
	}
}

// 正常衰减范围内，fork 后的结果必须与旧循环逐位一致。
func TestDecayBoundForkKeepsLegacyResults(t *testing.T) {
	const now = int64(1_700_000_000)
	const period = int64(10)

	cases := []struct {
		total int64
		ten   int64
		times int64
	}{
		{1e18, 1, 0}, {1e18, 1, 1}, {1e18, 1, 2}, {1e18, 1, 100}, {1e18, 1, 4000},
		{12345, 9999, 0}, {12345, 9999, 1}, {12345, 9999, 2}, {12345, 9999, 3},
		{1e8, 100, 0}, {1e8, 100, 1}, {1e8, 100, 10}, {1e8, 100, 100},
	}
	for _, c := range cases {
		start := now + period - c.times*period
		u := newLeftProportionUnfreeze(c.total, c.ten, period, start)
		for _, pair := range []struct {
			name    string
			bounded Means
			legacy  Means
		}{
			{"v1", &leftProportion{bounded: true}, &leftProportion{}},
			{"v2", &leftProportionV2{bounded: true}, &leftProportionV2{}},
		} {
			want, err := pair.legacy.calcFrozen(u, now)
			require.NoError(t, err)
			got, err := pair.bounded.calcFrozen(u, now)
			require.NoError(t, err)
			require.Equal(t, want, got, "%s total=%d ten=%d times=%d", pair.name, c.total, c.ten, c.times)
		}
	}
}

// StartTime 取 MinInt64 时旧实现会因减法溢出返回 TotalCount（永远锁死），fork 后按已解锁处理。
func TestDecayBoundForkHandlesOverflowStartTime(t *testing.T) {
	const now = int64(1_700_000_000)
	u := newLeftProportionUnfreeze(1e18, 2, 1, math.MinInt64)
	got, err := (&leftProportionV2{bounded: true}).calcFrozen(u, now)
	require.NoError(t, err)
	require.Zero(t, got)
}

// fork 后 v1 也要拒绝 >= 10000 的退化参数；fork 前保持原样。
func TestDecayBoundForkRejectsDegenerateTenThousandthOnCreate(t *testing.T) {
	from := &pty.UnfreezeCreate{MeansOpt: &pty.UnfreezeCreate_LeftProportion{
		LeftProportion: &pty.LeftProportion{Period: 1, TenThousandth: 10000},
	}}
	_, err := (&leftProportion{bounded: true}).setOpt(&pty.Unfreeze{}, from)
	require.Error(t, err)

	_, err = (&leftProportion{}).setOpt(&pty.Unfreeze{}, from)
	require.NoError(t, err)
}

// 分叉高度决定是否走有界分支。
func TestDecayBoundForkSelectionByHeight(t *testing.T) {
	cfg := newForkTestCfg(1<<40, 100)

	v1Before, err := newMeans(cfg, pty.LeftProportionX, 99)
	require.NoError(t, err)
	require.False(t, v1Before.(*leftProportion).bounded)

	v1After, err := newMeans(cfg, pty.LeftProportionX, 100)
	require.NoError(t, err)
	require.True(t, v1After.(*leftProportion).bounded)

	cfgV2 := newForkTestCfg(0, 100)
	v2Before, err := newMeans(cfgV2, pty.LeftProportionX, 99)
	require.NoError(t, err)
	require.False(t, v2Before.(*leftProportionV2).bounded)

	v2After, err := newMeans(cfgV2, pty.LeftProportionX, 100)
	require.NoError(t, err)
	require.True(t, v2After.(*leftProportionV2).bounded)
}
