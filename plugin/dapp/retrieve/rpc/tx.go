// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

type RetrieveBackupTx struct {
	BackupAddr  string `json:"backupAddr"`
	DefaultAddr string `json:"defaultAddr"`
	DelayPeriod int64  `json:"delayPeriod"`
	Fee         int64  `json:"fee"`
}

type RetrievePrepareTx struct {
	BackupAddr  string `json:"backupAddr"`
	DefaultAddr string `json:"defaultAddr"`
	Fee         int64  `json:"fee"`
}

type RetrievePerformTx struct {
	BackupAddr  string `json:"backupAddr"`
	DefaultAddr string `json:"defaultAddr"`
	Fee         int64  `json:"fee"`
}

type RetrieveCancelTx struct {
	BackupAddr  string `json:"backupAddr"`
	DefaultAddr string `json:"defaultAddr"`
	Fee         int64  `json:"fee"`
}
