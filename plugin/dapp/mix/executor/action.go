// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"

	"github.com/33cn/chain33/types"

	"github.com/golang/protobuf/proto"

	token "github.com/33cn/plugin/plugin/dapp/token/types"
)

type action struct {
	coinsAccount *account.DB
	db           dbm.KV
	localdb      dbm.KVDB
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
	api          client.QueueProtocolAPI
	tx           *types.Transaction
	exec         *Mix
}

func newAction(t *Mix, tx *types.Transaction) *action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &action{t.GetCoinsAccount(), t.GetStateDB(), t.GetLocalDB(), hash, fromaddr,
		t.GetBlockTime(), t.GetHeight(), dapp.ExecAddress(string(tx.Execer)), t.GetAPI(), tx, t}
}

func createAccount(cfg *types.Chain33Config, execer, symbol string, db dbm.KV) (*account.DB, error) {
	var accDB *account.DB
	if symbol == "" {
		accDB = account.NewCoinsAccount(cfg)
		accDB.SetDB(db)
		return accDB, nil
	}
	if execer == "" {
		execer = token.TokenX
	}
	return account.NewAccountDB(cfg, execer, symbol, db)
}

func isNotFound(err error) bool {
	if err != nil && (err == dbm.ErrNotFoundInDb || err == types.ErrNotFound) {
		return true
	}
	return false
}

func mergeReceipt(receipt1, receipt2 *types.Receipt) *types.Receipt {
	if receipt2 != nil {
		receipt1.KV = append(receipt1.KV, receipt2.KV...)
		receipt1.Logs = append(receipt1.Logs, receipt2.Logs...)
	}

	return receipt1
}

func makeReceipt(key []byte, logTy int32, data proto.Message) *types.Receipt {
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(data)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: logTy, Log: types.Encode(data)},
		},
	}
}
