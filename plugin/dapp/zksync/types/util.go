package types

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

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