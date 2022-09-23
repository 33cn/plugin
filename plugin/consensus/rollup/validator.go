package rollup

import (
	"bytes"
	"encoding/hex"
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
	blsKey           crypto.PrivKey
	validators       map[string]int

	blsDriver crypto.Crypto
	status    *rolluptypes.RollupStatus
}

func (v *validator) init(cfg Config, valPubs []string, status *rolluptypes.RollupStatus) {

	var err error
	v.blsDriver, err = crypto.Load(bls.Name, -1)
	if err != nil {
		panic("load bls driver err:" + err.Error())
	}
	privByte, err := common.FromHex(cfg.CommitBlsKey)
	if err != nil {
		panic("decode hex bls key err:" + err.Error())
	}
	key, err := v.blsDriver.PrivKeyFromBytes(privByte)
	if err != nil {
		panic("new bls priv key err:" + err.Error())
	}

	v.blsKey = key
	v.updateValidators(valPubs)

}

// 获取本节点下一个提交轮数
func (v *validator) getNextCommitRound() int64 {

	v.lock.RLock()
	defer v.lock.RUnlock()

	commitRound := v.status.GetCommitRound()
	valCount := int64(len(v.validators))
	for {
		commitRound++
		if int32(commitRound%valCount) == v.commitRoundIndex {
			break
		}
	}
	return commitRound
}

func (v *validator) updateRollupStatus(status *rolluptypes.RollupStatus) {

	v.lock.Lock()
	defer v.lock.Unlock()

	v.status = status
}

// 其他节点未提交相应round数据, 导致超时
func (v *validator) isRollupCommitTimeout() (currRound int64, timeout bool) {

	v.lock.RLock()
	defer v.lock.RUnlock()

	now := types.Now().Unix()

	if now-v.status.Timestamp >= rolluptypes.RollupCommitTimeout {
		return v.status.GetCommitRound(), true
	}

	return 0, false
}

func (v *validator) getValidatorCount() int {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return len(v.validators)
}

func (v *validator) updateValidators(valPubs []string) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.validators = make(map[string]int, len(valPubs))

	for i, pub := range valPubs {
		pub = rolluptypes.FormatHexPubKey(pub)
		v.validators[pub] = i
	}

	blsPub := hex.EncodeToString(v.blsKey.PubKey().Bytes())
	idx, ok := v.validators[blsPub]

	v.enable = ok
	v.commitRoundIndex = int32(idx)
}

func (v *validator) validateSignMsg(sign *rolluptypes.ValidatorSignMsg) bool {

	v.lock.RLock()
	defer v.lock.RUnlock()
	pub := hex.EncodeToString(sign.PubKey)

	_, ok := v.validators[pub]
	if !ok {
		rlog.Error("validateSignMsg invalid node", "round", sign.CommitRound, "pub", pub)
		return false
	}

	if err := v.blsDriver.Validate(sign.MsgHash, sign.PubKey, sign.Signature); err != nil {
		rlog.Error("validateSignMsg invalid sign",
			"round", sign.CommitRound, "pub", pub, "err", err)
		return false
	}
	return true
}

func (v *validator) sign(round int64, batch *rolluptypes.BlockBatch) *rolluptypes.ValidatorSignMsg {

	msg := common.Sha256(types.Encode(batch))
	sign := &rolluptypes.ValidatorSignMsg{}
	sign.Signature = v.blsKey.Sign(msg).Bytes()
	sign.PubKey = v.blsKey.PubKey().Bytes()
	sign.CommitRound = round
	sign.MsgHash = msg

	return sign
}

type aggreSignFunc = func(set *validatorSignMsgSet) (pubs [][]byte, aggreSign []byte)

func (v *validator) aggregateSign(set *validatorSignMsgSet) (pubs [][]byte, aggreSign []byte) {

	valCount := v.getValidatorCount()
	// 2/3 共识, 向上取整
	minSignCount := valCount * 2 / 3
	if valCount%3 != 0 {
		minSignCount++
	}
	if len(set.others)+1 < minSignCount {
		rlog.Debug("aggregateSign", "commitRound", set.self.CommitRound,
			"valCount", valCount, "signCount", len(set.others)+1)
		return nil, nil
	}

	pubs = make([][]byte, 0, len(set.others)+1)
	signs := make([]crypto.Signature, 0, len(set.others)+1)

	s, _ := v.blsDriver.SignatureFromBytes(set.self.Signature)
	signs = append(signs, s)
	pubs = append(pubs, set.self.PubKey)
	for i := 0; i < len(set.others); {
		sign := set.others[i]
		// 数据哈希不一致, 非法签名
		if !bytes.Equal(sign.MsgHash, set.self.MsgHash) {

			set.others = append(set.others[:i], set.others[i+1:]...)
			rlog.Error("aggregateSign msgHash not equal", "commitRound", set.self.CommitRound,
				"selfHash", hex.EncodeToString(set.self.MsgHash),
				"otherHash", hex.EncodeToString(sign.MsgHash))
			continue
		}
		s, _ = v.blsDriver.SignatureFromBytes(sign.GetSignature())
		signs = append(signs, s)
		pubs = append(pubs, sign.PubKey)
		i++
	}

	blsAggre := v.blsDriver.(crypto.AggregateCrypto)
	s, err := blsAggre.Aggregate(signs[:minSignCount])
	if err != nil {
		rlog.Error("aggregateSign", "commitRound", set.self.CommitRound, "aggre err", err)
		return nil, nil
	}

	return pubs[:minSignCount], s.Bytes()
}
