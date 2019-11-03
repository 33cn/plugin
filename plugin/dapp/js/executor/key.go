package executor

import (
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
)

func calcAllPrefix(cfg *types.Chain33Config, name string) ([]byte, []byte) {
	execer := cfg.ExecName("user." + ptypes.JsX + "." + name)
	state := types.CalcStatePrefix([]byte(execer))
	local := types.CalcLocalPrefix([]byte(execer))
	return state, local
}

func calcCodeKey(name string) []byte {
	return append([]byte("mavl-"+ptypes.JsX+"-code-"), []byte(name)...)
}
