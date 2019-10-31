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
	types.RegistorExecutor(IssuanceX, NewType())
	types.RegisterDappFork(IssuanceX, "Enable", 0)
}

// IssuanceType def
type IssuanceType struct {
	types.ExecTypeBase
}

// NewType method
func NewType() *IssuanceType {
	c := &IssuanceType{}
	c.SetChild(c)
	return c
}

// GetName 获取执行器名称
func (Issuance *IssuanceType) GetName() string {
	return IssuanceX
}

// GetLogMap method
func (Issuance *IssuanceType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogIssuanceCreate: {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceCreate"},
		TyLogIssuanceDebt:    {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceDebt"},
		TyLogIssuanceRepay:   {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceRepay"},
		TyLogIssuanceFeed:   {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceFeed"},
		TyLogIssuanceClose:  {Ty: reflect.TypeOf(ReceiptIssuance{}), Name: "LogIssuanceClose"},
	}
}

// GetPayload method
func (Issuance *IssuanceType) GetPayload() types.Message {
	return &IssuanceAction{}
}

// CreateTx method
func (Issuance IssuanceType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	llog.Debug("Issuance.CreateTx", "action", action)

	if action == "IssuanceCreate" {
		var param IssuanceCreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceCreateTx(&param)
	} else if action == "IssuanceDebt" {
		var param IssuanceDebtTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceDebtTx(&param)
	} else if action == "IssuanceRepay" {
		var param IssuanceRepayTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceRepayTx(&param)
	} else if action == "IssuancePriceFeed" {
		var param IssuanceFeedTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceFeedTx(&param)
	} else if action == "IssuanceClose" {
		var param IssuanceCloseTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceCloseTx(&param)
	} else if action == "IssuanceManage" {
		var param IssuanceManageTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawIssuanceManageTx(&param)
	} else {
		return nil, types.ErrNotSupport
	}
}

// GetTypeMap method
func (Issuance IssuanceType) GetTypeMap() map[string]int32 {
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
func CreateRawIssuanceCreateTx(parm *IssuanceCreateTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawIssuanceCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &IssuanceCreate{
		TotalBalance:  parm.TotalBalance,
	}
	create := &IssuanceAction{
		Ty:    IssuanceActionCreate,
		Value: &IssuanceAction_Create{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(IssuanceX)),
		Payload: types.Encode(create),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(IssuanceX)),
	}
	name := types.ExecName(IssuanceX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceDebtTx method
func CreateRawIssuanceDebtTx(parm *IssuanceDebtTx) (*types.Transaction, error) {
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
		Execer:  []byte(types.ExecName(IssuanceX)),
		Payload: types.Encode(debt),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(IssuanceX)),
	}
	name := types.ExecName(IssuanceX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceRepayTx method
func CreateRawIssuanceRepayTx(parm *IssuanceRepayTx) (*types.Transaction, error) {
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
		Execer:  []byte(types.ExecName(IssuanceX)),
		Payload: types.Encode(repay),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(IssuanceX)),
	}
	name := types.ExecName(IssuanceX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceFeedTx method
func CreateRawIssuanceFeedTx(parm *IssuanceFeedTx) (*types.Transaction, error) {
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
		Execer:  []byte(types.ExecName(IssuanceX)),
		Payload: types.Encode(feed),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(IssuanceX)),
	}
	name := types.ExecName(IssuanceX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceCloseTx method
func CreateRawIssuanceCloseTx(parm *IssuanceCloseTx) (*types.Transaction, error) {
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
		Execer:  []byte(types.ExecName(IssuanceX)),
		Payload: types.Encode(close),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(IssuanceX)),
	}

	name := types.ExecName(IssuanceX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawIssuanceManageTx method
func CreateRawIssuanceManageTx(parm *IssuanceManageTx) (*types.Transaction, error) {
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
		Execer:  []byte(types.ExecName(IssuanceX)),
		Payload: types.Encode(manage),
		Fee:     parm.Fee,
		To:      address.ExecAddress(types.ExecName(IssuanceX)),
	}

	name := types.ExecName(IssuanceX)
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}