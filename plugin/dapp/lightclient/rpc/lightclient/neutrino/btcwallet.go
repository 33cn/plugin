package neutrino

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	lighttypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/chain"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/btcsuite/btcwallet/wtxmgr"
)

const (
	transactionTypeDeposit  string = "deposit"
	transactionTypeWithdraw string = "withdraw"
	_                       string = "mergeBalance"
)

const (
	// 默认确认数
	defaultRequiredConfs = 6
	// 最小找零金额（粉尘限制）
	minChangeAmount = 546
	// UTXO锁定时长
	utxoLeaseDuration = 24 * time.Hour
)

const (
	withdrawOpReturnPrefix    = "rgbx:" + transactionTypeWithdraw + ":"
	withdrawOpReturnDataLen   = len(withdrawOpReturnPrefix) + 32
	withdrawOpReturnScriptLen = 1 + 1 + withdrawOpReturnDataLen
)

const (
	btcwalletMonitorBucket = "rgbx-btcwallet-monitor"
	minPendingHeightKey    = "min-pending-height"
)

// utxoWithdrawLockID UTXO锁定ID
var utxoLockID = wtxmgr.LockID{
	'R', 'G', 'B', 'X', '-', 'L', 'O', 'C', 'K',
	'-', 'I', 'D', '-', 'V', '1', '.', '0', '.', '0',
	0, 0, 0, 0,
}

// DepositNotification 充值通知
type DepositNotification struct {
	TxHash       chainhash.Hash
	Amount       btcutil.Amount
	FromAddress  string
	Chain33Addr  string // 绑定的Chain33地址
	OpReturnData string // 原始OP_RETURN数据
}

// WithdrawNotification 提现确认通知
type WithdrawNotification struct {
	TxHash        chainhash.Hash
	ToAddress     string
	Chain33TxHash string // Chain33提现交易哈希
	OpReturnData  string // 原始OP_RETURN数据
}

// withdrawRequest 提现请求
type withdrawRequest struct {
	chain33WithDrawHash []byte
	amount              btcutil.Amount
	feeRate             btcutil.Amount // sat/byte，0表示使用默认
	toAddress           string
	stickyUTXO          *UTXO
}

type btcWallet struct {
	*wallet.Wallet
	client      *neutrinoClient
	chainParams chaincfg.Params
	chainClient chain.Interface
	rpcClient   *rpcclient.Client
	db          walletdb.DB

	minPendingHeight int32
	processedHeight  int32

	// TSS相关
	tssAddress  btcutil.Address
	tssPubKey   *btcec.PublicKey
	tssPkScript []byte // 预计算的TSS地址脚本

	// 通知channel
	depositChan       chan *btcPendingTx
	withdrawChan      chan *btcPendingTx
	addPendingChan    chan *btcPendingTx
	removePendingChan chan chainhash.Hash

	// 配置
	requiredConfs int32

	// 交易监控
	pendingTxs map[chainhash.Hash]*btcPendingTx
	rescanDone bool
	// txLock     sync.RWMutex
}

type btcPendingTx struct {
	tx                    *wire.MsgTx
	submitTime            time.Time
	notified              bool
	confirmations         int32
	blockHeight           int32
	blockHash             chainhash.Hash
	txHash                chainhash.Hash
	txType                string // "deposit" or "withdraw"
	depositAmount         btcutil.Amount
	withdrawAmount        btcutil.Amount
	chain33DepositAddress string // Chain33充值地址
	withdrawAddress       string
	chain33WithdrawTxHash []byte // Chain33提现交易哈希
	// OP_RETURN数据
	opReturnData opReturnData // 原始OP_RETURN解析数据
}

func newBtcWallet(n *neutrinoClient) (*btcWallet, error) {
	bw := &btcWallet{
		client:            n,
		chainParams:       n.neutrinoCfg.ChainParams,
		depositChan:       make(chan *btcPendingTx, 100),
		withdrawChan:      make(chan *btcPendingTx, 100),
		addPendingChan:    make(chan *btcPendingTx, 100),
		removePendingChan: make(chan chainhash.Hash, 100),
		requiredConfs:     int32(n.cfg.BlockConfirmations),
		pendingTxs:        make(map[chainhash.Hash]*btcPendingTx),
	}

	if n.cfg.BtcRPC.Host != "" {
		connCfg, err := n.cfg.BtcRPC.toConnConfig()
		if err != nil {
			log.Error("newBtcWallet btc rpc conn config error", "err", err)
			return nil, err
		}
		rpcCli, err := rpcclient.New(connCfg, nil)
		if err != nil {
			log.Error("newBtcWallet create btc rpc client error", "err", err)
			return nil, err
		}
		bw.rpcClient = rpcCli
	}

	exist, db, err := openWalletDB(n.neutrinoCfg.DataDir, "btcwallet.db")
	if err != nil {
		log.Error("newBtcWallet open db error", "err", err)
		return nil, err
	}

	pubPass := []byte("hello")
	if !exist {
		err = wallet.CreateWatchingOnly(db, pubPass, &bw.chainParams, types.Now())
		if err != nil {
			log.Error("newBtcWallet create wallet error", "err", err)
			_ = db.Close()
			return nil, err
		}
	}

	w, err := wallet.Open(db, pubPass, nil, &bw.chainParams, 0)
	if err != nil {
		log.Error("newBtcWallet open wallet error", "err", err)
		_ = db.Close()
		return nil, err
	}
	log.Info("newBtcWallet open wallet success", "wallet birthtime", w.Manager.Birthday().Unix())
	bw.db = db
	bw.Wallet = w
	bw.chainClient = chain.NewNeutrinoClient(&bw.chainParams, n.neutrinoCS)
	return bw, nil
}

func (b *btcWallet) start() error {
	if err := b.chainClient.Start(); err != nil {
		log.Error("btcwallet chainclient start error", "err", err)
		return err
	}

	b.Wallet.Start()

	// 启动交易监听
	go b.monitorTransactions()

	return nil
}

func (b *btcWallet) stop() {
	b.Wallet.Stop()
	b.chainClient.Stop()
	if b.rpcClient != nil {
		b.rpcClient.Shutdown()
	}
	_ = b.db.Close()
}

// waitAndImportTSSAddress 等待TSS地址生成并导入
func (b *btcWallet) waitAndImportTSSAddress() {

	b.tssPubKey = b.client.tss.tssPublicKey
	b.tssPkScript = b.client.tss.pkScript
	b.tssAddress = b.client.tss.tssAddress
	log.Debug("waitAndImportTSSAddress", "address", b.tssAddress.String())
	b.client.waitUntilDone("waitImportTSSAddress", func() bool {
		// 显式导入TSS公钥到钱包
		if _, err := b.Wallet.AddressInfo(b.tssAddress); err != nil {
			if !waddrmgr.IsError(err, waddrmgr.ErrAddressNotFound) {
				log.Error("waitAndImportTSSAddress AddressInfo failed", "err", err)
				return false
			}
			err = b.importTSSPublicKey()
			if err != nil {
				log.Error("waitAndImportTSSAddress ImportPublicKey failed", "err", err)
				return false
			}
		}
		return true
	}, time.Second*3)
	log.Info("waitAndImportTSSAddress success", "address", b.tssAddress.String())
}

func (b *btcWallet) importTSSPublicKey() error {
	err := b.Wallet.ImportPublicKey(b.tssPubKey, waddrmgr.WitnessPubKey)
	if err == nil {
		return nil
	}
	// Old/external wallet DBs might miss BIP84 scope (m/84'/coin'). Create it
	// on-demand and retry importing the TSS key.
	if !waddrmgr.IsError(err, waddrmgr.ErrScopeNotFound) {
		return err
	}

	scope := waddrmgr.KeyScopeBIP0084
	schema, ok := waddrmgr.ScopeAddrMap[scope]
	if !ok {
		return err
	}
	if _, addErr := b.Wallet.AddScopeManager(scope, schema); addErr != nil {
		log.Warn("importTSSPublicKey AddScopeManager failed, retry import anyway",
			"scope", scope, "err", addErr)
	}
	return b.Wallet.ImportPublicKey(b.tssPubKey, waddrmgr.WitnessPubKey)
}

func (b *btcWallet) loadMinPendingHeight() int32 {
	var height int32
	err := walletdb.View(b.client.neutrinoCfg.Database, func(tx walletdb.ReadTx) error {
		bucket := tx.ReadBucket([]byte(btcwalletMonitorBucket))
		if bucket == nil {
			return walletdb.ErrBucketNotFound
		}
		val := bucket.Get([]byte(minPendingHeightKey))
		if len(val) == 0 {
			return nil
		}
		reply := &types.Int64{}
		if err := types.Decode(val, reply); err != nil {
			return err
		}
		height = int32(reply.GetData())
		return nil
	})
	log.Debug("loadMinPendingHeight", "height", height, "err", err)
	if err != nil && !errors.Is(err, walletdb.ErrBucketNotFound) {
		log.Error("loadMinPendingHeight", "err", err)
	}
	return height
}

func (b *btcWallet) saveMinPendingHeight(height int32) {
	err := walletdb.Update(b.client.neutrinoCfg.Database, func(tx walletdb.ReadWriteTx) error {
		bucket, err := tx.CreateTopLevelBucket([]byte(btcwalletMonitorBucket))
		if err != nil {
			return err
		}
		if height <= 0 {
			return bucket.Delete([]byte(minPendingHeightKey))
		}
		data := types.Encode(&types.Int64{Data: int64(height)})
		return bucket.Put([]byte(minPendingHeightKey), data)
	})
	log.Debug("saveMinPendingHeight", "height", height, "err", err)
	if err != nil {
		log.Error("saveMinPendingHeight", "err", err, "height", height)
	}
}

func (b *btcWallet) updateMinPendingHeight() {
	minHeight := int32(0)
	for _, pending := range b.pendingTxs {
		if pending.blockHeight <= 0 {
			continue
		}
		if minHeight == 0 || pending.blockHeight < minHeight {
			minHeight = pending.blockHeight
		}
	}
	log.Debug("updateMinPendingHeight", "minHeight", minHeight,
		"processedHeight", b.processedHeight, "bestHeight", b.client.getBestBlockHeight(),
		"b.minPendingHeight", b.minPendingHeight, "pendingTxs", len(b.pendingTxs))
	if minHeight > 0 && minHeight != b.minPendingHeight {
		b.minPendingHeight = minHeight
		b.saveMinPendingHeight(minHeight)
	}
}

// rescanFromHeight 从指定高度开始重新扫描, 需要注意和wallet.SynchronizeRPC的同步关系，可能造成死锁
func (b *btcWallet) rescanFromHeight(height int32) error {

	log.Debug("rescanFromHeight", "height", height)
	hash, err := b.chainClient.GetBlockHash(int64(height))
	if err != nil {
		return err
	}
	stamp := waddrmgr.BlockStamp{
		Hash:   *hash,
		Height: height,
	}
	if header, err := b.chainClient.GetBlockHeader(hash); err == nil {
		stamp.Timestamp = header.Timestamp
	}
	job := &wallet.RescanJob{
		InitialSync: true,
		Addrs:       []btcutil.Address{b.tssAddress},
		BlockStamp:  stamp,
	}
	select {
	case err := <-b.Wallet.SubmitRescan(job):
		log.Debug("rescanFromHeight submitRescan done")
		return err
	case <-b.client.ctx.Done():
		return types.ErrChannelClosed
	}

}

func (b *btcWallet) rescanWalletTxs(start, end int32, rescanChan chan *wallet.GetTransactionsResult) error {
	b.client.waitUntilDone("walletTxsRescan", func() bool {
		return b.Wallet.ChainSynced()
	}, time.Second*2)
	log.Debug("rescanWalletTxs", "start", start, "end", end, "bestHeight", b.client.getBestBlockHeight())
	startBlock := wallet.NewBlockIdentifierFromHeight(start)
	endBlock := wallet.NewBlockIdentifierFromHeight(end)
	res, err := b.Wallet.GetTransactions(startBlock, endBlock, "", b.client.ctx.Done())
	if err != nil {
		log.Error("rescanWalletTxs GetTransactions error", "err", err)
		return err
	}
	rescanChan <- res
	return nil

}

func (b *btcWallet) handleNotify(attachedBlocks []wallet.Block, unminedTxs []wallet.TransactionSummary) {

	log.Debug("handleNotify", "attachedBlocks", len(attachedBlocks), "unminedTxs", len(unminedTxs))
	// 处理已确认交易， attachedBlocks是btc链上新添加的区块
	for _, block := range attachedBlocks {
		log.Debug("handleNotify block", "height", block.Height, "hash", block.Hash.String(), "transactions", len(block.Transactions))
		for _, tx := range block.Transactions {
			b.handleTransaction(&tx, block.Height, *block.Hash)
		}
		b.processedHeight = block.Height
	}

	// 处理未确认交易（重置确认数）
	for _, tx := range unminedTxs {
		b.handleUnminedTransaction(*tx.Hash)
	}
}

// monitorTransactions 监听交易通知
func (b *btcWallet) monitorTransactions() {

	client := b.Wallet.NtfnServer.TransactionNotifications()
	b.Wallet.SynchronizeRPC(b.chainClient)
	b.waitAndImportTSSAddress()
	rescanHeight := b.loadMinPendingHeight()
	bestHeight := b.client.getBestBlockHeight()
	if rescanHeight < int32(b.client.cfg.BtcHeaderStartHeight) {
		rescanHeight = int32(b.client.cfg.BtcHeaderStartHeight)
		log.Info("monitorTransactions initial rescan", "height", rescanHeight, "bestHeight", bestHeight)
	} else {
		log.Debug("monitorTransactions resume from height", "height", rescanHeight, "bestHeight", bestHeight)
	}
	b.minPendingHeight = rescanHeight
	interval := b.client.cfg.BtcBlockInterval/2 + 1
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	rescanChan := make(chan *wallet.GetTransactionsResult, 1)
	firstProcessFlag := true
	for {
		select {
		case <-b.client.ctx.Done():
			client.Done()
			return
		case res := <-rescanChan:
			b.handleNotify(res.MinedTransactions, res.UnminedTransactions)
			b.rescanDone = true

		case ntfn := <-client.C:
			if ntfn == nil {
				continue
			}
			// 首次处理检测是否需要重新扫描
			if firstProcessFlag && len(ntfn.AttachedBlocks) > 0 {
				log.Debug("monitorTransactions first process", "height", ntfn.AttachedBlocks[0].Height, "rescanHeight", rescanHeight)
				firstProcessFlag = false
				if ntfn.AttachedBlocks[0].Height > rescanHeight {
					go b.rescanWalletTxs(rescanHeight, ntfn.AttachedBlocks[0].Height-1, rescanChan)
				} else {
					b.rescanDone = true
				}
			}
			b.handleNotify(ntfn.AttachedBlocks, ntfn.UnminedTransactions)

		case pending := <-b.addPendingChan:
			b.pendingTxs[pending.txHash] = pending
			log.Debug("addPendingTx", "txHash", pending.txHash.String(), "txType", pending.txType)
		case txHash := <-b.removePendingChan:
			delete(b.pendingTxs, txHash)
			log.Debug("removePendingTx", "txHash", txHash.String(), "txLen", len(b.pendingTxs),
				"processedHeight", b.processedHeight, "minPendingHeight", b.minPendingHeight)
			if len(b.pendingTxs) > 0 {
				b.updateMinPendingHeight()
			}
		case <-ticker.C:
			if b.rescanDone && len(b.pendingTxs) == 0 && b.processedHeight > b.minPendingHeight {
				b.minPendingHeight = b.processedHeight
				b.saveMinPendingHeight(b.minPendingHeight)
				log.Debug("monitorTransactions update minPendingHeight",
					"minPendingHeight", b.minPendingHeight, "bestHeight", b.client.getBestBlockHeight())
			}
			b.updateTransactionConfirmations()
		}
	}
}

// handleTransaction 处理单个交易
func (b *btcWallet) handleTransaction(tx *wallet.TransactionSummary, blockHeight int32, blockHash chainhash.Hash) {
	txHash := *tx.Hash
	// 检查提现交易是否已在pending缓存中（避免重复解析），如果存在，则更新区块高度和区块哈希
	pending, exists := b.pendingTxs[txHash]
	if exists {
		pending.blockHeight = blockHeight
		pending.blockHash = blockHash
		log.Debug("handleTransaction already in pending", "txHash", txHash.String())
		return
	}

	log.Debug("handleTransaction processing", "txHash", txHash.String(), "blockHeight", blockHeight,
		"inputs", len(tx.Tx.TxIn), "outputs", len(tx.Tx.TxOut))

	// 分析交易类型和相关信息
	pending = b.analyzeTransaction(tx.Hash, tx.Tx)
	if pending == nil {
		log.Debug("handleTransaction not deposit/withdraw", "txHash", txHash.String())
		return
	}
	pending.tx = tx.Tx
	pending.blockHeight = blockHeight
	pending.blockHash = blockHash
	pending.txHash = txHash
	// 记录关键交易信息
	log.Info("handleTransaction detected "+pending.txType, "blockHeight", blockHeight, "txHash", txHash.String(),
		"depositAmount", pending.depositAmount, "withdrawAmount", pending.withdrawAmount)

	// 添加到待确认列表
	b.pendingTxs[txHash] = pending
}

// handleUnminedTransaction 处理未确认交易（重置确认数）
func (b *btcWallet) handleUnminedTransaction(txHash chainhash.Hash) {
	if pending, exists := b.pendingTxs[txHash]; exists && pending.blockHeight > 0 {
		// 重置确认数和区块高度
		oldConfirmations := pending.confirmations
		oldBlockHeight := pending.blockHeight

		pending.confirmations = 0
		pending.blockHeight = -1
		pending.notified = false

		log.Debug("handleUnminedTransaction reset confirmations", "txHash", txHash.String(),
			"oldConfirmations", oldConfirmations, "oldBlockHeight", oldBlockHeight, "type", pending.txType)
	}
}

// updateTransactionConfirmations 更新已存在交易的确认数
func (b *btcWallet) updateTransactionConfirmations() {
	bestHeight := b.client.getBestBlockHeight()
	log.Debug("updateTransactionConfirmations", "bestBlock", bestHeight, "pendingTxs", len(b.pendingTxs))
	for txHash, pending := range b.pendingTxs {
		if pending.blockHeight > 0 && bestHeight > 0 {
			pending.confirmations = bestHeight - pending.blockHeight + 1
		}
		// 如果达到要求的确认数，发送通知
		if !pending.notified && pending.confirmations >= b.requiredConfs {
			log.Debug("updateTransactionConfirmations ready for notification", "txHash", txHash.String(), "type", pending.txType,
				"confirmations", pending.confirmations, "required", b.requiredConfs)
			b.sendTransactionNotification(txHash, pending)
			pending.notified = true
		}

	}

}

func (b *btcWallet) addPendingTx(pending *btcPendingTx) {
	b.addPendingChan <- pending
}

func (b *btcWallet) removePendingTx(txHash chainhash.Hash) {
	b.removePendingChan <- txHash
}

// OpReturnData OP_RETURN数据结构
type opReturnData struct {
	protocol string // "rgbx"
	action   string // "deposit" | "withdraw"
	payload  string // chain33地址 或 交易哈希
}

// parseOpReturn 解析OP_RETURN数据
func (b *btcWallet) parseOpReturn(pkScript []byte) (*opReturnData, error) {
	// 提取数据（跳过OP_RETURN和长度字节, 解析格式: protocol:action:payload
	dataStr := string(pkScript[2:])
	parts := strings.Split(dataStr, ":")
	if len(parts) < 3 {
		log.Error("parseOpReturn invalid data", "data", dataStr)
		return nil, fmt.Errorf("invalid OP_RETURN format: expected 3 parts, got %d", len(parts))
	}

	opData := &opReturnData{
		protocol: parts[0],
		action:   parts[1],
		payload:  parts[2],
	}

	return opData, nil
}

// analyzeTransaction 分析交易类型（优化版本）
// 返回: ("deposit"|"withdraw"|"", pendingTx)
func (b *btcWallet) analyzeTransaction(hash *chainhash.Hash, tx *wire.MsgTx) *btcPendingTx {
	info := &btcPendingTx{}

	// 检查输出：查找TSS地址、非TSS地址的输出和OP_RETURN
	var hasTssOutput bool
	var depositAmount, withdrawAmount btcutil.Amount
	var firstNonTssOutputAddress string
	var parsed *opReturnData
	var err error

	for i, output := range tx.TxOut {
		// 检查并解析 OP_RETURN 输出
		if len(output.PkScript) > 2 && output.PkScript[0] == txscript.OP_RETURN && parsed == nil {
			parsed, err = b.parseOpReturn(output.PkScript)
			if err != nil {
				log.Error("analyzeTransaction parseOpReturn failed", "txHash", hash.String(),
					"outputIndex", i, "err", err, "pkScript", hex.EncodeToString(output.PkScript))
			} else {

				info.opReturnData = *parsed
				log.Debug("analyzeTransaction parseOpReturn success", "txHash", hash.String(),
					"protocol", parsed.protocol, "action", parsed.action, "payloadLen", len(parsed.payload))
			}
			continue
		}

		// 直接比较脚本（预计算的TSS地址脚本）
		if bytes.Equal(output.PkScript, b.tssPkScript) {
			hasTssOutput = true
			depositAmount += btcutil.Amount(output.Value)
			log.Debug("analyzeTransaction TSS output found", "txHash", hash.String(), "outputIndex", i, "amount", btcutil.Amount(output.Value))
			continue
		}

		withdrawAmount += btcutil.Amount(output.Value)
		if firstNonTssOutputAddress == "" {
			// 提取第一个非TSS地址输出（提现地址）
			_, addrs, _, err := txscript.ExtractPkScriptAddrs(output.PkScript, &b.chainParams)
			if err == nil && len(addrs) > 0 {
				firstNonTssOutputAddress = addrs[0].String()
				log.Debug("analyzeTransaction non-TSS output found", "txHash", hash.String(),
					"outputIndex", i, "address", firstNonTssOutputAddress, "amount", btcutil.Amount(output.Value))
			}
		}
	}

	hasTssInput := false
	// 检查输入：直接从witness解析公钥验证是否来自TSS地址
	if witness := tx.TxIn[0].Witness; len(witness) == 2 &&
		bytes.Equal(witness[1], b.tssPubKey.SerializeCompressed()) {
		hasTssInput = true
	}

	log.Debug("analyzeTransaction analysis result", "txHash", hash.String(),
		"hasTssInput", hasTssInput, "hasTssOutput", hasTssOutput,
		"depositAmount", depositAmount, "firstNonTssAddress", firstNonTssOutputAddress)

	// 根据规则判断交易类型
	// 提现交易特征：有TSS输入，有非TSS输出
	if hasTssInput && firstNonTssOutputAddress != "" {
		info.withdrawAddress = firstNonTssOutputAddress
		info.txType = transactionTypeWithdraw
		info.withdrawAmount = withdrawAmount
		info.chain33WithdrawTxHash = []byte(info.opReturnData.payload)
		return info
	} else if hasTssOutput && !hasTssInput { // 充值交易特征：有TSS输出，无TSS输入
		info.depositAmount = depositAmount
		info.txType = transactionTypeDeposit
		info.chain33DepositAddress = info.opReturnData.payload
		return info
	}

	// 不符合充值或提现特征，可能是小额合并交易
	log.Debug("analyzeTransaction not deposit/withdraw", "txHash", hash.String(),
		"hasTssInput", hasTssInput, "hasTssOutput", hasTssOutput)
	return nil
}

// sendTransactionNotification 发送交易确认通知
func (b *btcWallet) sendTransactionNotification(_ chainhash.Hash, pending *btcPendingTx) {
	if pending.txType == "deposit" {
		b.depositChan <- pending
	} else {
		b.withdrawChan <- pending
	}
}

// buildWithdrawTx 构建提现交易
// 返回: (交易, 输入金额列表, 已锁定的UTXO列表, 错误)
// 注意: 如果返回错误，UTXO锁定会自动释放；如果成功，调用方需要在广播失败时调用releaseUTXOs
func (b *btcWallet) buildWithdrawTx(req *withdrawRequest) (*wire.MsgTx, []int64, []*UTXO, error) {

	// 解析目标地址
	toAddr, err := btcutil.DecodeAddress(req.toAddress, &b.chainParams)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid to address: %w", err)
	}

	// 构建输出
	pkScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create pk script failed: %w", err)
	}

	outputs := []*wire.TxOut{
		{
			Value:    int64(req.amount),
			PkScript: pkScript,
		},
	}

	// 手动选择UTXO并构建交易
	tx, inputAmounts, lockedUTXOs, err := b.buildTransaction(outputs, req.feeRate, req.chain33WithDrawHash, true, req.stickyUTXO)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Debug("buildWithdrawTx", "to", req.toAddress, "amount", req.amount, "feeRate", req.feeRate,
		"inputs", len(tx.TxIn), "outputs", len(tx.TxOut), "lockedUTXOs", len(lockedUTXOs))

	return tx, inputAmounts, lockedUTXOs, nil
}

// buildTransaction 手动构建交易
// 返回: (交易, 输入金额列表, 选中的UTXO列表, 错误)
func (b *btcWallet) buildTransaction(outputs []*wire.TxOut, feeRate btcutil.Amount, chain33Hash []byte,
	receiverPaysFee bool, stickyUTXO *UTXO) (*wire.MsgTx, []int64, []*UTXO, error) {
	// 计算输出总额
	var outputTotal btcutil.Amount
	for _, output := range outputs {
		outputTotal += btcutil.Amount(output.Value)
	}

	hashStr := hex.EncodeToString(chain33Hash)
	// 获取可用UTXO
	utxos, err := b.listUnspent()
	if err != nil {
		log.Error("buildTransaction listUnspent failed", "hash", hashStr, "targetAmount", outputTotal,
			"utxos", len(utxos), "err", err)
		return nil, nil, nil, fmt.Errorf("list unspent failed: %w", err)
	}

	// 选择并锁定UTXO（防止并发双花）
	selectionFeeRate := feeRate
	if receiverPaysFee {
		selectionFeeRate = 0
	}
	selectedUTXOs, inputTotal, err := b.selectAndLockUTXOs(utxos, outputTotal, selectionFeeRate, stickyUTXO)
	if err != nil {
		log.Error("buildTransaction selectAndLockUTXOs failed", "hash", hashStr, "targetAmount", outputTotal,
			"err", err)
		return nil, nil, nil, err
	}

	// 创建交易
	tx := wire.NewMsgTx(wire.TxVersion)

	// 添加输入
	inputAmounts := make([]int64, 0, len(selectedUTXOs))
	for _, utxo := range selectedUTXOs {
		tx.AddTxIn(wire.NewTxIn(&utxo.OutPoint, nil, nil))
		inputAmounts = append(inputAmounts, int64(utxo.Amount))
	}

	buf := make([]byte, 0, withdrawOpReturnDataLen)
	buf = append(buf, []byte(withdrawOpReturnPrefix)...)
	buf = append(buf, chain33Hash...)
	opScript, err := txscript.NullDataScript(buf)
	if err != nil {
		log.Error("build tx op script failed", "hash", hashStr, "err", err)
		return nil, nil, nil, err
	}
	tx.AddTxOut(wire.NewTxOut(0, opScript))
	// 添加输出
	for _, output := range outputs {
		tx.AddTxOut(output)
	}

	// 计算实际手续费
	fee := estimateBtcFee(tx, feeRate)

	// 计算找零
	change := inputTotal - outputTotal
	if receiverPaysFee {
		// 提现输出金额过小，则不构建交易
		if fee+minChangeAmount >= outputTotal {
			b.releaseUTXOsExcept(selectedUTXOs, stickyUTXO)
			log.Error("buildTransaction withdraw amount too small for fee", "hash", hashStr, "targetAmount", outputTotal,
				"fee", fee, "change", change)
			return nil, nil, nil, fmt.Errorf("withdraw amount too small for fee: amount %d, fee %d", outputTotal, fee)
		}
		outputs[0].Value = int64(outputTotal - fee)
	} else {
		change = inputTotal - outputTotal - fee
	}

	// 注意：如果找零过小，会被用于交易费，由平台承担
	// 还有种方案是将这部分补贴给提现地址，减少用户提现成本
	if change > minChangeAmount {
		// 添加找零输出（使用预计算的脚本）
		tx.AddTxOut(wire.NewTxOut(int64(change), b.tssPkScript))
	} else if change < 0 {
		// 资金不足，释放已锁定的UTXO
		b.releaseUTXOsExcept(selectedUTXOs, stickyUTXO)
		log.Error("buildTransaction insufficient funds", "hash", hashStr, "targetAmount", outputTotal,
			"fee", fee, "change", change)
		return nil, nil, nil, fmt.Errorf("insufficient funds: need %d, have %d", outputTotal+fee, inputTotal)
	}

	log.Debug("buildTransaction success", "inputs", len(selectedUTXOs), "inputTotal", inputTotal,
		"outputTotal", outputTotal, "fee", fee, "change", change)

	return tx, inputAmounts, selectedUTXOs, nil
}

// listUnspent 获取可用UTXO
func (b *btcWallet) listUnspent() ([]*UTXO, error) {

	// 获取钱包中的未花费输出
	unspentOutputs, err := b.Wallet.ListUnspent(b.requiredConfs, math.MaxInt32, "")
	if err != nil {
		return nil, fmt.Errorf("list unspent from wallet failed: %w", err)
	}

	var utxos []*UTXO
	totalAmount := btcutil.Amount(0)
	for _, output := range unspentOutputs {
		// 解析OutPoint
		txHash, err := chainhash.NewHashFromStr(output.TxID)
		if err != nil {
			log.Error("listUnspent invalid txid", "txid", output.TxID, "err", err)
			continue
		}

		// 转换金额
		amount, err := btcutil.NewAmount(output.Amount)
		if err != nil {
			log.Error("listUnspent invalid amount", "amount", output.Amount, "err", err)
			continue
		}

		// 解析脚本
		pkScript, err := hex.DecodeString(output.ScriptPubKey)
		if err != nil {
			log.Error("listUnspent invalid script", "script", output.ScriptPubKey, "err", err)
			continue
		}
		if !bytes.Equal(pkScript, b.tssPkScript) {
			log.Debug("listUnspent skip non-tss utxo", "txid", output.TxID, "vout", output.Vout, "amount", amount)
			continue
		}

		utxo := &UTXO{
			OutPoint: wire.OutPoint{
				Hash:  *txHash,
				Index: output.Vout,
			},
			Amount:   amount,
			PkScript: pkScript,
		}
		utxos = append(utxos, utxo)
		totalAmount += amount
	}

	log.Debug("listUnspent", "count", len(utxos), "totalAmount", totalAmount.ToBTC())
	return utxos, nil
}

// selectUTXOs 找到最少数量的UTXO组合
// 策略：按金额从大到小排序，依次选择直到满足需求
func (b *btcWallet) selectUTXOs(utxos []*UTXO, targetAmount, feeRate btcutil.Amount, stickyCount int) ([]*UTXO, btcutil.Amount, error) {
	// 按金额从小到大排序
	sorted := make([]*UTXO, len(utxos))
	copy(sorted, utxos)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Amount < sorted[j].Amount
	})

	//从小到大，找第一个大于等于amount的UTXO索引，如果没有则返回最大的
	findFirstGreaterOrEqual := func(sorted []*UTXO, amount btcutil.Amount) int {

		for i := 0; i < len(sorted); i++ {
			if sorted[i].Amount >= amount {
				return i
			}
		}
		return len(sorted) - 1
	}

	var selected []*UTXO
	var total btcutil.Amount
	for len(sorted) > 0 {
		// 估算当前手续费
		txSize := b.estimateTxSize(len(selected)+stickyCount+1, 2, withdrawOpReturnScriptLen)
		fee := btcutil.Amount(txSize) * feeRate
		needed := targetAmount + fee - total
		idx := findFirstGreaterOrEqual(sorted, needed)
		selected = append(selected, sorted[idx])
		total += sorted[idx].Amount
		sorted = sorted[:idx]
		// 检查是否满足需求
		if targetAmount+fee <= total {
			log.Debug("selectUTXOs success", "selected", len(selected), "total", total,
				"target", targetAmount, "fee", fee, "waste", total-targetAmount-fee)
			return selected, total, nil
		}
	}

	log.Debug("selectUTXOs insufficient funds", "target", targetAmount, "total", total)
	return nil, 0, fmt.Errorf("insufficient funds: need %d, have %d", targetAmount, total)
}

// estimateTxSize 估算交易大小
// inputCount: 输入数量
// p2wpkhOutputCount: P2WPKH输出数量
// opReturnScriptSize: OP_RETURN脚本长度（字节）
// 返回: 交易大小（字节）
func (b *btcWallet) estimateTxSize(inputCount, p2wpkhOutputCount, opReturnScriptSize int) int {
	// 基本交易大小
	baseSize := 10 // version(4) + locktime(4) + input_count(1) + output_count(1)

	// P2WPKH输入大小
	// - OutPoint: 36字节 (txid:32 + index:4)
	// - ScriptSig: 1字节 (空)
	// - Sequence: 4字节
	// - Witness: ~108字节 (signature:72 + pubkey:33 + witness_count:1 + lengths:2)
	inputSize := inputCount * (36 + 1 + 4 + 108)

	// P2WPKH输出大小
	// - Value: 8字节
	// - ScriptPubKey: 23字节 (OP_0 + 20字节pubkey hash)
	outputSize := p2wpkhOutputCount * (8 + 23)
	if opReturnScriptSize > 0 {
		outputSize += 8 + 1 + opReturnScriptSize // value + script len + script
	}

	return baseSize + inputSize + outputSize
}

// selectAndLockUTXOs 选择并锁定UTXO（用于提现交易）
// 使用wallet.LeaseOutput实现持久化锁定，防止UTXO被重复使用
func (b *btcWallet) selectAndLockUTXOs(utxos []*UTXO, targetAmount, feeRate btcutil.Amount, stickyUTXO *UTXO) ([]*UTXO, btcutil.Amount, error) {

	log.Debug("selectAndLockUTXOs", "total", len(utxos), "targetAmount", targetAmount,
		"feeRate", feeRate, "stickyUTXO", stickyUTXO != nil)

	stickyCount := 0
	if stickyUTXO != nil {
		targetAmount -= stickyUTXO.Amount
		// 如果剩余所需金额为0，且没有其他UTXO，则直接返回指定输入
		if len(utxos) == 0 && targetAmount <= 0 {
			return []*UTXO{stickyUTXO}, stickyUTXO.Amount, nil
		}
		stickyCount = 1
	}

	selected, total, err := b.selectUTXOs(utxos, targetAmount, feeRate, stickyCount)
	if err != nil {
		log.Error("selectAndLockUTXOs selectUTXOs failed", "utxos", len(utxos), "targetAmount", targetAmount,
			"feeRate", feeRate, "err", err)
		return nil, 0, err
	}
	// 如果存在指定输入，则将其添加到选中的UTXO列表中，并累加金额
	if stickyUTXO != nil {
		selected = append(selected, stickyUTXO)
		total += stickyUTXO.Amount
	}

	// 锁定选中的UTXO
	lockedUTXOs := make([]*UTXO, 0, len(selected))
	for _, utxo := range selected {
		// 使用wallet.LeaseOutput锁定UTXO
		expiry, err := b.Wallet.LeaseOutput(utxoLockID, utxo.OutPoint, utxoLeaseDuration)
		if err != nil {
			log.Error("selectAndLockUTXOs lease output failed", "outpoint", utxo.OutPoint.String(), "err", err)
			// 锁定失败，释放已锁定的UTXO
			b.releaseUTXOsExcept(lockedUTXOs, stickyUTXO)
			return nil, 0, fmt.Errorf("lease output failed: %w", err)
		}

		lockedUTXOs = append(lockedUTXOs, utxo)
		log.Debug("selectAndLockUTXOs locked UTXO", "outpoint", utxo.OutPoint.String(), "amount", utxo.Amount, "expiry", expiry)
	}

	log.Debug("selectAndLockUTXOs success", "utxos", len(utxos), "selected", len(lockedUTXOs),
		"totalAmount", total, "targetAmount", targetAmount, "feeRate", feeRate)

	return lockedUTXOs, total, nil
}

// releaseUTXOsExcept 释放UTXO锁定，可选保留指定输入不释放
func (b *btcWallet) releaseUTXOsExcept(utxos []*UTXO, keep *UTXO) {
	if len(utxos) == 0 {
		return
	}

	log.Debug("releaseUTXOs start", "count", len(utxos))

	for _, utxo := range utxos {
		if keep != nil && utxo.OutPoint == keep.OutPoint {
			log.Debug("releaseUTXOsExcept keep sticky UTXO", "outpoint", utxo.OutPoint.String(), "amount", utxo.Amount)
			continue
		}
		err := b.Wallet.ReleaseOutput(utxoLockID, utxo.OutPoint)
		if err != nil {
			log.Error("releaseUTXOs release output failed",
				"outpoint", utxo.OutPoint.String(),
				"err", err)
		}
	}
}

// UTXO 结构
type UTXO struct {
	OutPoint wire.OutPoint
	Amount   btcutil.Amount
	PkScript []byte
}

// broadcastTransaction 广播交易
// lockedUTXOs: 已锁定的UTXO列表，广播失败时会自动释放
func (b *btcWallet) broadcastTransaction(tx *wire.MsgTx, btcTxHash string) error {

	if b.rpcClient != nil {
		_, err := b.rpcClient.SendRawTransaction(tx, false)
		if err != nil {
			log.Error("BroadcastTransaction failed", "txHash", btcTxHash, "err", err)
			return err
		}
	} else {
		_, err := b.chainClient.SendRawTransaction(tx, false)
		if err != nil {
			log.Error("BroadcastTransaction failed", "txHash", btcTxHash, "err", err)
			return err
		}
	}
	log.Debug("broadcastTransaction success", "txHash", btcTxHash)
	return nil
}

func buildBtcSpv(txHash, blockHash string, blockTime int64, blockHeight uint64, txs [][]byte, txIndex uint32) *lighttypes.BtcSpv {
	return &lighttypes.BtcSpv{
		TxHash:      txHash,
		Time:        blockTime,
		Height:      blockHeight,
		BlockHash:   blockHash,
		TxIndex:     txIndex,
		BranchProof: merkle.GetMerkleBranch(txs, txIndex),
	}
}

func buildTxHashesFromVerbose(txIDs []string, targetTxHash string) ([][]byte, uint32, error) {
	txs := make([][]byte, 0, len(txIDs))
	txIndex := uint32(0)
	found := false
	for idx, txID := range txIDs {
		hash, err := chainhash.NewHashFromStr(txID)
		if err != nil {
			return nil, 0, err
		}
		txs = append(txs, hash.CloneBytes())
		if txID == targetTxHash {
			txIndex = uint32(idx)
			found = true
		}
	}
	if !found {
		return nil, 0, fmt.Errorf("tx not found in block")
	}
	return txs, txIndex, nil
}

func buildTxHashesFromBlockTxs(blockTxs []*wire.MsgTx, targetTxHash chainhash.Hash) ([][]byte, uint32, error) {
	txs := make([][]byte, 0, len(blockTxs))
	txIndex := uint32(0)
	found := false
	for idx, tx := range blockTxs {
		hash := tx.TxHash()
		txs = append(txs, hash.CloneBytes())
		if hash == targetTxHash {
			txIndex = uint32(idx)
			found = true
		}
	}
	if !found {
		return nil, 0, fmt.Errorf("tx not found in block")
	}
	return txs, txIndex, nil
}

// buildTxExistenceProof 计算交易存在性证明（SPV）
// 输入: pendingTx（需要包含 tx 和 blockHash）
// 输出: lighttypes.BtcSpv
func (b *btcWallet) buildTxExistenceProof(pending *btcPendingTx) (*lighttypes.BtcSpv, error) {
	if pending == nil || pending.tx == nil {
		return nil, fmt.Errorf("pending tx data missing")
	}
	if pending.blockHeight <= 0 {
		return nil, fmt.Errorf("invalid pending block height")
	}
	txHashStr := pending.tx.TxHash().String()
	if b.rpcClient != nil {
		block, err := b.rpcClient.GetBlockVerbose(&pending.blockHash)
		if err != nil {
			log.Error("buildTxExistenceProof GetBlockVerbose failed, fallback neutrino",
				"txHash", txHashStr, "blockHash", pending.blockHash.String(), "err", err)
		} else {
			txs, txIndex, err := buildTxHashesFromVerbose(block.Tx, txHashStr)
			if err != nil {
				log.Error("buildTxExistenceProof buildTxHashesFromVerbose failed",
					"txHash", txHashStr, "blockHash", pending.blockHash.String(), "err", err)
				return nil, err
			}
			return buildBtcSpv(
				txHashStr, pending.blockHash.String(), block.Time, uint64(block.Height), txs, txIndex,
			), nil
		}
	}

	block, err := b.chainClient.GetBlock(&pending.blockHash)
	if err != nil {
		log.Error("buildTxExistenceProof GetBlock failed",
			"txHash", txHashStr, "blockHash", pending.blockHash.String(), "err", err)
		return nil, err
	}

	txs, txIndex, err := buildTxHashesFromBlockTxs(block.Transactions, pending.txHash)
	if err != nil {
		log.Error("buildTxExistenceProof buildTxHashesFromBlockTxs failed",
			"txHash", txHashStr, "blockHash", pending.blockHash.String(), "err", err)
		return nil, err
	}
	return buildBtcSpv(
		txHashStr, pending.blockHash.String(), block.Header.Timestamp.Unix(), uint64(pending.blockHeight), txs, txIndex,
	), nil
}

// GetBalance 获取余额
func (b *btcWallet) getBalance() (btcutil.Amount, error) {

	// 获取已确认余额
	balance, err := b.Wallet.CalculateBalance(b.requiredConfs)
	if err != nil {
		return 0, fmt.Errorf("calculate balance failed: %w", err)
	}

	return balance, nil
}

// GetDepositChannel 获取充值通知channel
func (b *btcWallet) GetDepositChannel() <-chan *btcPendingTx {
	return b.depositChan
}

// GetWithdrawChannel 获取提现通知channel
func (b *btcWallet) GetWithdrawChannel() <-chan *btcPendingTx {
	return b.withdrawChan
}
