// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

//multisig 合约的错误码
var (
	ErrRequiredweight       = errors.New("ErrRequiredweight")
	ErrCreatAccountAddr     = errors.New("ErrCreatAccountAddr")
	ErrOwnerExist           = errors.New("ErrOwnerExist")
	ErrOwnerNotExist        = errors.New("ErrOwnerNotExist")
	ErrTotalWeightNotEnough = errors.New("ErrTotalWeightNotEnough")
	ErrIsNotOwner           = errors.New("ErrIsNotOwner")
	ErrDailyLimitIsZero     = errors.New("ErrDailyLimitIsZero")
	ErrInvalidTxid          = errors.New("ErrInvalidTxid")
	ErrTxidNotExist         = errors.New("ErrTxidNotExist")
	ErrTxHasExecuted        = errors.New("ErrTxHasExecuted")
	ErrDupConfirmed         = errors.New("ErrDupConfirmed")
	ErrConfirmNotExist      = errors.New("ErrConfirmNotExist")
	ErrExecerHashNoMatch    = errors.New("ErrExecerHashNoMatch")
	ErrPayLoadTypeNoMatch   = errors.New("ErrPayLoadTypeNoMatch")
	ErrTxHashNoMatch        = errors.New("ErrTxHashNoMatch")
	ErrAccCountNoMatch      = errors.New("ErrAccCountNoMatch")
	ErrAccountHasExist      = errors.New("ErrAccountHasExist")
	ErrOwnerNoMatch         = errors.New("ErrOwnerNoMatch")
	ErrDailyLimitNoMatch    = errors.New("ErrDailyLimitNoMatch")
	ErrExecutedNoMatch      = errors.New("ErrExecutedNoMatch")
	ErrActionTyNoMatch      = errors.New("ErrActionTyNoMatch")
	ErrTxTypeNoMatch        = errors.New("ErrTxTypeNoMatch")
	ErrTxidHasExist         = errors.New("ErrTxidHasExist")
	ErrOnlyOneOwner         = errors.New("ErrOnlyOneOwner")
	ErrOperateType          = errors.New("ErrOperateType")
	ErrNewOwnerExist        = errors.New("ErrNewOwnerExist")
	ErrOwnerLessThanTwo     = errors.New("ErrOwnerLessThanTwo")
	ErrAddrNotSupport       = errors.New("ErrAddrNotSupport")
	ErrMaxOwnerCount        = errors.New("ErrMaxOwnerCount")
	ErrInvalidSymbol        = errors.New("ErrInvalidSymbol")
	ErrInvalidExec          = errors.New("ErrInvalidExec")
	ErrInvalidWeight        = errors.New("ErrInvalidWeight")
	ErrInvalidDailyLimit    = errors.New("ErrInvalidDailyLimit")
)
