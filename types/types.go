package types

import (
	"reflect"

	log "github.com/inconshreveable/log15"
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

////创建解冻相关交易
//func (u UnfreezeType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
//	tlog.Debug("UnfreezeType.CreateTx", "action", action)
//	if action == Action_CreateUnfreeze {
//		var param UnfreezeCreate
//		err := json.Unmarshal(message, &param)
//		if err != nil {
//			tlog.Error("CreateTx", "Error", err)
//			return nil, types.ErrInputPara
//		}
//		return CreateRawGamePreCreateTx(&param)
//	} else if action == Action_WithdrawUnfreeze {
//		var param UnfreezeWithdraw
//		err := json.Unmarshal(message, &param)
//		if err != nil {
//			tlog.Error("CreateTx", "Error", err)
//			return nil, types.ErrInputPara
//		}
//		return CreateRawGamePreMatchTx(&param)
//	} else if action == Action_TerminateUnfreeze {
//		var param UnfreezeTerminate
//		err := json.Unmarshal(message, &param)
//		if err != nil {
//			tlog.Error("CreateTx", "Error", err)
//			return nil, types.ErrInputPara
//		}
//		return CreateRawGamePreCancelTx(&param)
//	} else {
//		return nil, types.ErrNotSupport
//	}
//	return nil, nil
//}
