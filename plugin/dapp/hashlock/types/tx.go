// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

type HashlockLockTx struct {
	Secret     string `json:"secret"`
	Amount     int64  `json:"amount"`
	Time       int64  `json:"time"`
	ToAddr     string `json:"toAddr"`
	ReturnAddr string `json:"returnAddr"`
	Fee        int64  `json:"fee"`
}

type HashlockUnlockTx struct {
	Secret string `json:"secret"`
	Fee    int64  `json:"fee"`
}

type HashlockSendTx struct {
	Secret string `json:"secret"`
	Fee    int64  `json:"fee"`
}
