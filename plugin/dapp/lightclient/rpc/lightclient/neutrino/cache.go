package neutrino

import (
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"sync"
)

type pendingTxCache struct {
	lock         sync.RWMutex
	pendingCache map[string]*rtypes.PendingTx
}

func newPendingTxCache(size int) *pendingTxCache {
	return &pendingTxCache{pendingCache: make(map[string]*rtypes.PendingTx, size)}
}

func (p *pendingTxCache) addTx(hash string, value *rtypes.PendingTx) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.pendingCache[hash] = value
}

func (p *pendingTxCache) getTx(hash string) *rtypes.PendingTx {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.pendingCache[hash]
}

func (p *pendingTxCache) removeTx(hash string) *rtypes.PendingTx {
	p.lock.Lock()
	defer p.lock.Unlock()
	tx := p.pendingCache[hash]
	delete(p.pendingCache, hash)
	return tx
}

func (p *pendingTxCache) getMinTxBlockHeight(minHeight int64) int64 {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for _, tx := range p.pendingCache {
		if tx.TxBlockHeight < minHeight {
			minHeight = tx.TxBlockHeight
		}
	}

	return minHeight
}
