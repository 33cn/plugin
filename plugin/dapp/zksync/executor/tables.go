package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

const (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-exchange-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-exchange"
)

var opt_account_tree = &table.Option{
	Prefix:  KeyPrefixLocalDB,
	Name:    "tree",
	Primary: "accountId",
	Index:   []string{"root_hash", "eth_address"},
}

// NewAccountTreeTable ...
func NewAccountTreeTable(kvdb db.KV) *table.Table {
	rowmeta := NewAccountTreeRow()
	table, err := table.NewTable(rowmeta, kvdb, opt_account_tree)
	if err != nil {
		panic(err)
	}
	return table
}

// AccountTreeRow table meta 结构
type AccountTreeRow struct {
	*zt.Leaf
}

func NewAccountTreeRow() *AccountTreeRow {
	return &AccountTreeRow{Leaf: &zt.Leaf{}}
}

//CreateRow 新建数据行
func (r *AccountTreeRow) CreateRow() *table.Row {
	return &table.Row{Data: &zt.Leaf{}}
}

//SetPayload 设置数据
func (r *AccountTreeRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*zt.Leaf); ok {
		r.Leaf = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *AccountTreeRow) Get(key string) ([]byte, error) {
	if key == "accountId" {
		return GetAccountIdPrimaryKey(r.GetAccountId()), nil
	} else if key == "root_hash" {
		return []byte(fmt.Sprintf("%s", r.GetRootHash())), nil
	} else if key == "eth_address" {
		return []byte(fmt.Sprintf("%s", r.GetEthAddress())), nil
	}
	return nil, types.ErrNotFound
}

func GetAccountIdPrimaryKey(accountId int32) []byte {
	return []byte(fmt.Sprintf("%022d", accountId))
}
