package neutrino

import (
	"bytes"
	"encoding/hex"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightninglabs/neutrino"
	"github.com/lightninglabs/neutrino/headerfs"
	"runtime"
	"sync"
	"time"
)

type rgbx struct {
	client                   *neutrinoClient
	commitTxKey              crypto.PrivKey
	commitAddressType        int32
	pendingTxConfirmedHeight int64
	pendingTxPullHeight      int64

	lock         sync.RWMutex
	pendingCache *pendingTxCache
	rescanChan   chan *utxoRescanInfo
	commitChan   chan *utxoSpendInfo
}

type utxoRescanInfo struct {
	pendingTxHash string
	out           wire.OutPoint
	pkScript      []byte
	startHeight   int32
	startTime     int64
}

type utxoSpendInfo struct {
	spendingHeight     uint32
	spendingInputIndex uint32
	pendingTxHash      string
	spendingTxHash     string
	spendingTx         *wire.MsgTx
}

func newRGBX() *rgbx {
	r := &rgbx{}
	r.pendingCache = newPendingTxCache(128)
	r.rescanChan = make(chan *utxoRescanInfo, 1024)
	r.commitChan = make(chan *utxoSpendInfo, 1024)
	return r
}

func (r *rgbx) init(cli *neutrinoClient) {
	r.client = cli
	var err error
	r.commitAddressType, err = address.GetAddressType(cli.commitAddr)
	if err != nil {
		panic("invalid address type for authAccount config, " + cli.commitAddr)
	}

	r.commitTxKey = cli.getKeyFromWallet(cli.commitAddr)

	// 等待同步
	cli.waitUntilTrue("wait chain33 sync", cli.isChain33Sync)
	height := cli.getRgbxConfirmedHeight()
	if height == nil {
		panic("getRgbxConfirmedHeight")
	}

	r.pendingTxConfirmedHeight = height.GetData()
	r.pendingTxPullHeight = height.GetData() + 1
	// 等待轻节点同步
	cli.waitUntilTrue("wait neutrino sync", cli.neutrinoCS.IsCurrent)
	cli.waitUntilTrue("wait neutrino best block", func() bool {
		if r.client.getBestBlock() != nil {
			return true
		}
		return false
	})

	go r.handleCommitPendingTx()
	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go r.handleRescanUtxo()
	}

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
			txNum := len(txs.GetPendingList())
			log.Debug("pullPendingTx", "txNum", txNum, "err", err, "req", req)
			if err != nil || txNum == 0 {
				continue
			}
			for _, tx := range txs.GetPendingList() {

				if tx.Confirmed {
					log.Debug("tx already confirmed", "txHash", hex.EncodeToString(tx.TxHash))
					continue
				}
				hash, _ := chainhash.NewHash(tx.Utxo.Hash)
				uri := &utxoRescanInfo{
					pendingTxHash: hex.EncodeToString(tx.TxHash),
					out: wire.OutPoint{
						Hash:  *hash,
						Index: tx.Utxo.Index,
					},
					pkScript:    tx.Utxo.PkScript,
					startTime:   tx.Timestamp,
					startHeight: 0,
				}
				r.pendingCache.addTx(uri.pendingTxHash, tx)
				r.rescanChan <- uri
			}

			lastTx := txs.GetPendingList()[txNum-1]
			req.StartHeight = lastTx.TxBlockHeight
			req.StartIndex = lastTx.TxIndex
		}
	}
}

func (r *rgbx) rescanUtxo(info *utxoRescanInfo) (success bool) {

	bestBlk := r.client.getBestBlock()
	startOpt := neutrino.StartTime(time.Unix(info.startTime, 0))
	if info.startHeight > 0 {
		startOpt = neutrino.StartBlock(&headerfs.BlockStamp{Height: info.startHeight})
	}
	spendReport, err := r.client.neutrinoCS.GetUtxo(
		neutrino.WatchInputs(neutrino.InputWithScript{
			OutPoint: info.out,
			PkScript: info.pkScript,
		}),
		startOpt)

	info.startHeight = bestBlk.Height

	if err != nil {
		log.Error("rescanUtxo", "pendingTxHash", info.pendingTxHash, "err", err)
		return false
	}

	if spendReport.SpendingTx == nil {
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
		"spendingHash", spendInfo.spendingTxHash)
	return true

}

func (r *rgbx) handleRescanUtxo() {

	rescanArr := make([]*utxoRescanInfo, 0, 16)
	ticker := time.NewTicker(time.Minute * 5)
	for {

		select {

		case <-r.client.ctx.Done():
			return

		case info := <-r.rescanChan:

			if !r.rescanUtxo(info) {
				rescanArr = append(rescanArr, info)
			}

		case <-ticker.C:
			tempArr := rescanArr
			rescanArr = rescanArr[:0]
			for _, info := range tempArr {

				if !r.rescanUtxo(info) {
					rescanArr = append(rescanArr, info)
				}
			}
		}

	}
}

func (r *rgbx) createConfirmPayload(info *utxoSpendInfo) *rtypes.ConfirmTx {

	pendTx := r.pendingCache.getTx(info.pendingTxHash)
	minTxBlockHeight := r.pendingCache.getMinTxBlockHeight(pendTx.TxBlockHeight)
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
	_ = info.spendingTx.SerializeNoWitness(buf)
	proof.SpendingTx = buf.Bytes()

	return &rtypes.ConfirmTx{
		ActionType:           pendTx.ActionType,
		ConfirmedBlockHeight: minTxBlockHeight - 1,
		TxBlockHeight:        pendTx.TxBlockHeight,
		TxIndex:              pendTx.TxIndex,
		TxHash:               pendTx.TxHash,
		Proof:                proof,
	}
}

func (r *rgbx) commitPendingTx(confirm *rtypes.ConfirmTx) (success bool) {

	action := &rtypes.RgbxAction{
		Ty:    rtypes.TyConfirmAction,
		Value: &rtypes.RgbxAction_Confirm{Confirm: confirm},
	}
	confirmHash := hex.EncodeToString(confirm.TxHash)
	tx, err := types.CallCreateTransaction(rtypes.RgbxX, rtypes.NameConfirmAction, action)
	if err != nil {
		log.Error("commitPendingTx callCreateTransaction", "confirmHash", confirmHash, "err", err)
		return false
	}

	tx, err = types.FormatTx(r.client.chain33Api.GetConfig(), rtypes.RgbxX, tx)
	if err != nil {
		log.Error("commitPendingTx formatTx", "confirmHash", confirmHash, "err", err)
		return false
	}
	tx.Fee, _ = tx.GetRealFee(r.client.getProperFeeRate())
	tx.Sign(types.EncodeSignID(secp256k1.ID, r.commitAddressType), r.commitTxKey)
	txHash := hex.EncodeToString(tx.Hash())
	_, err = r.client.chain33Api.SendTx(tx)
	if err != nil {
		log.Error("commitPendingTx SendTx", "txHash", txHash, "confirmHash", confirmHash, "err", err)
		return false
	}

	log.Debug("commitPendingTx success", "txHash", txHash, "confirmHash", confirmHash)
	return true
}

func (r *rgbx) handleCommitPendingTx() {

	ticker := time.NewTicker(time.Second * 10)
	confirmArr := make([]*rtypes.ConfirmTx, 0, 8)
	for {

		select {

		case <-r.client.ctx.Done():
			return

		case info := <-r.commitChan:

			confirm := r.createConfirmPayload(info)
			if !r.commitPendingTx(confirm) {
				confirmArr = append(confirmArr, confirm)
			}

		case <-ticker.C:
			tempArr := confirmArr
			confirmArr = confirmArr[:0]
			for _, confirm := range tempArr {

				if !r.commitPendingTx(confirm) {
					confirmArr = append(confirmArr, confirm)
				}
			}
		}

	}

}
