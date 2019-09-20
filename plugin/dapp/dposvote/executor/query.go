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

//Query_QueryVrfByCycleForTopN method
func (d *DPos) Query_QueryVrfByCycleForTopN(in *dty.DposVrfQuery) (types.Message, error) {
	return queryVrfByCycleForTopN(d.GetLocalDB(), in)
}

//Query_QueryVrfByCycleForPubkeys method
func (d *DPos) Query_QueryVrfByCycleForPubkeys(in *dty.DposVrfQuery) (types.Message, error) {
	return queryVrfByCycleForPubkeys(d.GetLocalDB(), in)
}

//Query_QueryCBInfoByCycle method
func (d *DPos) Query_QueryCBInfoByCycle(in *dty.DposCBQuery) (types.Message, error) {
	return queryCBInfoByCycle(d.GetLocalDB(), in)
}

//Query_QueryCBInfoByHeight method
func (d *DPos) Query_QueryCBInfoByHeight(in *dty.DposCBQuery) (types.Message, error) {
	return queryCBInfoByHeight(d.GetLocalDB(), in)
}

//Query_QueryCBInfoByHash method
func (d *DPos) Query_QueryCBInfoByHash(in *dty.DposCBQuery) (types.Message, error) {
	return queryCBInfoByHash(d.GetLocalDB(), in)
}

//Query_QueryTopNByVersion method
func (d *DPos) Query_QueryTopNByVersion(in *dty.TopNCandidatorsQuery) (types.Message, error) {
	return queryTopNByVersion(d.GetStateDB(), in)
}
