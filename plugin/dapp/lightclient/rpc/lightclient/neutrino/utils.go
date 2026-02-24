package neutrino

import (
	"fmt"
	"os"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

func (n *neutrinoClient) getCommitKey() crypto.PrivKey {
	n.lock.Lock()
	defer n.lock.Unlock()
	if n.commitKey != nil {
		return n.commitKey
	}
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

			resp, err = n.chain33Api.ExecWalletFunc("wallet", "DumpPrivkey", &types.ReqString{Data: n.commitAddr})
			if err != nil {
				log.Error("getKeyFromWallet", "addr", n.commitAddr, "dump priv key err", err)
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

func (n *neutrinoClient) getRgbxWithdrawAsset(txHash []byte) (*rtypes.WithdrawAsset, error) {
	if len(txHash) == 0 {
		return nil, types.ErrInvalidParam
	}
	reply, err := n.mainChainGrpc.QueryTransaction(n.ctx, &types.ReqHash{Hash: txHash})
	if err != nil {
		log.Error("getRgbxWithdrawAsset", "txHash", fmt.Sprintf("%x", txHash), "err", err)
		return nil, err
	}

	action := &rtypes.RgbxAction{}
	if err := types.Decode(reply.GetTx().Payload, action); err != nil {
		log.Error("getRgbxWithdrawAsset decode action", "txHash", fmt.Sprintf("%x", txHash), "err", err)
		return nil, err
	}
	if action.Ty != rtypes.TyWithDrawAsset {
		return nil, fmt.Errorf("withdraw action not found")
	}

	return action.GetWithdraw(), nil
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func estimateBtcFee(tx *wire.MsgTx, feeRate btcutil.Amount) btcutil.Amount {
	txSize := tx.SerializeSize() + len(tx.TxIn)*108 // 估算witness大小
	return btcutil.Amount(txSize) * feeRate
}
