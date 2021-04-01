package executor

import (
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

// "mavl-wasm-code-{name}"
func contractKey(name string) []byte {
	return append([]byte("mavl-"+types2.WasmX+"-code-"), []byte(name)...)
}

// "mavl-wasm-creator-{name}"
func contractCreatorKey(name string) []byte {
	return append([]byte("mavl-"+types2.WasmX+"-creator-"), []byte(name)...)
}

// "mavl-wasm-{contract}-"
func calcStatePrefix(contract string) []byte {
	var prefix []byte
	prefix = append(prefix, types.CalcStatePrefix([]byte(types2.WasmX))...)
	prefix = append(prefix, []byte(contract)...)
	prefix = append(prefix, '-')
	return prefix
}

// "LODB-wasm-{contract}-"
func calcLocalPrefix(contract string) []byte {
	var prefix []byte
	prefix = append(prefix, types.CalcLocalPrefix([]byte(types2.WasmX))...)
	prefix = append(prefix, []byte(contract)...)
	prefix = append(prefix, '-')
	return prefix
}
