package types

import (
	"github.com/33cn/chain33/types"
)

// action for executor
const (
	f3dActionStart = 0
	f3dActionDraw  = 1
	f3dActionBuy   = 2
)

const (
	TyLogf3dUnknown = iota
	TyLogf3dStart
	TyLogf3dDraw
	TyLogf3dBuy
)

var (
	typeMap = map[string]int32{
		"Start": f3dActionStart,
		"Draw":  f3dActionDraw,
		"Buy":   f3dActionBuy,
	}

	logMap    = map[int64]*types.LogInfo{}
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
	return typeMap
}

func (t *f3dType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}
