package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"

	"github.com/33cn/chain33/types"
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

func (r *rollup) checkCommit(cp *rtypes.CheckPoint) error {

	status, err := r.getRollupStatus(cp.GetChainTitle())
	if err != nil {
		elog.Error("checkCommit", "getRollupStatus err", err)
		return ErrGetRollupStatus
	}
	commitRound := cp.GetCommitRound()
	if commitRound != status.CommitRound+1 {
		elog.Error("checkCommit", "currRound", status.CommitRound, "commitRound", commitRound)
		return ErrInvalidCommitRound
	}
	// check validator
	pubs, err := r.getValidatorNodesBlsPubs(cp.GetChainTitle())
	if err != nil {
		elog.Error("checkCommit", "commitRound", commitRound, "getValidatorNodesBlsPubs err", err)
		return ErrGetValPubs
	}

	valPubs := make(map[string]struct{}, len(pubs))
	for _, pub := range pubs {
		valPubs[rtypes.FormatHexPubKey(pub)] = struct{}{}
	}

	blsPubs := make([]crypto.PubKey, 0, len(cp.GetValidatorPubs()))
	for _, pub := range cp.GetValidatorPubs() {
		_, valid := valPubs[hex.EncodeToString(pub)]
		if !valid {
			return ErrInvalidValidator
		}
		blsPub, _ := blsDriver.PubKeyFromBytes(pub)
		blsPubs = append(blsPubs, blsPub)
	}
	blsSig, _ := blsDriver.SignatureFromBytes(cp.GetAggregateValidatorSign())
	signMsg := common.Sha256(types.Encode(cp.GetBatch()))
	aggreDriver := blsDriver.(crypto.AggregateCrypto)
	err = aggreDriver.VerifyAggregatedOne(blsPubs, signMsg, blsSig)
	if err != nil {
		elog.Error("checkCommit", "commitRound", commitRound, "verify bls sig err", err)
		return ErrInvalidValidatorSign
	}

	// check aggregate tx
	txCount := len(cp.GetBatch().GetTxList())
	txPubs := make([]crypto.PubKey, 0, txCount)
	txHashList := make([][]byte, 0, txCount)
	for i, tx := range cp.GetBatch().GetTxList() {
		txHash := common.Sha256(tx)
		pub := cp.GetBatch().GetPubKeyList()[i]
		txPub, err := blsDriver.PubKeyFromBytes(pub)
		if err != nil {
			elog.Error("checkCommit", "commitRound", commitRound,
				"txHash", hex.EncodeToString(txHash),
				"txPub", hex.EncodeToString(pub), "err", err)
			return ErrInvalidBlsPub
		}
		txHashList = append(txHashList, txHash)
		txPubs = append(txPubs, txPub)
	}
	aggreTxSign, _ := blsDriver.SignatureFromBytes(cp.GetBatch().GetAggregateTxSign())
	err = aggreDriver.VerifyAggregatedN(txPubs, cp.GetBatch().GetTxList(), aggreTxSign)

	if err != nil {
		elog.Error("checkCommit", "commitRound", commitRound, "verify aggregate tx sig err", err)
		return ErrInvalidTxAggregateSign
	}

	return nil
}
