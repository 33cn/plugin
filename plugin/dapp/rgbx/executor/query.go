package executor

import (
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
