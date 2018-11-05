package types

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"time"

	log "github.com/inconshreveable/log15"
	"gitlab.33.cn/chain33/chain33/common/address"
	"gitlab.33.cn/chain33/chain33/types"
)

var name string

var tlog = log.New("module", name)

func init() {
	name = UnfreezeX
	// init executor type
	types.RegistorExecutor(name, &UnfreezeType{})
}

//getRealExecName
func getRealExecName(paraName string) string {
	return types.ExecName(paraName + UnfreezeX)
}

func NewType() *UnfreezeType {
	c := &UnfreezeType{}
	c.SetChild(c)
	return c
}

// exec
type UnfreezeType struct {
	types.ExecTypeBase
}

func (u *UnfreezeType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogCreateUnfreeze:    {reflect.TypeOf(ReceiptUnfreeze{}), "LogCreateUnfreeze"},
		TyLogWithdrawUnfreeze:  {reflect.TypeOf(ReceiptUnfreeze{}), "LogWithdrawUnfreeze"},
		TyLogTerminateUnfreeze: {reflect.TypeOf(ReceiptUnfreeze{}), "LogTerminateUnfreeze"},
	}
}

func (u *UnfreezeType) GetPayload() types.Message {
	return &UnfreezeAction{}
}

func (u *UnfreezeType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create":    UnfreezeActionCreate,
		"Withdraw":  UnfreezeActionWithdraw,
		"Terminate": UnfreezeActionTerminate,
	}
}

// TODO createTx接口暂时没法用，作为一个预留接口
func (u UnfreezeType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	tlog.Debug("UnfreezeType.CreateTx", "action", action)
	if action == Action_CreateUnfreeze {
		var param UnfreezeCreate
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateUnfreezeCreateTx(&param)
	} else if action == Action_WithdrawUnfreeze {
		var param UnfreezeWithdraw
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateUnfreezeWithdrawTx(&param)
	} else if action == Action_TerminateUnfreeze {
		var param UnfreezeTerminate
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateUnfreezeTerminateTx(&param)
	} else {
		return nil, types.ErrNotSupport
	}
	return nil, nil
}

func CreateUnfreezeCreateTx(parm *UnfreezeCreate) (*types.Transaction, error) {
	if parm == nil {
		tlog.Error("CreateUnfreezeCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	if parm.AssetExec == "" || parm.AssetSymbol == "" || parm.TotalCount <= 0 || parm.Means == "" {
		tlog.Error("CreateUnfreezeCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	if !supportMeans(parm.Means) {
		tlog.Error("CreateUnfreezeCreateTx not support means", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	create := &UnfreezeAction{
		Ty:    UnfreezeActionCreate,
		Value: &UnfreezeAction_Create{parm},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(types.GetParaName())),
		Payload: types.Encode(create),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	tx.SetRealFee(types.MinFee)
	return tx, nil
}

func CreateUnfreezeWithdrawTx(parm *UnfreezeWithdraw) (*types.Transaction, error) {
	if parm == nil {
		tlog.Error("CreateUnfreezeWithdrawTx", "parm", parm)
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
		Execer:  []byte(getRealExecName(types.GetParaName())),
		Payload: types.Encode(withdraw),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	tx.SetRealFee(types.MinFee)
	return tx, nil
}

func CreateUnfreezeTerminateTx(parm *UnfreezeTerminate) (*types.Transaction, error) {
	if parm == nil {
		tlog.Error("CreateUnfreezeTerminateTx", "parm", parm)
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
		Execer:  []byte(getRealExecName(types.GetParaName())),
		Payload: types.Encode(terminate),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	tx.SetRealFee(types.MinFee)
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
