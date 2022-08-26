package rollup

// Config rollup 配置
type Config struct {
	ValidatorBlsKey string
	CommitInterval  int32
}

type validatorSignMsgSet struct {
	msg   []byte
	pubs  [][]byte
	signs [][]byte
}
