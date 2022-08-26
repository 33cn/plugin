package rollup

import (
	"bytes"
	"sync"

	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

type batchCache struct {
	lock sync.RWMutex

	batchCache   map[int64]*rolluptypes.CommitBatch
	signMsgCache map[int64]*validatorSignMsgSet
}

func newCache() *batchCache {

	c := &batchCache{}
	c.batchCache = make(map[int64]*rolluptypes.CommitBatch, 32)
	c.signMsgCache = make(map[int64]*validatorSignMsgSet, 32)
	return c
}

func (c *batchCache) addCommitBatch(batch *rolluptypes.CommitBatch) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.batchCache[batch.CommitRound] = batch
}

func (c *batchCache) getCommitBatch(round int64) *rolluptypes.CommitBatch {

	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.batchCache[round]
}

func (c *batchCache) addSignMsg(msg []byte, sign *rolluptypes.ValidatorSignMsg) {

	c.lock.Lock()
	defer c.lock.Unlock()

	set, ok := c.signMsgCache[sign.CommitRound]

	if !ok {
		set = &validatorSignMsgSet{msg: msg}
		c.signMsgCache[sign.CommitRound] = set
	}

	// 检测是否有重复
	for _, pub := range set.pubs {
		if bytes.Equal(sign.PubKey, pub) {
			return
		}
	}
	set.pubs = append(set.pubs, sign.PubKey)
	set.signs = append(set.signs, sign.Signature)
}

func (c *batchCache) getSignSet(round int64) *validatorSignMsgSet {

	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.signMsgCache[round]
}

// remove already commit batch and sign
func (c *batchCache) remove() {

}

//

func (c *batchCache) getAggregateBatch(round int64, aggreFunc aggreSignFunc) *rolluptypes.CommitBatch {

	c.lock.RLock()
	defer c.lock.RUnlock()

	signSet := c.signMsgCache[round]
	pubs, aggreSign := aggreFunc(signSet)
	if pubs == nil {
		return nil
	}
	batch := c.batchCache[round]

	batch.ValidatorPubs = pubs
	batch.AggregateValidatorSign = aggreSign
	return batch
}
