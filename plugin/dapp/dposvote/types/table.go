package types

import (
	"fmt"

	"github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
)

/*
table  struct
data:  voter
index: FromAddr,Pubkey,Votes,Index,Time
*/

var optDposVoter = &table.Option{
	Prefix:  "LODB-dpos",
	Name:    "voter",
	Primary: "index",
	Index:   []string{"addr", "pubkey"},
}

//NewDposVoteTable 新建表
func NewDposVoteTable(kvdb db.KV) *table.Table {
	rowmeta := NewDposVoterRow()
	table, err := table.NewTable(rowmeta, kvdb, optDposVoter)
	if err != nil {
		panic(err)
	}
	return table
}

//DposVoterRow table meta 结构
type DposVoterRow struct {
	*DposVoter
}

//NewDposVoterRow 新建一个meta 结构
func NewDposVoterRow() *DposVoterRow {
	return &DposVoterRow{DposVoter: &DposVoter{}}
}

//CreateRow 新建数据行
func (tx *DposVoterRow) CreateRow() *table.Row {
	return &table.Row{Data: &DposVoter{}}
}

//SetPayload 设置数据
func (tx *DposVoterRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*DposVoter); ok {
		tx.DposVoter = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *DposVoterRow) Get(key string) ([]byte, error) {
	if key == "index" {
		return []byte(fmt.Sprintf("%018d", tx.Index)), nil
	} else if key == "addr" {
		return []byte(tx.FromAddr), nil
	} else if key == "pubkey" {
		return tx.Pubkey, nil
	}

	return nil, types.ErrNotFound
}

var optDposCandidator = &table.Option{
	Prefix:  "LODB-dpos",
	Name:    "candidator",
	Primary: "pubkey",
	Index:   []string{"status"},
}

//NewDposCandidatorTable 新建表
func NewDposCandidatorTable(kvdb db.KV) *table.Table {
	rowmeta := NewDposCandidatorRow()
	table, err := table.NewTable(rowmeta, kvdb, optDposCandidator)
	if err != nil {
		panic(err)
	}
	return table
}

//DposCandidatorRow table meta 结构
type DposCandidatorRow struct {
	*CandidatorInfo
}

//NewDposCandidatorRow 新建一个meta 结构
func NewDposCandidatorRow() *DposCandidatorRow {
	return &DposCandidatorRow{CandidatorInfo: &CandidatorInfo{}}
}

//CreateRow 新建数据行
func (tx *DposCandidatorRow) CreateRow() *table.Row {
	return &table.Row{Data: &CandidatorInfo{}}
}

//SetPayload 设置数据
func (tx *DposCandidatorRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*CandidatorInfo); ok {
		tx.CandidatorInfo = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *DposCandidatorRow) Get(key string) ([]byte, error) {
	if key == "pubkey" {
		return tx.Pubkey, nil
	} else if key == "status" {
		return []byte(fmt.Sprintf("%2d", tx.Status)), nil
	}

	return nil, types.ErrNotFound
}

var optDposVrfm = &table.Option{
	Prefix:  "LODB-dpos",
	Name:    "vrfm",
	Primary: "index",
	Index:   []string{"pubkey_cycle", "cycle"},
}

//NewDposVrfMTable 新建表
func NewDposVrfMTable(kvdb db.KV) *table.Table {
	rowmeta := NewDposVrfMRow()
	table, err := table.NewTable(rowmeta, kvdb, optDposVrfm)
	if err != nil {
		panic(err)
	}
	return table
}

//DposVrfMRow table meta 结构
type DposVrfMRow struct {
	*DposVrfM
}

//NewDposVrfMRow 新建一个meta 结构
func NewDposVrfMRow() *DposVrfMRow {
	return &DposVrfMRow{DposVrfM: &DposVrfM{}}
}

//CreateRow 新建数据行
func (tx *DposVrfMRow) CreateRow() *table.Row {
	return &table.Row{Data: &DposVrfM{}}
}

//SetPayload 设置数据
func (tx *DposVrfMRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*DposVrfM); ok {
		tx.DposVrfM = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *DposVrfMRow) Get(key string) ([]byte, error) {
	if key == "index" {
		return []byte(fmt.Sprintf("%018d", tx.Index)), nil
	} else if key == "pubkey_cycle" {
		return []byte(fmt.Sprintf("%X:%018d", tx.Pubkey, tx.Cycle)), nil
	} else if key == "cycle" {
		return []byte(fmt.Sprintf("%018d", tx.Cycle)), nil
	}

	return nil, types.ErrNotFound
}

var optDposVrfrp = &table.Option{
	Prefix:  "LODB-dpos",
	Name:    "vrfrp",
	Primary: "index",
	Index:   []string{"pubkey_cycle", "cycle"},
}

//NewDposVrfRPTable 新建表
func NewDposVrfRPTable(kvdb db.KV) *table.Table {
	rowmeta := NewDposVrfRPRow()
	table, err := table.NewTable(rowmeta, kvdb, optDposVrfrp)
	if err != nil {
		panic(err)
	}
	return table
}

//DposVrfRPRow table meta 结构
type DposVrfRPRow struct {
	*DposVrfRP
}

//NewDposVrfRPRow 新建一个meta 结构
func NewDposVrfRPRow() *DposVrfRPRow {
	return &DposVrfRPRow{DposVrfRP: &DposVrfRP{}}
}

//CreateRow 新建数据行
func (tx *DposVrfRPRow) CreateRow() *table.Row {
	return &table.Row{Data: &DposVrfRP{}}
}

//SetPayload 设置数据
func (tx *DposVrfRPRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*DposVrfRP); ok {
		tx.DposVrfRP = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *DposVrfRPRow) Get(key string) ([]byte, error) {
	if key == "index" {
		return []byte(fmt.Sprintf("%018d", tx.Index)), nil
	} else if key == "pubkey_cycle" {
		return []byte(fmt.Sprintf("%X:%018d", tx.Pubkey, tx.Cycle)), nil
	} else if key == "cycle" {
		return []byte(fmt.Sprintf("%018d", tx.Cycle)), nil
	}

	return nil, types.ErrNotFound
}

var optDposCb = &table.Option{
	Prefix:  "LODB-dpos",
	Name:    "cb",
	Primary: "cycle",
	Index:   []string{"height", "hash"},
}

//NewDposCBTable 新建表
func NewDposCBTable(kvdb db.KV) *table.Table {
	rowmeta := NewDposCBRow()
	table, err := table.NewTable(rowmeta, kvdb, optDposCb)
	if err != nil {
		panic(err)
	}
	return table
}

//DposCBRow table meta 结构
type DposCBRow struct {
	*DposCycleBoundaryInfo
}

//NewDposCBRow 新建一个meta 结构
func NewDposCBRow() *DposCBRow {
	return &DposCBRow{DposCycleBoundaryInfo: &DposCycleBoundaryInfo{}}
}

//CreateRow 新建数据行
func (tx *DposCBRow) CreateRow() *table.Row {
	return &table.Row{Data: &DposCycleBoundaryInfo{}}
}

//SetPayload 设置数据
func (tx *DposCBRow) SetPayload(data types.Message) error {
	if txdata, ok := data.(*DposCycleBoundaryInfo); ok {
		tx.DposCycleBoundaryInfo = txdata
		return nil
	}
	return types.ErrTypeAsset
}

//Get 按照indexName 查询 indexValue
func (tx *DposCBRow) Get(key string) ([]byte, error) {
	if key == "cycle" {
		return []byte(fmt.Sprintf("%018d", tx.Cycle)), nil
	} else if key == "height" {
		return []byte(fmt.Sprintf("%018d", tx.StopHeight)), nil
	} else if key == "hash" {
		return tx.StopHash, nil
	}

	return nil, types.ErrNotFound
}
