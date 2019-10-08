// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
	tokenE "github.com/33cn/plugin/plugin/dapp/token/executor"
)

// List control
const (
	ListDESC    = int32(0)
	ListASC     = int32(1)
	DefultCount = int32(20)  //默认一次取多少条记录
	MaxCount    = int32(100) //最多取100条
)

const (
	decimal               = types.Coin // 1e8
	MaxDebtCeiling        = 10000      // 最大借贷限额
	MinLiquidationRatio   = 0.3        // 最小质押比
	MaxStabilityFee       = 1000       // 最大稳定费
	MaxLiquidationPenalty = 1000       // 最大清算罚金
	MinCreatorAccount     = 1000000    // 借贷创建者账户最小ccny余额
)

// CollateralizeDB def
type CollateralizeDB struct {
	pty.Collateralize
}

// GetKVSet for CollateralizeDB
func (coll *CollateralizeDB) GetKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(&coll.Collateralize)
	kvset = append(kvset, &types.KeyValue{Key: Key(coll.CollateralizeId), Value: value})
	return kvset
}

// Save for CollateralizeDB
func (coll *CollateralizeDB) Save(db dbm.KV) {
	set := coll.GetKVSet()
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
}

// Key for Collateralize
func Key(id string) (key []byte) {
	key = append(key, []byte("mavl-"+pty.CollateralizeX+"-")...)
	key = append(key, []byte(id)...)
	return key
}

// Action struct
type Action struct {
	coinsAccount *account.DB  // bty账户
	tokenAccount *account.DB  // ccny账户
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
	difficulty   uint64
	index        int
	Collateralize   *Collateralize
}

// NewCollateralizeAction generate New Action
func NewCollateralizeAction(c *Collateralize, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	tokenDb, err := account.NewAccountDB(tokenE.GetName(), pty.CCNYTokenName, c.GetStateDB())
	if err != nil {
		clog.Error("NewCollateralizeAction", "Get Account DB error", "err", err)
		return nil
	}

	return &Action{
		coinsAccount: c.GetCoinsAccount(), tokenAccount:tokenDb, db: c.GetStateDB(),
		txhash: hash, fromaddr: fromaddr, blocktime: c.GetBlockTime(),
		height: c.GetHeight(), execaddr: dapp.ExecAddress(string(tx.Execer)),
		difficulty: c.GetDifficulty(), index: index, Collateralize: c}
}

// GetCollCommonRecipt generate logs for Collateralize common action
func (action *Action) GetCollCommonRecipt(collateralize *pty.Collateralize, preStatus int32) *pty.ReceiptCollateralize {
	c := &pty.ReceiptCollateralize{}
	c.CollateralizeId = collateralize.CollateralizeId
	c.PreStatus = preStatus
	c.Status = collateralize.Status
	return c
}

// GetCreateReceiptLog generate logs for Collateralize create action
func (action *Action) GetCreateReceiptLog(collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCollateralizeCreate

	c := action.GetCollCommonRecipt(collateralize, preStatus)
	c.CreateAddr = action.fromaddr

	log.Log = types.Encode(c)

	return log
}

// GetBorrowReceiptLog generate logs for Collateralize borrow action
// TODO
func (action *Action) GetBorrowReceiptLog(collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCollateralizeBorrow

	c := action.GetCollCommonRecipt(collateralize, preStatus)
	c.AccountAddr = action.fromaddr

	log.Log = types.Encode(c)

	return log
}

// GetRepayReceiptLog generate logs for Collateralize Repay action
// TODO
func (action *Action) GetRepayReceiptLog(Collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCollateralizeRepay

	c := action.GetCollCommonRecipt(Collateralize, preStatus)

	log.Log = types.Encode(c)

	return log
}

// GetAppendReceiptLog generate logs for Collateralize Repay action
// TODO
func (action *Action) GetAppendReceiptLog(Collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	return nil
}

// GetCloseReceiptLog generate logs for Collateralize close action
// TODO
func (action *Action) GetCloseReceiptLog(Collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	return nil
}

// GetIndex returns index in block
func (action *Action) GetIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

// CollateralizeCreate 创建借贷，持有一定数量ccny的用户可创建借贷，提供给其他用户借贷
func (action *Action) CollateralizeCreate(create *pty.CollateralizeCreate) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	collateralizeID := common.ToHex(action.txhash)

	// 参数校验
	if create.DebtCeiling > MaxDebtCeiling || create.DebtCeiling < 0 ||
		create.LiquidationRatio < MinLiquidationRatio || create.LiquidationRatio >= 1 ||
		create.StabilityFee > MaxStabilityFee || create.StabilityFee < 0 ||
		create.LiquidationPenalty > MaxLiquidationPenalty || create.LiquidationPenalty < 0 {
		return nil, pty.ErrRiskParam
	}

	// 检查ccny余额
	if !action.CheckExecTokenAccount(action.fromaddr, MinCreatorAccount, false) {
		return nil, types.ErrInsufficientBalance
	}

	// 查找ID是否重复
	_, err := findCollateralize(action.db, collateralizeID)
	if err != types.ErrNotFound {
		clog.Error("CollateralizeCreate", "CollateralizeCreate repeated", collateralizeID)
		return nil, pty.ErrCollateralizeRepeatHash
	}

	// TODO ccny是否需要冻结

	// 构造coll结构
	coll := &CollateralizeDB{}
	coll.CollateralizeId = collateralizeID
	coll.LiquidationRatio = create.LiquidationRatio
	coll.TotalBalance = create.TotalBalance
	coll.DebtCeiling = create.DebtCeiling
	coll.LiquidationPenalty = create.LiquidationPenalty
	coll.StabilityFee = create.StabilityFee
	coll.CreateAddr = action.fromaddr
	coll.Status = pty.CollateralizeActionCreate

	clog.Debug("CollateralizeCreate created", "CollateralizeID", collateralizeID, "TotalBalance", coll.TotalBalance)

	// 保存
	coll.Save(action.db)
	kv = append(kv, coll.GetKVSet()...)

	receiptLog := action.GetCreateReceiptLog(&coll.Collateralize, 0)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// 根据最近抵押物价格计算需要冻结的BTY数量
func getBtyNumToFrozen(value int64, price float32, ratio float32) (int64,error) {
	if price == 0 {
		clog.Error("Bty price should greate to 0")
		return 0, pty.ErrPriceZero
	}

    btyValue := float32(value)/ratio
    btyNum := int64(btyValue/price) + 1

    return btyNum, nil
}

// 计算清算价格
// value:借出ccny数量， colValue:抵押物数量， price:抵押物价格
func calcRepayPrice(value int64, colValue int64) float32 {
	liquidationRation := float32(value) / float32(colValue)
	repayPrice := liquidationRation * pty.CollateralizeRepayRatio

	return repayPrice
}

// 获取最近抵押物价格
func (action *Action)getLatestPrice(db dbm.KV, assetType int32) (float32, error) {
	data, err := db.Get(calcCollateralizeLatestPriceKey())
	if err != nil {
		clog.Debug("getLatestPrice", "get", err)
		return -1, err
	}
	var price pty.AssetPriceRecord
	//decode
	err = types.Decode(data, &price)
	if err != nil {
		clog.Debug("getLatestPrice", "decode", err)
		return -1, err
	}

	switch assetType {
	case pty.CollateralizeAssetTypeBty:
		return price.BtcPrice, nil
	case pty.CollateralizeAssetTypeBtc:
		return price.BtcPrice, nil
	case pty.CollateralizeAssetTypeEth:
		return price.EthPrice, nil
	default:
		return -1, pty.ErrAssetType
	}
}

// CheckExecAccountBalance 检查账户抵押物余额
func (action *Action) CheckExecAccountBalance(fromAddr string, ToFrozen, ToActive int64) bool {
	acc := action.coinsAccount.LoadExecAccount(fromAddr, action.execaddr)
	if acc.GetBalance() >= ToFrozen && acc.GetFrozen() >= ToActive {
		return true
	}
	return false
}

// CheckExecAccount 检查账户token余额
func (action *Action) CheckExecTokenAccount(addr string, amount int64, isFrozen bool) bool {
	acc := action.tokenAccount.LoadExecAccount(addr, action.execaddr)
	if isFrozen {
		if acc.GetFrozen() >= amount {
			return true
		}
	} else {
		if acc.GetBalance() >= amount {
			return true
		}
	}

	return false
}

// CollateralizeBorrow 用户质押bty借出ccny
// TODO 考虑同一用户多次借贷的场景
func (action *Action) CollateralizeBorrow(borrow *pty.CollateralizeBorrow) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 查找对应的借贷ID
	// TODO 是否需要合约自动查找？
	collateralize, err := findCollateralize(action.db, borrow.CollateralizeId)
	if err != nil {
		clog.Error("CollateralizeBorrow", "CollateralizeId", borrow.CollateralizeId)
		return nil, err
	}

	coll := &CollateralizeDB{*collateralize}
	preStatus := coll.Status

	// 状态检查
	if coll.Status == pty.CollateralizeStatusClose {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "status", coll.Status, "err", pty.ErrCollateralizeStatus)
		return nil, pty.ErrCollateralizeStatus
	}

	// 借贷金额检查
	if borrow.GetValue() <= 0 || borrow.GetValue() > coll.DebtCeiling {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "borrow value", borrow.GetValue(), "err", pty.ErrCollateralizeExceedDebtCeiling)
		return nil, pty.ErrCollateralizeExceedDebtCeiling
	}

	clog.Debug("CollateralizeBorrow", "value", borrow.GetValue())

	// 获取抵押物价格
	lastPrice, err := action.getLatestPrice(action.db, coll.CollType)
	if err != nil {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", err)
		return nil, err
	}

	// 根据价格和需要借贷的金额，计算需要质押的抵押物数量
	btyFrozen, err := getBtyNumToFrozen(borrow.Value, lastPrice, coll.LiquidationRatio)
	if err != nil {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", err)
		return nil, err
	}

	// 检查抵押物账户余额
	if !action.CheckExecAccountBalance(action.fromaddr, btyFrozen, 0) {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", types.ErrNoBalance)
		return nil, types.ErrNoBalance
	}

	// 抵押物转账
	receipt, err := action.coinsAccount.ExecTransfer(action.fromaddr, coll.CreateAddr, action.execaddr, btyFrozen*decimal)
	if err != nil {
		clog.Error("CollateralizeBorrow.ExecTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", btyFrozen)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 抵押物冻结
	receipt, err = action.coinsAccount.ExecFrozen(coll.CreateAddr, action.execaddr, btyFrozen)
	if err != nil {
		clog.Error("CollateralizeBorrow.Frozen", "addr", coll.CreateAddr, "execaddr", action.execaddr, "amount", btyFrozen)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 借出ccny
	receipt, err = action.tokenAccount.ExecTransfer(coll.CreateAddr, action.fromaddr, action.execaddr, borrow.Value)
	if err != nil {
		clog.Error("CollateralizeBorrow.ExecTokenTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", borrow.Value)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 构造借出记录
	borrowRecord := &pty.BorrowRecord{}
	borrowRecord.AccountAddr = action.fromaddr
	borrowRecord.CollateralValue = btyFrozen
	borrowRecord.StartTime = action.blocktime
	borrowRecord.CollateralPrice = lastPrice
	borrowRecord.DebtValue = borrow.Value
	borrowRecord.LiquidationPrice = coll.LiquidationRatio * lastPrice * pty.CollateralizeRepayRatio
	borrowRecord.Status = pty.CollateralizeUserStatusCreate

	// 记录当前借贷的最高自动清算价格
	if coll.LatestRepayPrice < borrowRecord.LiquidationPrice {
		coll.LatestRepayPrice = borrowRecord.LiquidationPrice
	}

	// 保存
	coll.BorrowRecords = append(coll.BorrowRecords, borrowRecord)
	coll.Status = pty.CollateralizeStatusCreated
	coll.Save(action.db)
	kv = append(kv, coll.GetKVSet()...)

	receiptLog := action.GetBorrowReceiptLog(&coll.Collateralize, preStatus)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// CollateralizeRepay 用户主动清算
func (action *Action) CollateralizeRepay(repay *pty.CollateralizeRepay) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	// 找到相应的借贷
	// TODO 是否需要合约自动查找
	Collateralize, err := findCollateralize(action.db, repay.CollateralizeId)
	if err != nil {
		clog.Error("CollateralizeRepay", "CollID", repay.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", "Can not find collateralize Id")
		return nil, err
	}

	coll := &CollateralizeDB{*Collateralize}

	preStatus := coll.Status

	// 状态检查
	if coll.Status != pty.CollateralizeStatusCreated {
		clog.Error("CollateralizeRepay", "CollID", repay.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", "status error", "Status", coll.Status)
		return nil, pty.ErrCollateralizeStatus
	}

	// 查找借出记录
	var borrowRecord *pty.BorrowRecord
	for _, record := range coll.BorrowRecords {
		if record.AccountAddr == action.fromaddr {
			borrowRecord = record
		}
	}

	if borrowRecord == nil {
		clog.Error("CollateralizeRepay", "CollID", repay.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", "Can not find borrow record")
		return nil, pty.ErrRecordNotExist
	}

	// 检查清算金额（默认全部清算，部分清算在其他接口）
	if repay.Value != borrowRecord.DebtValue {
		clog.Error("CollateralizeRepay", "CollID", repay.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "collateralize value", borrowRecord.DebtValue, "err", pty.ErrRepayValueInsufficient)
		return nil, pty.ErrRepayValueInsufficient
	}

	// 检查
	// TODO 暂时未考虑利息
	if !action.CheckExecTokenAccount(action.fromaddr, repay.Value, false) {
		clog.Error("CollateralizeRepay", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", types.ErrInsufficientBalance)
		return nil, types.ErrNoBalance
	}

	// ccny转移
	receipt, err = action.tokenAccount.ExecTransfer(action.fromaddr, coll.CreateAddr, action.execaddr, repay.Value)
	if err != nil {
		clog.Error("CollateralizeRepay.ExecTokenTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", repay.Value)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 抵押物归还
	receipt, err = action.coinsAccount.ExecTransferFrozen(coll.CreateAddr, action.execaddr, action.execaddr, borrowRecord.CollateralValue)
	if err != nil {
		clog.Error("CollateralizeRepay.ExecTransferFrozen", "addr", coll.CreateAddr, "execaddr", action.execaddr, "amount", borrowRecord.CollateralValue)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 借贷记录关闭
	borrowRecord.Status = pty.CollateralizeUserStatusClose

	// 保存
	coll.Save(action.db)
	kv = append(kv, coll.GetKVSet()...)

	receiptLog := action.GetRepayReceiptLog(&coll.Collateralize, preStatus)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// CollateralizeAppend 追加抵押物
func (action *Action) CollateralizeAppend(cAppend *pty.CollateralizeAppend) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	// 查找对应的借贷ID
	// TODO 是否需要合约自动查找？
	collateralize, err := findCollateralize(action.db, cAppend.CollateralizeId)
	if err != nil {
		clog.Error("CollateralizeAppend", "CollateralizeId", cAppend.CollateralizeId)
		return nil, err
	}

	coll := &CollateralizeDB{*collateralize}
	preStatus := coll.Status

	// 状态检查
	if coll.Status != pty.CollateralizeStatusCreated {
		clog.Error("CollateralizeAppend", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "status", coll.Status, "err", pty.ErrCollateralizeStatus)
		return nil, pty.ErrCollateralizeStatus
	}

	// 查找借出记录
	var borrowRecord *pty.BorrowRecord
	for _, record := range coll.BorrowRecords {
		if record.AccountAddr == action.fromaddr {
			borrowRecord = record
		}
	}

	if borrowRecord == nil {
		clog.Error("CollateralizeAppend", "CollID", cAppend.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", "Can not find borrow record")
		return nil, pty.ErrRecordNotExist
	}

	clog.Debug("CollateralizeAppend", "value", cAppend.CollateralValue)

	// 获取抵押物价格
	lastPrice, err := action.getLatestPrice(action.db, coll.CollType)
	if err != nil {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", err)
		return nil, err
	}

	// 检查抵押物账户余额
	if !action.CheckExecAccountBalance(action.fromaddr, cAppend.CollateralValue, 0) {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", types.ErrNoBalance)
		return nil, types.ErrNoBalance
	}

	// 抵押物转账
	receipt, err := action.coinsAccount.ExecTransfer(action.fromaddr, coll.CreateAddr, action.execaddr, cAppend.CollateralValue*decimal)
	if err != nil {
		clog.Error("CollateralizeBorrow.ExecTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", cAppend.CollateralValue)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 抵押物冻结
	receipt, err = action.coinsAccount.ExecFrozen(coll.CreateAddr, action.execaddr, cAppend.CollateralValue)
	if err != nil {
		clog.Error("CollateralizeBorrow.Frozen", "addr", coll.CreateAddr, "execaddr", action.execaddr, "amount", cAppend.CollateralValue)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 构造借出记录
	borrowRecord.CollateralValue += cAppend.CollateralValue
	borrowRecord.CollateralPrice = lastPrice
	borrowRecord.LiquidationPrice = calcRepayPrice(borrowRecord.DebtValue, borrowRecord.CollateralValue)

	// 记录当前借贷的最高自动清算价格
	if coll.LatestRepayPrice < borrowRecord.LiquidationPrice {
		coll.LatestRepayPrice = borrowRecord.LiquidationPrice
	}

	// 保存
	coll.BorrowRecords = append(coll.BorrowRecords, borrowRecord)
	coll.Status = pty.CollateralizeStatusCreated
	coll.Save(action.db)
	kv = append(kv, coll.GetKVSet()...)

	receiptLog := action.GetAppendReceiptLog(&coll.Collateralize, preStatus)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// TODO 部分清算

// CollateralizeFeed 喂价
func (action *Action) CollateralizeFeed(repay *pty.CollateralizeFeed) (*types.Receipt, error) {
	//TODO
	return nil, nil
}

// CollateralizeClose 终止借贷
func (action *Action) CollateralizeClose(draw *pty.CollateralizeClose) (*types.Receipt, error) {
	//TODO
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	//var receipt *types.Receipt

	Collateralize, err := findCollateralize(action.db, draw.CollateralizeId)
	if err != nil {
		clog.Error("CollateralizeBuy", "CollateralizeId", draw.CollateralizeId)
		return nil, err
	}

	coll := &CollateralizeDB{*Collateralize}
	preStatus := coll.Status

	if action.fromaddr != coll.CreateAddr {
		return nil, pty.ErrCollateralizeErrCloser
	}

	if coll.Status == pty.CollateralizeStatusClose {
		return nil, pty.ErrCollateralizeStatus
	}

	clog.Debug("CollateralizeClose", )

	//TODO
	coll.Save(action.db)
	kv = append(kv, coll.GetKVSet()...)

	receiptLog := action.GetCloseReceiptLog(&coll.Collateralize, preStatus)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

// 查找借贷
func findCollateralize(db dbm.KV, CollateralizeID string) (*pty.Collateralize, error) {
	data, err := db.Get(Key(CollateralizeID))
	if err != nil {
		clog.Debug("findCollateralize", "get", err)
		return nil, err
	}
	var coll pty.Collateralize
	//decode
	err = types.Decode(data, &coll)
	if err != nil {
		clog.Debug("findCollateralize", "decode", err)
		return nil, err
	}
	return &coll, nil
}
