package neutrino

import (
	"encoding/json"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
	"github.com/stretchr/testify/require"
	"testing"
)

var lightConfig = `

[rpc.sub.light]
clients=["neutrino"]
commitAddr=""

#[rpc.sub.light.neutrino]
#netType="regress"

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
	require.Equal(t, subCfg.NetType, "regress")

}
