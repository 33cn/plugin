// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package relayd

import (
	"errors"
	"strconv"
	"sync"
	"time"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/common/merkle"
	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

// BtcClient interface
type BtcClient interface {
	Start() error
	Stop() error
	GetLatestBlock() (*chainhash.Hash, uint64, error)
	GetBlockHeader(height uint64) (*ty.BtcHeader, error)
	GetSPV(height uint64, txHash string) (*ty.BtcSpv, error)
	GetTransaction(hash string) (*ty.BtcTransaction, error)
	Ping()
}

type (
	blockStamp struct {
		Height int32
		Hash   chainhash.Hash
	}

	blockMeta struct {
		blockStamp
		Time time.Time
	}

	clientConnected   struct{}
	blockConnected    blockMeta
	blockDisconnected blockMeta
)

type params struct {
	*chaincfg.Params
	RPCClientPort string
	RPCServerPort string
}

var mainNetParams = params{
	Params:        &chaincfg.MainNetParams,
	RPCClientPort: "8334",
	RPCServerPort: "8332",
}

type btcdClient struct {
	rpcClient           *rpcclient.Client
	connConfig          *rpcclient.ConnConfig
	chainParams         *chaincfg.Params
	reconnectAttempts   int
	enqueueNotification chan interface{}
	dequeueNotification chan interface{}
	currentBlock        chan *blockStamp
	quit                chan struct{}
	wg                  sync.WaitGroup
	started             bool
	quitMtx             sync.Mutex
}

func newBtcd(config *rpcclient.ConnConfig, reconnectAttempts int) (BtcClient, error) {
	if reconnectAttempts < 0 {
		return nil, errors.New("ReconnectAttempts must be positive")
	}
	client := &btcdClient{
		connConfig:          config,
		chainParams:         mainNetParams.Params,
		reconnectAttempts:   reconnectAttempts,
		enqueueNotification: make(chan interface{}),
		dequeueNotification: make(chan interface{}),
		currentBlock:        make(chan *blockStamp),
		quit:                make(chan struct{}),
	}
	ntfnCallbacks := &rpcclient.NotificationHandlers{
		OnClientConnected:   client.onClientConnect,
		OnBlockConnected:    client.onBlockConnected,
		OnBlockDisconnected: client.onBlockDisconnected,
	}
	rpcClient, err := rpcclient.New(client.connConfig, ntfnCallbacks)
	if err != nil {
		return nil, err
	}
	client.rpcClient = rpcClient
	return client, nil
}

func (b *btcdClient) Start() error {
	err := b.rpcClient.Connect(b.reconnectAttempts)
	if err != nil {
		return err
	}

	// Verify that the server is running on the expected network.
	net, err := b.rpcClient.GetCurrentNet()
	if err != nil {
		b.rpcClient.Disconnect()
		return err
	}
	if net != b.chainParams.Net {
		b.rpcClient.Disconnect()
		return errors.New("mismatched networks")
	}

	b.quitMtx.Lock()
	b.started = true
	b.quitMtx.Unlock()

	b.wg.Add(1)
	go b.handler()
	return nil
}

func (b *btcdClient) Stop() error {
	b.quitMtx.Lock()
	select {
	case <-b.quit:
	default:
		close(b.quit)
		b.rpcClient.Shutdown()

		if !b.started {
			close(b.dequeueNotification)
		}
	}
	b.quitMtx.Unlock()
	return nil
}

func (b *btcdClient) WaitForShutdown() {
	b.rpcClient.WaitForShutdown()
	b.wg.Wait()
}

func (b *btcdClient) Notifications() <-chan interface{} {
	return b.dequeueNotification
}

func (b *btcdClient) BlockStamp() (*blockStamp, error) {
	select {
	case bs := <-b.currentBlock:
		return bs, nil
	case <-b.quit:
		return nil, errors.New("disconnected")
	}
}

func (b *btcdClient) onClientConnect() {
	select {
	case b.enqueueNotification <- clientConnected{}:
	case <-b.quit:
	}
}

func (b *btcdClient) onBlockConnected(hash *chainhash.Hash, height int32, time time.Time) {
	select {
	case b.enqueueNotification <- blockConnected{
		blockStamp: blockStamp{
			Hash:   *hash,
			Height: height,
		},
		Time: time,
	}:
	case <-b.quit:
	}
}

func (b *btcdClient) onBlockDisconnected(hash *chainhash.Hash, height int32, time time.Time) {
	select {
	case b.enqueueNotification <- blockDisconnected{
		blockStamp: blockStamp{
			Hash:   *hash,
			Height: height,
		},
		Time: time,
	}:
	case <-b.quit:
	}
}

func (b *btcdClient) handler() {
	hash, height, err := b.rpcClient.GetBestBlock()
	if err != nil {
		b.Stop()
		b.wg.Done()
		return
	}

	bs := &blockStamp{Hash: *hash, Height: height}
	var notifications []interface{}
	enqueue := b.enqueueNotification
	var dequeue chan interface{}
	var next interface{}
	pingChan := time.After(time.Minute)
out:
	for {
		select {
		case n, ok := <-enqueue:
			if !ok {
				// If no notifications are queued for handling,
				// the queue is finished.
				if len(notifications) == 0 {
					break out
				}
				// nil channel so no more reads can occur.
				enqueue = nil
				continue
			}
			if len(notifications) == 0 {
				next = n
				dequeue = b.dequeueNotification
			}
			notifications = append(notifications, n)
			pingChan = time.After(time.Minute)

		case dequeue <- next:
			if n, ok := next.(blockConnected); ok {
				bs = &blockStamp{
					Height: n.Height,
					Hash:   n.Hash,
				}
			}

			notifications[0] = nil
			notifications = notifications[1:]
			if len(notifications) != 0 {
				next = notifications[0]
			} else {
				// If no more notifications can be enqueued, the
				// queue is finished.
				if enqueue == nil {
					break out
				}
				dequeue = nil
			}

		case <-pingChan:
			type sessionResult struct {
				err error
			}
			sessionResponse := make(chan sessionResult, 1)
			go func() {
				_, err := b.rpcClient.Session()
				sessionResponse <- sessionResult{err}
			}()

			select {
			case resp := <-sessionResponse:
				if resp.err != nil {
					b.Stop()
					break out
				}
				pingChan = time.After(time.Minute)

			case <-time.After(time.Minute):
				b.Stop()
				break out
			}

		case b.currentBlock <- bs:

		case <-b.quit:
			break out
		}
	}

	b.Stop()
	close(b.dequeueNotification)
	b.wg.Done()
}

// POSTClient creates the equivalent HTTP POST rpcclient.Client.
func (b *btcdClient) POSTClient() (*rpcclient.Client, error) {
	configCopy := *b.connConfig
	configCopy.HTTPPostMode = true
	return rpcclient.New(&configCopy, nil)
}

func (b *btcdClient) GetSPV(height uint64, txHash string) (*ty.BtcSpv, error) {
	hash, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return nil, err
	}

	ret, err := b.rpcClient.GetRawTransactionVerbose(hash)
	if err != nil {
		return nil, err
	}

	blockHash, err := chainhash.NewHashFromStr(ret.BlockHash)
	if err != nil {
		return nil, err
	}

	block, err := b.rpcClient.GetBlockVerbose(blockHash)
	if err != nil {
		return nil, err
	}
	var txIndex uint32
	txs := make([][]byte, 0, len(block.Tx))
	for index, tx := range block.Tx {
		if txHash == tx {
			txIndex = uint32(index)
		}
		hash, err := merkle.NewHashFromStr(tx)
		if err != nil {
			return nil, err
		}
		txs = append(txs, hash.CloneBytes())
	}
	proof := merkle.GetMerkleBranch(txs, txIndex)
	spv := &ty.BtcSpv{
		Hash:        txHash,
		Time:        block.Time,
		Height:      uint64(block.Height),
		BlockHash:   block.Hash,
		TxIndex:     txIndex,
		BranchProof: proof,
	}
	return spv, nil
}

func (b *btcdClient) GetTransaction(hash string) (*ty.BtcTransaction, error) {
	txHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, err
	}
	tx, err := b.rpcClient.GetRawTransactionVerbose(txHash)
	if err != nil {
		return nil, err
	}

	blockHash, err := chainhash.NewHashFromStr(tx.BlockHash)
	if err != nil {
		return nil, err
	}

	header, err := b.rpcClient.GetBlockHeaderVerbose(blockHash)
	if err != nil {
		return nil, err
	}

	btxTx := &ty.BtcTransaction{}
	btxTx.Hash = hash
	btxTx.Time = tx.Time
	btxTx.BlockHeight = uint64(header.Height)
	vin := make([]*ty.Vin, len(tx.Vin))
	for index, in := range tx.Vin {
		var v ty.Vin
		// v.Address = in.
		v.Value = uint64(in.Vout) * 1e8
		vin[index] = &v
	}
	btxTx.Vin = vin
	vout := make([]*ty.Vout, len(tx.Vout))
	for index, in := range tx.Vout {
		var out ty.Vout
		out.Value = uint64(in.Value) * 1e8
		out.Address = in.ScriptPubKey.Addresses[0]
		vout[index] = &out
	}
	btxTx.Vout = vout
	return btxTx, nil
}

func (b *btcdClient) GetBlockHeader(height uint64) (*ty.BtcHeader, error) {
	hash, err := b.rpcClient.GetBlockHash(int64(height))
	if err != nil {
		return nil, err
	}
	header, err := b.rpcClient.GetBlockHeaderVerbose(hash)
	if err != nil {
		return nil, err
	}

	bits, err := strconv.ParseInt(header.Bits, 16, 32)
	if err != nil {
		return nil, err
	}

	h := &ty.BtcHeader{
		Hash:          header.Hash,
		Confirmations: uint64(header.Confirmations),
		Height:        uint64(header.Height),
		MerkleRoot:    header.MerkleRoot,
		Time:          header.Time,
		Nonce:         header.Nonce,
		Bits:          bits,
		PreviousHash:  header.PreviousHash,
		NextHash:      header.NextHash,
		Version:       uint32(header.Version),
	}
	return h, nil

}

func (b *btcdClient) GetLatestBlock() (*chainhash.Hash, uint64, error) {
	hash, height, err := b.rpcClient.GetBestBlock()
	return hash, uint64(height), err
}

func (b *btcdClient) Ping() {
	hash, height, err := b.rpcClient.GetBestBlock()
	if err != nil {
		log.Error("btcdClient ping", "error", err)
	}

	log.Info("btcdClient ping", "latest Hash: ", hash.String(), "latest height", height)
}
