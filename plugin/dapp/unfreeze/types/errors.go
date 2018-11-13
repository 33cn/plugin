// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	ErrUnfreezeBeforeDue = errors.New("ErrUnfreezeBeforeDue")
	ErrUnfreezeEmptied   = errors.New("ErrUnfreezeEmptied")
	ErrUnfreezeMeans     = errors.New("ErrUnfreezeMeans")
	ErrUnfreezeID        = errors.New("ErrUnfreezeID")
	ErrNoUnfreezeItem    = errors.New("ErrNoUnfreezeItem")
	ErrNoPrivilege       = errors.New("ErrNoPrivilege")
)
