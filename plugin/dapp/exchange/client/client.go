package client

import (
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

type Cli interface {
	Query(fn string, msg proto.Message) ([]byte, error)
	Send(tx *types.Transaction, hexKey string) ([]*types.ReceiptLog, error)
}
