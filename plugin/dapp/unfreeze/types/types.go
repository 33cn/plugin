// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"time"

	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var name string

var tlog = log.New("module", name)

func init() {
	name = UnfreezeX
	types.AllowUserExec = append(types.AllowUserExec, []byte(UnfreezeX))
	// init executor type
	types.RegistorExecutor(name, NewType())
	types.RegisterDappFork(name, "Enable", 0)
	types.RegisterDappFork(name, ForkTerminatePartX, 1298600)
	types.RegisterDappFork(name, ForkUnfreezeIDX, 1450000)
}

//getRealExecName
func getRealExecName(paraName string) string {
	return types.ExecName(paraName + UnfreezeX)
}

// NewType 生成新的基础类型
func NewType() *UnfreezeType {
	c := &UnfreezeType{}
	c.SetChild(c)
	return c
}

// UnfreezeType 基础类型结构体
type UnfreezeType struct {
	types.ExecTypeBase
}

// GetName 获取执行器名称
func (u *UnfreezeType) GetName() string {
	return UnfreezeX
}

// GetLogMap 获得日志类型列表
func (u *UnfreezeType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogCreateUnfreeze:    {Ty: reflect.TypeOf(ReceiptUnfreeze{}), Name: "LogCreateUnfreeze"},
		TyLogWithdrawUnfreeze:  {Ty: reflect.TypeOf(ReceiptUnfreeze{}), Name: "LogWithdrawUnfreeze"},
		TyLogTerminateUnfreeze: {Ty: reflect.TypeOf(ReceiptUnfreeze{}), Name: "LogTerminateUnfreeze"},
	}
}

// GetPayload 获得空的Unfreeze 的 Payload
func (u *UnfreezeType) GetPayload() types.Message {
	return &UnfreezeAction{}
}

// GetTypeMap 获得Action 方法列表
func (u *UnfreezeType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create":    UnfreezeActionCreate,
		"Withdraw":  UnfreezeActionWithdraw,
		"Terminate": UnfreezeActionTerminate,
	}
}

// CreateTx 创建交易
func (u UnfreezeType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	tlog.Debug("UnfreezeType.CreateTx", "action", action)
	if action == Action_CreateUnfreeze {
		var param UnfreezeCreate
		err := types.JSONToPB(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return u.RPC_UnfreezeCreateTx(&param)
	} else if action == Action_WithdrawUnfreeze {
		var param UnfreezeWithdraw
		err := types.JSONToPB(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return u.RPC_UnfreezeWithdrawTx(&param)
	} else if action == Action_TerminateUnfreeze {
		var param UnfreezeTerminate
		err := types.JSONToPB(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return u.RPC_UnfreezeTerminateTx(&param)
	}

	return nil, types.ErrNotSupport
}

// RPC_UnfreezeCreateTx 创建冻结合约交易入口
func (u UnfreezeType) RPC_UnfreezeCreateTx(parm *UnfreezeCreate) (*types.Transaction, error) {
	return CreateUnfreezeCreateTx(types.GetParaName(), parm)
}

// CreateUnfreezeCreateTx 创建冻结合约交易
func CreateUnfreezeCreateTx(title string, parm *UnfreezeCreate) (*types.Transaction, error) {
	if parm == nil {
		tlog.Error("RPC_UnfreezeCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	if parm.AssetExec == "" || parm.AssetSymbol == "" || parm.TotalCount <= 0 || parm.Means == "" {
		tlog.Error("RPC_UnfreezeCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	if !supportMeans(parm.Means) {
		tlog.Error("RPC_UnfreezeCreateTx not support means", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	create := &UnfreezeAction{
		Ty:    UnfreezeActionCreate,
		Value: &UnfreezeAction_Create{parm},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(title)),
		Payload: types.Encode(create),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	tx.SetRealFee(types.GInt("MinFee"))
	return tx, nil
}

// RPC_UnfreezeWithdrawTx 创建提币交易入口
func (u UnfreezeType) RPC_UnfreezeWithdrawTx(parm *UnfreezeWithdraw) (*types.Transaction, error) {
	return CreateUnfreezeWithdrawTx(types.GetParaName(), parm)
}

// CreateUnfreezeWithdrawTx 创建提币交易
func CreateUnfreezeWithdrawTx(title string, parm *UnfreezeWithdraw) (*types.Transaction, error) {
	if parm == nil {
		tlog.Error("RPC_UnfreezeWithdrawTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	v := &UnfreezeWithdraw{
		UnfreezeID: parm.UnfreezeID,
	}
	withdraw := &UnfreezeAction{
		Ty:    UnfreezeActionWithdraw,
		Value: &UnfreezeAction_Withdraw{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(title)),
		Payload: types.Encode(withdraw),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	tx.SetRealFee(types.GInt("MinFee"))
	return tx, nil
}

// RPC_UnfreezeTerminateTx 创建终止冻结合约入口
func (u UnfreezeType) RPC_UnfreezeTerminateTx(parm *UnfreezeTerminate) (*types.Transaction, error) {
	return CreateUnfreezeTerminateTx(types.GetParaName(), parm)
}

// CreateUnfreezeTerminateTx 创建终止冻结合约
func CreateUnfreezeTerminateTx(title string, parm *UnfreezeTerminate) (*types.Transaction, error) {
	if parm == nil {
		tlog.Error("RPC_UnfreezeTerminateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	v := &UnfreezeTerminate{
		UnfreezeID: parm.UnfreezeID,
	}
	terminate := &UnfreezeAction{
		Ty:    UnfreezeActionTerminate,
		Value: &UnfreezeAction_Terminate{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(title)),
		Payload: types.Encode(terminate),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	tx.SetRealFee(types.GInt("MinFee"))
	return tx, nil
}

func supportMeans(means string) bool {
	for _, m := range SupportMeans {
		if m == means {
			return true
		}
	}
	return false
}
