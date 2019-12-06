package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	exchangetypes "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

/*
 * 执行器相关定义
 * 重载基类相关接口
 */

var (
	//日志
	elog = log.New("module", "exchange.executor")
)

var driverName = exchangetypes.ExchangeX

// Init register dapp
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newExchange, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// InitExecType Init Exec Type
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&exchange{}))
}

type exchange struct {
	drivers.DriverBase
}

func newExchange() drivers.Driver {
	t := &exchange{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName get driver name
func GetName() string {
	return newExchange().GetName()
}

func (e *exchange) GetDriverName() string {
	return driverName
}

// CheckTx 实现自定义检验交易接口，供框架调用
func (e *exchange) CheckTx(tx *types.Transaction, index int) error {
	//发送交易的时候就检查payload,做严格的参数检查
	var exchange exchangetypes.ExchangeAction
	types.Decode(tx.GetPayload(), &exchange)
	if exchange.Ty == exchangetypes.TyLimitOrderAction {
		limitOrder := exchange.GetLimitOrder()
		left := limitOrder.GetLeftAsset()
		right := limitOrder.GetRightAsset()
		price := Truncate(limitOrder.GetPrice())
		amount := limitOrder.GetAmount()
		op := limitOrder.GetOp()
		if !CheckExchangeAsset(left, right) {
			return exchangetypes.ErrAsset
		}
		if !CheckPrice(price) {
			return exchangetypes.ErrAssetPrice
		}
		if !types.CheckAmount(amount) {
			return exchangetypes.ErrAssetAmount
		}
		if !CheckOp(op) {
			return exchangetypes.ErrAssetOp
		}
	}
	return nil
}

// GetPayloadValue get payload value
func (e *exchange) GetPayloadValue() types.Message {
	return &exchangetypes.ExchangeAction{}
}
