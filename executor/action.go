package executor

import (
	"time"
	
	"gitlab.33.cn/chain33/chain33/account"
	dbm "gitlab.33.cn/chain33/chain33/common/db"
	uf "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	"gitlab.33.cn/chain33/chain33/system/dapp"
	"gitlab.33.cn/chain33/chain33/types"
)

type action struct {
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	index        int32
	execaddr     string
}

func newAction(u *Unfreeze, tx *types.Transaction, index int32) *action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &action{u.GetCoinsAccount(), u.GetStateDB(), hash, fromaddr,
		u.GetBlockTime(), u.GetHeight(), index, dapp.ExecAddress(string(tx.Execer))}
}

//创建解冻交易
func (a *action) UnfreezeCreate(create *uf.UnfreezeCreate) (*types.Receipt, error) {
	
	/*
	*参数检测
	*时间等
	 */
	if create.GetStartTime() <= time.Unix()

	//构造ID - txHash
	var unfreezeID string = a.txhash

	receipt, err := a.coinsAccount.TransferToExec(a.fromaddr, a.execaddr, create.TotalCount)
	if err != nil {
		uflog.Error("unfreeze create ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", create.TotalCount)
		return nil, err
	}
	unfreeze := &uf.Unfreeze{
		UnfreezeID:  unfreezeID,
		StartTime:   create.StartTime,
		TokenName:   create.TokenName,
		TotalCount:  create.TotalCount,
		Initiator:   a.fromaddr,
		Beneficiary: create.Beneficiary,
		Period:      create.Period,
		Means:       create.Means,
		Amount:      create.Amount,
	}
	a.saveStateDB(unfreeze)
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	receiptLog := a.getReceiptLog(unfreeze) //TODO 修改receiptLog
	logs = append(logs, receiptLog)

	return &types.Receipt{types.ExecOk, kv, logs}, nil
}

//提取解冻币
func (a *action) UnfreezeWithdraw(withdraw *uf.UnfreezeWithdraw) (*types.Receipt, error) {
	//TODO pseudocode
	/*
	*参数检测
	*检测该地址是否存在对应解冻交易ID
	 */

	/*
	*从合约转币到该地址（收币地址）
	 */
	value, err := a.db.Get(key(withdraw.GetUnfreezeID()))
	if err != nil {
		uflog.Error("unfreeze withdraw ", "execaddr", a.execaddr, "err", err)
		return nil, err
	}
	var unfreeze uf.Unfreeze
	err = types.Decode(value, &unfreeze)
	if err != nil {
		uflog.Error("unfreeze withdraw ", "execaddr", a.execaddr, "err", err)
		return nil, err
	}
	/*
	*检测可取款状态（时间）
	*计算可取款额
	*计算取款次数增加量
	 */
	var amount int64
	var withdrawTimes int32

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	receipt, err := a.coinsAccount.ExecTransferFrozen(unfreeze.Initiator, unfreeze.Beneficiary, a.execaddr, amount)
	if err != nil {
		uflog.Error("unfreeze withdraw ", "execaddr", a.execaddr, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	unfreeze.WithdrawTimes += withdrawTimes
	a.saveStateDB(&unfreeze)

	return &types.Receipt{types.ExecOk, kv, logs}, nil
}

//中止定期解冻
func (a *action) UnfreezeTerminate(terminate *uf.UnfreezeTerminate) (*types.Receipt, error) {
	//TODO pseudocode
	/*
	*参数检测
	*检测该地址是否存在对应解冻交易ID
	 */

	/*
	*从合约转币到该地址（发币地址）
	 */
	//计算合约中剩余币数
	var remain int64
	//获取发币地址
	var senderAddr string
	receipt, err := a.coinsAccount.ExecActive(senderAddr, a.execaddr, remain)
	if err != nil {
		uflog.Error("unfreeze terminate ", "addr", senderAddr, "execaddr", a.execaddr, "err", err)
		return nil, err
	}
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	/*修改receipt
	*修改数据库中状态
	 */
	return &types.Receipt{types.ExecOk, kv, logs}, nil
}

func (a *action) saveStateDB(unfreeze *uf.Unfreeze) {
	a.db.Set(key(unfreeze.GetUnfreezeID()), types.Encode(unfreeze))
}

func key(id string) (keys []byte) {
	keys = append(keys, []byte("mavl-"+uf.UnfreezeX+"-")...)
	keys = append(keys, []byte(id)...)
	return keys
}

func (a *action) getReceiptLog(unfreeze *uf.Unfreeze) *types.ReceiptLog {
	//TODO 判断不同类型receipt
	log := &types.ReceiptLog{}
	r := &uf.ReceiptUnfreeze{}
	r.TokenName = unfreeze.TokenName
	r.CreateAddr = unfreeze.Initiator
	r.ReceiveAddr = unfreeze.Beneficiary
	log.Log = types.Encode(r)
	return log
}

//查询可提币状态
func QueryWithdraw(stateDB dbm.KV, param *uf.QueryWithdraw) (types.Message, error) {
	//查询提币次数
	//计算当前可否提币
	return &types.Reply{}, nil
}
