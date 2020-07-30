package types

import (
	"fmt"

	"github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
)

var opt = &table.Option{
	Prefix:  "LODB-issuance",
	Name:    "issuer",
	Primary: "issuanceid",
	Index:   []string{"status"},
}

//NewIssuanceTable 新建表
func NewIssuanceTable(kvdb db.KV) *table.Table {
	rowmeta := NewIssuanceRow()
	table, err := table.NewTable(rowmeta, kvdb, opt)
	if err != nil {
		panic(err)
	}
	return table
}

//IssuanceRow table meta 结构
type IssuanceRow struct {
	*ReceiptIssuanceID
}

//NewIssuanceRow 新建一个meta 结构
func NewIssuanceRow() *IssuanceRow {
	return &IssuanceRow{ReceiptIssuanceID: &ReceiptIssuanceID{}}
}

//CreateRow 新建数据行
func (tx *IssuanceRow) CreateRow() *table.Row {
	return &table.Row{Data: &ReceiptIssuanceID{}}
}

//SetPayload 设置数据
func (tx *IssuanceRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*ReceiptIssuanceID); ok {
		tx.ReceiptIssuanceID = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *IssuanceRow) Get(key string) ([]byte, error) {
	if key == "issuanceid" {
		return []byte(tx.IssuanceId), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", tx.Status)), nil
	}
	return nil, types.ErrNotFound
}

var optRecord = &table.Option{
	Prefix:  "LODB-issuance",
	Name:    "debt",
	Primary: "debtid",
	Index:   []string{"status", "addr", "addr_status"},
}

// NewRecordTable 大户发行记录表
func NewRecordTable(kvdb db.KV) *table.Table {
	rowmeta := NewRecordRow()
	table, err := table.NewTable(rowmeta, kvdb, optRecord)
	if err != nil {
		panic(err)
	}
	return table
}

//IssuanceRecordRow table meta 结构
type IssuanceRecordRow struct {
	*ReceiptIssuance
}

//NewRecordRow 新建一个meta 结构
func NewRecordRow() *IssuanceRecordRow {
	return &IssuanceRecordRow{ReceiptIssuance: &ReceiptIssuance{}}
}

//CreateRow 新建数据行
func (tx *IssuanceRecordRow) CreateRow() *table.Row {
	return &table.Row{Data: &ReceiptIssuance{}}
}

//SetPayload 设置数据
func (tx *IssuanceRecordRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*ReceiptIssuance); ok {
		tx.ReceiptIssuance = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *IssuanceRecordRow) Get(key string) ([]byte, error) {
	if key == "debtid" {
		return []byte(tx.DebtId), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", tx.Status)), nil
	} else if key == "addr" {
		return []byte(tx.AccountAddr), nil
	} else if key == "addr_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.AccountAddr, tx.Status)), nil
	}
	return nil, types.ErrNotFound
}
