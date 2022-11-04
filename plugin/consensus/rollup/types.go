package rollup

import (
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

// Config rollup 配置
type Config struct {
	SignTxKey      string
	CommitBlsKey   string
	CommitInterval int32
	// BootHeight 指定平行链启动时主链的高度
	BootHeight int64
}

type validatorSignMsgSet struct {
	self   *rtypes.ValidatorSignMsg
	others []*rtypes.ValidatorSignMsg
}

type commitInfo struct {
	cp      *rtypes.CheckPoint
	crossTx *pt.CommitRollup
}

type crossTxInfo struct {
	txIndex        *pt.CrossTxIndex
	enterTimestamp int64
}
