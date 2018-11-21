// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrValidateCertFailed cert校验失败
	ErrValidateCertFailed = errors.New("ErrValidateCertFailed")
	// ErrGetHistoryCertData 获取证书错误
	ErrGetHistoryCertData = errors.New("ErrGetHistoryCertData")
	// ErrUnknowAuthSignType 无效签名类型
	ErrUnknowAuthSignType = errors.New("ErrUnknowAuthSignType")
	// ErrInitializeAuthority 初始化校验器失败
	ErrInitializeAuthority = errors.New("ErrInitializeAuthority")
)
