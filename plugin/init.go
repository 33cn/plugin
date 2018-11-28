// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plugin

import (
	_ "github.com/33cn/plugin/plugin/consensus/init"  // register consensus init package
	_ "github.com/33cn/plugin/plugin/crypto/init"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)
