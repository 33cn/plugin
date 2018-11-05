package executor

import (
	pty "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/types"
)

type Means interface {
	setOpt(unfreeze *pty.Unfreeze, from *pty.UnfreezeCreate) (*pty.Unfreeze, error)
	calcFrozen(unfreeze *pty.Unfreeze, now int64) (int64, error)
}

func newMeans(means string) (Means, error) {
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
	frozen := unfreeze.TotalCount
	for i := int64(0); i < unfreezeTimes; i++ {
		frozen = frozen * (10000 - means.TenThousandth) / 10000
	}
	return frozen, nil
}

func withdraw(unfreeze *pty.Unfreeze, frozen int64) (*pty.Unfreeze, int64) {
	amount := unfreeze.Remaining - frozen
	unfreeze.Remaining = frozen
	return unfreeze, amount
}
