// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// CollateralizeCreateTx for construction
type CollateralizeCreateTx struct {
	TotalBalance float64 `json:"totalBalance"`
	Fee          int64   `json:"fee"`
}

// CollateralizeBorrowTx for construction
type CollateralizeBorrowTx struct {
	CollateralizeID string  `json:"collateralizeId"`
	Value           float64 `json:"value"`
	Fee             int64   `json:"fee"`
}

// CollateralizeRepayTx for construction
type CollateralizeRepayTx struct {
	CollateralizeID string `json:"collateralizeId"`
	RecordID        string `json:"recordID"`
	Fee             int64  `json:"fee"`
}

// CollateralizeAppednTx for construction
type CollateralizeAppendTx struct {
	CollateralizeID string  `json:"collateralizeId"`
	RecordID        string  `json:"recordID"`
	Value           float64 `json:"value"`
	Fee             int64   `json:"fee"`
}

// CollateralizeFeedTx for construction
type CollateralizeFeedTx struct {
	Price  []float64 `json:"price"`
	Volume []int64   `json:"volume"`
	Fee    int64     `json:"fee"`
}

// CollateralizeRetrieveTx for construction
type CollateralizeRetrieveTx struct {
	CollateralizeID string  `json:"collateralizeId"`
	Balance         float64 `json:"Balance"`
	Fee             int64   `json:"fee"`
}

// CollateralizeManageTx for construction
type CollateralizeManageTx struct {
	DebtCeiling       float64 `json:"debtCeiling"`
	LiquidationRatio  float64 `json:"liquidationRatio"`
	StabilityFeeRatio float64 `json:"stabilityFeeRatio"`
	Period            int64   `json:"period"`
	TotalBalance      float64 `json:"totalBalance"`
	Fee               int64   `json:"fee"`
}
