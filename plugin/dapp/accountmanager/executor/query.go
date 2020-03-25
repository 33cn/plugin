package executor

import (
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

//根据ID查询账户信息
func (s *accountmanager) Query_QueryAccountByID(in *et.QueryAccountByID) (types.Message, error) {
	return findAccountByID(s.GetLocalDB(), in.AccountID)
}

//根据ID查询账户信息
func (s *accountmanager) Query_QueryAccountByAddr(in *et.QueryAccountByAddr) (types.Message, error) {
	return findAccountByAddr(s.GetLocalDB(), in.Addr)
}

//根据状态查询账户列表||  账户状态 1 正常， 2表示冻结, 3表示锁定 4,过期注销
func (s *accountmanager) Query_QueryAccountsByStatus(in *et.QueryAccountsByStatus) (types.Message, error) {
	return findAccountListByStatus(s.GetLocalDB(), in.Status, in.Direction, in.PrimaryKey)
}

//查询逾期注销的账户列表
func (s *accountmanager) Query_QueryExpiredAccounts(in *et.QueryExpiredAccounts) (types.Message, error) {
	return findAccountListByIndex(s.GetLocalDB(), in.ExpiredTime, in.PrimaryKey)
}

//根据ID查询账户余额
func (s *accountmanager) Query_QueryBalanceByID(in *et.QueryBalanceByID) (types.Message, error) {
	return queryBalanceByID(s.GetStateDB(), s.GetLocalDB(), s.GetAPI().GetConfig(), s.GetName(), in)
}
