package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

/*
table  struct
data:  autonomy board
index: status, addr
*/

var boardOpt = &table.Option{
	Prefix:  "LODB-autonomy",
	Name:    "board",
	Primary: "heightindex",
	Index:   []string{"addr", "status", "addr_status"},
}

//NewBoardTable 新建表
func NewBoardTable(kvdb db.KV) *table.Table {
	rowmeta := NewBoardRow()
	newTable, err := table.NewTable(rowmeta, kvdb, boardOpt)
	if err != nil {
		panic(err)
	}
	return newTable
}

//BoardRow table meta 结构
type BoardRow struct {
	*auty.AutonomyProposalBoard
}

//NewBoardRow 新建一个meta 结构
func NewBoardRow() *BoardRow {
	return &BoardRow{AutonomyProposalBoard: &auty.AutonomyProposalBoard{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存heightindex)
func (r *BoardRow) CreateRow() *table.Row {
	return &table.Row{Data: &auty.AutonomyProposalBoard{}}
}

//SetPayload 设置数据
func (r *BoardRow) SetPayload(data types.Message) error {
	if d, ok := data.(*auty.AutonomyProposalBoard); ok {
		r.AutonomyProposalBoard = d
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *BoardRow) Get(key string) ([]byte, error) {
	if key == "heightindex" {
		return []byte(dapp.HeightIndexStr(r.Height, int64(r.Index))), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", r.Status)), nil
	} else if key == "addr" {
		return []byte(r.Address), nil
	} else if key == "addr_status" {
		return []byte(fmt.Sprintf("%s:%2d", r.Address, r.Status)), nil
	}
	return nil, types.ErrNotFound
}
