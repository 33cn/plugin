// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	ErrInvalidTitle         = errors.New("ErrInvalidTitle")
	ErrTitleNotExist        = errors.New("ErrTitleNotExist")
	ErrNodeNotForTheTitle   = errors.New("ErrNodeNotForTheTitle")
	ErrParaBlockHashNoMatch = errors.New("ErrParaBlockHashNoMatch")
	ErrParaMinerBaseIndex   = errors.New("ErrParaMinerBaseIndex")
	ErrParaMinerTxType      = errors.New("ErrParaMinerTxType")
	ErrParaEmptyMinerTx     = errors.New("ErrParaEmptyMinerTx")
	ErrParaMinerExecErr     = errors.New("ErrParaMinerExecErr")
)
