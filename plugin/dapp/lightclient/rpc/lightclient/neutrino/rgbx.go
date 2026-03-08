package neutrino

import (
	"bytes"
	"encoding/hex"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightninglabs/neutrino"
	"github.com/lightninglabs/neutrino/headerfs"
)

type rgbx struct {
	client                   *neutrinoClient
	commitTxKey              crypto.PrivKey
	pendingTxConfirmedHeight int64
	pendingTxPullHeight      int64

	lock         sync.RWMutex
	pendingCache *pendingTxCache
	rescanChan   chan *utxoRescanInfo
	commitChan   chan *utxoSpendInfo
}

type utxoRescanInfo struct {
	pendingTxHash   string
	outPrint        string
	out             wire.OutPoint
	pkScript        []byte
	startHeight     int32
	startTime       int64
	confirmedHeight int32
}

type utxoSpendInfo struct {
	spendingHeight     uint32
	spendingInputIndex uint32
	pendingTxHash      string
	spendingTxHash     string
	spendingTx         *wire.MsgTx
	timeout            bool
}

func newRGBX() *rgbx {
	r := &rgbx{}
	r.pendingCache = newPendingTxCache(128)
	r.rescanChan = make(chan *utxoRescanInfo, 1024)
	r.commitChan = make(chan *utxoSpendInfo, 1024)
	return r
}

func (r *rgbx) start(cli *neutrinoClient) {
	log.Debug("init rgbx service start")
	r.client = cli
	r.commitTxKey = cli.getCommitKey()

	log.Debug("wait getRgbxConfirmedHeight")
	cli.waitUntilDone("wait getRgbxConfirmedHeight", func() bool {
		if height := cli.getRgbxConfirmedHeight(); height != nil {
			r.pendingTxConfirmedHeight = height.GetData()
			if r.pendingTxConfirmedHeight > 0 {
				r.pendingTxPullHeight = height.GetData() + 1
			}
			return true
		}
		return false
	}, 0)

	// 等待轻节点同步
	log.Debug("wait neutrino sync", "currHeight", cli.getBestBlock().Height)
	cli.waitUntilDone("wait neutrino sync", func() bool {
		return cli.neutrinoCS.IsCurrent()
	}, 0)
	log.Debug("wait neutrino best block")
	cli.waitUntilDone("wait neutrino best block", func() bool {
		return cli.getBestBlock() != nil
	}, 0)

	go r.pullPendingTx()
	go r.handleCommitPendingTx()
	for i := 0; i < runtime.NumCPU(); i++ {
		go r.handleRescanUtxo()
	}

	log.Debug("init rgbx service done")
}

func (r *rgbx) pullPendingTx() {
	ticker := time.NewTicker(time.Second * 10)
	req := &rtypes.ReqListPendingTx{
		StartHeight: r.pendingTxPullHeight,
		StartIndex:  0,
		Count:       128,
	}
	for {
		select {

		case <-r.client.ctx.Done():
			return
		case <-ticker.C:

			txs, err := r.client.getRgbxPendingTxs(req)
			if err != nil {
				log.Debug("pullPendingTx", "err", err, "req", req)
				continue
			}
			txNum := len(txs.GetPendingList())
			if txNum == 0 {
				continue
			}

			log.Debug("pullPendingTx", "txNum", txNum, "req", req)
			for _, tx := range txs.GetPendingList() {

				txHash := hex.EncodeToString(tx.TxHash)
				if tx.Confirmed {
					log.Debug("tx already confirmed", "txHash", txHash)
					continue
				}
				r.pendingCache.addTx(txHash, tx)
				if tx.GetActionType() == rtypes.TyWithDrawAsset {
					r.client.withdrawReqChan <- tx
					continue
				}

				hash, err := chainhash.NewHashFromStr(tx.Utxo.Hash)
				if err != nil {
					log.Error("pullPendingTx", "txHash", txHash, "err", err)
					r.pendingCache.removeTx(txHash)
					continue
				}
				r.rescanChan <- &utxoRescanInfo{
					pendingTxHash: txHash,
					out: wire.OutPoint{
						Hash:  *hash,
						Index: tx.Utxo.Index,
					},
					outPrint:    tx.Utxo.ToString(),
					pkScript:    tx.Utxo.PkScript,
					startTime:   tx.Timestamp,
					startHeight: 0,
				}
			}

			lastTx := txs.GetPendingList()[txNum-1]
			req.StartHeight = lastTx.TxBlockHeight
			req.StartIndex = lastTx.TxIndex
		}
	}
}

func (r *rgbx) rescanUtxo(info *utxoRescanInfo) (success bool) {
	bestBlk := r.client.getBestBlock()

	// get start height from startTime
	for height := bestBlk.Height; info.startHeight == 0 && height > 0; height-- {

		header, err := r.client.neutrinoCS.BlockHeaders.FetchHeaderByHeight(uint32(height))
		if err != nil {
			log.Error("rescanUtxo", "pendHash", info.pendingTxHash,
				"height", height, "FetchHeaderByHeight err", err)
			return false
		}

		if header.Timestamp.Unix() < info.startTime {
			info.startHeight = height + 1
		}
	}
	if info.startHeight > bestBlk.Height || info.confirmedHeight > bestBlk.Height {
		return false
	}

	log.Debug("rescanUtxo", "startHeight", info.startHeight,
		"pendHash", info.pendingTxHash, "utxo", info.outPrint)

	spendReport, err := r.client.neutrinoCS.GetUtxo(
		neutrino.WatchInputs(neutrino.InputWithScript{
			OutPoint: info.out,
			PkScript: info.pkScript,
		}),
		neutrino.StartBlock(&headerfs.BlockStamp{Height: info.startHeight}))
	if err != nil {
		log.Error("rescanUtxo", "pendingTxHash", info.pendingTxHash,
			"utxo", info.outPrint, "err", err)
		return false
	}

	if spendReport == nil || spendReport.SpendingTx == nil {
		info.startHeight = bestBlk.Height + 1
		return false
	}

	// make sure enough confirmations
	if spendReport.SpendingTxHeight+r.client.cfg.BlockConfirmations > uint32(bestBlk.Height) {
		info.confirmedHeight = int32(spendReport.SpendingTxHeight + r.client.cfg.BlockConfirmations)
		log.Debug("rescanUtxo confirmations", "bestHeight", bestBlk.Height,
			"confirmedHeight", info.confirmedHeight, "utxo", info.outPrint)
		return false
	}

	spendInfo := &utxoSpendInfo{
		spendingTxHash:     spendReport.SpendingTx.TxHash().String(),
		pendingTxHash:      info.pendingTxHash,
		spendingTx:         spendReport.SpendingTx,
		spendingHeight:     spendReport.SpendingTxHeight,
		spendingInputIndex: spendReport.SpendingInputIndex,
	}

	r.commitChan <- spendInfo
	log.Debug("rescanUtxo", "pendingHash", spendInfo.pendingTxHash,
		"spendingHash", spendInfo.spendingTxHash, "utxo", info.outPrint)
	return true
}

func (r *rgbx) handleRescanUtxo() {
	rescanArr := make([]*utxoRescanInfo, 0, 16)
	interval := r.client.cfg.BtcBlockInterval/2 + 1
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for {
		select {

		case <-r.client.ctx.Done():
			return

		case info := <-r.rescanChan:

			if !r.rescanUtxo(info) && !r.mayTimeout(info) {
				rescanArr = append(rescanArr, info)
			}

		case <-ticker.C:
			tempArr := rescanArr
			rescanArr = rescanArr[:0]
			for _, info := range tempArr {
				if !r.rescanUtxo(info) && !r.mayTimeout(info) {
					rescanArr = append(rescanArr, info)
				}
			}
		}
	}
}

func (r *rgbx) mayTimeout(info *utxoRescanInfo) bool {
	if r.client.cfg.MaxUtxoRescanTime == 0 ||
		types.Now().Unix() < info.startTime+r.client.cfg.MaxUtxoRescanTime {
		return false
	}

	spendInfo := &utxoSpendInfo{
		pendingTxHash: info.pendingTxHash,
		timeout:       true,
	}
	log.Debug("mayTimeout", "startTime", info.startTime, "pendingHash", info.pendingTxHash,
		"utxo", info.outPrint)
	r.commitChan <- spendInfo
	return true
}

func (r *rgbx) createConfirmPayload(info *utxoSpendInfo, pendTx *rtypes.PendingTx) (*rtypes.ConfirmTx, error) {

	minPendingHeight := r.pendingCache.getMinPendingHeight()
	confirm := &rtypes.ConfirmTx{
		ActionType:           pendTx.ActionType,
		ConfirmedBlockHeight: minPendingHeight - 1,
		TxBlockHeight:        pendTx.TxBlockHeight,
		TxIndex:              pendTx.TxIndex,
		TxHash:               pendTx.TxHash,
	}

	if info.timeout {
		confirm.Timeout = true
		return confirm, nil
	}

	proof := &rtypes.UtxoSpendingProof{
		SpendingInputIdx: info.spendingInputIndex,
		OpRetOutputIdx:   -1,
	}
	for idx, out := range info.spendingTx.TxOut {
		if out.PkScript[0] == txscript.OP_RETURN {
			proof.OpRetOutputIdx = int32(idx)
			proof.OpRetOutputPkScript = out.PkScript
		}
	}
	buf := bytes.NewBuffer(make([]byte, 0, info.spendingTx.SerializeSizeStripped()))
	err := info.spendingTx.SerializeNoWitness(buf)
	if err != nil {
		log.Error("createConfirmPayload", "pendingTxHash", info.pendingTxHash, "serialize spending tx err", err)
		return nil, err
	}
	proof.SpendingTx = buf.Bytes()
	confirm.UtxoProof = proof

	return confirm, nil
}

func (r *rgbx) commitPendingTx(request *confrimRequest) error {

	confirmHash := request.spendingInfo.pendingTxHash
	confirm, err := r.createConfirmPayload(request.spendingInfo, request.pendTx)
	if err != nil {
		log.Error("commitPendingTx createConfirmPayload", "confirmHash", confirmHash, "err", err)
		return err
	}
	err = r.client.submitMainchainTx(rtypes.RgbxX, rtypes.NameConfirmAction, confirm)
	if err != nil && !strings.Contains(err.Error(), "already confirmed") {
		log.Error("commitPendingTx submitMainchainTx", "confirmHash", confirmHash, "err", err)
		return err
	}

	r.pendingCache.removeTx(confirmHash)
	log.Debug("commitPendingTx success", "confirmHash", confirmHash)
	return nil
}

type confrimRequest struct {
	spendingInfo *utxoSpendInfo
	pendTx       *rtypes.PendingTx
}

func (r *rgbx) handleCommitPendingTx() {
	ticker := time.NewTicker(time.Second * 10)
	retryList := make([]*confrimRequest, 0, 8)
	for {
		select {

		case <-r.client.ctx.Done():
			return

		case info := <-r.commitChan:

			pendTx := r.pendingCache.getTx(info.pendingTxHash)
			if pendTx == nil {
				log.Error("handleCommitPendingTx pendingTx not found", "pendingTxHash", info.pendingTxHash)
				continue
			}
			request := &confrimRequest{spendingInfo: info, pendTx: pendTx}
			if err := r.commitPendingTx(request); err != nil {
				log.Error("handleCommitPendingTx commitPendingTx", "confirmHash", info.pendingTxHash, "err", err)
				retryList = append(retryList, request)
			}

		case <-ticker.C:
			tempRequests := retryList
			retryList = retryList[:0]
			for _, request := range tempRequests {
				if err := r.commitPendingTx(request); err != nil {
					log.Error("handleCommitPendingTx retry", "confirmHash", request.spendingInfo.pendingTxHash, "commitPendingTx err", err)
					retryList = append(retryList, request)
				}
			}
		}
	}
}
