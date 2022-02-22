package types

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"math/big"
	"strings"
)

func Str2Byte(v string) []byte {
	var f fr.Element
	f.SetString(v)
	b := f.Bytes()
	return b[:]
}

func Byte2Str(v []byte) string {
	var f fr.Element
	f.SetBytes(v)
	return f.String()
}

// HexAddr2Decimal 16进制地址转10进制
func HexAddr2Decimal(addr string) string {
	addrInt, _ := new(big.Int).SetString(strings.ToLower(addr), 16)
	return addrInt.String()
}