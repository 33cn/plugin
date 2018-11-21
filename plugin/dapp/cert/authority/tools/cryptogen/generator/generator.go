// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package generator

import "crypto/x509"

// CAGenerator CA生成器接口
type CAGenerator interface {
	SignCertificate(baseDir, name string, sans []string, pub interface{}) (*x509.Certificate, error)

	GenerateLocalUser(baseDir, name string) error
}
