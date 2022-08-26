package rollup

import (
	"sync"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/crypto/bls"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

type validator struct {
	lock             sync.RWMutex
	enable           bool
	commitRoundIndex int32
	privKey          crypto.PrivKey
	validators       map[string]struct{}

	blsDriver crypto.Crypto
	status    *rolluptypes.RollupStatus
}

func (v *validator) init(cfg Config, vals []byte) {

	var err error
	v.blsDriver, err = crypto.Load(bls.Name, -1)
	if err != nil {
		panic("load bls err" + err.Error())
	}
}

// 获取本节点下一个提交轮数
func (v *validator) getNextCommitRound() int64 {

	v.lock.RLock()
	defer v.lock.RUnlock()

	commitRound := v.status.CurrCommitRound
	valCount := int64(len(v.validators))
	for {
		commitRound++
		if int32(commitRound%valCount) == v.commitRoundIndex {
			break
		}
	}
	return commitRound
}

// 其他节点未提交相应round数据, 导致超时
func (v *validator) isRollupCommitTimeout() (currRound int64, timeout bool) {

	v.lock.RLock()
	defer v.lock.RUnlock()

	now := types.Now().Unix()

	if now-v.status.Timestamp >= rolluptypes.RollupCommitTimeout {
		return v.status.CurrCommitRound, true
	}

	return 0, false
}

func (v *validator) getValidatorCount() int {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return len(v.validators)
}

func (v *validator) updateValidators(vals []byte) {
	v.lock.Lock()
	defer v.lock.Unlock()
}

func (v *validator) validateSignMsg(msg []byte, sign *rolluptypes.ValidatorSignMsg) bool {

	v.lock.RLock()
	defer v.lock.RUnlock()
	// check round

	// check pub key

	// check bls sign

	return true
}

func (v *validator) sign(round int64, batch *rolluptypes.BlockBatch) ([]byte, *rolluptypes.ValidatorSignMsg) {

	msg := common.Sha256(types.Encode(batch))
	sign := &rolluptypes.ValidatorSignMsg{}
	sign.Signature = v.privKey.Sign(msg).Bytes()
	sign.PubKey = v.privKey.PubKey().Bytes()
	sign.CommitRound = round

	return msg, sign
}

type aggreSignFunc = func(set *validatorSignMsgSet) (pubs [][]byte, aggreSign []byte)

func (v *validator) aggregateSign(set *validatorSignMsgSet) (pubs [][]byte, aggreSign []byte) {

	valCount := v.getValidatorCount()
	// 2/3 共识
	minSignCount := valCount * 2 / 3
	if len(set.signs) < minSignCount {
		rlog.Debug("aggregateSign", "valCount", valCount, "signCount", len(set.signs))
		return nil, nil
	}

	signs := make([]crypto.Signature, 0, minSignCount)
	for _, sign := range set.signs[:minSignCount] {
		s, _ := v.blsDriver.SignatureFromBytes(sign)
		signs = append(signs, s)
	}

	blsAggre := v.blsDriver.(crypto.AggregateCrypto)
	s, err := blsAggre.Aggregate(signs)
	if err != nil {
		rlog.Error("aggregateSign", "aggre err", err)
		return nil, nil
	}

	return set.pubs[:minSignCount], s.Bytes()
}
