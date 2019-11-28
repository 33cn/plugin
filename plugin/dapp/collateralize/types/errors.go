// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

// Errors for lottery
var (
	ErrRiskParam                      = errors.New("ErrRiskParam")
	ErrCollateralizeStatus            = errors.New("ErrCollateralizeStatus")
	ErrCollateralizeExceedDebtCeiling = errors.New("ErrCollateralizeExceedDebtCeiling")
	ErrPriceInvalid                   = errors.New("ErrPriceInvalid")
	ErrAssetType                      = errors.New("ErrAssetType")
	ErrRecordNotExist                 = errors.New("ErrRecordNotExist")
	ErrCollateralizeErrCloser         = errors.New("ErrCollateralizeErrCloser")
	ErrRepayValueInsufficient         = errors.New("ErrRepayValueInsufficient")
	ErrCollateralizeAccountExist      = errors.New("ErrCollateralizeAccountExist")
	ErrCollateralizeLowBalance        = errors.New("ErrCollateralizeLowBalance")
	ErrCollateralizeBalanceInvalid    = errors.New("ErrCollateralizeBalanceInvalid")
	ErrPermissionDeny                 = errors.New("ErrPermissionDeny")
	ErrCollateralizeRecordNotEmpty    = errors.New("ErrCollateralizeRecordNotEmpty")
)
