package executor

import (
	"errors"
	"testing"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	wcom "github.com/33cn/chain33/wallet/common"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
	"github.com/stretchr/testify/require"
)

var (
	testHeight    = int64(100)
	testBlockTime = int64(1539918074)

	// 测试的私钥
	testPrivateKeys = []string{
		"0x8dea7332c7bb3e3b0ce542db41161fd021e3cfda9d7dabacf24f98f2dfd69558",
		"0x920976ffe83b5a98f603b999681a0bc790d97e22ffc4e578a707c2234d55cc8a",
		"0xb59f2b02781678356c231ad565f73699753a28fd3226f1082b513ebf6756c15c",
	}
	// 测试的地址
	testAddrs = []string{
		"1EDDghAtgBsamrNEtNmYdQzC1QEhLkr87t",
		"13cS5G1BDN2YfGudsxRxr7X25yu6ZdgxMU",
		"1JSRSwp16NvXiTjYBYK9iUQ9wqp3sCxz2p",
	}
	// 测试的隐私公钥对
	testPubkeyPairs = []string{
		"92fe6cfec2e19cd15f203f83b5d440ddb63d0cb71559f96dc81208d819fea85886b08f6e874fca15108d244b40f9086d8c03260d4b954a40dfb3cbe41ebc7389",
		"6326126c968a93a546d8f67d623ad9729da0e3e4b47c328a273dfea6930ffdc87bcc365822b80b90c72d30e955e7870a7a9725e9a946b9e89aec6db9455557eb",
		"44bf54abcbae297baf3dec4dd998b313eafb01166760f0c3a4b36509b33d3b50239de0a5f2f47c2fc98a98a382dcd95a2c5bf1f4910467418a3c2595b853338e",
	}

	privKeys = make([]crypto.PrivKey, len(testPrivateKeys))
	testCfg  = types.NewChain33Config(types.GetDefaultCfgstring())
)

func init() {
	log.SetLogLevel("error")
	Init(vty.VoteX, testCfg, nil)
	for i, priv := range testPrivateKeys {
		privKeys[i], _ = decodePrivKey(priv)
	}
}

type testExecMock struct {
	dbDir    string
	localDB  dbm.KVDB
	stateDB  dbm.DB
	exec     *vote
	policy   wcom.WalletBizPolicy
	cfg      *types.Chain33Config
	q        queue.Queue
	qapi     client.QueueProtocolAPI
	execType types.ExecutorType
}

type testcase struct {
	payload            types.Message
	expectExecErr      error
	expectCheckErr     error
	expectExecLocalErr error
	expectExecDelErr   error
	priv               crypto.PrivKey
	execType           int
	index              int
}

// InitEnv init env
func (mock *testExecMock) InitEnv() {

	mock.cfg = testCfg
	util.ResetDatadir(mock.cfg.GetModuleConfig(), "$TEMP/")
	mock.q = queue.New("channel")
	mock.q.SetConfig(mock.cfg)
	mock.qapi, _ = client.New(mock.q.Client(), nil)
	mock.initExec()

}

func (mock *testExecMock) FreeEnv() {
	util.CloseTestDB(mock.dbDir, mock.stateDB)
}

func (mock *testExecMock) initExec() {
	mock.dbDir, mock.stateDB, mock.localDB = util.CreateTestDB()
	exec := newVote()
	exec.SetAPI(mock.qapi)
	exec.SetStateDB(mock.stateDB)
	exec.SetLocalDB(mock.localDB)
	exec.SetEnv(testHeight, testBlockTime, 1539918074)
	mock.exec = exec.(*vote)
	mock.execType = types.LoadExecutorType(vty.VoteX)
}

func decodePrivKey(priv string) (crypto.PrivKey, error) {
	c, err := crypto.Load(crypto.GetName(types.SECP256K1), -1)
	if err != nil {
		return nil, err
	}
	bytes, err := common.FromHex(priv[:])
	if err != nil {
		return nil, err
	}
	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func createTx(mock *testExecMock, payload types.Message, privKey crypto.PrivKey) (*types.Transaction, error) {

	action, _ := getActionName(payload)
	tx, err := mock.execType.CreateTransaction(action, payload)
	if err != nil {
		return nil, errors.New("createTxErr:" + err.Error())
	}
	tx, err = types.FormatTx(mock.cfg, vty.VoteX, tx)
	if err != nil {
		return nil, errors.New("formatTxErr:" + err.Error())
	}
	tx.Sign(int32(types.SECP256K1), privKey)
	return tx, nil
}

func getActionName(param types.Message) (string, error) {

	if _, ok := param.(*vty.CreateGroup); ok {
		return vty.NameCreateGroupAction, nil
	} else if _, ok := param.(*vty.UpdateGroup); ok {
		return vty.NameUpdateGroupAction, nil
	} else if _, ok := param.(*vty.CreateVote); ok {
		return vty.NameCreateVoteAction, nil
	} else if _, ok := param.(*vty.CommitVote); ok {
		return vty.NameCommitVoteAction, nil
	} else if _, ok := param.(*vty.CloseVote); ok {
		return vty.NameCloseVoteAction, nil
	} else if _, ok := param.(*vty.UpdateMember); ok {
		return vty.NameUpdateMemberAction, nil
	} else {
		return "", types.ErrActionNotSupport
	}
}

func TestUtil(t *testing.T) {

	heightIndex := dapp.HeightIndexStr(100, 0)
	require.Equal(t, IDLen, len(formatGroupID(heightIndex)))
	require.Equal(t, IDLen, len(formatVoteID(heightIndex)))

	strs := []string{"a", "b", "c"}
	require.True(t, checkSliceItemExist("a", strs))
	require.False(t, checkSliceItemExist("d", strs))
	require.False(t, checkSliceItemDuplicate(strs))
	strs = append(strs, "c")
	require.True(t, checkSliceItemDuplicate(strs))
	members := make([]*vty.GroupMember, 0)
	for _, addr := range testAddrs {
		members = append(members, &vty.GroupMember{
			Addr: addr,
		})
	}
	require.True(t, checkMemberExist(testAddrs[0], members))
	require.False(t, checkMemberExist("testaddr", members))
	require.Equal(t, addrLen, len(testAddrs[0]))
}

func TestFilterVoteWithStatus(t *testing.T) {

	currentTime := types.Now().Unix()
	voteList := []*vty.VoteInfo{
		{
			BeginTimestamp: currentTime + 1,
		},
		{
			EndTimestamp: currentTime + 1,
		},
		{
			BeginTimestamp: currentTime,
			EndTimestamp:   currentTime,
		},
		{
			Status: voteStatusClosed,
		},
	}
	statusList := []uint32{voteStatusPending, voteStatusOngoing, voteStatusFinished, voteStatusClosed}
	voteList = filterVoteWithStatus(voteList, 0, currentTime)
	for i, info := range voteList {
		require.Equal(t, statusList[i], info.Status)
	}
	for _, status := range statusList {
		list := filterVoteWithStatus(voteList, status, currentTime)
		require.Equal(t, 1, len(list))
		require.Equal(t, status, list[0].Status)
	}
}
