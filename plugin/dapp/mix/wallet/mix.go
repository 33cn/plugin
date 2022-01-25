// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/consensys/gnark/backend/groth16"

	"github.com/33cn/chain33/system/dapp"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"

	wcom "github.com/33cn/chain33/wallet/common"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

func (p *mixPolicy) getPrivKeyByAddr(addr string) (crypto.PrivKey, error) {
	//获取指定地址在钱包里的账户信息
	Accountstor, err := p.store.GetAccountByAddr(addr)
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
	operater := p.getWalletOperate()
	password := []byte(operater.GetPassword())
	privkey := wcom.CBCDecrypterPrivkey(password, prikeybyte)
	//通过privkey生成一个pubkey然后换算成对应的addr
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
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

func (p *mixPolicy) getAccountPrivacyKey(addr string) (*mixTy.WalletAddrPrivacy, error) {
	if data, _ := p.store.getAccountPrivacy(addr); data != nil {
		privacyInfo := &mixTy.AccountPrivacyKey{}
		password := []byte(p.getWalletOperate().GetPassword())
		decrypted, err := decryptDataWithPading(password, data)
		if err != nil {
			return p.savePrivacyPair(addr)
		}

		//有可能修改了秘钥，如果解密失败，需要重新设置
		err = types.Decode(decrypted, privacyInfo)
		if err != nil {
			return p.savePrivacyPair(addr)
		}

		return &mixTy.WalletAddrPrivacy{Privacy: privacyInfo, Addr: addr}, nil
	}

	return p.savePrivacyPair(addr)

}

func (p *mixPolicy) savePrivacyPair(addr string) (*mixTy.WalletAddrPrivacy, error) {
	priv, err := p.getPrivKeyByAddr(addr)
	if err != nil {
		return nil, errors.Wrapf(err, "savePrivacyPair addr=%s", addr)
	}

	bizlog.Info("savePrivacyPair", "pri", common.ToHex(priv.Bytes()), "addr", addr)
	newPrivacy := newPrivacyKey(priv.Bytes())

	password := []byte(p.getWalletOperate().GetPassword())
	encryptered := encryptDataWithPadding(password, types.Encode(newPrivacy))
	//save the privacy created to wallet db
	p.store.setAccountPrivacy(addr, encryptered)
	return &mixTy.WalletAddrPrivacy{Privacy: newPrivacy, Addr: addr}, nil
}

//查询钱包里面所有的地址对应的PrivacyKeys
func (p *mixPolicy) getWalletPrivacyKeys() ([]*mixTy.WalletAddrPrivacy, error) {
	//通过Account前缀查找获取钱包中的所有账户信息
	WalletAccStores, err := p.store.GetAccountByPrefix("Account")
	if err != nil || len(WalletAccStores) == 0 {
		bizlog.Info("getPrivacyKeyPairs", "store getAccountByPrefix error", err)
		return nil, err
	}

	var infoPriRes []*mixTy.WalletAddrPrivacy
	for _, AccStore := range WalletAccStores {
		if len(AccStore.Addr) != 0 {
			if privacyInfo, err := p.getAccountPrivacyKey(AccStore.Addr); err == nil {
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

func (p *mixPolicy) getRescanStatus() string {
	status := p.store.getRescanNoteStatus()
	return mixTy.MixWalletRescanStatus(status).String()
}

func (p *mixPolicy) tryRescanNotes() error {
	//未使能，直接使能
	if !p.store.getPrivacyEnable() {
		//p.store.enablePrivacy()
		return errors.Wrap(types.ErrNotAllow, "privacy need enable firstly")
	}
	operater := p.getWalletOperate()
	if operater.IsWalletLocked() {
		return types.ErrWalletIsLocked
	}
	status := p.store.getRescanNoteStatus()
	if status == int32(mixTy.MixWalletRescanStatus_SCANNING) {
		return errors.Wrap(types.ErrNotAllow, "mix wallet is scanning")
	}

	p.store.setRescanNoteStatus(int32(mixTy.MixWalletRescanStatus_SCANNING))

	go p.rescanNotes()

	return nil
}

//从localdb中把Mix合约的交易按升序都获取出来依次处理
func (p *mixPolicy) rescanNotes() {
	var txInfo mixTy.LocalMixTx
	i := 0
	operater := p.getWalletOperate()
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
		reqInfo.Count = int32(maxTxHashsPerTime)
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

		p.processPrivcyTxs(&ReqHashes)
		if txcount < int(maxTxHashsPerTime) {
			break
		}
	}

	p.store.setRescanNoteStatus(int32(mixTy.MixWalletRescanStatus_FINISHED))
	return
}

func (p *mixPolicy) processPrivcyTxs(ReqHashes *types.ReqHashes) {
	//通过txhashs获取对应的txdetail
	txDetails, err := p.getWalletOperate().GetAPI().GetTransactionByHash(ReqHashes)
	if err != nil {
		bizlog.Error("processPrivcyTx", "GetTransactionByHash error", err)
		return
	}

	for _, tx := range txDetails.Txs {
		if tx.Receipt.Ty != types.ExecOk {
			bizlog.Error("processPrivcyTx wrong tx", "receipt ty", tx.Receipt.Ty, "hash", common.ToHex(tx.Tx.Hash()))
			continue
		}
		set, err := p.processMixTx(tx.Tx, tx.Height, tx.Index)
		if err != nil {
			bizlog.Error("processPrivcyTx", "processMixTx error", err)
			continue
		}
		p.store.setKvs(set)
	}
}

func (p *mixPolicy) enablePrivacy(addrs []string) (*mixTy.ReqEnablePrivacyRst, error) {
	if 0 == len(addrs) {
		WalletAccStores, err := p.store.GetAccountByPrefix("Account")
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
		_, err := p.getAccountPrivacyKey(addr)
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
	p.store.enablePrivacy()
	return &rep, nil
}

func (p *mixPolicy) showAccountNoteInfo(req *mixTy.WalletMixIndexReq) (*mixTy.WalletNoteResp, error) {
	resp, err := p.listMixInfos(req)
	if err != nil {
		return nil, err
	}
	return resp.(*mixTy.WalletNoteResp), nil
}

func (p *mixPolicy) createRawTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	switch req.ActionTy {
	case mixTy.MixActionDeposit:
		return p.createDepositTx(req)
	case mixTy.MixActionWithdraw:
		return p.createWithdrawTx(req)
	case mixTy.MixActionAuth:
		return p.createAuthTx(req)
	case mixTy.MixActionTransfer:
		return p.createTransferTx(req)
	default:
		return nil, errors.Wrapf(types.ErrInvalidParam, "action=%d", req.ActionTy)
	}

}

func (p *mixPolicy) createZkKeyFile(req *mixTy.CreateZkKeyFileReq) (*types.ReplyString, error) {

	ccs, err := getCircuit(mixTy.VerifyType(req.Ty))
	if err != nil {
		return nil, err
	}

	pkName, vkName, err := getCircuitKeyFileName(mixTy.VerifyType(req.Ty))
	if err != nil {
		return nil, err
	}

	var bufPk, bufVk bytes.Buffer
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return nil, errors.Wrapf(err, "setup")
	}

	pk.WriteTo(&bufPk)
	vk.WriteTo(&bufVk)

	file := filepath.Join(req.SavePath, pkName)
	fPk, err := os.Create(file)
	if err != nil {
		return nil, errors.Wrapf(err, "create file")
	}
	defer fPk.Close()
	fPk.WriteString(hex.EncodeToString(bufPk.Bytes()))

	file = filepath.Join(req.SavePath, vkName)
	fVk, err := os.Create(file)
	if err != nil {
		return nil, errors.Wrapf(err, "create file")
	}
	defer fVk.Close()
	fVk.WriteString(hex.EncodeToString(bufVk.Bytes()))

	var res types.ReplyString
	res.Data = hex.EncodeToString(bufVk.Bytes())

	return &res, nil
}
