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
	// 本执行器名称
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

// 定义本执行器类型
type EchoType struct {
	types.ExecTypeBase
}

// 初始化本执行器类型
func NewType() *EchoType {
	c := &EchoType{}
	c.SetChild(c)
	return c
}

// 返回本执行器的负载类型
func (b *EchoType) GetPayload() types.Message {
	return &EchoAction{}
}

// 返回本执行器名称
func (b *EchoType) GetName() string {
	return EchoX
}

// 返回本执行器中的action字典，支持双向查找
func (b *EchoType) GetTypeMap() map[string]int32 {
	return actionName
}

// 返回本执行器的日志类型信息，用于rpc解析日志数据
func (b *EchoType) GetLogMap() map[int64]*types.LogInfo {
	return logInfo
}
