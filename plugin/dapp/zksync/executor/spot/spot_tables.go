package spot

import (
	"fmt"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

type spot struct {
	//fromaddr string
	//height   int64
	//statedb  dbm.KV
	//localdb  dbm.KV
	prefix et.DBprefix
}

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

func (s *spot) getDepthOpt() *table.Option {
	return &table.Option{
		Prefix:  s.prefix.GetLocaldbPrefix(),
		Name:    "depth",
		Primary: "price",
		Index:   nil,
	}
}

func (s *spot) getOrderOpt() *table.Option {
	return &table.Option{
		Prefix:  s.prefix.GetLocaldbPrefix(),
		Name:    "order",
		Primary: "orderID",
		Index:   []string{"market_order", "addr_status"},
	}
}

func (s *spot) getHistoryOpt() *table.Option {
	return &table.Option{
		Prefix:  s.prefix.GetLocaldbPrefix(),
		Name:    "history",
		Primary: "index",
		Index:   []string{"name", "addr_status"},
	}
}

//NewMarketDepthTable 新建表
func NewMarketDepthTable(kvdb dbm.KV, p et.DBprefix) *table.Table {
	s := spot{prefix: p}
	rowmeta := NewMarketDepthRow()
	table, err := table.NewTable(rowmeta, kvdb, s.getDepthOpt())
	if err != nil {
		panic(err)
	}
	return table
}

//NewMarketOrderTable ...
func NewMarketOrderTable(kvdb dbm.KV, p et.DBprefix) *table.Table {
	s := spot{prefix: p}
	rowmeta := NewOrderRow()
	table, err := table.NewTable(rowmeta, kvdb, s.getOrderOpt())
	if err != nil {
		panic(err)
	}
	return table
}

//NewHistoryOrderTable ...
func NewHistoryOrderTable(kvdb dbm.KV, p et.DBprefix) *table.Table {
	s := spot{prefix: p}
	rowmeta := NewHistoryOrderRow()
	table, err := table.NewTable(rowmeta, kvdb, s.getHistoryOpt())
	if err != nil {
		panic(err)
	}
	return table
}

//OrderRow table meta 结构
type OrderRow struct {
	*et.SpotOrder
}

//NewOrderRow 新建一个meta 结构
func NewOrderRow() *OrderRow {
	return &OrderRow{SpotOrder: &et.SpotOrder{}}
}

//CreateRow ...
func (r *OrderRow) CreateRow() *table.Row {
	return &table.Row{Data: &et.SpotOrder{}}
}

//SetPayload 设置数据
func (r *OrderRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*et.SpotOrder); ok {
		r.SpotOrder = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *OrderRow) Get(key string) ([]byte, error) {
	if key == "orderID" {
		return []byte(fmt.Sprintf("%022d", r.OrderID)), nil
	} else if key == "market_order" {
		return []byte(fmt.Sprintf("%08d:%08d:%d:%016d", r.GetLimitOrder().LeftAsset, r.GetLimitOrder().RightAsset, r.GetLimitOrder().Op, r.GetLimitOrder().Price)), nil
	} else if key == "addr_status" {
		return []byte(fmt.Sprintf("%s:%d", r.Addr, r.Status)), nil
	}
	return nil, types.ErrNotFound
}

//HistoryOrderRow table meta 结构
type HistoryOrderRow struct {
	*et.SpotOrder
}

//NewHistoryOrderRow ...
func NewHistoryOrderRow() *HistoryOrderRow {
	return &HistoryOrderRow{SpotOrder: &et.SpotOrder{Value: &et.SpotOrder_LimitOrder{LimitOrder: &et.SpotLimitOrder{}}}}
}

//CreateRow ...
func (m *HistoryOrderRow) CreateRow() *table.Row {
	return &table.Row{Data: &et.SpotOrder{Value: &et.SpotOrder_LimitOrder{LimitOrder: &et.SpotLimitOrder{}}}}
}

//SetPayload 设置数据
func (m *HistoryOrderRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*et.SpotOrder); ok {
		m.SpotOrder = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (m *HistoryOrderRow) Get(key string) ([]byte, error) {
	if key == "index" {
		return []byte(fmt.Sprintf("%022d", m.Index)), nil
	} else if key == "name" {
		return []byte(fmt.Sprintf("%08d:%08d", m.GetLimitOrder().LeftAsset, m.GetLimitOrder().RightAsset)), nil
	} else if key == "addr_status" {
		return []byte(fmt.Sprintf("%s:%d", m.Addr, m.Status)), nil
	}
	return nil, types.ErrNotFound
}

//MarketDepthRow table meta 结构
type MarketDepthRow struct {
	*et.SpotMarketDepth
}

//NewMarketDepthRow 新建一个meta 结构
func NewMarketDepthRow() *MarketDepthRow {
	return &MarketDepthRow{SpotMarketDepth: &et.SpotMarketDepth{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存eventid)
func (m *MarketDepthRow) CreateRow() *table.Row {
	return &table.Row{Data: &et.SpotMarketDepth{}}
}

//SetPayload 设置数据
func (m *MarketDepthRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*et.SpotMarketDepth); ok {
		m.SpotMarketDepth = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (m *MarketDepthRow) Get(key string) ([]byte, error) {
	if key == "price" {
		return []byte(fmt.Sprintf("%08d:%08d:%d:%016d", m.LeftAsset, m.RightAsset, m.Op, m.Price)), nil
	}
	return nil, types.ErrNotFound
}
