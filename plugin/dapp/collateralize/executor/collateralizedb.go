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
	issuanceE "github.com/33cn/plugin/plugin/dapp/issuance/types"
)

// List control
const (
	ListDESC    = int32(0)   // list降序
	ListASC     = int32(1)   // list升序
	DefultCount = int32(20)  // 默认一次取多少条记录
	MaxCount    = int32(100) // 最多取100条
)

const (
	Coin                      = types.Coin      // 1e8
	DefaultDebtCeiling        = 10000           // 默认借贷限额
	DefaultLiquidationRatio   = 0.4             // 默认质押比
	DefaultStabilityFeeRation = 0.08            // 默认稳定费
	DefaultPeriod             = 3600 * 24 * 365 // 默认合约限期
	PriceWarningRate          = 1.3             // 价格提前预警率
	ExpireWarningTime         = 3600 * 24 * 10  // 提前10天超时预警
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

// Key for CollateralizeConfig
func ConfigKey() (key []byte) {
	key = append(key, []byte("mavl-"+pty.CollateralizeX+"config")...)
	return key
}

// Key for CollateralizeAddrConfig
func AddrKey() (key []byte) {
	key = append(key, []byte("mavl-"+issuanceE.IssuanceX+"addr")...)
	return key
}

// Action struct
type Action struct {
	coinsAccount *account.DB  // bty账户
	tokenAccount *account.DB  // ccny账户
	db           dbm.KV
	localDB      dbm.Lister
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
		coinsAccount: c.GetCoinsAccount(), tokenAccount:tokenDb, db: c.GetStateDB(), localDB:c.GetLocalDB(),
		txhash: hash, fromaddr: fromaddr, blocktime: c.GetBlockTime(), height: c.GetHeight(),
		execaddr: dapp.ExecAddress(string(tx.Execer)), difficulty: c.GetDifficulty(), index: index, Collateralize: c}
}

// GetCollCommonRecipt generate logs for Collateralize common action
func (action *Action) GetCollCommonRecipt(collateralize *pty.Collateralize, preStatus int32) *pty.ReceiptCollateralize {
	c := &pty.ReceiptCollateralize{}
	c.CollateralizeId = collateralize.CollateralizeId
	c.PreStatus = preStatus
	c.Status = collateralize.Status
	c.Index = action.GetIndex()
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
func (action *Action) GetBorrowReceiptLog(collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCollateralizeBorrow

	c := action.GetCollCommonRecipt(collateralize, preStatus)
	c.AccountAddr = action.fromaddr

	log.Log = types.Encode(c)

	return log
}

// GetRepayReceiptLog generate logs for Collateralize Repay action
func (action *Action) GetRepayReceiptLog(collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCollateralizeRepay

	c := action.GetCollCommonRecipt(collateralize, preStatus)

	log.Log = types.Encode(c)

	return log
}

// GetAppendReceiptLog generate logs for Collateralize append action
func (action *Action) GetAppendReceiptLog(collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCollateralizeAppend

	c := action.GetCollCommonRecipt(collateralize, preStatus)

	log.Log = types.Encode(c)

	return log
}

// GetFeedReceiptLog generate logs for Collateralize price feed action
func (action *Action) GetFeedReceiptLog(collateralize *pty.Collateralize, borrowRecord *pty.BorrowRecord) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCollateralizeFeed

	c := action.GetCollCommonRecipt(collateralize, borrowRecord.PreStatus)
	c.AccountAddr = borrowRecord.AccountAddr
	c.RecordStatus = borrowRecord.Status

	log.Log = types.Encode(c)

	return log
}

// GetCloseReceiptLog generate logs for Collateralize close action
func (action *Action) GetCloseReceiptLog(Collateralize *pty.Collateralize, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogCollateralizeClose

	c := action.GetCollCommonRecipt(Collateralize, preStatus)

	log.Log = types.Encode(c)

	return log
}

// GetIndex returns index in block
func (action *Action) GetIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

func getLatestLiquidationPrice(coll *pty.Collateralize) float32 {
	var latest float32
	for _, collRecord := range coll.BorrowRecords {
		if collRecord.LiquidationPrice > latest {
			latest = collRecord.LiquidationPrice
		}
	}

	return latest
}

func getLatestExpireTime(coll *pty.Collateralize) int64 {
	var latest int64 = 0x7fffffffffffffff

	for _, collRecord := range coll.BorrowRecords {
		if collRecord.ExpireTime < latest {
			latest = collRecord.ExpireTime
		}
	}

	return latest
}

// CollateralizeConfig 设置全局借贷参数（管理员权限）
func (action *Action) CollateralizeManage(manage *pty.CollateralizeManage) (*types.Receipt, error) {
	var kv []*types.KeyValue
	var receipt *types.Receipt

	// 是否配置管理用户
	if !isRightAddr(configKey, action.fromaddr, action.db) {
		clog.Error("CollateralizeManage", "addr", action.fromaddr, "error", "Address has no permission to config")
		return nil, pty.ErrPermissionDeny
	}

	// 配置借贷参数
	if manage.DebtCeiling < 0 || manage.LiquidationRatio < 0 || manage.LiquidationRatio >= 1 ||
		manage.StabilityFeeRatio < 0 || manage.StabilityFeeRatio >= 1 {
		return nil, pty.ErrRiskParam
	}

	collConfig := &pty.CollateralizeManage{}
	collConfig.StabilityFeeRatio = manage.StabilityFeeRatio
	collConfig.Period = manage.Period
	collConfig.LiquidationRatio = manage.LiquidationRatio
	collConfig.DebtCeiling = manage.DebtCeiling

	value := types.Encode(collConfig)
	action.db.Set(ConfigKey(), value)
	kv = append(kv, &types.KeyValue{Key: ConfigKey(), Value: value})

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: nil}
	return receipt, nil
}

func (action *Action) getCollateralizeConfig() (*pty.CollateralizeManage, error) {
	data, err := action.db.Get(ConfigKey())
	if err != nil {
		clog.Debug("getCollateralizeConfig", "error", err)
		return nil, err
	}

	var collCfg pty.CollateralizeManage
	err = types.Decode(data, &collCfg)
	if err != nil {
		clog.Debug("getCollateralizeConfig", "decode", err)
		return nil, err
	}
	return &collCfg, nil
}


func (action *Action) getSuperAddr() []string {
	data, err := action.db.Get(AddrKey())
	if err != nil {
		clog.Error("getSuperAddr", "error", err)
		return nil
	}

	var addrStore pty.CollateralizeAddr
	err = types.Decode(data, &addrStore)
	if err != nil {
		clog.Debug("getSuperAddr", "decode", err)
		return nil
	}

	return addrStore.SuperAddrs
}

func isSuperAddr(super []string, addr string) bool {
	if super == nil || len(super) == 0 {
		return false
	}

	for _, superAddr := range super {
		if superAddr == addr {
			return true
		}
	}

	return false
}

// CollateralizeCreate 创建借贷，持有一定数量ccny的用户可创建借贷，提供给其他用户借贷
func (action *Action) CollateralizeCreate(create *pty.CollateralizeCreate) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	superAddr := action.getSuperAddr()
	if !isSuperAddr(superAddr, action.fromaddr) {
		clog.Error("CollateralizeCreate", "error", "CollateralizeCreate need super address")
		return nil, pty.ErrPermissionDeny
	}

	collateralizeID := common.ToHex(action.txhash)

	// 检查ccny余额
	if !action.CheckExecTokenAccount(action.fromaddr, create.TotalBalance, false) {
		return nil, types.ErrInsufficientBalance
	}

	// 查找ID是否重复
	_, err := queryCollateralizeByID(action.db, collateralizeID)
	if err != types.ErrNotFound {
		clog.Error("CollateralizeCreate", "CollateralizeCreate repeated", collateralizeID)
		return nil, pty.ErrCollateralizeRepeatHash
	}

	// 冻结ccny
	receipt, err = action.tokenAccount.ExecFrozen(action.fromaddr, action.execaddr, create.TotalBalance)
	if err != nil {
		clog.Error("CollateralizeCreate.Frozen", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", create.TotalBalance)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 获取借贷配置
	var collcfg *pty.CollateralizeManage
	cfg, err := action.getCollateralizeConfig()
	if err != nil {
		collcfg = &pty.CollateralizeManage{DebtCeiling:DefaultDebtCeiling, LiquidationRatio:DefaultLiquidationRatio, StabilityFeeRatio:DefaultStabilityFeeRation, Period:DefaultPeriod}
	} else {
		collcfg = cfg
	}

	// 构造coll结构
	coll := &CollateralizeDB{}
	coll.CollateralizeId = collateralizeID
	coll.LiquidationRatio = collcfg.LiquidationRatio
	coll.TotalBalance = create.TotalBalance
	coll.DebtCeiling = collcfg.DebtCeiling
	coll.StabilityFeeRatio = collcfg.StabilityFeeRatio
	coll.Period = collcfg.Period
	coll.Balance = create.TotalBalance
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
		return 0, pty.ErrPriceInvalid
	}

    btyValue := float32(value)/ratio
    btyNum := int64(btyValue/price) + 1

    return btyNum, nil
}

// 计算清算价格
// value:借出ccny数量， colValue:抵押物数量， price:抵押物价格
func calcLiquidationPrice(value int64, colValue int64) float32 {
	liquidationRation := float32(value) / float32(colValue)
	liquidationPrice := liquidationRation * pty.CollateralizePreLiquidationRatio

	return liquidationPrice
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
	collateralize, err := queryCollateralizeByID(action.db, borrow.CollateralizeId)
	if err != nil {
		clog.Error("CollateralizeBorrow", "CollateralizeId", borrow.CollateralizeId, "err", err)
		return nil, err
	}

	// 状态检查
	if collateralize.Status == pty.CollateralizeStatusClose {
		clog.Error("CollateralizeBorrow", "CollID", collateralize.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "status", collateralize.Status, "err", pty.ErrCollateralizeStatus)
		return nil, pty.ErrCollateralizeStatus
	}

	// 一个地址在一期借贷中只允许借出一次
	for _, record := range collateralize.BorrowRecords {
		if record.AccountAddr == action.fromaddr {
			clog.Error("CollateralizeBorrow","CollateralizeId", borrow.CollateralizeId, action.fromaddr, "execaddr", action.execaddr, "err", pty.ErrCollateralizeAccountExist)
			return nil, err
		}
	}
	coll := &CollateralizeDB{*collateralize}
	preStatus := coll.Status

	// 借贷金额检查
	if borrow.GetValue() <= 0 {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "borrow value", borrow.GetValue(), "err", types.ErrInvalidParam)
		return  nil, types.ErrInvalidParam
	}

	// 借贷金额不超过个人限额
	if borrow.GetValue() > coll.DebtCeiling {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "borrow value", borrow.GetValue(), "err", pty.ErrCollateralizeExceedDebtCeiling)
		return nil, pty.ErrCollateralizeExceedDebtCeiling
	}

	// 借贷金额不超过当前可借贷金额
	if borrow.GetValue() > coll.Balance {
		clog.Error("CollateralizeBorrow", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "borrow value", borrow.GetValue(), "err", pty.ErrCollateralizeLowBalance)
		return nil, pty.ErrCollateralizeLowBalance
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
	receipt, err := action.coinsAccount.ExecTransfer(action.fromaddr, coll.CreateAddr, action.execaddr, btyFrozen*Coin)
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
	borrowRecord.LiquidationPrice = coll.LiquidationRatio * lastPrice * pty.CollateralizePreLiquidationRatio
	borrowRecord.Status = pty.CollateralizeUserStatusCreate
	borrowRecord.ExpireTime = action.blocktime + coll.Period

	// 记录当前借贷的最高自动清算价格
	if coll.LatestLiquidationPrice < borrowRecord.LiquidationPrice {
		coll.LatestLiquidationPrice = borrowRecord.LiquidationPrice
	}

	// 保存
	coll.BorrowRecords = append(coll.BorrowRecords, borrowRecord)
	coll.Status = pty.CollateralizeStatusCreated
	coll.Balance -= borrow.Value
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
	collateralize, err := queryCollateralizeByID(action.db, repay.CollateralizeId)
	if err != nil {
		clog.Error("CollateralizeRepay", "CollID", repay.CollateralizeId, "err", err)
		return nil, err
	}

	coll := &CollateralizeDB{*collateralize}

	preStatus := coll.Status

	// 状态检查
	if coll.Status != pty.CollateralizeStatusCreated {
		clog.Error("CollateralizeRepay", "CollID", repay.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", "status error", "Status", coll.Status)
		return nil, pty.ErrCollateralizeStatus
	}

	// 查找借出记录
	var borrowRecord *pty.BorrowRecord
	var index int
	for i, record := range coll.BorrowRecords {
		if record.AccountAddr == action.fromaddr {
			borrowRecord = record
			index = i
			break
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

	// 借贷金额+利息
	realRepay := repay.Value + int64(float32(repay.Value) * coll.StabilityFeeRatio) + 1

	// 检查
	if !action.CheckExecTokenAccount(action.fromaddr, realRepay, false) {
		clog.Error("CollateralizeRepay", "CollID", coll.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", types.ErrInsufficientBalance)
		return nil, types.ErrNoBalance
	}

	// ccny转移
	receipt, err = action.tokenAccount.ExecTransfer(action.fromaddr, coll.CreateAddr, action.execaddr, realRepay)
	if err != nil {
		clog.Error("CollateralizeRepay.ExecTokenTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", realRepay)
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
	borrowRecord.PreStatus = borrowRecord.Status
	borrowRecord.Status = pty.CollateralizeUserStatusClose

	// 保存
	coll.Balance += repay.Value
	coll.BorrowRecords = append(coll.BorrowRecords[:index], coll.BorrowRecords[index+1:]...)
	coll.InvalidRecords = append(coll.InvalidRecords, borrowRecord)
	coll.LatestLiquidationPrice = getLatestLiquidationPrice(&coll.Collateralize)
	coll.LatestExpireTime = getLatestExpireTime(&coll.Collateralize)
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
	collateralize, err := queryCollateralizeByID(action.db, cAppend.CollateralizeId)
	if err != nil {
		clog.Error("CollateralizeAppend", "CollateralizeId", cAppend.CollateralizeId, "err", err)
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
	receipt, err := action.coinsAccount.ExecTransfer(action.fromaddr, coll.CreateAddr, action.execaddr, cAppend.CollateralValue*Coin)
	if err != nil {
		clog.Error("CollateralizeBorrow.ExecTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", cAppend.CollateralValue, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 抵押物冻结
	receipt, err = action.coinsAccount.ExecFrozen(coll.CreateAddr, action.execaddr, cAppend.CollateralValue)
	if err != nil {
		clog.Error("CollateralizeBorrow.Frozen", "addr", coll.CreateAddr, "execaddr", action.execaddr, "amount", cAppend.CollateralValue, "err", err)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 构造借出记录
	borrowRecord.CollateralValue += cAppend.CollateralValue
	borrowRecord.CollateralPrice = lastPrice
	borrowRecord.LiquidationPrice = calcLiquidationPrice(borrowRecord.DebtValue, borrowRecord.CollateralValue)
	if borrowRecord.LiquidationPrice * PriceWarningRate < lastPrice {
		// 告警解除
		if borrowRecord.Status == pty.CollateralizeUserStatusWarning {
			borrowRecord.PreStatus = borrowRecord.Status
			borrowRecord.Status = pty.CollateralizeStatusCreated
		}
	}

	// 记录当前借贷的最高自动清算价格
	coll.LatestLiquidationPrice = getLatestLiquidationPrice(&coll.Collateralize)
	coll.LatestExpireTime = getLatestExpireTime(&coll.Collateralize)
	coll.Save(action.db)

	kv = append(kv, coll.GetKVSet()...)

	receiptLog := action.GetAppendReceiptLog(&coll.Collateralize, preStatus)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		return nil, err
	}
	return value, nil
}

const (
	priceFeedKey = "collateralize-price-feed"
	guarantorKey = "collateralize-guarantor"
	configKey    = "collateralize-config"
)

func isRightAddr(key string, addr string, db dbm.KV) bool {
	value, err := getManageKey(key, db)
	if err != nil {
		clog.Error("isRightAddr", "Key", key)
		return false
	}
	if value == nil {
		clog.Error("isRightAddr", "key", key, "error", "Found key nil value")
		return false
	}

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		clog.Error("isRightAddr", "Decode", value)
		return false
	}

	for _, op := range item.GetArr().Value {
		if op == addr {
			return true
		}
	}
	return false

}

func getGuarantorAddr(db dbm.KV) (string, error) {
	value, err := getManageKey(guarantorKey, db)
	if err != nil {
		clog.Error("CollateralizePriceFeed", "getGuarantorAddr", err)
		return "", err
	}
	if value == nil {
		clog.Error("CollateralizePriceFeed guarantorKey found nil value")
		return "", err
	}

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		clog.Error("CollateralizePriceFeed", "getGuarantorAddr", err)
		return "", err
	}

	return item.GetAddr(), nil
}

// 系统清算
func (action *Action) systemLiquidation(coll *pty.Collateralize, price float32) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	collDB := &CollateralizeDB{*coll}
	for index, borrowRecord := range coll.BorrowRecords {
		if borrowRecord.LiquidationPrice * PriceWarningRate < price {
			if borrowRecord.Status == pty.CollateralizeUserStatusSystemLiquidate {
				borrowRecord.Status = borrowRecord.PreStatus
				borrowRecord.PreStatus = pty.CollateralizeUserStatusSystemLiquidate
			}
			continue
		}

		if borrowRecord.LiquidationPrice >= price {
			getGuarantorAddr, err := getGuarantorAddr(action.db)
			if err != nil {
				if err != nil {
					clog.Error("systemLiquidation", "getGuarantorAddr", err)
					continue
				}
			}

			// 抵押物转移
			receipt, err := action.coinsAccount.ExecTransferFrozen(action.fromaddr, getGuarantorAddr, action.execaddr, borrowRecord.CollateralValue)
			if err != nil {
				clog.Error("systemLiquidation", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", borrowRecord.CollateralValue, "err", err)
				continue
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)

			// 借贷记录清算
			borrowRecord.LiquidateTime = action.blocktime
			borrowRecord.PreStatus = borrowRecord.Status
			borrowRecord.Status = pty.CollateralizeUserStatusSystemLiquidate
			coll.BorrowRecords = append(coll.BorrowRecords[:index], coll.BorrowRecords[index+1:]...)
			coll.InvalidRecords = append(coll.InvalidRecords, borrowRecord)
		} else {
			borrowRecord.PreStatus = borrowRecord.Status
			borrowRecord.Status = pty.CollateralizeUserStatusWarning
		}

		log := action.GetFeedReceiptLog(coll, borrowRecord)
		logs = append(logs, log)
	}

	// 保存
	coll.LatestLiquidationPrice = getLatestLiquidationPrice(coll)
	coll.LatestExpireTime = getLatestExpireTime(coll)
	collDB.Save(action.db)
	kv = append(kv, collDB.GetKVSet()...)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// 超时清算
func (action *Action) expireLiquidation(coll *pty.Collateralize) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	collDB := &CollateralizeDB{*coll}
	for index, borrowRecord := range coll.BorrowRecords {
		if borrowRecord.ExpireTime - ExpireWarningTime > action.blocktime {
			continue
		}

		if borrowRecord.ExpireTime >= action.blocktime {
			getGuarantorAddr, err := getGuarantorAddr(action.db)
			if err != nil {
				if err != nil {
					clog.Error("systemLiquidation", "getGuarantorAddr", err)
					continue
				}
			}

			// 抵押物转移
			receipt, err := action.coinsAccount.ExecTransferFrozen(action.fromaddr, getGuarantorAddr, action.execaddr, borrowRecord.CollateralValue)
			if err != nil {
				clog.Error("systemLiquidation", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", borrowRecord.CollateralValue, "err", err)
				continue
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)

			// 借贷记录清算
			borrowRecord.LiquidateTime = action.blocktime
			borrowRecord.PreStatus = borrowRecord.Status
			borrowRecord.Status = pty.CollateralizeUserStatusExpireLiquidate
			coll.BorrowRecords = append(coll.BorrowRecords[:index], coll.BorrowRecords[index+1:]...)
			coll.InvalidRecords = append(coll.InvalidRecords, borrowRecord)
		} else {
			borrowRecord.PreStatus = borrowRecord.Status
			borrowRecord.Status = pty.CollateralizeUserStatusExpire
		}

		log := action.GetFeedReceiptLog(coll, borrowRecord)
		logs = append(logs, log)
	}

	// 保存
	coll.LatestLiquidationPrice = getLatestLiquidationPrice(coll)
	coll.LatestExpireTime = getLatestExpireTime(coll)
	collDB.Save(action.db)
	kv = append(kv, collDB.GetKVSet()...)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// 价格计算策略
func pricePolicy(feed *pty.CollateralizeFeed) float32 {
	var totalPrice float32
	var totalVolume int64
	for _, volume := range feed.Volume {
		totalVolume += volume
	}

	for i, price := range feed.Price {
		totalPrice += price * float32(float64(feed.Volume[i])/float64(totalVolume))
	}

	return totalPrice
}

// CollateralizeFeed 喂价
func (action *Action) CollateralizeFeed(feed *pty.CollateralizeFeed) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if feed == nil || len(feed.Price) == 0 || len(feed.Price) != len(feed.Volume) {
		clog.Error("CollateralizePriceFeed", types.ErrInvalidParam)
		return nil, types.ErrInvalidParam
	}

	// 是否后台管理用户
	if !isRightAddr(priceFeedKey, action.fromaddr, action.db) {
		clog.Error("CollateralizePriceFeed", "addr", action.fromaddr, "error", "Address has no permission to feed price")
		return nil, pty.ErrPermissionDeny
	}

	price := pricePolicy(feed)
	if price == 0 || price == -1 {
		clog.Error("CollateralizePriceFeed", "price", price, "err", pty.ErrPriceInvalid)
		return nil, pty.ErrPriceInvalid
	}

	collIDRecords, err := queryCollateralizeByStatus(action.localDB, pty.CollateralizeStatusCreated)
	if err != nil {
		clog.Error("CollateralizePriceFeed", "get collateralize record error", err)
		return nil, err
	}

	for _, collID := range collIDRecords {
		coll, err := queryCollateralizeByID(action.db, collID.CollateralizeId)
		if err != nil {
			clog.Error("CollateralizePriceFeed", "Collateralize ID", coll.CollateralizeId, "get collateralize record by id error", err)
			continue
		}

		// 超时清算判断
		if coll.LatestExpireTime - ExpireWarningTime <= action.blocktime {
			receipt, err := action.expireLiquidation(coll)
			if err != nil {
				clog.Error("CollateralizePriceFeed", "Collateralize ID", coll.CollateralizeId, "expire liquidation error", err)
				continue
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}

		// 系统清算判断
		receipt, err := action.systemLiquidation(coll, price)
		if err != nil {
			clog.Error("CollateralizePriceFeed", "Collateralize ID", coll.CollateralizeId, "system liquidation error", err)
			continue
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	var priceRecord pty.AssetPriceRecord
	if feed.CollType == pty.CollateralizeAssetTypeBty {
		priceRecord.BtyPrice = price
	} else if feed.CollType == pty.CollateralizeAssetTypeBtc {
		priceRecord.BtcPrice = price
	} else if feed.CollType == pty.CollateralizeAssetTypeEth {
		priceRecord.EthPrice = price
	} else {
		priceRecord.BtyPrice = price
	}
	priceRecord.RecordTime = action.blocktime

	// 喂价记录
	pricekv := &types.KeyValue{Key: calcCollateralizePriceKey(string(action.blocktime)), Value: types.Encode(&priceRecord)}
	action.db.Set(pricekv.Key, pricekv.Value)
	kv = append(kv, pricekv)

	// 最近喂价记录
	pricekv = &types.KeyValue{Key: calcCollateralizeLatestPriceKey(), Value: types.Encode(&priceRecord)}
	action.db.Set(pricekv.Key, pricekv.Value)
	kv = append(kv, pricekv)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// CollateralizeClose 终止借贷
func (action *Action) CollateralizeClose(close *pty.CollateralizeClose) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	collateralize, err := queryCollateralizeByID(action.db, close.CollateralizeId)
	if err != nil {
		clog.Error("CollateralizeClose", "CollateralizeId", close.CollateralizeId, "err", err)
		return nil, err
	}

	for _, borrowRecord := range collateralize.BorrowRecords {
		if borrowRecord.Status != pty.CollateralizeUserStatusClose {
			clog.Error("CollateralizeClose", "CollateralizeId", close.CollateralizeId, "addr", action.fromaddr, "execaddr", action.execaddr, "err", pty.ErrCollateralizeRecordNotEmpty)
			return nil, pty.ErrCollateralizeRecordNotEmpty
		}
	}

	clog.Debug("CollateralizeClose", "ID", close.CollateralizeId)

	coll := &CollateralizeDB{*collateralize}
	preStatus := coll.Status
	coll.Status = pty.CollateralizeStatusClose
	coll.Save(action.db)
	kv = append(kv, coll.GetKVSet()...)

	receiptLog := action.GetCloseReceiptLog(&coll.Collateralize, preStatus)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

// 查找借贷
func queryCollateralizeByID(db dbm.KV, CollateralizeID string) (*pty.Collateralize, error) {
	data, err := db.Get(Key(CollateralizeID))
	if err != nil {
		clog.Debug("queryCollateralizeByID", "error", err)
		return nil, err
	}

	var coll pty.Collateralize
	err = types.Decode(data, &coll)
	if err != nil {
		clog.Debug("queryCollateralizeByID", "decode", err)
		return nil, err
	}
	return &coll, nil
}

func queryCollateralizeByStatus(localdb dbm.Lister, status int32) ([]*pty.CollateralizeRecord, error) {
	data, err := localdb.List(calcCollateralizeStatusPrefix(status), nil, DefultCount, ListDESC)
	if err != nil {
		clog.Debug("queryCollateralizesByStatus", "error", err)
		return nil, err
	}

	var colls []*pty.CollateralizeRecord
	var coll pty.CollateralizeRecord
	for _, collBytes := range data {
		err = types.Decode(collBytes, &coll)
		if err != nil {
			clog.Debug("queryCollateralizesByStatus", "decode", err)
			return nil, err
		}
		colls = append(colls, &coll)
	}

	return colls, nil
}

func queryCollateralizeByAddr(localdb dbm.Lister, addr string) ([]*pty.CollateralizeRecord, error) {
	data, err := localdb.List(calcCollateralizeAddrPrefix(addr), nil, DefultCount, ListDESC)
	if err != nil {
		clog.Debug("queryCollateralizesByAddr", "error", err)
		return nil, err
	}

	var colls []*pty.CollateralizeRecord
	var coll pty.CollateralizeRecord
	for _, collBytes := range data {
		err = types.Decode(collBytes, &coll)
		if err != nil {
			clog.Debug("queryCollateralizesByAddr", "decode", err)
			return nil, err
		}
		colls = append(colls, &coll)
	}

	return colls, nil
}

func queryCollateralizeRecordByStatus(localdb dbm.Lister, status int32) ([]*pty.CollateralizeRecord, error) {
	data, err := localdb.List(calcCollateralizeRecordStatusPrefix(status), nil, DefultCount, ListDESC)
	if err != nil {
		clog.Debug("queryCollateralizeRecordByStatus", "error", err)
		return nil, err
	}

	var colls []*pty.CollateralizeRecord
	var coll pty.CollateralizeRecord
	for _, collBytes := range data {
		err = types.Decode(collBytes, &coll)
		if err != nil {
			clog.Debug("queryCollateralizesByStatus", "decode", err)
			return nil, err
		}
		colls = append(colls, &coll)
	}

	return colls, nil
}
