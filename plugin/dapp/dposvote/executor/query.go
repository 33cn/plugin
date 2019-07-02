// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
)

//Query_QueryCandidatorByPubkeys method
func (d *DPos) Query_QueryCandidatorByPubkeys(in *dty.CandidatorQuery) (types.Message, error) {
	return queryCands(d.GetLocalDB(), in)
}

//Query_QueryCandidatorByTopN method
func (d *DPos) Query_QueryCandidatorByTopN(in *dty.CandidatorQuery) (types.Message, error) {
	return queryTopNCands(d.GetLocalDB(), in)
}

//Query_QueryVote method
func (d *DPos) Query_QueryVote(in *dty.DposVoteQuery) (types.Message, error) {
	return queryVote(d.GetLocalDB(), in)
}

//Query_QueryVrfByTime method
func (d *DPos) Query_QueryVrfByTime(in *dty.DposVrfQuery) (types.Message, error) {
	return queryVrfByTime(d.GetLocalDB(), in)
}

//Query_QueryVrfByCycle method
func (d *DPos) Query_QueryVrfByCycle(in *dty.DposVrfQuery) (types.Message, error) {
	return queryVrfByCycle(d.GetLocalDB(), in)
}
