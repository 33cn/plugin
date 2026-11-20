// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

// 定理：对 1 <= T0 < 2^63、0 < a < 10000、p > 0，任意 s（含 s > t 的零轮情况），有
//
//	calcLeftProportionBounded(T0, a, p, s, t) = R_old(t)
//
// 这里 R_old 是 ForkUnfreezeDecayBound 之前的逐步衰减循环。

// legacyFrozen 复刻旧实现：n = (now+period-start)/period 次浮点乘除后截断。
// 该函数只用于测试，等价于旧代码 leftProportion/leftProportionV2 的 calcFrozen。
func legacyFrozen(total, ten, period, start, now int64) int64 {
	times := (now + period - start) / period
	frozen := float64(total)
	for i := int64(0); i < times; i++ {
		frozen = frozen * float64(10000-ten) / 10000
	}
	return int64(frozen)
}

// startForTimes 返回使 (now+period-start)/period == times 的 StartTime。
func startForTimes(now, period, times int64) int64 {
	return now + period - times*period
}

// firstZeroTimes 用参考实现二分出"第一次归零"的轮数 k0，不依赖被测代码的 decaySteps 公式。
// 0 < ten < 10000 时 legacyFrozen 关于轮数单调不增，因此可以二分。
func firstZeroTimes(total, ten, period, now, hi int64) int64 {
	lo := int64(0)
	for lo < hi {
		mid := (lo + hi) / 2
		if legacyFrozen(total, ten, period, startForTimes(now, period, mid), now) == 0 {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

// 参考实现本身必须与 fork 前的真实代码路径一致，否则后面的对照没有意义。
func TestLegacyReferenceMatchesPreForkPath(t *testing.T) {
	const now, period = int64(1_700_000_000), int64(10)
	for _, c := range []struct {
		total, ten, times int64
	}{
		{1, 1, 0}, {1, 1, 1}, {12345, 9999, 3}, {1e18, 7, 100}, {math.MaxInt64, 1, 1000},
	} {
		start := startForTimes(now, period, c.times)
		want := legacyFrozen(c.total, c.ten, period, start, now)
		u := &pty.Unfreeze{
			TotalCount: c.total,
			Means:      pty.LeftProportionX,
			StartTime:  start,
			MeansOpt: &pty.Unfreeze_LeftProportion{LeftProportion: &pty.LeftProportion{
				Period:        period,
				TenThousandth: c.ten,
			}},
		}
		for _, m := range []Means{&leftProportion{}, &leftProportionV2{}} {
			got, err := m.calcFrozen(u, now)
			require.NoError(t, err)
			require.Equal(t, want, got, "total=%d ten=%d times=%d", c.total, c.ten, c.times)
		}
	}
}

// 小步数穷举：覆盖全部 ten 取值（1..9999），确认有界实现逐位等于旧循环。
func TestBoundedMatchesLegacySmallSteps(t *testing.T) {
	const now, period = int64(1_700_000_000), int64(10)
	totals := []int64{1, 2, 1000, 1 << 40, 1e18, math.MaxInt64}

	for ten := int64(1); ten < 10000; ten++ {
		for _, total := range totals {
			for times := int64(0); times <= 6; times++ {
				start := startForTimes(now, period, times)
				got := calcLeftProportionBounded(total, ten, period, start, now)
				want := legacyFrozen(total, ten, period, start, now)
				require.Equal(t, want, got, "total=%d ten=%d times=%d", total, ten, times)
			}
		}
	}
}

// 收敛边界：用参考实现独立二分出第一次归零的轮数 k0，检查 k0 附近的逐位一致，
// 并确认 k0 不超过实现里的全局上界 436662。
func TestBoundedMatchesLegacyAtThreshold(t *testing.T) {
	const now, period = int64(1_700_000_000), int64(10)
	cases := []struct {
		total, ten int64
	}{
		{1, 1}, {2, 1}, {999, 1}, {1 << 20, 1}, {1e9, 1}, {1e15, 1}, {1e18, 1}, {math.MaxInt64, 1},
		{1e18, 7}, {1e18, 100}, {1e18, 2500}, {1e18, 9999}, {math.MaxInt64, 9999},
		{1234567890123456789, 3}, {9007199254740993, 2},
	}

	for _, c := range cases {
		k0 := firstZeroTimes(c.total, c.ten, period, now, 500_000)
		require.Greater(t, k0, int64(0), "total=%d ten=%d", c.total, c.ten)
		require.LessOrEqual(t, k0, int64(436_662), "total=%d ten=%d k0=%d", c.total, c.ten, k0)

		for _, times := range []int64{k0 - 1, k0, k0 + 1, k0 + 100} {
			if times < 1 {
				continue
			}
			start := startForTimes(now, period, times)
			got := calcLeftProportionBounded(c.total, c.ten, period, start, now)
			want := legacyFrozen(c.total, c.ten, period, start, now)
			require.Equal(t, want, got, "total=%d ten=%d times=%d k0=%d", c.total, c.ten, times, k0)
		}
	}

	// 最坏情况（T0 = 2^63-1、a = 1）第一次归零的轮数，与实现里的上界对应。
	require.Equal(t, int64(436_661), firstZeroTimes(math.MaxInt64, 1, period, now, 500_000))
}

// 随机对照：一半样本取小步数，一半取在收敛边界附近。
func TestBoundedMatchesLegacyRandom(t *testing.T) {
	const now, period = int64(1_700_000_000), int64(10)
	rng := rand.New(rand.NewSource(20261120))

	for i := 0; i < 1000; i++ {
		total := rng.Int63n(math.MaxInt64-1) + 1
		ten := rng.Int63n(9999) + 1
		times := rng.Int63n(64) + 1
		if i%5 == 0 {
			k := int64(math.Ceil(math.Log(float64(total)) / -math.Log(1-float64(ten)/10000)))
			times = k + rng.Int63n(5) - 2
			if times < 1 {
				times = 1
			}
		}
		start := startForTimes(now, period, times)
		got := calcLeftProportionBounded(total, ten, period, start, now)
		want := legacyFrozen(total, ten, period, start, now)
		require.Equal(t, want, got, "total=%d ten=%d times=%d", total, ten, times)
	}
}

// 结果只依赖轮数 n，与 (period, start) 的具体拆分无关。
func TestBoundedDependsOnlyOnStepCount(t *testing.T) {
	const now, total, ten = int64(1_700_000_000), int64(1e12), int64(37)
	for _, times := range []int64{1, 2, 50, 1000, 200_000} {
		var first int64
		for i, period := range []int64{1, 7, 1000} {
			start := startForTimes(now, period, times)
			got := calcLeftProportionBounded(total, ten, period, start, now)
			require.Equal(t, legacyFrozen(total, ten, period, start, now), got,
				"times=%d period=%d", times, period)
			if i == 0 {
				first = got
			} else {
				require.Equal(t, first, got, "times=%d period=%d", times, period)
			}
		}
	}
}

// n = 0（StartTime 在未来）也必须与旧实现逐位一致：旧循环此时返回
// int64(float64(TotalCount))，对 TotalCount > 2^53 会丢掉浮点往返的尾差。
func TestBoundedMatchesLegacyBeforeStart(t *testing.T) {
	const now, period = int64(1_700_000_000), int64(10)
	for _, total := range []int64{1, 1 << 53, 1<<53 + 1, 9_000_000_000_000_000_001, math.MaxInt64} {
		for _, ten := range []int64{1, 2, 5000, 9999} {
			start := now + 1 // 未来才开始
			got := calcLeftProportionBounded(total, ten, period, start, now)
			want := legacyFrozen(total, ten, period, start, now)
			require.Equal(t, want, got, "total=%d ten=%d", total, ten)
		}
	}
}
