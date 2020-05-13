package types

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strings"

	log "github.com/33cn/chain33/common/log/log15"
)

var (
	//日志
	clog = log.New("module", "common")
)

func Float64ToBytes(float float64) []byte {
	result := make([]byte, 8)
	binary.LittleEndian.PutUint64(result, math.Float64bits(float))
	return result
}

func BytesToFloat64(bytes []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(bytes))
}

func AddressIsEmpty(address string) bool {
	if address == "" {
		return true
	}

	var aa2 string
	return address == aa2
}

func AddToStringMap(in *StringMap, validator string) *StringMap {
	inStringMap := append(in.GetValidators(), validator)
	stringMapRes := new(StringMap)
	stringMapRes.Validators = inStringMap
	return stringMapRes
}

func DivideSpecifyTimes(start, time int64) int64 {
	for i := 0; i < int(time); i++ {
		start /= 10
	}
	return start
}

func MultiplySpecifyTimes(start float64, time int64) float64 {
	for i := 0; i < int(time); i++ {
		start *= 10
	}
	return start
}

func Toeth(amount string, decimal int64) float64 {

	bf := big.NewFloat(0)
	var ok bool
	bf, ok = bf.SetString(TrimZeroAndDot(amount))
	if !ok {
		return 0
	}
	bf = bf.Quo(bf, big.NewFloat(MultiplySpecifyTimes(1, decimal)))
	f, _ := bf.Float64()
	return f
}

//将eth单位的金额转为wei单位
func ToWei(amount float64, decimal int64) *big.Int {

	var ok bool
	bn := big.NewInt(1)
	if decimal > 4 {
		bn, ok = bn.SetString(TrimZeroAndDot(fmt.Sprintf("%.0f", MultiplySpecifyTimes(math.Trunc(amount*1e4), decimal-4))), 10)
	} else {
		bn, ok = bn.SetString(TrimZeroAndDot(fmt.Sprintf("%.0f", MultiplySpecifyTimes(amount, decimal))), 10)
	}
	if ok {
		return bn
	}

	return nil
}

func TrimZeroAndDot(s string) string {
	if strings.Contains(s, ".") {
		var trimDotStr string
		trimZeroStr := strings.TrimRight(s, "0")
		trimDotStr = strings.TrimRight(trimZeroStr, ".")
		return trimDotStr
	}

	return s
}

func CheckPower(power int64) bool {
	if power <= 0 || power > 100 {
		return false
	}
	return true
}
