// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/address"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

// 本地数据版本
// 1. v0 手动生成
// 2. v1 用table生成 LocalOrder
// 3. v2 用table生成 LocalOrderV2 比v1 用更多的索引，支持v0需要的数据索引，即将v2 包含 v0 v1 数据
//       预期在6.4升级后 v0 格式将不存在

/*
现有接口
 1.  查询地址对应的买单 （无分页）
   1.1 只指定地址   -> owner
   1.2 同时指定地址和token  -> owner_asset
   1.3 显示一个用户成交的所有买单 -> owner
   1.4 显示一个用户成交的指定一个或者多个token所有买单 -> owner_asset 不支持多个
 2. 分状态查询地址的买单： 状态 地址 （无分页） -> owner_status
 3. 显示一个token 指定数量的买单 GetTokenBuyOrderByStatus  -> asset_inBuy_status
 4. 显示指定token出售者的一个或多个token 或 不指定token 的卖单 （无分页） -> owner_asset/owner_asset_isSell 不支持多个
 5. 显示指定状态下的某地址卖单 （无分页）  -> owner_isSell_status
 6. 显示一个token 指定数量的卖单    -> asset_isSell
 7. 根据状态分页列出某地址的订单（包括买单卖单） owner_status
*/

//
var optV2 = &table.Option{
	Prefix:  "LODB-trade",
	Name:    "order_v2",
	Primary: "txIndex",
	// asset 指定交易对 price_exec + price_symbol + asset_exec+asset_symbol
	// status: 设计为可以同时查询几种的并集 , 存储为前缀， 需要提前设计需要合并的， 用前缀表示
	//    进行中，  撤销，  部分成交 ， 全部成交，  完成状态统一前缀. 数字和原来不一样
	//      00     10     11          12         1*
	// 排序特点： 在不用key排序时，需要生成排序用的组合索引， 前面n个索引用来区分前缀， 后一个索引用来排序
	Index: []string{
		"key",                 // 内部查询用
		"asset",               // 按资产统计订单
		"asset_isSell_status", // 接口 3
		// "asset_status", 可能需求， 用于资产的交易历史
		// "asset_isSell",
		"owner",              // 接口 1.1， 1.3
		"owner_asset",        // 接口 1.2, 1.4, 4, 7
		"owner_asset_isSell", // 接口 4
		"owner_asset_status", // 新需求， 在
		"owner_isSell",       // 接口 6
		// "owner_isSell_statusPrefix", // 状态可以定制组合, 成交历史需求
		"owner_status",             // 接口 2
		"assset_isSell_isFinished", // 用 isFinish, 进行订单是否完成的列表功能
		"owner_asset_isFinished",
		"owner_isFinished",
		// "owner_statusPrefix", // 状态可以定制组合 , 成交历史需求
		// 增加更多的key， 把老接口的数据的key 也生成，可以去掉老接口的实现
		// https://chain.33.cn/document/105 文档1.8 sell & asset-price & status, order by price
		// https://chain.33.cn/document/105 文档1.3 buy  & asset-price & status, order by price
		"asset_isSell_status_price",
		// 文档1.2 文档1.5 按 用户状态来 addr-status buy or sell
		"owner_isSell_status",
	},
}

// OrderV2Row order row
type OrderV2Row struct {
	*pty.LocalOrder
}

// NewOrderV2Row create row
func NewOrderV2Row() *OrderV2Row {
	return &OrderV2Row{LocalOrder: nil}
}

// CreateRow create row
func (r *OrderV2Row) CreateRow() *table.Row {
	return &table.Row{Data: &pty.LocalOrder{}}
}

// SetPayload set payload
func (r *OrderV2Row) SetPayload(data types.Message) error {
	if d, ok := data.(*pty.LocalOrder); ok {
		r.LocalOrder = d
		return nil
	}

	return types.ErrTypeAsset
}

// Get get index key
func (r *OrderV2Row) Get(key string) ([]byte, error) {
	switch key {
	case "txIndex":
		return []byte(r.TxIndex), nil
	case "key":
		return []byte(r.Key), nil
	case "asset":
		return []byte(r.asset()), nil
	case "asset_isSell_status":
		return []byte(fmt.Sprintf("%s_%d_%s", r.asset(), r.isSell(), r.status())), nil
	case "owner":
		return []byte(address.FormatAddrKey(r.Owner)), nil
	case "owner_asset":
		return []byte(fmt.Sprintf("%s_%s", address.FormatAddrKey(r.Owner), r.asset())), nil
	case "owner_asset_isSell":
		return []byte(fmt.Sprintf("%s_%s_%d", address.FormatAddrKey(r.Owner), r.asset(), r.isSell())), nil
	case "owner_asset_status":
		return []byte(fmt.Sprintf("%s_%s_%s", address.FormatAddrKey(r.Owner), r.asset(), r.status())), nil
	case "owner_isSell":
		return []byte(fmt.Sprintf("%s_%d", address.FormatAddrKey(r.Owner), r.isSell())), nil
	case "owner_isSell_status":
		return []byte(fmt.Sprintf("%s_%d_%s", address.FormatAddrKey(r.Owner), r.isSell(), r.status())), nil
	case "owner_status":
		return []byte(fmt.Sprintf("%s_%s", address.FormatAddrKey(r.Owner), r.status())), nil
	//case "owner_statusPrefix":
	//	return []byte(fmt.Sprintf("%s_%d", r.Owner, r.isSell())), nil
	case "assset_isSell_isFinished":
		return []byte(fmt.Sprintf("%s_%d_%d", address.FormatAddrKey(r.Owner), r.isSell(), r.isFinished())), nil
	case "owner_asset_isFinished":
		return []byte(fmt.Sprintf("%s_%s_%d", address.FormatAddrKey(r.Owner), r.asset(), r.isFinished())), nil
	case "owner_isFinished":
		return []byte(fmt.Sprintf("%s_%d", address.FormatAddrKey(r.Owner), r.isFinished())), nil
	case "asset_isSell_status_price":
		return []byte(fmt.Sprintf("%s_%d_%s_%s", r.asset(), r.isSell(), r.status(), r.price())), nil
	default:
		return nil, types.ErrNotFound
	}
}

// 老接口查询参数： Price 用主币
func (r *OrderV2Row) asset() string {
	return r.LocalOrder.PriceExec + "." + r.LocalOrder.PriceSymbol + "_" + r.LocalOrder.AssetExec + "." + r.LocalOrder.AssetSymbol
}

func (r *OrderV2Row) isSell() int {
	if r.IsSellOrder {
		return 1
	}
	return 0
}

func (r *OrderV2Row) isFinished() int {
	if r.IsFinished {
		return 1
	}
	return 0
}

func (r *OrderV2Row) price() string {
	// 在计算前缀时，返回空
	if r.AmountPerBoardlot == 0 {
		return ""
	}
	p := calcPriceOfToken(r.PricePerBoardlot, r.AmountPerBoardlot)
	return fmt.Sprintf("%018d", p)
}

// status: 设计为可以同时查询几种的并集 , 存储为前缀， 需要提前设计需要合并的， 用前缀表示
//    进行中，  撤销，  部分成交 ， 全部成交，  完成状态统一前缀. 数字和原来不一样
//      01     10     11          12        19 -> 1*
func (r *OrderV2Row) status() string {
	if r.Status == pty.TradeOrderStatusOnBuy || r.Status == pty.TradeOrderStatusOnSale {
		return "01" // 试图用1 可以匹配所有完成的
	} else if r.Status == pty.TradeOrderStatusSoldOut || r.Status == pty.TradeOrderStatusBoughtOut {
		return "12"
	} else if r.Status == pty.TradeOrderStatusRevoked || r.Status == pty.TradeOrderStatusBuyRevoked {
		return "10"
	} else if r.Status == pty.TradeOrderStatusSellHalfRevoked || r.Status == pty.TradeOrderStatusBuyHalfRevoked {
		return "11"
	} else if r.Status == pty.TradeOrderStatusGroupComplete {
		return "1" // 1* match complete
	}

	return "XX"
}

// NewOrderTableV2 create order table
func NewOrderTableV2(kvdb dbm.KV) *table.Table {
	rowMeta := NewOrderV2Row()
	rowMeta.SetPayload(&pty.LocalOrder{})
	t, err := table.NewTable(rowMeta, kvdb, optV2)
	if err != nil {
		panic(err)
	}
	return t
}

func listV2(db dbm.KVDB, indexName string, data *pty.LocalOrder, count, direction int32) ([]*table.Row, error) {
	query := NewOrderTableV2(db).GetQuery(db)
	var primary []byte
	if len(data.TxIndex) > 0 {
		primary = []byte(data.TxIndex)
	}

	cur := &OrderV2Row{LocalOrder: data}
	index, err := cur.Get(indexName)
	if err != nil {
		tradelog.Error("query List failed", "key", string(primary), "param", data, "err", err)
		return nil, err
	}
	tradelog.Debug("query List dbg", "indexName", indexName, "index", string(index), "primary", primary, "count", count, "direction", direction)
	rows, err := query.ListIndex(indexName, index, primary, count, direction)
	if err != nil {
		tradelog.Error("query List failed", "key", string(primary), "param", data, "err", err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, types.ErrNotFound
	}
	return rows, nil
}
