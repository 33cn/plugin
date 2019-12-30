package test

import (
	"github.com/33cn/chain33/types"
	"github.com/gogo/protobuf/proto"
)

type Cli interface {
	Send(tx *types.Transaction, hexKey string) ([]*types.ReceiptLog, error)
	Query(fn string, msg proto.Message) ([]byte, error)
}
