// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

//----------------------------- data for testing ---------------------------------

var (
	sellorderOnsale = pty.SellOrder{
		TokenSymbol:       "Tokensymbol",
		Address:           "Address",
		AmountPerBoardlot: 20 * 1e8,                   // Amountperboardlot int64  `protobuf:"varint,3,opt,name=amountperboardlot" json:"amountperboardlot,omitempty"`
		MinBoardlot:       2,                          // Minboardlot       int64  `protobuf:"varint,4,opt,name=minboardlot" json:"minboardlot,omitempty"`
		PricePerBoardlot:  1 * 1e8,                    //Priceperboardlot  int64  `protobuf:"varint,5,opt,name=priceperboardlot" json:"priceperboardlot,omitempty"`
		TotalBoardlot:     60,                         // Totalboardlot     int64  `protobuf:"varint,6,opt,name=totalboardlot" json:"totalboardlot,omitempty"`
		SoldBoardlot:      2,                          // Soldboardlot      int64  `protobuf:"varint,7,opt,name=soldboardlot" json:"soldboardlot,omitempty"`
		Starttime:         0,                          //Starttime         int64  `protobuf:"varint,8,opt,name=starttime" json:"starttime,omitempty"`
		Stoptime:          0,                          //Stoptime          int64  `protobuf:"varint,9,opt,name=stoptime" json:"stoptime,omitempty"`
		Crowdfund:         false,                      //Crowdfund         bool   `protobuf:"varint,10,opt,name=crowdfund" json:"crowdfund,omitempty"`
		SellID:            "IAMSELLID",                // Sellid            string `protobuf:"bytes,11,opt,name=sellid" json:"sellid,omitempty"`
		Status:            pty.TradeOrderStatusOnSale, //Status            int32  `protobuf:"varint,12,opt,name=status" json:"status,omitempty"`
		Height:            100,                        //Height            int64  `protobuf:"varint,13,opt,name=height" json:"height,omitempty"`
		AssetExec:         "token",
	}

	sellorderSoldOut = pty.SellOrder{
		TokenSymbol:       "Tokensymbol",
		Address:           "Address",
		AmountPerBoardlot: 20 * 1e8,                    // Amountperboardlot int64  `protobuf:"varint,3,opt,name=amountperboardlot" json:"amountperboardlot,omitempty"`
		MinBoardlot:       2,                           // Minboardlot       int64  `protobuf:"varint,4,opt,name=minboardlot" json:"minboardlot,omitempty"`
		PricePerBoardlot:  1 * 1e8,                     //Priceperboardlot  int64  `protobuf:"varint,5,opt,name=priceperboardlot" json:"priceperboardlot,omitempty"`
		TotalBoardlot:     60,                          // Totalboardlot     int64  `protobuf:"varint,6,opt,name=totalboardlot" json:"totalboardlot,omitempty"`
		SoldBoardlot:      2,                           // Soldboardlot      int64  `protobuf:"varint,7,opt,name=soldboardlot" json:"soldboardlot,omitempty"`
		Starttime:         0,                           //Starttime         int64  `protobuf:"varint,8,opt,name=starttime" json:"starttime,omitempty"`
		Stoptime:          0,                           //Stoptime          int64  `protobuf:"varint,9,opt,name=stoptime" json:"stoptime,omitempty"`
		Crowdfund:         false,                       //Crowdfund         bool   `protobuf:"varint,10,opt,name=crowdfund" json:"crowdfund,omitempty"`
		SellID:            "IAMSELLID",                 // Sellid            string `protobuf:"bytes,11,opt,name=sellid" json:"sellid,omitempty"`
		Status:            pty.TradeOrderStatusSoldOut, //Status            int32  `protobuf:"varint,12,opt,name=status" json:"status,omitempty"`
		Height:            100,                         //Height            int64  `protobuf:"varint,13,opt,name=height" json:"height,omitempty"`
		AssetExec:         "token",
	}

	sellorderRevoked = pty.SellOrder{
		TokenSymbol:       "Tokensymbol",
		Address:           "Address",
		AmountPerBoardlot: 20 * 1e8,                    // Amountperboardlot int64  `protobuf:"varint,3,opt,name=amountperboardlot" json:"amountperboardlot,omitempty"`
		MinBoardlot:       2,                           // Minboardlot       int64  `protobuf:"varint,4,opt,name=minboardlot" json:"minboardlot,omitempty"`
		PricePerBoardlot:  1 * 1e8,                     //Priceperboardlot  int64  `protobuf:"varint,5,opt,name=priceperboardlot" json:"priceperboardlot,omitempty"`
		TotalBoardlot:     60,                          // Totalboardlot     int64  `protobuf:"varint,6,opt,name=totalboardlot" json:"totalboardlot,omitempty"`
		SoldBoardlot:      2,                           // Soldboardlot      int64  `protobuf:"varint,7,opt,name=soldboardlot" json:"soldboardlot,omitempty"`
		Starttime:         0,                           //Starttime         int64  `protobuf:"varint,8,opt,name=starttime" json:"starttime,omitempty"`
		Stoptime:          0,                           //Stoptime          int64  `protobuf:"varint,9,opt,name=stoptime" json:"stoptime,omitempty"`
		Crowdfund:         false,                       //Crowdfund         bool   `protobuf:"varint,10,opt,name=crowdfund" json:"crowdfund,omitempty"`
		SellID:            "IAMSELLID",                 // Sellid            string `protobuf:"bytes,11,opt,name=sellid" json:"sellid,omitempty"`
		Status:            pty.TradeOrderStatusRevoked, //Status            int32  `protobuf:"varint,12,opt,name=status" json:"status,omitempty"`
		Height:            100,                         //Height            int64  `protobuf:"varint,13,opt,name=height" json:"height,omitempty"`
		AssetExec:         "token",
	}
)

// TODO 几个测试数据 linter 不报错, 修改好后写测试可能需要用
func Test_Order(t *testing.T) {
	assert.NotNil(t, &sellorderOnsale)
	assert.NotNil(t, &sellorderSoldOut)
	assert.NotNil(t, &sellorderRevoked)
}

func TestPriceCheck(t *testing.T) {
	cases := []struct {
		exec   string
		symbol string
		result bool
	}{
		{"coins", "bty", true},
		{"", "bty", false},
		{"coins", "", false},
		{"token", "TEST", true},
	}

	for _, c := range cases {
		assert.Equal(t, c.result, checkPrice(chain33TestCfg, chain33TestCfg.GetDappFork(pty.TradeX, pty.ForkTradePriceX), c.exec, c.symbol))
	}
}
