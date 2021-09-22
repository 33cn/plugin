package executor

import (
	"fmt"

	"github.com/33cn/chain33/system/dapp"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	mix "github.com/33cn/plugin/plugin/dapp/mix/types"
)

/*
table  struct
data:  self consens stage
index: status
*/

var txBoardOpt = &table.Option{
	Prefix:  "LODB-mix",
	Name:    "tx",
	Primary: "txIndex",
	Index:   []string{"height", "hash"},
}

//NewStageTable 新建表
func NewMixTxTable(kvdb db.KV) *table.Table {
	rowmeta := NewMixTxRow()
	table, err := table.NewTable(rowmeta, kvdb, txBoardOpt)
	if err != nil {
		panic(err)
	}
	return table
}

//MixRow table meta 结构
type MixTxRow struct {
	*mix.LocalMixTx
}

//NewMixTxRow 新建一个meta 结构
func NewMixTxRow() *MixTxRow {
	return &MixTxRow{LocalMixTx: &mix.LocalMixTx{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存heightindex)
func (r *MixTxRow) CreateRow() *table.Row {
	return &table.Row{Data: &mix.LocalMixTx{}}
}

//SetPayload 设置数据
func (r *MixTxRow) SetPayload(data types.Message) error {
	if d, ok := data.(*mix.LocalMixTx); ok {
		r.LocalMixTx = d
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *MixTxRow) Get(key string) ([]byte, error) {
	switch key {
	case "txIndex":
		return []byte(dapp.HeightIndexStr(r.Height, r.Index)), nil
	case "height":
		return []byte(fmt.Sprintf("%022d", r.Height)), nil
	case "hash":
		return []byte(r.Hash), nil

	default:
		return nil, types.ErrNotFound
	}

}
