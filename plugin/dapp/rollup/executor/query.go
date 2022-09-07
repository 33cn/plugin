package executor

import (
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/pkg/errors"
)

func (r *rollup) Query_GetValidatorPubs(title *rtypes.ChainTitle) (types.Message, error) {

	if title.GetValue() == "" {
		return nil, errors.Wrap(types.ErrInvalidParam, "emptyTitle")
	}

	blsPubs, err := r.getValidatorNodesBlsPubs(title.GetValue())
	reply := &rtypes.ValidatorPubs{BlsPubs: blsPubs}
	return reply, err
}

func (r *rollup) Query_GetRollupStatus(title *rtypes.ChainTitle) (types.Message, error) {
	if title.GetValue() == "" {
		return nil, errors.Wrap(types.ErrInvalidParam, "emptyTitle")
	}
	return r.getRollupStatus(title.GetValue())
}
