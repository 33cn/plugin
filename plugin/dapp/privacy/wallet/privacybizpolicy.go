// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of policy source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"sync"
	"sync/atomic"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	privacytypes "github.com/33cn/plugin/plugin/dapp/privacy/types"
)

var (
	bizlog = log15.New("module", "wallet.privacy")
	// MaxTxHashsPerTime 单词处理的最大哈希书
	MaxTxHashsPerTime int64 = 100
	// maxTxNumPerBlock 单个区块最大数
	maxTxNumPerBlock int64 = types.MaxTxsPerBlock
)

func init() {
	wcom.RegisterPolicy(privacytypes.PrivacyX, New())
}

// New 创建一盒钱包业务策略
func New() wcom.WalletBizPolicy {
	return &privacyPolicy{
		mtx:            &sync.Mutex{},
		rescanwg:       &sync.WaitGroup{},
		rescanUTXOflag: privacytypes.UtxoFlagNoScan,
	}
}

type privacyPolicy struct {
	mtx            *sync.Mutex
	store          *privacyStore
	walletOperate  wcom.WalletOperate
	rescanwg       *sync.WaitGroup
	rescanUTXOflag int32
}

func (policy *privacyPolicy) setWalletOperate(walletBiz wcom.WalletOperate) {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	policy.walletOperate = walletBiz
}

func (policy *privacyPolicy) getWalletOperate() wcom.WalletOperate {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	return policy.walletOperate
}

// Init 初始化处理
func (policy *privacyPolicy) Init(walletOperate wcom.WalletOperate, sub []byte) {
	policy.setWalletOperate(walletOperate)
	policy.store = newStore(walletOperate.GetDBStore())
	// 启动定时检查超期FTXO的协程
	walletOperate.GetWaitGroup().Add(1)
	go policy.checkWalletStoreData()
}

// OnCreateNewAccount 在账号创建时做一些处理
func (policy *privacyPolicy) OnCreateNewAccount(acc *types.Account) {
	wg := policy.getWalletOperate().GetWaitGroup()
	wg.Add(1)
	go policy.rescanReqTxDetailByAddr(acc.Addr, wg)
}

// OnImportPrivateKey 在私钥导入时做一些处理
func (policy *privacyPolicy) OnImportPrivateKey(acc *types.Account) {
	wg := policy.getWalletOperate().GetWaitGroup()
	wg.Add(1)
	go policy.rescanReqTxDetailByAddr(acc.Addr, wg)
}

// OnAddBlockFinish 在区块被添加成功时做一些处理
func (policy *privacyPolicy) OnAddBlockFinish(block *types.BlockDetail) {

}

// OnDeleteBlockFinish 在区块被删除成功时做一些处理
func (policy *privacyPolicy) OnDeleteBlockFinish(block *types.BlockDetail) {

}

// OnClose 在钱包关闭时做一些处理
func (policy *privacyPolicy) OnClose() {

}

// OnSetQueueClient 在钱包消息队列初始化时做一些处理
func (policy *privacyPolicy) OnSetQueueClient() {
	version := policy.store.getVersion()
	if version < PRIVACYDBVERSION {
		policy.rescanAllTxAddToUpdateUTXOs()
		policy.store.setVersion()
	}
}

// OnWalletLocked 在钱包加锁时做一些处理
func (policy *privacyPolicy) OnWalletLocked() {
}

// OnWalletUnlocked 在钱包解锁时做一些处理
func (policy *privacyPolicy) OnWalletUnlocked(WalletUnLock *types.WalletUnLock) {
}

// Call 调用隐私的方法
func (policy *privacyPolicy) Call(funName string, in types.Message) (ret types.Message, err error) {
	switch funName {
	case "GetUTXOScaningFlag":
		isok := policy.GetRescanFlag() == privacytypes.UtxoFlagScaning
		ret = &types.Reply{IsOk: isok}
	default:
		err = types.ErrNotSupport
	}
	return
}

// SignTransaction 对隐私交易进行签名
func (policy *privacyPolicy) SignTransaction(key crypto.PrivKey, req *types.ReqSignRawTx) (needSysSign bool, signtxhex string, err error) {
	needSysSign = false
	bytes, err := common.FromHex(req.GetTxHex())
	if err != nil {
		bizlog.Error("SignTransaction", "common.FromHex error", err)
		return
	}
	tx := new(types.Transaction)
	if err = types.Decode(bytes, tx); err != nil {
		bizlog.Error("SignTransaction", "Decode Transaction error", err)
		return
	}
	signParam := &privacytypes.PrivacySignatureParam{}
	if err = types.Decode(tx.Signature.Signature, signParam); err != nil {
		bizlog.Error("SignTransaction", "Decode PrivacySignatureParam error", err)
		return
	}
	action := new(privacytypes.PrivacyAction)
	if err = types.Decode(tx.Payload, action); err != nil {
		bizlog.Error("SignTransaction", "Decode PrivacyAction error", err)
		return
	}
	if action.Ty != signParam.ActionType {
		bizlog.Error("SignTransaction", "action type ", action.Ty, "signature action type ", signParam.ActionType)
		return
	}
	switch action.Ty {
	case privacytypes.ActionPublic2Privacy:
		// 隐私交易的公对私动作，不存在交易组的操作
		tx.Sign(int32(policy.getWalletOperate().GetSignType()), key)

	case privacytypes.ActionPrivacy2Privacy, privacytypes.ActionPrivacy2Public:
		// 隐私交易的私对私、私对公需要进行特殊签名
		if err = policy.signatureTx(tx, action.GetInput(), signParam.GetUtxobasics(), signParam.GetRealKeyInputs()); err != nil {
			return
		}
	default:
		bizlog.Error("SignTransaction", "Invalid action type ", action.Ty)
		err = types.ErrInvalidParam
	}
	signtxhex = common.ToHex(types.Encode(tx))
	return
}

type buildStoreWalletTxDetailParam struct {
	tokenname    string
	block        *types.BlockDetail
	tx           *types.Transaction
	index        int
	newbatch     db.Batch
	senderRecver string
	isprivacy    bool
	addDelType   int32
	sendRecvFlag int32
	utxos        []*privacytypes.UTXO
}

// OnAddBlockTx 响应区块交易添加的处理
func (policy *privacyPolicy) OnAddBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbbatch db.Batch) *types.WalletTxDetail {
	policy.addDelPrivacyTxsFromBlock(tx, index, block, dbbatch, AddTx)
	// 自己处理掉所有事务，部需要外部处理了
	return nil
}

// OnDeleteBlockTx 响应删除区块交易的处理
func (policy *privacyPolicy) OnDeleteBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbbatch db.Batch) *types.WalletTxDetail {
	policy.addDelPrivacyTxsFromBlock(tx, index, block, dbbatch, DelTx)
	// 自己处理掉所有事务，部需要外部处理了
	return nil
}

// GetRescanFlag get rescan utxo flag
func (policy *privacyPolicy) GetRescanFlag() int32 {
	return atomic.LoadInt32(&policy.rescanUTXOflag)
}

// SetRescanFlag set rescan utxos flag
func (policy *privacyPolicy) SetRescanFlag(flag int32) {
	atomic.StoreInt32(&policy.rescanUTXOflag, flag)
}
