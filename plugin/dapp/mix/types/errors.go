// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrZkVerifyFail zk verify fail
	ErrZkVerifyFail = errors.New("ErrZkVerifyFail")
	//ErrInputParaNotMatch input paras not match
	ErrInputParaNotMatch = errors.New("ErrInputParaNotMatch")
	//ErrLeafNotFound not found leaf
	ErrLeafNotFound = errors.New("ErrLeafNotFound")
	//ErrTreeRootHashNotFound not found leaf
	ErrTreeRootHashNotFound = errors.New("ErrTreeRootHashNotFound")

	//ErrNulliferHashExist exist
	ErrNulliferHashExist = errors.New("ErrNulliferHashExist")
	//ErrAuthorizeHashExist exist
	ErrAuthorizeHashExist = errors.New("ErrAuthorizeHashExist")

	//ErrSpendInOutValueNotMatch spend input and output value not match
	ErrSpendInOutValueNotMatch = errors.New("ErrSpendInOutValueNotMatch")
)
