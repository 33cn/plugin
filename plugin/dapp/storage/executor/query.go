package executor

import (
	"github.com/33cn/chain33/types"
	storagetypes "github.com/33cn/plugin/plugin/dapp/storage/types"
)


//从statedb 读取原始数据
func (s *storage) Query_QueryStorage(in *storagetypes.QueryStorage) (types.Message, error) {
	return QueryStorage(s.GetStateDB(), in)
}

//通过状态查询ids
func (s *storage) Query_BatchQueryStorage(in *storagetypes.BatchQueryStorage) (types.Message, error) {
	return BatchQueryStorage(s.GetStateDB(), in)
}