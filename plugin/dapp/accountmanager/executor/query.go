package executor

import (
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

//Query_QueryAccountByID 根据ID查询账户信息
func (a *Accountmanager) Query_QueryAccountByID(in *et.QueryAccountByID) (types.Message, error) {
	return findAccountByID(a.GetLocalDB(), in.AccountID)
}

//Query_QueryAccountByAddr 根据ID查询账户信息
func (a *Accountmanager) Query_QueryAccountByAddr(in *et.QueryAccountByAddr) (types.Message, error) {
	return findAccountByAddr(a.GetLocalDB(), in.Addr)
}

//Query_QueryAccountsByStatus 根据状态查询账户列表||  账户状态 1 正常， 2表示冻结, 3表示锁定 4,过期注销
func (a *Accountmanager) Query_QueryAccountsByStatus(in *et.QueryAccountsByStatus) (types.Message, error) {
	return findAccountListByStatus(a.GetLocalDB(), in.Status, in.Direction, in.PrimaryKey)
}

//Query_QueryExpiredAccounts 查询逾期注销的账户列表
func (a *Accountmanager) Query_QueryExpiredAccounts(in *et.QueryExpiredAccounts) (types.Message, error) {
	return findAccountListByIndex(a.GetLocalDB(), in.ExpiredTime, in.PrimaryKey)
}

//Query_QueryBalanceByID 根据ID查询账户余额
func (a *Accountmanager) Query_QueryBalanceByID(in *et.QueryBalanceByID) (types.Message, error) {
	return queryBalanceByID(a.GetStateDB(), a.GetLocalDB(), a.GetAPI().GetConfig(), a.GetName(), in)
}
