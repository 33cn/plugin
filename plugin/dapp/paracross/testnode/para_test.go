package testnode

import (
	"testing"

	"github.com/33cn/chain33/util"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

func TestParaNode(t *testing.T) {
	para := NewParaNode(nil, nil)
	//通过rpc 发生信息
	txs := util.GenNoneTxs(para.Para.GetGenesisKey(), 10)
	for i := 0; i < len(txs); i++ {
		para.Para.SendTxRPC(txs[i])
	}
	para.Para.WaitHeight(1)
	txs = util.GenNoneTxs(para.Para.GetGenesisKey(), 10)
	for i := 0; i < len(txs); i++ {
		para.Para.SendTxRPC(txs[i])
	}
	para.Para.WaitHeight(2)
}
