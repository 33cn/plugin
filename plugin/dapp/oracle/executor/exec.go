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

func (o *oracle) Exec_EventPublish(payload *oty.EventPublish, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newOracleAction(o, tx, index)
	return action.eventPublish(payload)
}

func (o *oracle) Exec_EventAbort(payload *oty.EventAbort, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newOracleAction(o, tx, index)
	return action.eventAbort(payload)
}

func (o *oracle) Exec_ResultPrePublish(payload *oty.ResultPrePublish, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newOracleAction(o, tx, index)
	return action.resultPrePublish(payload)
}

func (o *oracle) Exec_ResultAbort(payload *oty.ResultAbort, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newOracleAction(o, tx, index)
	return action.resultAbort(payload)
}

func (o *oracle) Exec_ResultPublish(payload *oty.ResultPublish, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newOracleAction(o, tx, index)
	return action.resultPublish(payload)
}
