// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

// Errors for lottery
var (
	ErrRiskParam                 = errors.New("ErrRiskParam")
	ErrIssuanceRepeatHash        = errors.New("ErrIssuanceRepeatHash")
	ErrIssuanceStatus            = errors.New("ErrIssuanceStatus")
	ErrIssuanceExceedDebtCeiling = errors.New("ErrIssuanceExceedDebtCeiling")
	ErrPriceInvalid              = errors.New("ErrPriceInvalid")
	ErrAssetType                 = errors.New("ErrAssetType")
	ErrRecordNotExist            = errors.New("ErrRecordNotExist")
	ErrIssuanceErrCloser         = errors.New("ErrIssuanceErrCloser")
	ErrRepayValueInsufficient    = errors.New("ErrRepayValueInsufficient")
	ErrIssuanceAccountExist      = errors.New("ErrIssuanceAccountExist")
	ErrIssuanceLowBalance        = errors.New("ErrIssuanceLowBalance")
	ErrIssuanceBalanceInvalid    = errors.New("ErrIssuanceBalanceInvalid")
	ErrPermissionDeny            = errors.New("ErrPermissionDeny")
	ErrIssuanceRecordNotEmpty    = errors.New("ErrIssuanceRecordNotEmpty")
)
