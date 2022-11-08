package executor

import (
	"errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

//Exec_CommitRollup exec commit rollup
func (e *Paracross) Exec_CommitRollup(payload *pt.CommitRollup, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.commitRollup(payload)
}

//ExecLocal_CommitRollup exec local commit rollup
func (e *Paracross) ExecLocal_CommitRollup(payload *pt.CommitRollup, tx *types.Transaction,
	receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {

	set := &types.LocalDBSet{}
	return set, nil
}

//ExecDelLocal_CommitRollup exec local commit rollup
func (e *Paracross) ExecDelLocal_CommitRollup(payload *pt.CommitRollup, tx *types.Transaction,
	receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	return set, nil
}

func (a *action) getRollupStatus(title string) (*rtypes.RollupStatus, error) {

	req := &rtypes.ChainTitle{Value: title}

	reply, err := a.api.Query(rtypes.RollupX, "GetRollupStatus", req)
	status := reply.(*rtypes.RollupStatus)
	return status, err
}

func (a *action) getRollupCommitRound(title string, commitRound int64) (*rtypes.CommitRoundInfo, error) {

	req := &rtypes.ReqGetCommitRound{
		CommitRound: commitRound,
		ChainTitle:  title,
	}

	reply, err := a.api.Query(rtypes.RollupX, "GetCommitRoundInfo", req)
	status := reply.(*rtypes.CommitRoundInfo)
	return status, err
}

var (
	ErrInvalidCommitRound   = errors.New("ErrInvalidCommitRound")
	ErrInvalidChain         = errors.New("ErrInvalidChain")
	ErrGetRollupCrossTx     = errors.New("ErrGetRollupCrossTx")
	ErrGetRollupCommitRound = errors.New("ErrGetRollupCommitRound")
	ErrCrossTxCheckHash     = errors.New("ErrCrossTxCheckHash")
)

func (a *action) commitRollup(commit *pt.CommitRollup) (*types.Receipt, error) {
	clog.Debug("commitRollup", "title", commit.GetChainTitle(), "commitRound", commit.GetCommitRound())
	if a.api.GetConfig().IsPara() {
		return nil, ErrInvalidChain
	}

	status, err := a.getRollupStatus(commit.GetChainTitle())

	if err != nil || status.CommitRound != commit.GetCommitRound() {

		clog.Error("commitRollup", "currRound", status.GetCommitRound(),
			"commitRound", commit.GetCommitRound(), "getRollupStatus err", err)
		return nil, ErrInvalidCommitRound
	}

	commitRound, err := a.getRollupCommitRound(commit.GetChainTitle(), commit.GetCommitRound())
	if err != nil {
		clog.Error("commitRollup", "commitRound", commit.GetCommitRound(), "getRollupCommitRound err", err)
		return nil, ErrGetRollupCommitRound
	}

	crossTxHashes, crossTxs, err := a.getRollupCrossTxs(commit.GetTxIndices())

	if err != nil {
		clog.Error("commitRollup", "commitRound", commit.GetCommitRound(), "getRollupCrossTxs err", err)
		return nil, ErrGetRollupCrossTx
	}

	checkHash := common.ToHex(CalcTxHashsHash(crossTxHashes))

	if commitRound.CrossTxCheckHash != checkHash {
		clog.Error("commitRollup", "commitRound", commit.GetCommitRound(),
			"calcHash", checkHash, "commitHash", commitRound.CrossTxCheckHash)

		for i, hash := range crossTxHashes {
			clog.Error("commitRollup cross tx info", "index", commit.GetTxIndices()[i].String(), "txhash", common.ToHex(hash))
		}
		return nil, ErrCrossTxCheckHash
	}

	crossTxResults, _ := common.FromHex(commitRound.CrossTxResults)

	receipt, err := a.execCrossTxs(commit.GetChainTitle(), commit.GetCommitRound(), crossTxs, crossTxResults)
	if err != nil {

		clog.Error("commitRollup", "commitRound", commit.GetCommitRound(), "execCrossTxs err", err)
		return nil, err
	}
	receipt.Ty = types.ExecOk
	return receipt, nil
}

func (a *action) getRollupCrossTxs(idxArr []*pt.CrossTxIndex) ([][]byte, []*types.Transaction, error) {

	blkCrossTxCache := make(map[int64][]*types.Transaction, len(idxArr)/2)
	crossTxs := make([]*types.Transaction, 0, len(idxArr))
	crossTxHashes := make([][]byte, 0, len(idxArr))
	cfg := a.api.GetConfig()
	for _, txIdx := range idxArr {

		// first get from cache
		blkCrossTxs, ok := blkCrossTxCache[txIdx.BlockHeight]
		if !ok {

			// get block from blockchain
			detail, err := getBlockByHeight(a.api, txIdx.BlockHeight)
			if err != nil {
				return nil, nil, err
			}

			blkCrossTxs = FilterParaCrossTxs(FilterTxsForPara(a.api.GetConfig(), detail.FilterParaTxsByTitle(cfg, cfg.GetTitle())))
			blkCrossTxCache[txIdx.BlockHeight] = blkCrossTxs
		}

		crossTx := blkCrossTxs[txIdx.FilterIndex]
		crossTxs = append(crossTxs, crossTx)
		crossTxHashes = append(crossTxHashes, crossTx.Hash())

	}

	return crossTxHashes, crossTxs, nil

}
