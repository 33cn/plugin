package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/address"
)

const (
	KeyPrefixContract = "mavl-wasm-%s:%s"
	KeyPrefixLocal    = "LODB-wasm-%s:%s"
)

func contractKey(name string) []byte {
	contractAddr := address.ExecAddress(name)
	key := "mavl-wasm-address:" + contractAddr
	return []byte(key)
}

func (w *Wasm) formatStateKey(key []byte) []byte {
	addr := address.ExecAddress(w.contractAddr)
	return []byte(fmt.Sprintf(KeyPrefixContract, addr, string(key)))
}

func (w *Wasm) formatLocalKey(key []byte) []byte {
	addr := address.ExecAddress(w.contractAddr)
	return []byte(fmt.Sprintf(KeyPrefixLocal, addr, string(key)))
}

func (w *Wasm) contractExist(name string) bool {
	_, err := w.GetStateDB().Get(contractKey(name))
	return err == nil
}

func (w *Wasm) getContract(name string) ([]byte, error) {
	return w.GetStateDB().Get(contractKey(name))
}
