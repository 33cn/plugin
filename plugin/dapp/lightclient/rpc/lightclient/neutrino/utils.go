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
	n.commitKeyMu.RLock()
	defer n.commitKeyMu.RUnlock()
	return n.commitKey
}

// initCommitKey 在后台 goroutine 中等待钱包解锁后导出 commit 私钥。
// 先加写锁再启动 goroutine，确保 getCommitKey() 的调用者在 key 就绪之前阻塞，
// 从而保证 Start() 中后续依赖 commitKey 的流程（tss/rgbx）不会在 key 为空时执行。
func (n *neutrinoClient) initCommitKey() {
	n.initCommitKeyOnce.Do(func() {
		n.commitKeyMu.Lock()
		go func() {
			defer n.commitKeyMu.Unlock()
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
	})
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
