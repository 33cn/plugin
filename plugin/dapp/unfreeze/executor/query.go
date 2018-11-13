package executor

import (
	"time"

	dbm "gitlab.33.cn/chain33/chain33/common/db"
	pty "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/types"
)

func (u *Unfreeze) Query_GetUnfreezeWithdraw(in *types.ReqString) (types.Message, error) {
	return QueryWithdraw(u.GetStateDB(), in.GetData())
}

func (u *Unfreeze) Query_GetUnfreeze(in *types.ReqString) (types.Message, error) {
	return QueryUnfreeze(u.GetStateDB(), in.GetData())
}

//查询可提币状态
func QueryWithdraw(stateDB dbm.KV, unfreezeID string) (types.Message, error) {
	unfreeze, err := loadUnfreeze(unfreezeID, stateDB)
	if err != nil {
		uflog.Error("QueryWithdraw ", "unfreezeID", unfreezeID, "err", err)
		return nil, err
	}
	currentTime := time.Now().Unix()
	reply := &pty.ReplyQueryUnfreezeWithdraw{UnfreezeID: unfreezeID}
	available, err := getWithdrawAvailable(unfreeze, currentTime)
	if err != nil {
		return nil, err
	}

	reply.AvailableAmount = available
	return reply, nil
}

func getWithdrawAvailable(unfreeze *pty.Unfreeze, calcTime int64) (int64, error) {
	means, err := newMeans(unfreeze.Means)
	if err != nil {
		return 0, err
	}
	frozen, err := means.calcFrozen(unfreeze, calcTime)
	if err != nil {
		return 0, err
	}
	_, amount := withdraw(unfreeze, frozen)
	return amount, nil
}

func QueryUnfreeze(stateDB dbm.KV, unfreezeID string) (types.Message, error) {
	unfreeze, err := loadUnfreeze(unfreezeID, stateDB)
	if err != nil {
		uflog.Error("QueryUnfreeze ", "unfreezeID", unfreezeID, "err", err)
		return nil, err
	}

	return unfreeze, nil
}
