package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	aty "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

const (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-accountmanager-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-accountmanager"
)

var opt_account = &table.Option{
	Prefix:  KeyPrefixLocalDB,
	Name:    "account",
	Primary: "index",
	Index:   []string{"status", "accountID", "addr"},
}

//状态数据库中存储具体账户信息
func calcAccountKey(accountID string) []byte {
	key := fmt.Sprintf("%s"+"accountID:%s", KeyPrefixStateDB, accountID)
	return []byte(key)
}

//NewAccountTable ...
func NewAccountTable(kvdb db.KV) *table.Table {
	rowmeta := NewAccountRow()
	table, err := table.NewTable(rowmeta, kvdb, opt_account)
	if err != nil {
		panic(err)
	}
	return table
}

//AccountRow account table meta 结构
type AccountRow struct {
	*aty.Account
}

//NewAccountRow 新建一个meta 结构
func NewAccountRow() *AccountRow {
	return &AccountRow{Account: &aty.Account{}}
}

//CreateRow 新建数据行(注意index 数据一定也要保存到数据中,不能就保存eventid)
func (m *AccountRow) CreateRow() *table.Row {
	return &table.Row{Data: &aty.Account{}}
}

//SetPayload 设置数据
func (m *AccountRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*aty.Account); ok {
		m.Account = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (m *AccountRow) Get(key string) ([]byte, error) {
	if key == "accountID" {
		return []byte(m.AccountID), nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%d", m.Status)), nil
	} else if key == "index" {
		return []byte(fmt.Sprintf("%015d", m.GetIndex())), nil
	} else if key == "addr" {
		return []byte(m.GetAddr()), nil
	}
	return nil, types.ErrNotFound
}
