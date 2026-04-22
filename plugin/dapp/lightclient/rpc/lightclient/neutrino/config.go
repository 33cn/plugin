package neutrino

import (
	"os"
	"path/filepath"
	"time"

	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/lightninglabs/neutrino"
)

// defaultBlockCacheSize is the size (in bytes) of blocks that will be
// keep in memory if no size is specified.
const defalutBlockCacheSize = 20 * 1024 * 1024 //20 MB

type config struct {

	// IsOfficialNode 是否为官方主节点
	IsOfficialNode bool
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
	// BtcHeaderStartHeight 首次启动提交btc header的起始高度
	BtcHeaderStartHeight uint64 `json:"btcHeaderStartHeight"`
	// MaxUtxoRescanTime, utxo 检索最大时长，hour, 0为永不超时
	MaxUtxoRescanTime int64 `json:"maxUtxoRescanTime"`
	// BtcFullNodeRPC 可选，比特币全节点 RPC 配置，用于查询 block 构造SPV
	BtcRPC btcRPCConfig `json:"btcRPC"`
	// Tss tss config
	Tss tssConfig `json:"tss"`
}

type btcRPCConfig struct {
	// Host 例如: 127.0.0.1:8332 或 btc-node.example.com:8332
	Host string `json:"host"`
	// User/Pass 对应 bitcoin.conf 的 rpcuser/rpcpassword
	User string `json:"user"`
	Pass string `json:"pass"`
	// Mode: ws(默认) 或 http
	Mode string `json:"mode"`
	// DisableTLS 是否禁用 TLS
	DisableTLS bool `json:"disableTLS"`
	// CertFile TLS 证书文件（可选）
	CertFile string `json:"certFile"`
}

func (c *btcRPCConfig) toConnConfig() (*rpcclient.ConnConfig, error) {
	endpoint := "ws"
	httpPostMode := false
	if c.Mode == "http" {
		endpoint = "http"
		httpPostMode = true
	}
	conn := &rpcclient.ConnConfig{
		Host:         c.Host,
		Endpoint:     endpoint,
		User:         c.User,
		Pass:         c.Pass,
		DisableTLS:   c.DisableTLS,
		HTTPPostMode: httpPostMode,
	}
	if c.CertFile == "" {
		return conn, nil
	}
	certs, err := os.ReadFile(c.CertFile)
	if err != nil {
		return nil, err
	}
	conn.Certificates = certs
	return conn, nil
}

type tssConfig struct {
	// Peers peers name
	Peers []string `json:"peers"`
	// Threshold peer threshold
	Threshold uint32 `json:"threshold"`
	// Rank peer rank
	Rank uint32 `json:"rank"`
}

func (c config) getChainParams() chaincfg.Params {

	params := ltypes.GetBtcChainParams(c.NetName)
	return *params
}

func (n *neutrinoClient) initNeutrinoConfig(chainCfg *types.Chain33Config) error {

	if n.cfg.BlockCacheSize < 1024*1024 {
		n.cfg.BlockCacheSize = defalutBlockCacheSize
	}
	if n.cfg.MaxPeer < 1 {
		n.cfg.MaxPeer = 8
	}

	if n.cfg.BtcBlockInterval <= 0 {
		n.cfg.BtcBlockInterval = 600
	}
	if n.cfg.BtcHeaderStartHeight == 0 {
		n.cfg.BtcHeaderStartHeight = 1
	}
	// convert to second
	if n.cfg.MaxUtxoRescanTime > 0 {
		n.cfg.MaxUtxoRescanTime *= int64(time.Hour / time.Second)
	}
	dbPath := filepath.Join(chainCfg.GetModuleConfig().BlockChain.DbPath, "lightclient")
	_ = os.MkdirAll(dbPath, 0755)
	_, db, err := openWalletDB(dbPath, "neutrino.db")
	if err != nil {
		log.Error("initNeutrinoConfig open db error", "err", err)
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

func openWalletDB(path, dbName string) (exist bool, db walletdb.DB, err error) {

	dbPath := filepath.Join(path, dbName)
	exist, err = fileExists(dbPath)
	if err != nil {
		return false, nil, err
	}
	if exist {
		db, err = walletdb.Open("bdb", dbPath, true, time.Second*10, false)
	} else {
		db, err = walletdb.Create("bdb", dbPath, false, time.Second*10, false)
	}
	return exist, db, err
}
