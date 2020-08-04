package test

import (
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

//Cli interface
type Cli interface {
	Send(tx *types.Transaction, hexKey string) ([]*types.ReceiptLog, error)
	Query(fn string, msg proto.Message) ([]byte, error)
	GetExecAccount(addr string, exec string, symbol string) (*types.Account, error) // 获取 addr 在本合约中 exec 执行器下面的 symbol 子账户
}
