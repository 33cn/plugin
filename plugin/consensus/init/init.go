// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package init

import (
	_ "github.com/33cn/plugin/plugin/consensus/para"
	_ "github.com/33cn/plugin/plugin/consensus/pbft"
	_ "github.com/33cn/plugin/plugin/consensus/raft"
	_ "github.com/33cn/plugin/plugin/consensus/tendermint"
	_ "github.com/33cn/plugin/plugin/consensus/ticket"
)
