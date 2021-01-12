// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/33cn/chain33/system/dapp"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"

	wcom "github.com/33cn/chain33/wallet/common"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	mimcbn256 "github.com/consensys/gnark/crypto/hash/mimc/bn256"
	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
)

// newPrivacyWithPrivKey create privacy from private key
//payment, payPrivKey=hash(privkey), payPubkey=hash(payPrivKey)
//DH crypt key, prikey=payPrikey, pubKey=payPrikey*G
func newPrivacyWithPrivKey(privKey []byte) (*mixTy.AccountPrivacyKey, error) {
	payPrivacyKey := MimcHashByte([][]byte{privKey})
	paymentKey := &mixTy.PaymentKeyPair{}
	paymentKey.SpendKey = getFrString(payPrivacyKey)
	paymentKey.PayKey = getFrString(MimcHashByte([][]byte{payPrivacyKey}))

	shareSecretKey := &mixTy.ShareSecretKeyPair{}
	ecdh := NewCurveBn256ECDH()
	shareSecretKey.PrivKey, shareSecretKey.ReceivingPk = ecdh.GenerateKey(payPrivacyKey)

	privacy := &mixTy.AccountPrivacyKey{}
	privacy.PaymentKey = paymentKey
	privacy.ShareSecretKey = shareSecretKey

	return privacy, nil
}

//CEC加密需要保证明文是秘钥的倍数，如果不是，则需要填充明文，在解密时候把填充物去掉
//填充算法有pkcs5,pkcs7, 比如Pkcs5的思想填充的值为填充的长度，比如加密he,不足8
//则填充为he666666, 解密后直接算最后一个值为6，把解密值的后6个Byte去掉即可
func pKCS5Padding(plainText []byte, blockSize int) []byte {
	if blockSize < 32 {
		blockSize = 32
	}
	padding := blockSize - (len(plainText) % blockSize)
	fmt.Println("pading", "passsize", blockSize, "plaintext", len(plainText), "pad", padding)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	fmt.Println("padding", padding, "text", common.ToHex(padText[:]))
	newText := append(plainText, padText...)
	return newText
}

func pKCS5UnPadding(plainText []byte) ([]byte, error) {
	length := len(plainText)
	number := int(plainText[length-1])
	if number > length {
		return nil, types.ErrInvalidParam
	}
	return plainText[:length-number], nil
}

func encryptDataWithPadding(password, data []byte) []byte {
	paddingText := pKCS5Padding(data, len(password))
	return wcom.CBCEncrypterPrivkey(password, paddingText)
}

func encryptData(receiverPubKey *mixTy.PubKey, data []byte) (*mixTy.PubKey, []byte, error) {
	ecdh := NewCurveBn256ECDH()
	//generate ephemeral priv/pub key
	ephPriv, ephPub := ecdh.GenerateKey(nil)
	password, _ := ecdh.GenerateSharedSecret(ephPriv, receiverPubKey)

	return ephPub, encryptDataWithPadding(password, data), nil

}

func decryptDataWithPading(password, data []byte) ([]byte, error) {
	plainData := wcom.CBCDecrypterPrivkey(password, data)
	return pKCS5UnPadding(plainData)
}

func decryptData(selfPrivKey *mixTy.PrivKey, oppositePubKey *mixTy.PubKey, cryptData []byte) ([]byte, error) {
	ecdh := NewCurveBn256ECDH()
	password, _ := ecdh.GenerateSharedSecret(selfPrivKey, oppositePubKey)

	return decryptDataWithPading(password, cryptData)
}

func getByte(v string) []byte {
	var fr fr_bn256.Element
	fr.SetString(v)
	return fr.Bytes()
}
func getFrString(v []byte) string {
	var f fr_bn256.Element
	f.SetBytes(v)
	return f.String()
}

func MimcHashString(params []string) []byte {
	var sum []byte
	for _, k := range params {
		fmt.Println("input:", k)
		sum = append(sum, getByte(k)...)
	}
	hash := mimcHashCalc(sum)
	fmt.Println("hash=", getFrString(hash))
	return hash

}

func MimcHashByte(params [][]byte) []byte {
	var sum []byte
	for _, k := range params {
		sum = append(sum, k...)
	}
	hash := mimcHashCalc(sum)
	fmt.Println("hash=", getFrString(hash))
	return hash

}

func mimcHashCalc(sum []byte) []byte {
	return mimcbn256.Sum("seed", sum)
}

func (policy *mixPolicy) getPrivKeyByAddr(addr string) (crypto.PrivKey, error) {
	//获取指定地址在钱包里的账户信息
	Accountstor, err := policy.store.GetAccountByAddr(addr)
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
	operater := policy.getWalletOperate()
	password := []byte(operater.GetPassword())
	privkey := wcom.CBCDecrypterPrivkey(password, prikeybyte)
	//通过privkey生成一个pubkey然后换算成对应的addr
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
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

func (policy *mixPolicy) getAccountPrivacyKey(addr string) (*mixTy.WalletAddrPrivacy, error) {
	if data, _ := policy.store.getAccountPrivacy(addr); data != nil {
		privacyInfo := &mixTy.AccountPrivacyKey{}
		password := []byte(policy.getWalletOperate().GetPassword())
		decrypted, err := decryptDataWithPading(password, data)
		if err != nil {
			return policy.savePrivacyPair(addr)
		}

		//有可能修改了秘钥，如果解密失败，需要重新设置
		err = types.Decode(decrypted, privacyInfo)
		if err != nil {
			return policy.savePrivacyPair(addr)
		}

		return &mixTy.WalletAddrPrivacy{Privacy: privacyInfo, Addr: addr}, nil
	}

	return policy.savePrivacyPair(addr)

}

func (policy *mixPolicy) savePrivacyPair(addr string) (*mixTy.WalletAddrPrivacy, error) {
	priv, err := policy.getPrivKeyByAddr(addr)
	if err != nil {
		return nil, errors.Wrapf(err, "savePrivacyPair addr=%s", addr)
	}

	bizlog.Info("savePrivacyPair", "pri", common.ToHex(priv.Bytes()), "addr", addr)
	newPrivacy, err := newPrivacyWithPrivKey(priv.Bytes())
	if err != nil {
		return nil, err
	}

	password := []byte(policy.getWalletOperate().GetPassword())
	bizlog.Info("savePrivacyPair", "newprivacy", newPrivacy.PaymentKey.PayKey, "password", common.ToHex(password))
	encryptered := encryptDataWithPadding(password, types.Encode(newPrivacy))
	bizlog.Info("savePrivacyPair--2")
	//save the privacy created to wallet db
	policy.store.setAccountPrivacy(addr, encryptered)
	return &mixTy.WalletAddrPrivacy{Privacy: newPrivacy, Addr: addr}, nil
}

//查询钱包里面所有的地址对应的PrivacyKeys
func (policy *mixPolicy) getWalletPrivacyKeys() ([]*mixTy.WalletAddrPrivacy, error) {
	//通过Account前缀查找获取钱包中的所有账户信息
	WalletAccStores, err := policy.store.GetAccountByPrefix("Account")
	if err != nil || len(WalletAccStores) == 0 {
		bizlog.Info("getPrivacyKeyPairs", "store getAccountByPrefix error", err)
		return nil, err
	}

	var infoPriRes []*mixTy.WalletAddrPrivacy
	for _, AccStore := range WalletAccStores {
		if len(AccStore.Addr) != 0 {
			if privacyInfo, err := policy.getAccountPrivacyKey(AccStore.Addr); err == nil {
				infoPriRes = append(infoPriRes, privacyInfo)
			}
		}
	}

	if 0 == len(infoPriRes) {
		bizlog.Error("mixCoin getPrivacyKeyPairs null")
		return nil, nil
	}

	return infoPriRes, nil

}

func (policy *mixPolicy) getRescanStatus() string {
	status := policy.store.getRescanNoteStatus()
	return mixTy.MixWalletRescanStatus(status).String()
}

func (policy *mixPolicy) tryRescanNotes() error {
	//未使能，直接使能
	if !policy.store.getPrivacyEnable() {
		//policy.store.enablePrivacy()
		return errors.Wrap(types.ErrNotAllow, "privacy need enable firstly")
	}
	operater := policy.getWalletOperate()
	if operater.IsWalletLocked() {
		return types.ErrWalletIsLocked
	}
	status := policy.store.getRescanNoteStatus()
	if status == int32(mixTy.MixWalletRescanStatus_SCANNING) {
		return errors.Wrap(types.ErrNotAllow, "mix wallet is scanning")
	}

	policy.store.setRescanNoteStatus(int32(mixTy.MixWalletRescanStatus_SCANNING))

	go policy.rescanNotes()

	return nil
}

//从localdb中把Mix合约的交易按升序都获取出来依次处理
func (policy *mixPolicy) rescanNotes() {
	var txInfo mixTy.LocalMixTx
	i := 0
	operater := policy.getWalletOperate()
	for {
		select {
		case <-operater.GetWalletDone():
			return
		default:
		}

		//首先从execs模块获取地址对应的所有UTXOs,
		// 1 先获取隐私合约地址相关交易
		var reqInfo mixTy.MixTxListReq
		reqInfo.Direction = 0
		reqInfo.Count = int32(MaxTxHashsPerTime)
		if i == 0 {
			reqInfo.Height = -1

		} else {
			reqInfo.Height = txInfo.GetHeight()
			reqInfo.TxIndex = dapp.HeightIndexStr(txInfo.GetHeight(), txInfo.GetIndex())
		}
		i++
		//请求交易信息
		msg, err := operater.GetAPI().Query(mixTy.MixX, "ListMixTxs", &reqInfo)
		if err != nil {
			bizlog.Error("ListMixTxs", "error", err, "height", reqInfo.Height, "index", reqInfo.TxIndex)
			break
		}
		mixTxInfos := msg.(*mixTy.MixTxListResp)
		if mixTxInfos == nil {
			bizlog.Info("rescanNotes mix privacy ReqTxInfosByAddr ReplyTxInfos is nil")
			break
		}
		txcount := len(mixTxInfos.Txs)

		var ReqHashes types.ReqHashes
		ReqHashes.Hashes = make([][]byte, len(mixTxInfos.Txs))
		for index, tx := range mixTxInfos.Txs {
			hash, err := common.FromHex(tx.Hash)
			if err != nil {
				bizlog.Error("rescanNotes mix decode hash", "hash", tx.Hash)
			}
			ReqHashes.Hashes[index] = hash
		}

		if txcount > 0 {
			txInfo.Hash = mixTxInfos.Txs[txcount-1].GetHash()
			txInfo.Height = mixTxInfos.Txs[txcount-1].GetHeight()
			txInfo.Index = mixTxInfos.Txs[txcount-1].GetIndex()
		}

		policy.getPrivacyTxDetailByHashs(&ReqHashes)
		if txcount < int(MaxTxHashsPerTime) {
			break
		}
	}

	policy.store.setRescanNoteStatus(int32(mixTy.MixWalletRescanStatus_FINISHED))
	return
}

func (policy *mixPolicy) getPrivacyTxDetailByHashs(ReqHashes *types.ReqHashes) {
	//通过txhashs获取对应的txdetail
	txDetails, err := policy.getWalletOperate().GetAPI().GetTransactionByHash(ReqHashes)
	if err != nil {
		bizlog.Error("getPrivacyTxDetailByHashs", "GetTransactionByHash error", err)
		return
	}

	for _, tx := range txDetails.Txs {
		policy.processMixTx(tx.Tx, tx.Height, tx.Index)
	}
}

func (policy *mixPolicy) enablePrivacy(addrs []string) (*mixTy.ReqEnablePrivacyRst, error) {
	if 0 == len(addrs) {
		WalletAccStores, err := policy.store.GetAccountByPrefix("Account")
		if err != nil || len(WalletAccStores) == 0 {
			bizlog.Info("enablePrivacy", "GetAccountByPrefix:err", err)
			return nil, types.ErrNotFound
		}
		for _, WalletAccStore := range WalletAccStores {
			bizlog.Info("enablePrivacy", "addr", WalletAccStore.Addr)
			addrs = append(addrs, WalletAccStore.Addr)
		}
	} else {
		addrs = append(addrs, addrs...)
		bizlog.Info("enablePrivacy", "addrs", addrs)
	}

	var rep mixTy.ReqEnablePrivacyRst
	for _, addr := range addrs {
		str := ""
		isOK := true
		_, err := policy.getAccountPrivacyKey(addr)
		if err != nil {
			isOK = false
			str = err.Error()
		}

		priAddrResult := &mixTy.PrivacyAddrResult{
			Addr: addr,
			IsOK: isOK,
			Msg:  str,
		}

		rep.Results = append(rep.Results, priAddrResult)
	}
	policy.store.enablePrivacy()
	return &rep, nil
}

func (policy *mixPolicy) showAccountNoteInfo(addrs []string) (*mixTy.WalletIndexResp, error) {
	var resps mixTy.WalletIndexResp
	for _, addr := range addrs {
		var req mixTy.WalletMixIndexReq
		req.Account = addr
		resp, err := policy.listMixInfos(&req)
		if err != nil {
			return nil, err
		}
		resps.Datas = append(resps.Datas, resp.(*mixTy.WalletIndexResp).Datas...)
	}
	return &resps, nil
}

//对secretData 编码为string,同时增加随机值
func encodeSecretData(secret *mixTy.SecretData) (*mixTy.EncodedSecretData, error) {
	if secret == nil {
		return nil, errors.Wrap(types.ErrInvalidParam, "para is nil")
	}
	if len(secret.PaymentPubKey) <= 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "spendPubKey is nil")
	}
	var val big.Int
	ret, succ := val.SetString(secret.Amount, 10)
	if !succ {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong amount = %s", secret.Amount)
	}
	if ret.Sign() <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount = %s, need bigger than 0", secret.Amount)
	}

	//获取随机值
	var fr fr_bn256.Element
	fr.SetRandom()
	secret.NoteRandom = fr.String()
	code := types.Encode(secret)
	var resp mixTy.EncodedSecretData

	resp.Encoded = common.ToHex(code)
	resp.RawData = secret

	return &resp, nil

}

//产生随机秘钥和receivingPk对data DH加密，返回随机秘钥的公钥
func encryptSecretData(req *mixTy.EncryptSecretData) (*mixTy.DHSecret, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode secret")
	}
	epk, crypt, err := encryptData(req.ReceivingPk, secret)
	if err != nil {
		return nil, errors.Wrap(err, "encrypt")
	}
	return &mixTy.DHSecret{Epk: epk, Secret: common.ToHex(crypt)}, nil
}

func decryptSecretData(req *mixTy.DecryptSecretData) (*mixTy.SecretData, error) {
	secret, err := common.FromHex(req.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "decode req.secret")
	}
	decrypt, err := decryptData(req.ReceivingPriKey, req.Epk, secret)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt secret")
	}
	var raw mixTy.SecretData
	err = types.Decode(decrypt, &raw)
	if err != nil {
		return nil, errors.Wrap(mixTy.ErrDecryptDataFail, "decode decrypt.secret")
	}
	return &raw, nil
}
