package rollup

import (
	"bytes"
	"encoding/hex"
	"sync"
	"time"

	"github.com/33cn/chain33/client"

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
	h.pulledHeight = r.cfg.StartHeight
	if status.CrossTxSyncedHeight > r.cfg.StartHeight {
		h.pulledHeight = status.CrossTxSyncedHeight
	}
}

func (h *crossTxHandler) addMainChainCrossTx(mainHeight int64, filterTxs []*types.Transaction) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.pulledHeight = mainHeight
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
func (h *crossTxHandler) removePackedCrossTx(hashList [][]byte) ([]*pt.CrossTxIndex, error) {

	if len(hashList) <= 0 {
		return nil, nil
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	idxList := make([]*pt.CrossTxIndex, 0, len(hashList))

	for _, hash := range hashList {

		short := shortHash(hash)
		info, ok := h.txIdxCache[short]
		if !ok {
			rlog.Error("removePackedCrossTx not exist", "hash", hex.EncodeToString(hash))
			return nil, types.ErrNotFound
		}

		// 缓存移除
		delete(h.txIdxCache, short)
		idxList = append(idxList, info.txIndex)

	}

	return idxList, nil
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
		if now-info.enterTimestamp >= 600 {

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
	defaultReservedMainHeight = 12
	maxPullIntervalOnce       = 128
)

func (h *crossTxHandler) pullCrossTx() {

	start := h.pulledHeight + 1
	for {

		select {
		case <-h.ru.ctx.Done():
			return
		default:
		}

		mainHeader, err := h.ru.mainChainGrpc.GetLastHeader(h.ru.ctx, &types.ReqNil{})
		if err != nil {
			rlog.Error("pullCrossTx", "start", start, "getLastHeader err", err)
			time.Sleep(time.Second * 5)
			continue
		}

		// 预留一定高度, 降低回滚概率
		end := mainHeader.Height - defaultReservedMainHeight
		if end < start {
			rlog.Debug("pullCrossTx wait for reserved block 1m")
			time.Sleep(time.Second * 5)
			continue
		}

		if end >= start+maxPullIntervalOnce {
			end = start + maxPullIntervalOnce - 1
		}

		rlog.Debug("pullCrossTx", "start", start, "end", end)
		details, err := h.ru.fetchCrossTx(start, end)
		if err != nil {
			rlog.Error("pullCrossTx", "start", start, "end", end, "fetchCrossTx err", err)
			time.Sleep(time.Second * 5)
			continue
		}

		for _, detail := range details.GetItems() {

			crossTxs := filterParaCrossTx(filterParaTx(h.ru.chainCfg, detail))
			h.addMainChainCrossTx(detail.Header.Height, crossTxs)
			h.send2Mempool(detail.Header.Height, crossTxs)
		}
		start = end + 1
	}
}

func (h *crossTxHandler) send2Mempool(mainHeight int64, txs []*types.Transaction) {

	if len(txs) == 0 {
		return
	}
	var errTxs []*types.Transaction
	for _, tx := range txs {

		// 交易组情况, 只需要第一笔发送至mempool
		if tx.GroupCount > 0 && !bytes.Equal(tx.Hash(), tx.Header) {
			rlog.Debug("send2Mempool txgroup", "mainHeight", mainHeight,
				"txHash", hex.EncodeToString(tx.Hash()))
			continue
		}
		api := h.ru.base.GetAPI().(*client.QueueProtocol)
		// 发送至mempool失败, 可能情况是该交易已经打包但未提交状态, 此时节点重启
		_, err := api.SendTx2Mempool(tx)
		if err != nil {
			errTxs = append(errTxs, tx)
			rlog.Error("send2Mempool error", "mainHeight", mainHeight,
				"txHash", hex.EncodeToString(tx.Hash()), "err", err)
		}
	}
}
