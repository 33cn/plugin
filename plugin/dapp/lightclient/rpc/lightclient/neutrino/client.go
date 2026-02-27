// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package neutrino integrate btc light client neutrino
package neutrino

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	"github.com/lightninglabs/neutrino/headerfs"

	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
	"github.com/lightninglabs/neutrino"
)

var log = log15.New("module", "lightclient.neutrino")

var _ lightclient.Lighter = &neutrinoClient{}

func init() {

	lightclient.Register("neutrino", newClient)
}

func newClient() lightclient.Lighter {
	return &neutrinoClient{}
}

type neutrinoClient struct {
	ctx               context.Context
	qclient           queue.Client
	chain33Api        client.QueueProtocolAPI
	mainChainGrpc     types.Chain33Client
	cfg               config
	commitAddressType int32
	commitAddr        string
	commitKey         crypto.PrivKey
	neutrinoCfg       neutrino.Config
	tss               *tssService
	neutrinoCS        *neutrino.ChainService
	bw                *btcWallet
	rgbx              *rgbx
	bestBlock         *headerfs.BlockStamp
	lock              sync.RWMutex
	chain33FeeRate    int64
	withdrawReqChan   chan *rtypes.PendingTx
}

// Init init client context
func (n *neutrinoClient) Init(ctx context.Context, q queue.Queue, cfg *lightclient.Config) error {

	n.ctx = ctx
	n.qclient = q.Client()
	n.chain33Api, _ = client.New(n.qclient, nil)
	n.chain33FeeRate = 100000
	n.commitAddr = cfg.CommitAddr
	commitAddressType, err := address.GetAddressType(n.commitAddr)
	if err != nil {
		panic("invalid address type for authAccount config, " + n.commitAddr)
	}
	n.commitAddressType = commitAddressType
	if cfg.CommitKey != "" {
		_, n.commitKey = getPrivKey(secp256k1.Name, cfg.CommitKey)
	}
	subCfg, _ := json.Marshal(cfg.Neutrino)
	types.MustDecode(subCfg, &n.cfg)
	chainCfg := q.GetConfig()
	n.mainChainGrpc, err = grpcclient.NewMainChainClient(chainCfg, "")
	if err != nil {
		panic("init main chain grpc client err:" + err.Error())
	}
	n.tss = newTssService(n)
	if !n.cfg.IsOfficialNode {
		return nil
	}
	n.withdrawReqChan = make(chan *rtypes.PendingTx, 256)
	err = n.initNeutrinoConfig(chainCfg)
	if err != nil {
		log.Error("Init", "initNeutrinoConfig error", err)
		return err
	}
	cs, err := neutrino.NewChainService(n.neutrinoCfg)
	if err != nil {
		log.Error("Init", "NewChainService error", err)
		_ = n.neutrinoCfg.Database.Close()
		return err
	}
	n.neutrinoCS = cs
	bw, err := newBtcWallet(n)
	if err != nil {
		log.Error("Init", "newBtcWallet error", err)
		return err
	}
	n.bw = bw
	n.rgbx = newRGBX()
	return nil

}

// Start starting routine
func (n *neutrinoClient) Start() {

	n.initCommitKey()
	n.tss.start()
	go n.subMsg()
	if !n.cfg.IsOfficialNode {
		return
	}
	if err := n.neutrinoCS.Start(); err != nil {
		log.Error("Start", "neutrinoCS start error", err)
		_ = n.neutrinoCfg.Database.Close()
		panic(err)
	}
	if err := n.bw.start(); err != nil {
		log.Error("Start", "btcwallet start error", err)
		n.bw.stop()
		panic(err)
	}
	go n.cleanUp()
	go n.handleBestBlock()
	go n.submitBitcoinHeaders()
	go n.depositWatcher()
	go n.withdrawalProcessor()
	n.rgbx.start(n)
}

// handle subscription messages
func (n *neutrinoClient) subMsg() {

	n.qclient.Sub(moduleName)
	for {

		select {
		case <-n.ctx.Done():
			return
		case msg := <-n.qclient.Recv():

			if msg == nil {
				log.Error("SubMsg", "err", "receive nil msg")
				return
			}
			data, ok := msg.Data.(*types.TopicData)
			if msg.Ty == types.EventReceiveSubData && ok && data.Topic == tssSignNotifyTopic {
				n.tss.subChan <- data
			}
			log.Error("SubMsg receive invalid msg", "ty", msg.Ty, "ok", ok)
		}
	}
}

func (n *neutrinoClient) cleanUp() {

	<-n.ctx.Done()
	if err := n.neutrinoCS.Stop(); err != nil {
		log.Error("cleanUp Unable to stop neutrino server", "err", err)
	}
	n.bw.stop()
	if err := n.neutrinoCfg.Database.Close(); err != nil {
		log.Error("cleanUp Unable to close neutrino db", "err", err)
	}
}

func (n *neutrinoClient) handleBestBlock() {

	n.syncBestBlock()
	interval := time.Duration(n.cfg.BtcBlockInterval)/3 + 1
	ticker := time.NewTicker(time.Second * interval)
	for {

		select {

		case <-n.ctx.Done():
			return
		case <-ticker.C:

			n.syncBestBlock()
		}
	}
}

func (n *neutrinoClient) syncBestBlock() {
	blk, err := n.neutrinoCS.BestBlock()
	if err != nil {
		log.Error("syncBestBlock", "err", err)
		return
	}
	log.Debug("syncBestBlock", "height", blk.Height, "hash", blk.Hash.String())
	n.setBestBlock(blk)
}

func (n *neutrinoClient) getBestBlock() *headerfs.BlockStamp {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.bestBlock
}

func (n *neutrinoClient) setBestBlock(blk *headerfs.BlockStamp) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if blk != nil {
		n.bestBlock = blk
	}
}
