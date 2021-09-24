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
data:  autonomy change
index: status, addr
*/

var changeOpt = &table.Option{
	Prefix:  "LODB-autonomy",
	Name:    "change",
	Primary: "heightindex",
	Index:   []string{"addr", "status", "addr_status"},
}

//NewChangeTable 新建表
func NewChangeTable(kvdb db.KV) *table.Table {
	rowmeta := NewChangeRow()
	newTable, err := table.NewTable(rowmeta, kvdb, changeOpt)
	if err != nil {
		panic(err)
	}
	return newTable
}

//ChangeRow table meta 结构
type ChangeRow struct {
	*auty.AutonomyProposalChange
}

//NewChangeRow 新建一个meta 结构
func NewChangeRow() *ChangeRow {
	return &ChangeRow{AutonomyProposalChange: &auty.AutonomyProposalChange{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存heightindex)
func (r *ChangeRow) CreateRow() *table.Row {
	return &table.Row{Data: &auty.AutonomyProposalChange{}}
}

//SetPayload 设置数据
func (r *ChangeRow) SetPayload(data types.Message) error {
	if d, ok := data.(*auty.AutonomyProposalChange); ok {
		r.AutonomyProposalChange = d
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *ChangeRow) Get(key string) ([]byte, error) {
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
