package rollup

import (
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

func (r *RollUp) getValidatorPubKeys() *rtypes.ValidatorPubs {

	req := &rtypes.ChainTitle{Value: r.chainCfg.GetTitle()}

	reply, err := r.mainChainGrpc.QueryChain(r.ctx, &types.ChainExecutor{
		Driver:   rtypes.RollupX,
		FuncName: "GetValidatorPubs",
		Param:    types.Encode(req),
	})

	if err != nil || !reply.GetIsOk() {
		rlog.Error("getValidatorPubKeys", "msg", string(reply.GetMsg()), "query err", err)
		return nil
	}

	res := &rtypes.ValidatorPubs{}

	err = types.Decode(reply.GetMsg(), res)
	if err != nil {
		rlog.Error("getValidatorPubKeys", "decode err", err)
		return nil
	}

	return res
}

func (r *RollUp) getRollupStatus() *rtypes.RollupStatus {

	req := &rtypes.ChainTitle{Value: r.chainCfg.GetTitle()}

	reply, err := r.mainChainGrpc.QueryChain(r.ctx, &types.ChainExecutor{
		Driver:   rtypes.RollupX,
		FuncName: "GetRollupStatus",
		Param:    types.Encode(req),
	})

	if err != nil || !reply.GetIsOk() {
		rlog.Error("getRollupStatus", "msg", string(reply.GetMsg()), "query err", err)
		return nil
	}

	res := &rtypes.RollupStatus{}

	err = types.Decode(reply.GetMsg(), res)
	if err != nil {
		rlog.Error("getRollupStatus", "decode err", err)
		return nil
	}
	return res
}
