// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrRelayOrderNotExist relay order not exist
	ErrRelayOrderNotExist = errors.New("ErrRelayOrderNotExist")
	// ErrRelayOrderOnSell relay order  on sell status
	ErrRelayOrderOnSell = errors.New("ErrRelayOrderOnSell")
	// ErrRelayOrderStatusErr relay order status err
	ErrRelayOrderStatusErr = errors.New("ErrRelayOrderStatusErr")
	// ErrRelayOrderParamErr relay order parameter err
	ErrRelayOrderParamErr = errors.New("ErrRelayOrderParamErr")
	// ErrRelayOrderSoldout order has been sold
	ErrRelayOrderSoldout = errors.New("ErrRelayOrderSoldout")
	// ErrRelayOrderRevoked order revoked
	ErrRelayOrderRevoked = errors.New("ErrRelayOrderRevoked")
	// ErrRelayOrderConfirming order is confirming, not time out
	ErrRelayOrderConfirming = errors.New("ErrRelayOrderConfirming")
	// ErrRelayOrderFinished order has finished
	ErrRelayOrderFinished = errors.New("ErrRelayOrderFinished")
	// ErrRelayReturnAddr relay order return addr error
	ErrRelayReturnAddr = errors.New("ErrRelayReturnAddr")
	// ErrRelayVerify order is verifying
	ErrRelayVerify = errors.New("ErrRelayVerify")
	// ErrRelayVerifyAddrNotFound order verify addr not found
	ErrRelayVerifyAddrNotFound = errors.New("ErrRelayVerifyAddrNotFound")
	// ErrRelayWaitBlocksErr order wait block not enough
	ErrRelayWaitBlocksErr = errors.New("ErrRelayWaitBlocks")
	// ErrRelayCoinTxHashUsed order confirm tx has been used
	ErrRelayCoinTxHashUsed = errors.New("ErrRelayCoinTxHashUsed")
	// ErrRelayBtcTxTimeErr btc tx time not reasonable
	ErrRelayBtcTxTimeErr = errors.New("ErrRelayBtcTxTimeErr")
	// ErrRelayBtcHeadSequenceErr btc header sequence not continuous
	ErrRelayBtcHeadSequenceErr = errors.New("ErrRelayBtcHeadSequenceErr")
	// ErrRelayBtcHeadHashErr btc header hash not correct
	ErrRelayBtcHeadHashErr = errors.New("ErrRelayBtcHeadHashErr")
	// ErrRelayBtcHeadBitsErr rcv btc header bit not correct
	ErrRelayBtcHeadBitsErr = errors.New("ErrRelayBtcHeadBitsErr")
	// ErrRelayBtcHeadNewBitsErr calc btc header new bits error
	ErrRelayBtcHeadNewBitsErr = errors.New("ErrRelayBtcHeadNewBitsErr")
)
