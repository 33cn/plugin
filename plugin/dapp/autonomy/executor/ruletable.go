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
data:  autonomy rule
index: status, addr
*/

var ruleOpt = &table.Option{
	Prefix:  "LODB-autonomy",
	Name:    "rule",
	Primary: "heightindex",
	Index:   []string{"addr", "status", "addr_status"},
}

//NewRuleTable 新建表
func NewRuleTable(kvdb db.KV) *table.Table {
	rowmeta := NewRuleRow()
	newTable, err := table.NewTable(rowmeta, kvdb, ruleOpt)
	if err != nil {
		panic(err)
	}
	return newTable
}

//RuleRow table meta 结构
type RuleRow struct {
	*auty.AutonomyProposalRule
}

//NewRuleRow 新建一个meta 结构
func NewRuleRow() *RuleRow {
	return &RuleRow{AutonomyProposalRule: &auty.AutonomyProposalRule{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存heightindex)
func (r *RuleRow) CreateRow() *table.Row {
	return &table.Row{Data: &auty.AutonomyProposalRule{}}
}

//SetPayload 设置数据
func (r *RuleRow) SetPayload(data types.Message) error {
	if d, ok := data.(*auty.AutonomyProposalRule); ok {
		r.AutonomyProposalRule = d
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *RuleRow) Get(key string) ([]byte, error) {
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
