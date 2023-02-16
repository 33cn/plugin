package rollup

import (
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
	filterIdx := int32(0)
	for _, tx := range filterTxs {
		// 只记录跨链交易索引信息
		if !isCrossChainTx(tx) {
			continue
		}
		info := &crossTxInfo{
			enterTimestamp: timestamp,
			txIndex: &pt.CrossTxIndex{
				BlockHeight: mainHeight,
				FilterIndex: filterIdx,
			},
		}
		h.txIdxCache[shortHash(tx.Hash())] = info
		filterIdx++
	}
}

// 从缓存中删除平行链已打包执行的跨链交易, 返回跨链交易在主链区块的索引信息
func (h *crossTxHandler) removePackedCrossTx(hashList [][]byte) []*pt.CrossTxIndex {

	if len(hashList) <= 0 {
		return nil
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	idxList := make([]*pt.CrossTxIndex, 0, len(hashList))

	for _, hash := range hashList {

		short := shortHash(hash)
		info, ok := h.txIdxCache[short]
		var txIdx *pt.CrossTxIndex
		if ok {
			txIdx = info.txIndex
			// 缓存移除
			delete(h.txIdxCache, short)
		} else {
			rlog.Error("removePackedCrossTx not exist", "hash", hex.EncodeToString(hash))
			txIdx = &pt.CrossTxIndex{}
		}
		txIdx.TxHash = hash
		idxList = append(idxList, txIdx)
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
	reservedHeight := h.ru.cfg.ReservedMainHeight
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
		end := mainHeader.Height - reservedHeight
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

	for i := 0; i < len(txs); i++ {

		tx := txs[i]
		// 交易组情况, 组装
		if tx.GetGroupCount() > 1 {
			gtxs := &types.Transactions{Txs: txs[i : i+int(tx.GetGroupCount())]}
			tx = gtxs.Tx()
			i += int(tx.GetGroupCount()) - 1
		}
		// 平行链发送交易有转发主链逻辑, 指定转发到mempool需要调用特定接口
		api := h.ru.base.GetAPI().(*client.QueueProtocol)
		_, err := api.Send2Mempool(tx)
		// 发送至mempool失败, 可能情况是该交易已经打包但未提交状态, 此时节点重启导致
		if err != nil {
			rlog.Error("send2Mempool error", "mainHeight", mainHeight,
				"txHash", hex.EncodeToString(tx.Hash()), "err", err)
		}
	}
}
