package executor

import (
	"errors"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
)

const maxListCount = 1000

func (r *rgbx) Query_ListPendingTx(req *rtypes.ReqListPendingTx) (types.Message, error) {

	if req.GetCount() <= 0 || req.GetCount() > maxListCount {
		return nil, types.ErrInvalidParam
	}
	startKey := formatPendingTxKey(req.GetStartHeight(), req.GetStartIndex())
	values, err := r.GetLocalDB().List([]byte(pendingTxKeyPrefix), startKey, req.GetCount(), 1)
	if err != nil {
		elog.Error("Query_GetPendingTxs", "list err", err, "req", req.String())
		return nil, err
	}

	pendingTxs := &rtypes.PendingTxs{}
	for _, v := range values {
		tx := &rtypes.PendingTx{}
		err := types.Decode(v, tx)
		if err != nil {
			elog.Error("Query_GetPendingTxs", "decode err", err)
			continue
		}

		pendingTxs.PendingList = append(pendingTxs.PendingList, tx)
	}

	return pendingTxs, nil
}

func (r *rgbx) Query_GetConfirmedHeight(_ *types.ReqNil) (types.Message, error) {

	v, err := r.GetLocalDB().Get([]byte(confirmedHeightKey))

	reply := &types.Int64{}
	if errors.Is(err, types.ErrNotFound) {
		return reply, nil
	}
	if err != nil {
		elog.Error("Query_GetConfirmedHeight", "get db err", err)
		return nil, err
	}

	err = types.Decode(v, reply)
	if err != nil {
		elog.Error("Query_GetConfirmedHeight", "decode err", err)
		return nil, err
	}

	return reply, nil

}

func (r *rgbx) Query_GetAsset(req *types.ReqString) (types.Message, error) {

	symbol := req.GetData()
	v, err := r.GetStateDB().Get(formatAssetKey(symbol))
	reply := &rtypes.Asset{}
	if err != nil {
		elog.Error("Query_GetAsset", "symbol", symbol, "get db err", err)
		return nil, err
	}

	err = types.Decode(v, reply)
	if err != nil {
		elog.Error("Query_GetAsset", "symbol", symbol, "decode err", err)
		return nil, err
	}

	return reply, nil
}

func (r *rgbx) Query_GetPendingTx(req *rtypes.ReqGetPendingTx) (types.Message, error) {

	v, err := r.GetLocalDB().Get(formatPendingTxKey(req.GetHeight(), req.GetIndex()))

	reply := &rtypes.PendingTx{}
	if err != nil {
		elog.Error("Query_GetPendingTx", "height", req.GetHeight(),
			"index", req.GetIndex(), "get db err", err)
		return nil, err
	}

	err = types.Decode(v, reply)
	if err != nil {
		elog.Error("Query_GetPendingTx", "height", req.GetHeight(),
			"index", req.GetIndex(), "decode err", err)
		return nil, err
	}

	return reply, nil
}
