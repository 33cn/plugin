// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"encoding/hex"

	"github.com/33cn/chain33/system/dapp"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"

	wcom "github.com/33cn/chain33/wallet/common"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	mimcbn256 "github.com/consensys/gnark/crypto/hash/mimc/bn256"
)

const CECBLOCKSIZE = 32

/*
 从secp256k1根私钥创建支票需要的私钥和公钥
 payPrivKey = rootPrivKey *G_X25519 这样很难泄露rootPrivKey

 支票收款key： ReceiveKey= hash(payPrivKey)  --或者*G的X坐标值, 看哪个电路少？
 DH加解密key: encryptPubKey= payPrivKey *G_X25519, 也是很安全的，只是电路里面目前不支持x25519
*/
func newPrivacyKey(rootPrivKey []byte) *mixTy.AccountPrivacyKey {
	ecdh := X25519()
	key := ecdh.PublicKey(rootPrivKey)
	payPrivKey := key.([32]byte)

	//payPrivKey := mimcHashByte([][]byte{rootPrivKey})
	paymentKey := &mixTy.PaymentKeyPair{}
	paymentKey.SpendKey = mixTy.Byte2Str(payPrivKey[:])
	paymentKey.ReceiveKey = mixTy.Byte2Str(mimcHashByte([][]byte{payPrivKey[:]}))

	encryptKeyPair := &mixTy.EncryptKeyPair{}
	pubkey := ecdh.PublicKey(payPrivKey)
	//需要Hex编码，不要使用fr.string, 模范围不同
	encryptKeyPair.PrivKey = hex.EncodeToString(payPrivKey[:])
	pubData := pubkey.([32]byte)
	encryptKeyPair.PubKey = hex.EncodeToString(pubData[:])

	privacy := &mixTy.AccountPrivacyKey{}
	privacy.PaymentKey = paymentKey
	privacy.EncryptKey = encryptKeyPair

	return privacy
}

//CEC加密需要保证明文是秘钥的倍数，如果不是，则需要填充明文，在解密时候把填充物去掉
//填充算法有pkcs5,pkcs7, 比如Pkcs5的思想填充的值为填充的长度，比如加密he,不足8
//则填充为he666666, 解密后直接算最后一个值为6，把解密值的后6个Byte去掉即可
func pKCS5Padding(plainText []byte, blockSize int) []byte {
	if blockSize < CECBLOCKSIZE {
		blockSize = CECBLOCKSIZE
	}
	padding := blockSize - (len(plainText) % blockSize)
	//fmt.Println("pading", "passsize", blockSize, "plaintext", len(plainText), "pad", padding)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	//fmt.Println("padding", padding, "text", common.ToHex(padText[:]))
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

func encryptData(peerPubKey string, data []byte) (*mixTy.DHSecret, error) {
	ecdh := X25519()
	oncePriv, oncePub, err := ecdh.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "x25519 generate key")
	}

	peerPubByte, err := hex.DecodeString(peerPubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "encrypt Decode peer pubkey=%s", peerPubKey)
	}
	password := ecdh.ComputeSecret(oncePriv, peerPubByte)
	encrypt := encryptDataWithPadding(password, data)

	pubData := oncePub.([32]byte)
	return &mixTy.DHSecret{PeerKey: hex.EncodeToString(pubData[:]), Secret: common.ToHex(encrypt)}, nil

}

func decryptDataWithPading(password, data []byte) ([]byte, error) {
	plainData := wcom.CBCDecrypterPrivkey(password, data)
	return pKCS5UnPadding(plainData)
}

func decryptData(selfPrivKey string, peerPubKey string, cryptData []byte) ([]byte, error) {
	ecdh := X25519()
	self, err := hex.DecodeString(selfPrivKey)
	if err != nil {
		return nil, errors.Wrapf(err, "decrypt Decode self prikey=%s", selfPrivKey)
	}
	peer, err := hex.DecodeString(peerPubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "decrypt Decode peer pubkey=%s", peerPubKey)
	}
	password := ecdh.ComputeSecret(self, peer)
	return decryptDataWithPading(password, cryptData)
}

func mimcHashString(params []string) []byte {
	var sum []byte
	for _, k := range params {
		//fmt.Println("input:", k)
		sum = append(sum, mixTy.Str2Byte(k)...)
	}
	hash := mimcHashCalc(sum)
	//fmt.Println("hash=", getFrString(hash))
	return hash

}

func mimcHashByte(params [][]byte) []byte {
	var sum []byte
	for _, k := range params {
		sum = append(sum, k...)
	}
	hash := mimcHashCalc(sum)
	//fmt.Println("hash=", getFrString(hash))
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
	newPrivacy := newPrivacyKey(priv.Bytes())

	password := []byte(policy.getWalletOperate().GetPassword())
	encryptered := encryptDataWithPadding(password, types.Encode(newPrivacy))
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

		policy.processPrivcyTxs(&ReqHashes)
		if txcount < int(MaxTxHashsPerTime) {
			break
		}
	}

	policy.store.setRescanNoteStatus(int32(mixTy.MixWalletRescanStatus_FINISHED))
	return
}

func (policy *mixPolicy) processPrivcyTxs(ReqHashes *types.ReqHashes) {
	//通过txhashs获取对应的txdetail
	txDetails, err := policy.getWalletOperate().GetAPI().GetTransactionByHash(ReqHashes)
	if err != nil {
		bizlog.Error("processPrivcyTx", "GetTransactionByHash error", err)
		return
	}

	for _, tx := range txDetails.Txs {
		if tx.Receipt.Ty != types.ExecOk {
			bizlog.Error("processPrivcyTx wrong tx", "receipt ty", tx.Receipt.Ty, "hash", common.ToHex(tx.Tx.Hash()))
			continue
		}
		set, err := policy.processMixTx(tx.Tx, tx.Height, tx.Index)
		if err != nil {
			bizlog.Error("processPrivcyTx", "processMixTx error", err)
			continue
		}
		policy.store.setKvs(set)
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
		resps.Notes = append(resps.Notes, resp.(*mixTy.WalletIndexResp).Notes...)
	}
	return &resps, nil
}

func (policy *mixPolicy) createRawTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	switch req.ActionTy {
	case mixTy.MixActionDeposit:
		return policy.createDepositTx(req)
	case mixTy.MixActionWithdraw:
		return policy.createWithdrawTx(req)
	case mixTy.MixActionAuth:
		return policy.createAuthTx(req)
	case mixTy.MixActionTransfer:
		return policy.createTransferTx(req)
	default:
		return nil, errors.Wrapf(types.ErrInvalidParam, "action=%d", req.ActionTy)
	}

}
