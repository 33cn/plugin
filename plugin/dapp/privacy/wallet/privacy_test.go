// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"encoding/hex"
	"sync"
	"time"
	"unsafe"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	privacy "github.com/33cn/plugin/plugin/dapp/privacy/crypto"
	ty "github.com/33cn/plugin/plugin/dapp/privacy/types"
)

type PrivacyMock struct {
	walletOp  wcom.WalletOperate
	store     *privacyStore
	policy    *privacyPolicy
	tokenName string
	password  string
}

func (mock *PrivacyMock) Init(walletOp wcom.WalletOperate, password string) {
	mock.policy = &privacyPolicy{mtx: &sync.Mutex{}, rescanwg: &sync.WaitGroup{}}
	mock.tokenName = types.BTY
	mock.walletOp = walletOp
	mock.password = password
	mock.policy.Init(walletOp, nil)
	mock.store = mock.policy.store
}

func (mock *PrivacyMock) getPrivKeyByAddr(addr string) (crypto.PrivKey, error) {
	//获取指定地址在钱包里的账户信息
	Accountstor, err := mock.store.getAccountByAddr(addr)
	if err != nil {
		bizlog.Error("ProcSendToAddress", "GetAccountByAddr err:", err)
		return nil, err
	}

	//通过password解密存储的私钥
	prikeybyte, err := common.FromHex(Accountstor.GetPrivkey())
	if err != nil || len(prikeybyte) == 0 {
		bizlog.Error("ProcSendToAddress", "FromHex err", err)
		return nil, err
	}

	password := []byte(mock.password)
	privkey := wcom.CBCDecrypterPrivkey(password, prikeybyte)
	//通过privkey生成一个pubkey然后换算成对应的addr
	cr, err := crypto.New(types.GetSignName("", mock.walletOp.GetSignType()))
	if err != nil {
		bizlog.Error("ProcSendToAddress", "err", err)
		return nil, err
	}
	priv, err := cr.PrivKeyFromBytes(privkey)
	if err != nil {
		bizlog.Error("ProcSendToAddress", "PrivKeyFromBytes err", err)
		return nil, err
	}
	return priv, nil
}

func (mock *PrivacyMock) getPrivacykeyPair(addr string) (*privacy.Privacy, error) {
	if accPrivacy, _ := mock.store.getWalletAccountPrivacy(addr); accPrivacy != nil {
		privacyInfo := &privacy.Privacy{}
		copy(privacyInfo.ViewPubkey[:], accPrivacy.ViewPubkey)
		decrypteredView := wcom.CBCDecrypterPrivkey([]byte(mock.password), accPrivacy.ViewPrivKey)
		copy(privacyInfo.ViewPrivKey[:], decrypteredView)
		copy(privacyInfo.SpendPubkey[:], accPrivacy.SpendPubkey)
		decrypteredSpend := wcom.CBCDecrypterPrivkey([]byte(mock.password), accPrivacy.SpendPrivKey)
		copy(privacyInfo.SpendPrivKey[:], decrypteredSpend)

		return privacyInfo, nil
	}
	_, err := mock.getPrivKeyByAddr(addr)
	if err != nil {
		return nil, err
	}
	return nil, ty.ErrPrivacyNotEnabled

}

func (mock *PrivacyMock) getPrivacyKeyPairsOfWallet() ([]addrAndprivacy, error) {
	//通过Account前缀查找获取钱包中的所有账户信息
	WalletAccStores, err := mock.walletOp.GetWalletAccounts()
	if err != nil || len(WalletAccStores) == 0 {
		return nil, err
	}

	var infoPriRes []addrAndprivacy
	for _, AccStore := range WalletAccStores {
		if len(AccStore.Addr) != 0 {
			if privacyInfo, err := mock.getPrivacykeyPair(AccStore.Addr); err == nil {
				var priInfo addrAndprivacy
				priInfo.Addr = &AccStore.Addr
				priInfo.PrivacyKeyPair = privacyInfo
				infoPriRes = append(infoPriRes, priInfo)
			}
		}
	}

	if 0 == len(infoPriRes) {
		return nil, ty.ErrPrivacyNotEnabled
	}

	return infoPriRes, nil
}

func (mock *PrivacyMock) CreateUTXOs(sender string, pubkeypair string, amount int64, height int64, count int) {
	privacyInfo, _ := mock.policy.getPrivacyKeyPairs()
	dbbatch := mock.store.NewBatch(true)
	for n := 0; n < count; n++ {
		tx := mock.createPublic2PrivacyTx(&ty.ReqCreatePrivacyTx{
			AssetExec:  "coins",
			Tokenname:  mock.tokenName,
			ActionType: 1,
			Amount:     amount,
			From:       sender,
			Pubkeypair: pubkeypair,
		})
		if tx == nil {
			return
		}

		txhash := tx.Hash()
		txhashstr := hex.EncodeToString(txhash)
		var privateAction ty.PrivacyAction
		if err := types.Decode(tx.GetPayload(), &privateAction); err != nil {
			return
		}
		privacyOutput := privateAction.GetOutput()
		RpubKey := privacyOutput.GetRpubKeytx()
		totalUtxosLeft := len(privacyOutput.Keyoutput)
		utxoProcessed := make([]bool, len(privacyOutput.Keyoutput))
		for _, info := range privacyInfo {
			privacykeyParirs := info.PrivacyKeyPair

			var utxos []*ty.UTXO
			for indexoutput, output := range privacyOutput.Keyoutput {
				if utxoProcessed[indexoutput] {
					continue
				}
				priv, _ := privacy.RecoverOnetimePriKey(RpubKey, privacykeyParirs.ViewPrivKey, privacykeyParirs.SpendPrivKey, int64(indexoutput))
				recoverPub := priv.PubKey().Bytes()[:]
				if bytes.Equal(recoverPub, output.Onetimepubkey) {
					totalUtxosLeft--
					utxoProcessed[indexoutput] = true
					info2store := &ty.PrivacyDBStore{
						Txhash:           txhash,
						Tokenname:        mock.tokenName,
						Amount:           output.Amount,
						OutIndex:         int32(indexoutput),
						TxPublicKeyR:     RpubKey,
						OnetimePublicKey: output.Onetimepubkey,
						Owner:            *info.Addr,
						Height:           height,
						Txindex:          0,
						Blockhash:        common.Sha256([]byte("some test for hash")),
					}

					utxoGlobalIndex := &ty.UTXOGlobalIndex{
						Outindex: int32(indexoutput),
						Txhash:   txhash,
					}

					utxoCreated := &ty.UTXO{
						Amount: output.Amount,
						UtxoBasic: &ty.UTXOBasic{
							UtxoGlobalIndex: utxoGlobalIndex,
							OnetimePubkey:   output.Onetimepubkey,
						},
					}

					utxos = append(utxos, utxoCreated)
					mock.store.setUTXO(info2store, txhashstr, dbbatch)
				}
			}
		}
	}
	dbbatch.Write()
}

func (mock *PrivacyMock) createPublic2PrivacyTx(req *ty.ReqCreatePrivacyTx) *types.Transaction {
	cfg := mock.walletOp.GetAPI().GetConfig()
	viewPubSlice, spendPubSlice, err := parseViewSpendPubKeyPair(req.GetPubkeypair())
	if err != nil {
		return nil
	}
	amount := req.GetAmount()
	viewPublic := (*[32]byte)(unsafe.Pointer(&viewPubSlice[0]))
	spendPublic := (*[32]byte)(unsafe.Pointer(&spendPubSlice[0]))
	privacyOutput, err := generateOuts(viewPublic, spendPublic, nil, nil, amount, amount, 0)
	if err != nil {
		return nil
	}

	value := &ty.Public2Privacy{
		Tokenname: req.Tokenname,
		Amount:    amount,
		Note:      req.GetNote(),
		Output:    privacyOutput,
	}
	action := &ty.PrivacyAction{
		Ty:    ty.ActionPublic2Privacy,
		Value: &ty.PrivacyAction_Public2Privacy{Public2Privacy: value},
	}
	tx := &types.Transaction{
		Execer:  []byte(ty.PrivacyX),
		Payload: types.Encode(action),
		Nonce:   mock.walletOp.Nonce(),
		To:      address.ExecAddress(ty.PrivacyX),
		ChainID: cfg.GetChainID(),
	}
	txSize := types.Size(tx) + ty.SignatureSize
	realFee := int64((txSize+1023)>>ty.Size1Kshiftlen) * cfg.GetMinTxFeeRate()
	tx.Fee = realFee
	tx.SetExpire(cfg, time.Hour)
	return tx
}
