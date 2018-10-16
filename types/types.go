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

func (at *UnfreezeType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogCreateUnfreeze:    {reflect.TypeOf(ReceiptUnfreeze{}), "LogCreateUnfreeze"},
		TyLogWithdrawUnfreeze:  {reflect.TypeOf(ReceiptUnfreeze{}), "LogWithdrawUnfreeze"},
		TyLogTerminateUnfreeze: {reflect.TypeOf(ReceiptUnfreeze{}), "LogTerminateUnfreeze"},
	}
}

func (g *UnfreezeType) GetPayload() types.Message {
	return &UnfreezeAction{}
}

func (g *UnfreezeType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create":    UnfreezeActionCreate,
		"Withdraw":  UnfreezeActionWithdraw,
		"Terminate": UnfreezeActionTerminate,
	}
}
