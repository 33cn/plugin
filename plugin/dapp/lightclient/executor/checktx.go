package executor

import (
	"encoding/hex"
	"errors"

	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	"github.com/btcsuite/btcd/blockchain"
)

// CheckTx 实现自定义检验交易接口，供框架调用
func (l *lightclient) CheckTx(tx *types.Transaction, index int) error {

	action := &ltypes.LightClientAction{}

	err := types.Decode(tx.GetPayload(), action)
	if err != nil {
		elog.Error("CheckTx", "txHash", hex.EncodeToString(tx.Hash()), "Decode payload error", err)
		return ErrDecodeAction
	}

	if action.Ty == ltypes.TyBtcHeadersAction {

		err = l.checkBtcHeaders(tx, action.GetBtcHeaders())
	} else {
		err = types.ErrActionNotSupport
	}
	if err != nil {
		elog.Error("CheckTx", "txHash", hex.EncodeToString(tx.Hash()), "actionName", tx.ActionName(), "err", err)
	}
	return err
}

func (l *lightclient) checkBtcHeaders(tx *types.Transaction, headers *ltypes.BtcHeaders) error {

	if tx.From() != lightCfg.CommitAddress {

		elog.Error("checkBtcHeaders", "from", tx.From(), "configAddress", lightCfg.CommitAddress)
		return ErrIllegalCommitAddress
	}

	prevHeader, err := getBtcLastHeader(l.GetStateDB())
	if err != nil {
		elog.Error("checkBtcHeaders", "getBtcLastHeader err", err)
		return ErrBtcGetLastHeader
	}

	if len(headers.GetHeaders()) < 1 {
		elog.Error("checkBtcHeaders", "err", "commit empty headers")
		return types.ErrInvalidParam
	}

	params := ltypes.GetBtcChainParams(lightCfg.BtcNetName)
	chainCtx := newBtcChainContext(params)
	timeSource := blockchain.NewMedianTime()
	isBootstrap := prevHeader == nil
	var prevCtx blockchain.HeaderCtx
	if prevHeader != nil {
		prevCtx = newBtcHeaderContext(prevHeader, nil, l.GetLocalDB())
	}

	for _, h := range headers.GetHeaders() {

		if h == nil {
			elog.Error("checkBtcHeaders nil header")
			return types.ErrInvalidParam
		}
		if prevHeader != nil && (prevHeader.Height+1 != h.GetHeight() || prevHeader.Hash != h.PreviousHash) {
			elog.Error("checkBtcHeaders", "prevHeight", prevHeader.Height, "prevHash", prevHeader.Hash,
				"commitHeight", h.GetHeight(), "commitPrevHash", h.GetPreviousHash())
			return ErrBtcHeaderDisorder
		}

		btcHeader, err := toWireHeader(h)
		if err != nil {
			elog.Error("checkBtcHeaders", "height", h.GetHeight(), "hash", h.GetHash(), "toWireHeader err", err)
			return ErrToBtcWireHeader
		}
		hash := btcHeader.BlockHash()
		if hash.String() != h.Hash {
			elog.Error("checkBtcHeaders", "expectHash", hash, "height", h.GetHeight(), "hash", h.GetHash(), "err", err)
			return ErrInvalidBtcBlockHash
		}

		if err = blockchain.CheckBlockHeaderSanity(btcHeader, params.PowLimit, timeSource, blockchain.BFNone); err != nil {
			elog.Error("checkBtcHeaders CheckBlockHeaderSanity", "height", h.GetHeight(), "hash", h.GetHash(), "err", err)
			return mapBtcHeaderVerifyErr(err)
		}
		// 首次提交时(prevCtx=nil)无法获取前置上下文，首个header只做sanity校验。
		if prevCtx != nil {
			// 首次导入且刚好在难度调整点，如果缺少历史祖先，则仅校验sanity与顺序。
			if isBootstrap && !canValidateHeaderContext(prevCtx, chainCtx) {
				prevHeader = h
				prevCtx = newBtcHeaderContext(h, prevCtx, l.GetLocalDB())
				continue
			}
			if err = blockchain.CheckBlockHeaderContext(btcHeader, prevCtx, blockchain.BFNone, chainCtx, true); err != nil {
				elog.Error("checkBtcHeaders CheckBlockHeaderContext", "height", h.GetHeight(), "hash", h.GetHash(), "err", err)
				return mapBtcHeaderVerifyErr(err)
			}
		}
		prevHeader = h
		prevCtx = newBtcHeaderContext(h, prevCtx, l.GetLocalDB())

	}

	return nil

}

func mapBtcHeaderVerifyErr(err error) error {
	var ruleErr blockchain.RuleError
	if errors.As(err, &ruleErr) && ruleErr.ErrorCode == blockchain.ErrUnexpectedDifficulty {
		return ErrBtcTargetBits
	}
	return ErrInvalidBtcBlockHash
}

func canValidateHeaderContext(prevCtx blockchain.HeaderCtx, chainCtx *btcChainContext) bool {
	if prevCtx == nil {
		return false
	}
	if (prevCtx.Height()+1)%chainCtx.BlocksPerRetarget() != 0 {
		return true
	}
	return prevCtx.RelativeAncestorCtx(chainCtx.BlocksPerRetarget()-1) != nil
}
