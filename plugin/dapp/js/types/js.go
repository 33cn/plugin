package types

import (
	"errors"
	"reflect"

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
	TyLogJs = 10000
)

// JsCreator 配置项 创建js合约的管理员
const JsCreator = "js-creator"

var (
	typeMap = map[string]int32{
		"Create": jsActionCreate,
		"Call":   jsActionCall,
	}
	logMap = map[int64]*types.LogInfo{
		TyLogJs: {Ty: reflect.TypeOf(jsproto.JsLog{}), Name: "TyLogJs"},
	}
)

//JsX 插件名字
var JsX = "jsvm"

//错误常量
var (
	ErrDupName            = errors.New("ErrDupName")
	ErrJsReturnNotObject  = errors.New("ErrJsReturnNotObject")
	ErrJsReturnKVSFormat  = errors.New("ErrJsReturnKVSFormat")
	ErrJsReturnLogsFormat = errors.New("ErrJsReturnLogsFormat")
	//ErrInvalidFuncFormat 错误的函数调用格式(没有_)
	ErrInvalidFuncFormat = errors.New("chain33.js: invalid function name format")
	//ErrInvalidFuncPrefix not exec_ execloal_ query_
	ErrInvalidFuncPrefix = errors.New("chain33.js: invalid function prefix format")
	//ErrFuncNotFound 函数没有找到
	ErrFuncNotFound = errors.New("chain33.js: invalid function name not found")
	ErrSymbolName   = errors.New("chain33.js: ErrSymbolName")
	ErrExecerName   = errors.New("chain33.js: ErrExecerName")
	ErrDBType       = errors.New("chain33.js: ErrDBType")
	// ErrJsCreator
	ErrJsCreator = errors.New("ErrJsCreator")
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
