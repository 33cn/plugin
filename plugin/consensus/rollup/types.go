package rollup

import (
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

// Config rollup 配置
type Config struct {
	FullDataCommit bool `json:"fullDataCommit,omitempty"`
	// StartHeight 指定平行链启动时主链的高度
	StartHeight        int64  `json:"startHeight,omitempty"`
	MaxCommitInterval  int64  `json:"maxCommitInterval,omitempty"`
	AuthAccount        string `json:"authAccount,omitempty"`
	AuthKey            string `json:"authKey,omitempty"`
	AddressID          int32  `json:"addressID,omitempty"`
	ReservedMainHeight int64  `json:"reservedMainHeight,omitempty"`
	// CrossTxSyncedHeight 手动指定跨链交易同步的起始主链高度, 大于 0 时覆盖链上记录的值, 0 表示不覆盖
	CrossTxSyncedHeight int64 `json:"crossTxSyncedHeight,omitempty"`
}

type validatorSignMsgSet struct {
	self   *rtypes.ValidatorSignMsg
	others []*rtypes.ValidatorSignMsg
}

type commitInfo struct {
	cp      *rtypes.CheckPoint
	crossTx *pt.RollupCrossTx
}

type crossTxInfo struct {
	txIndex        *pt.CrossTxIndex
	enterTimestamp int64
}
