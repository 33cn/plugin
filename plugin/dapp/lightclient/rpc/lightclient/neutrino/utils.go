package neutrino

import (
	"fmt"
	"os"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

func (n *neutrinoClient) getCommitKey() crypto.PrivKey {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.commitKey
}

func (n *neutrinoClient) initCommitKey() {
	if n.commitKey != nil {
		return
	}
	n.lock.Lock()
	go func() {
		defer n.lock.Unlock()
		ticker := time.NewTicker(time.Second * 3)
		defer ticker.Stop()
		for {
			select {
			case <-n.ctx.Done():
				return
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
				_, key, err := getPrivKey(secp256k1.Name, resp.(*types.ReplyString).Data)
				if err != nil {
					log.Error("getKeyFromWallet", "addr", n.commitAddr, "getPrivKey err", err)
					continue
				}
				n.commitKey = key
				return
			}
		}
	}()
}

func getPrivKey(cryptoName, privKey string) (crypto.Crypto, crypto.PrivKey, error) {
	if privKey == "" {
		return nil, nil, fmt.Errorf("getPrivKey: empty privKey")
	}
	driver, err := crypto.Load(cryptoName, -1)
	if err != nil {
		return nil, nil, fmt.Errorf("getPrivKey load crypto driver err: %w", err)
	}
	privByte, err := common.FromHex(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("getPrivKey decode hex key err: %w", err)
	}
	key, err := driver.PrivKeyFromBytes(privByte)
	if err != nil {
		return nil, nil, fmt.Errorf("getPrivKey priv key from bytes err: %w", err)
	}

	return driver, key, nil
}

func (n *neutrinoClient) waitUntilDone(_ string, done func() bool, interval time.Duration) {
	if done() {
		return
	}
	if interval <= 0 {
		interval = time.Second * 3
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {

		case <-ticker.C:
			if done() {
				return
			}
		case <-n.ctx.Done():
			return
		}
	}
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
