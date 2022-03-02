// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"sync"
	"time"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	ty "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

var (
	bizlog = log15.New("module", "wallet.paracross")
)

func init() {
	wcom.RegisterPolicy(ty.ParaX, New())
}

// New new instance
func New() wcom.WalletBizPolicy {
	return &ParaPolicy{mtx: &sync.Mutex{}}
}

// ParaPolicy 结构体
type ParaPolicy struct {
	mtx           *sync.Mutex
	walletOperate wcom.WalletOperate
	minertimeout  *time.Timer
}

func (policy *ParaPolicy) setWalletOperate(walletBiz wcom.WalletOperate) {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	policy.walletOperate = walletBiz
}

func (policy *ParaPolicy) getWalletOperate() wcom.WalletOperate {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	return policy.walletOperate
}

func (policy *ParaPolicy) getAPI() client.QueueProtocolAPI {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	return policy.walletOperate.GetAPI()
}

// Init initial
func (policy *ParaPolicy) Init(walletBiz wcom.WalletOperate, sub []byte) {
	policy.setWalletOperate(walletBiz)
}

// OnWalletLocked process lock event
func (policy *ParaPolicy) OnWalletLocked() {
	var walletsatus types.WalletStatus
	wallet := policy.getWalletOperate()
	walletsatus.IsWalletLock = wallet.IsWalletLocked()
	NotifyConsensus(policy.getAPI(), "WalletStatus", types.Encode(&walletsatus))
	bizlog.Info("OnWalletLocked", "IsWalletLock", walletsatus.IsWalletLock)
}

//解锁超时处理，需要区分整个钱包的解锁或者只挖矿的解锁
func (policy *ParaPolicy) resetTimeout(Timeout int64) {
	if policy.minertimeout == nil {
		policy.minertimeout = time.AfterFunc(time.Second*time.Duration(Timeout), func() {
			var walletsatus types.WalletStatus
			wallet := policy.getWalletOperate()
			walletsatus.IsWalletLock = wallet.IsWalletLocked()
			NotifyConsensus(policy.getAPI(), "WalletStatus", types.Encode(&walletsatus))
			bizlog.Info("resetTimeout", "IsWalletLock", walletsatus.IsWalletLock)
		})
	} else {
		policy.minertimeout.Reset(time.Second * time.Duration(Timeout))
	}
}

// OnWalletUnlocked process unlock event,只处理wallet锁
func (policy *ParaPolicy) OnWalletUnlocked(param *types.WalletUnLock) {
	if !param.WalletOrTicket {
		if param.Timeout != 0 {
			policy.resetTimeout(param.Timeout)
		}
		var walletsatus types.WalletStatus
		wallet := policy.getWalletOperate()
		walletsatus.IsWalletLock = wallet.IsWalletLocked()
		NotifyConsensus(policy.getAPI(), "WalletStatus", types.Encode(&walletsatus))
		bizlog.Info("OnWalletUnlocked", "IsWalletLock", walletsatus.IsWalletLock)
	}

}

// OnCreateNewAccount 通知para共识有新账户创建
func (policy *ParaPolicy) OnCreateNewAccount(acc *types.Account) {
	NotifyConsensus(policy.getAPI(), "CreateNewAccount", types.Encode(acc))
	bizlog.Info("OnCreateNewAccount", "Addr", acc.Addr)
}

// OnImportPrivateKey 通知para共识有新账户导入
func (policy *ParaPolicy) OnImportPrivateKey(acc *types.Account) {
	NotifyConsensus(policy.getAPI(), "CreateNewAccount", types.Encode(acc))
	bizlog.Info("OnImportPrivateKey", "Addr", acc.Addr)
}

// NotifyConsensus 通知para共识模块做相应的处理
func NotifyConsensus(api client.QueueProtocolAPI, FuncName string, param []byte) {
	bizlog.Info("Wallet Notify Consensus")
	api.Notify("consensus", types.EventConsensusQuery, &types.ChainExecutor{
		Driver:   "para",
		FuncName: FuncName,
		Param:    param,
	})
}

// OnClose close
func (policy *ParaPolicy) OnClose() {
}

// OnSetQueueClient on set queue client
func (policy *ParaPolicy) OnSetQueueClient() {
}

// Call call
func (policy *ParaPolicy) Call(funName string, in types.Message) (ret types.Message, err error) {
	err = types.ErrNotSupport
	return
}

// OnAddBlockTx add Block tx
func (policy *ParaPolicy) OnAddBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbbatch db.Batch) *types.WalletTxDetail {
	return policy.proceWalletTxDetail(block, tx, index)
}

// OnDeleteBlockTx on delete block
func (policy *ParaPolicy) OnDeleteBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbbatch db.Batch) *types.WalletTxDetail {
	return policy.proceWalletTxDetail(block, tx, index)
}

// SignTransaction sign tx
func (policy *ParaPolicy) SignTransaction(key crypto.PrivKey, req *types.ReqSignRawTx) (needSysSign bool, signtx string, err error) {
	needSysSign = true
	return
}

// OnAddBlockFinish process finish block
func (policy *ParaPolicy) OnAddBlockFinish(block *types.BlockDetail) {
}

// OnDeleteBlockFinish process finish block
func (policy *ParaPolicy) OnDeleteBlockFinish(block *types.BlockDetail) {
}

func (policy *ParaPolicy) proceWalletTxDetail(block *types.BlockDetail, tx *types.Transaction, index int32) *types.WalletTxDetail {
	receipt := block.Receipts[index]
	amount, err := tx.Amount()
	if err != nil {
		bizlog.Error("proceWalletTxDetail:tx.Amount()", "err", err)
	}
	wtxdetail := &types.WalletTxDetail{
		Tx:         tx,
		Height:     block.Block.Height,
		Index:      int64(index),
		Receipt:    receipt,
		Blocktime:  block.Block.BlockTime,
		ActionName: tx.ActionName(),
		Amount:     amount,
		Payload:    nil,
	}
	if len(wtxdetail.Fromaddr) <= 0 {
		//from addr
		fromaddress := tx.From()
		if len(fromaddress) != 0 && policy.walletOperate.AddrInWallet(fromaddress) {
			wtxdetail.Fromaddr = fromaddress
		}
	}
	if len(wtxdetail.Fromaddr) <= 0 {
		toaddr := tx.GetTo()
		if len(toaddr) != 0 && policy.walletOperate.AddrInWallet(toaddr) {
			wtxdetail.Fromaddr = toaddr
		}
	}
	return wtxdetail
}
