// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	ErrValidateCertFailed  = errors.New("ErrValidateCertFailed")
	ErrGetHistoryCertData  = errors.New("ErrGetHistoryCertData")
	ErrUnknowAuthSignType  = errors.New("ErrUnknowAuthSignType")
	ErrInitializeAuthority = errors.New("ErrInitializeAuthority")
)
