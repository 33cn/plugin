package score

import (
	clog "github.com/33cn/chain33/common/log"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/mempool"
	"github.com/33cn/chain33/types"
)

func SetLogLevel(level string) {
	clog.SetLogLevel(level)
}

func DisableLog() {
	mlog.SetHandler(log.DiscardHandler())
}

//--------------------------------------------------------------------------------
// Module Mempool

type Mempool struct {
	subConfig subConfig
}

type subConfig struct {
	PoolCacheSize      int64 `json:"poolCacheSize"`
	MinTxFee           int64 `json:"minTxFee"`
	MaxTxNumPerAccount int64 `json:"maxTxNumPerAccount"`
	TimeParam          int64 `json:"timeParam"`
	PriceConstant      int64 `json:"priceConstant"`
	PricePower         int64 `json:"pricePower"`
}

func init() {
	drivers.Reg("trade", New)
}

//New 创建timeline cache 结构的 mempool
func New(cfg *types.Mempool, sub []byte) queue.Module {
	c := drivers.NewMempool(cfg)
	var subcfg subConfig
	types.MustDecode(sub, &subcfg)
	if subcfg.PoolCacheSize == 0 {
		subcfg.PoolCacheSize = cfg.PoolCacheSize
	}
	c.SetQueueCache(NewScoreQueue(subcfg))
	return c
}
