// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	//TradeX :
	TradeX = "trade"
	tlog   = log.New("module", TradeX)

	actionName = map[string]int32{
		"SellLimit":  TradeSellLimit,
		"BuyMarket":  TradeBuyMarket,
		"RevokeSell": TradeRevokeSell,
		"BuyLimit":   TradeBuyLimit,
		"SellMarket": TradeSellMarket,
		"RevokeBuy":  TradeRevokeBuy,
	}

	logInfo = map[int64]*types.LogInfo{
		TyLogTradeSellLimit:  {Ty: reflect.TypeOf(ReceiptTradeSellLimit{}), Name: "LogTradeSell"},
		TyLogTradeBuyMarket:  {Ty: reflect.TypeOf(ReceiptTradeBuyMarket{}), Name: "LogTradeBuyMarket"},
		TyLogTradeSellRevoke: {Ty: reflect.TypeOf(ReceiptTradeSellRevoke{}), Name: "LogTradeSellRevoke"},
		TyLogTradeSellMarket: {Ty: reflect.TypeOf(ReceiptSellMarket{}), Name: "LogTradeSellMarket"},
		TyLogTradeBuyLimit:   {Ty: reflect.TypeOf(ReceiptTradeBuyLimit{}), Name: "LogTradeBuyLimit"},
		TyLogTradeBuyRevoke:  {Ty: reflect.TypeOf(ReceiptTradeBuyRevoke{}), Name: "LogTradeBuyRevoke"},
	}
)

func (t *tradeType) GetName() string {
	return TradeX
}

func (t *tradeType) GetTypeMap() map[string]int32 {
	return actionName
}

func (t *tradeType) GetLogMap() map[int64]*types.LogInfo {
	return logInfo
}

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(TradeX))
	types.RegistorExecutor(TradeX, newType())
	types.RegisterDappFork(TradeX, "Enable", 100899)
	types.RegisterDappFork(TradeX, ForkTradeBuyLimitX, 301000)
	types.RegisterDappFork(TradeX, ForkTradeAssetX, 1010000)
	types.RegisterDappFork(TradeX, ForkTradeIDX, 1450000)
}

type tradeType struct {
	types.ExecTypeBase
}

func newType() *tradeType {
	c := &tradeType{}
	c.SetChild(c)
	return c
}

func (t *tradeType) GetPayload() types.Message {
	return &Trade{}
}

//ActionName :
func (t *tradeType) ActionName(tx *types.Transaction) string {
	var action Trade
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return "unknown-err"
	}
	if action.Ty == TradeSellLimit && action.GetSellLimit() != nil {
		return "selltoken"
	} else if action.Ty == TradeBuyMarket && action.GetBuyMarket() != nil {
		return "buytoken"
	} else if action.Ty == TradeRevokeSell && action.GetRevokeSell() != nil {
		return "revokeselltoken"
	} else if action.Ty == TradeBuyLimit && action.GetBuyLimit() != nil {
		return "buylimittoken"
	} else if action.Ty == TradeSellMarket && action.GetSellMarket() != nil {
		return "sellmarkettoken"
	} else if action.Ty == TradeRevokeBuy && action.GetRevokeBuy() != nil {
		return "revokebuytoken"
	}
	return "unknown"
}

func (t *tradeType) Amount(tx *types.Transaction) (int64, error) {
	//TODO: 补充和完善token和trade分支的amount的计算, added by hzj
	var trade Trade
	err := types.Decode(tx.GetPayload(), &trade)
	if err != nil {
		return 0, types.ErrDecode
	}

	if TradeSellLimit == trade.Ty && trade.GetSellLimit() != nil {
		return 0, nil
	} else if TradeBuyMarket == trade.Ty && trade.GetBuyMarket() != nil {
		return 0, nil
	} else if TradeRevokeSell == trade.Ty && trade.GetRevokeSell() != nil {
		return 0, nil
	}
	return 0, nil
}

func (t *tradeType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	//var tx *types.Transaction
	if action == "TradeSellLimit" {
		var param TradeSellTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawTradeSellTx(&param)
	} else if action == "TradeBuyMarket" {
		var param TradeBuyTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawTradeBuyTx(&param)
	} else if action == "TradeSellRevoke" {
		var param TradeRevokeTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawTradeRevokeTx(&param)
	} else if action == "TradeBuyLimit" {
		var param TradeBuyLimitTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawTradeBuyLimitTx(&param)
	} else if action == "TradeSellMarket" {
		var param TradeSellMarketTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawTradeSellMarketTx(&param)
	} else if action == "TradeRevokeBuy" {
		var param TradeRevokeBuyTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawTradeRevokeBuyTx(&param)
	}

	return nil, types.ErrNotSupport
}

//CreateRawTradeSellTx : 创建卖单交易
func CreateRawTradeSellTx(parm *TradeSellTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := &TradeForSell{
		TokenSymbol:       parm.TokenSymbol,
		AmountPerBoardlot: parm.AmountPerBoardlot,
		MinBoardlot:       parm.MinBoardlot,
		PricePerBoardlot:  parm.PricePerBoardlot,
		TotalBoardlot:     parm.TotalBoardlot,
		Starttime:         0,
		Stoptime:          0,
		Crowdfund:         false,
		AssetExec:         parm.AssetExec,
	}
	sell := &Trade{
		Ty:    TradeSellLimit,
		Value: &Trade_SellLimit{v},
	}
	return types.CreateFormatTx(types.ExecName(TradeX), types.Encode(sell))
}

//CreateRawTradeBuyTx :创建想指定卖单发起的买单交易
func CreateRawTradeBuyTx(parm *TradeBuyTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := &TradeForBuy{SellID: parm.SellID, BoardlotCnt: parm.BoardlotCnt}
	buy := &Trade{
		Ty:    TradeBuyMarket,
		Value: &Trade_BuyMarket{v},
	}
	return types.CreateFormatTx(types.ExecName(TradeX), types.Encode(buy))
}

//CreateRawTradeRevokeTx :创建取消卖单的交易
func CreateRawTradeRevokeTx(parm *TradeRevokeTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}

	v := &TradeForRevokeSell{SellID: parm.SellID}
	buy := &Trade{
		Ty:    TradeRevokeSell,
		Value: &Trade_RevokeSell{v},
	}
	return types.CreateFormatTx(types.ExecName(TradeX), types.Encode(buy))
}

//CreateRawTradeBuyLimitTx :创建买单交易
func CreateRawTradeBuyLimitTx(parm *TradeBuyLimitTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := &TradeForBuyLimit{
		TokenSymbol:       parm.TokenSymbol,
		AmountPerBoardlot: parm.AmountPerBoardlot,
		MinBoardlot:       parm.MinBoardlot,
		PricePerBoardlot:  parm.PricePerBoardlot,
		TotalBoardlot:     parm.TotalBoardlot,
		AssetExec:         parm.AssetExec,
	}
	buyLimit := &Trade{
		Ty:    TradeBuyLimit,
		Value: &Trade_BuyLimit{v},
	}
	return types.CreateFormatTx(types.ExecName(TradeX), types.Encode(buyLimit))
}

//CreateRawTradeSellMarketTx : 创建向指定买单出售token的卖单交易
func CreateRawTradeSellMarketTx(parm *TradeSellMarketTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := &TradeForSellMarket{BuyID: parm.BuyID, BoardlotCnt: parm.BoardlotCnt}
	sellMarket := &Trade{
		Ty:    TradeSellMarket,
		Value: &Trade_SellMarket{v},
	}
	return types.CreateFormatTx(types.ExecName(TradeX), types.Encode(sellMarket))
}

//CreateRawTradeRevokeBuyTx : 取消发起的买单交易
func CreateRawTradeRevokeBuyTx(parm *TradeRevokeBuyTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}

	v := &TradeForRevokeBuy{BuyID: parm.BuyID}
	buy := &Trade{
		Ty:    TradeRevokeBuy,
		Value: &Trade_RevokeBuy{v},
	}
	return types.CreateFormatTx(types.ExecName(TradeX), types.Encode(buy))
}
