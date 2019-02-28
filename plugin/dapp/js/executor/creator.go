// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"

	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
)

func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	return db.Get([]byte(manageKey))
}

func checkPriv(addr, key string, db dbm.KV) error {
	value, err := getManageKey(key, db)
	if err != nil {
		return err
	}
	if value == nil {
		return ptypes.ErrJsCreator
	}

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		return err
	}

	for _, op := range item.GetArr().Value {
		if op == addr {
			return nil
		}
	}

	return ptypes.ErrJsCreator
}
