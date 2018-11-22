// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

// hashlock errors
var (
	ErrHashlockAmount       = errors.New("ErrHashlockAmount")
	ErrHashlockHash         = errors.New("ErrHashlockHash")
	ErrHashlockStatus       = errors.New("ErrHashlockStatus")
	ErrTime                 = errors.New("ErrTime")
	ErrHashlockReturnAddrss = errors.New("ErrHashlockReturnAddrss")
	ErrHashlockTime         = errors.New("ErrHashlockTime")
	ErrHashlockReapeathash  = errors.New("ErrHashlockReapeathash")
	ErrHashlockSendAddress  = errors.New("ErrHashlockSendAddress")
)
