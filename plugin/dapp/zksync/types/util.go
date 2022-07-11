package types

import (
	"encoding/hex"
	"github.com/pkg/errors"
	"math/big"
	"strings"

	"github.com/33cn/chain33/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
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
func HexAddr2Decimal(addr string) (string, bool) {
	addrInt, ok := new(big.Int).SetString(strings.ToLower(addr), 16)
	if !ok {
		return "", false
	}
	return addrInt.String(), true
}

// DecimalAddr2Hex 10进制地址转16进制
func DecimalAddr2Hex(addr string) (string, bool) {
	addrInt, ok := new(big.Int).SetString(strings.ToLower(addr), 10)
	if !ok {
		return "", false
	}
	return hex.EncodeToString(addrInt.Bytes()), true
}

//DecodePacVal decode pac val with man+exp format
func DecodePacVal(p []byte, expBitWidth int) string {
	var v, expMask big.Int
	v.SetBytes(p)
	expMask.SetString(strings.Repeat("1", expBitWidth), 2)
	expV := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).And(&v, &expMask), nil)
	manV := big.NewInt(0).Rsh(&v, uint(expBitWidth))
	return Byte2Str(big.NewInt(0).Mul(manV, expV).Bytes())
}

func SplitNFTContent(contentHash string) (*big.Int, *big.Int, string, error) {
	hexContent := strings.ToLower(contentHash)
	if hexContent[0:2] == "0x" || hexContent[0:2] == "0X" {
		hexContent = hexContent[2:]
	}

	if len(hexContent) != 64 {
		return nil, nil, "", errors.Wrapf(types.ErrInvalidParam, "contentHash not 64 len, %s", hexContent)
	}
	part1, ok := big.NewInt(0).SetString(hexContent[:32], 16)
	if !ok {
		return nil, nil, "", errors.Wrapf(types.ErrInvalidParam, "contentHash.preHalf hex err, %s", hexContent[:32])
	}
	part2, ok := big.NewInt(0).SetString(hexContent[32:], 16)
	if !ok {
		return nil, nil, "", errors.Wrapf(types.ErrInvalidParam, "contentHash.postHalf hex err, %s", hexContent[32:])
	}
	return part1, part2, hexContent, nil
}
