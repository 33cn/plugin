// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	//ErrTSellBalanceNotEnough :
	ErrTSellBalanceNotEnough = errors.New("ErrTradeSellBalanceNotEnough")
	//ErrTSellOrderNotExist :
	ErrTSellOrderNotExist = errors.New("ErrTradeSellOrderNotExist")
	//ErrTSellOrderNotStart :
	ErrTSellOrderNotStart = errors.New("ErrTradeSellOrderNotStart")
	//ErrTSellOrderNotEnough :
	ErrTSellOrderNotEnough = errors.New("ErrTradeSellOrderNotEnough")
	//ErrTSellOrderSoldout :
	ErrTSellOrderSoldout = errors.New("ErrTradeSellOrderSoldout")
	//ErrTSellOrderRevoked :
	ErrTSellOrderRevoked = errors.New("ErrTradeSellOrderRevoked")
	//ErrTSellOrderExpired :
	ErrTSellOrderExpired = errors.New("ErrTradeSellOrderExpired")
	//ErrTSellOrderRevoke :
	ErrTSellOrderRevoke = errors.New("ErrTradeSellOrderRevokeNotAllowed")
	//ErrTSellNoSuchOrder :
	ErrTSellNoSuchOrder = errors.New("ErrTradeSellNoSuchOrder")
	//ErrTBuyOrderNotExist :
	ErrTBuyOrderNotExist = errors.New("ErrTradeBuyOrderNotExist")
	//ErrTBuyOrderNotEnough :
	ErrTBuyOrderNotEnough = errors.New("ErrTradeBuyOrderNotEnough")
	//ErrTBuyOrderSoldout :
	ErrTBuyOrderSoldout = errors.New("ErrTradeBuyOrderSoldout")
	//ErrTBuyOrderRevoked :
	ErrTBuyOrderRevoked = errors.New("ErrTradeBuyOrderRevoked")
	//ErrTBuyOrderRevoke :
	ErrTBuyOrderRevoke = errors.New("ErrTradeBuyOrderRevokeNotAllowed")
	//ErrTCntLessThanMinBoardlot :
	ErrTCntLessThanMinBoardlot = errors.New("ErrTradeCountLessThanMinBoardlot")
)
