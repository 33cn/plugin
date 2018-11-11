// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sm2

import (
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/crypto/sm2"
)

type sm2Driver struct {
	sm2.Driver
}

const Name = "auth_sm2"
const ID = 258

func init() {
	crypto.Register(Name, &sm2Driver{})
	crypto.RegisterType(Name, ID)
}
