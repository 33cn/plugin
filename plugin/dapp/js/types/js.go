package types

import (
	"errors"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
)

// action for executor
const (
	jsActionCreate = 0
	jsActionCall   = 1
)

//日志类型
const (
	TyLogJsCreate = iota + 1
	TyLogJsCall
)

var (
	typeMap = map[string]int32{
		"Create": jsActionCreate,
		"Call":   jsActionCall,
	}
	logMap = map[int64]*types.LogInfo{}
)

//JsX 插件名字
var JsX = "js"

//错误常量
var (
	ErrDupName           = errors.New("ErrDupName")
	ErrJsReturnNotObject = errors.New("ErrJsReturnNotObject")
	ErrJsReturnKVSFormat = errors.New("ErrJsReturnKVSFormat")
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(JsX))
	types.RegistorExecutor(JsX, NewType())
}

//JsType 类型
type JsType struct {
	types.ExecTypeBase
}

//NewType 新建一个plugin 类型
func NewType() *JsType {
	c := &JsType{}
	c.SetChild(c)
	return c
}

//GetPayload 获取 交易构造
func (t *JsType) GetPayload() types.Message {
	return &jsproto.JsAction{}
}

//GetTypeMap 获取类型映射
func (t *JsType) GetTypeMap() map[string]int32 {
	return typeMap
}

//GetLogMap 获取日志映射
func (t *JsType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}
