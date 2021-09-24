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
data:  autonomy project
index: status, addr
*/

var projectOpt = &table.Option{
	Prefix:  "LODB-autonomy",
	Name:    "project",
	Primary: "heightindex",
	Index:   []string{"addr", "status", "addr_status"},
}

//NewProjectTable 新建表
func NewProjectTable(kvdb db.KV) *table.Table {
	rowmeta := NewProjectRow()
	newTable, err := table.NewTable(rowmeta, kvdb, projectOpt)
	if err != nil {
		panic(err)
	}
	return newTable
}

//ProjectRow table meta 结构
type ProjectRow struct {
	*auty.AutonomyProposalProject
}

//NewProjectRow 新建一个meta 结构
func NewProjectRow() *ProjectRow {
	return &ProjectRow{AutonomyProposalProject: &auty.AutonomyProposalProject{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存heightindex)
func (r *ProjectRow) CreateRow() *table.Row {
	return &table.Row{Data: &auty.AutonomyProposalProject{}}
}

//SetPayload 设置数据
func (r *ProjectRow) SetPayload(data types.Message) error {
	if d, ok := data.(*auty.AutonomyProposalProject); ok {
		r.AutonomyProposalProject = d
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *ProjectRow) Get(key string) ([]byte, error) {
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
