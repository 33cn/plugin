package executor

import (
	"fmt"
	"strings"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	paratypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_CheckTx(t *testing.T) {

	r := newRollup()
	action := &rtypes.RollupAction{}

	tx := &types.Transaction{Payload: []byte("testdata")}
	require.Equal(t, types.ErrActionNotSupport, r.CheckTx(tx, 0))

	tx.Payload = types.Encode(action)
	require.Equal(t, types.ErrActionNotSupport, r.CheckTx(tx, 0))
	action.Ty = rtypes.TyCommitAction
	tx.Payload = types.Encode(action)
	require.NotNil(t, r.CheckTx(tx, 0))
}

func Test_checkCommit(t *testing.T) {

	title := "user.p.para"
	r := &rollup{}
	cp := &rtypes.CheckPoint{ChainTitle: "invalidTitle"}
	require.Equal(t, ErrNullCommitData, r.checkCommit(cp))

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(state)
	header := &types.Header{Height: 1}
	cp.Batch = &rtypes.BlockBatch{BlockHeaders: []*types.Header{header}}
	require.Equal(t, ErrChainTitle, r.checkCommit(cp))
	cp.ChainTitle = title
	require.Equal(t, ErrOutOfOrderCommit, r.checkCommit(cp))
	_ = state.Set(formatRollupStatusKey(title), []byte("errorData"))
	require.Equal(t, ErrGetRollupStatus, r.checkCommit(cp))
	_ = state.Set(formatRollupStatusKey(title), types.Encode(&rtypes.RollupStatus{}))
	cp.CommitRound = 1
	api.On("Query", mock.Anything, "GetNodeGroupStatus", mock.Anything).Return(nil, types.ErrActionNotSupport).Once()
	require.Equal(t, ErrGetValPubs, r.checkCommit(cp))

	parentHash := []byte("parentHash")
	status := &rtypes.RollupStatus{
		CommitBlockHash:   common.ToHex(parentHash),
		CommitBlockHeight: 1,
	}
	_ = state.Set(formatRollupStatusKey(title), types.Encode(status))
	require.Equal(t, ErrOutOfOrderCommit, r.checkCommit(cp))
	header.ParentHash = parentHash
	require.Equal(t, ErrOutOfOrderCommit, r.checkCommit(cp))
	header.Height = 1
	status.BlockFragIndex = 1
	status.CommitBlockHash = calcBlockHash(header)
	header1 := &types.Header{Height: 2, ParentHash: common.Sha256(types.Encode(header))}
	cp.Batch.BlockHeaders = []*types.Header{header, header1}
	_ = state.Set(formatRollupStatusKey(title), types.Encode(status))
	//blsDriver := bls.Driver{}
	priv, _ := blsDriver.GenKey()
	paraNodeStatus := &paratypes.ParaNodeGroupStatus{BlsPubKeys: common.ToHex(priv.PubKey().Bytes())}
	api.On("Query", mock.Anything, "GetNodeGroupStatus", mock.Anything).Return(paraNodeStatus, nil)

	require.Equal(t, ErrInvalidValidator, r.checkCommit(cp))
	cp.ValidatorPubs = [][]byte{[]byte("invalidPub")}
	require.Equal(t, ErrInvalidValidator, r.checkCommit(cp))
	cp.ValidatorPubs = [][]byte{priv.PubKey().Bytes()}
	require.Equal(t, ErrInvalidValidatorSign, r.checkCommit(cp))
	cp.AggregateValidatorSign = priv.Sign(common.Sha256(types.Encode(cp.GetBatch()))).Bytes()
	require.Nil(t, r.checkCommit(cp))

	var valSigs, txSigs []crypto.Signature
	var privs []crypto.PrivKey
	var blsPubKeys []string
	for i := 0; i < 10; i++ {
		data := types.Encode(&types.Transaction{Payload: []byte(fmt.Sprintf("test%d", i))})
		cp.Batch.TxList = append(cp.Batch.TxList, data)
		priv, _ = blsDriver.GenKey()
		privs = append(privs, priv)
		blsPubKeys = append(blsPubKeys, common.ToHex(priv.PubKey().Bytes()))
		cp.Batch.PubKeyList = append(cp.Batch.PubKeyList, priv.PubKey().Bytes())
		txSigs = append(txSigs, priv.Sign(data))
	}
	aggSig, err := blsDriver.(crypto.AggregateCrypto).Aggregate(txSigs)
	require.Nil(t, err)
	cp.Batch.AggregateTxSign = aggSig.Bytes()
	batchHash := common.Sha256(types.Encode(cp.GetBatch()))
	for i := range cp.Batch.TxList {
		valSigs = append(valSigs, privs[i].Sign(batchHash))
	}

	paraNodeStatus.BlsPubKeys = strings.Join(blsPubKeys, ",")
	cp.ValidatorPubs = cp.Batch.PubKeyList[:7]
	aggSig, err = blsDriver.(crypto.AggregateCrypto).Aggregate(valSigs[:7])
	require.Nil(t, err)
	cp.AggregateValidatorSign = aggSig.Bytes()
	require.Nil(t, r.checkCommit(cp))
}
