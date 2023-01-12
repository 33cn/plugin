package rollup

import (
	"context"
	"fmt"
	"testing"

	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/require"

	_ "github.com/33cn/plugin/plugin/dapp/init"
)

func newTestNode(t *testing.T) (*testnode.Chain33Mock, *RollUp) {

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.GetModuleConfig().RPC.GrpcBindAddr = fmt.Sprintf("localhost:%d", 9965)
	node := testnode.NewWithRPC(cfg, nil)
	r := &RollUp{}
	grpc, err := grpcclient.NewMainChainClient(cfg, cfg.GetModuleConfig().RPC.GrpcBindAddr)
	require.Nil(t, err)
	cfg.SetTitleOnlyForTest("user.p.test")
	r.chainCfg = cfg
	r.mainChainGrpc = grpc
	r.ctx = context.Background()

	return node, r
}

func Test_getValidatorPubKeys(t *testing.T) {

	mock, r := newTestNode(t)
	defer mock.Close()

	pubs := r.getValidatorPubKeys()
	require.Nil(t, pubs)

}

func Test_getRollupStatus(t *testing.T) {

	mock, r := newTestNode(t)
	defer mock.Close()

	status := r.getRollupStatus()

	require.NotNil(t, status)
	require.True(t, status.CommitRound == 0)
	require.True(t, status.CommitBlockHeight == 0)
}

func Test_createTx(t *testing.T) {

	mock, r := newTestNode(t)
	defer mock.Close()

	tx, err := r.createTx(rtypes.RollupX, "errAction", nil)
	require.NotNil(t, err)
	require.Nil(t, tx)

	tx, err = r.createTx(rtypes.RollupX, rtypes.NameCommitAction, nil)
	require.Nil(t, err)
	action := &rtypes.RollupAction{}
	err = types.Decode(tx.Payload, action)
	require.Nil(t, err)
	require.True(t, rtypes.TyCommitAction == action.GetTy())
}

func Test_getProperFeeRate(t *testing.T) {

	mock, r := newTestNode(t)
	defer mock.Close()

	rate := r.getProperFeeRate()
	require.Equal(t, r.chainCfg.GetMinTxFeeRate(), rate)
}

func Test_sendTx2MainChain(t *testing.T) {

	mock, r := newTestNode(t)
	defer mock.Close()

	tx, err := r.createTx(rtypes.RollupX, rtypes.NameCommitAction, nil)
	require.Nil(t, err)

	err = r.sendTx2MainChain(tx)
	require.NotNil(t, err)
	tx = util.CreateNoneTx(r.chainCfg, mock.GetGenesisKey())
	err = r.sendTx2MainChain(tx)
	require.Nil(t, err)
}

func Test_fetchCrossTx(t *testing.T) {

	mock, r := newTestNode(t)
	defer mock.Close()

	mock.SendTx(util.CreateNoneTx(r.chainCfg, mock.GetGenesisKey()))
	err := mock.WaitHeightTimeout(1, 3)
	require.Nil(t, err)

	details, err := r.fetchCrossTx(0, mock.GetLastBlock().Height)

	require.Nil(t, err)
	require.Equal(t, 2, len(details.GetItems()))

}
