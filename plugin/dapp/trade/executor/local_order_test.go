package executor

import (
	"testing"

	"github.com/33cn/chain33/system/dapp"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
	//"github.com/33cn/chain33/common/db"
	//"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/assert"
)

var order1 = &pty.LocalOrder{
	AssetSymbol:       "A",
	Owner:             "O1",
	AmountPerBoardlot: 1,
	MinBoardlot:       1,
	PricePerBoardlot:  1,
	TotalBoardlot:     10,
	TradedBoardlot:    0,
	BuyID:             "B1",
	Status:            pty.TradeOrderStatusOnBuy,
	SellID:            "",
	TxHash:            nil,
	Height:            1,
	Key:               "B1",
	BlockTime:         1,
	IsSellOrder:       false,
	AssetExec:         "a",
	TxIndex:           dapp.HeightIndexStr(1, 1),
	IsFinished:        false,
}

var order2 = &pty.LocalOrder{
	AssetSymbol:       "A",
	Owner:             "O1",
	AmountPerBoardlot: 1,
	MinBoardlot:       1,
	PricePerBoardlot:  1,
	TotalBoardlot:     10,
	TradedBoardlot:    0,
	BuyID:             "B2",
	Status:            pty.TradeOrderStatusOnBuy,
	SellID:            "",
	TxHash:            nil,
	Height:            2,
	Key:               "B2",
	BlockTime:         2,
	IsSellOrder:       false,
	AssetExec:         "a",
	TxIndex:           dapp.HeightIndexStr(2, 1),
	IsFinished:        false,
}

func TestListAll(t *testing.T) {
	dir, ldb, tdb := util.CreateTestDB()
	t.Log(dir, ldb, tdb)
	odb := NewOrderTable(tdb)
	odb.Add(order1)
	odb.Add(order2)
	kvs, err := odb.Save()
	assert.Nil(t, err)
	t.Log(kvs)
	ldb.Close()
}
