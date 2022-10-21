package rollup

import (
	"encoding/hex"
	"sync"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

// cross tx handler
type ctxHandler struct {
	lock sync.RWMutex
	// main chain blockHeight => cross tx list
	ctxCache map[int64]*crossTxInfo

	// tx hash => tx index info
	txIdxCache map[string]*pt.CrossTxIndex

	pulledHeight int64
}

func newCtxHandler() *ctxHandler {

	h := &ctxHandler{}

	h.ctxCache = make(map[int64]*crossTxInfo, 16)
	h.txIdxCache = make(map[string]*pt.CrossTxIndex, 32)

	return h
}

func (h *ctxHandler) addMainChainCrossTx(mainHeight int64, filterTxs []*types.Transaction) *crossTxInfo {
	h.lock.Lock()
	defer h.lock.Unlock()
	info, ok := h.ctxCache[mainHeight]
	if !ok {
		info = &crossTxInfo{}
	}
	info.txList = filterTxs
	info.validatorSyncCount++
	h.ctxCache[mainHeight] = info

	for idx, tx := range filterTxs {
		h.txIdxCache[shortHash(tx.Hash())] = &pt.CrossTxIndex{
			BlockHeight: mainHeight,
			FilterIndex: int32(idx),
		}
	}

	return info
}

// 跨链交易由主链拉取, 记录其他验证节点拉取主链高度信息
func (h *ctxHandler) addValidatorSyncHeight(height int64) *crossTxInfo {

	h.lock.Lock()
	defer h.lock.Unlock()
	info, ok := h.ctxCache[height]
	if !ok {
		info = &crossTxInfo{}
	}
	info.validatorSyncCount++
	return info
}

// 从缓存中删除平行链已打包执行的跨链交易, 返回跨链交易在主链区块的索引信息
func (h *ctxHandler) removePackedCrossTx(ctxList []*types.Transaction) []*pt.CrossTxIndex {

	h.lock.Lock()
	defer h.lock.Unlock()
	idxList := make([]*pt.CrossTxIndex, 0, len(ctxList))

	for _, ctx := range ctxList {

		short := shortHash(ctx.Hash())
		idxInfo, ok := h.txIdxCache[short]
		if !ok {
			rlog.Error("removePackedCrossTx not exist", "hash", hex.EncodeToString(ctx.Hash()))
			continue
		}

		// 缓存移除
		delete(h.txIdxCache, short)
		idxList = append(idxList, idxInfo)
		ctxInfo := h.ctxCache[idxInfo.BlockHeight]
		ctxInfo.packedTxCount++
		// 所有跨链交易已在平行链打包, 删除缓存
		if ctxInfo.packedTxCount >= int32(len(ctxInfo.txList)) {
			delete(h.ctxCache, idxInfo.BlockHeight)
		}

	}

	return idxList
}

func (h *ctxHandler) getCrossTxSyncedHeight() int64 {

	h.lock.RLock()
	defer h.lock.RUnlock()

	syncedHeight := h.pulledHeight

	for height := range h.ctxCache {
		if height < syncedHeight {
			syncedHeight = height
		}
	}

	return syncedHeight
}
