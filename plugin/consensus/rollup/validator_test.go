package rollup

import (
	"encoding/hex"
	"testing"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/crypto/bls"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/require"
)

func newTestVal() (*validator, *rtypes.ValidatorPubs, string) {

	cfg := Config{}
	val := &validator{}

	blsDrv := &bls.Driver{}
	commitAddr, priv := util.Genaddress()
	cfg.AuthKey = common.ToHex(priv.Bytes())
	valPubs := &rtypes.ValidatorPubs{}
	for i := 0; i < 3; i++ {
		newPriv, _ := blsDrv.GenKey()
		valPubs.BlsPubs = append(valPubs.BlsPubs, hex.EncodeToString(newPriv.PubKey().Bytes()))
	}
	_, blsKey := bls.MustPrivKeyFromBytes(priv.Bytes())
	valPubs.BlsPubs = append(valPubs.BlsPubs, common.ToHex(blsKey.PubKey().Bytes()))
	val.init(cfg.AuthKey, valPubs, &rtypes.RollupStatus{})
	return val, valPubs, commitAddr
}

func TestInitValidator(t *testing.T) {

	val, valPubs, commitAddr := newTestVal()
	require.Equal(t, 4, val.getValidatorCount())
	require.True(t, val.enable)
	require.Equal(t, int32(3), val.commitRoundIndex)
	require.Equal(t, common.Sha256(types.Encode(valPubs)), val.valPubHash)
	require.Equal(t, commitAddr, val.commitAddr)
}

func TestIsMyCommitTurn(t *testing.T) {

	val, _, commitAddr := newTestVal()
	maxInterval := int64(10)
	val.status.Timestamp = types.Now().Unix()
	// 非本节点提交
	nextRound, isTurn := val.isMyCommitTurn(maxInterval)
	require.True(t, -1 == nextRound)
	require.False(t, isTurn)

	// 本节点提交
	val.status.CommitRound = 2
	nextRound, isTurn = val.isMyCommitTurn(maxInterval)
	require.True(t, 3 == nextRound)
	require.True(t, isTurn)

	// 预计超时, 上一轮非本节点提交
	val.status.CommitRound = 3
	val.status.Timestamp = types.Now().Unix() - maxInterval - 120
	nextRound, isTurn = val.isMyCommitTurn(maxInterval)
	require.True(t, -1 == nextRound)
	require.False(t, isTurn)

	// 预计超时, 且上一轮由本节点提交
	val.status.CommitAddr = commitAddr
	nextRound, isTurn = val.isMyCommitTurn(maxInterval)
	require.True(t, 4 == nextRound)
	require.True(t, isTurn)

	// 完全超时, 触发提交
	val.status.CommitAddr = "test-addr"
	val.status.Timestamp = types.Now().Unix() - maxInterval - 300
	require.True(t, 4 == nextRound)
	require.True(t, isTurn)
}

func TestUpdateValidator(t *testing.T) {

	val, valPubs, _ := newTestVal()

	val.updateValidators(valPubs)

	valPubs.BlsPubs = valPubs.BlsPubs[:3]

	val.updateValidators(valPubs)
	require.Equal(t, common.Sha256(types.Encode(valPubs)), val.valPubHash)
	require.Equal(t, 3, val.getValidatorCount())
	require.False(t, val.enable)
	require.True(t, 0 == val.commitRoundIndex)
	select {
	case <-val.exit:
	default:
		t.Error("validator exit chan not closed")
	}
}

func TestValidatorSign(t *testing.T) {

	val, valPubs, _ := newTestVal()
	round := int64(10)
	batch := &rtypes.BlockBatch{AggregateTxSign: []byte("test-sign")}
	sign := val.sign(round, batch)
	require.Equal(t, round, sign.CommitRound)
	require.Equal(t, common.Sha256(types.Encode(batch)), sign.MsgHash)

	sign.PubKey = []byte("test-pubkey")
	require.False(t, val.validateSignMsg(sign))

	valPub, err := common.FromHex(valPubs.BlsPubs[0])
	require.Nil(t, err)
	sign.PubKey = valPub
	require.False(t, val.validateSignMsg(sign))

	sign.PubKey = val.blsKey.PubKey().Bytes()
	require.True(t, val.validateSignMsg(sign))
	require.False(t, val.validateSignMsg(nil))
}

func TestAggreSign(t *testing.T) {

	val, valPubs, _ := newTestVal()

	pubs, _ := val.aggregateSign(nil)
	signSet := &validatorSignMsgSet{}
	pubs, sign := val.aggregateSign(signSet)
	require.Nil(t, pubs)
	require.Nil(t, sign)

	// 只有一个节点
	val.validators = make(map[string]int)
	val.validators[valPubs.BlsPubs[3]] = 0
	signSet.self = val.sign(1, &rtypes.BlockBatch{})
	pubs, sign = val.aggregateSign(signSet)
	require.Equal(t, 1, len(pubs))
	blsAggre := val.blsDriver.(crypto.AggregateCrypto)
	sig, err := val.blsDriver.SignatureFromBytes(sign)
	require.Nil(t, err)
	pubKeys := []crypto.PubKey{val.blsKey.PubKey()}
	err = blsAggre.VerifyAggregatedOne(pubKeys, signSet.self.MsgHash, sig)
	require.Nil(t, err)
	require.Nil(t, val.blsDriver.Validate(signSet.self.MsgHash, pubs[0], sign))

	// 有2个节点, 签名数量不足
	val.validators["other-node"] = 1
	pubs, _ = val.aggregateSign(signSet)
	require.Nil(t, pubs)

	// 包含非法签名, 正确签名数量不足
	errSign := &rtypes.ValidatorSignMsg{MsgHash: []byte("err-hash")}
	signSet.others = append(signSet.others, errSign)
	pubs, _ = val.aggregateSign(signSet)
	require.Nil(t, pubs)
	require.Equal(t, 0, len(signSet.others))

	// 包非法签名, 但正确签名数量满足要求, 正确流程
	blsKey, _ := val.blsDriver.GenKey()
	pubKeys = append(pubKeys, blsKey.PubKey())
	val.blsKey = blsKey
	otherSign := val.sign(1, &rtypes.BlockBatch{})
	signSet.others = append(signSet.others, errSign, otherSign)
	pubs, sign = val.aggregateSign(signSet)
	require.Equal(t, 1, len(signSet.others))
	require.Equal(t, 2, len(pubs))
	sig, err = val.blsDriver.SignatureFromBytes(sign)
	require.Nil(t, err)
	err = blsAggre.VerifyAggregatedOne(pubKeys, signSet.self.MsgHash, sig)
	require.Nil(t, err)
}
