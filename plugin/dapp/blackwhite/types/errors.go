// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	ErrIncorrectStatus  = errors.New("ErrIncorrectStatus")
	ErrRepeatPlayerAddr = errors.New("ErrRepeatPlayerAddress")
	ErrNoTimeoutDone    = errors.New("ErrNoTimeoutDone")
	ErrNoExistAddr      = errors.New("ErrNoExistAddress")
	ErrNoLoopSeq        = errors.New("ErrBlackwhiteFinalloopLessThanSeq")
)
