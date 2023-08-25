package state

import (
	"errors"
	"fmt"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	ticket "github.com/33cn/plugin/plugin/dapp/ticket/executor"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

//CloseBindMiner 关闭绑定挖矿，先把自己的票关闭掉，然后重新建立绑定关系
func (mdb *MemoryStateDB) CloseBindMiner(from, bind common.Address, amount int64) (bool, error) {
	//step 1 关闭挖矿
	receipt, err := mdb.closeMiner(from, bind, amount)
	if err != nil {
		return false, err
	}

	//step2 重新建立绑定绑定关系
	receipt2, err := mdb.bindMiner(from, from, 0)
	if err != nil {
		return false, err
	}

	receipt.KV = append(receipt.KV, receipt2.GetKV()...)
	receipt.Logs = append(receipt.Logs, receipt2.GetLogs()...)

	mdb.addChange(ticketChange{
		baseChange: baseChange{},
		data:       receipt.GetKV(),
		logs:       receipt.GetLogs(),
	})

	return true, nil
}

//TransferToExec 向特定执行器转账
func (mdb *MemoryStateDB) TransferToExec(from common.Address, exec string, amount int64) (bool, error) {
	receipt, err := mdb.CoinsAccount.TransferToExec(from.String(), address.ExecAddress(exec), amount)
	if err != nil {
		return false, err
	}

	mdb.addChange(transferChange{
		baseChange: baseChange{},
		amount:     amount,
		data:       receipt.GetKV(),
		logs:       receipt.GetLogs(),
	})

	return true, nil
}

//CreateBindMiner ,创建本地址与挖矿地址的绑定关系,并转移amount 数量的币去挖矿
func (mdb *MemoryStateDB) CreateBindMiner(from, bind common.Address, amount int64) (bool, error) {
	receipt, err := mdb.bindMiner(from, bind, amount)
	if err != nil {
		return false, err
	}
	mdb.addChange(ticketChange{
		baseChange: baseChange{},
		amount:     amount,
		data:       receipt.GetKV(),
		logs:       receipt.GetLogs(),
	})

	return true, nil
}

func (mdb *MemoryStateDB) bindMiner(from, bind common.Address, amount int64) (*types.Receipt, error) {
	if from == bind { //授权地址和签名地址不能是同一个地址
		return nil, types.ErrFromAddr
	}
	log15.Info("bindMiner ++++++STEP 1")
	cfg := ty.GetTicketMinerParam(mdb.GetConfig(), mdb.blockHeight)
	fee := mdb.GetConfig().GetCoinPrecision()

	if amount > 0 && (amount-2*fee)/cfg.TicketPrice < 1 { //至少一张票
		return nil, errors.New("insufficient balance to buy a ticket")
	}
	log15.Info("bindMiner ++++++STEP 2", "ReturnAddress:", from.String())
	//check address
	if err := address.CheckAddress(bind.String(), mdb.blockHeight); err != nil {
		return nil, err
	}
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	tNewBind := &ty.TicketBind{MinerAddress: bind.String(), ReturnAddress: from.String()}
	oldbind := mdb.getBindLog(tNewBind, mdb.getBind(from.String()))
	logs = append(logs, oldbind)
	mdb.saveBind(mdb.StateDB, tNewBind)
	bindKv := mdb.getBindKV(tNewBind)
	kvs = append(kvs, bindKv...)
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	log15.Info("bindMiner ++++++STEP 3")
	//如果填写金额参数，则认为在绑定的时候一次性转移相应的coins 到执行器
	if amount > 0 {
		//transfer to exec ----->send ticket exector
		receipt2, err := mdb.CoinsAccount.TransferToExec(from.String(), address.ExecAddress("ticket"), amount)
		if err != nil {
			return nil, err
		}

		receipt.KV = append(receipt.KV, receipt2.GetKV()...)
		receipt.Logs = append(receipt.Logs, receipt2.GetLogs()...)

	}
	log15.Info("bindMiner ++++++STEP 4")
	return receipt, nil
}

func (mdb *MemoryStateDB) closeMiner(from, bind common.Address, amount int64) (*types.Receipt, error) {
	//step 1 获取当前的挖票数
	tickets, err := mdb.listBindTicket(from.String())
	if err != nil {
		return nil, err
	}
	dbtickets := make([]*ticket.DB, len(tickets))
	cfg := ty.GetTicketMinerParam(mdb.GetConfig(), mdb.GetBlockHeight())
	for i, tk := range tickets {
		//ticket 的生成时间超过 2天,可提款,ticket 的状态必须是minded或者Opened 状态
		if tk.Status != ty.TicketMined && tk.Status != ty.TicketOpened {
			log15.Error("CloseBindMiner", "id", tk.GetTicketId(), "status", tk.GetStatus())
			return nil, ty.ErrTicketClosed
		}

		if !tk.IsGenesis {
			//分成两种情况
			//开启挖矿
			if tk.Status == ty.TicketOpened && mdb.blockTime-tk.GetCreateTime() < cfg.TicketWithdrawTime {
				return nil, ty.ErrTime
			}
			//已经挖矿成功了
			if tk.Status == ty.TicketMined && mdb.blockTime-tk.GetCreateTime() < cfg.TicketWithdrawTime {
				return nil, ty.ErrTime
			}

			if tk.Status == ty.TicketMined && mdb.blockTime-tk.GetMinerTime() < cfg.TicketMinerWaitTime {
				return nil, ty.ErrTime
			}
		}

		prevstatus := tk.Status
		tk.Status = ty.TicketClosed
		tkdb := &ticket.DB{Ticket: *tk}
		//更新节点状态
		tkdb.SetPrevstatus(prevstatus)
		dbtickets[i] = tkdb
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	for _, tk := range dbtickets {
		if tk.GetPrevStatus() == 1 { //如果之前的状态是OPENED，则MinerValue设置为0
			//MinerValue 为挖到币的实际数量
			tk.MinerValue = 0
		}
		//返回的挖矿金额 GetRealPrice：3000 coins +挖到矿的数量
		retValue := tk.GetRealPrice(mdb.GetConfig()) + tk.MinerValue
		receipt1, err := mdb.CoinsAccount.ExecActive(from.String(), address.ExecAddress("ticket"), retValue)
		if err != nil {

			return nil, err
		}
		logs = append(logs, tk.GetReceiptLog())
		kv = append(kv, tk.GetKVSet()...)
		logs = append(logs, receipt1.Logs...)
		kv = append(kv, receipt1.KV...)

		//如果ticket 已经挖矿成功了，那么要解冻发展基金部分币
		if tk.GetPrevStatus() == 2 {
			receipt2, err := mdb.CoinsAccount.ExecActive(mdb.GetConfig().GetFundAddr(), address.ExecAddress("ticket"), cfg.CoinDevFund)
			if err != nil {
				log15.Error("TicketClose.ExecActive fund", "addr", mdb.GetConfig().GetFundAddr(), "execaddr", address.ExecAddress("ticket"), "value", retValue)
				return nil, err
			}
			logs = append(logs, receipt2.Logs...)
			kv = append(kv, receipt2.KV...)

		}

		tk.Save(mdb.StateDB)
	}

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil

}
func bindKey(addr string) []byte {
	var bindKey []byte
	ticketBindKeyPrefix := []byte("mavl-ticket-tbind-")
	bindKey = append(bindKey, ticketBindKeyPrefix...)
	bindKey = append(bindKey, []byte(addr)...)
	return bindKey
}

//getBind 返回老的绑定地址，即上一个绑定关系的地址,如果没有则返回为空
func (mdb *MemoryStateDB) getBind(addr string) string {
	log15.Info("getBind++++++++++STEP1 addr:", addr)
	var bindKey []byte
	ticketBindKeyPrefix := []byte("mavl-ticket-tbind-")
	bindKey = append(bindKey, ticketBindKeyPrefix...)
	bindKey = append(bindKey, []byte(addr)...)
	value, err := mdb.StateDB.Get(bindKey)
	if err != nil || value == nil {
		return ""
	}
	log15.Info("getBind++++++++++STEP2 value:", value)
	var bind ty.TicketBind
	err = types.Decode(value, &bind)
	if err != nil {
		panic(err)
	}
	log15.Info("getBind++++++++++STEP2 bind.MinerAddress:", bind.MinerAddress)
	return bind.MinerAddress
}

//getBindLog old 是该地址之前绑定的账户地址，如果之前没有绑定挖矿，自己挖矿的则是返回自己的挖矿地址
//tbind 是新的绑定关系
func (mdb *MemoryStateDB) getBindLog(tbind *ty.TicketBind, old string) *types.ReceiptLog {

	log15.Info("getBindLog++++++++++STEP1 old:", old)
	log := &types.ReceiptLog{}
	log.Ty = ty.TyLogTicketBind
	r := &ty.ReceiptTicketBind{}
	r.ReturnAddress = tbind.ReturnAddress
	r.OldMinerAddress = old
	r.NewMinerAddress = tbind.MinerAddress
	log.Log = types.Encode(r)
	return log
}

//getBindKV 创建KV key=[]byte(mavl-ticket-tbind-ReturnAddress), value=types.Encode(tbind)
func (mdb *MemoryStateDB) getBindKV(tbind *ty.TicketBind) (kvset []*types.KeyValue) {
	value := types.Encode(tbind)
	kvset = append(kvset, &types.KeyValue{Key: bindKey(tbind.ReturnAddress), Value: value})
	return kvset
}

func (mdb *MemoryStateDB) saveBind(db dbm.KV, tbind *ty.TicketBind) {
	set := mdb.getBindKV(tbind)
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
}

func (mdb *MemoryStateDB) listBindTicket(returnAddr string) ([]*ty.Ticket, error) {
	var tickets []*ty.Ticket
	// status 0 -> 未成熟 1 -> 可挖矿 2 -> 已挖成功 3-> 已关闭
	values, err := mdb.LocalDB.List(calcTicketPrefix(returnAddr, 1), nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return tickets, nil
	}
	for i := 0; i < len(values); i++ {
		//ids.TicketIds = append(ids.TicketIds, string(values[i]))
		data, err := mdb.StateDB.Get(ticket.Key(string(values[i])))
		if err != nil {
			return tickets, err
		}
		var ticketInfo ty.Ticket
		//decode
		err = types.Decode(data, &ticketInfo)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticketInfo)
	}

	return tickets, nil

}
func calcTicketPrefix(addr string, status int32) []byte {
	key := fmt.Sprintf("LODB-ticket-tl:%s:%d", address.FormatAddrKey(addr), status)
	return []byte(key)
}
