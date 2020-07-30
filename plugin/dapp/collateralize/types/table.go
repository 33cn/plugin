package types

import (
	"fmt"

	"github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
)

var opt = &table.Option{
	Prefix:  "LODB-collateralize",
	Name:    "coller",
	Primary: "collateralizeid",
	Index:   []string{"status", "addr", "addr_status"},
}

//NewCollateralizeTable 新建表
func NewCollateralizeTable(kvdb db.KV) *table.Table {
	rowmeta := NewCollatetalizeRow()
	table, err := table.NewTable(rowmeta, kvdb, opt)
	if err != nil {
		panic(err)
	}
	return table
}

//CollatetalizeRow table meta 结构
type CollatetalizeRow struct {
	*ReceiptCollateralize
}

//NewCollatetalizeRow 新建一个meta 结构
func NewCollatetalizeRow() *CollatetalizeRow {
	return &CollatetalizeRow{ReceiptCollateralize: &ReceiptCollateralize{}}
}

//CreateRow 新建数据行
func (tx *CollatetalizeRow) CreateRow() *table.Row {
	return &table.Row{Data: &ReceiptCollateralize{}}
}

//SetPayload 设置数据
func (tx *CollatetalizeRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*ReceiptCollateralize); ok {
		tx.ReceiptCollateralize = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *CollatetalizeRow) Get(key string) ([]byte, error) {
	if key == "collateralizeid" {
		return []byte(tx.CollateralizeId), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", tx.Status)), nil
	} else if key == "addr" {
		return []byte(tx.AccountAddr), nil
	} else if key == "addr_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.AccountAddr, tx.Status)), nil
	}
	return nil, types.ErrNotFound
}

var optRecord = &table.Option{
	Prefix:  "LODB-collateralize",
	Name:    "borrow",
	Primary: "borrowid",
	Index:   []string{"status", "addr", "addr_status", "id_status", "id_addr"},
}

// NewRecordTable 借贷记录表
func NewRecordTable(kvdb db.KV) *table.Table {
	rowmeta := NewRecordRow()
	table, err := table.NewTable(rowmeta, kvdb, optRecord)
	if err != nil {
		panic(err)
	}
	return table
}

//CollateralizeRecordRow table meta 结构
type CollateralizeRecordRow struct {
	*ReceiptCollateralize
}

//NewRecordRow 新建一个meta 结构
func NewRecordRow() *CollateralizeRecordRow {
	return &CollateralizeRecordRow{ReceiptCollateralize: &ReceiptCollateralize{}}
}

//CreateRow 新建数据行
func (tx *CollateralizeRecordRow) CreateRow() *table.Row {
	return &table.Row{Data: &ReceiptCollateralize{}}
}

//SetPayload 设置数据
func (tx *CollateralizeRecordRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*ReceiptCollateralize); ok {
		tx.ReceiptCollateralize = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *CollateralizeRecordRow) Get(key string) ([]byte, error) {
	if key == "borrowid" {
		return []byte(tx.RecordId), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", tx.Status)), nil
	} else if key == "addr" {
		return []byte(tx.AccountAddr), nil
	} else if key == "addr_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.AccountAddr, tx.Status)), nil
	} else if key == "id_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.CollateralizeId, tx.Status)), nil
	} else if key == "id_addr" {
		return []byte(fmt.Sprintf("%s:%s", tx.CollateralizeId, tx.AccountAddr)), nil
	}
	return nil, types.ErrNotFound
}
