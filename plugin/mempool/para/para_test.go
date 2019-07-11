package para_test

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	paratest "github.com/33cn/plugin/plugin/dapp/paracross/testnode"
	"github.com/33cn/plugin/plugin/mempool/para"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

func TestClose(t *testing.T) {
	mem := para.NewMempool(nil)
	n := 1000
	done := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		go func() {
			mem.Close()
			done <- struct{}{}
		}()
	}
	for i := 0; i < n; i++ {
		<-done
	}
}

func TestParaNodeMempool(t *testing.T) {
	main := testnode.New("", nil)
	main.Listen()

	cfg, sub := types.InitCfgString(paratest.DefaultConfig)
	testnode.ModifyParaClient(sub, main.GetCfg().RPC.GrpcBindAddr)
	cfg.Mempool.Name = "para"
	para := testnode.NewWithConfig(cfg, sub, nil)
	para.Listen()
	mockpara := paratest.NewParaNode(main, para)
	tx := util.CreateTxWithExecer(mockpara.Para.GetGenesisKey(), "user.p.guodun.none")
	hash := mockpara.Para.SendTx(tx)
	assert.Equal(t, tx.Hash(), hash)

	_, err := mockpara.Para.GetAPI().GetMempool(&types.ReqGetMempool{})
	assert.Equal(t, err, types.ErrActionNotSupport)
	t.Log(err)
}
