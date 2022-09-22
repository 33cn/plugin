package rollup

import rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"

// Config rollup 配置
type Config struct {
	SignTxKey      string
	CommitBlsKey   string
	CommitInterval int32
}

type validatorSignMsgSet struct {
	self   *rtypes.ValidatorSignMsg
	others []*rtypes.ValidatorSignMsg
}
