// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"math"
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

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(CollateralizeX, "Enable", 0)
	cfg.RegisterDappFork(CollateralizeX, ForkCollateralizeTableUpdate, 0)
}

//InitExecutor ...
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
		TyLogCollateralizeCreate:   {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeCreate"},
		TyLogCollateralizeBorrow:   {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeBorrow"},
		TyLogCollateralizeRepay:    {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeRepay"},
		TyLogCollateralizeAppend:   {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeAppend"},
		TyLogCollateralizeFeed:     {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeFeed"},
		TyLogCollateralizeRetrieve: {Ty: reflect.TypeOf(ReceiptCollateralize{}), Name: "LogCollateralizeRetrieve"},
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
	} else if action == "CollateralizeRetrieve" {
		var param CollateralizeRetrieveTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeRetrieveTx(cfg, &param)
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
		"Create":   CollateralizeActionCreate,
		"Borrow":   CollateralizeActionBorrow,
		"Repay":    CollateralizeActionRepay,
		"Append":   CollateralizeActionAppend,
		"Feed":     CollateralizeActionFeed,
		"Retrieve": CollateralizeActionRetrieve,
		"Manage":   CollateralizeActionManage,
	}
}

// CreateRawCollateralizeCreateTx method
func CreateRawCollateralizeCreateTx(cfg *types.Chain33Config, parm *CollateralizeCreateTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeCreate{
		TotalBalance: int64(math.Trunc((parm.TotalBalance+0.0000001)*1e4)) * 1e4,
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
		Value:           int64(math.Trunc((parm.Value+0.0000001)*1e4)) * 1e4,
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
		RecordId:        parm.RecordID,
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
		RecordId:        parm.RecordID,
		CollateralValue: int64(math.Trunc((parm.Value+0.0000001)*1e4)) * 1e4,
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
		Volume: parm.Volume,
	}

	for _, r := range parm.Price {
		v.Price = append(v.Price, int64(math.Trunc(r*1e4)))
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

// CreateRawCollateralizeRetrieveTx method
func CreateRawCollateralizeRetrieveTx(cfg *types.Chain33Config, parm *CollateralizeRetrieveTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeCloseTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeRetrieve{
		CollateralizeId: parm.CollateralizeID,
		Balance:         int64(math.Trunc((parm.Balance+0.0000001)*1e4)) * 1e4,
	}
	close := &CollateralizeAction{
		Ty:    CollateralizeActionRetrieve,
		Value: &CollateralizeAction_Retrieve{v},
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
		DebtCeiling:       int64(math.Trunc((parm.DebtCeiling+0.0000001)*1e4)) * 1e4,
		LiquidationRatio:  int64(math.Trunc((parm.LiquidationRatio + 0.0000001) * 1e4)),
		StabilityFeeRatio: int64(math.Trunc((parm.StabilityFeeRatio + 0.0000001) * 1e4)),
		Period:            parm.Period,
		TotalBalance:      int64(math.Trunc((parm.TotalBalance+0.0000001)*1e4)) * 1e4,
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
