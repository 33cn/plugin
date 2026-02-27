// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package neutrino

import (
	"bytes"
	"time"

	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
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
	var retryList []*pendingTx
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
func (n *neutrinoClient) commitDepositTx(pendingTx *pendingTx) error {
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
			BlockHash:   pendingTx.blockHash.CloneBytes(),
			BlockHeight: int64(pendingTx.blockHeight),
			TxIndex:     int64(spv.GetTxIndex()),
			MerkleProof: spv.GetBranchProof(),
		},
	}
	n.submitMainchainTxUntilSuccess(rtypes.RgbxX, rtypes.NameDepositAssetAction, deposit)
	log.Debug("commitDepositTx submit deposit success", "txHash", pendingTx.txHash.String(),
		"depositAddr", deposit.GetDepositAddress(), "amount", deposit.GetAmount())
	return nil
}

// withdrawalProcessor 监听chain33主链上比特币提现请求, 构造提现交易到比特币网络，并向chain33主链提交rgbx confirm交易
func (n *neutrinoClient) withdrawalProcessor() {

	withdrawalChan := n.bw.GetWithdrawChannel()
	for {
		select {
		case <-n.ctx.Done():
			return
		case pendingTx := <-withdrawalChan:

			if pendingTx == nil {
				continue
			}
		}
	}
}
