package neutrino

import (
	"github.com/33cn/chain33/client"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/lightninglabs/neutrino"
	"os"
	"path/filepath"
	"time"
)

// defaultBlockCacheSize is the size (in bytes) of blocks that will be
// keep in memory if no size is specified.
const defalutBlockCacheSize = 20 * 1024 * 1024 //20 MB

type config struct {
	// MaxPeers is the maximum number of connections the client maintains.
	MaxPeer int `json:"maxPeer"`
	// BlockCacheSize indicates the size (in bytes) of blocks the block
	// cache will hold in memory at most. If a BlockCache is provided then
	// BlockCacheSize is ignored.
	BlockCacheSize uint64 `json:"blockCacheSize"`
	// NetName btc network name
	NetName string `json:"netName"`
	// AddPeers is a slice of hosts that should be connected to on startup,
	// and be maintained as persistent peers.
	AddPeers []string `json:"addPeers"`
	// ConnectPeers is a slice of hosts that should be connected to on
	// startup, and be established as persistent peers.
	//
	// NOTE: If specified, we'll *only* connect to this set of peers and
	// won't attempt to automatically seek outbound peers.
	ConnectPeers []string `json:"connectPeers"`

	// BtcBlockInterval 区块时间间隔，second
	BtcBlockInterval uint32 `json:"btcBlockInterval"`
	// BlockConfirmations 区块确认数
	BlockConfirmations uint32 `json:"blockConfirmations"`
	// MaxUtxoRescanTime, utxo 检索最大时长，hour, 0为永不超时
	MaxUtxoRescanTime int64 `json:"maxUtxoRescanTime"`
}

func (c config) getChainParams() chaincfg.Params {

	if c.NetName == "simnet" {
		return chaincfg.SimNetParams
	} else if c.NetName == "testnet3" {
		return chaincfg.TestNet3Params
	} else if c.NetName == "regtest" {
		return chaincfg.RegressionNetParams
	} else if c.NetName == "signet" {
		return chaincfg.SigNetParams
	}
	return chaincfg.MainNetParams
}

func (n *neutrinoClient) initNeutrinoConfig(api client.QueueProtocolAPI) error {

	if n.cfg.BlockCacheSize < 1024*1024 {
		n.cfg.BlockCacheSize = defalutBlockCacheSize
	}
	if n.cfg.MaxPeer < 1 {
		n.cfg.MaxPeer = 8
	}

	if n.cfg.BtcBlockInterval <= 0 {
		n.cfg.BtcBlockInterval = 600
	}
	// convert to second
	if n.cfg.MaxUtxoRescanTime > 0 {
		n.cfg.MaxUtxoRescanTime *= int64(time.Hour / time.Second)
	}

	dbPath := filepath.Join(api.GetConfig().GetModuleConfig().BlockChain.DbPath, "lightclient")
	_ = os.MkdirAll(dbPath, 0755)
	dbName := filepath.Join(dbPath, "neutrino.db")
	db, err := walletdb.Create("bdb", dbName, true, time.Second*10)
	if err != nil {
		log.Error("getNeutrinoConfig Create db error", "err", err)
		return err
	}

	neutrino.MaxPeers = n.cfg.MaxPeer
	neutrino.BanDuration = time.Hour * 48

	n.neutrinoCfg = neutrino.Config{
		DataDir:      dbPath,
		Database:     db,
		ChainParams:  n.cfg.getChainParams(),
		ConnectPeers: n.cfg.ConnectPeers,
		AddPeers:     n.cfg.AddPeers,
		//BlockCache:         lru.NewCache[wire.InvVect, *neutrino.CacheableBlock](clientCfg.BlockCacheSize),
		BlockCacheSize: n.cfg.BlockCacheSize,
	}

	return nil

}
