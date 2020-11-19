package wallet

import (
	mixtypes "github.com/33cn/plugin/plugin/dapp/mix/types"
)

// The main interface for ECDH key exchange.
type ECDH interface {
	GenerateKey([]byte) (*mixtypes.PrivKey, *mixtypes.PubKey)
	GenerateSharedSecret(*mixtypes.PrivKey, *mixtypes.PubKey) ([]byte, error)
}
