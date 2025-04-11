package neutrino

import (
	"github.com/33cn/chain33/client"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/lightninglabs/neutrino"
	"path/filepath"
	"strings"
	"time"
)

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
