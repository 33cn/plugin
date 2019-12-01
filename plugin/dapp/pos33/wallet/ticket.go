// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

var (
	minerAddrWhiteList = make(map[string]bool)
	bizlog             = log15.New("module", "wallet.pos33")
)

func init() {
	wcom.RegisterPolicy(ty.Pos33TicketX, New())
}

// New new instance
func New() wcom.WalletBizPolicy {
	return &ticketPolicy{mtx: &sync.Mutex{}}
}

type ticketPolicy struct {
	mtx                     *sync.Mutex
	walletOperate           wcom.WalletOperate
	store                   *ticketStore
	needFlush               bool
	miningPos33TicketTicker *time.Ticker
	autoMinerFlag           int32
	isPos33TicketLocked     int32
	minertimeout            *time.Timer
	cfg                     *subConfig
}

type subConfig struct {
	MinerWaitTime  string   `json:"minerWaitTime"`
	ForceMining    bool     `json:"forceMining"`
	Minerdisable   bool     `json:"minerdisable"`
	Minerwhitelist []string `json:"minerwhitelist"`
}

func (policy *ticketPolicy) initMingPos33TicketTicker(wait time.Duration) {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	bizlog.Info("initMingPos33TicketTicker", "Duration", wait)
	policy.miningPos33TicketTicker = time.NewTicker(wait)
}

func (policy *ticketPolicy) getMingPos33TicketTicker() *time.Ticker {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	return policy.miningPos33TicketTicker
}

func (policy *ticketPolicy) setWalletOperate(walletBiz wcom.WalletOperate) {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	policy.walletOperate = walletBiz
}

func (policy *ticketPolicy) getWalletOperate() wcom.WalletOperate {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	return policy.walletOperate
}

func (policy *ticketPolicy) getAPI() client.QueueProtocolAPI {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	return policy.walletOperate.GetAPI()
}

// IsAutoMining check auto mining
func (policy *ticketPolicy) IsAutoMining() bool {
	return policy.isAutoMining()
}

// IsPos33TicketLocked check lock status
func (policy *ticketPolicy) IsTicketLocked() bool {
	return atomic.LoadInt32(&policy.isPos33TicketLocked) != 0
}

// Init initial
func (policy *ticketPolicy) Init(walletBiz wcom.WalletOperate, sub []byte) {
	policy.setWalletOperate(walletBiz)
	policy.store = newStore(walletBiz.GetDBStore())
	policy.needFlush = false
	policy.isPos33TicketLocked = 1
	policy.autoMinerFlag = policy.store.GetAutoMinerFlag()
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}
	policy.cfg = &subcfg
	policy.initMinerWhiteList(walletBiz.GetConfig())
	wait := 2 * time.Minute
	if subcfg.MinerWaitTime != "" {
		d, err := time.ParseDuration(subcfg.MinerWaitTime)
		if err == nil {
			wait = d
		}
	}
	policy.initMingPos33TicketTicker(wait)
	walletBiz.RegisterMineStatusReporter(policy)
	// 启动自动挖矿
	walletBiz.GetWaitGroup().Add(1)
	go policy.autoMining()
}

// OnClose close
func (policy *ticketPolicy) OnClose() {
	policy.getMingPos33TicketTicker().Stop()
}

// OnSetQueueClient on set queue client
func (policy *ticketPolicy) OnSetQueueClient() {

}

// Call call
func (policy *ticketPolicy) Call(funName string, in types.Message) (ret types.Message, err error) {
	err = types.ErrNotSupport
	return
}

// OnAddBlockTx add Block tx
func (policy *ticketPolicy) OnAddBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbbatch db.Batch) *types.WalletTxDetail {
	receipt := block.Receipts[index]
	amount, _ := tx.Amount()
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
		pubkey := tx.Signature.GetPubkey()
		address := address.PubKeyToAddress(pubkey)
		//from addr
		fromaddress := address.String()
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

	if policy.checkNeedFlushPos33Ticket(tx, receipt) {
		policy.needFlush = true
	}
	return wtxdetail
}

// OnDeleteBlockTx on delete block
func (policy *ticketPolicy) OnDeleteBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbbatch db.Batch) *types.WalletTxDetail {
	receipt := block.Receipts[index]
	amount, _ := tx.Amount()
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
		pubkey := tx.Signature.GetPubkey()
		address := address.PubKeyToAddress(pubkey)
		//from addr
		fromaddress := address.String()
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

	if policy.checkNeedFlushPos33Ticket(tx, receipt) {
		policy.needFlush = true
	}
	return wtxdetail
}

// SignTransaction sign tx
func (policy *ticketPolicy) SignTransaction(key crypto.PrivKey, req *types.ReqSignRawTx) (needSysSign bool, signtx string, err error) {
	needSysSign = true
	return
}

// OnWalletLocked process lock event
func (policy *ticketPolicy) OnWalletLocked() {
	// 钱包锁住时，不允许挖矿
	atomic.CompareAndSwapInt32(&policy.isPos33TicketLocked, 0, 1)
	FlushPos33Ticket(policy.getAPI())
}

//解锁超时处理，需要区分整个钱包的解锁或者只挖矿的解锁
func (policy *ticketPolicy) resetTimeout(Timeout int64) {
	if policy.minertimeout == nil {
		policy.minertimeout = time.AfterFunc(time.Second*time.Duration(Timeout), func() {
			//wallet.isPos33TicketLocked = true
			atomic.CompareAndSwapInt32(&policy.isPos33TicketLocked, 0, 1)
		})
	} else {
		policy.minertimeout.Reset(time.Second * time.Duration(Timeout))
	}
}

// OnWalletUnlocked process unlock event
func (policy *ticketPolicy) OnWalletUnlocked(param *types.WalletUnLock) {
	if param.WalletOrTicket {
		atomic.CompareAndSwapInt32(&policy.isPos33TicketLocked, 1, 0)
		if param.Timeout != 0 {
			policy.resetTimeout(param.Timeout)
		}
	}
	// 钱包解锁时，需要刷新，通知挖矿
	FlushPos33Ticket(policy.getAPI())
}

// OnCreateNewAccount process create new account event
func (policy *ticketPolicy) OnCreateNewAccount(acc *types.Account) {
}

// OnImportPrivateKey 导入key的时候flush ticket
func (policy *ticketPolicy) OnImportPrivateKey(acc *types.Account) {
	FlushPos33Ticket(policy.getAPI())
}

// OnAddBlockFinish process finish block
func (policy *ticketPolicy) OnAddBlockFinish(block *types.BlockDetail) {
	if policy.needFlush {
		// 新增区块，由于ticket具有锁定期，所以这里不需要刷新
		//policy.flushPos33Ticket()
	}
	policy.needFlush = false
}

// OnDeleteBlockFinish process finish block
func (policy *ticketPolicy) OnDeleteBlockFinish(block *types.BlockDetail) {
	if policy.needFlush {
		FlushPos33Ticket(policy.getAPI())
	}
	policy.needFlush = false
}

// FlushPos33Ticket flush ticket
func FlushPos33Ticket(api client.QueueProtocolAPI) {
	bizlog.Info("wallet FLUSH TICKET")
	api.Notify("consensus", types.EventConsensusQuery, &types.ChainExecutor{
		Driver:   "pos33",
		FuncName: "FlushPos33Ticket",
		Param:    types.Encode(&types.ReqNil{}),
	})
}

func (policy *ticketPolicy) needFlushPos33Ticket(tx *types.Transaction, receipt *types.ReceiptData) bool {
	pubkey := tx.Signature.GetPubkey()
	addr := address.PubKeyToAddress(pubkey)
	return policy.store.checkAddrIsInWallet(addr.String())
}

func (policy *ticketPolicy) checkNeedFlushPos33Ticket(tx *types.Transaction, receipt *types.ReceiptData) bool {
	if receipt.Ty != types.ExecOk {
		return false
	}
	return policy.needFlushPos33Ticket(tx, receipt)
}

func (policy *ticketPolicy) forceClosePos33Ticket(height int64, minerAddr string) (*types.ReplyHashes, error) {
	if minerAddr != "" {
		return policy.forceClosePos33TicketByReturnAddr(height, minerAddr)
	}

	return policy.forceCloseAllPos33Ticket(height)
}

//通过minerAddr地址找到绑定的ticket的returnAddr，通过returnAddr的私钥close ticket，防止miner的私钥丢失场景
func (policy *ticketPolicy) forceClosePos33TicketByReturnAddr(height int64, minerAddr string) (*types.ReplyHashes, error) {
	tListMiner, err := policy.getForceClosePos33Tickets(minerAddr)
	if err != nil {
		return nil, err
	}
	tListMap := make(map[string][]*ty.Pos33Ticket)
	for _, ticket := range tListMiner {
		tListMap[ticket.ReturnAddress] = append(tListMap[ticket.ReturnAddress], ticket)
	}

	keys, err := policy.getWalletOperate().GetAllPrivKeys()
	if err != nil {
		return nil, err
	}

	var hashes types.ReplyHashes
	for _, key := range keys {
		returnAddr := address.PubKeyToAddress(key.PubKey().Bytes()).String()
		if len(tListMap[returnAddr]) == 0 {
			continue
		}
		hash, err := policy.forceClosePos33TicketList(height, key, tListMap[returnAddr])
		if err != nil {
			bizlog.Error("forceClosePos33TicketByAddr", "error", err, "returnAddr", returnAddr, "minerAddr", minerAddr)
			continue
		}
		if hash == nil {
			continue
		}
		hashes.Hashes = append(hashes.Hashes, hash)
	}
	return &hashes, nil
}

func (policy *ticketPolicy) forceCloseAllPos33Ticket(height int64) (*types.ReplyHashes, error) {
	var err error

	keys, err := policy.getWalletOperate().GetAllPrivKeys()
	if err != nil {
		return nil, err
	}

	var hashes types.ReplyHashes
	for _, key := range keys {
		addr := address.PubKeyToAddress(key.PubKey().Bytes()).String()
		tlist, err := policy.getForceClosePos33Tickets(addr)
		if err != nil {
			bizlog.Error("forceCloseAllPos33Ticket getPos33Tickets", "error", err, "addr", addr)
			continue
		}
		hash, err := policy.forceClosePos33TicketList(height, key, tlist)
		if err != nil {
			bizlog.Error("forceCloseAllPos33Ticket", "error", err, "addr", addr)
			continue
		}
		if hash == nil {
			continue
		}
		hashes.Hashes = append(hashes.Hashes, hash)
	}
	return &hashes, nil
}

func (policy *ticketPolicy) getPos33Tickets(addr string, status int32) ([]*ty.Pos33Ticket, error) {
	reqaddr := &ty.Pos33TicketList{Addr: addr, Status: status}
	api := policy.getAPI()
	msg, err := api.Query(ty.Pos33TicketX, "Pos33TicketList", reqaddr)
	if err != nil {
		bizlog.Error("getPos33Tickets", "addr", addr, "status", status, "Query error", err)
		return nil, err
	}
	reply := msg.(*ty.ReplyPos33TicketList)
	return reply.Tickets, nil
}

func (policy *ticketPolicy) getForceClosePos33Tickets(addr string) ([]*ty.Pos33Ticket, error) {
	if addr == "" {
		return nil, nil
	}
	tlist1, err1 := policy.getPos33Tickets(addr, 1)
	if err1 != nil && err1 != types.ErrNotFound {
		return nil, err1
	}
	tlist2, err2 := policy.getPos33Tickets(addr, 2)
	if err2 != nil && err2 != types.ErrNotFound {
		return nil, err1
	}

	return append(tlist1, tlist2...), nil
}

func (policy *ticketPolicy) forceClosePos33TicketList(height int64, priv crypto.PrivKey, tlist []*ty.Pos33Ticket) ([]byte, error) {
	var ids []string
	var tl []*ty.Pos33Ticket
	now := types.Now().Unix()
	chain33Cfg := policy.walletOperate.GetAPI().GetConfig()
	cfg := ty.GetPos33TicketMinerParam(chain33Cfg, height)
	for _, t := range tlist {
		if !t.IsGenesis {
			if t.Status == 1 && now-t.GetCreateTime() < cfg.Pos33TicketWithdrawTime {
				continue
			}
			if t.Status == 2 && now-t.GetCreateTime() < cfg.Pos33TicketWithdrawTime {
				continue
			}
			if t.Status == 2 && now-t.GetMinerTime() < cfg.Pos33TicketMinerWaitTime {
				continue
			}
		}
		tl = append(tl, t)
	}
	for i := 0; i < len(tl); i++ {
		ids = append(ids, tl[i].TicketId)
	}
	if len(ids) > 0 {
		return policy.closePos33Tickets(priv, ids)
	}
	return nil, nil
}

//通过rpc 精选close 操作
func (policy *ticketPolicy) closePos33Tickets(priv crypto.PrivKey, ids []string) ([]byte, error) {
	//每次最多close 200个
	end := 200
	if end > len(ids) {
		end = len(ids)
	}
	bizlog.Info("closePos33Tickets", "ids", ids[0:end])
	ta := &ty.Pos33TicketAction{}
	tclose := &ty.Pos33TicketClose{TicketId: ids[0:end]}
	ta.Value = &ty.Pos33TicketAction_Tclose{Tclose: tclose}
	ta.Ty = ty.Pos33TicketActionClose
	return policy.getWalletOperate().SendTransaction(ta, []byte(ty.Pos33TicketX), priv, "")
}

func (policy *ticketPolicy) getPos33TicketsByStatus(status int32) ([]*ty.Pos33Ticket, [][]byte, error) {
	operater := policy.getWalletOperate()
	accounts, err := operater.GetWalletAccounts()
	if err != nil {
		return nil, nil, err
	}
	operater.GetMutex().Lock()
	defer operater.GetMutex().Unlock()
	ok, err := operater.CheckWalletStatus()
	if !ok && err != types.ErrOnlyTicketUnLocked {
		return nil, nil, err
	}
	//循环遍历所有的账户-->保证钱包已经解锁
	var tickets []*ty.Pos33Ticket
	var privs [][]byte
	for _, acc := range accounts {
		t, err := policy.getPos33Tickets(acc.Addr, status)
		if err == types.ErrNotFound {
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		if t != nil {
			priv, err := operater.GetPrivKeyByAddr(acc.Addr)
			if err != nil {
				return nil, nil, err
			}
			privs = append(privs, priv.Bytes())
			tickets = append(tickets, t...)
		}
	}
	if len(tickets) == 0 {
		return nil, nil, ty.ErrNoPos33Ticket
	}
	return tickets, privs, nil
}

func (policy *ticketPolicy) setAutoMining(flag int32) {
	atomic.StoreInt32(&policy.autoMinerFlag, flag)
}

func (policy *ticketPolicy) isAutoMining() bool {
	return atomic.LoadInt32(&policy.autoMinerFlag) == 1
}

func (policy *ticketPolicy) closePos33TicketsByAddr(height int64, priv crypto.PrivKey) ([]byte, error) {
	addr := address.PubKeyToAddress(priv.PubKey().Bytes()).String()
	tlist, err := policy.getPos33Tickets(addr, 2)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	var ids []string
	var tl []*ty.Pos33Ticket
	now := types.Now().Unix()
	chain33Cfg := policy.walletOperate.GetAPI().GetConfig()
	cfg := ty.GetPos33TicketMinerParam(chain33Cfg, height)
	for _, t := range tlist {
		if !t.IsGenesis {
			if now-t.GetCreateTime() < cfg.Pos33TicketWithdrawTime {
				continue
			}
			if now-t.GetMinerTime() < cfg.Pos33TicketMinerWaitTime {
				continue
			}
		}
		tl = append(tl, t)
	}
	for i := 0; i < len(tl); i++ {
		ids = append(ids, tl[i].TicketId)
	}
	if len(ids) > 0 {
		return policy.closePos33Tickets(priv, ids)
	}
	return nil, nil
}

func (policy *ticketPolicy) closeAllPos33Tickets(height int64) (int, error) {
	operater := policy.getWalletOperate()
	keys, err := operater.GetAllPrivKeys()
	if err != nil {
		return 0, err
	}
	var hashes [][]byte
	for _, key := range keys {
		hash, err := policy.closePos33TicketsByAddr(height, key)
		if err != nil {
			bizlog.Error("close Pos33Tickets By Addr", "err", err)
			continue
		}
		if hash == nil {
			continue
		}
		hashes = append(hashes, hash)
	}
	if len(hashes) > 0 {
		operater.WaitTxs(hashes)
		return len(hashes), nil
	}
	return 0, nil
}

func (policy *ticketPolicy) closePos33Ticket(height int64) (int, error) {
	return policy.closeAllPos33Tickets(height)
}

func (policy *ticketPolicy) processFee(priv crypto.PrivKey) error {
	addr := address.PubKeyToAddress(priv.PubKey().Bytes()).String()
	operater := policy.getWalletOperate()
	acc1, err := operater.GetBalance(addr, "coins")
	if err != nil {
		return err
	}
	acc2, err := operater.GetBalance(addr, ty.Pos33TicketX)
	if err != nil {
		return err
	}
	toaddr := address.ExecAddress(ty.Pos33TicketX)
	//如果acc2 的余额足够，那题withdraw 部分钱做手续费
	if (acc1.Balance < (types.Coin / 2)) && (acc2.Balance > types.Coin) {
		_, err := operater.SendToAddress(priv, toaddr, -types.Coin, "pos33->coins", false, "")
		if err != nil {
			return err
		}
	}
	return nil
}

//手续费处理
func (policy *ticketPolicy) processFees() error {
	keys, err := policy.getWalletOperate().GetAllPrivKeys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		e := policy.processFee(key)
		if e != nil {
			err = e
		}
	}
	return err
}

func (policy *ticketPolicy) withdrawFromPos33TicketOne(priv crypto.PrivKey) ([]byte, error) {
	addr := address.PubKeyToAddress(priv.PubKey().Bytes()).String()
	operater := policy.getWalletOperate()
	acc, err := operater.GetBalance(addr, ty.Pos33TicketX)
	if err != nil {
		return nil, err
	}
	if acc.Balance > 0 {
		hash, err := operater.SendToAddress(priv, address.ExecAddress(ty.Pos33TicketX), -acc.Balance, "autominer->withdraw", false, "")
		if err != nil {
			return nil, err
		}
		return hash.GetHash(), nil
	}
	return nil, nil
}

func (policy *ticketPolicy) openticket(mineraddr, returnaddr string, priv crypto.PrivKey, count int32) ([]byte, error) {
	bizlog.Info("openticket", "mineraddr", mineraddr, "returnaddr", returnaddr, "count", count)
	if count > ty.Pos33TicketCountOpenOnce {
		count = ty.Pos33TicketCountOpenOnce
		bizlog.Info("openticket", "Update count", "wait for another open")
	}

	ta := &ty.Pos33TicketAction{}
	topen := &ty.Pos33TicketOpen{MinerAddress: mineraddr, ReturnAddress: returnaddr, Count: count, RandSeed: types.Now().UnixNano()}
	hashList := make([][]byte, int(count))
	for i := 0; i < int(count); i++ {
		privHash := common.Sha256([]byte(fmt.Sprintf("%x:%d:%d", priv.Bytes(), i, topen.RandSeed)))
		pubHash := common.Sha256(privHash)
		hashList[i] = pubHash
	}
	topen.PubHashes = hashList
	ta.Value = &ty.Pos33TicketAction_Topen{Topen: topen}
	ta.Ty = ty.Pos33TicketActionOpen
	return policy.walletOperate.SendTransaction(ta, []byte(ty.Pos33TicketX), priv, "")
}

func (policy *ticketPolicy) buyPos33TicketOne(height int64, priv crypto.PrivKey) ([]byte, int, error) {
	//ticket balance and coins balance
	addr := address.PubKeyToAddress(priv.PubKey().Bytes()).String()
	operater := policy.getWalletOperate()
	acc1, err := operater.GetBalance(addr, "coins")
	if err != nil {
		return nil, 0, err
	}
	acc2, err := operater.GetBalance(addr, ty.Pos33TicketX)
	if err != nil {
		return nil, 0, err
	}
	//留一个币作为手续费，如果手续费不够了，不能挖矿
	//判断手续费是否足够，如果不足要及时补充。
	chain33Cfg := policy.walletOperate.GetAPI().GetConfig()
	cfg := ty.GetPos33TicketMinerParam(chain33Cfg, height)
	fee := types.Coin
	if acc1.Balance+acc2.Balance-2*fee >= cfg.Pos33TicketPrice {
		// 如果可用余额+冻结余额，可以凑成新票，则转币到冻结余额
		if (acc1.Balance+acc2.Balance-2*fee)/cfg.Pos33TicketPrice > acc2.Balance/cfg.Pos33TicketPrice {
			//第一步。转移币到 ticket
			toaddr := address.ExecAddress(ty.Pos33TicketX)
			amount := acc1.Balance - 2*fee
			//必须大于0，才需要转移币
			var hash *types.ReplyHash
			if amount > 0 {
				bizlog.Info("buyPos33TicketOne.send", "toaddr", toaddr, "amount", amount)
				hash, err = policy.walletOperate.SendToAddress(priv, toaddr, amount, "coins->pos33", false, "")

				if err != nil {
					return nil, 0, err
				}
				operater.WaitTx(hash.Hash)
			}
		}

		acc, err := operater.GetBalance(addr, ty.Pos33TicketX)
		if err != nil {
			return nil, 0, err
		}
		count := acc.Balance / cfg.Pos33TicketPrice
		if count > 0 {
			txhash, err := policy.openticket(addr, addr, priv, int32(count))
			return txhash, int(count), err
		}
	}
	return nil, 0, nil
}

func (policy *ticketPolicy) buyPos33Ticket(height int64) ([][]byte, int, error) {
	privs, err := policy.getWalletOperate().GetAllPrivKeys()
	if err != nil {
		bizlog.Error("buyPos33Ticket.getAllPrivKeys", "err", err)
		return nil, 0, err
	}
	count := 0
	var hashes [][]byte
	bizlog.Debug("ticketPolicy buyPos33Ticket begin")
	for _, priv := range privs {
		hash, n, err := policy.buyPos33TicketOne(height, priv)
		if err != nil {
			bizlog.Error("ticketPolicy buyPos33Ticket buyPos33TicketOne", "err", err)
			continue
		}
		count += n
		if hash != nil {
			hashes = append(hashes, hash)
		}
		bizlog.Debug("ticketPolicy buyPos33Ticket", "Address", address.PubKeyToAddress(priv.PubKey().Bytes()).String(), "txhash", hex.EncodeToString(hash), "n", n)
	}
	bizlog.Debug("ticketPolicy buyPos33Ticket end")
	return hashes, count, nil
}

func (policy *ticketPolicy) getMinerColdAddr(addr string) ([]string, error) {
	reqaddr := &types.ReqString{Data: addr}
	api := policy.walletOperate.GetAPI()
	msg, err := api.Query(ty.Pos33TicketX, "MinerSourceList", reqaddr)
	if err != nil {
		bizlog.Error("getMinerColdAddr", "Query error", err)
		return nil, err
	}
	reply := msg.(*types.ReplyStrings)
	return reply.Datas, nil
}

func (policy *ticketPolicy) initMinerWhiteList(cfg *types.Wallet) {
	if len(policy.cfg.Minerwhitelist) == 0 {
		minerAddrWhiteList["*"] = true
		return
	}
	if len(policy.cfg.Minerwhitelist) == 1 && policy.cfg.Minerwhitelist[0] == "*" {
		minerAddrWhiteList["*"] = true
		return
	}
	for _, addr := range policy.cfg.Minerwhitelist {
		minerAddrWhiteList[addr] = true
	}
}

func checkMinerWhiteList(addr string) bool {
	if _, ok := minerAddrWhiteList["*"]; ok {
		return true
	}

	if _, ok := minerAddrWhiteList[addr]; ok {
		return true
	}
	return false
}

func (policy *ticketPolicy) buyMinerAddrPos33TicketOne(height int64, priv crypto.PrivKey) ([][]byte, int, error) {
	addr := address.PubKeyToAddress(priv.PubKey().Bytes()).String()
	//判断是否绑定了coldaddr
	addrs, err := policy.getMinerColdAddr(addr)
	if err != nil {
		return nil, 0, err
	}
	total := 0
	var hashes [][]byte
	chain33Cfg := policy.walletOperate.GetAPI().GetConfig()
	cfg := ty.GetPos33TicketMinerParam(chain33Cfg, height)
	for i := 0; i < len(addrs); i++ {
		bizlog.Info("sourceaddr", "addr", addrs[i])
		ok := checkMinerWhiteList(addrs[i])
		if !ok {
			bizlog.Info("buyMinerAddrPos33TicketOne Cold Addr not in MinerWhiteList", "addr", addrs[i])
			continue
		}
		acc, err := policy.getWalletOperate().GetBalance(addrs[i], ty.Pos33TicketX)
		if err != nil {
			return nil, 0, err
		}
		count := acc.Balance / cfg.Pos33TicketPrice
		if count > 0 {
			txhash, err := policy.openticket(addr, addrs[i], priv, int32(count))
			if err != nil {
				return nil, 0, err
			}
			total += int(count)
			if txhash != nil {
				hashes = append(hashes, txhash)
			}
		}
	}
	return hashes, total, nil
}

func (policy *ticketPolicy) buyMinerAddrPos33Ticket(height int64) ([][]byte, int, error) {
	privs, err := policy.getWalletOperate().GetAllPrivKeys()
	if err != nil {
		bizlog.Error("buyMinerAddrPos33Ticket.getAllPrivKeys", "err", err)
		return nil, 0, err
	}
	count := 0
	var hashes [][]byte
	bizlog.Debug("ticketPolicy buyMinerAddrPos33Ticket begin")
	for _, priv := range privs {
		hashlist, n, err := policy.buyMinerAddrPos33TicketOne(height, priv)
		if err != nil {
			if err != types.ErrNotFound {
				bizlog.Error("buyMinerAddrPos33TicketOne", "err", err)
			}
			continue
		}
		count += n
		if hashlist != nil {
			hashes = append(hashes, hashlist...)
		}
		bizlog.Debug("ticketPolicy buyMinerAddrPos33Ticket", "Address", address.PubKeyToAddress(priv.PubKey().Bytes()).String(), "n", n)
	}
	bizlog.Debug("ticketPolicy buyMinerAddrPos33Ticket end")
	return hashes, count, nil
}

func (policy *ticketPolicy) withdrawFromPos33Ticket() (hashes [][]byte, err error) {
	privs, err := policy.getWalletOperate().GetAllPrivKeys()
	if err != nil {
		bizlog.Error("withdrawFromPos33Ticket.getAllPrivKeys", "err", err)
		return nil, err
	}
	for _, priv := range privs {
		hash, err := policy.withdrawFromPos33TicketOne(priv)
		if err != nil {
			bizlog.Error("withdrawFromPos33TicketOne", "err", err)
			continue
		}
		if hash != nil {
			hashes = append(hashes, hash)
		}
	}
	return hashes, nil
}

//检查周期 --> 10分
//开启挖矿：
//1. 自动把成熟的ticket关闭
//2. 查找超过1万余额的账户，自动购买ticket
//3. 查找mineraddress 和他对应的 账户的余额（不在1中），余额超过1万的自动购买ticket 挖矿
//
//停止挖矿：
//1. 自动把成熟的ticket关闭
//2. 查找ticket 可取的余额
//3. 取出ticket 里面的钱
func (policy *ticketPolicy) autoMining() {
	bizlog.Info("Begin auto mining")
	defer bizlog.Info("End auto mining")
	operater := policy.getWalletOperate()
	defer operater.GetWaitGroup().Done()

	// 只有ticket共识下ticket相关的操作才有效
	cfg := policy.walletOperate.GetAPI().GetConfig()
	q := types.Conf(cfg, "config.consensus")
	if q != nil {
		cons := q.GStr("name")
		if strings.Compare(strings.TrimSpace(cons), ty.Pos33TicketX) != 0 {
			bizlog.Info("consensus is not ticket, exit mining")
			return
		}
	}

	lastHeight := int64(0)
	miningPos33TicketTicker := policy.getMingPos33TicketTicker()
	for {
		select {
		case <-miningPos33TicketTicker.C:
			if policy.cfg.Minerdisable {
				bizlog.Info("autoMining, GetMinerdisable() is true, exit autoMining()")
				break
			}
			if !(operater.IsCaughtUp() || policy.cfg.ForceMining) {
				bizlog.Error("wallet IsCaughtUp false")
				break
			}
			//判断高度是否增长
			height := operater.GetBlockHeight()
			if height <= lastHeight {
				bizlog.Error("wallet Height not inc", "height", height, "lastHeight", lastHeight)
				break
			}
			lastHeight = height
			bizlog.Info("BEG miningPos33Ticket")
			if policy.isAutoMining() {
				/*
					n1, err := policy.closePos33Ticket(lastHeight + 1)
					if err != nil {
						bizlog.Error("closePos33Ticket", "err", err)
					}
				*/
				n1 := 0
				err := policy.processFees()
				if err != nil {
					bizlog.Error("processFees", "err", err)
				}
				hashes1, n2, err := policy.buyPos33Ticket(lastHeight + 1)
				if err != nil {
					bizlog.Error("buyPos33Ticket", "err", err)
				}
				hashes2, n3, err := policy.buyMinerAddrPos33Ticket(lastHeight + 1)
				if err != nil {
					bizlog.Error("buyMinerAddrPos33Ticket", "err", err)
				}
				hashes := append(hashes1, hashes2...)
				if len(hashes) > 0 {
					operater.WaitTxs(hashes)
				}
				if n1+n2+n3 > 0 {
					FlushPos33Ticket(policy.getAPI())
				}
			} else {
				n1, err := policy.closePos33Ticket(lastHeight + 1)
				if err != nil {
					bizlog.Error("closePos33Ticket", "err", err)
				}
				err = policy.processFees()
				if err != nil {
					bizlog.Error("processFees", "err", err)
				}
				hashes, err := policy.withdrawFromPos33Ticket()
				if err != nil {
					bizlog.Error("withdrawFromPos33Ticket", "err", err)
				}
				if len(hashes) > 0 {
					operater.WaitTxs(hashes)
				}
				if n1 > 0 {
					FlushPos33Ticket(policy.getAPI())
				}
			}
			bizlog.Info("END miningPos33Ticket")
		case <-operater.GetWalletDone():
			return
		}
	}
}
