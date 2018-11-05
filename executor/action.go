package executor

import (
	"time"

	dbm "gitlab.33.cn/chain33/chain33/common/db"
	pty "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/types"
)

//查询可提币状态
func QueryUnfreezeWithdraw(stateDB dbm.KV, param *pty.QueryUnfreezeWithdraw) (types.Message, error) {
	//查询提币次数
	//计算当前可否提币
	unfreezeID := param.UnfreezeID
	value, err := stateDB.Get([]byte(unfreezeID))
	if err != nil {
		uflog.Error("QueryWithdraw ", "unfreezeID", unfreezeID, "err", err)
		return nil, err
	}
	var unfreeze pty.Unfreeze
	err = types.Decode(value, &unfreeze)
	if err != nil {
		uflog.Error("QueryWithdraw ", "unfreezeID", unfreezeID, "err", err)
		return nil, err
	}
	currentTime := time.Now().Unix()
	reply := &pty.ReplyQueryUnfreezeWithdraw{UnfreezeID: unfreezeID}
	available, err := getWithdrawAvailable(&unfreeze, currentTime)
	if err != nil {
		reply.AvailableAmount = 0
	} else {
		reply.AvailableAmount = available
	}
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
