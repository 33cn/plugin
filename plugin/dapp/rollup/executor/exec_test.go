package executor

import (
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/require"
)

var cfg *types.Chain33Config

func init() {

	cfg = types.NewChain33Config(types.ReadFile("../../../../chain33.toml"))
	Init(driverName, cfg, nil)
}

func Test_rollup(t *testing.T) {
	r := newRollup()
	require.Equal(t, driverName, r.GetDriverName())
	require.Equal(t, driverName, GetName())
}

func TestRollup_Exec_Commit(t *testing.T) {

	r := newRollup()
	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(state)

	_ = state.Set(formatRollupStatusKey(""), []byte("test"))
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameCommitAction, &rtypes.CheckPoint{})
	require.Nil(t, err)
	_, err = r.Exec(tx, 0)
	require.Equal(t, ErrGetRollupStatus, err)

	_ = state.Set(formatRollupStatusKey(""), types.Encode(&rtypes.RollupStatus{}))
	cp := &rtypes.CheckPoint{Batch: &rtypes.BlockBatch{BlockHeaders: []*types.Header{{}}}}
	tx, err = r.GetExecutorType().CreateTransaction(rtypes.NameCommitAction, cp)
	require.Nil(t, err)
	_, err = r.Exec(tx, 0)
	require.Nil(t, err)
}

func TestRollup_ExecLocal_CommitBatch(t *testing.T) {

	r := newRollup()
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameCommitAction, &rtypes.CheckPoint{})
	require.Nil(t, err)
	_, err = r.ExecLocal(tx, nil, 0)
	require.Nil(t, err)
}

func TestRollup_ExecDelLocal(t *testing.T) {

	r := newRollup()
	dir, state, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)

	r.SetLocalDB(local)
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameCommitAction, &rtypes.CheckPoint{})
	require.Nil(t, err)
	_, err = r.ExecDelLocal(tx, nil, 0)
	require.Nil(t, err)
}
