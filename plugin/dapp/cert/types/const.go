// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

var (
	// CertX cert执行器名
	CertX = "cert"
	// ExecerCert cert执行器字节
	ExecerCert = []byte(CertX)
	actionName = map[string]int32{
		"New":    CertActionNew,
		"Update": CertActionUpdate,
		"Normal": CertActionNormal,
	}
)
