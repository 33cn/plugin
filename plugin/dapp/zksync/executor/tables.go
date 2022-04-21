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
	KeyPrefixStateDB = "mavl-zksync-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-zksync"
)

var opt_account_tree = &table.Option{
	Prefix:  KeyPrefixLocalDB,
	Name:    "tree",
	Primary: "address",
	Index:   []string{"chain33_address", "eth_address"},
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
	if key == "address" {
		return GetLocalChain33EthPrimaryKey(r.GetChain33Addr(), r.GetEthAddress()), nil
	} else if key == "chain33_address" {
		return []byte(fmt.Sprintf("%s", r.GetChain33Addr())), nil
	} else if key == "eth_address" {
		return []byte(fmt.Sprintf("%s", r.GetEthAddress())), nil
	}
	return nil, types.ErrNotFound
}

var opt_zksync_info = &table.Option{
	Prefix:  KeyPrefixLocalDB,
	Name:    "zksync",
	Primary: "height_index_opIndex",
	Index:   []string{"height", "txHash"},
}

// NewZksyncInfoTable ...
func NewZksyncInfoTable(kvdb db.KV) *table.Table {
	rowmeta := NewZksyncInfoRow()
	table, err := table.NewTable(rowmeta, kvdb, opt_zksync_info)
	if err != nil {
		panic(err)
	}
	return table
}

// AccountTreeRow table meta 结构
type ZksyncInfoRow struct {
	*zt.OperationInfo
}

func NewZksyncInfoRow() *ZksyncInfoRow {
	return &ZksyncInfoRow{OperationInfo: &zt.OperationInfo{}}
}

//CreateRow 新建数据行
func (r *ZksyncInfoRow) CreateRow() *table.Row {
	return &table.Row{Data: &zt.OperationInfo{}}
}

//SetPayload 设置数据
func (r *ZksyncInfoRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*zt.OperationInfo); ok {
		r.OperationInfo = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *ZksyncInfoRow) Get(key string) ([]byte, error) {
	if key == "height_index_opIndex" {
		return []byte(fmt.Sprintf("%016d.%016d.%016d", r.GetBlockHeight(), r.GetTxIndex(), r.GetOpIndex())), nil
	} else if key == "height" {
		return []byte(fmt.Sprintf("%016d", r.GetBlockHeight())), nil
	} else if key == "txHash" {
		return []byte(fmt.Sprintf("%s", r.GetTxHash())), nil
	}
	return nil, types.ErrNotFound
}



var opt_commit_proof = &table.Option{
	Prefix:  KeyPrefixLocalDB,
	Name:    "proof",
	Primary: "proofId",
	Index:   []string{"height", "root"},
}

// NewCommitProofTable ...
func NewCommitProofTable(kvdb db.KV) *table.Table {
	rowmeta := NewCommitProofRow()
	table, err := table.NewTable(rowmeta, kvdb, opt_commit_proof)
	if err != nil {
		panic(err)
	}
	return table
}

// CommitProofRow table meta 结构
type CommitProofRow struct {
	*zt.ZkCommitProof
}

func NewCommitProofRow() *CommitProofRow {
	return &CommitProofRow{ZkCommitProof: &zt.ZkCommitProof{}}
}

//CreateRow 新建数据行
func (r *CommitProofRow) CreateRow() *table.Row {
	return &table.Row{Data: &zt.ZkCommitProof{}}
}

//SetPayload 设置数据
func (r *CommitProofRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*zt.ZkCommitProof); ok {
		r.ZkCommitProof = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *CommitProofRow) Get(key string) ([]byte, error) {
	if key == "proofId" {
		return []byte(fmt.Sprintf("%016d", r.GetProofId())), nil
	} else if key == "root" {
		return []byte(fmt.Sprintf("%s", r.GetNewTreeRoot())), nil
	} else if key == "height" {
		return []byte(fmt.Sprintf("%016d", r.GetBlockEnd())), nil
	}
	return nil, types.ErrNotFound
}


var opt_history_account_tree = &table.Option{
	Prefix:  KeyPrefixLocalDB,
	Name:    "historyTree",
	Primary: "proofId_accountId",
	Index:   []string{"proofId"},
}

// NewHistoryAccountTreeTable ...
func NewHistoryAccountTreeTable(kvdb db.KV) *table.Table {
	rowmeta := NewHistoryAccountTreeRow()
	table, err := table.NewTable(rowmeta, kvdb, opt_history_account_tree)
	if err != nil {
		panic(err)
	}
	return table
}

// HistoryAccountTreeRow table meta 结构
type HistoryAccountTreeRow struct {
	*zt.HistoryLeaf
}

func NewHistoryAccountTreeRow() *HistoryAccountTreeRow {
	return &HistoryAccountTreeRow{HistoryLeaf: &zt.HistoryLeaf{}}
}

//CreateRow 新建数据行
func (r *HistoryAccountTreeRow) CreateRow() *table.Row {
	return &table.Row{Data: &zt.HistoryLeaf{}}
}

//SetPayload 设置数据
func (r *HistoryAccountTreeRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*zt.HistoryLeaf); ok {
		r.HistoryLeaf = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (r *HistoryAccountTreeRow) Get(key string) ([]byte, error) {
	if key == "proofId_accountId" {
		return []byte(fmt.Sprintf("%016d.%16d", r.GetProofId(), r.GetAccountId())), nil
	} else if key == "proofId" {
		return []byte(fmt.Sprintf("%016d", r.GetProofId())), nil
	}
	return nil, types.ErrNotFound
}


