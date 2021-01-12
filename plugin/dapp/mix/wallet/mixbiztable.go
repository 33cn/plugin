package wallet

import (
	"fmt"

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

var boardOpt = &table.Option{
	Prefix:  "LODB-mix",
	Name:    "wallet",
	Primary: "heightindex",
	Index: []string{
		"noteHash",
		"nullifier",
		"authSpendHash",
		"spender",
		"account",
		"status"},
}

//NewStageTable 新建表
func NewMixTable(kvdb db.KV) *table.Table {
	rowmeta := NewMixRow()
	table, err := table.NewTable(rowmeta, kvdb, boardOpt)
	if err != nil {
		panic(err)
	}
	return table
}

//MixRow table meta 结构
type MixRow struct {
	*mix.WalletDbMixInfo
}

//NewStageRow 新建一个meta 结构
func NewMixRow() *MixRow {
	return &MixRow{WalletDbMixInfo: &mix.WalletDbMixInfo{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存heightindex)
func (r *MixRow) CreateRow() *table.Row {
	return &table.Row{Data: &mix.WalletDbMixInfo{}}
}

//SetPayload 设置数据
func (r *MixRow) SetPayload(data types.Message) error {
	if d, ok := data.(*mix.WalletDbMixInfo); ok {
		r.WalletDbMixInfo = d
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *MixRow) Get(key string) ([]byte, error) {
	switch key {
	case "heightindex":
		return []byte(r.TxIndex), nil
	case "noteHash":
		return []byte(r.Info.NoteHash), nil
	case "nullifier":
		return []byte(r.Info.Nullifier), nil
	case "authSpendHash":
		return []byte(r.Info.AuthSpendHash), nil
	case "spender":
		return []byte(r.Info.Spender), nil
	case "account":
		return []byte(r.Info.Account), nil
	case "status":
		return []byte(fmt.Sprintf("%2d", r.Info.Status)), nil
	default:
		return nil, types.ErrNotFound
	}

}
