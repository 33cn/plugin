package spot

import (
	"encoding/hex"
	"fmt"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

type orderInit func(*et.SpotOrder) *et.SpotOrder

func createOrder(or *et.SpotOrder, entrustAddr string, inits []orderInit) *et.SpotOrder {
	or.Status = et.Ordered
	or.EntrustAddr = entrustAddr
	//	Executed: 0,
	//	AVGPrice: 0,
	for _, initFun := range inits {
		or = initFun(or)
	}
	return or
}

func PreCreateLimitOrder(payload *et.SpotLimitOrder) *Order {
	or := &et.SpotOrder{
		Value:   &et.SpotOrder_LimitOrder{LimitOrder: payload},
		Ty:      et.TyLimitOrderAction,
		Balance: payload.GetAmount(),
	}
	return NewOrder(or, nil)
}

func PreCreateAssetLimitOrder(payload *et.SpotAssetLimitOrder) *Order {
	or := &et.SpotOrder{
		Value:   &et.SpotOrder_AssetLimitOrder{AssetLimitOrder: payload},
		Ty:      et.TyAssetLimitOrderAction,
		Balance: payload.GetAmount(),
	}
	return NewOrder(or, nil)
}

func PreCreateNftOrder(payload *et.SpotNftOrder, ty int32) *Order {
	or := &et.SpotOrder{
		Value:   &et.SpotOrder_NftOrder{NftOrder: payload},
		Ty:      ty,
		Balance: payload.GetAmount(),
	}
	return NewOrder(or, nil)
}

func PreCreateNftTakerOrder(payload *et.SpotNftTakerOrder, ty int, order2 *Order) *Order {
	or := &et.SpotOrder{
		Value:   &et.SpotOrder_NftTakerOrder{NftTakerOrder: payload},
		Ty:      int32(ty),
		Balance: order2.order.Balance,
	}
	return NewOrder(or, nil)
}

func CreateNftOrder(payload *et.SpotNftOrder, ty int32) *et.SpotOrder {
	or := &et.SpotOrder{
		Value:   &et.SpotOrder_NftOrder{NftOrder: payload},
		Ty:      ty,
		Balance: payload.GetAmount(),
	}
	return or
}

func CreateNftTakerOrder(payload *et.SpotNftTakerOrder, ty int32, order2 *Order) *et.SpotOrder {
	or := &et.SpotOrder{
		Value:   &et.SpotOrder_NftTakerOrder{NftTakerOrder: payload},
		Ty:      ty,
		Balance: order2.order.Balance,
	}
	return or
}

type Order struct {
	order *et.SpotOrder

	repo *orderSRepo
}

func NewOrder(order *et.SpotOrder, orderdb *orderSRepo) *Order {
	return &Order{
		repo:  orderdb,
		order: order,
	}
}

func (o *Order) checkRevoke(fromaddr string) error {
	if o.order.Addr != fromaddr {
		elog.Error("RevokeOrder.OrderCheck", "addr", fromaddr, "order.addr", o.order.Addr, "order.status", o.order.Status)
		return et.ErrAddr
	}
	if o.order.Status == et.Completed || o.order.Status == et.Revoked {
		elog.Error("RevokeOrder.OrderCheck", "addr", fromaddr, "order.addr", o.order.Addr, "order.status", o.order.Status)
		return et.ErrOrderSatus
	}
	return nil
}

func (o *Order) calcFrozenToken(rightPrecision int64) (*et.ZkAsset, uint64) {
	order := o.order
	price := o.GetPrice()
	balance := order.GetBalance()

	left, right := o.GetAsset()
	if o.GetOp() == et.OpBuy {
		amount := CalcActualCost(et.OpBuy, balance, price, rightPrecision)
		amount += SafeMul(amount, int64(order.Rate), rightPrecision)
		return right, uint64(amount)
	}
	return left, uint64(balance)
}

// buy 按最大量判断余额是否够
// 因为在吃单时, 价格是变动的, 所以实际锁定的量是会浮动的
// 实现上, 按最大量判断余额是否够, 在成交时, 按实际需要量扣除. 最后变成挂单时, 进行锁定
func (o *Order) NeedToken(precision int64) (uint64, int64) {
	or := o.order.GetLimitOrder()
	if or.GetOp() == et.OpBuy {
		amount := SafeMul(or.GetAmount(), or.GetPrice(), precision)
		fee := calcMtfFee(amount, int32(o.order.TakerRate), precision)
		total := SafeAdd(amount, int64(fee))
		return or.RightAsset, total
	}

	/* if payload.GetOp() == et.OpSell */
	return or.LeftAsset, or.GetAmount()
}

func (o *Order) Revoke(blockTime int64, txhash []byte, txindex int) (*types.Receipt, error) {
	order := o.order
	order.Status = et.Revoked
	order.UpdateTime = blockTime
	order.RevokeHash = hex.EncodeToString(txhash)
	kvs := o.repo.GetOrderKvSet(order)

	re := &et.ReceiptSpotMatch{
		Order: order,
		Index: int64(txindex),
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyRevokeOrderLog, Log: types.Encode(re)}
	return &types.Receipt{KV: kvs, Logs: []*types.ReceiptLog{receiptlog}}, nil
}

func (o *Order) isActiveOrder() bool {
	return o.order.Status == et.Ordered
}

func (o *Order) orderUpdate(matchDetail *et.MatchInfo) {
	matched := matchDetail.Matched

	// fee and AVGPrice
	o.order.DigestedFee += matchDetail.FeeTaker
	o.order.AVGPrice = matchDetail.Price

	// status
	if matched == o.order.GetBalance() {
		o.order.Status = et.Completed
	} else {
		o.order.Status = et.Ordered
	}

	// order matched
	o.order.Executed = matched
	o.order.Balance -= matched
}

func (o *Order) Traded(matchDetail *et.MatchInfo, blocktime int64) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	o.orderUpdate(matchDetail)
	o.order.UpdateTime = blocktime
	kvs := o.repo.GetOrderKvSet(o.order)
	return []*types.ReceiptLog{}, kvs, nil
}

func (o *Order) GetOp() int32 {
	switch o.order.Ty {
	case et.TyLimitOrderAction:
		return o.order.GetLimitOrder().GetOp()
	case et.TyAssetLimitOrderAction:
		return o.order.GetAssetLimitOrder().GetOp()
	case et.TyNftOrderAction:
		return o.order.GetNftOrder().GetOp()
	}
	panic("Not support op")
}

func (o *Order) GetPrice() int64 {
	switch o.order.Ty {
	case et.TyLimitOrderAction:
		return o.order.GetLimitOrder().GetPrice()
	case et.TyAssetLimitOrderAction:
		return o.order.GetAssetLimitOrder().GetPrice()
	case et.TyNftOrderAction:
		return o.order.GetNftOrder().GetPrice()
	}
	panic("Not support price")
}

func (o *Order) GetAsset() (*et.ZkAsset, *et.ZkAsset) {
	switch o.order.Ty {
	case et.TyLimitOrderAction:
		return NewZkAsset(o.order.GetLimitOrder().LeftAsset), NewZkAsset(o.order.GetLimitOrder().RightAsset)
	case et.TyAssetLimitOrderAction:
		return o.order.GetAssetLimitOrder().LeftAsset, o.order.GetAssetLimitOrder().RightAsset
	case et.TyNftOrderAction:
		return NewZkNftAsset(o.order.GetNftOrder().LeftAsset), NewZkAsset(o.order.GetNftOrder().RightAsset)
	case et.TyNftOrder2Action:
		return NewEvmNftAsset(o.order.GetNftOrder().LeftAsset), NewZkAsset(o.order.GetNftOrder().RightAsset)
	}
	panic("Not support GetAsset")
}

func (o *Order) GetZkOrder() *et.ZkOrder {
	switch o.order.Ty {
	case et.TyLimitOrderAction:
		return o.order.GetLimitOrder().Order
	case et.TyAssetLimitOrderAction:
		return o.order.GetAssetLimitOrder().Order
	case et.TyNftOrderAction:
		return o.order.GetNftOrder().Order
	case et.TyNftOrder2Action:
		return o.order.GetNftOrder().Order
	}
	panic("Not support GetAsset")
}

// statedb: order, account
// localdb: market-depth, market-orders, history-orders

func calcOrderKey(prefix string, orderID int64) []byte {
	return []byte(fmt.Sprintf("%s"+orderKeyFmt, prefix, orderID))
}

func FindOrderByOrderID(statedb dbm.KV, localdb dbm.KV, dbprefix et.DBprefix, orderID int64) (*et.SpotOrder, error) {
	return newOrderSRepo(statedb, dbprefix).findOrderBy(orderID)
}

func FindOrderByOrderNftID(statedb dbm.KV, localdb dbm.KV, dbprefix et.DBprefix, orderID int64) (*et.SpotOrder, error) {
	return newOrderSRepo(statedb, dbprefix).findNftOrderBy(orderID)
}

// orderSRepo statedb repo
type orderSRepo struct {
	statedb  dbm.KV
	dbprefix et.DBprefix
}

func newOrderSRepo(statedb dbm.KV, dbprefix et.DBprefix) *orderSRepo {
	return &orderSRepo{
		statedb:  statedb,
		dbprefix: dbprefix,
	}
}

func (repo *orderSRepo) orderKey(orderID int64) []byte {
	return calcOrderKey(repo.dbprefix.GetStatedbPrefix(), orderID)
}

func (repo *orderSRepo) findOrderBy(orderID int64) (*et.SpotOrder, error) {
	key := repo.orderKey(orderID)
	data, err := repo.statedb.Get(key)
	if err != nil {
		elog.Error("findOrderByOrderID.Get", "orderID", orderID, "err", err.Error())
		return nil, err
	}
	var order et.SpotOrder
	err = types.Decode(data, &order)
	if err != nil {
		elog.Error("findOrderByOrderID.Decode", "orderID", orderID, "err", err.Error())
		return nil, err
	}
	order.Executed = order.GetLimitOrder().Amount - order.Balance
	return &order, nil
}

func (repo *orderSRepo) findNftOrderBy(orderID int64) (*et.SpotOrder, error) {
	key := repo.orderKey(orderID)
	data, err := repo.statedb.Get(key)
	if err != nil {
		elog.Error("findNftOrderBy.Get", "orderID", orderID, "err", err.Error())
		return nil, err
	}
	var order et.SpotOrder
	err = types.Decode(data, &order)
	if err != nil {
		elog.Error("findNftOrderBy.Decode", "orderID", orderID, "err", err.Error())
		return nil, err
	}
	if order.GetNftOrder() == nil {
		elog.Error("findNftOrderBy", "order", "nil")
		return nil, err
	}
	order.Executed = order.GetNftOrder().Amount - order.Balance
	return &order, nil
}

func (repo *orderSRepo) GetOrderKvSet(order *et.SpotOrder) (kvset []*types.KeyValue) {
	kvset = append(kvset, &types.KeyValue{Key: repo.orderKey(order.OrderID), Value: types.Encode(order)})
	return kvset
}

//OpSwap reverse
func OpSwap(op int32) int32 {
	if op == et.OpBuy {
		return et.OpSell
	}
	return et.OpBuy
}

//Direction
//Buying depth is in reverse order by price, from high to low
func Direction(op int32) int32 {
	if op == et.OpBuy {
		return et.ListDESC
	}
	return et.ListASC
}
