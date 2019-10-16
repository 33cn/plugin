package testnode

import (
	"testing"

	"github.com/33cn/chain33/util"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
	"github.com/33cn/chain33/util/testnode"
	"github.com/33cn/chain33/types"
	"strings"
)

func TestParaNode(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"" , 1))
	cfg.GetModuleConfig().Consensus.Name = "ticket"

	main := testnode.NewWithConfig(cfg, nil)
	main.Listen()

	para := NewParaNode(main, nil)
	paraCfg := para.Para.GetAPI().GetConfig()
	defer para.Close()
	//通过rpc 发生信息
	tx := util.CreateTxWithExecer(paraCfg, para.Para.GetGenesisKey(), "user.p.guodun.none")
	para.Para.SendTxRPC(tx)
	para.Para.WaitHeight(1)
	tx = util.CreateTxWithExecer(paraCfg, para.Para.GetGenesisKey(), "user.p.guodun.none")
	para.Para.SendTxRPC(tx)
	para.Para.WaitHeight(2)
}
