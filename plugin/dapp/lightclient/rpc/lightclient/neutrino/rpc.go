package neutrino

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
)

func (n *neutrinoClient) getCrossChainInfo() *rtypes.CrossChainInfo {

	req := &types.ReqString{Data: "btc"}
	reply, err := n.mainChainGrpc.QueryChain(n.ctx, &types.ChainExecutor{
		Driver:   rtypes.RgbxX,
		FuncName: "GetCrossChainInfo",
		Param:    types.Encode(req),
	})

	if err != nil || !reply.GetIsOk() {
		log.Error("getCrossChainInfo", "msg", string(reply.GetMsg()), "query err", err)
		return nil
	}

	msg := &rtypes.CrossChainInfo{}

	err = types.Decode(reply.GetMsg(), msg)
	if err != nil {
		log.Error("getValidatorPubKeys", "decode err", err)
		return nil
	}

	return msg
}

func (n *neutrinoClient) createTx(exec, action string, payload []byte) (*types.Transaction, error) {

	req := &types.CreateTxIn{
		Execer:     []byte(exec),
		Payload:    payload,
		ActionName: action,
	}
	reply, err := n.mainChainGrpc.CreateTransaction(n.ctx, req)
	if err != nil {
		log.Error("createTx", "exec", exec, "action", action, "err", err)
		return nil, err
	}
	tx := &types.Transaction{}
	err = types.Decode(reply.GetData(), tx)
	return tx, err
}

func (n *neutrinoClient) getProperFeeRate() int64 {

	reply, err := n.mainChainGrpc.GetProperFee(n.ctx, &types.ReqProperFee{})
	if err != nil {
		log.Error("getProperFeeRate", "err", err)
	} else {
		n.chain33FeeRate = reply.GetProperFee()
	}

	return n.chain33FeeRate
}

func (n *neutrinoClient) sendTx2MainChain(tx *types.Transaction) error {

	reply, err := n.mainChainGrpc.SendTransaction(n.ctx, tx)
	if err == nil && !reply.GetIsOk() {
		err = errors.New(string(reply.GetMsg()))
	}
	return err
}

func (n *neutrinoClient) getRgbxConfirmedHeight() *types.Int64 {
	reply, err := n.mainChainGrpc.QueryChain(n.ctx, &types.ChainExecutor{
		Driver:   rtypes.RgbxX,
		FuncName: "GetConfirmedHeight",
	})
	if err != nil {
		log.Error("getRgbxConfirmedHeight", "query err", err)
		return nil
	}

	data := &types.Int64{}
	err = types.Decode(reply.GetMsg(), data)
	if err != nil {
		log.Error("getRgbxConfirmedHeight", "decode err", err)
		return nil
	}
	return data
}

func (n *neutrinoClient) getRgbxPendingTxs(req *rtypes.ReqListPendingTx) (*rtypes.PendingTxs, error) {
	reply, err := n.mainChainGrpc.QueryChain(n.ctx, &types.ChainExecutor{
		Driver:   rtypes.RgbxX,
		FuncName: "ListPendingTx",
		Param:    types.Encode(req),
	})
	if err != nil {
		log.Error("getRgbxPendingTxs", "query err", err, "req", req.String())
		return nil, err
	}

	data := &rtypes.PendingTxs{}
	err = types.Decode(reply.GetMsg(), data)
	if err != nil {
		log.Error("getRgbxPendingTxs", "decode err", err)
		return nil, err
	}
	return data, nil
}

func (n *neutrinoClient) getTxDetail(txHash []byte) (*types.TransactionDetail, error) {
	return n.mainChainGrpc.QueryTransaction(n.ctx, &types.ReqHash{Hash: txHash})
}

func (n *neutrinoClient) getRgbxWithdrawAsset(txHash []byte) (*rtypes.WithdrawAsset, error) {
	if len(txHash) == 0 {
		return nil, types.ErrInvalidParam
	}
	reply, err := n.getTxDetail(txHash)
	if err != nil {
		log.Error("getRgbxWithdrawAsset", "txHash", fmt.Sprintf("%x", txHash), "get txDetail err", err)
		return nil, err
	}

	action := &rtypes.RgbxAction{}
	if err := types.Decode(reply.GetTx().Payload, action); err != nil {
		log.Error("getRgbxWithdrawAsset decode action", "txHash", fmt.Sprintf("%x", txHash), "err", err)
		return nil, err
	}
	if action.Ty != rtypes.TyWithDrawAsset {
		log.Error("getRgbxWithdrawAsset invalid action", "txHash", fmt.Sprintf("%x", txHash), "action", action.String())
		return nil, fmt.Errorf("withdraw action not found")
	}

	return action.GetWithdraw(), nil
}

func (n *neutrinoClient) getRgbxPendingTx(height, index int64) (*rtypes.PendingTx, error) {
	req := &rtypes.ReqGetPendingTx{
		Height: height,
		Index:  index,
	}
	reply, err := n.mainChainGrpc.QueryChain(n.ctx, &types.ChainExecutor{
		Driver:   rtypes.RgbxX,
		FuncName: "GetPendingTx",
		Param:    types.Encode(req),
	})
	if err != nil {
		log.Error("getRgbxPendingTx", "height", height, "index", index, "query err", err)
		return nil, err
	}
	data := &rtypes.PendingTx{}
	if err = types.Decode(reply.GetMsg(), data); err != nil {
		log.Error("getRgbxPendingTx", "height", height, "index", index, "decode err", err)
		return nil, err
	}
	return data, nil
}

func (n *neutrinoClient) getRgbxPendingTxByHash(txHash []byte) (*rtypes.PendingTx, error) {
	if len(txHash) == 0 {
		return nil, types.ErrInvalidParam
	}
	detail, err := n.getTxDetail(txHash)
	if err != nil {
		log.Error("getRgbxPendingTxByHash", "txHash", fmt.Sprintf("%x", txHash), "get txDetail err", err)
		return nil, err
	}
	return n.getRgbxPendingTx(detail.GetHeight(), detail.GetIndex())
}

func (n *neutrinoClient) submitMainchainTx(exec string, action string, payload types.Message) (string, error) {
	tx, err := n.createTx(exec, action, types.Encode(payload))
	if err != nil {
		log.Error("submitMainchainTx", "createTx err", err)
		return "", err
	}
	txHash := hex.EncodeToString(tx.Hash())
	tx.Fee, err = tx.GetRealFee(n.getProperFeeRate())
	if err != nil {
		log.Error("submitMainchainTx", "txHash", txHash, "GetRealFee err", err)
		return "", err
	}
	tx.Sign(types.EncodeSignID(secp256k1.ID, n.commitAddressType), n.getCommitKey())
	err = n.sendTx2MainChain(tx)
	if err != nil {
		log.Error("submitMainchainTx", "txHash", txHash, "sendTx2MainChain err", err)
		return "", err
	}
	return txHash, nil
}

func (n *neutrinoClient) submitMainchainTxUntilSuccess(exec string, action string, payload types.Message) {

	n.waitUntilDone("submitMainchainTxUntilSuccess", func() bool {
		_, err := n.submitMainchainTx(exec, action, payload)
		return err == nil
	}, 0)
}

func (n *neutrinoClient) getMainchainHeight() int64 {
	reply, err := n.mainChainGrpc.GetLastHeader(n.ctx, &types.ReqNil{})
	if err != nil {
		log.Error("getMainchainHeight", "query err", err)
		return 0
	}
	return reply.GetHeight()
}
