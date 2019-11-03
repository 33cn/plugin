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
	llog = log.New("module", "exectype."+CollateralizeX)
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(CollateralizeX))
	types.RegFork(CollateralizeX, InitFork)
	types.RegExec(CollateralizeX, InitExecutor)
}

func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(CollateralizeX, "Enable", 0)
}

func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(CollateralizeX, NewType(cfg))
}

// CollateralizeType def
type CollateralizeType struct {
	types.ExecTypeBase
}

// NewType method
func NewType(cfg *types.Chain33Config) *CollateralizeType {
	c := &CollateralizeType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetName 获取执行器名称
func (collateralize *CollateralizeType) GetName() string {
	return CollateralizeX
}

// GetLogMap method
func (collateralize *CollateralizeType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogCollateralizeCreate: {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeCreate"},
		TyLogCollateralizeBorrow:    {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeBorrow"},
		TyLogCollateralizeRepay:   {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeRepay"},
		TyLogCollateralizeAppend:   {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeAppend"},
		TyLogCollateralizeFeed:   {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeFeed"},
		TyLogCollateralizeClose:  {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeClose"},
	}
}

// GetPayload method
func (collateralize *CollateralizeType) GetPayload() types.Message {
	return &CollateralizeAction{}
}

// CreateTx method
func (collateralize CollateralizeType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	llog.Debug("Collateralize.CreateTx", "action", action)
	cfg := collateralize.GetConfig()

	if action == "CollateralizeCreate" {
		var param CollateralizeCreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeCreateTx(cfg, &param)
	} else if action == "CollateralizeBorrow" {
		var param CollateralizeBorrowTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeBorrowTx(cfg, &param)
	} else if action == "CollateralizeRepay" {
		var param CollateralizeRepayTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeRepayTx(cfg, &param)
	} else if action == "CollateralizeAppend" {
		var param CollateralizeAppendTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeAppendTx(cfg, &param)
	} else if action == "CollateralizePriceFeed" {
		var param CollateralizeFeedTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeFeedTx(cfg, &param)
	} else if action == "CollateralizeClose" {
		var param CollateralizeCloseTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeCloseTx(cfg, &param)
	} else if action == "CollateralizeManage" {
		var param CollateralizeManageTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeManageTx(cfg, &param)
	} else {
		return nil, types.ErrNotSupport
	}
}

// GetTypeMap method
func (collateralize CollateralizeType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create": CollateralizeActionCreate,
		"Borrow": CollateralizeActionBorrow,
		"Repay":  CollateralizeActionRepay,
		"Append": CollateralizeActionAppend,
		"Feed":   CollateralizeActionFeed,
		"Close":  CollateralizeActionClose,
		"Manage": CollateralizeActionManage,
	}
}

// CreateRawCollateralizeCreateTx method
func CreateRawCollateralizeCreateTx(cfg *types.Chain33Config, parm *CollateralizeCreateTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeCreate{
		TotalBalance:  parm.TotalBalance,
	}
	create := &CollateralizeAction{
		Ty:    CollateralizeActionCreate,
		Value: &CollateralizeAction_Create{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(CollateralizeX)),
		Payload: types.Encode(create),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(CollateralizeX)),
	}
	name := cfg.ExecName(CollateralizeX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeBorrowTx method
func CreateRawCollateralizeBorrowTx(cfg *types.Chain33Config, parm *CollateralizeBorrowTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeBorrowTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeBorrow{
		CollateralizeId: parm.CollateralizeID,
		Value:    parm.Value,
	}
	borrow := &CollateralizeAction{
		Ty:    CollateralizeActionBorrow,
		Value: &CollateralizeAction_Borrow{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(CollateralizeX)),
		Payload: types.Encode(borrow),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(CollateralizeX)),
	}
	name := cfg.ExecName(CollateralizeX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeRepayTx method
func CreateRawCollateralizeRepayTx(cfg *types.Chain33Config, parm *CollateralizeRepayTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeRepayTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeRepay{
		CollateralizeId: parm.CollateralizeID,
		RecordId:parm.RecordID,
	}
	repay := &CollateralizeAction{
		Ty:    CollateralizeActionRepay,
		Value: &CollateralizeAction_Repay{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(CollateralizeX)),
		Payload: types.Encode(repay),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(CollateralizeX)),
	}
	name := cfg.ExecName(CollateralizeX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeAppendTx method
func CreateRawCollateralizeAppendTx(cfg *types.Chain33Config, parm *CollateralizeAppendTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeAppendTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeAppend{
		CollateralizeId: parm.CollateralizeID,
		RecordId:parm.RecordID,
		CollateralValue: parm.Value,
	}
	append := &CollateralizeAction{
		Ty:    CollateralizeActionAppend,
		Value: &CollateralizeAction_Append{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(CollateralizeX)),
		Payload: types.Encode(append),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(CollateralizeX)),
	}
	name := cfg.ExecName(CollateralizeX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeFeedTx method
func CreateRawCollateralizeFeedTx(cfg *types.Chain33Config, parm *CollateralizeFeedTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizePriceFeedTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeFeed{
		Price: parm.Price,
		Volume: parm.Volume,
	}
	feed := &CollateralizeAction{
		Ty:    CollateralizeActionFeed,
		Value: &CollateralizeAction_Feed{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(CollateralizeX)),
		Payload: types.Encode(feed),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(CollateralizeX)),
	}
	name := cfg.ExecName(CollateralizeX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeCloseTx method
func CreateRawCollateralizeCloseTx(cfg *types.Chain33Config, parm *CollateralizeCloseTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeCloseTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeClose{
		CollateralizeId: parm.CollateralizeID,
	}
	close := &CollateralizeAction{
		Ty:    CollateralizeActionClose,
		Value: &CollateralizeAction_Close{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(CollateralizeX)),
		Payload: types.Encode(close),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(CollateralizeX)),
	}

	name := cfg.ExecName(CollateralizeX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeManageTx method
func CreateRawCollateralizeManageTx(cfg *types.Chain33Config, parm *CollateralizeManageTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeManageTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeManage{
		DebtCeiling:          parm.DebtCeiling,
		LiquidationRatio:     parm.LiquidationRatio,
		StabilityFeeRatio:    parm.StabilityFeeRatio,
		Period:               parm.Period,
	}

	manage := &CollateralizeAction{
		Ty:    CollateralizeActionManage,
		Value: &CollateralizeAction_Manage{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(cfg.ExecName(CollateralizeX)),
		Payload: types.Encode(manage),
		Fee:     parm.Fee,
		To:      address.ExecAddress(cfg.ExecName(CollateralizeX)),
	}

	name := cfg.ExecName(CollateralizeX)
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
