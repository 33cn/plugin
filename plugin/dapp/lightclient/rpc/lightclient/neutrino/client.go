// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package neutrino integrate btc light client neutrino
package neutrino

import (
	"context"
	"github.com/lightninglabs/neutrino/headerfs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcwallet/walletdb"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
	"github.com/lightninglabs/neutrino"
)

var log = log15.New("module", "neutrino")

var _ lightclient.Lighter = &neutrinoClient{}

func init() {

	lightclient.Register("neutrino", newClient)
}

func newClient() lightclient.Lighter {
	return &neutrinoClient{}
}

type neutrinoClient struct {
	ctx            context.Context
	chain33Api     client.QueueProtocolAPI
	cfg            config
	commitAddr     string
	neutrinoCfg    neutrino.Config
	neutrinoCS     *neutrino.ChainService
	bestBlock      *headerfs.BlockStamp
	lock           sync.Mutex
	chain33FeeRate int64
}

// defaultBlockCacheSize is the size (in bytes) of blocks that will be
// keep in memory if no size is specified.
const defalutBlockCacheSize = 20 * 1024 * 1024 //20 MB

type config struct {
	MaxPeer        int      `json:"maxPeer"`
	BlockCacheSize uint64   `json:"blockCacheSize"`
	NetType        string   `json:"netType"`
	AddPeers       []string `json:"addPeers"`
	ConnectPeers   []string `json:"connectPeers"`
}

func (c config) getChainParams() chaincfg.Params {

	if strings.Contains(c.NetType, "simple") {
		return chaincfg.SimNetParams
	} else if strings.Contains(c.NetType, "test") {
		return chaincfg.TestNet3Params
	} else if strings.Contains(c.NetType, "regress") {
		return chaincfg.RegressionNetParams
	}
	return chaincfg.MainNetParams
}

// Init init client context
func (n *neutrinoClient) Init(ctx context.Context, api client.QueueProtocolAPI, commitAddr string, jsonCfg []byte) error {

	n.ctx = ctx
	n.chain33Api = api
	n.chain33FeeRate = 100000

	types.MustDecode(jsonCfg, &n.cfg)

	if n.cfg.BlockCacheSize < 1024*1024 {
		n.cfg.BlockCacheSize = defalutBlockCacheSize
	}

	if n.cfg.MaxPeer < 1 {
		n.cfg.MaxPeer = 8
	}

	neutrinoCfg, err := initNeutrinoConfig(api, n.cfg)
	if err != nil {
		log.Error("Init", "initNeutrinoConfig error", err)
		return err
	}
	n.neutrinoCfg = neutrinoCfg

	cs, err := neutrino.NewChainService(neutrinoCfg)
	if err != nil {
		log.Error("Init", "NewChainService error", err)
		_ = neutrinoCfg.Database.Close()
		return err
	}
	n.neutrinoCS = cs
	return nil

}

func initNeutrinoConfig(api client.QueueProtocolAPI, clientCfg config) (neutrino.Config, error) {

	dbPath := filepath.Join(api.GetConfig().GetModuleConfig().BlockChain.DbPath, "lightclient")
	dbName := filepath.Join(dbPath, "neutrino.db")
	db, err := walletdb.Create("bdb", dbName)
	if err != nil {
		log.Error("getNeutrinoConfig Create db error", "err", err)
		return neutrino.Config{}, err
	}

	neutrino.MaxPeers = clientCfg.MaxPeer
	neutrino.BanDuration = time.Hour * 48

	cfg := neutrino.Config{
		DataDir:      dbPath,
		Database:     db,
		ChainParams:  clientCfg.getChainParams(),
		ConnectPeers: clientCfg.ConnectPeers,
		AddPeers:     clientCfg.AddPeers,
		//BlockCache:         lru.NewCache[wire.InvVect, *neutrino.CacheableBlock](clientCfg.BlockCacheSize),
		BlockCacheSize: clientCfg.BlockCacheSize,
	}

	return cfg, nil

}

// Start starting routine
func (n *neutrinoClient) Start() {

	if err := n.neutrinoCS.Start(); err != nil {
		log.Error("Start", "neutrinoCS start error", err)
		_ = n.neutrinoCfg.Database.Close()
		return
	}

	newRGBX().init(n)
	go n.cleanUp()
}

func (n *neutrinoClient) cleanUp() {

	for {

		select {

		case <-n.ctx.Done():
			if err := n.neutrinoCS.Stop(); err != nil {
				log.Error("cleanUp Unable to stop neutrino server", "err", err)
			}
			if err := n.neutrinoCfg.Database.Close(); err != nil {
				log.Error("cleanUp Unable to close neutrino db", "err", err)
			}
		}
	}

}

func (n *neutrinoClient) handleBestBlock() {

	ticker := time.NewTicker(time.Second * 30)
	for {

		select {

		case <-n.ctx.Done():
			return
		case <-ticker.C:

			blk, err := n.neutrinoCS.BestBlock()
			if err != nil {
				log.Error("handleBestBlock", "err", err)
				continue
			}
			n.setBestBlock(blk)
		}
	}
}

func (n *neutrinoClient) getBestBlock() *headerfs.BlockStamp {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.bestBlock
}

func (n *neutrinoClient) setBestBlock(blk *headerfs.BlockStamp) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if blk != nil {
		n.bestBlock = blk
	}
}
