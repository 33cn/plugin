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
	types.RegistorExecutor(CollateralizeX, NewType())
	types.RegisterDappFork(CollateralizeX, "Enable", 0)
}

// CollateralizeType def
type CollateralizeType struct {
	types.ExecTypeBase
}

// NewType method
func NewType() *CollateralizeType {
	c := &CollateralizeType{}
	c.SetChild(c)
	return c
}

// GetName 获取执行器名称
func (Collateralize *CollateralizeType) GetName() string {
	return CollateralizeX
}

// GetLogMap method
func (Collateralize *CollateralizeType) GetLogMap() map[int64]*types.LogInfo {
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
func (Collateralize *CollateralizeType) GetPayload() types.Message {
	return &CollateralizeAction{}
}

// CreateTx method
func (Collateralize CollateralizeType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	llog.Debug("Collateralize.CreateTx", "action", action)

	if action == "CollateralizeCreate" {
		var param CollateralizeCreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeCreateTx(&param)
	} else if action == "CollateralizeBorrow" {
		var param CollateralizeBorrowTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeBorrowTx(&param)
	} else if action == "CollateralizeRepay" {
		var param CollateralizeRepayTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeRepayTx(&param)
	} else if action == "CollateralizeAppend" {
		var param CollateralizeAppendTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeAppendTx(&param)
	} else if action == "CollateralizeFeed" {
		var param CollateralizeFeedTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeFeedTx(&param)
	} else if action == "CollateralizeClose" {
		var param CollateralizeCloseTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawCollateralizeCloseTx(&param)
	} else {
		return nil, types.ErrNotSupport
	}
}

// GetTypeMap method
func (Collateralize CollateralizeType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create": CollateralizeActionCreate,
		"Borrow": CollateralizeActionBorrow,
		"Repay":  CollateralizeActionRepay,
		"Append": CollateralizeActionAppend,
		"Feed":   CollateralizeActionFeed,
		"Close":  CollateralizeActionClose,
	}
}

// CreateRawCollateralizeCreateTx method
func CreateRawCollateralizeCreateTx(parm *CollateralizeCreateTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeCreate{
		DebtCeiling:    parm.DebtCeiling,
		LiquidationRatio:   parm.LiquidationRatio,
		StabilityFee:  parm.StabilityFee,
		LiquidationPenalty: parm.LiquidationPenalty,
		TotalBalance:  parm.TotalBalance,
	}
	create := &CollateralizeAction{
		Ty:    CollateralizeActionCreate,
		Value: &CollateralizeAction_Create{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(CollateralizeX)),
		Payload: types.Encode(create),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(CollateralizeX)),
	}
	name := types.ExecName(CollateralizeX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeBorrowTx method
func CreateRawCollateralizeBorrowTx(parm *CollateralizeBorrowTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeBorrowTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeBorrow{
		CollateralizeId: parm.CollateralizeID,
		Value:    parm.Value,
	}
	Borrow := &CollateralizeAction{
		Ty:    CollateralizeActionBorrow,
		Value: &CollateralizeAction_Borrow{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(CollateralizeX)),
		Payload: types.Encode(Borrow),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(CollateralizeX)),
	}
	name := types.ExecName(CollateralizeX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeRepayTx method
func CreateRawCollateralizeRepayTx(parm *CollateralizeRepayTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeRepayTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeRepay{
		CollateralizeId: parm.CollateralizeID,
		Value: parm.Value,
	}
	Repay := &CollateralizeAction{
		Ty:    CollateralizeActionRepay,
		Value: &CollateralizeAction_Repay{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(CollateralizeX)),
		Payload: types.Encode(Repay),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(CollateralizeX)),
	}
	name := types.ExecName(CollateralizeX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeAppendTx method
func CreateRawCollateralizeAppendTx(parm *CollateralizeAppendTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizeAppendTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeAppend{
		CollateralizeId: parm.CollateralizeID,
		CollateralValue: parm.Value,
	}
	Repay := &CollateralizeAction{
		Ty:    CollateralizeActionAppend,
		Value: &CollateralizeAction_Append{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(CollateralizeX)),
		Payload: types.Encode(Repay),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(CollateralizeX)),
	}
	name := types.ExecName(CollateralizeX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeFeedTx method
func CreateRawCollateralizeFeedTx(parm *CollateralizeFeedTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawCollateralizePriceFeedTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &CollateralizeFeed{
		Price: parm.Price,
		Volume: parm.Volume,
	}
	Feed := &CollateralizeAction{
		Ty:    CollateralizeActionFeed,
		Value: &CollateralizeAction_Feed{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(CollateralizeX)),
		Payload: types.Encode(Feed),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(CollateralizeX)),
	}
	name := types.ExecName(CollateralizeX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawCollateralizeCloseTx method
func CreateRawCollateralizeCloseTx(parm *CollateralizeCloseTx) (*types.Transaction, error) {
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
		Execer:  []byte(types.ExecName(CollateralizeX)),
		Payload: types.Encode(close),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(CollateralizeX)),
	}

	name := types.ExecName(CollateralizeX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
