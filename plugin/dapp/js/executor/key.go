package executor

import (
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
)

func calcLocalPrefix(execer []byte) []byte {
	s := append([]byte("LODB-"), execer...)
	s = append(s, byte('-'))
	return s
}

func calcStatePrefix(execer []byte) []byte {
	s := append([]byte("mavl-"), execer...)
	s = append(s, byte('-'))
	return s
}

func calcAllPrefix(name string) ([]byte, []byte) {
	execer := types.ExecName("user." + ptypes.JsX + "." + name)
	state := calcStatePrefix([]byte(execer))
	local := calcLocalPrefix([]byte(execer))
	return state, local
}

func calcCodeKey(name string) []byte {
	return append([]byte("mavl-"+ptypes.JsX+"-code-"), []byte(name)...)
}

func calcRollbackKey(hash []byte) []byte {
	return append([]byte("LODB-"+ptypes.JsX+"-rollback-"), hash...)
}
