/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import (
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
)

func (f *f3d) Query_QueryLastRoundInfo(in *pt.QueryF3DLastRound) (types.Message, error) {
	return queryList(f.GetLocalDB(), f.GetStateDB(), in)
}

func (f *f3d) Query_QueryRoundInfoByRound(in *pt.QueryF3DByRound) (types.Message, error) {
	return queryList(f.GetLocalDB(), f.GetStateDB(), in)
}

func (f *f3d) Query_QueryKeyCountByRoundAndAddr(in *pt.QueryKeyCountByRoundAndAddr) (types.Message, error) {
	return queryList(f.GetLocalDB(), f.GetStateDB(), in)
}

func (f *f3d) Query_QueryBuyRecordByRoundAndAddr(in *pt.QueryBuyRecordByRoundAndAddr) (types.Message, error) {
	return queryList(f.GetLocalDB(), f.GetStateDB(), in)
}
