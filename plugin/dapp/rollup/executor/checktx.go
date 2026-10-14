package executor

import (
	"encoding/hex"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"

	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

// CheckTx 实现自定义检验交易接口，供框架调用
func (r *rollup) CheckTx(tx *types.Transaction, index int) error {
	txHash := hex.EncodeToString(tx.Hash())
	var action rtypes.RollupAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		elog.Error("rollup CheckTx", "txHash", txHash, "Decode payload error", err)
		return types.ErrActionNotSupport
	}

	if action.Ty == rtypes.TyCommitAction {

		err = r.checkCommit(action.GetCommit())
	} else {
		err = types.ErrActionNotSupport
	}
	if err != nil {
		elog.Error("rollup CheckTx", "txHash", txHash, "actionName", tx.ActionName(), "err", err)
	}
	return err
}

// isValidatorDedupForkActive 2/3 门槛是否按去重后的验证人公钥数计算，见 types.ForkSecurityHardening
func (r *rollup) isValidatorDedupForkActive() bool {
	return r.isSecurityHardeningForkActive()
}

// isSecurityHardeningForkActive 本分叉下的提交校验(分段续接/title 规范/公钥列表长度)是否按加固后的规则执行
func (r *rollup) isSecurityHardeningForkActive() bool {
	return r.GetAPI().GetConfig().IsDappFork(r.GetHeight(), rtypes.RollupX, rtypes.ForkSecurityHardening)
}

// validatorSignThreshold 返回 2/3 门槛判定所用的 (提交侧计数, 验证人侧计数)。
// dedup 为 false 时沿用历史行为: 直接数提交列表与成员列表的条目数;
// 为 true 时两侧都按去重后的公钥数(见 types.ForkSecurityHardening)。
func validatorSignThreshold(submitted [][]byte, valPubs map[string]struct{}, memberCount int, dedup bool) (int, int) {
	if !dedup {
		return len(submitted), memberCount
	}
	distinct := make(map[string]struct{}, len(submitted))
	for _, pub := range submitted {
		distinct[hex.EncodeToString(pub)] = struct{}{}
	}
	return len(distinct), len(valPubs)
}

func (r *rollup) checkCommit(cp *rtypes.CheckPoint) error {

	commitRound := cp.GetCommitRound()
	if len(cp.GetBatch().GetBlockHeaders()) < 1 {
		elog.Error("checkCommit null data", "commitRound", commitRound,
			"blkHeaders", len(cp.GetBatch().GetBlockHeaders()))
		return ErrNullCommitData
	}

	// 必须是平行链: 分叉后要求规范形式的平行链title(user.p.<name>.),
	// 与执行侧一致, 防止用真实链title的前缀/去掉结尾'.'的形式建立提交
	if r.isSecurityHardeningForkActive() {
		if !ptypes.IsValidParaTitle(cp.GetChainTitle()) {
			return ErrChainTitle
		}
	} else if !strings.HasPrefix(cp.GetChainTitle(), types.ParaKeyX) {
		return ErrChainTitle
	}

	status, err := GetRollupStatus(r.GetStateDB(), cp.GetChainTitle())
	if err != nil {
		elog.Error("checkCommit", "title", cp.GetChainTitle(),
			"commitRound", commitRound, "getRollupStatus err", err)
		return ErrGetRollupStatus
	}

	// 检查提交是否有序, 区块高度及区块哈希
	errorOrder := func(cause string, checkHeight int64) error {
		elog.Error("checkCommit", "title", cp.GetChainTitle(),
			"commitRound", commitRound, "cause", cause, "checkHeight", checkHeight,
			"status", status.String())
		return ErrOutOfOrderCommit
	}

	if commitRound != status.CommitRound+1 {
		return errorOrder("commit round", 0)
	}

	parentHash := status.CommitBlockHash
	parentHeight := status.CommitBlockHeight

	// 首次提交无状态数据记录, 父区块哈希值提前设定
	if len(status.CommitBlockHash) == 0 {
		parentHash = common.ToHex(cp.GetBatch().GetBlockHeaders()[0].ParentHash)
	}

	// 区块数据过大被分割情况, 分割区块的区块头信息会被重复提交
	if status.GetBlockFragIndex() > 0 {
		if status.CommitBlockHash != calcBlockHash(cp.GetBatch().GetBlockHeaders()[0]) {
			return errorOrder("fragment block hash", cp.GetBatch().GetBlockHeaders()[0].GetHeight())
		}
		parentHash = common.ToHex(cp.GetBatch().GetBlockHeaders()[0].ParentHash)
		parentHeight = status.CommitBlockHeight - 1
	}

	for _, header := range cp.GetBatch().GetBlockHeaders() {

		if common.ToHex(header.GetParentHash()) != parentHash {

			elog.Error("checkCommit", "height", header.GetHeight(),
				"expectHash", parentHash, "actualHash", common.ToHex(header.GetParentHash()))
			return errorOrder("block hash", header.GetHeight())
		}
		parentHash = calcBlockHash(header)
		if parentHeight+1 != header.GetHeight() {
			return errorOrder("block height", header.GetHeight())
		}
		parentHeight++
	}

	// check validator
	pubs, err := r.getValidatorNodesBlsPubs(cp.GetChainTitle())
	if err != nil {
		elog.Error("checkCommit", "title", cp.GetChainTitle(), "commitRound", commitRound,
			"getValidatorNodesBlsPubs err", err)
		return ErrGetValPubs
	}

	valPubs := make(map[string]struct{}, len(pubs))
	for _, pub := range pubs {
		valPubs[rtypes.FormatHexPubKey(pub)] = struct{}{}
	}

	blsPubs := make([]crypto.PubKey, 0, len(cp.GetValidatorPubs()))
	for _, pub := range cp.GetValidatorPubs() {
		hexPub := hex.EncodeToString(pub)
		_, valid := valPubs[hexPub]
		if !valid {
			elog.Error("checkCommit", "title", cp.GetChainTitle(),
				"commitRound", commitRound, "invalid validator pub", hexPub)
			return ErrInvalidValidator
		}
		blsPub, _ := blsDriver.PubKeyFromBytes(pub)
		blsPubs = append(blsPubs, blsPub)
	}

	signCount, valCount := validatorSignThreshold(cp.GetValidatorPubs(), valPubs, len(pubs),
		r.isValidatorDedupForkActive())
	if signCount < valCount*2/3+1 {
		elog.Error("checkCommit", "title", cp.GetChainTitle(), "commitRound", commitRound,
			"signCount", signCount, "valCount", valCount, "err", "not enough validator")
		return ErrInvalidValidator
	}
	blsSig, _ := blsDriver.SignatureFromBytes(cp.GetAggregateValidatorSign())
	signMsg := common.Sha256(types.Encode(cp.GetBatch()))
	aggreDriver := blsDriver.(crypto.AggregateCrypto)
	err = aggreDriver.VerifyAggregatedOne(blsPubs, signMsg, blsSig)
	if err != nil {
		elog.Error("checkCommit", "title", cp.GetChainTitle(),
			"commitRound", commitRound, "verify bls sig err", err)
		return ErrInvalidValidatorSign
	}

	// 精简模式, 不校验交易数据
	if len(cp.GetBatch().GetTxList()) == 0 {
		return nil
	}
	// 分叉后校验公钥列表与交易列表一一对应, 避免下面按下标取公钥时越界
	if r.isSecurityHardeningForkActive() &&
		len(cp.GetBatch().GetPubKeyList()) != len(cp.GetBatch().GetTxList()) {
		elog.Error("checkCommit", "title", cp.GetChainTitle(), "commitRound", commitRound,
			"txCount", len(cp.GetBatch().GetTxList()), "pubKeyCount", len(cp.GetBatch().GetPubKeyList()),
			"err", "pubkey list length not match tx list")
		return ErrTxPubKeyCountNotMatch
	}
	txCount := len(cp.GetBatch().GetTxList())
	txPubs := make([]crypto.PubKey, 0, txCount)
	for i, tx := range cp.GetBatch().GetTxList() {
		pub := cp.GetBatch().GetPubKeyList()[i]
		txPub, err := blsDriver.PubKeyFromBytes(pub)
		if err != nil {
			elog.Error("checkCommit", "title", cp.GetChainTitle(), "commitRound", commitRound,
				"txHash", hex.EncodeToString(common.Sha256(tx)),
				"txPub", hex.EncodeToString(pub), "err", err)
			return ErrInvalidBlsPub
		}
		txPubs = append(txPubs, txPub)
	}

	aggreTxSign, _ := blsDriver.SignatureFromBytes(cp.GetBatch().GetAggregateTxSign())
	err = aggreDriver.VerifyAggregatedN(txPubs, cp.GetBatch().GetTxList(), aggreTxSign)

	if err != nil {
		elog.Error("checkCommit", "title", cp.GetChainTitle(),
			"commitRound", commitRound, "verify aggregate tx sig err", err)
		return ErrInvalidTxAggregateSign
	}

	return nil
}
