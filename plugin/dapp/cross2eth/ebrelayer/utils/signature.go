package utils

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	//solsha3 "github.com/miguelmota/go-solidity-sha3"
)

func SignClaim4Evm(hash common.Hash, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	rawSignature, _ := prefixMessage(hash, privateKey)
	signature := hexutil.Bytes(rawSignature)
	return signature, nil
}

func prefixMessage(message common.Hash, key *ecdsa.PrivateKey) ([]byte, []byte) {
	//prefixed := solsha3.SoliditySHA3WithPrefix(message[:])
	prefixed := SoliditySHA3WithPrefix(message[:])
	sig, err := crypto.Sign(prefixed, key)
	if err != nil {
		panic(err)
	}
	return sig, prefixed
}
