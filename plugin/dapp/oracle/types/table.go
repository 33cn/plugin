package types

import (
	"fmt"

	"github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
)

/*
table  struct
data:  oracle
index: addr,status,index,type
*/

var opt = &table.Option{
	Prefix:  "LODB",
	Name:    "oracle",
	Primary: "eventid",
	Index:   []string{"status", "addr_status", "type_status"},
}

//NewTable 新建表
func NewTable(kvdb db.KV) *table.Table {
	rowmeta := NewOracleRow()
	table, err := table.NewTable(rowmeta, kvdb, opt)
	if err != nil {
		panic(err)
	}
	return table
}

//OracleRow table meta 结构
type OracleRow struct {
	*ReceiptOracle
}

//NewOracleRow 新建一个meta 结构
func NewOracleRow() *OracleRow {
	return &OracleRow{ReceiptOracle: &ReceiptOracle{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存eventid)
func (tx *OracleRow) CreateRow() *table.Row {
	return &table.Row{Data: &ReceiptOracle{}}
}

//SetPayload 设置数据
func (tx *OracleRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*ReceiptOracle); ok {
		tx.ReceiptOracle = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *OracleRow) Get(key string) ([]byte, error) {
	if key == "eventid" {
		return []byte(tx.EventID), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", tx.Status)), nil
	} else if key == "addr_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.Addr, tx.Status)), nil
	} else if key == "type_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.Type, tx.Status)), nil
	}
	return nil, types.ErrNotFound
}
