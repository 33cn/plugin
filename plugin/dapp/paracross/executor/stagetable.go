package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

/*
table  struct
data:  self consens stage
index: status
*/

var boardOpt = &table.Option{
	Prefix:  "LODB-paracross",
	Name:    "stage",
	Primary: "heightindex",
	Index:   []string{"id", "status"},
}

//NewStageTable 新建表
func NewStageTable(kvdb db.KV) *table.Table {
	rowmeta := NewStageRow()
	table, err := table.NewTable(rowmeta, kvdb, boardOpt)
	if err != nil {
		panic(err)
	}
	return table
}

//StageRow table meta 结构
type StageRow struct {
	*pt.LocalSelfConsStageInfo
}

//NewStageRow 新建一个meta 结构
func NewStageRow() *StageRow {
	return &StageRow{LocalSelfConsStageInfo: &pt.LocalSelfConsStageInfo{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存heightindex)
func (r *StageRow) CreateRow() *table.Row {
	return &table.Row{Data: &pt.LocalSelfConsStageInfo{}}
}

//SetPayload 设置数据
func (r *StageRow) SetPayload(data types.Message) error {
	if d, ok := data.(*pt.LocalSelfConsStageInfo); ok {
		r.LocalSelfConsStageInfo = d
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *StageRow) Get(key string) ([]byte, error) {
	if key == "heightindex" {
		return []byte(r.TxIndex), nil
	} else if key == "id" {
		return []byte(r.Stage.Id), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", r.Stage.Status)), nil
	}

	return nil, types.ErrNotFound
}
