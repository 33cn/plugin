// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// CollateralizeCreateTx for construction
type CollateralizeCreateTx struct {
	TotalBalance        int64 `json:"totalBalance"`
	Fee                 int64  `json:"fee"`
}

// CollateralizeBorrowTx for construction
type CollateralizeBorrowTx struct {
	CollateralizeID string `json:"collateralizeId"`
	Value    int64  `json:"value"`
	Fee       int64  `json:"fee"`
}

// CollateralizeRepayTx for construction
type CollateralizeRepayTx struct {
	CollateralizeID string `json:"collateralizeId"`
	RecordID string `json:"recordID"`
	Fee       int64  `json:"fee"`
}

// CollateralizeAppednTx for construction
type CollateralizeAppendTx struct {
	CollateralizeID string `json:"collateralizeId"`
	RecordID string `json:"recordID"`
	Value    int64  `json:"value"`
	Fee       int64  `json:"fee"`
}

// CollateralizeFeedTx for construction
type CollateralizeFeedTx struct {
	Price     []float32  `json:"price"`
	Volume    []int64  `json:"volume"`
	Fee       int64  `json:"fee"`
}

// CollateralizeCloseTx for construction
type CollateralizeCloseTx struct {
	CollateralizeID string `json:"collateralizeId"`
	Fee       int64  `json:"fee"`
}

// CollateralizeManageTx for construction
type CollateralizeManageTx struct {
	DebtCeiling         int64 `json:"debtCeiling"`
	LiquidationRatio    float32 `json:"liquidationRatio"`
	StabilityFeeRatio   float32 `json:"stabilityFeeRatio"`
	Period              int64 `json:"period"`
	Fee                 int64  `json:"fee"`
}
