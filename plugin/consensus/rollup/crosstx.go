package rollup

import (
	"encoding/hex"
	"sync"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

// cross tx handler
type crossTxHandler struct {
	ru   *RollUp
	lock sync.RWMutex

	// tx hash => cross tx index info
	txIdxCache map[string]*crossTxInfo

	pulledHeight int64
}

func (h *crossTxHandler) init() {

	h.txIdxCache = make(map[string]*crossTxInfo, 32)

}

func (h *crossTxHandler) addMainChainCrossTx(mainHeight int64, filterTxs []*types.Transaction) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.pulledHeight = mainHeight
	if len(filterTxs) == 0 {
		return
	}
	timestamp := types.Now().Unix()
	for idx, tx := range filterTxs {
		info := &crossTxInfo{
			enterTimestamp: timestamp,
			txIndex: &pt.CrossTxIndex{
				BlockHeight: mainHeight,
				FilterIndex: int32(idx),
			},
		}
		h.txIdxCache[shortHash(tx.Hash())] = info
	}
}

// 从缓存中删除平行链已打包执行的跨链交易, 返回跨链交易在主链区块的索引信息
func (h *crossTxHandler) removePackedCrossTx(ctxList []*types.Transaction) []*pt.CrossTxIndex {

	h.lock.Lock()
	defer h.lock.Unlock()
	idxList := make([]*pt.CrossTxIndex, 0, len(ctxList))

	for _, ctx := range ctxList {

		short := shortHash(ctx.Hash())
		info, ok := h.txIdxCache[short]
		if !ok {
			rlog.Error("removePackedCrossTx not exist", "hash", hex.EncodeToString(ctx.Hash()))
			continue
		}

		// 缓存移除
		delete(h.txIdxCache, short)
		idxList = append(idxList, info.txIndex)

	}

	return idxList
}

// 刷新平行链同步主链跨链交易区块高度, 根据缓存中跨链交易记录决定
func (h *crossTxHandler) refreshCrossTxSyncedHeight() int64 {

	h.lock.Lock()
	defer h.lock.Unlock()

	syncedHeight := h.pulledHeight
	var expiredTxs []string
	// 缓存中存在, 表示该跨链交易还未被打包, 返回上一个高度
	now := types.Now().Unix()
	for hash, info := range h.txIdxCache {

		// 10min 在缓存中未打包, 标记为过期
		if now-info.enterTimestamp > 600 {

			rlog.Error("refreshCrossTxSyncedHeight expired", "shortHash", hex.EncodeToString([]byte(hash)),
				"txIndex", info.txIndex.String())
			expiredTxs = append(expiredTxs, hash)
			continue
		}

		if info.txIndex.BlockHeight <= syncedHeight {
			syncedHeight = info.txIndex.BlockHeight - 1
		}

	}
	// 移除标记为过期的交易
	for _, hash := range expiredTxs {
		delete(h.txIdxCache, hash)
	}

	return syncedHeight
}

func (h *crossTxHandler) removeErrTxs(errList []*types.Transaction) {

	if len(errList) <= 0 {
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	for _, tx := range errList {

		rlog.Error("removeErrTxs", "txHash", hex.EncodeToString(tx.Hash()))
		delete(h.txIdxCache, shortHash(tx.Hash()))
	}
}

func (h *crossTxHandler) pullCrossTx() {

}

func (h *crossTxHandler) send2Mempool() {

}
