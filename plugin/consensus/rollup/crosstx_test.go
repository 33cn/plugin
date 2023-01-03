package rollup

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/33cn/chain33/rpc/grpcclient"
	_ "github.com/33cn/chain33/system/consensus/init"
	_ "github.com/33cn/chain33/system/dapp/init"
	_ "github.com/33cn/chain33/system/mempool/init"
	_ "github.com/33cn/chain33/system/store/init"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/require"
)

func newTestHandler() *crossTxHandler {
	ru := &RollUp{}
	h := &crossTxHandler{}
	h.init(ru, &rtypes.RollupStatus{})
	return h
}

func TestCrossTxHandler(t *testing.T) {

	h := newTestHandler()

	tx := &types.Transaction{Payload: []byte("test")}
	h.addMainChainCrossTx(2, nil)
	require.Equal(t, 0, len(h.txIdxCache))
	h.addMainChainCrossTx(2, []*types.Transaction{tx})
	require.Equal(t, 1, len(h.txIdxCache))
	idxArr := h.removePackedCrossTx([][]byte{tx.Hash(), []byte("test")})
	require.Equal(t, 0, len(h.txIdxCache))
	require.Equal(t, 1, len(idxArr))
	require.Equal(t, int64(2), idxArr[0].BlockHeight)
	require.Equal(t, int32(0), idxArr[0].FilterIndex)
	h.removePackedCrossTx(nil)
	require.Equal(t, 0, len(h.txIdxCache))
}

func TestRefreshSyncedHeight(t *testing.T) {

	h := newTestHandler()
	tx := &types.Transaction{Payload: []byte("test")}
	h.addMainChainCrossTx(2, []*types.Transaction{tx})
	require.Equal(t, 1, len(h.txIdxCache))
	info := h.txIdxCache[shortHash(tx.Hash())]
	require.Equal(t, int64(1), h.refreshSyncedHeight())
	info.enterTimestamp = types.Now().Unix() - 600
	require.Equal(t, int64(2), h.refreshSyncedHeight())
	require.Equal(t, 0, len(h.txIdxCache))
}

func TestRemoveErrTx(t *testing.T) {

	h := newTestHandler()
	tx := &types.Transaction{Payload: []byte("test")}
	h.addMainChainCrossTx(2, []*types.Transaction{tx})
	require.Equal(t, 1, len(h.txIdxCache))

	h.removeErrTxs([]*types.Transaction{tx})
	require.Equal(t, 0, len(h.txIdxCache))
}

func TestPullCrossTx(t *testing.T) {

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.GetModuleConfig().RPC.GrpcBindAddr = fmt.Sprintf("localhost:%d", 9965)
	node := testnode.NewWithRPC(cfg, nil)
	defer node.Close()
	h := newTestHandler()
	grpc, err := grpcclient.NewMainChainClient(cfg, cfg.GetModuleConfig().RPC.GrpcBindAddr)
	require.Nil(t, err)
	cfg.SetTitleOnlyForTest("user.p.para")
	h.ru.chainCfg = cfg
	h.ru.mainChainGrpc = grpc
	h.ru.ctx = context.Background()
	txs := util.GenNoneTxs(cfg, node.GetGenesisKey(), 20)
	for i := 0; i < len(txs); i++ {
		node.GetAPI().SendTx(txs[i])
		node.WaitHeight(int64(i + 1))
	}

	go h.pullCrossTx()
	start := types.Now().Unix()
	for {
		h.lock.Lock()
		pulled := h.pulledHeight
		h.lock.Unlock()
		if pulled == 8 {
			return
		}
		if types.Now().Unix()-start >= 5 {
			t.Errorf("test timeout, pullHeight= %d", pulled)
			return
		}
		time.Sleep(time.Millisecond)
	}
}
