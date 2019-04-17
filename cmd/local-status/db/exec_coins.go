package db

import "github.com/33cn/chain33/types"
import (
	rpcTypes "github.com/33cn/chain33/rpc/types"
)

type coinsConvert struct {
	block *rpcTypes.BlockDetail
}

func (e *coinsConvert) Convert(ty int64, jsonString string) (key []string, prev, current []byte, err error) {
	if ty == types.TyLogFee {
		return LogFeeConvert([]byte(jsonString))
	}
	return CommonConverts(ty, []byte(jsonString))
}
