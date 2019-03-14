package score

import (
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/mempool"
	"github.com/33cn/chain33/types"
)

//--------------------------------------------------------------------------------
// Module Mempool

type subConfig struct {
	PoolCacheSize int64 `json:"poolCacheSize"`
	TimeParam     int64 `json:"timeParam"`
	PriceConstant int64 `json:"priceConstant"`
	PricePower    int64 `json:"pricePower"`
	ProperFee     int64 `json:"properFee"`
}

func init() {
	drivers.Reg("score", New)
}

//New 创建score cache 结构的 mempool
func New(cfg *types.Mempool, sub []byte) queue.Module {
	c := drivers.NewMempool(cfg)
	var subcfg subConfig
	types.MustDecode(sub, &subcfg)
	if subcfg.PoolCacheSize == 0 {
		subcfg.PoolCacheSize = cfg.PoolCacheSize
	}
	if subcfg.ProperFee == 0 {
		subcfg.ProperFee = cfg.MinTxFee
	}
	c.SetQueueCache(NewQueue(subcfg))
	return c
}
