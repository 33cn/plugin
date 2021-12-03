// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"math"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/issuance/types"
	tokenE "github.com/33cn/plugin/plugin/dapp/token/executor"
	"github.com/shopspring/decimal"
)

// List control
const (
	ListDESC     = int32(0)   // list降序
	ListASC      = int32(1)   // list升序
	DefaultCount = int32(20)  // 默认一次取多少条记录
	MaxCount     = int32(100) // 最多取100条
)

//setting
const (
	DefaultDebtCeiling      = 100000          // 默认借贷限额
	DefaultLiquidationRatio = 0.25 * 1e4      // 默认质押比
	DefaultPeriod           = 3600 * 24 * 365 // 默认合约限期
	PriceWarningRate        = 1.3 * 1e4       // 价格提前预警率
	ExpireWarningTime       = 3600 * 24 * 10  // 提前10天超时预警
)

func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		return nil, err
	}
	return value, nil
}

func getGuarantorAddr(db dbm.KV) (string, error) {
	value, err := getManageKey(pty.GuarantorKey, db)
	if err != nil {
		clog.Error("IssuancePriceFeed", "getGuarantorAddr", err)
		return "", err
	}
	if value == nil {
		clog.Error("IssuancePriceFeed guarantorKey found nil value")
		return "", err
	}

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		clog.Error("IssuancePriceFeed", "getGuarantorAddr", err)
		return "", err
	}

	return item.GetArr().Value[0], nil
}

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

func isSuperAddr(addr string, db dbm.KV) bool {
	data, err := db.Get(AddrKey())
	if err != nil {
		clog.Error("getSuperAddr", "error", err)
		return false
	}

	var item types.ConfigItem
	err = types.Decode(data, &item)
	if err != nil {
		clog.Error("isSuperAddr", "Decode", data)
		return false
	}

	for _, op := range item.GetArr().Value {
		if op == addr {
			return true
		}
	}

	return false
}

// IssuanceDB def
type IssuanceDB struct {
	pty.Issuance
}

// GetKVSet for IssuanceDB
func (issu *IssuanceDB) GetKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(&issu.Issuance)
	kvset = append(kvset, &types.KeyValue{Key: Key(issu.IssuanceId), Value: value})
	return kvset
}

// Save for IssuanceDB
func (issu *IssuanceDB) Save(db dbm.KV) {
	set := issu.GetKVSet()
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
}

// Key for Issuance
func Key(id string) (key []byte) {
	key = append(key, []byte("mavl-"+pty.IssuanceX+"-")...)
	key = append(key, []byte(id)...)
	return key
}

// AddrKey for IssuanceAddrConfig
func AddrKey() (key []byte) {
	key = append(key, []byte("mavl-"+pty.IssuanceX+"-addr")...)
	return key
}

// PriceKey for IssuancePriceFeed
func PriceKey() (key []byte) {
	key = append(key, []byte("mavl-"+pty.IssuanceX+"-price")...)
	return key
}

// Action struct
type Action struct {
	coinsAccount *account.DB // bty账户
	tokenAccount *account.DB // ccny账户
	db           dbm.KV
	localDB      dbm.KVDB
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
	difficulty   uint64
	index        int
	Issuance     *Issuance
}

// NewIssuanceAction generate New Action
func NewIssuanceAction(c *Issuance, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	cfg := c.GetAPI().GetConfig()
	tokenDb, err := account.NewAccountDB(cfg, tokenE.GetName(), pty.CCNYTokenName, c.GetStateDB())
	if err != nil {
		clog.Error("NewIssuanceAction", "Get Account DB error", "error", err)
		return nil
	}

	return &Action{
		coinsAccount: c.GetCoinsAccount(), tokenAccount: tokenDb, db: c.GetStateDB(), localDB: c.GetLocalDB(),
		txhash: hash, fromaddr: fromaddr, blocktime: c.GetBlockTime(), height: c.GetHeight(),
		execaddr: dapp.ExecAddress(string(tx.Execer)), difficulty: c.GetDifficulty(), index: index, Issuance: c}
}

// GetCreateReceiptLog generate logs for Issuance create action
func (action *Action) GetCreateReceiptLog(issuance *pty.Issuance) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogIssuanceCreate

	c := &pty.ReceiptIssuance{}
	c.IssuanceId = issuance.IssuanceId
	c.Status = issuance.Status

	log.Log = types.Encode(c)

	return log
}

// GetDebtReceiptLog generate logs for Issuance debt action
func (action *Action) GetDebtReceiptLog(issuance *pty.Issuance, debtRecord *pty.DebtRecord) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogIssuanceDebt

	c := &pty.ReceiptIssuance{}
	c.IssuanceId = issuance.IssuanceId
	c.AccountAddr = action.fromaddr
	c.DebtId = debtRecord.DebtId
	c.Status = debtRecord.Status

	log.Log = types.Encode(c)

	return log
}

// GetRepayReceiptLog generate logs for Issuance Repay action
func (action *Action) GetRepayReceiptLog(issuance *pty.Issuance, debtRecord *pty.DebtRecord) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogIssuanceRepay

	c := &pty.ReceiptIssuance{}
	c.IssuanceId = issuance.IssuanceId
	c.AccountAddr = action.fromaddr
	c.DebtId = debtRecord.DebtId
	c.Status = debtRecord.Status

	log.Log = types.Encode(c)

	return log
}

// GetFeedReceiptLog generate logs for Issuance price feed action
func (action *Action) GetFeedReceiptLog(issuance *pty.Issuance, debtRecord *pty.DebtRecord) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogIssuanceFeed

	c := &pty.ReceiptIssuance{}
	c.IssuanceId = issuance.IssuanceId
	c.AccountAddr = debtRecord.AccountAddr
	c.DebtId = debtRecord.DebtId
	c.Status = debtRecord.Status

	log.Log = types.Encode(c)

	return log
}

// GetCloseReceiptLog generate logs for Issuance close action
func (action *Action) GetCloseReceiptLog(issuance *pty.Issuance) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogIssuanceClose

	c := &pty.ReceiptIssuance{}
	c.IssuanceId = issuance.IssuanceId
	c.Status = issuance.Status

	log.Log = types.Encode(c)

	return log
}

// GetIndex returns index in block
func (action *Action) GetIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

func getLatestLiquidationPrice(issu *pty.Issuance) int64 {
	var latest int64
	for _, collRecord := range issu.DebtRecords {
		if collRecord.LiquidationPrice > latest {
			latest = collRecord.LiquidationPrice
		}
	}

	return latest
}

func getLatestExpireTime(issu *pty.Issuance) int64 {
	var latest int64 = 0x7fffffffffffffff

	for _, collRecord := range issu.DebtRecords {
		if collRecord.ExpireTime < latest {
			latest = collRecord.ExpireTime
		}
	}

	return latest
}

// IssuanceManage 设置全局借贷参数（管理员权限）
func (action *Action) IssuanceManage(manage *pty.IssuanceManage) (*types.Receipt, error) {
	var kv []*types.KeyValue
	var receipt *types.Receipt

	// 是否配置管理用户
	if !isRightAddr(pty.ManageKey, action.fromaddr, action.db) {
		clog.Error("IssuanceManage", "addr", action.fromaddr, "error", "Address has no permission to config")
		return nil, pty.ErrPermissionDeny
	}

	// 添加大户地址
	var item types.ConfigItem
	data, err := action.db.Get(AddrKey())
	if err != nil {
		if err != types.ErrNotFound {
			clog.Error("IssuanceManage", "error", err)
			return nil, err
		}
		emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
		arr := types.ConfigItem_Arr{Arr: emptyValue}
		item.Value = &arr
		item.GetArr().Value = append(item.GetArr().Value, manage.SuperAddrs...)

		value := types.Encode(&item)
		action.db.Set(AddrKey(), value)
		kv = append(kv, &types.KeyValue{Key: AddrKey(), Value: value})
	} else {
		err = types.Decode(data, &item)
		if err != nil {
			clog.Debug("IssuanceManage", "decode", err)
			return nil, err
		}
		item.GetArr().Value = append(item.GetArr().Value, manage.SuperAddrs...)
		value := types.Encode(&item)
		action.db.Set(AddrKey(), value)
		kv = append(kv, &types.KeyValue{Key: AddrKey(), Value: value})
	}

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: nil}
	return receipt, nil
}

func (action *Action) getSuperAddr() []string {
	data, err := action.db.Get(AddrKey())
	if err != nil {
		clog.Error("getSuperAddr", "error", err)
		return nil
	}

	var addrStore pty.IssuanceManage
	err = types.Decode(data, &addrStore)
	if err != nil {
		clog.Debug("getSuperAddr", "decode", err)
		return nil
	}

	return addrStore.SuperAddrs
}

// IssuanceCreate 创建借贷，持有一定数量ccny的用户可创建借贷，提供给其他用户借贷
func (action *Action) IssuanceCreate(create *pty.IssuanceCreate) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	cfg := action.Issuance.GetAPI().GetConfig()
	// 是否配置管理用户
	if !isRightAddr(pty.FundKey, action.fromaddr, action.db) {
		clog.Error("IssuanceCreate", "addr", action.fromaddr, "error", "Address has no permission to create")
		return nil, pty.ErrPermissionDeny
	}

	// 参数检查
	if create.GetTotalBalance() <= 0 {
		clog.Error("IssuanceCreate", "addr", action.fromaddr, "execaddr", action.execaddr, "total balance", create.GetTotalBalance(), "error", types.ErrAmount)
		return nil, types.ErrAmount
	}
	if create.DebtCeiling < 0 || create.LiquidationRatio < 0 || create.Period < 0 {
		clog.Error("IssuanceCreate", "addr", action.fromaddr, "execaddr", action.execaddr, "error", types.ErrInvalidParam)
		return nil, types.ErrInvalidParam
	}

	// 检查ccny余额
	if !action.CheckExecTokenAccount(action.fromaddr, create.TotalBalance, false) {
		return nil, types.ErrInsufficientBalance
	}

	// 查找ID是否重复
	issuanceID := common.ToHex(action.txhash)
	_, err := queryIssuanceByID(action.db, issuanceID)
	if err != types.ErrNotFound {
		clog.Error("IssuanceCreate", "IssuanceCreate repeated", issuanceID)
		return nil, pty.ErrIssuanceRepeatHash
	}

	// 冻结ccny
	receipt, err = action.tokenAccount.ExecFrozen(action.fromaddr, action.execaddr, create.TotalBalance)
	if err != nil {
		clog.Error("IssuanceCreate.Frozen", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", create.TotalBalance)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 构造coll结构
	issu := &IssuanceDB{}
	issu.IssuanceId = issuanceID
	issu.TotalBalance = create.TotalBalance
	if create.LiquidationRatio != 0 {
		issu.LiquidationRatio = create.LiquidationRatio
	} else {
		issu.LiquidationRatio = DefaultLiquidationRatio
	}
	if create.DebtCeiling != 0 {
		issu.DebtCeiling = create.DebtCeiling
	} else {
		issu.DebtCeiling = DefaultDebtCeiling * cfg.GetCoinPrecision()
	}
	if create.Period != 0 {
		issu.Period = create.Period
	} else {
		issu.Period = DefaultPeriod
	}
	issu.Balance = create.TotalBalance
	issu.CreateTime = action.blocktime
	issu.IssuerAddr = action.fromaddr
	issu.Status = pty.IssuanceActionCreate

	clog.Debug("IssuanceCreate created", "IssuanceID", issuanceID, "TotalBalance", issu.TotalBalance)

	// 保存
	issu.Save(action.db)
	kv = append(kv, issu.GetKVSet()...)

	receiptLog := action.GetCreateReceiptLog(&issu.Issuance)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// 根据最近抵押物价格计算需要冻结的BTY数量
func getBtyNumToFrozen(value int64, price int64, ratio int64) (int64, error) {
	if price == 0 {
		clog.Error("Bty price should greate to 0")
		return 0, pty.ErrPriceInvalid
	}

	btyValue := (value * 1e4) / (price * ratio)
	return btyValue * 1e4, nil
}

// 获取最近抵押物价格
func getLatestPrice(db dbm.KV) (int64, error) {
	data, err := db.Get(PriceKey())
	if err != nil {
		clog.Error("getLatestPrice", "get", err)
		return -1, err
	}
	var price pty.IssuanceAssetPriceRecord
	//decode
	err = types.Decode(data, &price)
	if err != nil {
		clog.Error("getLatestPrice", "decode", err)
		return -1, err
	}

	return price.BtyPrice, nil
}

// CheckExecAccountBalance 检查账户抵押物余额
func (action *Action) CheckExecAccountBalance(fromAddr string, ToFrozen, ToActive int64) bool {
	acc := action.coinsAccount.LoadExecAccount(fromAddr, action.execaddr)
	if acc.GetBalance() >= ToFrozen && acc.GetFrozen() >= ToActive {
		return true
	}
	return false
}

// CheckExecTokenAccount 检查账户token余额
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

// IssuanceDebt 大户质押bty借出ccny
func (action *Action) IssuanceDebt(debt *pty.IssuanceDebt) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if !isSuperAddr(action.fromaddr, action.db) {
		clog.Error("IssuanceDebt", "error", "IssuanceDebt need super address")
		return nil, pty.ErrPermissionDeny
	}

	// 查找对应的借贷ID
	issuance, err := queryIssuanceByID(action.db, debt.IssuanceId)
	if err != nil {
		clog.Error("IssuanceDebt", "IssuanceId", debt.IssuanceId, "error", err)
		return nil, err
	}

	// 状态检查
	if issuance.Status == pty.IssuanceStatusClose {
		clog.Error("IssuanceDebt", "CollID", issuance.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "status", issuance.Status, "error", pty.ErrIssuanceStatus)
		return nil, pty.ErrIssuanceStatus
	}

	issu := &IssuanceDB{*issuance}

	// 借贷金额检查
	if debt.GetValue() <= 0 {
		clog.Error("IssuanceDebt", "CollID", issu.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "debt value", debt.GetValue(), "error", types.ErrInvalidParam)
		return nil, types.ErrInvalidParam
	}

	// 借贷金额不超过个人限额
	userBalance, _ := queryIssuanceUserBalance(action.db, action.localDB, action.fromaddr)
	if debt.GetValue()+userBalance > issu.DebtCeiling {
		clog.Error("IssuanceDebt", "CollID", issu.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr,
			"debt value", debt.GetValue(), "current balance", userBalance, "error", pty.ErrIssuanceExceedDebtCeiling)
		return nil, pty.ErrIssuanceExceedDebtCeiling
	}

	// 借贷金额不超过当前可借贷金额
	if debt.GetValue() > issu.Balance {
		clog.Error("IssuanceDebt", "CollID", issu.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "debt value", debt.GetValue(), "error", pty.ErrIssuanceLowBalance)
		return nil, pty.ErrIssuanceLowBalance
	}
	clog.Debug("IssuanceDebt", "value", debt.GetValue())

	// 获取抵押物价格
	lastPrice, err := getLatestPrice(action.db)
	if err != nil {
		clog.Error("IssuanceDebt.getLatestPrice", "CollID", issu.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "error", err)
		return nil, err
	}

	// 精度转换 #1024
	// 先将token由token精度转成精度8
	valueReal := debt.GetValue()
	cfg := action.Issuance.GetAPI().GetConfig()
	if cfg.IsDappFork(action.Issuance.GetHeight(), pty.IssuanceX, pty.ForkIssuancePrecision) {
		precisionNum := int(math.Log10(float64(cfg.GetTokenPrecision())))
		valueReal = decimal.NewFromInt(valueReal).Shift(int32(-precisionNum)).Shift(8).IntPart()
	}
	// 根据价格和需要借贷的金额，计算需要质押的抵押物数量
	btyFrozen, err := getBtyNumToFrozen(valueReal, lastPrice, issu.LiquidationRatio)
	if err != nil {
		clog.Error("IssuanceDebt.getBtyNumToFrozen", "CollID", issu.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "error", err)
		return nil, err
	}
	// 再将bty由精度8转成coins精度
	if cfg.IsDappFork(action.Issuance.GetHeight(), pty.IssuanceX, pty.ForkIssuancePrecision) {
		precisionNum := int(math.Log10(float64(cfg.GetCoinPrecision())))
		btyFrozen = decimal.NewFromInt(btyFrozen).Shift(-8).Shift(int32(precisionNum)).IntPart()
	}

	// 检查抵押物账户余额
	if !action.CheckExecAccountBalance(action.fromaddr, btyFrozen, 0) {
		clog.Error("IssuanceDebt.CheckExecAccountBalance", "CollID", issu.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "btyFrozen", btyFrozen, "error", types.ErrNoBalance)
		return nil, types.ErrNoBalance
	}

	// 抵押物转账
	receipt, err := action.coinsAccount.ExecTransfer(action.fromaddr, issu.IssuerAddr, action.execaddr, btyFrozen)
	if err != nil {
		clog.Error("IssuanceDebt.ExecTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", btyFrozen)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 抵押物冻结
	receipt, err = action.coinsAccount.ExecFrozen(issu.IssuerAddr, action.execaddr, btyFrozen)
	if err != nil {
		clog.Error("IssuanceDebt.Frozen", "addr", issu.IssuerAddr, "execaddr", action.execaddr, "amount", btyFrozen)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 借出ccny
	receipt, err = action.tokenAccount.ExecTransferFrozen(issu.IssuerAddr, action.fromaddr, action.execaddr, debt.Value)
	if err != nil {
		clog.Error("IssuanceDebt.ExecTokenTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", debt.Value)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 构造借出记录
	debtRecord := &pty.DebtRecord{}
	debtRecord.AccountAddr = action.fromaddr
	debtRecord.DebtId = common.ToHex(action.txhash)
	debtRecord.IssuId = issu.IssuanceId
	debtRecord.CollateralValue = btyFrozen
	debtRecord.StartTime = action.blocktime
	debtRecord.CollateralPrice = lastPrice
	debtRecord.DebtValue = debt.Value
	debtRecord.LiquidationPrice = (issu.LiquidationRatio * lastPrice * pty.IssuancePreLiquidationRatio) / 1e8
	debtRecord.Status = pty.IssuanceUserStatusCreate
	debtRecord.ExpireTime = action.blocktime + issu.Period

	// 记录当前借贷的最高自动清算价格
	if issu.LatestLiquidationPrice < debtRecord.LiquidationPrice {
		issu.LatestLiquidationPrice = debtRecord.LiquidationPrice
	}

	// 保存
	issu.DebtRecords = append(issu.DebtRecords, debtRecord)
	issu.CollateralValue += btyFrozen
	issu.DebtValue += debt.Value
	issu.Balance -= debt.Value
	issu.LatestExpireTime = getLatestExpireTime(&issu.Issuance)
	issu.Save(action.db)
	kv = append(kv, issu.GetKVSet()...)

	receiptLog := action.GetDebtReceiptLog(&issu.Issuance, debtRecord)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// IssuanceRepay 用户主动清算
func (action *Action) IssuanceRepay(repay *pty.IssuanceRepay) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	// 找到相应的借贷
	issuance, err := queryIssuanceByID(action.db, repay.IssuanceId)
	if err != nil {
		clog.Error("IssuanceRepay", "CollID", repay.IssuanceId, "error", err)
		return nil, err
	}

	issu := &IssuanceDB{*issuance}

	// 状态检查
	if issu.Status != pty.IssuanceStatusCreated {
		clog.Error("IssuanceRepay", "CollID", repay.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "error", "status error", "Status", issu.Status)
		return nil, pty.ErrIssuanceStatus
	}

	// 查找借出记录
	var debtRecord *pty.DebtRecord
	var index int
	for i, record := range issu.DebtRecords {
		if record.DebtId == repay.DebtId {
			debtRecord = record
			index = i
			break
		}
	}

	if debtRecord == nil {
		clog.Error("IssuanceRepay", "CollID", repay.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "error", "Can not find debt record")
		return nil, pty.ErrRecordNotExist
	}

	// 检查
	if !action.CheckExecTokenAccount(action.fromaddr, debtRecord.DebtValue, false) {
		clog.Error("IssuanceRepay", "CollID", issu.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "amount", debtRecord.DebtValue, "error", types.ErrInsufficientBalance)
		return nil, types.ErrNoBalance
	}

	// ccny转移
	receipt, err = action.tokenAccount.ExecTransfer(action.fromaddr, issu.IssuerAddr, action.execaddr, debtRecord.DebtValue)
	if err != nil {
		clog.Error("IssuanceRepay.ExecTokenTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", debtRecord.DebtValue)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 冻结ccny
	receipt, err = action.tokenAccount.ExecFrozen(issu.IssuerAddr, action.execaddr, debtRecord.DebtValue)
	if err != nil {
		clog.Error("IssuanceCreate.Frozen", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", debtRecord.DebtValue)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 抵押物归还
	receipt, err = action.coinsAccount.ExecTransferFrozen(issu.IssuerAddr, action.fromaddr, action.execaddr, debtRecord.CollateralValue)
	if err != nil {
		clog.Error("IssuanceRepay.ExecTransferFrozen", "addr", issu.IssuerAddr, "execaddr", action.execaddr, "amount", debtRecord.CollateralValue)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 借贷记录关闭
	debtRecord.PreStatus = debtRecord.Status
	debtRecord.Status = pty.IssuanceUserStatusClose

	// 保存
	issu.Balance += debtRecord.DebtValue
	issu.CollateralValue -= debtRecord.CollateralValue
	issu.DebtValue -= debtRecord.DebtValue
	issu.DebtRecords = append(issu.DebtRecords[:index], issu.DebtRecords[index+1:]...)
	issu.InvalidRecords = append(issu.InvalidRecords, debtRecord)
	issu.LatestLiquidationPrice = getLatestLiquidationPrice(&issu.Issuance)
	issu.LatestExpireTime = getLatestExpireTime(&issu.Issuance)
	issu.Save(action.db)
	kv = append(kv, issu.GetKVSet()...)

	receiptLog := action.GetRepayReceiptLog(&issu.Issuance, debtRecord)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func removeLiquidateRecord(debtRecords []*pty.DebtRecord, remove pty.DebtRecord) []*pty.DebtRecord {
	for index, record := range debtRecords {
		if record.DebtId == remove.DebtId {
			debtRecords = append(debtRecords[:index], debtRecords[index+1:]...)
			break
		}
	}

	return debtRecords
}

// 系统清算
func (action *Action) systemLiquidation(issu *pty.Issuance, price int64) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var removeRecord []*pty.DebtRecord

	for _, debtRecord := range issu.DebtRecords {
		if (debtRecord.LiquidationPrice*PriceWarningRate)/1e4 < price {
			// 价格恢复，告警记录恢复
			if debtRecord.Status == pty.IssuanceUserStatusWarning {
				debtRecord.PreStatus = debtRecord.Status
				debtRecord.Status = pty.IssuanceUserStatusCreate
				log := action.GetFeedReceiptLog(issu, debtRecord)
				logs = append(logs, log)
			}
			continue
		}

		// 价格低于清算线，记录清算
		if debtRecord.LiquidationPrice >= price {
			// 价格低于清算线，记录清算
			clog.Debug("systemLiquidation", "issuance id", debtRecord.IssuId, "record id", debtRecord.DebtId, "account", debtRecord.AccountAddr, "price", price)

			getGuarantorAddr, err := getGuarantorAddr(action.db)
			if err != nil {
				if err != nil {
					clog.Error("systemLiquidation", "getGuarantorAddr", err)
					continue
				}
			}

			// 抵押物转移
			receipt, err := action.coinsAccount.ExecTransferFrozen(issu.IssuerAddr, getGuarantorAddr, action.execaddr, debtRecord.CollateralValue)
			if err != nil {
				clog.Error("systemLiquidation", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", debtRecord.CollateralValue, "error", err)
				continue
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)

			// 借贷记录清算
			debtRecord.LiquidateTime = action.blocktime
			debtRecord.PreStatus = debtRecord.Status
			debtRecord.Status = pty.IssuanceUserStatusSystemLiquidate
			issu.InvalidRecords = append(issu.InvalidRecords, debtRecord)
			removeRecord = append(removeRecord, debtRecord)

			log := action.GetFeedReceiptLog(issu, debtRecord)
			logs = append(logs, log)

			continue
		}

		// 价格高于清算线，且还不处于告警状态，记录告警
		if debtRecord.Status != pty.IssuanceUserStatusWarning {
			debtRecord.PreStatus = debtRecord.Status
			debtRecord.Status = pty.IssuanceUserStatusWarning

			log := action.GetFeedReceiptLog(issu, debtRecord)
			logs = append(logs, log)
		}
	}

	// 删除被清算的记录
	for _, record := range removeRecord {
		issu.DebtRecords = removeLiquidateRecord(issu.DebtRecords, *record)
	}

	// 保存
	issu.LatestLiquidationPrice = getLatestLiquidationPrice(issu)
	issu.LatestExpireTime = getLatestExpireTime(issu)
	collDB := &IssuanceDB{*issu}
	collDB.Save(action.db)
	kv = append(kv, collDB.GetKVSet()...)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// 超时清算
func (action *Action) expireLiquidation(issu *pty.Issuance) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var removeRecord []*pty.DebtRecord

	for _, debtRecord := range issu.DebtRecords {
		if debtRecord.ExpireTime-ExpireWarningTime > action.blocktime {
			continue
		}

		// 超过超时时间，记录清算
		if debtRecord.ExpireTime <= action.blocktime {
			// 超过清算线，记录清算
			clog.Debug("expireLiquidation", "issuance id", debtRecord.IssuId, "record id", debtRecord.DebtId, "account", debtRecord.AccountAddr, "time", action.blocktime)

			getGuarantorAddr, err := getGuarantorAddr(action.db)
			if err != nil {
				if err != nil {
					clog.Error("expireLiquidation", "getGuarantorAddr", err)
					continue
				}
			}

			// 抵押物转移
			receipt, err := action.coinsAccount.ExecTransferFrozen(issu.IssuerAddr, getGuarantorAddr, action.execaddr, debtRecord.CollateralValue)
			if err != nil {
				clog.Error("expireLiquidation", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", debtRecord.CollateralValue, "error", err)
				continue
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)

			// 借贷记录清算
			debtRecord.LiquidateTime = action.blocktime
			debtRecord.PreStatus = debtRecord.Status
			debtRecord.Status = pty.IssuanceUserStatusExpireLiquidate
			issu.InvalidRecords = append(issu.InvalidRecords, debtRecord)
			removeRecord = append(removeRecord, debtRecord)
			log := action.GetFeedReceiptLog(issu, debtRecord)
			logs = append(logs, log)

			continue
		}

		// 还没记录超时告警，记录告警
		if debtRecord.Status != pty.IssuanceUserStatusExpire {
			debtRecord.PreStatus = debtRecord.Status
			debtRecord.Status = pty.IssuanceUserStatusExpire
			log := action.GetFeedReceiptLog(issu, debtRecord)
			logs = append(logs, log)
		}
	}

	// 删除被清算的记录
	for _, record := range removeRecord {
		issu.DebtRecords = removeLiquidateRecord(issu.DebtRecords, *record)
	}

	// 保存
	issu.LatestLiquidationPrice = getLatestLiquidationPrice(issu)
	issu.LatestExpireTime = getLatestExpireTime(issu)
	collDB := &IssuanceDB{*issu}
	collDB.Save(action.db)
	kv = append(kv, collDB.GetKVSet()...)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// 价格计算策略
func pricePolicy(feed *pty.IssuanceFeed) int64 {
	var totalPrice int64
	var totalVolume int64
	for _, volume := range feed.Volume {
		totalVolume += volume
	}

	if totalVolume == 0 {
		clog.Error("issuance price feed volume empty")
		return 0
	}
	for i, price := range feed.Price {
		totalPrice += (price * feed.Volume[i]) / totalVolume
	}

	return totalPrice
}

// IssuanceFeed 喂价
func (action *Action) IssuanceFeed(feed *pty.IssuanceFeed) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if feed == nil || len(feed.Price) == 0 || len(feed.Price) != len(feed.Volume) {
		clog.Error("IssuancePriceFeed", types.ErrInvalidParam)
		return nil, types.ErrInvalidParam
	}

	// 是否后台管理用户
	if !isRightAddr(pty.PriceFeedKey, action.fromaddr, action.db) {
		clog.Error("IssuancePriceFeed", "addr", action.fromaddr, "error", "Address has no permission to feed price")
		return nil, pty.ErrPermissionDeny
	}

	price := pricePolicy(feed)
	if price <= 0 {
		clog.Error("IssuancePriceFeed", "price", price, "error", pty.ErrPriceInvalid)
		return nil, pty.ErrPriceInvalid
	}

	ids, err := queryIssuanceByStatus(action.localDB, pty.IssuanceStatusCreated, "")
	if err != nil {
		clog.Debug("IssuancePriceFeed", "get issuance record error", err)
	}

	for _, collID := range ids {
		issu, err := queryIssuanceByID(action.db, collID)
		if err != nil {
			clog.Error("IssuancePriceFeed", "Issuance ID", issu.IssuanceId, "get issuance record by id error", err)
			continue
		}

		// 超时清算判断
		if issu.LatestExpireTime-ExpireWarningTime <= action.blocktime {
			receipt, err := action.expireLiquidation(issu)
			if err != nil {
				clog.Error("IssuancePriceFeed", "Issuance ID", issu.IssuanceId, "expire liquidation error", err)
				continue
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}

		// 系统清算判断
		receipt, err := action.systemLiquidation(issu, price)
		if err != nil {
			clog.Error("IssuancePriceFeed", "Issuance ID", issu.IssuanceId, "system liquidation error", err)
			continue
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	var priceRecord pty.IssuanceAssetPriceRecord
	priceRecord.BtyPrice = price
	priceRecord.RecordTime = action.blocktime

	// 最近喂价记录
	pricekv := &types.KeyValue{Key: PriceKey(), Value: types.Encode(&priceRecord)}
	action.db.Set(pricekv.Key, pricekv.Value)
	kv = append(kv, pricekv)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// IssuanceClose 终止借贷
func (action *Action) IssuanceClose(close *pty.IssuanceClose) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	issuance, err := queryIssuanceByID(action.db, close.IssuanceId)
	if err != nil {
		clog.Error("IssuanceClose", "IssuanceId", close.IssuanceId, "error", err)
		return nil, err
	}

	if action.fromaddr != issuance.IssuerAddr {
		clog.Error("IssuanceClose", "IssuanceId", issuance.IssuanceId, "error", "account error", "create", issuance.IssuerAddr, "from", action.fromaddr)
		return nil, pty.ErrPermissionDeny
	}

	for _, debtRecord := range issuance.DebtRecords {
		if debtRecord.Status != pty.IssuanceUserStatusClose {
			clog.Error("IssuanceClose", "IssuanceId", close.IssuanceId, "addr", action.fromaddr, "execaddr", action.execaddr, "error", pty.ErrIssuanceRecordNotEmpty)
			return nil, pty.ErrIssuanceRecordNotEmpty
		}
	}

	// 解冻ccny
	if issuance.Balance > 0 {
		receipt, err = action.tokenAccount.ExecActive(action.fromaddr, action.execaddr, issuance.Balance)
		if err != nil {
			clog.Error("IssuanceClose.ExecActive", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", issuance.Balance, "error", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}
	clog.Debug("IssuanceClose", "ID", close.IssuanceId)

	issu := &IssuanceDB{*issuance}
	issu.Status = pty.IssuanceStatusClose

	issu.Save(action.db)
	kv = append(kv, issu.GetKVSet()...)

	receiptLog := action.GetCloseReceiptLog(&issu.Issuance)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

// 根据ID查找发行信息
func queryIssuanceByID(db dbm.KV, issuanceID string) (*pty.Issuance, error) {
	data, err := db.Get(Key(issuanceID))
	if err != nil {
		return nil, err
	}

	var issu pty.Issuance
	err = types.Decode(data, &issu)
	if err != nil {
		clog.Error("queryIssuanceByID", "decode", err)
		return nil, err
	}
	return &issu, nil
}

// 根据发行状态查找发行ID
func queryIssuanceByStatus(localdb dbm.KVDB, status int32, issuanceID string) ([]string, error) {
	query := pty.NewIssuanceTable(localdb).GetQuery(localdb)
	var primary []byte
	if len(issuanceID) > 0 {
		primary = []byte(issuanceID)
	}

	var data = &pty.ReceiptIssuanceID{
		IssuanceId: issuanceID,
		Status:     status,
	}
	rows, err := query.List("status", data, primary, DefaultCount, ListDESC)
	if err != nil {
		clog.Error("queryIssuanceByStatus.List", "error", err)
		return nil, err
	}

	var ids []string
	for _, row := range rows {
		ids = append(ids, string(row.Primary))
	}

	return ids, nil
}

// 精确查找发行记录
func queryIssuanceRecordByID(db dbm.KV, issuanceID string, debtID string) (*pty.DebtRecord, error) {
	issu, err := queryIssuanceByID(db, issuanceID)
	if err != nil {
		clog.Error("queryIssuanceRecordByID", "error", err)
		return nil, err
	}

	for _, record := range issu.DebtRecords {
		if record.DebtId == debtID {
			return record, nil
		}
	}

	for _, record := range issu.InvalidRecords {
		if record.DebtId == debtID {
			return record, nil
		}
	}

	return nil, types.ErrNotFound
}

// 根据发行状态查找
func queryIssuanceRecordsByStatus(db dbm.KV, localdb dbm.KVDB, status int32, debtID string) ([]*pty.DebtRecord, error) {
	query := pty.NewRecordTable(localdb).GetQuery(localdb)
	var primary []byte
	if len(debtID) > 0 {
		primary = []byte(debtID)
	}

	var data = &pty.ReceiptIssuance{
		Status: status,
	}
	rows, err := query.List("status", data, primary, DefaultCount, ListDESC)
	if err != nil {
		clog.Error("queryIssuanceRecordsByStatus.List", "index", "status", "error", err)
		return nil, err
	}

	var records []*pty.DebtRecord
	for _, row := range rows {
		record, err := queryIssuanceRecordByID(db, row.Data.(*pty.ReceiptIssuance).IssuanceId, row.Data.(*pty.ReceiptIssuance).DebtId)
		if err != nil {
			clog.Error("queryIssuanceRecordsByStatus", "decode", err)
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// 根据用户地址查找
func queryIssuanceRecordByAddr(db dbm.KV, localdb dbm.KVDB, addr string, status int32, debtID string) ([]*pty.DebtRecord, error) {
	query := pty.NewRecordTable(localdb).GetQuery(localdb)
	var primary []byte
	if len(debtID) > 0 {
		primary = []byte(debtID)
	}

	var data = &pty.ReceiptIssuance{
		AccountAddr: addr,
		Status:      status,
	}

	var rows []*table.Row
	var err error
	if status == 0 {
		rows, err = query.List("addr", data, primary, DefaultCount, ListDESC)
		if err != nil {
			clog.Error("queryIssuanceRecordByAddr.List", "index", "addr", "error", err)
			return nil, err
		}
	} else {
		rows, err = query.List("addr_status", data, primary, DefaultCount, ListDESC)
		if err != nil {
			clog.Error("queryIssuanceRecordByAddr.List", "index", "addr_status", "error", err)
			return nil, err
		}
	}

	var records []*pty.DebtRecord
	for _, row := range rows {
		record, err := queryIssuanceRecordByID(db, row.Data.(*pty.ReceiptIssuance).IssuanceId, row.Data.(*pty.ReceiptIssuance).DebtId)
		if err != nil {
			clog.Error("queryIssuanceRecordByAddr", "decode", err)
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

func queryIssuanceUserBalanceStatus(db dbm.KV, localdb dbm.KVDB, addr string, status int32) (int64, error) {
	var totalBalance int64
	query := pty.NewRecordTable(localdb).GetQuery(localdb)
	var primary []byte
	var data = &pty.ReceiptIssuance{
		AccountAddr: addr,
		Status:      status,
	}

	var rows []*table.Row
	var err error

	for {
		rows, err = query.List("addr_status", data, primary, DefaultCount, ListDESC)
		if err != nil {
			return -1, err
		}

		for _, row := range rows {
			record, err := queryIssuanceRecordByID(db, row.Data.(*pty.ReceiptIssuance).IssuanceId, row.Data.(*pty.ReceiptIssuance).DebtId)
			if err != nil {
				continue
			}
			totalBalance += record.DebtValue
		}

		if len(rows) < int(DefaultCount) {
			break
		}
		primary = []byte(rows[DefaultCount-1].Data.(*pty.ReceiptIssuance).DebtId)
	}

	return totalBalance, nil
}

func queryIssuanceUserBalance(db dbm.KV, localdb dbm.KVDB, addr string) (int64, error) {
	var totalBalance int64

	balance, err := queryIssuanceUserBalanceStatus(db, localdb, addr, pty.IssuanceUserStatusCreate)
	if err != nil {
		if err != types.ErrNotFound {
			clog.Error("queryIssuanceUserBalance", "err", err)
		}
	} else {
		totalBalance += balance
	}

	balance, err = queryIssuanceUserBalanceStatus(db, localdb, addr, pty.IssuanceUserStatusWarning)
	if err != nil {
		if err != types.ErrNotFound {
			clog.Error("queryIssuanceUserBalance", "err", err)
		}
	} else {
		totalBalance += balance
	}

	balance, err = queryIssuanceUserBalanceStatus(db, localdb, addr, pty.IssuanceUserStatusExpire)
	if err != nil {
		if err != types.ErrNotFound {
			clog.Error("queryIssuanceUserBalance", "err", err)
		}
	} else {
		totalBalance += balance
	}

	return totalBalance, nil
}
