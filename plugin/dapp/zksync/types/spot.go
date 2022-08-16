package types

import (
	"fmt"
	"math/big"
	"runtime"

	"github.com/33cn/chain33/types"
)

var (
	ErrAssetAmount  = fmt.Errorf("%s", "The asset amount is not valid!")
	ErrAssetPrice   = fmt.Errorf("%s", "The asset price is not valid!")
	ErrAssetOp      = fmt.Errorf("%s", "The asset op is not define!")
	ErrAssetBalance = fmt.Errorf("%s", "Insufficient balance!")
	ErrOrderSatus   = fmt.Errorf("%s", "The order status is reovked or completed!")
	ErrAddr         = fmt.Errorf("%s", "Wrong Addr!")
	ErrAsset        = fmt.Errorf("%s", "The asset's execer or symbol can't be nil,The same assets cannot be exchanged!")
	ErrCount        = fmt.Errorf("%s", "The param count can't large  20")
	ErrDirection    = fmt.Errorf("%s", "The direction only 0 or 1!")
	ErrStatus       = fmt.Errorf("%s", "The status only in  0 , 1, 2!")
	ErrOrderID      = fmt.Errorf("%s", "Wrong OrderID!")

	ErrCfgFmt   = fmt.Errorf("%s", "ErrCfgFmt")
	ErrBindAddr = fmt.Errorf("%s", "The address is not bound")

	// 资产处理
	ErrDexNotEnough  = fmt.Errorf("%s", "token balance not enough")
	ErrSpotFeeConfig = fmt.Errorf("%s", "spot fee config")
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TySpotNilAction = iota + 1000
	TyLimitOrderAction
	TyMarketOrderAction
	TyRevokeOrderAction
	TyExchangeBindAction
	TyEntrustOrderAction
	TyEntrustRevokeOrderAction
	TyNftOrderAction
	TyNftTakerOrderAction
	// evmxgo nft order
	TyNftOrder2Action
	TyNftTakerOrder2Action
	TyAssetLimitOrderAction

	NameLimitOrderAction         = "LimitOrder"
	NameMarketOrderAction        = "MarketOrder"
	NameRevokeOrderAction        = "RevokeOrder"
	NameExchangeBindAction       = "ExchangeBind"
	NameEntrustOrderAction       = "EntrustOrder"
	NameEntrustRevokeOrderAction = "EntrustRevokeOrder"
	NameNftOrderAction           = "NftOrder"
	NameNftTakerOrderAction      = "NftTakerOrder"
	NameNftOrder2Action          = "NftOrder2"
	NameNftTakerOrder2Action     = "NftTakerOrder2"
	NameAssetLimitOrderAction    = "AssetLimitOrder"

	FuncNameQueryMarketDepth      = "QueryMarketDepth"
	FuncNameQueryHistoryOrderList = "QueryHistoryOrderList"
	FuncNameQueryOrder            = "QueryOrder"
	FuncNameQueryOrderList        = "QueryOrderList"
)

// log类型id值
const (
	TySpotUnknowLog = iota + 1000
	TyLimitOrderLog
	TyMarketOrderLog
	TyRevokeOrderLog
	TyNftOrderLog
	TyNftTakerOrderLog
	TyExchangeBindLog
	TySpotTradeLog

	// account logs
	TyDexAccountFrozen = iota + 1100
	TyDexAccountActive
	TyDexAccountBurn
	TyDexAccountMint
	TyDexAccountTransfer
	TyDexAccountTransferFrozen
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

/*
const (
	ListDESC = int32(0)
	ListASC  = int32(1)
	ListSeek = int32(2)
) */

const (
	//Count 单次list还回条数
	Count = int32(10)
	//MaxMatchCount 系统最大撮合深度
	MaxMatchCount = 100

	QeuryCountLmit = 20
	PriceLimit     = 1e16
)

var (
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	//ForkFix Forks
	ForkParamV1 = "ForkParamV1"
)

// SpotInitFork defines register fork
func SpotInitFork(cfg *types.Chain33Config) {
	//cfg.RegisterDappFork(ExecName, ForkParamV1, 0)
	return
}

// config part
var MverPrefix = "mver.exec.sub." + ExecName // [mver.exec.sub.zkspot]

func CheckIsNormalToken(id uint64) bool {
	return id < SystemNFTTokenId
}

func CheckIsNFTToken(id uint64) bool {
	return id > SystemNFTTokenId
}

//CheckPrice price  1<=price<=1e16
func CheckPrice(price int64) bool {
	if price > int64(PriceLimit) || price < 1 {
		return false
	}
	return true
}

//CheckOp ...
func CheckOp(op int32) bool {
	if op == OpBuy || op == OpSell {
		return true
	}
	return false
}

//CheckCount ...
func CheckCount(count int32) bool {
	return count <= QeuryCountLmit && count >= 0
}

//CheckAmount 最小交易 1coin
func CheckAmount(amount, coinPrecision int64) bool {
	if amount < 1 || amount >= types.MaxCoin*coinPrecision {
		return false
	}
	return true
}

//CheckDirection ...
func CheckDirection(direction int32) bool {
	if direction == ListASC || direction == ListDESC {
		return true
	}
	return false
}

//CheckStatus ...
func CheckStatus(status int32) bool {
	if status == Ordered || status == Completed || status == Revoked {
		return true
	}
	return false
}

//CheckExchangeAsset
func CheckExchangeAsset(coinExec string, left, right uint64) bool {
	if left == right {
		return false
	}
	return true
}

func CheckLimitOrder(cfg *types.Chain33Config, limitOrder *SpotLimitOrder) error {
	left := limitOrder.GetLeftAsset()
	right := limitOrder.GetRightAsset()
	price := limitOrder.GetPrice()
	amount := limitOrder.GetAmount()
	op := limitOrder.GetOp()
	if !CheckExchangeAsset(cfg.GetCoinExec(), left, right) {
		return ErrAsset
	}
	if !CheckPrice(price) {
		return ErrAssetPrice
	}
	if !CheckAmount(amount, cfg.GetCoinPrecision()) {
		return ErrAssetAmount
	}
	if !CheckOp(op) {
		return ErrAssetOp
	}
	return nil
}

func CheckAssetLimitOrder(cfg *types.Chain33Config, order *SpotAssetLimitOrder) error {
	//left := order.GetLeftAsset()
	//right := order.GetRightAsset()
	price := order.GetPrice()
	amount := order.GetAmount()
	op := order.GetOp()
	//if !CheckExchangeAsset(cfg.GetCoinExec(), left, right) {
	//	return ErrAsset
	//}
	if !CheckPrice(price) {
		return ErrAssetPrice
	}
	if !CheckAmount(amount, cfg.GetCoinPrecision()) {
		return ErrAssetAmount
	}
	if !CheckOp(op) {
		return ErrAssetOp
	}
	return nil
}

func CheckNftOrder(cfg *types.Chain33Config, limitOrder *SpotNftOrder) error {
	left := limitOrder.GetLeftAsset()
	right := limitOrder.GetRightAsset()
	price := limitOrder.GetPrice()
	amount := limitOrder.GetAmount()
	if !CheckExchangeAsset(cfg.GetCoinExec(), left, right) {
		return ErrAsset
	}
	if !CheckPrice(price) {
		return ErrAssetPrice
	}
	if !CheckAmount(amount, cfg.GetCoinPrecision()) {
		return ErrAssetAmount
	}
	if !(CheckIsNFTToken(left) && CheckIsNFTToken(right)) {
		return ErrAsset
	}
	return nil
}

func CheckNftOrder2(cfg *types.Chain33Config, limitOrder *SpotNftOrder) error {
	//left := limitOrder.GetLeftAsset()
	right := limitOrder.GetRightAsset()
	price := limitOrder.GetPrice()
	amount := limitOrder.GetAmount()
	//if !CheckExchangeAsset(cfg.GetCoinExec(), left, right) {
	//	return ErrAsset
	//}
	if !CheckPrice(price) {
		return ErrAssetPrice
	}
	if !CheckAmount(amount, cfg.GetCoinPrecision()) {
		return ErrAssetAmount
	}
	if CheckIsNFTToken(right) {
		return ErrAsset
	}
	return nil
}

type DBprefix interface {
	GetLocaldbPrefix() string
	GetStatedbPrefix() string
}

type TxInfo struct {
	Index    int
	Hash     []byte
	From     string
	To       string
	ExecAddr string
	Tx       *types.Transaction
}

// eth precision : 1e18, chain33 precision : 1e8
const (
	precisionDiff = 1e10
)

func AmountFromZksync(s string) (uint64, error) {
	zkAmount, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return 0, ErrAssetAmount
	}
	chain33Amount := new(big.Int).Div(zkAmount, big.NewInt(precisionDiff))
	if !chain33Amount.IsUint64() {
		return 0, ErrAssetAmount
	}
	return chain33Amount.Uint64(), nil
}

func AmountToZksync(a uint64) string {
	amount := new(big.Int).Mul(new(big.Int).SetUint64(a), big.NewInt(precisionDiff))
	return amount.String()
}

func NftAmountToZksync(a uint64) string {
	return new(big.Int).SetUint64(a).String()
}

func NftAmountFromZksync(s string) (uint64, error) {
	zkAmount, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return 0, ErrAssetAmount
	}
	return zkAmount.Uint64(), nil
}

func MergeReceipt(receipt1, receipt2 *types.Receipt) *types.Receipt {
	if receipt2 != nil {
		receipt1.KV = append(receipt1.KV, receipt2.KV...)
		receipt1.Logs = append(receipt1.Logs, receipt2.Logs...)
	}

	return receipt1
}

func GetStack() string {
	buf := make([]byte, 1<<12)
	runtime.Stack(buf, false)
	return string(buf)
}
