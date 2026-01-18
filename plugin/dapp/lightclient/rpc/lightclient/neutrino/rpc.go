package neutrino

import (
	"errors"

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
