// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"time"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

// Query_GetUnfreezeWithdraw 查询合约可提币量
func (u *Unfreeze) Query_GetUnfreezeWithdraw(in *types.ReqString) (types.Message, error) {
	return QueryWithdraw(u.GetStateDB(), in.GetData())
}