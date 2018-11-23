// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/types"
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

func init() {

}

// 分叉不好构造， 直接生成对应的kv 记录进行对比
// 在save时有值的， 需要在del 时设置为空；或save 时设置为空， 在del 时有值的
func check(t *testing.T, kvSave, kvDel []*types.KeyValue) {
	kvmapSave := map[string]string{}
	for _, kv := range kvSave {
		if string(kv.Value) != "IAMSELLID" && kv.Value != nil {
			t.Error("onsale error")
		}
		kvmapSave[string(kv.Key)] = string(kv.Value)
	}

	for _, kv := range kvDel {
		v, ok := kvmapSave[string(kv.Key)]
		if !ok {
			t.Error("error 1")
		}
		if len(v) == 0 && len(kv.Value) == 0 {
			t.Error("error 2")
		}
		if len(v) != 0 && len(kv.Value) != 0 {
			t.Error("error 3")
		}
	}
}

func TestOnsaleSaveDel(t *testing.T) {
	kvOnsale := genSaveSellKv(&sellorderOnsale)
	kvOnsaleDel := genDeleteSellKv(&sellorderOnsale)
	check(t, kvOnsale, kvOnsaleDel)
}

func TestSoldOutSaveDel(t *testing.T) {
	kv := genSaveSellKv(&sellorderSoldOut)
	kvDel := genDeleteSellKv(&sellorderSoldOut)
	check(t, kv, kvDel)
}

func TestRevokeSaveDel(t *testing.T) {
	kv := genSaveSellKv(&sellorderRevoked)
	kvDel := genDeleteSellKv(&sellorderRevoked)
	check(t, kv, kvDel)
}
