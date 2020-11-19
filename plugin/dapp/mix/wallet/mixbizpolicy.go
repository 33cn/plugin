// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of policy source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"sync"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

var (
	bizlog = log15.New("module", "wallet.mix")
	// MaxTxHashsPerTime 单词处理的最大哈希书
	MaxTxHashsPerTime int64 = 100
	// maxTxNumPerBlock 单个区块最大数
	maxTxNumPerBlock int64 = types.MaxTxsPerBlock
)

func init() {
	wcom.RegisterPolicy(mixTy.MixX, New())
}

// New 创建一盒钱包业务策略
func New() wcom.WalletBizPolicy {
	return &mixPolicy{
		mtx:      &sync.Mutex{},
		rescanwg: &sync.WaitGroup{},
	}
}

type mixPolicy struct {
	mtx           *sync.Mutex
	store         *mixStore
	walletOperate wcom.WalletOperate
	rescanwg      *sync.WaitGroup
}

func (policy *mixPolicy) setWalletOperate(walletBiz wcom.WalletOperate) {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	policy.walletOperate = walletBiz
}

func (policy *mixPolicy) getWalletOperate() wcom.WalletOperate {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	return policy.walletOperate
}

// Init 初始化处理
func (policy *mixPolicy) Init(walletOperate wcom.WalletOperate, sub []byte) {
	policy.setWalletOperate(walletOperate)
	policy.store = newStore(walletOperate.GetDBStore())

}

// OnCreateNewAccount 在账号创建时做一些处理
func (policy *mixPolicy) OnCreateNewAccount(acc *types.Account) {

}

// OnImportPrivateKey 在私钥导入时做一些处理
func (policy *mixPolicy) OnImportPrivateKey(acc *types.Account) {

}

// OnAddBlockFinish 在区块被添加成功时做一些处理
func (policy *mixPolicy) OnAddBlockFinish(block *types.BlockDetail) {

}

// OnDeleteBlockFinish 在区块被删除成功时做一些处理
func (policy *mixPolicy) OnDeleteBlockFinish(block *types.BlockDetail) {

}

// OnClose 在钱包关闭时做一些处理
func (policy *mixPolicy) OnClose() {

}

// OnSetQueueClient 在钱包消息队列初始化时做一些处理
func (policy *mixPolicy) OnSetQueueClient() {

}

// OnWalletLocked 在钱包加锁时做一些处理
func (policy *mixPolicy) OnWalletLocked() {
}

// OnWalletUnlocked 在钱包解锁时做一些处理
func (policy *mixPolicy) OnWalletUnlocked(WalletUnLock *types.WalletUnLock) {
}

// OnAddBlockTx 响应区块交易添加的处理
func (policy *mixPolicy) OnAddBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbBatch db.Batch) *types.WalletTxDetail {
	dbSet, err := policy.execAutoLocalMix(tx, block.Receipts[index], int(index), block.Block.Height)
	if err != nil {
		return nil
	}
	for _, kv := range dbSet.KV {
		dbBatch.Set(kv.Key, kv.Value)
	}
	// 自己处理掉所有事务，部需要外部处理了
	return nil
}

// OnDeleteBlockTx 响应删除区块交易的处理
func (policy *mixPolicy) OnDeleteBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbBatch db.Batch) *types.WalletTxDetail {
	dbSet, err := policy.execAutoDelLocal(tx)
	if err != nil {
		return nil
	}
	for _, kv := range dbSet.KV {
		dbBatch.Set(kv.Key, kv.Value)
	}

	return nil
}

// Call 调用隐私的方法
func (policy *mixPolicy) Call(funName string, in types.Message) (ret types.Message, err error) {
	switch funName {
	case "GetScanFlag":

		isok := policy.store.getRescanNoteStatus() == int32(mixTy.MixWalletRescanStatus_SCANNING)
		ret = &types.Reply{IsOk: isok}
	default:
		err = types.ErrNotSupport
	}
	return
}

// SignTransaction 对隐私交易进行签名
func (policy *mixPolicy) SignTransaction(key crypto.PrivKey, req *types.ReqSignRawTx) (needSysSign bool, signtxhex string, err error) {

	return false, "", types.ErrNotSupport
}
