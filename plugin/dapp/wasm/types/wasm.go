package types

import (
	"reflect"
	"regexp"

	"github.com/33cn/chain33/types"
)

var NameReg *regexp.Regexp

const (
	WasmX      = "wasm"
	NameRegExp = "^[a-z0-9]+$"
	//TODO: max size to define
	MaxCodeSize = 1 << 20
)

// action for executor
const (
	WasmActionCreate = iota + 1
	WasmActionUpdate
	WasmActionCall
)

// log ty for executor
const (
	TyLogWasmCreate = iota + 100
	TyLogWasmUpdate
	TyLogWasmCall
	TyLogCustom
	TyLogLocalData
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(WasmX))
	types.RegFork(WasmX, InitFork)
	types.RegExec(WasmX, InitExecutor)

	NameReg, _ = regexp.Compile(NameRegExp)
}

func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(WasmX, "Enable", 0)
}

func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(WasmX, NewType(cfg))
}

type WasmType struct {
	types.ExecTypeBase
}

func NewType(cfg *types.Chain33Config) *WasmType {
	c := &WasmType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

func (t *WasmType) GetPayload() types.Message {
	return &WasmAction{}
}

func (t *WasmType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create": WasmActionCreate,
		"Update": WasmActionUpdate,
		"Call":   WasmActionCall,
	}
}

func (t *WasmType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogWasmCreate: {Ty: reflect.TypeOf(CreateContractLog{}), Name: "LogWasmCreate"},
		TyLogWasmUpdate: {Ty: reflect.TypeOf(UpdateContractLog{}), Name: "LogWasmUpdate"},
		TyLogWasmCall:   {Ty: reflect.TypeOf(CallContractLog{}), Name: "LogWasmCall"},
		TyLogCustom:     {Ty: reflect.TypeOf(CustomLog{}), Name: "LogWasmCustom"},
		TyLogLocalData:  {Ty: reflect.TypeOf(LocalDataLog{}), Name: "LogWasmLocalData"},
	}
}
