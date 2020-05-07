// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"errors"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/valnode/types"
)

const managerKey = "tendermint-manager"

// Exec_Node method
func (val *ValNode) Exec_Node(node *pty.ValNode, tx *types.Transaction, index int) (*types.Receipt, error) {
	if !isValidManager(tx.From(), val.GetStateDB()) {
		return nil, errors.New("not valid manager")
	}
	if len(node.GetPubKey()) == 0 {
		return nil, errors.New("validator pubkey is empty")
	}
	if node.GetPower() < 0 {
		return nil, errors.New("validator power must not be negative")
	}
	receipt := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: nil}
	return receipt, nil
}

// Exec_BlockInfo method
func (val *ValNode) Exec_BlockInfo(blockInfo *pty.TendermintBlockInfo, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: nil}
	return receipt, nil
}

func getConfigKey(key string, db dbm.KV) ([]byte, error) {
	configKey := types.ConfigKey(key)
	value, err := db.Get([]byte(configKey))
	if err != nil {
		clog.Error("getConfigKey not find", "configKey", configKey, "err", err)
		return nil, err
	}
	return value, nil
}

func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		clog.Info("getManageKey not find", "manageKey", manageKey, "err", err)
		return getConfigKey(key, db)
	}
	return value, nil
}

func isValidManager(addr string, db dbm.KV) bool {
	value, err := getManageKey(managerKey, db)
	if err != nil {
		clog.Error("isValidManager nil key", "managerKey", managerKey)
		return false
	}
	if value == nil {
		clog.Error("isValidManager nil value")
		return false
	}

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		clog.Error("isValidManager decode fail", "err", err)
		return false
	}

	for _, op := range item.GetArr().Value {
		if op == addr {
			return true
		}
	}
	return false
}
