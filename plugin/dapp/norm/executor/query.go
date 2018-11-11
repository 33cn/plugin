package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/norm/types"
)

func (n *Norm) Query_NormGet(in *pty.NormGetKey) (types.Message, error) {
	value, err := n.GetStateDB().Get(Key(in.Key))
	if err != nil {
		return nil, types.ErrNotFound
	}
	return &types.ReplyString{string(value)}, nil
}
