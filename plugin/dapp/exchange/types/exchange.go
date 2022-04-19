package types

import (
	"fmt"
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
	TyExchangeBindAction
	TyEntrustOrderAction
	TyEntrustRevokeOrderAction

	NameLimitOrderAction         = "LimitOrder"
	NameMarketOrderAction        = "MarketOrder"
	NameRevokeOrderAction        = "RevokeOrder"
	NameExchangeBindAction       = "ExchangeBind"
	NameEntrustOrderAction       = "EntrustOrder"
	NameEntrustRevokeOrderAction = "EntrustRevokeOrder"

	FuncNameQueryMarketDepth      = "QueryMarketDepth"
	FuncNameQueryHistoryOrderList = "QueryHistoryOrderList"
	FuncNameQueryOrder            = "QueryOrder"
	FuncNameQueryOrderList        = "QueryOrderList"
)

// log类型id值
const (
	TyUnknowLog = iota + 200
	TyLimitOrderLog
	TyMarketOrderLog
	TyRevokeOrderLog

	TyExchangeBindLog
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
	//Count 单次list还回条数
	Count = int32(10)
	//MaxMatchCount 系统最大撮合深度
	MaxMatchCount = 100
)

var (
	//ExchangeX 执行器名称定义
	ExchangeX = "exchange"
	//定义actionMap
	actionMap = map[string]int32{
		NameLimitOrderAction:         TyLimitOrderAction,
		NameMarketOrderAction:        TyMarketOrderAction,
		NameRevokeOrderAction:        TyRevokeOrderAction,
		NameExchangeBindAction:       TyExchangeBindAction,
		NameEntrustOrderAction:       TyEntrustOrderAction,
		NameEntrustRevokeOrderAction: TyEntrustRevokeOrderAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		TyLimitOrderLog:   {Ty: reflect.TypeOf(ReceiptExchange{}), Name: "TyLimitOrderLog"},
		TyMarketOrderLog:  {Ty: reflect.TypeOf(ReceiptExchange{}), Name: "TyMarketOrderLog"},
		TyRevokeOrderLog:  {Ty: reflect.TypeOf(ReceiptExchange{}), Name: "TyRevokeOrderLog"},
		TyExchangeBindLog: {Ty: reflect.TypeOf(ReceiptExchangeBind{}), Name: "TyExchangeBindLog"},
	}
	//tlog = log.New("module", "exchange.types")

	//ForkFix Forks
	//ForkFix1 = "ForkFix1"

	ForkParamV1 = "ForkParamV1"
	ForkParamV2 = "ForkParamV2"
	ForkParamV3 = "ForkParamV3"
	ForkParamV4 = "ForkParamV4"
	ForkParamV5 = "ForkParamV5"
	ForkParamV6 = "ForkParamV6"
	ForkParamV7 = "ForkParamV7"
	ForkParamV8 = "ForkParamV8"
	ForkParamV9 = "ForkParamV9"
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
	//cfg.RegisterDappFork(ExchangeX, ForkFix1, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV1, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV2, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV3, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV4, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV5, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV6, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV7, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV8, 0)
	cfg.RegisterDappFork(ExchangeX, ForkParamV9, 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(ExchangeX, NewType(cfg))
}

//ExchangeType ...
type ExchangeType struct {
	types.ExecTypeBase
}

//NewType ...
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

// GetTypeMap 获取合约action的id和name信息
func (e *ExchangeType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (e *ExchangeType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}

var MverPrefix = "mver.exec.sub." + ExchangeX // [mver.exec.sub.exchange]

type Econfig struct {
	Banks     []string
	Coins     []CoinCfg
	Exchanges map[string]*Trade // 现货交易、杠杠交易
}

type CoinCfg struct {
	Coin   string
	Execer string
	Name   string
}

// 交易对配置
type Trade struct {
	Symbol       string
	PriceDigits  int32
	AmountDigits int32
	Taker        int32
	Maker        int32
	MinFee       int64
}

func (f *Econfig) GetFeeAddr() string {
	return f.Banks[0]
}

func (f *Econfig) GetCoinName(asset *Asset) string {
	for _, v := range f.Coins {
		if v.Coin == asset.GetSymbol() && v.Execer == asset.GetExecer() {
			return v.Name
		}
	}
	return asset.Symbol
}

func (f *Econfig) GetSymbol(left, right *Asset) string {
	return fmt.Sprintf("%v_%v", f.GetCoinName(left), f.GetCoinName(right))
}

func (f *Econfig) GetTrade(or *LimitOrder) *Trade {
	symbol := f.GetSymbol(or.LeftAsset, or.RightAsset)
	c, ok := f.Exchanges[symbol]
	if !ok {
		return nil
	}
	return c
}

func (t *Trade) GetPriceDigits() int32 {
	if t == nil {
		return 0
	}
	return t.PriceDigits
}

func (t *Trade) GetAmountDigits() int32 {
	if t == nil {
		return 0
	}
	return t.AmountDigits
}

func (t *Trade) GetTaker() int32 {
	if t == nil {
		return 100000
	}
	return t.Taker
}

func (t *Trade) GetMaker() int32 {
	if t == nil {
		return 100000
	}
	return t.Maker
}

func (t *Trade) GetMinFee() int64 {
	if t == nil {
		return 0
	}
	return t.MinFee
}
