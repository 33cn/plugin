/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import (
	"github.com/33cn/chain33/types"
	oty "github.com/33cn/plugin/plugin/dapp/oracle/types"
)

//从statedb 读取原始数据
func (o *oracle) Query_QueryOraclesByIDs(in *oty.QueryOracleInfos) (types.Message, error) {
	return getOracleLisByIDs(o.GetStateDB(), in)
}

//通过状态查询ids
func (o *oracle) Query_QueryEventIDsByStatus(in *oty.QueryEventID) (types.Message, error) {
	eventIds, err := getEventIDListByStatus(o.GetLocalDB(), in.Status, in.EventID)
	if err != nil {
		return nil, err
	}

	return eventIds, nil
}

//通过状态 和 地址查询
func (o *oracle) Query_QueryEventIDsByAddrAndStatus(in *oty.QueryEventID) (types.Message, error) {
	eventIds, err := getEventIDListByAddrAndStatus(o.GetLocalDB(), in.Addr, in.Status, in.EventID)
	if err != nil {
		return nil, err
	}

	return eventIds, nil
}

//通过类型和状态查询
func (o *oracle) Query_QueryEventIDsByTypeAndStatus(in *oty.QueryEventID) (types.Message, error) {
	eventIds, err := getEventIDListByTypeAndStatus(o.GetLocalDB(), in.Type, in.Status, in.EventID)
	if err != nil {
		return nil, err
	}
	return eventIds, nil
}
