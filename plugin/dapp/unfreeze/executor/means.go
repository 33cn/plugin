// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

// Means 解冻算法接口
type Means interface {
	setOpt(unfreeze *pty.Unfreeze, from *pty.UnfreezeCreate) (*pty.Unfreeze, error)
	calcFrozen(unfreeze *pty.Unfreeze, now int64) (int64, error)
}

func newMeans(means string, height int64) (Means, error) {
	if types.IsDappFork(height, pty.UnfreezeX, "ForkTerminatePart") {
		if means == "FixAmount" {
			return &fixAmountV2{}, nil
		} else if means == "LeftProportion" {
			return &leftProportionV2{}, nil
		}
		return nil, types.ErrNotSupport
	}
	if means == "FixAmount" {
		return &fixAmount{}, nil
	} else if means == "LeftProportion" {
		return &leftProportion{}, nil
	}
	return nil, types.ErrNotSupport
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
}

func (opt *leftProportion) setOpt(unfreeze *pty.Unfreeze, from *pty.UnfreezeCreate) (*pty.Unfreeze, error) {
	o := from.GetLeftProportion()
	if o == nil {
		return nil, types.ErrInvalidParam
	}
	if o.Period <= 0 || o.TenThousandth <= 0 {
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
	unfreezeTimes := (now + means.Period - unfreeze.StartTime) / means.Period
	frozen := float64(unfreeze.TotalCount)
	for i := int64(0); i < unfreezeTimes; i++ {
		frozen = frozen * float64(10000-means.TenThousandth) / 10000
	}
	return int64(frozen), nil
}
