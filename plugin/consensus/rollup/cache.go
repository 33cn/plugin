package rollup

import (
	"bytes"
	"sync"

	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

type batchCache struct {
	lock            sync.RWMutex
	currCommitRound int64
	batchCache      map[int64]*rtypes.CommitBatch
	signMsgCache    map[int64]*validatorSignMsgSet
}

func newCache(commitRound int64) *batchCache {

	c := &batchCache{currCommitRound: commitRound}
	c.batchCache = make(map[int64]*rtypes.CommitBatch, 32)
	c.signMsgCache = make(map[int64]*validatorSignMsgSet, 32)
	return c
}

func (c *batchCache) addCommitBatch(batch *rtypes.CommitBatch) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.batchCache[batch.CommitRound] = batch
}

func (c *batchCache) getCommitBatch(round int64) *rtypes.CommitBatch {

	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.batchCache[round]
}

func (c *batchCache) addValidatorSign(isSelf bool, sign *rtypes.ValidatorSignMsg) {
	c.lock.Lock()
	defer c.lock.Unlock()
	set, ok := c.signMsgCache[sign.CommitRound]
	if !ok {
		set = &validatorSignMsgSet{others: make([]*rtypes.ValidatorSignMsg, 0, 8)}
		c.signMsgCache[sign.CommitRound] = set
	}
	if isSelf {
		set.self = sign
		return
	}

	// 检测是否有重复
	for _, other := range set.others {
		if bytes.Equal(sign.PubKey, other.PubKey) {
			return
		}
	}
	set.others = append(set.others, sign)
}

func (c *batchCache) getLocalSignSet(round int64) *validatorSignMsgSet {

	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.signMsgCache[round]
}

// clean already commit batch and sign
func (c *batchCache) remove(commitRound int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for i := c.currCommitRound; i <= commitRound; i++ {
		delete(c.signMsgCache, i)
		delete(c.batchCache, i)
	}
	c.currCommitRound = commitRound
}

//

func (c *batchCache) getAggregateBatch(round int64, aggreSign aggreSignFunc) *rtypes.CommitBatch {

	c.lock.RLock()
	defer c.lock.RUnlock()

	signSet := c.signMsgCache[round]
	pubs, aSign := aggreSign(signSet)
	if pubs == nil {
		return nil
	}
	batch := c.batchCache[round]

	batch.ValidatorPubs = pubs
	batch.AggregateValidatorSign = aSign
	return batch
}
