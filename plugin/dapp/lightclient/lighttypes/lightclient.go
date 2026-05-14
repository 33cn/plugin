package lighttypes

import (
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TyUnknowAction = iota + 100
	TyBtcHeadersAction

	NameBtcHeadersAction = "BtcHeaders"
)

// log类型id值
const (
	TyUnknownLog = iota + 100
	TyBtcHeadersLog

	NameBtcHeadersLog = "BtcHeadersLog"
)

var (
	//LightclientX 执行器名称定义
	LightclientX = "lightclient"
	//定义actionMap
	actionMap = map[string]int32{
		NameBtcHeadersAction: TyBtcHeadersAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		TyBtcHeadersLog: {Ty: reflect.TypeOf(BtcHeadersLog{}), Name: NameBtcHeadersLog},
	}
	tlog = log.New("module", "lightclient.types")
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(LightclientX))
	//注册合约启用高度
	types.RegFork(LightclientX, InitFork)
	types.RegExec(LightclientX, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(LightclientX, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(LightclientX, NewType(cfg))
}

type lightClientType struct {
	types.ExecTypeBase
}

func NewType(cfg *types.Chain33Config) *lightClientType {
	c := &lightClientType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (l *lightClientType) GetPayload() types.Message {
	return &LightClientAction{}
}

// GetTypeMap 获取合约action的id和name信息
func (l *lightClientType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (l *lightClientType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}
