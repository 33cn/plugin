package neutrino

import (
	"github.com/33cn/chain33/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcwallet/chain"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/btcsuite/btcwallet/walletdb"
)

type btcWallet struct {
	*wallet.Wallet
	chainParams chaincfg.Params
	chainClient chain.Interface
	db          walletdb.DB
}

func newBtcWallet(n *neutrinoClient) (*btcWallet, error) {

	bw := &btcWallet{}
	bw.chainParams = n.neutrinoCfg.ChainParams
	exist, db, err := openWalletDB(n.neutrinoCfg.DataDir, "btcwallet.db")
	if err != nil {
		log.Error("newBtcWallet open db error", "err", err)
		return nil, err
	}
	pubPass := []byte("hello")
	if !exist {
		err := wallet.CreateWatchingOnly(db, pubPass, &bw.chainParams, types.Now())
		if err != nil {
			log.Error("newBtcWallet create wallet error", "err", err)
			_ = db.Close()
			return nil, err
		}
	}

	w, err := wallet.Open(db, pubPass, nil, &bw.chainParams, 250)
	if err != nil {
		log.Error("newBtcWallet open wallet error", "err", err)
		_ = db.Close()
		return nil, err
	}
	bw.db = db
	bw.Wallet = w
	bw.chainClient = chain.NewNeutrinoClient(&bw.chainParams, n.neutrinoCS)
	return bw, nil
}

func (b *btcWallet) start() (err error) {

	if err = b.chainClient.Start(); err != nil {
		log.Error("btcwallet chainclient start error", "err", err)
		return err
	}
	b.Wallet.Start()
	b.Wallet.SynchronizeRPC(b.chainClient)
	return nil
}

func (b *btcWallet) stop() {
	b.Wallet.Stop()
	b.chainClient.Stop()
	_ = b.db.Close()
}
