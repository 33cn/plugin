// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"time"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
	token "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/pkg/errors"
)

const (
	lockingTime   = 12 * time.Hour //as currently one BTC tx may need wait quite long time
	lockBtcHeight = 12 * 6
	lockBtyAmount = 100 //coins
)

type relayLog struct {
	ty.RelayOrder
	*types.Chain33Config
}

func newRelayLog(order *ty.RelayOrder, cfg *types.Chain33Config) *relayLog {
	return &relayLog{RelayOrder: *order, Chain33Config: cfg}
}

func (r *relayLog) save(db dbm.KV) []*types.KeyValue {
	set := r.getKVSet()
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}

	return set
}

func (r *relayLog) getKVSet() (kvSet []*types.KeyValue) {
	value := types.Encode(&r.RelayOrder)
	key := []byte(r.Id)
	kvSet = append(kvSet, &types.KeyValue{Key: key, Value: value})

	if r.XTxHash != "" {
		key = []byte(calcCoinHash(r.XTxHash))
		kvSet = append(kvSet, &types.KeyValue{Key: key, Value: value})
	}

	return kvSet
}

func (r *relayLog) receiptLog(relayLogType int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = relayLogType

	receipt := &ty.ReceiptRelayLog{
		OrderId:         r.Id,
		CurStatus:       r.Status.String(),
		PreStatus:       r.PreStatus.String(),
		CreaterAddr:     r.CreaterAddr,
		LocalCoinAmount: types.FormatAmount2FloatDisplay(int64(r.LocalCoinAmount), r.GetCoinPrecision(), true),
		CoinOperation:   r.Operation,
		XCoin:           r.XCoin,
		XAmount:         types.FormatAmount2FloatDisplay(int64(r.XAmount), r.GetCoinPrecision(), true),
		XAddr:           r.XAddr,
		XTxHash:         r.XTxHash,
		XBlockWaits:     r.XBlockWaits,
		CreateTime:      r.CreateTime,
		AcceptAddr:      r.AcceptAddr,
		AcceptTime:      r.AcceptTime,
		ConfirmTime:     r.ConfirmTime,
		FinishTime:      r.FinishTime,
		XHeight:         r.XHeight,
		LocalCoinExec:   r.LocalCoinExec,
		LocalCoinSymbol: r.LocalCoinSymbol,
	}

	log.Log = types.Encode(receipt)

	return log
}

type relayDB struct {
	coinsAccount *account.DB
	db           dbm.KV
	txHash       []byte
	fromAddr     string
	blockTime    int64
	height       int64
	execAddr     string
	btc          *btcStore
	api          client.QueueProtocolAPI
}

func newRelayDB(r *relay, tx *types.Transaction) *relayDB {
	hash := tx.Hash()
	fromAddr := tx.From()
	btc := newBtcStore(r.GetLocalDB())
	return &relayDB{r.GetCoinsAccount(), r.GetStateDB(), hash,
		fromAddr, r.GetBlockTime(), r.GetHeight(), dapp.ExecAddress(string(tx.Execer)), btc, r.GetAPI()}
}

func (action *relayDB) getOrderByID(orderID string) (*ty.RelayOrder, error) {
	value, err := action.db.Get([]byte(calcRelayOrderID(orderID)))
	if err != nil {
		return nil, err
	}

	var order ty.RelayOrder
	if err = types.Decode(value, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (action *relayDB) getOrderByCoinHash(hash []byte) (*ty.RelayOrder, error) {
	value, err := action.db.Get(hash)
	if err != nil {
		return nil, err
	}

	var order ty.RelayOrder
	if err = types.Decode(value, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (action *relayDB) createAccount(exec, symbol string) (*account.DB, error) {
	var accDB *account.DB
	cfg := action.api.GetConfig()

	if symbol == "" {
		accDB = account.NewCoinsAccount(cfg)
		accDB.SetDB(action.db)
		return accDB, nil
	}
	if exec == "" {
		exec = token.TokenX
	}
	return account.NewAccountDB(cfg, exec, symbol, action.db)
}

func (action *relayDB) create(order *ty.RelayCreate) (*types.Receipt, error) {
	var receipt *types.Receipt
	var err error

	cfg := action.api.GetConfig()
	accDb, err := action.createAccount(order.LocalCoinExec, order.LocalCoinSymbol)
	if err != nil {
		return nil, errors.Wrapf(err, "relay create,exec=%s,sym=%s", order.LocalCoinExec, order.LocalCoinSymbol)
	}

	if order.Operation == ty.RelayOrderBuy {
		receipt, err = accDb.ExecFrozen(action.fromAddr, action.execAddr, int64(order.LocalCoinAmount))
		if err != nil {
			relaylog.Error("account.ExecFrozen relay ", "addrFrom", action.fromAddr, "execAddr", action.execAddr, "amount", order.LocalCoinAmount)
			return nil, err
		}

	} else {
		receipt, err = accDb.ExecFrozen(action.fromAddr, action.execAddr, lockBtyAmount*cfg.GetCoinPrecision())
		if err != nil {
			relaylog.Error("account.ExecFrozen relay ", "addrFrom", action.fromAddr, "execAddr", action.execAddr, "amount", lockBtyAmount)
			return nil, err
		}
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	uOrder := &ty.RelayOrder{
		Id:              calcRelayOrderID(common.ToHex(action.txHash)),
		Status:          ty.RelayOrderStatus_pending,
		PreStatus:       ty.RelayOrderStatus_init,
		Operation:       order.Operation,
		LocalCoinAmount: order.LocalCoinAmount,
		CreaterAddr:     action.fromAddr,
		XCoin:           order.XCoin,
		XAmount:         order.XAmount,
		XAddr:           order.XAddr,
		XBlockWaits:     order.XBlockWaits,
		CreateTime:      action.blockTime,
		Height:          action.height,
		LocalCoinExec:   order.LocalCoinExec,
		LocalCoinSymbol: order.LocalCoinSymbol,
	}

	height, err := action.btc.getLastBtcHeadHeight()
	if err != nil {
		return nil, err
	}
	uOrder.XHeight = uint64(height)

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	relayLog := newRelayLog(uOrder, action.api.GetConfig())
	sellOrderKV := relayLog.save(action.db)
	logs = append(logs, relayLog.receiptLog(ty.TyLogRelayCreate))
	kv = append(kv, sellOrderKV...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (action *relayDB) checkRevokeOrder(order *ty.RelayOrder) error {
	nowTime := time.Unix(action.blockTime, 0)
	var nowBtcHeight, subHeight int64

	nowBtcHeight, err := action.btc.getLastBtcHeadHeight()
	if err != nil {
		return err
	}

	if nowBtcHeight > 0 && order.XHeight > 0 && nowBtcHeight > int64(order.XHeight) {
		subHeight = nowBtcHeight - int64(order.XHeight)
	}

	if order.Status == ty.RelayOrderStatus_locking {
		acceptTime := time.Unix(order.AcceptTime, 0)
		if nowTime.Sub(acceptTime) < lockingTime && subHeight < lockBtcHeight {
			relaylog.Error("relay revoke locking", "duration", nowTime.Sub(acceptTime), "lockingTime", lockingTime, "subHeight", subHeight)
			return ty.ErrRelayBtcTxTimeErr
		}
	}

	if order.Status == ty.RelayOrderStatus_confirming {
		confirmTime := time.Unix(order.ConfirmTime, 0)
		if nowTime.Sub(confirmTime) < 4*lockingTime && subHeight < 4*lockBtcHeight {
			relaylog.Error("relay revoke confirming ", "duration", nowTime.Sub(confirmTime), "confirmTime", 4*lockingTime, "subHeight", subHeight)
			return ty.ErrRelayBtcTxTimeErr
		}
	}

	return nil

}

func (action *relayDB) revokeCreate(revoke *ty.RelayRevoke) (*types.Receipt, error) {
	order, err := action.getOrderByID(revoke.OrderId)
	if err != nil {
		return nil, ty.ErrRelayOrderNotExist
	}

	err = action.checkRevokeOrder(order)
	if err != nil {
		return nil, err
	}

	if order.Status == ty.RelayOrderStatus_init {
		return nil, ty.ErrRelayOrderStatusErr
	}

	if order.Status == ty.RelayOrderStatus_pending && revoke.Action == ty.RelayUnlock {
		return nil, ty.ErrRelayOrderParamErr
	}

	if order.Status == ty.RelayOrderStatus_finished {
		return nil, ty.ErrRelayOrderSoldout
	}
	if order.Status == ty.RelayOrderStatus_canceled {
		return nil, ty.ErrRelayOrderRevoked
	}

	if action.fromAddr != order.CreaterAddr {
		return nil, ty.ErrRelayReturnAddr
	}

	cfg := action.api.GetConfig()
	accDb, err := action.createAccount(order.LocalCoinExec, order.LocalCoinSymbol)
	if err != nil {
		return nil, errors.Wrapf(err, "relay revokeCreate,exec=%s,sym=%s", order.LocalCoinExec, order.LocalCoinSymbol)
	}

	var receipt *types.Receipt
	var receiptTransfer *types.Receipt
	if order.Operation == ty.RelayOrderBuy {
		receipt, err = accDb.ExecActive(order.CreaterAddr, action.execAddr, int64(order.LocalCoinAmount))
		if err != nil {
			relaylog.Error("revoke create", "addrFrom", order.CreaterAddr, "execAddr", action.execAddr, "amount", order.LocalCoinAmount)
			return nil, err
		}
	} else if order.Status != ty.RelayOrderStatus_pending {
		receipt, err = accDb.ExecActive(order.AcceptAddr, action.execAddr, int64(order.LocalCoinAmount))
		if err != nil {
			relaylog.Error("revoke create", "addrFrom", order.AcceptAddr, "execAddr", action.execAddr, "amount", order.LocalCoinAmount)
			return nil, err
		}

		receiptTransfer, err = accDb.ExecTransferFrozen(order.CreaterAddr, order.AcceptAddr, action.execAddr, lockBtyAmount*cfg.GetCoinPrecision())
		if err != nil {
			relaylog.Error("revokeAccept", "from", order.AcceptAddr, "to", order.CreaterAddr, "execAddr", action.execAddr, "amount", lockBtyAmount)
			return nil, err
		}
	}

	order.PreStatus = order.Status
	if revoke.Action == ty.RelayUnlock {
		order.Status = ty.RelayOrderStatus_pending
	} else {
		order.Status = ty.RelayOrderStatus_canceled
	}

	relayLog := newRelayLog(order, action.api.GetConfig())
	orderKV := relayLog.save(action.db)

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	if receipt != nil {
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}
	if receiptTransfer != nil {
		logs = append(logs, receiptTransfer.Logs...)
		kv = append(kv, receiptTransfer.KV...)
	}
	logs = append(logs, relayLog.receiptLog(ty.TyLogRelayRevokeCreate))
	kv = append(kv, orderKV...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (action *relayDB) accept(accept *ty.RelayAccept) (*types.Receipt, error) {
	order, err := action.getOrderByID(accept.OrderId)
	if err != nil {
		return nil, ty.ErrRelayOrderNotExist
	}

	if order.Status == ty.RelayOrderStatus_canceled {
		return nil, ty.ErrRelayOrderRevoked
	}
	if order.Status != ty.RelayOrderStatus_pending {
		return nil, ty.ErrRelayOrderSoldout
	}

	accDb, err := action.createAccount(order.LocalCoinExec, order.LocalCoinSymbol)
	if err != nil {
		return nil, errors.Wrapf(err, "relay accept,exec=%s,sym=%s", order.LocalCoinExec, order.LocalCoinSymbol)
	}

	cfg := action.api.GetConfig()
	var receipt *types.Receipt
	if order.Operation == ty.RelayOrderBuy {
		receipt, err = accDb.ExecFrozen(action.fromAddr, action.execAddr, lockBtyAmount*cfg.GetCoinPrecision())
		if err != nil {
			relaylog.Error("relay accept frozen fail ", "addrFrom", action.fromAddr, "execAddr", action.execAddr, "amount", lockBtyAmount)
			return nil, err
		}

	} else {
		if accept.XAddr == "" {
			relaylog.Error("accept, for sell operation, coinAddr needed")
			return nil, ty.ErrRelayOrderParamErr
		}

		order.XAddr = accept.XAddr
		order.XBlockWaits = accept.XBlockWaits

		receipt, err = accDb.ExecFrozen(action.fromAddr, action.execAddr, int64(order.LocalCoinAmount))
		if err != nil {
			relaylog.Error("relay accept frozen fail", "addrFrom", action.fromAddr, "execAddr", action.execAddr, "amount", order.LocalCoinAmount)
			return nil, err
		}
	}

	order.PreStatus = order.Status
	order.Status = ty.RelayOrderStatus_locking
	order.AcceptAddr = action.fromAddr
	order.AcceptTime = action.blockTime

	height, err := action.btc.getLastBtcHeadHeight()
	if err != nil {
		return nil, err
	}
	order.XHeight = uint64(height)

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	relayLog := newRelayLog(order, action.api.GetConfig())
	sellOrderKV := relayLog.save(action.db)

	logs = append(logs, relayLog.receiptLog(ty.TyLogRelayAccept))
	kv = append(kv, sellOrderKV...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil

}

func (action *relayDB) relayRevoke(revoke *ty.RelayRevoke) (*types.Receipt, error) {
	if revoke.Target == ty.RelayRevokeCreate {
		return action.revokeCreate(revoke)
	}

	return action.revokeAccept(revoke)
}

func (action *relayDB) revokeAccept(revoke *ty.RelayRevoke) (*types.Receipt, error) {
	order, err := action.getOrderByID(revoke.OrderId)
	if err != nil {
		return nil, ty.ErrRelayOrderNotExist
	}

	if order.Status == ty.RelayOrderStatus_pending || order.Status == ty.RelayOrderStatus_canceled {
		return nil, ty.ErrRelayOrderRevoked
	}
	if order.Status == ty.RelayOrderStatus_finished {
		return nil, ty.ErrRelayOrderFinished
	}

	cfg := action.api.GetConfig()
	err = action.checkRevokeOrder(order)
	if err != nil {
		return nil, err
	}

	if action.fromAddr != order.AcceptAddr {
		return nil, ty.ErrRelayReturnAddr
	}

	accDb, err := action.createAccount(order.LocalCoinExec, order.LocalCoinSymbol)
	if err != nil {
		return nil, errors.Wrapf(err, "relay revokeAccept,exec=%s,sym=%s", order.LocalCoinExec, order.LocalCoinSymbol)
	}

	var receipt *types.Receipt
	if order.Operation == ty.RelayOrderSell {
		receipt, err = accDb.ExecActive(order.AcceptAddr, action.execAddr, int64(order.LocalCoinAmount))
		if err != nil {
			relaylog.Error("revokeAccept", "addrFrom", order.AcceptAddr, "execAddr", action.execAddr, "amount", order.LocalCoinAmount)
			return nil, err
		}
		order.XAddr = ""
	}

	var receiptTransfer *types.Receipt
	if order.Operation == ty.RelayOrderBuy {
		receiptTransfer, err = accDb.ExecTransferFrozen(order.AcceptAddr, order.CreaterAddr, action.execAddr, lockBtyAmount*cfg.GetCoinPrecision())
		if err != nil {
			relaylog.Error("revokeAccept", "from", order.AcceptAddr, "to", order.CreaterAddr, "execAddr", action.execAddr, "amount", lockBtyAmount)
			return nil, err
		}
	}

	order.PreStatus = order.Status
	order.Status = ty.RelayOrderStatus_pending
	order.AcceptAddr = ""
	order.AcceptTime = 0

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	if receipt != nil {
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}
	if receiptTransfer != nil {
		logs = append(logs, receiptTransfer.Logs...)
		kv = append(kv, receiptTransfer.KV...)
	}
	relayLog := newRelayLog(order, action.api.GetConfig())
	sellOrderKV := relayLog.save(action.db)
	logs = append(logs, relayLog.receiptLog(ty.TyLogRelayRevokeAccept))
	kv = append(kv, sellOrderKV...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (action *relayDB) confirmTx(confirm *ty.RelayConfirmTx) (*types.Receipt, error) {
	order, err := action.getOrderByID(confirm.OrderId)
	if err != nil {
		return nil, ty.ErrRelayOrderNotExist
	}

	if order.Status == ty.RelayOrderStatus_pending {
		return nil, ty.ErrRelayOrderOnSell
	}
	if order.Status == ty.RelayOrderStatus_finished {
		return nil, ty.ErrRelayOrderSoldout
	}
	if order.Status == ty.RelayOrderStatus_canceled {
		return nil, ty.ErrRelayOrderRevoked
	}

	//report Error if coinTxHash has been used and not same orderID, if same orderID, means to modify the txHash
	coinTxOrder, _ := action.getOrderByCoinHash([]byte(calcCoinHash(confirm.TxHash)))
	if coinTxOrder != nil {
		if coinTxOrder.Id != confirm.OrderId {
			relaylog.Error("confirmTx", "coinTxHash", confirm.TxHash, "has been used in other order", coinTxOrder.Id)
			return nil, ty.ErrRelayCoinTxHashUsed
		}
	}

	var confirmAddr string
	if order.Operation == ty.RelayOrderBuy {
		confirmAddr = order.AcceptAddr
	} else {
		confirmAddr = order.CreaterAddr
	}
	if action.fromAddr != confirmAddr {
		return nil, ty.ErrRelayReturnAddr
	}

	order.PreStatus = order.Status
	order.Status = ty.RelayOrderStatus_confirming
	order.ConfirmTime = action.blockTime
	order.XTxHash = confirm.TxHash
	height, err := action.btc.getLastBtcHeadHeight()
	if err != nil {
		relaylog.Error("confirmTx Get Last BTC", "orderid", confirm.OrderId)
		return nil, err
	}
	order.XHeight = uint64(height)

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	relayLog := newRelayLog(order, action.api.GetConfig())
	sellOrderKV := relayLog.save(action.db)
	logs = append(logs, relayLog.receiptLog(ty.TyLogRelayConfirmTx))
	kv = append(kv, sellOrderKV...)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil

}

func (action *relayDB) verifyTx(verify *ty.RelayVerify) (*types.Receipt, error) {
	order, err := action.getOrderByID(verify.OrderId)
	if err != nil {
		return nil, ty.ErrRelayOrderNotExist
	}

	if order.Status == ty.RelayOrderStatus_finished {
		return nil, ty.ErrRelayOrderSoldout
	}
	if order.Status == ty.RelayOrderStatus_canceled {
		return nil, ty.ErrRelayOrderRevoked
	}
	if order.Status == ty.RelayOrderStatus_pending || order.Status == ty.RelayOrderStatus_locking {
		return nil, ty.ErrRelayOrderOnSell
	}

	cfg := action.api.GetConfig()
	err = action.btc.verifyBtcTx(verify, order)
	if err != nil {
		return nil, err
	}

	accDb, err := action.createAccount(order.LocalCoinExec, order.LocalCoinSymbol)
	if err != nil {
		return nil, errors.Wrapf(err, "relay revokeAccept,exec=%s,sym=%s", order.LocalCoinExec, order.LocalCoinSymbol)
	}

	var receipt *types.Receipt
	if order.Operation == ty.RelayOrderBuy {
		receipt, err = accDb.ExecTransferFrozen(order.CreaterAddr, order.AcceptAddr, action.execAddr, int64(order.LocalCoinAmount))
		if err != nil {
			relaylog.Error("verify buy transfer fail", "error", err.Error())
			return nil, err
		}

	} else {
		receipt, err = accDb.ExecTransferFrozen(order.AcceptAddr, order.CreaterAddr, action.execAddr, int64(order.LocalCoinAmount))
		if err != nil {
			relaylog.Error("verify sell transfer fail", "error", err.Error())
			return nil, err
		}
	}

	var receiptTransfer *types.Receipt
	if order.Operation == ty.RelayOrderBuy {
		receiptTransfer, err = accDb.ExecActive(order.AcceptAddr, action.execAddr, lockBtyAmount*cfg.GetCoinPrecision())
		if err != nil {
			relaylog.Error("verify exec active", "from", order.AcceptAddr, "amount", lockBtyAmount)
			return nil, err
		}

	} else {
		receiptTransfer, err = accDb.ExecActive(order.CreaterAddr, action.execAddr, lockBtyAmount*cfg.GetCoinPrecision())
		if err != nil {
			relaylog.Error("verify exec active", "from", order.CreaterAddr, "amount", lockBtyAmount)
			return nil, err
		}
	}

	order.PreStatus = order.Status
	order.Status = ty.RelayOrderStatus_finished
	order.FinishTime = action.blockTime
	order.FinishTxHash = common.ToHex(action.txHash)

	relayLog := newRelayLog(order, action.api.GetConfig())
	orderKV := relayLog.save(action.db)

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	logs = append(logs, receipt.Logs...)
	logs = append(logs, receiptTransfer.Logs...)
	logs = append(logs, relayLog.receiptLog(ty.TyLogRelayFinishTx))
	kv = append(kv, receipt.KV...)
	kv = append(kv, receiptTransfer.KV...)
	kv = append(kv, orderKV...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil

}

func saveBtcLastHead(db dbm.KV, head *ty.RelayLastRcvBtcHeader) (set []*types.KeyValue) {
	if head == nil || head.Header == nil {
		return nil
	}

	value := types.Encode(head)
	key := []byte(btcLastHead)
	set = append(set, &types.KeyValue{Key: key, Value: value})

	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
	return set
}

func getBtcLastHead(db dbm.KV) (*ty.RelayLastRcvBtcHeader, error) {
	value, err := db.Get([]byte(btcLastHead))
	if err != nil {
		return nil, err
	}
	var head ty.RelayLastRcvBtcHeader
	if err = types.Decode(value, &head); err != nil {
		return nil, err
	}

	return &head, nil
}

func (action *relayDB) saveBtcHeader(headers *ty.BtcHeaders, localDb dbm.KVDB) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var preHead = &ty.RelayLastRcvBtcHeader{}
	var receipt = &ty.ReceiptRelayRcvBTCHeaders{}

	cfg := action.api.GetConfig()
	subconfig := types.ConfSub(cfg, driverName)
	if action.fromAddr != subconfig.GStr("genesis") {
		return nil, types.ErrFromAddr
	}

	lastHead, err := getBtcLastHead(action.db)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}

	if lastHead != nil {
		preHead.Header = lastHead.Header
		preHead.BaseHeight = lastHead.BaseHeight

		receipt.LastHeight = lastHead.Header.Height
		receipt.LastBaseHeight = lastHead.BaseHeight
	}

	log := &types.ReceiptLog{}
	log.Ty = ty.TyLogRelayRcvBTCHead

	for _, head := range headers.BtcHeader {
		err := verifyBlockHeader(head, preHead, localDb)
		if err != nil {
			return nil, err
		}

		preHead.Header = head
		if head.IsReset {
			preHead.BaseHeight = head.Height
		}
		receipt.Headers = append(receipt.Headers, head)
	}

	receipt.NewHeight = preHead.Header.Height
	receipt.NewBaseHeight = preHead.BaseHeight

	log.Log = types.Encode(receipt)
	logs = append(logs, log)
	kv = saveBtcLastHead(action.db, preHead)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}
