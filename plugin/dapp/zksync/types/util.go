package types

import (
	"encoding/hex"
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

func Byte2Uint64(v []byte) uint64 {
	return new(big.Int).SetBytes(v).Uint64()
}

// HexAddr2Decimal 16进制地址转10进制
func HexAddr2Decimal(addr string) string {
	addrInt, _ := new(big.Int).SetString(strings.ToLower(addr), 16)
	return addrInt.String()
}

// DecimalAddr2Hex 10进制地址转16进制
func DecimalAddr2Hex(addr string) string {
	addrInt, _ := new(big.Int).SetString(strings.ToLower(addr), 10)
	return hex.EncodeToString(addrInt.Bytes())
}