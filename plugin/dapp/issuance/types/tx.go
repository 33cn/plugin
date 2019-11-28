// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// IssuanceCreateTx for construction
type IssuanceCreateTx struct {
	DebtCeiling         float64 `json:"debtCeiling"`
	LiquidationRatio    float64 `json:"liquidationRatio"`
	Period              int64 `json:"period"`
	TotalBalance        float64 `json:"totalBalance"`
	Fee                 int64  `json:"fee"`
}

// IssuanceDebtTx for construction
type IssuanceDebtTx struct {
	IssuanceID string `json:"issuanceId"`
	Value    float64  `json:"value"`
	Fee       int64  `json:"fee"`
}

// IssuanceRepayTx for construction
type IssuanceRepayTx struct {
	IssuanceID string `json:"issuanceId"`
	DebtID     string `json:"debtId"`
	Fee       int64  `json:"fee"`
}

// IssuanceFeedTx for construction
type IssuanceFeedTx struct {
	Price     []float64  `json:"price"`
	Volume    []int64  `json:"volume"`
	Fee       int64  `json:"fee"`
}

// IssuanceCloseTx for construction
type IssuanceCloseTx struct {
	IssuanceID string `json:"issuanceId"`
	Fee       int64  `json:"fee"`
}

// IssuanceManageTx for construction
type IssuanceManageTx struct {
	Addr                []string `json:"addr"`
	Fee                 int64  `json:"fee"`
}
