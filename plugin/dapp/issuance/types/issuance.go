// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"reflect"

	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	llog = log.New("module", "exectype."+IssuanceX)
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(IssuanceX))
	types.RegFork(IssuanceX, InitFork)
	types.RegExec(IssuanceX, InitExecutor)
}

func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(IssuanceX, "Enable", 0)
}

func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(IssuanceX, NewType(cfg))
}

// IssuanceType def
type IssuanceType struct {
	types.ExecTypeBase
}

// NewType method
func NewType(cfg *types.Chain33Config) *IssuanceType {
	c := &IssuanceType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetName 获取执行器名称
func (issuance *IssuanceType) GetName() string {
	return IssuanceX
}

// GetLogMap method
func (issuance *IssuanceType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogIssuanceCreate: {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceCreate"},
		TyLogIssuanceDebt:    {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceDebt"},
		TyLogIssuanceRepay:   {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceRepay"},
		TyLogIssuanceFeed:   {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceFeed"},
		TyLogIssuanceClose:  {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceClose"},
	}
}

// GetPayload method
func (issuance *IssuanceType) GetPayload() types.Message {
	return &IssuanceAction{}
}

// CreateTx method
func (issuance IssuanceType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	llog.Debug("Issuance.CreateTx", "action", action)
	cfg := issuance.GetConfig()

	if action == "IssuanceCreate" {
		var param IssuanceCreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceCreateTx(cfg, &param)
	} else if action == "IssuanceDebt" {
		var param IssuanceDebtTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceDebtTx(cfg, &param)
	} else if action == "IssuanceRepay" {
		var param IssuanceRepayTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceRepayTx(cfg, &param)
	} else if action == "IssuancePriceFeed" {
		var param IssuanceFeedTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceFeedTx(cfg, &param)
	} else if action == "IssuanceClose" {
		var param IssuanceCloseTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceCloseTx(cfg, &param)
	} else if action == "IssuanceManage" {
		var param IssuanceManageTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceManageTx(cfg, &param)
	} else {
		return nil, types.ErrNotSupport
	}
}

// GetTypeMap method
func (issuance IssuanceType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create": IssuanceActionCreate,
		"Debt": IssuanceActionDebt,
		"Repay":  IssuanceActionRepay,
		"Feed":   IssuanceActionFeed,
		"Close":  IssuanceActionClose,
		"Manage": IssuanceActionManage,
	}
}

// CreateRawIssuanceCreateTx method
func CreateRawIssuanceCreateTx(cfg *types.Chain33Config, parm *IssuanceCreateTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawIssuanceCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &IssuanceCreate{
		TotalBalance:  parm.TotalBalance,
		DebtCeiling:   parm.DebtCeiling,
		LiquidationRatio: parm.LiquidationRatio,
		Period: parm.Period,
	}
	create := &IssuanceAction{
		Ty:    IssuanceActionCreate,
		Value: &IssuanceAction_Create{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(IssuanceX)),
		Payload: types.Encode(create),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(IssuanceX)),
	}
	name := cfg.ExecName(IssuanceX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceDebtTx method
func CreateRawIssuanceDebtTx(cfg *types.Chain33Config, parm *IssuanceDebtTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawIssuanceBorrowTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &IssuanceDebt{
		IssuanceId: parm.IssuanceID,
		Value:    parm.Value,
	}
	debt := &IssuanceAction{
		Ty:    IssuanceActionDebt,
		Value: &IssuanceAction_Debt{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(IssuanceX)),
		Payload: types.Encode(debt),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(IssuanceX)),
	}
	name := cfg.ExecName(IssuanceX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceRepayTx method
func CreateRawIssuanceRepayTx(cfg *types.Chain33Config, parm *IssuanceRepayTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawIssuanceRepayTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &IssuanceRepay{
		IssuanceId: parm.IssuanceID,
		DebtId: parm.DebtID,
	}
	repay := &IssuanceAction{
		Ty:    IssuanceActionRepay,
		Value: &IssuanceAction_Repay{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(IssuanceX)),
		Payload: types.Encode(repay),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(IssuanceX)),
	}
	name := cfg.ExecName(IssuanceX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceFeedTx method
func CreateRawIssuanceFeedTx(cfg *types.Chain33Config, parm *IssuanceFeedTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawIssuancePriceFeedTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &IssuanceFeed{
		Price: parm.Price,
		Volume: parm.Volume,
	}
	feed := &IssuanceAction{
		Ty:    IssuanceActionFeed,
		Value: &IssuanceAction_Feed{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(IssuanceX)),
		Payload: types.Encode(feed),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(IssuanceX)),
	}
	name := cfg.ExecName(IssuanceX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceCloseTx method
func CreateRawIssuanceCloseTx(cfg *types.Chain33Config, parm *IssuanceCloseTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawIssuanceCloseTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &IssuanceClose{
		IssuanceId: parm.IssuanceID,
	}
	close := &IssuanceAction{
		Ty:    IssuanceActionClose,
		Value: &IssuanceAction_Close{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(IssuanceX)),
		Payload: types.Encode(close),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(IssuanceX)),
	}

	name := cfg.ExecName(IssuanceX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceManageTx method
func CreateRawIssuanceManageTx(cfg *types.Chain33Config, parm *IssuanceManageTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawIssuanceManageTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &IssuanceManage{SuperAddrs:parm.Addr}

	manage := &IssuanceAction{
		Ty:    IssuanceActionManage,
		Value: &IssuanceAction_Manage{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(IssuanceX)),
		Payload: types.Encode(manage),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(IssuanceX)),
	}

	name := cfg.ExecName(IssuanceX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}