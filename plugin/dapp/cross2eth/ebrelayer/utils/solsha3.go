package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"golang.org/x/crypto/sha3"
)

// TODO: refactor and remove duplicated code

// Address address
func Address(input interface{}) []byte {
	switch v := input.(type) {
	case common.Address:
		return v.Bytes()[:]
	case string:
		v = strings.TrimPrefix(v, "0x")
		if v == "" || v == "0" {
			return []byte{0}
		}

		if len(v)%2 == 1 {
			v = "0" + v
		}

		decoded, err := hex.DecodeString(v)
		if err != nil {
			panic(err)
		}

		return decoded
	case []byte:
		return v
	}

	if isArray(input) {
		return AddressArray(input)
	}

	return common.HexToAddress("").Bytes()[:]
}

// AddressArray address
func AddressArray(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Address(val), 32)
		values = append(values, result...)
	}
	return values
}

// Uint256 uint256
func Uint256(input interface{}) []byte {
	switch v := input.(type) {
	case *big.Int:
		return math.U256Bytes(v)
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		return math.U256Bytes(bn)
	}

	if isArray(input) {
		return Uint256Array(input)
	}

	return common.RightPadBytes([]byte(""), 32)
}

// Uint256Array uint256 array
func Uint256Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Uint256(val), 32)
		values = append(values, result...)
	}
	return values
}

// Uint128 uint128
func Uint128(input interface{}) []byte {
	switch v := input.(type) {
	case *big.Int:
		return common.LeftPadBytes(v.Bytes(), 16)
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		return common.LeftPadBytes(bn.Bytes(), 16)
	}

	if isArray(input) {
		return Uint128Array(input)
	}

	return common.LeftPadBytes([]byte(""), 16)
}

// Uint128Array uint128
func Uint128Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Uint128(val), 32)
		values = append(values, result...)
	}
	return values
}

// Uint64 uint64
func Uint64(input interface{}) []byte {
	b := new(bytes.Buffer)
	switch v := input.(type) {
	case *big.Int:
		binary.Write(b, binary.BigEndian, v.Uint64())
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.Write(b, binary.BigEndian, bn.Uint64())
	case uint64:
		binary.Write(b, binary.BigEndian, v)
	case uint32:
		binary.Write(b, binary.BigEndian, uint64(v))
	case uint16:
		binary.Write(b, binary.BigEndian, uint64(v))
	case uint8:
		binary.Write(b, binary.BigEndian, uint64(v))
	case uint:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int64:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int32:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int16:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int8:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int:
		binary.Write(b, binary.BigEndian, uint64(v))
	default:
		binary.Write(b, binary.BigEndian, uint64(0))
	}

	if isArray(input) {
		return Uint64Array(input)
	}

	return b.Bytes()
}

// Uint64Array uint64 array
func Uint64Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Uint64(val), 32)
		values = append(values, result...)
	}
	return values
}

// Uint32 uint32
func Uint32(input interface{}) []byte {
	b := new(bytes.Buffer)
	switch v := input.(type) {
	case *big.Int:
		binary.Write(b, binary.BigEndian, uint32(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.Write(b, binary.BigEndian, uint32(bn.Uint64()))
	case uint64:
		binary.Write(b, binary.BigEndian, uint32(v))
	case uint32:
		binary.Write(b, binary.BigEndian, uint32(v))
	case uint16:
		binary.Write(b, binary.BigEndian, uint32(v))
	case uint8:
		binary.Write(b, binary.BigEndian, uint32(v))
	case uint:
		binary.Write(b, binary.BigEndian, uint32(v))
	case int64:
		binary.Write(b, binary.BigEndian, uint32(v))
	case int32:
		binary.Write(b, binary.BigEndian, v)
	case int16:
		binary.Write(b, binary.BigEndian, uint32(v))
	case int8:
		binary.Write(b, binary.BigEndian, uint32(v))
	case int:
		binary.Write(b, binary.BigEndian, uint32(v))
	default:
		binary.Write(b, binary.BigEndian, uint32(0))
	}

	if isArray(input) {
		return Uint32Array(input)
	}

	return b.Bytes()
}

// Uint32Array uint32 array
func Uint32Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Uint32(val), 32)
		values = append(values, result...)
	}
	return values
}

// Uint16 uint16
func Uint16(input interface{}) []byte {
	b := new(bytes.Buffer)
	switch v := input.(type) {
	case *big.Int:
		binary.Write(b, binary.BigEndian, uint16(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.Write(b, binary.BigEndian, uint16(bn.Uint64()))
	case uint64:
		binary.Write(b, binary.BigEndian, uint16(v))
	case uint32:
		binary.Write(b, binary.BigEndian, uint16(v))
	case uint16:
		binary.Write(b, binary.BigEndian, v)
	case uint8:
		binary.Write(b, binary.BigEndian, uint16(v))
	case uint:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int64:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int32:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int16:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int8:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int:
		binary.Write(b, binary.BigEndian, uint16(v))
	default:
		binary.Write(b, binary.BigEndian, uint16(0))
	}

	if isArray(input) {
		return Uint16Array(input)
	}

	return b.Bytes()
}

// Uint16Array uint16 array
func Uint16Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Uint16(val), 32)
		values = append(values, result...)
	}
	return values
}

// Uint8 uint8
func Uint8(input interface{}) []byte {
	b := new(bytes.Buffer)
	switch v := input.(type) {
	case *big.Int:
		binary.Write(b, binary.BigEndian, uint8(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.Write(b, binary.BigEndian, uint8(bn.Uint64()))
	case uint64:
		binary.Write(b, binary.BigEndian, uint8(v))
	case uint32:
		binary.Write(b, binary.BigEndian, uint8(v))
	case uint16:
		binary.Write(b, binary.BigEndian, uint8(v))
	case uint8:
		binary.Write(b, binary.BigEndian, v)
	case uint:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int64:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int32:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int16:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int8:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int:
		binary.Write(b, binary.BigEndian, uint8(v))
	default:
		binary.Write(b, binary.BigEndian, uint8(0))
	}

	if isArray(input) {
		return Uint8Array(input)
	}

	return b.Bytes()
}

// Uint8Array uint8 array
func Uint8Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Uint8(val), 32)
		values = append(values, result...)
	}
	return values
}

// Int256 int256
func Int256(input interface{}) []byte {
	switch v := input.(type) {
	case *big.Int:
		return common.LeftPadBytes(v.Bytes(), 32)
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint64:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint32:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint16:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint8:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int64:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int32:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int16:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int8:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	}

	if isArray(input) {
		return Int256Array(input)
	}

	return common.LeftPadBytes([]byte{}, 32)
}

// Int256Array int256 array
func Int256Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Int256(val), 32)
		values = append(values, result...)
	}
	return values
}

// Int128 int128
func Int128(input interface{}) []byte {
	switch v := input.(type) {
	case *big.Int:
		return common.LeftPadBytes(v.Bytes(), 16)
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint64:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint32:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint16:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint8:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int64:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int32:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int16:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int8:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	}

	if isArray(input) {
		return Int128Array(input)
	}

	return common.LeftPadBytes([]byte{}, 16)
}

// Int128Array int128 array
func Int128Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Int128(val), 32)
		values = append(values, result...)
	}
	return values
}

// Int64 int64
func Int64(input interface{}) []byte {
	b := make([]byte, 8)
	switch v := input.(type) {
	case *big.Int:
		binary.BigEndian.PutUint64(b, v.Uint64())
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.BigEndian.PutUint64(b, bn.Uint64())
	case uint64:
		binary.BigEndian.PutUint64(b, v)
	case uint32:
		binary.BigEndian.PutUint64(b, uint64(v))
	case uint16:
		binary.BigEndian.PutUint64(b, uint64(v))
	case uint8:
		binary.BigEndian.PutUint64(b, uint64(v))
	case uint:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int64:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int32:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int16:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int8:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int:
		binary.BigEndian.PutUint64(b, uint64(v))
	default:
		binary.BigEndian.PutUint64(b, uint64(0))
	}

	if isArray(input) {
		return Int64Array(input)
	}

	return b
}

// Int64Array int64 array
func Int64Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Int64(val), 32)
		values = append(values, result...)
	}
	return values
}

// Int32 int32
func Int32(input interface{}) []byte {
	b := make([]byte, 4)
	switch v := input.(type) {
	case *big.Int:
		binary.BigEndian.PutUint32(b, uint32(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.BigEndian.PutUint32(b, uint32(bn.Uint64()))
	case uint64:
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint32:
		binary.BigEndian.PutUint32(b, v)
	case uint16:
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint8:
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int64:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int32:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int16:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int8:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int:
		binary.BigEndian.PutUint32(b, uint32(v))
	default:
		binary.BigEndian.PutUint32(b, uint32(0))
	}

	if isArray(input) {
		return Int32Array(input)
	}

	return b
}

// Int32Array int32
func Int32Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Int32(val), 32)
		values = append(values, result...)
	}
	return values
}

// Int16 int16
func Int16(input interface{}) []byte {
	b := make([]byte, 2)
	switch v := input.(type) {
	case *big.Int:
		binary.BigEndian.PutUint16(b, uint16(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.BigEndian.PutUint16(b, uint16(bn.Uint64()))
	case uint64:
		binary.BigEndian.PutUint16(b, uint16(v))
	case uint32:
		binary.BigEndian.PutUint16(b, uint16(v))
	case uint16:
		binary.BigEndian.PutUint16(b, v)
	case uint8:
		binary.BigEndian.PutUint16(b, uint16(v))
	case uint:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int64:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int32:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int16:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int8:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int:
		binary.BigEndian.PutUint16(b, uint16(v))
	default:
		binary.BigEndian.PutUint16(b, uint16(0))
	}

	if isArray(input) {
		return Int16Array(input)
	}

	return b
}

// Int16Array int16 array
func Int16Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Int16(val), 32)
		values = append(values, result...)
	}
	return values
}

// Int8 int8
func Int8(input interface{}) []byte {
	b := make([]byte, 1)
	switch v := input.(type) {
	case *big.Int:
		b[0] = byte(int8(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		b[0] = byte(int8(bn.Uint64()))
	case uint64:
		b[0] = byte(int8(v))
	case uint32:
		b[0] = byte(int8(v))
	case uint16:
		b[0] = byte(int8(v))
	case uint8:
		b[0] = byte(int8(v))
	case uint:
		b[0] = byte(int8(v))
	case int64:
		b[0] = byte(int8(v))
	case int32:
		b[0] = byte(int8(v))
	case int16:
		b[0] = byte(int8(v))
	case int8:
		b[0] = byte(v)
	case int:
		b[0] = byte(int8(v))
	default:
		b[0] = byte(int8(0))
	}

	if isArray(input) {
		return Int8Array(input)
	}

	return b
}

// Int8Array int8 array
func Int8Array(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Int8(val), 32)
		values = append(values, result...)
	}
	return values
}

// Bytes32 bytes32
func Bytes32(input interface{}) []byte {
	switch v := input.(type) {
	case [32]byte:
		return common.RightPadBytes(v[:], 32)
	case []byte:
		return common.RightPadBytes(v, 32)
	case string:
		if isHex(v) {
			hexb, _ := hex.DecodeString(strings.TrimPrefix(v, "0x"))
			return common.RightPadBytes(hexb, 32)
		}
		str := fmt.Sprintf("%x", v)
		hexb, _ := hex.DecodeString(str)
		return common.RightPadBytes(hexb, 32)
	default:
		return common.RightPadBytes([]byte(""), 32)
	}
}

// Bytes16 bytes16
func Bytes16(input interface{}) []byte {
	switch v := input.(type) {
	case [16]byte:
		return common.RightPadBytes(v[:], 16)
	case []byte:
		return common.RightPadBytes(v, 16)
	case string:
		str := fmt.Sprintf("%x", v)
		hexb, _ := hex.DecodeString(str)
		return common.RightPadBytes(hexb, 16)
	default:
		return common.RightPadBytes([]byte(""), 16)
	}
}

// String string
func String(input interface{}) []byte {
	switch v := input.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	}

	if isArray(input) {
		return StringArray(input)
	}

	return []byte("")
}

// StringArray string
func StringArray(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := String(val)
		values = append(values, result...)
	}
	return values
}

// Bool bool
func Bool(input interface{}) []byte {
	switch v := input.(type) {
	case bool:
		if v {
			return []byte{0x1}
		}
		return []byte{0x0}
	}

	if isArray(input) {
		return BoolArray(input)
	}

	return []byte{0x0}
}

// BoolArray bool array
func BoolArray(input interface{}) []byte {
	var values []byte
	s := reflect.ValueOf(input)
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i).Interface()
		result := common.LeftPadBytes(Bool(val), 32)
		values = append(values, result...)
	}
	return values
}

// SoliditySHA3 solidity sha3
func SoliditySHA3(data ...interface{}) []byte {
	types, ok := data[0].([]string)
	if len(data) > 1 && ok {
		rest := data[1:]
		if len(rest) == len(types) {
			return solsha3(types, data[1:]...)
		}
		iface, ok := data[1].([]interface{})
		if ok {
			return solsha3(types, iface...)
		}
	}

	var v [][]byte
	for _, item := range data {
		v = append(v, item.([]byte))
	}
	return solsha3Legacy(v...)
}

// solsha3Legacy solidity sha3
func solsha3Legacy(data ...[]byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	bs := concatByteSlices(data...)

	hash.Write(bs)
	return hash.Sum(nil)
}

// SoliditySHA3WithPrefix solidity sha3 with prefix
func SoliditySHA3WithPrefix(data []byte) []byte {
	result := SoliditySHA3(
		concatByteSlices(
			[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(data))),
			data,
		),
	)

	return result
}

// ConcatByteSlices concat byte slices
func ConcatByteSlices(arrays ...[]byte) []byte {
	return concatByteSlices(arrays...)
}

func concatByteSlices(arrays ...[]byte) []byte {
	var result []byte

	for _, b := range arrays {
		result = append(result, b...)
	}

	return result
}

func isArray(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Array ||
		reflect.TypeOf(value).Kind() == reflect.Slice
}

// solsha3 solidity sha3
func solsha3(types []string, values ...interface{}) []byte {

	var b [][]byte
	for i, typ := range types {
		b = append(b, pack(typ, values[i], false))
	}

	hash := sha3.NewLegacyKeccak256()
	bs := concatByteSlices(b...)

	hash.Write(bs)
	return hash.Sum(nil)
}

func padZeros(value []byte, width int) []byte {
	return common.LeftPadBytes(value, width)
}

var zeros = "0000000000000000000000000000000000000000000000000000000000000000"

func pack(typ string, value interface{}, _isArray bool) []byte {
	switch typ {
	case "address":
		if _isArray {
			return padZeros(Address(value), 32)
		}

		return Address(value)
	case "string":
		return String(value)
	case "bool":
		if _isArray {
			return padZeros(Bool(value), 32)
		}

		return Bool(value)
	}

	regexNumber := regexp.MustCompile(`^(u?int)([0-9]*)$`)
	matches := regexNumber.FindAllStringSubmatch(typ, -1)
	if len(matches) > 0 {
		match := matches[0]
		var err error
		size := 256
		if len(match) > 1 {
			//signed = match[1] == "int"
		}
		if len(match) > 2 {
			size, err = strconv.Atoi(match[2])
			if err != nil {
				panic(err)
			}
		}

		_ = size
		if (size%8 != 0) || size == 0 || size > 256 {
			panic("invalid number type " + typ)
		}

		if _isArray {
			size = 256
		}

		var v []byte
		if strings.HasPrefix(typ, "int8") {
			v = Int8(value)
		} else if strings.HasPrefix(typ, "int16") {
			v = Int16(value)
		} else if strings.HasPrefix(typ, "int32") {
			v = Int32(value)
		} else if strings.HasPrefix(typ, "int64") {
			v = Int64(value)
		} else if strings.HasPrefix(typ, "int128") {
			v = Int128(value)
		} else if strings.HasPrefix(typ, "int256") {
			v = Int256(value)
		} else if strings.HasPrefix(typ, "uint8") {
			v = Uint8(value)
		} else if strings.HasPrefix(typ, "uint16") {
			v = Uint16(value)
		} else if strings.HasPrefix(typ, "uint32") {
			v = Uint32(value)
		} else if strings.HasPrefix(typ, "uint128") {
			v = Uint128(value)
		} else if strings.HasPrefix(typ, "uint64") {
			v = Uint64(value)
		} else if strings.HasPrefix(typ, "uint256") {
			v = Uint256(value)
		}
		return padZeros(v, size/8)
	}

	regexBytes := regexp.MustCompile(`^bytes([0-9]+)$`)
	matches = regexBytes.FindAllStringSubmatch(typ, -1)
	if len(matches) > 0 {
		match := matches[0]

		size, err := strconv.Atoi(match[1])
		if err != nil {
			panic(err)
		}

		_ = size

		strSize := strconv.Itoa(size)
		if strSize != match[1] || size == 0 || size > 32 {
			panic("invalid number type " + typ)
		}

		//if (bytes_1.arrayify(value).byteLength !== size) {
		//throw new Error('invalid value for ' + type);
		//}

		if _isArray {
			s := reflect.ValueOf(value)
			v := s.Index(0).Bytes()
			z := make([]byte, 64)
			copy(z[:], v[:])
			return z[:]
		}

		str, isString := value.(string)
		if isString && isHex(str) {
			s := strings.TrimPrefix(str, "0x")
			if len(s)%2 == 1 {
				s = "0" + s
			}
			hexb, err := hex.DecodeString(s)
			if err != nil {
				panic(err)
			}
			z := make([]byte, size)
			copy(z[:], hexb)
			return z
		} else if isString {
			s := reflect.ValueOf(value)
			z := make([]byte, size)
			copy(z[:], s.Bytes())
			return z
		}

		s := reflect.ValueOf(value)
		z := make([]byte, size)
		b := make([]byte, s.Len())
		for i := 0; i < s.Len(); i++ {
			b[i] = s.Index(i).Interface().(byte)
		}
		copy(z[:], b[:])
		return z
	}

	regexArray := regexp.MustCompile(`^(.*)\[([0-9]*)\]$`)
	matches = regexArray.FindAllStringSubmatch(typ, -1)
	if len(matches) > 0 {
		match := matches[0]

		_ = match
		if isArray(value) {
			baseType := match[1]
			k := reflect.ValueOf(value)
			count := k.Len()
			var err error
			if len(match) > 1 && match[2] != "" {
				count, err = strconv.Atoi(match[2])
				if err != nil {
					panic(err)
				}
			}
			if count != k.Len() {
				panic("invalid value for " + typ)
			}

			var result [][]byte
			for i := 0; i < k.Len(); i++ {
				val := k.Index(i).Interface()

				result = append(result, pack(baseType, val, true))
			}

			return concatByteSlices(result...)
		}
	}
	return nil
}

// Pack ...
func Pack(types []string, values []interface{}) []byte {
	if len(types) != len(values) {
		panic("type/value count mismatch")
	}

	var tight [][]byte
	for i, typ := range types {
		tight = append(tight, pack(typ, values[i], false))
	}

	return concatByteSlices(tight...)
}

func keccak256(types []string, values []interface{}) {
	//return keccak256_1.keccak256(pack(types, values))
}

func isHex(str string) bool {
	return strings.HasPrefix(str, "0x")
}
