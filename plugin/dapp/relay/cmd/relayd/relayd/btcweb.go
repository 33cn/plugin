// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package relayd

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/common/merkle"
	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/valyala/fasthttp"
)

type btcWeb struct {
	urlRoot    string
	httpClient *fasthttp.Client
}

func newBtcWeb() (BtcClient, error) {
	b := &btcWeb{
		urlRoot:    "https://blockchain.info",
		httpClient: &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	return b, nil
}

func (b *btcWeb) Start() error {
	return nil
}

func (b *btcWeb) Stop() error {
	return nil
}

func (b *btcWeb) GetBlockHeader(height uint64) (*ty.BtcHeader, error) {
	block, err := b.getBlock(height)
	if err != nil {
		return nil, err
	}
	return block.BtcHeader(), nil
}

func (b *btcWeb) getBlock(height uint64) (*block, error) {
	if height < 0 {
		return nil, errors.New("height < 0")
	}
	url := fmt.Sprintf("%s/block-height/%d?format=json", b.urlRoot, height)
	data, err := b.requestURL(url)
	if err != nil {
		return nil, err
	}
	var blocks = blocks{}
	err = json.Unmarshal(data, &blocks)
	if err != nil {
		return nil, err
	}
	block := blocks.Blocks[0]
	return &block, nil
}

func (b *btcWeb) GetLatestBlock() (*chainhash.Hash, uint64, error) {
	url := b.urlRoot + "/latestblock"
	data, err := b.requestURL(url)
	if err != nil {
		return nil, 0, err
	}
	var blocks = latestBlock{}
	err = json.Unmarshal(data, &blocks)
	if err != nil {
		return nil, 0, err
	}

	hash, err := chainhash.NewHashFromStr(blocks.Hash)
	if err != nil {
		return nil, 0, err
	}

	return hash, blocks.Height, nil
}

func (b *btcWeb) Ping() {
	hash, height, err := b.GetLatestBlock()
	if err != nil {
		log.Error("btcWeb ping", "error", err)
	}
	log.Info("btcWeb ping", "latest Hash: ", hash.String(), "latest height", height)
}

func (b *btcWeb) GetTransaction(hash string) (*ty.BtcTransaction, error) {
	url := b.urlRoot + "/rawtx/" + hash
	data, err := b.requestURL(url)
	if err != nil {
		return nil, err
	}
	var tx = transactionResult{}
	err = json.Unmarshal(data, &tx)
	if err != nil {
		return nil, err
	}
	return tx.BtcTransaction(), nil
}

func (b *btcWeb) GetSPV(height uint64, txHash string) (*ty.BtcSpv, error) {
	block, err := b.getBlock(height)
	if err != nil {
		return nil, err
	}
	var txIndex uint32
	txs := make([][]byte, 0, len(block.Tx))
	for index, tx := range block.Tx {
		if txHash == tx.Hash {
			txIndex = uint32(index)
		}
		hash, err := merkle.NewHashFromStr(tx.Hash)
		if err != nil {
			return nil, err
		}
		txs = append(txs, hash.CloneBytes())
	}
	proof := merkle.GetMerkleBranch(txs, txIndex)
	spv := &ty.BtcSpv{
		Hash:        txHash,
		Time:        block.Time,
		Height:      block.Height,
		BlockHash:   block.Hash,
		TxIndex:     txIndex,
		BranchProof: proof,
	}
	return spv, nil
}

func (b *btcWeb) requestURL(url string) ([]byte, error) {
	status, body, err := b.httpClient.Get(nil, url)
	if err != nil {
		return nil, err
	}
	if status != fasthttp.StatusOK {
		return nil, fmt.Errorf("%d", status)
	}
	return body, nil
}
