package rollup

import (
	"encoding/hex"
	"sync"
	"time"

	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"

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

func (h *crossTxHandler) init(r *RollUp, status *rtypes.RollupStatus) {

	h.ru = r
	h.txIdxCache = make(map[string]*crossTxInfo, 32)
	h.pulledHeight = r.cfg.BootHeight
	if status.CrossTxSyncedHeight > r.cfg.BootHeight {
		h.pulledHeight = status.CrossTxSyncedHeight
	}
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
func (h *crossTxHandler) removePackedCrossTx(hashList [][]byte) []*pt.CrossTxIndex {

	h.lock.Lock()
	defer h.lock.Unlock()
	idxList := make([]*pt.CrossTxIndex, 0, len(hashList))

	for _, hash := range hashList {

		short := shortHash(hash)
		info, ok := h.txIdxCache[short]
		if !ok {
			rlog.Error("removePackedCrossTx not exist", "hash", hex.EncodeToString(hash))
			continue
		}

		// 缓存移除
		delete(h.txIdxCache, short)
		idxList = append(idxList, info.txIndex)

	}

	return idxList
}

// 刷新平行链同步主链跨链交易区块高度, 根据缓存中跨链交易记录决定
func (h *crossTxHandler) refreshSyncedHeight() int64 {

	h.lock.Lock()
	defer h.lock.Unlock()

	syncedHeight := h.pulledHeight
	var expiredTxs []string
	// 缓存中存在, 表示该跨链交易还未被打包, 返回上一个高度
	now := types.Now().Unix()
	for hash, info := range h.txIdxCache {

		// 10min 在缓存中未打包, 标记为过期
		if now-info.enterTimestamp > 600 {

			rlog.Error("refreshSyncedHeight expired", "shortHash", hex.EncodeToString([]byte(hash)),
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

const (
	reservedMainHeight  = 12
	maxPullIntervalOnce = 128
)

func (h *crossTxHandler) pullCrossTx() {

	ticker := time.NewTicker(time.Minute)
	start := h.pulledHeight + 1
	for {

		select {

		case <-h.ru.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:

			mainHeader, err := h.ru.mainChainGrpc.GetLastHeader(h.ru.ctx, nil)
			if err != nil {
				rlog.Error("pullCrossTx", "start", start, "getLastHeader err", err)
				continue
			}

			// 预留一定高度, 降低回滚概率
			end := mainHeader.Height - reservedMainHeight

			if end < start {
				continue
			}

			if end > start+maxPullIntervalOnce {
				end = start + maxPullIntervalOnce
			}

			details, err := h.ru.fetchCrossTx(start, end)
			if err != nil {
				rlog.Error("pullCrossTx", "start", start, "end", end, "fetchCrossTx err", err)
				continue
			}

			start = end + 1

			for _, detail := range details.GetItems() {

				crossTxs := filterParaCrossTx(filterParaTx(h.ru.chainCfg, detail))
				h.addMainChainCrossTx(detail.Header.Height, crossTxs)
				h.send2Mempool(detail.Header.Height, crossTxs)
			}

		}

	}
}

func (h *crossTxHandler) send2Mempool(mainHeight int64, txs []*types.Transaction) {

	//TODO 处理交易组情况

	var errTxs []*types.Transaction

	for _, tx := range txs {

		reply, err := h.ru.base.GetAPI().SendTx(tx)
		if err != nil || !reply.GetIsOk() {
			errTxs = append(errTxs, tx)
			rlog.Error("send2Mempool", "mainHeight", mainHeight, "txHash", hex.EncodeToString(tx.Hash()))
		}
	}

	h.removeErrTxs(errTxs)

}
