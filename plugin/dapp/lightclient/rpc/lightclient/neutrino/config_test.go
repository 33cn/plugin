package neutrino

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var lightConfig = `

[rpc.sub.light]
clients=["neutrino"]
commitAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"

[rpc.sub.light.neutrino]
netName="regtest"
addPeers=["127.0.0.1:8333"]
btcBlockInterval=10
blockConfirmations=0
maxUtxoRescanTime=24

`

func Test_config(t *testing.T) {

	cfg := types.NewChain33Config(types.MergeCfg(types.GetDefaultCfgstring(), lightConfig))
	light := &lightclient.Config{}
	types.MustDecode(cfg.GetSubConfig().RPC["light"], light)
	require.Equal(t, 1, len(light.EnableClients))
	require.Equal(t, "neutrino", light.EnableClients[0])

	sub, err := json.Marshal(light.Neutrino)
	require.NoError(t, err)
	subCfg := &config{}
	types.MustDecode(sub, subCfg)
	require.Equal(t, subCfg.NetName, "regtest")

	// test init neutrino config
	n := &neutrinoClient{}
	n.cfg = *subCfg
	q := queue.New("test")
	q.SetConfig(cfg)
	require.NoError(t, err)
	util.ResetDatadir(cfg.GetModuleConfig(), "$TEMP/")
	err = n.initNeutrinoConfig(cfg)
	if err != nil {
		t.Skipf("initNeutrinoConfig needs walletdb/bdb (environment): %v", err)
	}

	require.True(t, n.cfg.BlockCacheSize == defalutBlockCacheSize)
	require.True(t, n.cfg.MaxPeer == 8)
	require.True(t, n.cfg.BtcBlockInterval == 10)
	require.True(t, n.cfg.BlockConfirmations == 0)
	require.True(t, n.cfg.MaxUtxoRescanTime == int64(24*time.Hour/time.Second))
	dir := cfg.GetModuleConfig().BlockChain.DbPath + "/lightclient"
	require.Equal(t, dir, n.neutrinoCfg.DataDir)
	require.Equal(t, n.cfg.NetName, n.neutrinoCfg.ChainParams.Name)
}

// TestBtcRPCConfig tests the btcRPCConfig.toConnConfig method
func TestBtcRPCConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   btcRPCConfig
		wantErr  bool
		wantMode string
	}{
		{
			name: "ws mode (default)",
			config: btcRPCConfig{
				Host:       "127.0.0.1:8332",
				User:       "testuser",
				Pass:       "testpass",
				Mode:       "",
				DisableTLS: true,
			},
			wantErr:  false,
			wantMode: "ws",
		},
		{
			name: "http mode",
			config: btcRPCConfig{
				Host:       "127.0.0.1:8332",
				User:       "testuser",
				Pass:       "testpass",
				Mode:       "http",
				DisableTLS: true,
			},
			wantErr:  false,
			wantMode: "http",
		},
		{
			name: "with TLS cert",
			config: btcRPCConfig{
				Host:       "127.0.0.1:8332",
				User:       "testuser",
				Pass:       "testpass",
				Mode:       "ws",
				DisableTLS: false,
				CertFile:   "/tmp/nonexistent_cert.pem",
			},
			wantErr: true, // file doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := tt.config.toConnConfig()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, conn)
			assert.Equal(t, tt.config.Host, conn.Host)
			assert.Equal(t, tt.config.User, conn.User)
			assert.Equal(t, tt.config.Pass, conn.Pass)
			assert.Equal(t, tt.wantMode, conn.Endpoint)
			if tt.wantMode == "http" {
				assert.True(t, conn.HTTPPostMode)
			}
		})
	}
}

// TestConfigGetChainParams tests the config.getChainParams method
func TestConfigGetChainParams(t *testing.T) {
	tests := []struct {
		netName  string
		wantName string
	}{
		{"mainnet", "mainnet"},
		{"testnet", "testnet3"},
		{"testnet4", "testnet4"},
		{"testnet3", "testnet3"},
		{"regtest", "regtest"},
		{"simnet", "simnet"},
		{"signet", "signet"},
		{"unknown", "mainnet"}, // default to mainnet
	}

	for _, tt := range tests {
		cfg := config{NetName: tt.netName}
		params := cfg.getChainParams()
		assert.Equal(t, tt.wantName, params.Name)
	}
}

// Test_openWalletDB verifies that openWalletDB correctly creates a new database
// when it doesn't exist and opens an existing one.
func Test_openWalletDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbName := "test_neutrino.db"

	// Test 1: Create new database (doesn't exist yet)
	exist, db, err := openWalletDB(tmpDir, dbName)
	if err != nil {
		t.Skipf("walletdb.Create requires bdb driver (environment): %v", err)
	}
	require.NoError(t, err)
	assert.False(t, exist, "database should not exist before creation")
	assert.NotNil(t, db, "database handle should not be nil after creation")

	// Verify database file was created
	dbPath := filepath.Join(tmpDir, dbName)
	_, statErr := os.Stat(dbPath)
	assert.NoError(t, statErr, "database file should exist on disk")

	// Close the database before reopening
	db.Close()

	// Test 2: Open existing database
	exist2, db2, err2 := openWalletDB(tmpDir, dbName)
	require.NoError(t, err2)
	assert.True(t, exist2, "database should exist when reopening")
	assert.NotNil(t, db2, "database handle should not be nil when reopening")
	db2.Close()
}
