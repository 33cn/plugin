// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"math"

	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

// Means 解冻算法接口
type Means interface {
	setOpt(unfreeze *pty.Unfreeze, from *pty.UnfreezeCreate) (*pty.Unfreeze, error)
	calcFrozen(unfreeze *pty.Unfreeze, now int64) (int64, error)
}

func newMeans(cfg *types.Chain33Config, means string, height int64) (Means, error) {
	bounded := cfg.IsDappFork(height, pty.UnfreezeX, pty.ForkUnfreezeDecayBoundX)
	if cfg.IsDappFork(height, pty.UnfreezeX, "ForkTerminatePart") {
		if means == "FixAmount" {
			return &fixAmountV2{}, nil
		} else if means == "LeftProportion" {
			return &leftProportionV2{bounded: bounded}, nil
		}
		return nil, types.ErrNotSupport
	}
	if means == "FixAmount" {
		return &fixAmount{}, nil
	} else if means == "LeftProportion" {
		return &leftProportion{bounded: bounded}, nil
	}
	return nil, types.ErrNotSupport
}

// maxDecaySteps 是剩余比例必然归零的迭代次数上界。
// 对 0 < TotalCount <= MaxInt64、0 < TenThousandth < 10000，每步衰减因子
// (10000-TenThousandth)/10000 <= 0.9999，n = 436661 时
// MaxInt64 * 0.9999^436661 < 1（实测 436661 步后为 0.999989），此后只会更小，
// int64 截断恒为 0。该值与 StartTime、Period 无关。
const maxDecaySteps = int64(436_662)

// decaySteps 返回剩余值首次小于 1 的步数上界，O(1) 计算且不会低估
// （+4 覆盖 2n 次舍入与 log 的浮点误差）。调用前需保证 0 < TenThousandth < 10000。
func decaySteps(total, tenThousandth int64) int64 {
	if total <= 1 {
		return 1
	}
	r := 1 - float64(tenThousandth)/10000
	n := int64(math.Ceil(math.Log(float64(total))/-math.Log(r))) + 4
	if n > maxDecaySteps {
		return maxDecaySteps
	}
	return n
}

// calcLeftProportionBounded 是 ForkUnfreezeDecayBound 之后使用的有界实现。
// 迭代次数只由金额和衰减率决定（< decaySteps），与用户提供的 StartTime 无关；
// 对 0 < TenThousandth < 10000 且旧实现不发生整数溢出的输入，结果与旧循环逐位一致，
// 包括轮数为 0（StartTime 在未来）时旧循环的 int64(float64(total)) 浮点往返。
func calcLeftProportionBounded(total, tenThousandth, period, start, now int64) int64 {
	if total <= 0 {
		return 0
	}
	if tenThousandth <= 0 {
		// 无衰减：与旧循环相同，结果经 float64 往返截断
		return int64(float64(total))
	}
	if tenThousandth >= 10000 {
		// 每期衰减 >= 100%，只可能来自 v1 时代的历史仓位。旧实现会按轮数奇偶在
		// ±TotalCount 之间跳变、甚至溢出到 ±Inf，这里统一视为已完全解锁。
		return 0
	}
	if start > now {
		// 轮数为 0：与旧循环一致，返回 int64(float64(total))（>2^53 时有浮点往返尾差）
		return int64(float64(total))
	}
	// now-start 溢出说明已经过了 1e19 个周期，剩余必然为 0。
	if start < 0 && now > math.MaxInt64+start {
		return 0
	}
	// 旧公式 (now+period-start)/period 在 now >= start 时等于 steps+1。
	steps := (now - start) / period
	if steps >= maxDecaySteps || steps+1 >= decaySteps(total, tenThousandth) {
		return 0
	}
	frozen := float64(total)
	for i := int64(0); i <= steps; i++ {
		frozen = frozen * float64(10000-tenThousandth) / 10000
	}
	return int64(frozen)
}

type fixAmount struct {
}

func (opt *fixAmount) setOpt(unfreeze *pty.Unfreeze, from *pty.UnfreezeCreate) (*pty.Unfreeze, error) {
	o := from.GetFixAmount()
	if o == nil {
		return nil, types.ErrInvalidParam
	}
	if o.Amount <= 0 || o.Period <= 0 {
		return nil, types.ErrInvalidParam
	}
	unfreeze.MeansOpt = &pty.Unfreeze_FixAmount{FixAmount: from.GetFixAmount()}
	return unfreeze, nil
}

func (opt *fixAmount) calcFrozen(unfreeze *pty.Unfreeze, now int64) (int64, error) {
	means := unfreeze.GetFixAmount()
	if means == nil {
		return 0, types.ErrInvalidParam
	}
	unfreezeTimes := (now + means.Period - unfreeze.StartTime) / means.Period
	unfreezeAmount := means.Amount * unfreezeTimes
	if unfreeze.TotalCount <= unfreezeAmount {
		return 0, nil
	}
	return unfreeze.TotalCount - unfreezeAmount, nil
}

type leftProportion struct {
	// bounded 为 true 表示 ForkUnfreezeDecayBound 已启用
	bounded bool
}

func (opt *leftProportion) setOpt(unfreeze *pty.Unfreeze, from *pty.UnfreezeCreate) (*pty.Unfreeze, error) {
	o := from.GetLeftProportion()
	if o == nil {
		return nil, types.ErrInvalidParam
	}
	// fork 之后 v1 也拒绝 >= 10000 的退化参数，与 v2 保持一致
	if o.Period <= 0 || o.TenThousandth <= 0 || (opt.bounded && o.TenThousandth >= 10000) {
		return nil, types.ErrInvalidParam
	}
	unfreeze.MeansOpt = &pty.Unfreeze_LeftProportion{LeftProportion: from.GetLeftProportion()}
	return unfreeze, nil
}

func (opt *leftProportion) calcFrozen(unfreeze *pty.Unfreeze, now int64) (int64, error) {
	means := unfreeze.GetLeftProportion()
	if means == nil {
		return 0, types.ErrInvalidParam
	}
	if opt.bounded {
		return calcLeftProportionBounded(unfreeze.TotalCount, means.TenThousandth, means.Period,
			unfreeze.StartTime, now), nil
	}
	unfreezeTimes := (now + means.Period - unfreeze.StartTime) / means.Period
	frozen := float64(unfreeze.TotalCount)
	for i := int64(0); i < unfreezeTimes; i++ {
		frozen = frozen * float64(10000-means.TenThousandth) / 10000
	}
	return int64(frozen), nil
}

func withdraw(unfreeze *pty.Unfreeze, frozen int64) (*pty.Unfreeze, int64) {
	if unfreeze.Remaining == 0 {
		return unfreeze, 0
	}
	amount := unfreeze.Remaining - frozen
	unfreeze.Remaining = frozen
	return unfreeze, amount
}

type fixAmountV2 struct {
}

func (opt *fixAmountV2) setOpt(unfreeze *pty.Unfreeze, from *pty.UnfreezeCreate) (*pty.Unfreeze, error) {
	o := from.GetFixAmount()
	if o == nil {
		return nil, types.ErrInvalidParam
	}
	if o.Amount <= 0 || o.Period <= 0 || unfreeze.TotalCount < o.Amount {
		return nil, types.ErrInvalidParam
	}
	unfreeze.MeansOpt = &pty.Unfreeze_FixAmount{FixAmount: from.GetFixAmount()}
	return unfreeze, nil
}

func (opt *fixAmountV2) calcFrozen(unfreeze *pty.Unfreeze, now int64) (int64, error) {
	means := unfreeze.GetFixAmount()
	if means == nil {
		return 0, types.ErrInvalidParam
	}
	if unfreeze.Terminated {
		return 0, nil
	}
	unfreezeTimes := (now + means.Period - unfreeze.StartTime) / means.Period
	unfreezeAmount := means.Amount * unfreezeTimes
	if unfreeze.TotalCount <= unfreezeAmount {
		return 0, nil
	}
	return unfreeze.TotalCount - unfreezeAmount, nil
}

type leftProportionV2 struct {
	// bounded 为 true 表示 ForkUnfreezeDecayBound 已启用
	bounded bool
}

func (opt *leftProportionV2) setOpt(unfreeze *pty.Unfreeze, from *pty.UnfreezeCreate) (*pty.Unfreeze, error) {
	o := from.GetLeftProportion()
	if o == nil {
		return nil, types.ErrInvalidParam
	}
	if o.Period <= 0 || o.TenThousandth <= 0 || o.TenThousandth >= 10000 {
		return nil, types.ErrInvalidParam
	}
	unfreeze.MeansOpt = &pty.Unfreeze_LeftProportion{LeftProportion: from.GetLeftProportion()}
	return unfreeze, nil
}

func (opt *leftProportionV2) calcFrozen(unfreeze *pty.Unfreeze, now int64) (int64, error) {
	means := unfreeze.GetLeftProportion()
	if means == nil {
		return 0, types.ErrInvalidParam
	}
	if unfreeze.Terminated {
		return 0, nil
	}
	if opt.bounded {
		return calcLeftProportionBounded(unfreeze.TotalCount, means.TenThousandth, means.Period,
			unfreeze.StartTime, now), nil
	}
	unfreezeTimes := (now + means.Period - unfreeze.StartTime) / means.Period
	frozen := float64(unfreeze.TotalCount)
	for i := int64(0); i < unfreezeTimes; i++ {
		frozen = frozen * float64(10000-means.TenThousandth) / 10000
	}
	return int64(frozen), nil
}
