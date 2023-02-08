package executor

import (
	"testing"

	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/stretchr/testify/assert"
)

func TestGetTreeSideAmount(t *testing.T) {
	amount := "0"
	totalAmount := "0"
	fee := "0"
	sysDecimal := 8
	tokenDecimal := 18
	_, _, _, err := GetTreeSideAmount(amount, totalAmount, fee, sysDecimal, tokenDecimal)
	assert.Nil(t, err)
}

func TestCheckPackValue(t *testing.T) {
	amount := "1234567812345"
	//amount :="12345678123445"
	count := 18
	for i := 0; i < count; i++ {
		amount += "0"
	}
	err := checkPackValue(amount, zt.PacAmountManBitWidth)
	assert.NotNil(t, err)
}
