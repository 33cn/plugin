package neutrino

import (
	"encoding/json"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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
	api, err := client.New(q.Client(), nil)
	require.NoError(t, err)
	util.ResetDatadir(cfg.GetModuleConfig(), "$TEMP/")
	err = n.initNeutrinoConfig(api)
	require.NoError(t, err)

	require.True(t, n.cfg.BlockCacheSize == defalutBlockCacheSize)
	require.True(t, n.cfg.MaxPeer == 8)
	require.True(t, n.cfg.BtcBlockInterval == 10)
	require.True(t, n.cfg.BlockConfirmations == 0)
	require.True(t, n.cfg.MaxUtxoRescanTime == int64(24*time.Hour/time.Second))
	dir := cfg.GetModuleConfig().BlockChain.DbPath + "/lightclient"
	require.Equal(t, dir, n.neutrinoCfg.DataDir)
	require.Equal(t, n.cfg.NetName, n.neutrinoCfg.ChainParams.Name)
}
