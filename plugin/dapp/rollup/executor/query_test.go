package executor

import (
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	paratypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRollup_Query_GetValidatorPubs(t *testing.T) {

	r := newRollup()
	funcName := "GetValidatorPubs"
	_, err := r.Query(funcName, nil)
	require.Equal(t, ErrChainTitle, err)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)

	api.On("Query", mock.Anything, "GetNodeGroupStatus", mock.Anything).Return(nil, types.ErrActionNotSupport).Once()
	_, err = r.Query(funcName, types.Encode(&rtypes.ChainTitle{Value: "user.p.test"}))
	require.Equal(t, types.ErrActionNotSupport, err)
	api.On("Query", mock.Anything, "GetNodeGroupStatus", mock.Anything).Return(&paratypes.ParaNodeGroupStatus{}, nil).Once()
	_, err = r.Query(funcName, types.Encode(&rtypes.ChainTitle{Value: "user.p.test"}))
	require.Nil(t, err)
}

func TestRollup_Query_GetRollupStatus(t *testing.T) {

	r := newRollup()
	funcName := "GetRollupStatus"
	_, err := r.Query(funcName, nil)
	require.Equal(t, ErrChainTitle, err)
	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(state)
	title := "user.p.test"
	status, err := r.Query(funcName, types.Encode(&rtypes.ChainTitle{Value: title}))
	require.Nil(t, err)
	require.True(t, status.(*rtypes.RollupStatus).CommitRound == 0)
	require.True(t, status.(*rtypes.RollupStatus).CommitBlockHeight == 0)
	_ = state.Set(formatRollupStatusKey(title), []byte("errorData"))
	_, err = r.Query(funcName, types.Encode(&rtypes.ChainTitle{Value: title}))
	require.NotNil(t, err)
}

func TestRollup_Query_GetCommitRoundInfo(t *testing.T) {

	r := newRollup()
	funcName := "GetCommitRoundInfo"
	_, err := r.Query(funcName, nil)
	require.Equal(t, ErrChainTitle, err)
	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(state)
	title := "user.p.test"
	_, err = r.Query(funcName, types.Encode(&rtypes.ReqGetCommitRound{ChainTitle: title}))
	require.Equal(t, types.ErrNotFound, err)

	_ = state.Set(formatCommitRoundInfoKey(title, 1), []byte("errorData"))
	_, err = r.Query(funcName, types.Encode(&rtypes.ReqGetCommitRound{ChainTitle: title, CommitRound: 1}))
	require.NotNil(t, err)

	_ = state.Set(formatCommitRoundInfoKey(title, 1), types.Encode(&rtypes.CommitRoundInfo{CommitRound: 1}))
	info, err := r.Query(funcName, types.Encode(&rtypes.ReqGetCommitRound{ChainTitle: title, CommitRound: 1}))
	require.Nil(t, err)
	require.True(t, info.(*rtypes.CommitRoundInfo).CommitRound == 1)
}
