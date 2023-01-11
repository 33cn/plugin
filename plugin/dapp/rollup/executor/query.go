package executor

import (
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

func (r *rollup) Query_GetValidatorPubs(title *rtypes.ChainTitle) (types.Message, error) {

	if title.GetValue() == "" {
		return nil, ErrChainTitle
	}

	blsPubs, err := r.getValidatorNodesBlsPubs(title.GetValue())
	reply := &rtypes.ValidatorPubs{BlsPubs: blsPubs}
	return reply, err
}

func (r *rollup) Query_GetRollupStatus(title *rtypes.ChainTitle) (types.Message, error) {
	if title.GetValue() == "" {
		return nil, ErrChainTitle
	}
	return r.getRollupStatus(title.GetValue())
}

func (r *rollup) Query_GetCommitRoundInfo(req *rtypes.ReqGetCommitRound) (types.Message, error) {

	if req.GetChainTitle() == "" {
		return nil, ErrChainTitle
	}

	return r.getRoundInfo(req.GetChainTitle(), req.GetCommitRound())
}
