package types

import (
	"reflect"

	"github.com/33cn/chain33/types"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TyUnknowAction = iota + 200
	TyLimitOrderAction
	TyMarketOrderAction
	TyRevokeOrderAction

	NameLimitOrderAction  = "LimitOrder"
	NameMarketOrderAction = "MarketOrder"
	NameRevokeOrderAction = "RevokeOrder"

	FuncNameQueryMarketDepth        = "QueryMarketDepth"
	FuncNameQueryCompletedOrderList = "QueryCompletedOrderList"
	FuncNameQueryOrder              = "QueryOrder"
	FuncNameQueryOrderList          = "QueryOrderList"
)

// log类型id值
const (
	TyUnknownLog = iota + 200
	TyLimitOrderLog
	TyMarketOrderLog
	TyRevokeOrderLog
)

// OP
const (
	OpBuy = iota + 1
	OpSell
)

//order status
const (
	Ordered = iota
	Completed
	Revoked
)

//const
const (
	ListDESC = int32(0)
	ListASC  = int32(1)
	ListSeek = int32(2)
)

const (
	//单次list还回条数
	Count = int32(5)
	//系统最大撮合深度
	MaxCount = 100
)

var (
	//ExchangeX 执行器名称定义
	ExchangeX = "exchange"
	//定义actionMap
	actionMap = map[string]int32{
		NameLimitOrderAction:  TyLimitOrderAction,
		NameMarketOrderAction: TyMarketOrderAction,
		NameRevokeOrderAction: TyRevokeOrderAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		TyLimitOrderLog:  {Ty: reflect.TypeOf(ReceiptExchange{}), Name: "TyLimitOrderLog"},
		TyMarketOrderLog: {Ty: reflect.TypeOf(ReceiptExchange{}), Name: "TyMarketOrderLog"},
		TyRevokeOrderLog: {Ty: reflect.TypeOf(ReceiptExchange{}), Name: "TyRevokeOrderLog"},
	}
	//tlog = log.New("module", "exchange.types")
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(ExchangeX))
	//注册合约启用高度
	types.RegFork(ExchangeX, InitFork)
	types.RegExec(ExchangeX, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(ExchangeX, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(ExchangeX, NewType(cfg))
}

type ExchangeType struct {
	types.ExecTypeBase
}

func NewType(cfg *types.Chain33Config) *ExchangeType {
	c := &ExchangeType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (e *ExchangeType) GetPayload() types.Message {
	return &ExchangeAction{}
}

// GeTypeMap 获取合约action的id和name信息
func (e *ExchangeType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (e *ExchangeType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}
