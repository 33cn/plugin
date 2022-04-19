package chain33

import (
	"errors"
	"fmt"

	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	chain33Types "github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	x2ethTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	btcec_secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	chain33AccountKey = []byte("Chain33Account4Relayer")
	start             = int(1)
)

//GetAccount ...
func (chain33Relayer *Relayer4Chain33) GetAccount(passphrase string) (privateKey, addr string, err error) {
	accountInfo, err := chain33Relayer.db.Get(chain33AccountKey)
	if nil != err {
		return "", "", err
	}
	ethAccount := &x2ethTypes.Account4Relayer{}
	if err := chain33Types.Decode(accountInfo, ethAccount); nil != err {
		return "", "", err
	}
	decryptered := wcom.CBCDecrypterPrivkey([]byte(passphrase), ethAccount.Privkey)
	privateKey = chain33Common.ToHex(decryptered)
	addr = ethAccount.Addr
	return
}

//GetAccountAddr ...
func (chain33Relayer *Relayer4Chain33) GetAccountAddr() (addr string, err error) {
	accountInfo, err := chain33Relayer.db.Get(chain33AccountKey)
	if nil != err {
		relayerLog.Info("GetValidatorAddr", "Failed to get account from db due to:", err.Error())
		return "", err
	}
	ethAccount := &x2ethTypes.Account4Relayer{}
	if err := chain33Types.Decode(accountInfo, ethAccount); nil != err {
		relayerLog.Info("GetValidatorAddr", "Failed to decode due to:", err.Error())
		return "", err
	}
	addr = ethAccount.Addr
	return
}

func (chain33Relayer *Relayer4Chain33) ImportPrivateKey(passphrase, privateKeyStr string) error {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(privateKeyStr)
	if nil != err {
		return err
	}
	priKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return err
	}

	chain33Relayer.rwLock.Lock()
	chain33Relayer.privateKey4Chain33 = priKey
	temp, _ := btcec_secp256k1.PrivKeyFromBytes(btcec_secp256k1.S256(), priKey.Bytes())
	chain33Relayer.privateKey4Chain33_ecdsa = temp.ToECDSA()
	chain33Relayer.rwLock.Unlock()
	chain33Relayer.unlockChan <- start
	addr := address.PubKeyToAddr(address.DefaultID, priKey.PubKey().Bytes())

	encryptered := wcom.CBCEncrypterPrivkey([]byte(passphrase), privateKeySli)
	account := &x2ethTypes.Account4Relayer{
		Privkey: encryptered,
		Addr:    addr,
	}
	encodedInfo := chain33Types.Encode(account)
	return chain33Relayer.db.SetSync(chain33AccountKey, encodedInfo)
}

//StoreAccountWithNewPassphase ...
func (chain33Relayer *Relayer4Chain33) StoreAccountWithNewPassphase(newPassphrase, oldPassphrase string) error {
	accountInfo, err := chain33Relayer.db.Get(chain33AccountKey)
	if nil != err {
		relayerLog.Info("StoreAccountWithNewPassphase", "pls check account is created already, err", err)
		return err
	}
	ethAccount := &x2ethTypes.Account4Relayer{}
	if err := chain33Types.Decode(accountInfo, ethAccount); nil != err {
		return err
	}
	decryptered := wcom.CBCDecrypterPrivkey([]byte(oldPassphrase), ethAccount.Privkey)
	encryptered := wcom.CBCEncrypterPrivkey([]byte(newPassphrase), decryptered)
	ethAccount.Privkey = encryptered
	encodedInfo := chain33Types.Encode(ethAccount)
	return chain33Relayer.db.SetSync(chain33AccountKey, encodedInfo)
}

//RestorePrivateKeys ...
func (chain33Relayer *Relayer4Chain33) RestorePrivateKeys(passPhase string) (err error) {
	accountInfo, err := chain33Relayer.db.Get(chain33AccountKey)
	if nil == err {
		Chain33Account := &x2ethTypes.Account4Relayer{}
		if err := chain33Types.Decode(accountInfo, Chain33Account); nil == err {
			decryptered := wcom.CBCDecrypterPrivkey([]byte(passPhase), Chain33Account.Privkey)
			var driver secp256k1.Driver
			priKey, err := driver.PrivKeyFromBytes(decryptered)
			if nil != err {
				errInfo := fmt.Sprintf("Failed to PrivKeyFromBytes due to:%s", err.Error())
				relayerLog.Info("RestorePrivateKeys", "Failed to PrivKeyFromBytes:", err.Error())
				return errors.New(errInfo)
			}
			chain33Relayer.rwLock.Lock()
			chain33Relayer.privateKey4Chain33 = priKey
			chain33Relayer.privateKey4Chain33_ecdsa, err = crypto.ToECDSA(priKey.Bytes())
			if nil != err {
				return err
			}
			chain33Relayer.rwLock.Unlock()
		}
	}

	chain33Relayer.rwLock.RLock()
	if nil != chain33Relayer.privateKey4Chain33 {
		chain33Relayer.unlockChan <- start
	}
	chain33Relayer.rwLock.RUnlock()

	return nil
}
