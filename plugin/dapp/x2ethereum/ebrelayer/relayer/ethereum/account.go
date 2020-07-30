package ethereum

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	chain33Common "github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	chain33Types "github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	x2ethTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58/base58"
	"github.com/pborman/uuid"
	"golang.org/x/crypto/ripemd160"
)

var (
	chain33AccountKey = []byte("Chain33Account4EthRelayer")
	start             = int(1)
)

//Key ...
type Key struct {
	ID uuid.UUID // Version 4 "random" for unique id not derived from key data
	// to simplify lookups we also store the address
	Address common.Address
	// we only store privkey as pubkey/address can be derived from it
	// privkey in this struct is always in plaintext
	PrivateKey *ecdsa.PrivateKey
}

//NewAccount ...
func (ethRelayer *Relayer4Ethereum) NewAccount(passphrase string) (privateKeystr, addr string, err error) {
	_, privateKeystr, addr, err = newKeyAndStore(ethRelayer.db, crand.Reader, passphrase)
	if err != nil {
		return "", "", err
	}
	return
}

//GetAccount ...
func (ethRelayer *Relayer4Ethereum) GetAccount(passphrase string) (privateKey, addr string, err error) {
	accountInfo, err := ethRelayer.db.Get(chain33AccountKey)
	if nil != err {
		return "", "", err
	}
	Chain33Account := &x2ethTypes.Account4Relayer{}
	if err := chain33Types.Decode(accountInfo, Chain33Account); nil != err {
		return "", "", err
	}
	decryptered := wcom.CBCDecrypterPrivkey([]byte(passphrase), Chain33Account.Privkey)
	privateKey = chain33Common.ToHex(decryptered)
	addr = Chain33Account.Addr
	return
}

//GetValidatorAddr ...
func (ethRelayer *Relayer4Ethereum) GetValidatorAddr() (validators x2ethTypes.ValidatorAddr4EthRelayer, err error) {
	var chain33AccountAddr string
	accountInfo, err := ethRelayer.db.Get(chain33AccountKey)
	if nil == err {
		ethAccount := &x2ethTypes.Account4Relayer{}
		if err := chain33Types.Decode(accountInfo, ethAccount); nil == err {
			chain33AccountAddr = ethAccount.Addr
		}
	}

	if 0 == len(chain33AccountAddr) {
		return x2ethTypes.ValidatorAddr4EthRelayer{}, x2ethTypes.ErrNoValidatorConfigured
	}

	validators = x2ethTypes.ValidatorAddr4EthRelayer{
		Chain33Validator: chain33AccountAddr,
	}
	return
}

//RestorePrivateKeys ...
func (ethRelayer *Relayer4Ethereum) RestorePrivateKeys(passPhase string) (err error) {
	accountInfo, err := ethRelayer.db.Get(chain33AccountKey)
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
			ethRelayer.rwLock.Lock()
			ethRelayer.privateKey4Chain33 = priKey
			ethRelayer.rwLock.Unlock()
		}
	}

	ethRelayer.rwLock.RLock()
	if nil != ethRelayer.privateKey4Chain33 {
		ethRelayer.unlockchan <- start
	}
	ethRelayer.rwLock.RUnlock()

	return nil
}

//StoreAccountWithNewPassphase ...
func (ethRelayer *Relayer4Ethereum) StoreAccountWithNewPassphase(newPassphrase, oldPassphrase string) error {
	accountInfo, err := ethRelayer.db.Get(chain33AccountKey)
	if nil != err {
		relayerLog.Info("StoreAccountWithNewPassphase", "pls check account is created already, err", err)
		return err
	}
	Chain33Account := &x2ethTypes.Account4Relayer{}
	if err := chain33Types.Decode(accountInfo, Chain33Account); nil != err {
		return err
	}
	decryptered := wcom.CBCDecrypterPrivkey([]byte(oldPassphrase), Chain33Account.Privkey)
	encryptered := wcom.CBCEncrypterPrivkey([]byte(newPassphrase), decryptered)
	Chain33Account.Privkey = encryptered
	encodedInfo := chain33Types.Encode(Chain33Account)
	return ethRelayer.db.SetSync(chain33AccountKey, encodedInfo)
}

//ImportChain33PrivateKey ...
func (ethRelayer *Relayer4Ethereum) ImportChain33PrivateKey(passphrase, privateKeyStr string) error {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(privateKeyStr)
	if nil != err {
		return err
	}
	priKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return err
	}

	ethRelayer.rwLock.Lock()
	ethRelayer.privateKey4Chain33 = priKey
	ethRelayer.rwLock.Unlock()
	ethRelayer.unlockchan <- start
	addr, err := pubKeyToAddress4Bty(priKey.PubKey().Bytes())
	if nil != err {
		return err
	}

	encryptered := wcom.CBCEncrypterPrivkey([]byte(passphrase), privateKeySli)
	account := &x2ethTypes.Account4Relayer{
		Privkey: encryptered,
		Addr:    addr,
	}
	encodedInfo := chain33Types.Encode(account)
	return ethRelayer.db.SetSync(chain33AccountKey, encodedInfo)
}

//checksum: first four bytes of double-SHA256.
func checksum(input []byte) (cksum [4]byte) {
	h := sha256.New()
	_, err := h.Write(input)
	if err != nil {
		return
	}
	intermediateHash := h.Sum(nil)
	h.Reset()
	_, err = h.Write(intermediateHash)
	if err != nil {
		return
	}
	finalHash := h.Sum(nil)
	copy(cksum[:], finalHash[:])
	return
}

func pubKeyToAddress4Bty(pub []byte) (addr string, err error) {
	if len(pub) != 33 && len(pub) != 65 { //压缩格式 与 非压缩格式
		return "", fmt.Errorf("invalid public key byte")
	}

	sha256h := sha256.New()
	_, err = sha256h.Write(pub)
	if err != nil {
		return "", err
	}
	//160hash
	ripemd160h := ripemd160.New()
	_, err = ripemd160h.Write(sha256h.Sum([]byte("")))
	if err != nil {
		return "", err
	}
	//添加版本号
	hash160res := append([]byte{0}, ripemd160h.Sum([]byte(""))...)

	//添加校验码
	cksum := checksum(hash160res)
	address := append(hash160res, cksum[:]...)

	//地址进行base58编码
	addr = base58.Encode(address)
	return
}

func newKeyAndStore(db dbm.DB, rand io.Reader, passphrase string) (privateKey *ecdsa.PrivateKey, privateKeyStr, addr string, err error) {
	key, err := newKey(rand)
	if err != nil {
		return nil, "", "", err
	}
	privateKey = key.PrivateKey
	privateKeyBytes := math.PaddedBigBytes(key.PrivateKey.D, 32)
	Encryptered := wcom.CBCEncrypterPrivkey([]byte(passphrase), privateKeyBytes)
	ethAccount := &x2ethTypes.Account4Relayer{
		Privkey: Encryptered,
		Addr:    key.Address.Hex(),
	}
	_ = db

	privateKeyStr = chain33Common.ToHex(privateKeyBytes)
	addr = ethAccount.Addr
	return
}

func newKey(rand io.Reader) (*Key, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand)
	if err != nil {
		return nil, err
	}
	return newKeyFromECDSA(privateKeyECDSA), nil
}

func newKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *Key {
	id := uuid.NewRandom()
	key := &Key{
		ID:         id,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key
}
