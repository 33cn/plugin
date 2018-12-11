package types

import (
	"github.com/33cn/chain33/types"
	"reflect"
)

// action for executor
const (
	F3dActionStart = iota + 1
	F3dActionDraw
	F3dActionBuy
)

const (
	TyLogf3dUnknown = iota + 100
	TyLogf3dStart
	TyLogf3dDraw
	TyLogf3dBuy
)

// query func name
const (
	FuncNameQueryLastRoundInfo           = "QueryLastRoundInfo"
	FuncNameQueryRoundInfoByRound        = "QueryRoundInfoByRound"
	FuncNameQueryKeyCountByRoundAndAddr  = "QueryKeyCountByRoundAndAddr"
	FuncNameQueryBuyRecordByRoundAndAddr = "QueryBuyRecordByRoundAndAddr"
)

var (
	logMap = map[string]int32{
		"Start": F3dActionStart,
		"Draw":  F3dActionDraw,
		"Buy":   F3dActionBuy,
	}

	typeMap   = map[int64]*types.LogInfo{}
	F3DX      = "f3d"
	ExecerF3D = []byte(F3DX)
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(F3DX))
	types.RegistorExecutor(F3DX, NewType())
	types.RegisterDappFork(F3DX, "Enable", 0)
}

type f3dType struct {
	types.ExecTypeBase
}

func NewType() *f3dType {
	c := &f3dType{}
	c.SetChild(c)
	return c
}

func (t *f3dType) GetPayload() types.Message {
	return &F3DAction{}
}

func (t *f3dType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Start": F3dActionStart,
		"Draw":  F3dActionDraw,
		"Buy":   F3dActionBuy,
	}
}

func (t *f3dType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogf3dStart: {Ty: reflect.TypeOf(ReceiptF3D{}), Name: "LogStartF3d"},
		TyLogf3dDraw:  {Ty: reflect.TypeOf(ReceiptF3D{}), Name: "LogDrawF3d"},
		TyLogf3dBuy:   {Ty: reflect.TypeOf(ReceiptF3D{}), Name: "LogBuyF3d"},
	}
}
