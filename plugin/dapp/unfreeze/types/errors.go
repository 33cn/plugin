// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrUnfreezeEmptied 没有可提币量
	ErrUnfreezeEmptied = errors.New("ErrUnfreezeEmptied")
	// ErrUnfreezeMeans 解冻币算法错误
	ErrUnfreezeMeans = errors.New("ErrUnfreezeMeans")
	// ErrUnfreezeID 冻结合约ID错误
	ErrUnfreezeID = errors.New("ErrUnfreezeID")
	// ErrNoPrivilege 没有权限
	ErrNoPrivilege = errors.New("ErrNoPrivilege")
	// ErrTerminated 已经被取消过了
	ErrTerminated = errors.New("ErrTerminated")
)
