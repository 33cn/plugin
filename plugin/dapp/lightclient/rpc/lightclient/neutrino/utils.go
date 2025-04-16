package neutrino

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"time"
)

func (n *neutrinoClient) getKeyFromWallet(addr string) crypto.PrivKey {

	ticker := time.NewTicker(time.Second * 3)
	for {

		select {
		case <-n.ctx.Done():
			return nil
		case <-ticker.C:

			resp, err := n.chain33Api.ExecWalletFunc("wallet", "GetWalletStatus", &types.ReqNil{})
			if err != nil {
				log.Error("getKeyFromWallet", "GetWalletStatus err", err)
				continue
			}
			if !resp.(*types.WalletStatus).GetIsHasSeed() {
				log.Info("getKeyFromWallet, wait wallet save seed...")
				continue
			}

			if resp.(*types.WalletStatus).GetIsWalletLock() {
				log.Info("getKeyFromWallet, wait wallet unlock...")
				continue
			}

			resp, err = n.chain33Api.ExecWalletFunc("wallet", "DumpPrivkey", &types.ReqString{Data: addr})
			if err != nil {
				log.Info("getKeyFromWallet", "addr", addr, "dump priv key err", err)
				continue
			}
			_, key := getPrivKey(secp256k1.Name, resp.(*types.ReplyString).Data)
			return key
		}
	}

}

func getPrivKey(cryptoName, privKey string) (crypto.Crypto, crypto.PrivKey) {

	if privKey == "" {
		panic("getPrivKey: empty  privKey")
	}
	driver, err := crypto.Load(cryptoName, -1)
	if err != nil {
		panic("getPrivKey load crypto driver err:" + err.Error())
	}
	privByte, err := common.FromHex(privKey)
	if err != nil {
		panic("getPrivKey decode hex key err:" + err.Error())
	}
	key, err := driver.PrivKeyFromBytes(privByte)
	if err != nil {
		panic("getPrivKey priv key from bytes err:" + err.Error())
	}

	return driver, key
}

func (n *neutrinoClient) isChain33Sync() bool {

	reply, err := n.chain33Api.IsSync()

	if err != nil {
		log.Error("isChain33Sync", "err", err)
		return false
	}

	return reply.GetIsOk()
}

func (n *neutrinoClient) waitTask(taskName string, isSatisfied func() bool) {

	ticker := time.NewTicker(time.Second * 5)
	for {

		select {

		case <-ticker.C:
			if isSatisfied() {
				return
			}
			log.Debug("waitTask", "taskName", taskName)
		case <-n.ctx.Done():
			return
		}
	}
}

func (n *neutrinoClient) getRgbxConfirmedHeight() *types.Int64 {

	reply, err := n.chain33Api.QueryChain(&types.ChainExecutor{
		Driver:   rtypes.RgbxX,
		FuncName: "GetConfirmedHeight",
	})

	if err != nil {
		log.Error("getRgbxConfirmedHeight", "query err", err)
		return nil
	}

	return reply.(*types.Int64)
}

func (n *neutrinoClient) getRgbxPendingTxs(req *rtypes.ReqListPendingTx) (*rtypes.PendingTxs, error) {

	reply, err := n.chain33Api.QueryChain(&types.ChainExecutor{
		Driver:   rtypes.RgbxX,
		FuncName: "ListPendingTx",
		Param:    types.Encode(req),
	})

	if err != nil {
		log.Error("getRgbxPendingTxs", "query err", err, "req", req.String())
		return nil, err
	}

	return reply.(*rtypes.PendingTxs), nil
}

func (n *neutrinoClient) getProperFeeRate() int64 {

	reply, err := n.chain33Api.GetProperFee(&types.ReqProperFee{})
	if err != nil {
		log.Error("getProperFeeRate", "err", err)
	} else {
		n.chain33FeeRate = reply.GetProperFee()
	}

	return n.chain33FeeRate
}
