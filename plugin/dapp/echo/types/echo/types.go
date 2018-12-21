package echo

import (
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

// 定义本执行器支持的Action种类
const (
	ActionPing = iota
	ActionPang
)

// 定义本执行器生成的log类型
const (
	TyLogPing = 100001
	TyLogPang = 100002
)

var (
	// EchoX 本执行器名称
	EchoX = "echo"
	// 定义本执行器支持的Action对应关系
	actionName = map[string]int32{
		"Ping": ActionPing,
		"Pang": ActionPang,
	}
	// 定义本执行器的Log收据解析结构
	logInfo = map[int64]*types.LogInfo{
		TyLogPing: {Ty: reflect.TypeOf(PingLog{}), Name: "PingLog"},
		TyLogPang: {Ty: reflect.TypeOf(PangLog{}), Name: "PangLog"},
	}
)
var elog = log.New("module", EchoX)

func init() {
	// 将本执行器添加到系统白名单
	types.AllowUserExec = append(types.AllowUserExec, []byte(EchoX))
	// 向系统注册本执行器类型
	types.RegistorExecutor(EchoX, NewType())
}

// Type 定义本执行器类型
type Type struct {
	types.ExecTypeBase
}

// NewType 初始化本执行器类型
func NewType() *Type {
	c := &Type{}
	c.SetChild(c)
	return c
}

// GetPayload 返回本执行器的负载类型
func (b *Type) GetPayload() types.Message {
	return &EchoAction{}
}

// GetName 返回本执行器名称
func (b *Type) GetName() string {
	return EchoX
}

// GetTypeMap 返回本执行器中的action字典，支持双向查找
func (b *Type) GetTypeMap() map[string]int32 {
	return actionName
}

// GetLogMap 返回本执行器的日志类型信息，用于rpc解析日志数据
func (b *Type) GetLogMap() map[int64]*types.LogInfo {
	return logInfo
}
