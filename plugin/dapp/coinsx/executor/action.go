// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	manager "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	coinTy "github.com/33cn/plugin/plugin/dapp/coinsx/types"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
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
	exec         *Coinsx
}

func newAction(t *Coinsx, tx *types.Transaction) *action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &action{t.GetCoinsAccount(), t.GetStateDB(), t.GetLocalDB(), hash, fromaddr,
		t.GetBlockTime(), t.GetHeight(), dapp.ExecAddress(string(tx.Execer)), t.GetAPI(), tx, t}
}

func makeManagerStatusReceipt(prev, current *coinTy.ManagerStatus) *types.Receipt {
	key := calcManagerStatusKey()
	log := &coinTy.ReceiptManagerStatus{
		Prev: prev,
		Curr: current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  coinTy.TyCoinsxManagerStatusLog,
				Log: types.Encode(log),
			},
		},
	}
}

// isSuperManager is supper manager or not
func isSuperManager(cfg *types.Chain33Config, addr string) bool {
	confManager := types.ConfSub(cfg, manager.ManageX)
	for _, m := range confManager.GStrList("superManager") {
		if addr == m {
			return true
		}
	}
	return false
}

func getSuperManager(cfg *types.Chain33Config) []string {
	confManager := types.ConfSub(cfg, manager.ManageX)
	return confManager.GStrList("superManager")
}

func getManagerStatus(db dbm.KV) (*coinTy.ManagerStatus, error) {
	key := calcManagerStatusKey()
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	var status coinTy.ManagerStatus
	err = types.Decode(val, &status)
	return &status, err
}

func (a *action) configTransfer(config *coinTy.TransferFlagConfig) (*types.Receipt, error) {
	if config.Flag != coinTy.TransferFlag_DISABLE && config.Flag != coinTy.TransferFlag_ENABLE {
		return nil, errors.Wrapf(types.ErrInvalidParam, "flag=%d", config.Flag)
	}

	stat, err := getManagerStatus(a.db)
	if err == types.ErrNotFound {
		stat = &coinTy.ManagerStatus{TransferFlag: config.Flag}
		return makeManagerStatusReceipt(nil, stat), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "get manager status")
	}

	if stat != nil && stat.TransferFlag == config.Flag {
		return nil, errors.Wrapf(types.ErrInvalidParam, "same flag set, cur=%d,config=%d", stat.TransferFlag, config.Flag)
	}

	copyStat := proto.Clone(stat).(*coinTy.ManagerStatus)
	stat.TransferFlag = config.Flag
	return makeManagerStatusReceipt(copyStat, stat), nil

}

//过滤重复地址
func filterAddrs(addrs []string) []string {
	f := make(map[string]bool)
	var newAddrs []string
	for _, k := range addrs {
		if !f[k] {
			f[k] = true
			newAddrs = append(newAddrs, k)
		}
	}
	return newAddrs

}

func (a *action) addAccounts(addrs []string) (*types.Receipt, error) {
	curStat, err := getManagerStatus(a.db)
	if err == types.ErrNotFound {
		stat := &coinTy.ManagerStatus{TransferFlag: coinTy.TransferFlag_DISABLE}
		stat.ManagerAccounts = append(stat.ManagerAccounts, addrs...)
		return makeManagerStatusReceipt(nil, stat), nil
	}
	if err != nil {
		return nil, err
	}

	copyStat := proto.Clone(curStat).(*coinTy.ManagerStatus)
	curStat.ManagerAccounts = append(curStat.ManagerAccounts, addrs...)
	curStat.ManagerAccounts = filterAddrs(curStat.ManagerAccounts)

	return makeManagerStatusReceipt(copyStat, curStat), nil

}

//删除掉指定地址
func filterByAddrs(curr, del []string) []string {
	f := make(map[string]bool)
	for _, k := range del {
		f[k] = true
	}
	var newAddrs []string
	for _, k := range curr {
		if !f[k] {
			newAddrs = append(newAddrs, k)
		}
	}
	return newAddrs
}

func (a *action) delAccounts(addrs []string) (*types.Receipt, error) {
	curStat, err := getManagerStatus(a.db)
	if err != nil || err == types.ErrNotFound {
		return nil, err
	}
	copyStat := proto.Clone(curStat).(*coinTy.ManagerStatus)
	curStat.ManagerAccounts = filterByAddrs(curStat.ManagerAccounts, addrs)

	return makeManagerStatusReceipt(copyStat, curStat), nil
}

func (a *action) configAccounts(config *coinTy.ManagerAccountsConfig) (*types.Receipt, error) {
	if config.Op != coinTy.AccountOp_ADD && config.Op != coinTy.AccountOp_DEL {
		return nil, errors.Wrapf(types.ErrInvalidParam, "unsupport op=%d ", config.Op)
	}

	if len(config.Accounts) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "config accounts=%s", config.Accounts)
	}

	addrs := strings.Split(config.Accounts, ",")
	if config.Op == coinTy.AccountOp_ADD {
		return a.addAccounts(addrs)
	}

	return a.delAccounts(addrs)
}

func (a *action) config(config *coinTy.CoinsConfig, tx *types.Transaction, index int) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	from := tx.From()
	//from 必须是超级管理员
	if !isSuperManager(cfg, from) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from=%s is not super manager", from)
	}
	switch config.Ty {
	case coinTy.ConfigType_TRANSFER:
		return a.configTransfer(config.GetTransferFlag())
	case coinTy.ConfigType_ACCOUNTS:
		return a.configAccounts(config.GetManagerAccounts())

	}
	return nil, errors.Wrapf(types.ErrInvalidParam, "config type=%d not support", config.Ty)
}

func isManager(addr string, managers []string) bool {
	for _, m := range managers {
		if addr == m {
			return true
		}
	}
	return false
}

/*
   p2p转账受限规则，to执行器也受限制
1. 超级管理员或者配置了管理员账号 转账不受限
2. 如果未配置转账使能标志或配置为DISABLE，都受限制
3. 配置ENABLE,则不受限
*/
func checkTransferEnable(cfg *types.Chain33Config, db dbm.KV, from, to string) bool {
	//节点配置的超级管理员转账不受限
	suppers := getSuperManager(cfg)
	if isManager(from, suppers) || isManager(to, suppers) {
		return true
	}

	stat, err := getManagerStatus(db)
	if stat != nil {
		//如果转账不受限，则可以任意转账
		if stat.TransferFlag == coinTy.TransferFlag_ENABLE {
			return true
		}
		//如果转账受限，则任一方是管理员才允许转账
		if isManager(from, stat.ManagerAccounts) || isManager(to, stat.ManagerAccounts) {
			return true
		}
	}
	if err != nil && err != types.ErrNotFound {
		clog.Error("checkTransferEnable", "err", err)
	}
	return false
}
