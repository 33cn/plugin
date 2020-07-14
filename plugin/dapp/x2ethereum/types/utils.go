package types

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/33cn/chain33/common/address"
)

//MultiplySpecifyTimes ...
func MultiplySpecifyTimes(start float64, time int64) float64 {
	for i := 0; i < int(time); i++ {
		start *= 10
	}
	return start
}

//Toeth ...
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

//ToWei 将eth单位的金额转为wei单位
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

//TrimZeroAndDot ...
func TrimZeroAndDot(s string) string {
	if strings.Contains(s, ".") {
		var trimDotStr string
		trimZeroStr := strings.TrimRight(s, "0")
		trimDotStr = strings.TrimRight(trimZeroStr, ".")
		return trimDotStr
	}

	return s
}

//CheckPower ...
func CheckPower(power int64) bool {
	if power <= 0 || power > 100 {
		return false
	}
	return true
}

//DivideDot ...
func DivideDot(in string) (left, right string, err error) {
	if strings.Contains(in, ".") {
		ss := strings.Split(in, ".")
		return ss[0], ss[1], nil
	}
	return "", "", errors.New("Divide error")
}

//IsExecAddrMatch ...
func IsExecAddrMatch(name string, to string) bool {
	toaddr := address.ExecAddress(name)
	return toaddr == to
}
