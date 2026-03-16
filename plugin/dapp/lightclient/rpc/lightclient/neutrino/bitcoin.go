// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package neutrino

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/walletdb"
)

// headerPublisher 提交比特币区块头到链上
func (n *neutrinoClient) submitBitcoinHeaders() {

	nextSubmitHeight := n.cfg.BtcHeaderStartHeight
	n.waitUntilDone("wait getChain33BtcLastHeader", func() bool {
		if lastHeader := n.getChain33BtcLastHeader(); lastHeader != nil {
			if lastHeader.GetHeight() > 0 {
				nextSubmitHeight = lastHeader.GetHeight() + 1
			}
			return true
		}
		return false
	}, 0)

	log.Info("submitBitcoinHeaders start", "nextSubmitHeight", nextSubmitHeight)

	interval := n.cfg.BtcBlockInterval/3 + 1
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()
	const batchSize = 16
	for {
		select {
		case <-n.ctx.Done():
			return

		case <-ticker.C:

			best := n.getBestBlock()
			if best == nil {
				continue
			}

			confirmedHeight := uint64(best.Height) - uint64(n.cfg.BlockConfirmations)
			headers := make([]*ltypes.BtcHeader, 0, batchSize)
			for height := nextSubmitHeight; height <= confirmedHeight && len(headers) < batchSize; height++ {
				header, err := n.neutrinoCS.BlockHeaders.FetchHeaderByHeight(uint32(height))
				if err != nil {
					log.Error("submitBitcoinHeaders FetchHeaderByHeight", "height", height, "err", err)
					break
				}
				headerHash := header.BlockHash().String()
				headers = append(headers, &ltypes.BtcHeader{
					Hash:          headerHash,
					Confirmations: uint64(best.Height) - height + 1,
					Height:        height,
					Version:       uint32(header.Version),
					MerkleRoot:    header.MerkleRoot.String(),
					Time:          header.Timestamp.Unix(),
					Nonce:         uint64(header.Nonce),
					Bits:          int64(header.Bits),
					PreviousHash:  header.PrevBlock.String(),
				})
			}
			if len(headers) > 0 {
				payload := &ltypes.BtcHeaders{Headers: headers}
				n.submitMainchainTxUntilSuccess(ltypes.LightclientX, ltypes.NameBtcHeadersAction, payload)
				lastHeader := headers[len(headers)-1]
				nextSubmitHeight = lastHeader.GetHeight() + 1
				log.Debug("submitBitcoinHeaders", "commitHeight", lastHeader.GetHeight(),
					"hash", lastHeader.GetHash(), "nextSubmitHeight", nextSubmitHeight)
			}

		}
	}
}

func (n *neutrinoClient) getChain33BtcLastHeader() *ltypes.BtcHeader {
	reply, err := n.mainChainGrpc.QueryChain(n.ctx, &types.ChainExecutor{
		Driver:   ltypes.LightclientX,
		FuncName: "GetBtcLastHeader",
	})
	if err != nil {
		log.Error("getChain33BtcLastHeader", "query err", err)
		return nil
	}
	data := &ltypes.BtcHeader{}
	err = types.Decode(reply.GetMsg(), data)
	if err != nil {
		log.Error("getChain33BtcLastHeader", "decode err", err)
		return nil
	}
	return data
}

// depositWatcher 监听比特币充值交易, 并向chain33主链提交rgbx deposit交易
func (n *neutrinoClient) depositWatcher() {

	depositChan := n.bw.GetDepositChannel()
	retryTicker := time.NewTicker(time.Second * 30)
	var retryList []*btcPendingTx
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-retryTicker.C:
			temp := retryList
			retryList = retryList[:0]
			for _, pendingTx := range temp {
				if err := n.commitDepositTx(pendingTx); err != nil {
					log.Error("depositWatcher commitDepositTx retry", "txHash", pendingTx.txHash.String(), "err", err)
					retryList = append(retryList, pendingTx)
				}
			}
		case pendingTx := <-depositChan:

			// 如果chain33DepositAddress为空，则使用第一个输入的utxo地址
			if pendingTx.chain33DepositAddress == "" {
				firstInputUtxo := pendingTx.tx.TxIn[0].PreviousOutPoint
				pendingTx.chain33DepositAddress = rtypes.FormatUtxo(firstInputUtxo.Hash.String(), firstInputUtxo.Index)
			}
			if pendingTx.depositAmount <= 0 {
				log.Error("depositWatcher invalid deposit amount", "txHash", pendingTx.txHash.String(),
					"amount", pendingTx.depositAmount)
				continue
			}

			if err := n.commitDepositTx(pendingTx); err != nil {
				log.Error("depositWatcher commitDepositTx", "txHash", pendingTx.txHash.String(), "err", err)
				retryList = append(retryList, pendingTx)
			}
		}
	}
}
func (n *neutrinoClient) commitDepositTx(pendingTx *btcPendingTx) error {
	if state := n.getDepositState(pendingTx.txHash[:]); bytes.Equal(state, depositStatusProcessed) {
		log.Debug("commitDepositTx already processed", "txHash", pendingTx.txHash.String())
		n.bw.removePendingTx(pendingTx.txHash)
		return nil
	}
	spv, err := n.bw.buildTxExistenceProof(pendingTx)
	if err != nil {
		log.Error("commitDepositTx buildTxExistenceProof", "txHash", pendingTx.txHash.String(), "err", err)
		return err
	}
	buf := bytes.NewBuffer(make([]byte, 0, pendingTx.tx.SerializeSizeStripped()))
	if err = pendingTx.tx.SerializeNoWitness(buf); err != nil {
		log.Error("commitDepositTx SerializeNoWitness", "txHash", pendingTx.txHash.String(), "err", err)
		return err
	}
	deposit := &rtypes.DepositAsset{
		Amount:         int64(pendingTx.depositAmount),
		DepositAddress: pendingTx.chain33DepositAddress,
		AssetSymbol:    rtypes.BTCSymbol,
		TxProof: &rtypes.BtcTxProof{
			TxData:      buf.Bytes(),
			BlockHash:   pendingTx.blockHash.String(),
			BlockHeight: uint64(pendingTx.blockHeight),
			TxIndex:     spv.GetTxIndex(),
			MerkleProof: spv.GetBranchProof(),
		},
	}
	n.submitMainchainTxUntilSuccess(rtypes.RgbxX, rtypes.NameDepositAssetAction, deposit)
	if err = n.setDepositState(pendingTx.txHash[:], depositStatusProcessed); err != nil {
		log.Error("commitDepositTx setDepositState processed", "txHash", pendingTx.txHash.String(), "err", err)
	}
	n.bw.removePendingTx(pendingTx.txHash)
	log.Debug("commitDepositTx submit deposit success", "btxHash", pendingTx.txHash.String(),
		"depositAddr", deposit.GetDepositAddress(), "amount", deposit.GetAmount())
	return nil
}

func pending2WithdrawRequest(chain33Pending *rtypes.PendingTx) *withdrawRequest {
	return &withdrawRequest{
		chain33WithDrawHash: chain33Pending.GetTxHash(),
		amount:              btcutil.Amount(chain33Pending.GetAmount()),
		feeRate:             btcutil.Amount(chain33Pending.GetFeeRate()),
		toAddress:           chain33Pending.GetTargetAddress(),
	}
}

func (n *neutrinoClient) processWithdrawRequest(req *withdrawRequest) error {

	txHash := hex.EncodeToString(req.chain33WithDrawHash)
	tx, inputAmounts, lockedUTXOs, err := n.bw.buildWithdrawTx(req)
	if err != nil {
		log.Error("processWithdrawRequest buildWithdrawTx", "txHash", txHash, "err", err)
		return err
	}
	stickyUTXO := lockedUTXOs[len(lockedUTXOs)-1]
	expectedHash := n.getExpectedWithdrawHash(stickyUTXO.OutPoint.String())
	if len(expectedHash) > 0 && !bytes.Equal(expectedHash, req.chain33WithDrawHash) {
		log.Error("processWithdrawRequest sticky input mismatch", "expected", hex.EncodeToString(expectedHash),
			"actual", txHash, "stickyOutPoint", stickyUTXO.OutPoint.String())
		return fmt.Errorf("invalid sticky input")
	}

	// 提现交易构建后，则和最后一个utxo强绑定，后续不能更改
	if req.stickyUTXO == nil {
		req.stickyUTXO = lockedUTXOs[len(lockedUTXOs)-1]
		if err = n.setWithdrawStickyUTXO(req.chain33WithDrawHash, req.stickyUTXO); err != nil {
			log.Error("processWithdrawRequest setWithdrawStickyUTXO", "txHash", txHash, "stickyUTXO", req.stickyUTXO.OutPoint.String(), "err", err)
		}
	}
	// 主节点也进行验证，保证各节点执行相同的逻辑
	err = n.tss.validateWithdrawTx(tx, inputAmounts, req)
	if err != nil {
		n.bw.releaseUTXOsExcept(lockedUTXOs, req.stickyUTXO)
		log.Error("processSignBtcTx validateWithdrawTx", "txHash", txHash, "err", err)
		return err
	}
	btcTxHash := tx.TxHash().String()
	if err = n.tss.processSignBtcTx(tx, transactionTypeWithdraw, inputAmounts, req.chain33WithDrawHash); err != nil {
		n.bw.releaseUTXOsExcept(lockedUTXOs, req.stickyUTXO)
		log.Error("processWithdrawRequest processSignBtcTx", "txHash", txHash, "btcTxHash", btcTxHash, "err", err)
		return err
	}

	if err = n.bw.broadcastTransaction(tx, btcTxHash); err != nil {
		n.bw.releaseUTXOsExcept(lockedUTXOs, req.stickyUTXO)
		log.Error("processWithdrawRequest broadcastTransaction", "txHash", txHash, "btcTxHash", btcTxHash, "err", err)
		return err
	}
	n.bw.addPendingTx(&btcPendingTx{
		tx:                    tx,
		submitTime:            types.Now(),
		confirmations:         0,
		blockHeight:           -1,
		txHash:                tx.TxHash(),
		txType:                transactionTypeWithdraw,
		withdrawAddress:       req.toAddress,
		chain33WithdrawTxHash: req.chain33WithDrawHash,
	})

	if err = n.setWithdrawState(req.chain33WithDrawHash, withdrawStatusSent); err != nil {
		log.Error("processWithdrawRequest setWithdrawState", "txHash", txHash, "btcTxHash", btcTxHash, "err", err)
	}
	log.Debug("processWithdrawRequest success", "txHash", txHash, "btcTxHash", btcTxHash)
	return nil
}

type confirmWithdraw struct {
	btcPending          *btcPendingTx
	pendingTxBlockIndex *rtypes.TxBlockIndex
	confirmTx           *rtypes.ConfirmTx
}

var (
	withdrawStateBucket      = []byte("rgbx-withdraw-state")
	withdrawStatusSent       = []byte("broadcasted")
	withdrawStatusConfirmed  = []byte("confirmed")
	withdrawStickyUTXOBucket = []byte("rgbx-withdraw-sticky-utxo")

	depositStateBucket     = []byte("rgbx-deposit-state")
	depositStatusProcessed = []byte("processed")
)

func (n *neutrinoClient) getWithdrawState(chain33TxHash []byte) []byte {
	var data []byte
	err := walletdb.View(n.neutrinoCfg.Database, func(tx walletdb.ReadTx) error {
		bucket := tx.ReadBucket(withdrawStateBucket)
		if bucket == nil {
			return walletdb.ErrBucketNotFound
		}
		data = bucket.Get(chain33TxHash)
		return nil
	})
	if err != nil && !errors.Is(err, walletdb.ErrBucketNotFound) {
		log.Error("getWithdrawState", "txHash", hex.EncodeToString(chain33TxHash), "err", err)
	}
	return data
}

func (n *neutrinoClient) setWithdrawState(txHash []byte, status []byte) error {

	return walletdb.Update(n.neutrinoCfg.Database, func(tx walletdb.ReadWriteTx) error {
		bucket, err := tx.CreateTopLevelBucket(withdrawStateBucket)
		if err != nil {
			return err
		}
		return bucket.Put(txHash, status)
	})
}

func encodeOutPoint(op *wire.OutPoint) []byte {
	return []byte(op.String())
}

func decodeOutPoint(data []byte) (*wire.OutPoint, error) {
	return wire.NewOutPointFromString(string(data))
}

func (n *neutrinoClient) getWithdrawStickyUTXO(chain33TxHash []byte) *UTXO {
	var data []byte
	err := walletdb.View(n.neutrinoCfg.Database, func(tx walletdb.ReadTx) error {
		bucket := tx.ReadBucket(withdrawStickyUTXOBucket)
		if bucket == nil {
			return walletdb.ErrBucketNotFound
		}
		data = bucket.Get(chain33TxHash)
		return nil
	})
	if err != nil && !errors.Is(err, walletdb.ErrBucketNotFound) {
		log.Error("getWithdrawStickyUTXO", "txHash", hex.EncodeToString(chain33TxHash), "err", err)
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	u := &rtypes.Utxo{}
	err = types.Decode(data, u)
	if err != nil {
		log.Error("getWithdrawStickyUTXO decode", "txHash", hex.EncodeToString(chain33TxHash), "err", err)
		return nil
	}
	outPoint, err := wire.NewOutPointFromString(u.OutPoint)
	if err != nil {
		log.Error("getWithdrawStickyUTXO NewOutPointFromString", "txHash", hex.EncodeToString(chain33TxHash), "err", err)
		return nil
	}
	return &UTXO{
		OutPoint: *outPoint,
		Amount:   btcutil.Amount(u.Amount),
		PkScript: u.PkScript,
	}
}

func (n *neutrinoClient) getExpectedWithdrawHash(outPoint string) []byte {
	var data []byte
	err := walletdb.View(n.neutrinoCfg.Database, func(tx walletdb.ReadTx) error {
		bucket := tx.ReadBucket(withdrawStickyUTXOBucket)
		if bucket == nil {
			return walletdb.ErrBucketNotFound
		}
		data = bucket.Get([]byte(outPoint))
		return nil
	})
	if err != nil && !errors.Is(err, walletdb.ErrBucketNotFound) {
		log.Error("getExpectedWithdrawHash", "outPoint", outPoint, "err", err)
		return nil
	}
	return data
}

func (n *neutrinoClient) setWithdrawStickyUTXO(chain33TxHash []byte, utxo *UTXO) error {
	u := &rtypes.Utxo{
		OutPoint: utxo.OutPoint.String(),
		Amount:   int64(utxo.Amount),
		PkScript: utxo.PkScript,
	}
	return walletdb.Update(n.neutrinoCfg.Database, func(tx walletdb.ReadWriteTx) error {
		bucket, err := tx.CreateTopLevelBucket(withdrawStickyUTXOBucket)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(utxo.OutPoint.String()), chain33TxHash)
		if err != nil {
			return err
		}
		return bucket.Put(chain33TxHash, types.Encode(u))
	})
}

func (n *neutrinoClient) clearWithdrawFirstInput(chain33TxHash []byte) error {
	return walletdb.Update(n.neutrinoCfg.Database, func(tx walletdb.ReadWriteTx) error {
		bucket := tx.ReadWriteBucket(withdrawStickyUTXOBucket)
		if bucket == nil {
			return nil
		}
		return bucket.Delete(chain33TxHash)
	})
}

func (n *neutrinoClient) getDepositState(txHash []byte) []byte {
	var data []byte
	err := walletdb.View(n.neutrinoCfg.Database, func(tx walletdb.ReadTx) error {
		bucket := tx.ReadBucket(depositStateBucket)
		if bucket == nil {
			return walletdb.ErrBucketNotFound
		}
		data = bucket.Get(txHash)
		return nil
	})
	if err != nil && !errors.Is(err, walletdb.ErrBucketNotFound) {
		log.Error("getDepositState", "txHash", hex.EncodeToString(txHash), "err", err)
	}
	return data
}

func (n *neutrinoClient) setDepositState(txHash []byte, status []byte) error {
	return walletdb.Update(n.neutrinoCfg.Database, func(tx walletdb.ReadWriteTx) error {
		bucket, err := tx.CreateTopLevelBucket(depositStateBucket)
		if err != nil {
			return err
		}
		return bucket.Put(txHash, status)
	})
}

func (n *neutrinoClient) getPendingTxBlockIndex(txHash []byte) *rtypes.TxBlockIndex {
	hashStr := hex.EncodeToString(txHash)
	pendingTx := n.rgbx.pendingCache.getTx(hashStr)
	if pendingTx != nil {
		return &rtypes.TxBlockIndex{
			BlockHeight: pendingTx.TxBlockHeight,
			TxIndex:     pendingTx.TxIndex,
		}
	}
	txDetail, err := n.getTxDetail(txHash)
	if err != nil {
		log.Error("getPendingTxBlockIndex getTxDetail", "txHash", hashStr, "err", err)
		return nil
	}
	txBlockIndex := &rtypes.TxBlockIndex{BlockHeight: txDetail.GetHeight(), TxIndex: txDetail.GetIndex()}
	return txBlockIndex
}

// withdrawalProcessor 监听chain33主链上比特币提现请求, 构造提现交易到比特币网络，并向chain33主链提交rgbx confirm交易
func (n *neutrinoClient) withdrawalProcessor() {
	withdrawalChan := n.bw.GetWithdrawChannel()
	retryTicker := time.NewTicker(time.Second * 30)
	defer retryTicker.Stop()
	withdrawReqChan := n.withdrawReqChan
	withdrawReqList := make([]*withdrawRequest, 0, 16)
	confirmRetryList := make([]*confirmWithdraw, 0, 16)

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-retryTicker.C:
			tempWithdraws := withdrawReqList
			withdrawReqList = withdrawReqList[:0]
			for _, p := range tempWithdraws {
				if err := n.processWithdrawRequest(p); err != nil {
					withdrawReqList = append(withdrawReqList, p)
				}
			}
			tempConfirms := confirmRetryList
			confirmRetryList = confirmRetryList[:0]
			for _, p := range tempConfirms {
				if !n.processWithdrawConfirm(p) {
					confirmRetryList = append(confirmRetryList, p)
				}
			}
		case pendingBtc := <-withdrawalChan: // 向chain33主链提交确认提现交易
			confirm := &confirmWithdraw{
				btcPending: pendingBtc,
			}
			if !n.processWithdrawConfirm(confirm) {
				confirmRetryList = append(confirmRetryList, confirm)
			}

		case pending := <-withdrawReqChan: // 向btc链提交提现交易
			if len(n.getWithdrawState(pending.GetTxHash())) > 0 {
				log.Debug("withdrawalProcessor hasWithdrawState", "txHash", hex.EncodeToString(pending.GetTxHash()))
				continue
			}
			req := pending2WithdrawRequest(pending)
			req.stickyUTXO = n.getWithdrawStickyUTXO(pending.GetTxHash())
			if err := n.processWithdrawRequest(req); err != nil {
				withdrawReqList = append(withdrawReqList, req)
			}
		}
	}
}

func (n *neutrinoClient) commitWithdrawConfirm(confirm *rtypes.ConfirmTx, confirmHash string) (string, error) {

	txHash, err := n.submitMainchainTx(rtypes.RgbxX, rtypes.NameConfirmAction, confirm)
	if err != nil && !strings.Contains(err.Error(), "already confirmed") {
		return "", err
	}
	n.rgbx.pendingCache.removeTx(confirmHash)
	if err := n.setWithdrawState(confirm.TxHash, withdrawStatusConfirmed); err != nil {
		log.Error("commitWithdrawConfirm setWithdrawState", "txHash", txHash,
			"confirmHash", confirmHash, "err", err)
	}
	return txHash, nil
}

func (n *neutrinoClient) buildWithdrawConfirm(btcPending *btcPendingTx, pendingTxBlockIndex *rtypes.TxBlockIndex) *rtypes.ConfirmTx {
	if pendingTxBlockIndex == nil {
		return nil
	}
	spv, err := n.bw.buildTxExistenceProof(btcPending)
	if err != nil {
		log.Error("buildWithdrawConfirmPayload buildTxExistenceProof", "btcTxHash", btcPending.txHash.String(),
			"chain33WithdrawTxHash", hex.EncodeToString(btcPending.chain33WithdrawTxHash), "err", err)
		return nil
	}
	buf := bytes.NewBuffer(make([]byte, 0, btcPending.tx.SerializeSizeStripped()))
	if err = btcPending.tx.SerializeNoWitness(buf); err != nil {
		log.Error("buildWithdrawConfirmPayload SerializeNoWitness", "btcTxHash", btcPending.txHash.String(),
			"chain33WithdrawTxHash", hex.EncodeToString(btcPending.chain33WithdrawTxHash), "err", err)
		return nil
	}
	minPendingHeight := n.rgbx.pendingCache.getMinPendingHeight()
	return &rtypes.ConfirmTx{
		ActionType:           rtypes.TyWithDrawAsset,
		ConfirmedBlockHeight: minPendingHeight - 1,
		TxBlockHeight:        pendingTxBlockIndex.GetBlockHeight(),
		TxIndex:              pendingTxBlockIndex.GetTxIndex(),
		TxHash:               btcPending.chain33WithdrawTxHash,
		BtcTxProof: &rtypes.BtcTxProof{
			TxData:      buf.Bytes(),
			BlockHash:   btcPending.blockHash.String(),
			BlockHeight: uint64(btcPending.blockHeight),
			TxIndex:     spv.GetTxIndex(),
			MerkleProof: spv.GetBranchProof(),
		},
	}
}

func (n *neutrinoClient) processWithdrawConfirm(confirm *confirmWithdraw) bool {

	if confirm.pendingTxBlockIndex == nil {
		confirm.pendingTxBlockIndex = n.getPendingTxBlockIndex(confirm.btcPending.chain33WithdrawTxHash)
	}

	if confirm.confirmTx == nil {
		confirm.confirmTx = n.buildWithdrawConfirm(confirm.btcPending, confirm.pendingTxBlockIndex)
	}
	if confirm.confirmTx == nil {
		return false
	}
	confirmHash := hex.EncodeToString(confirm.confirmTx.TxHash)
	if state := n.getWithdrawState(confirm.confirmTx.GetTxHash()); bytes.Equal(state, withdrawStatusConfirmed) {
		n.rgbx.pendingCache.removeTx(confirmHash)
		n.bw.removePendingTx(confirm.btcPending.txHash)
		if err := n.clearWithdrawFirstInput(confirm.confirmTx.GetTxHash()); err != nil {
			log.Error("processWithdrawConfirm clearWithdrawFirstInput", "confirmHash", confirmHash, "err", err)
		}
		log.Debug("processWithdrawConfirm already confirmed local state", "confirmHash", confirmHash)
		return true
	}

	txHash, err := n.commitWithdrawConfirm(confirm.confirmTx, confirmHash)
	if err != nil {
		log.Error("processWithdrawConfirm commitWithdrawConfirm", "txHash", txHash, "confirmHash", confirmHash, "err", err)
		return false
	}
	n.bw.removePendingTx(confirm.btcPending.txHash)
	if err := n.clearWithdrawFirstInput(confirm.confirmTx.GetTxHash()); err != nil {
		log.Error("processWithdrawConfirm clearWithdrawFirstInput", "confirmHash", confirmHash, "err", err)
	}
	log.Debug("processWithdrawConfirm success", "txHash", txHash,
		"btcTxHash", confirm.btcPending.txHash.String(), "confirmHash", confirmHash)
	return true

}
