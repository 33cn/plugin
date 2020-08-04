package chain33

import (
	chain33Common "github.com/33cn/chain33/common"
	"github.com/ethereum/go-ethereum/crypto"

	//dbm "github.com/33cn/chain33/common/db"
	chain33Types "github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	x2ethTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
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

//ImportPrivateKey ...
func (chain33Relayer *Relayer4Chain33) ImportPrivateKey(passphrase, privateKeyStr string) (addr string, err error) {
	privateKeySlice, err := chain33Common.FromHex(privateKeyStr)
	if nil != err {
		return "", err
	}
	privateKey, err := crypto.ToECDSA(privateKeySlice)
	if nil != err {
		return "", err
	}

	ethSender := crypto.PubkeyToAddress(privateKey.PublicKey)
	chain33Relayer.privateKey4Ethereum = privateKey
	chain33Relayer.ethSender = ethSender
	chain33Relayer.unlock <- start

	addr = chain33Common.ToHex(ethSender.Bytes())
	encryptered := wcom.CBCEncrypterPrivkey([]byte(passphrase), privateKeySlice)
	ethAccount := &x2ethTypes.Account4Relayer{
		Privkey: encryptered,
		Addr:    addr,
	}
	encodedInfo := chain33Types.Encode(ethAccount)
	err = chain33Relayer.db.SetSync(chain33AccountKey, encodedInfo)

	return
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
func (chain33Relayer *Relayer4Chain33) RestorePrivateKeys(passphrase string) error {
	accountInfo, err := chain33Relayer.db.Get(chain33AccountKey)
	if nil != err {
		relayerLog.Info("No private key saved for Relayer4Chain33")
		return nil
	}
	ethAccount := &x2ethTypes.Account4Relayer{}
	if err := chain33Types.Decode(accountInfo, ethAccount); nil != err {
		relayerLog.Info("RestorePrivateKeys", "Failed to decode due to:", err.Error())
		return err
	}
	decryptered := wcom.CBCDecrypterPrivkey([]byte(passphrase), ethAccount.Privkey)
	privateKey, err := crypto.ToECDSA(decryptered)
	if nil != err {
		relayerLog.Info("RestorePrivateKeys", "Failed to ToECDSA:", err.Error())
		return err
	}

	chain33Relayer.rwLock.Lock()
	chain33Relayer.privateKey4Ethereum = privateKey
	chain33Relayer.ethSender = crypto.PubkeyToAddress(privateKey.PublicKey)
	chain33Relayer.rwLock.Unlock()
	chain33Relayer.unlock <- start
	return nil
}

//func (chain33Relayer *Relayer4Chain33) UpdatePrivateKey(Passphrase, privateKey string) error {
//	return nil
//}
