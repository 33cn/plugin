package types

import (
	"fmt"

	"github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
)

/*
table  struct
data:  guess
index: addr,status,addr_status,admin,admin_status,category_status
*/

var opt = &table.Option{
	Prefix:  "LODB",
	Name:    "guess",
	Primary: "gameid",
	Index:   []string{"addr", "status", "addr_status", "admin", "admin_status", "category_status"},
}

//NewTable 新建表
func NewTable(kvdb db.KV) *table.Table {
	rowmeta := NewGuessRow()
	table, err := table.NewTable(rowmeta, kvdb, opt)
	if err != nil {
		panic(err)
	}
	return table
}

//OracleRow table meta 结构
type GuessRow struct {
	*ReceiptGuessGame
}

//NewOracleRow 新建一个meta 结构
func NewGuessRow() *GuessRow {
	return &GuessRow{ReceiptGuessGame: &ReceiptGuessGame{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存eventid)
func (tx *GuessRow) CreateRow() *table.Row {
	return &table.Row{Data: &ReceiptGuessGame{}}
}

//SetPayload 设置数据
func (tx *GuessRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*ReceiptGuessGame); ok {
		tx.ReceiptGuessGame = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *GuessRow) Get(key string) ([]byte, error) {
	if key == "gameid" {
		return []byte(tx.GameID), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", tx.Status)), nil
	} else if key == "addr" {
		return []byte(tx.Addr), nil
	} else if key == "addr_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.Addr, tx.Status)), nil
	} else if key == "admin" {
		return []byte(tx.AdminAddr), nil
	} else if key == "admin_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.AdminAddr, tx.Status)), nil
	} else if key == "category_status" {
		return []byte(fmt.Sprintf("%s:%2d", tx.Category, tx.Status)), nil
	}
	return nil, types.ErrNotFound
}
