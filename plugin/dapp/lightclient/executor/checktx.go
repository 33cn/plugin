package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/common/difficulty"
	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
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
	if err != nil && err != types.ErrNotFound {
		elog.Error("checkBtcHeaders", "getBtcLastHeader err", err)
		return ErrBtcGetLastHeader
	}

	if len(headers.GetHeaders()) < 1 {
		elog.Error("checkBtcHeaders", "err", "commit empty headers")
		return types.ErrInvalidParam
	}

	for _, h := range headers.GetHeaders() {

		if h == nil {
			elog.Error("checkBtcHeaders nil header")
			return types.ErrInvalidParam
		}
		if prevHeader.Height > 0 && (prevHeader.Height+1 != h.GetHeight() || prevHeader.Hash != h.PreviousHash) {
			elog.Error("checkBtcHeaders", "prevHeight", prevHeader.Height, "prevHash", prevHeader.Hash,
				"commitHeight", h.GetHeight(), "commitPrevHash", h.GetPreviousHash())
			return ErrBtcHeaderDisorder
		}

		targetBits := calcNextRequiredDifficulty(prevHeader.Time, prevHeader.Bits, prevHeader.Height, l.GetLocalDB())
		if targetBits != 0 && h.Bits != targetBits {
			elog.Error("checkBtcHeaders", "height", h.GetHeight(), "hash", h.GetHash(),
				"targetBits", targetBits, "commitBits", h.Bits)
			return ErrBtcTargetBits
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

		// The block hash must be less than the claimed target.
		hashNum := difficulty.HashToBig(hash[:])
		target := difficulty.CompactToBig(uint32(h.Bits))
		if hashNum.Cmp(target) > 0 {
			elog.Error("checkBtcHeaders", hash, "height", h.GetHeight(), "hash", h.GetHash(), "bits", h.Bits)
			return ErrInvalidBtcBlockHash
		}
		prevHeader = h

	}

	return nil

}
