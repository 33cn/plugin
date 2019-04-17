package db

import (
	"testing"
	"github.com/33cn/plugin/cmd/local-status/exec"
	"github.com/stretchr/testify/assert"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
)

func Test_LogFeeConvert(t *testing.T) {
	acc := types.ReceiptAccountTransfer{
		Prev: &types.Account{
			Currency:             0,
			Balance:              3,
			Frozen:               2,
			Addr:                 "14Xw2mYPjmMr5vqs3cFdhb9oz5MMxKJGUH",
		},
		Current: &types.Account{
			Currency:             0,
			Balance:              4,
			Frozen:               1,
			Addr:                 "14Xw2mYPjmMr5vqs3cFdhb9oz5MMxKJGUH",
		},
	}
	j, _ := types.PBToJSON(&acc)
	v := &rpctypes.ReceiptLogResult{
		Ty: types.TyLogFee,
		TyName: "LogFee",
		Log: j,
		RawLog: common.ToHex(j),
	}
	k, _, _, err := exec.ParseLog("ticket", v)
	assert.Nil(t, err)
	assert.Equal(t, "coins-bty/coins-14Xw2mYPjmMr5vqs3cFdhb9oz5MMxKJGUH", k)
}