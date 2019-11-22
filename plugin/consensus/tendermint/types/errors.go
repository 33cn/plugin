// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrHeightLessThanOne error type
	ErrHeightLessThanOne = errors.New("ErrHeightLessThanOne")
	// ErrBaseTxType error type
	ErrBaseTxType = errors.New("ErrBaseTxType")
	// ErrBlockInfoTx error type
	ErrBlockInfoTx = errors.New("ErrBlockInfoTx")
	// ErrBaseExecErr error type
	ErrBaseExecErr = errors.New("ErrBaseExecErr")
	// ErrLastBlockID error type
	ErrLastBlockID = errors.New("ErrLastBlockID")
)
