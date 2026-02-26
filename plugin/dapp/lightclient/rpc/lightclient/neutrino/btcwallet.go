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
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/chain"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/btcsuite/btcwallet/wtxmgr"
)

const (
	transactionTypeDeposit      string = "deposit"
	transactionTypeWithdraw     string = "withdraw"
	transactionTypeMergeBalance string = "mergeBalance"
)

const (
	// 默认确认数
	defaultRequiredConfs = 6
	// 最小找零金额（粉尘限制）
	minChangeAmount = 546
	// 默认手续费率 sat/byte
	defaultFeeRate = 20
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
var utxoWithdrawLockID = wtxmgr.LockID{
	'R', 'G', 'B', 'X', '-', 'W', 'I', 'T', 'H',
	'D', 'R', 'A', 'W', '-', 'L', 'O', 'C', 'K',
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
}

type btcWallet struct {
	*wallet.Wallet
	client      *neutrinoClient
	chainParams chaincfg.Params
	chainClient chain.Interface
	db          walletdb.DB

	monitorStartHeight int32

	// TSS相关
	tssAddress  btcutil.Address
	tssPubKey   *btcec.PublicKey
	tssPkScript []byte // 预计算的TSS地址脚本

	// 通知channel
	depositChan    chan *pendingTx
	withdrawChan   chan *pendingTx
	addPendingChan chan *pendingTx

	// 配置
	requiredConfs int32
	feeRate       btcutil.Amount

	// 交易监控
	pendingTxs map[chainhash.Hash]*pendingTx
	// txLock     sync.RWMutex
}

type pendingTx struct {
	tx                    *wire.MsgTx
	submitTime            time.Time
	confirmations         int32
	blockHeight           int32
	blockHash             chainhash.Hash
	txHash                chainhash.Hash
	txType                string // "deposit" or "withdraw"
	depositAmount         btcutil.Amount
	withdrawAmount        btcutil.Amount
	chain33DepositAddress string // Chain33充值地址
	withdrawAddress       string
	chain33WithDrawHash   string // Chain33提现交易哈希
	// OP_RETURN数据
	opReturnData opReturnData // 原始OP_RETURN解析数据
}

func newBtcWallet(n *neutrinoClient) (*btcWallet, error) {
	bw := &btcWallet{
		client:         n,
		chainParams:    n.neutrinoCfg.ChainParams,
		depositChan:    make(chan *pendingTx, 100),
		withdrawChan:   make(chan *pendingTx, 100),
		addPendingChan: make(chan *pendingTx, 100),
		requiredConfs:  defaultRequiredConfs,
		feeRate:        defaultFeeRate,
		pendingTxs:     make(map[chainhash.Hash]*pendingTx),
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

	w, err := wallet.Open(db, pubPass, nil, &bw.chainParams, 250)
	if err != nil {
		log.Error("newBtcWallet open wallet error", "err", err)
		_ = db.Close()
		return nil, err
	}

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
	b.Wallet.SynchronizeRPC(b.chainClient)

	// 等待TSS地址生成
	go b.waitAndImportTSSAddress()

	// 启动交易监听
	go b.monitorTransactions()

	return nil
}

func (b *btcWallet) stop() {
	b.Wallet.Stop()
	b.chainClient.Stop()
	_ = b.db.Close()
	close(b.depositChan)
	close(b.withdrawChan)
}

// waitAndImportTSSAddress 等待TSS地址生成并导入
func (b *btcWallet) waitAndImportTSSAddress() {
	for {
		if !b.client.tss.isDKGCompleted() {

			time.Sleep(time.Second * 3)
			continue
		}
		addr := b.client.tss.getTssAddress()

		// 保存TSS地址、公钥和脚本
		b.tssAddress = addr
		b.tssPubKey = b.client.tss.tssPublicKey
		b.tssPkScript = b.client.tss.pkScript

		// 显式导入TSS公钥到钱包
		if _, err := b.Wallet.AddressInfo(addr); err != nil {
			if !waddrmgr.IsError(err, waddrmgr.ErrAddressNotFound) {
				log.Error("waitAndImportTSSAddress AddressInfo failed", "err", err)
				time.Sleep(time.Second * 3)
				continue
			}
			err = b.Wallet.ImportPublicKey(b.tssPubKey, waddrmgr.WitnessPubKey)
			if err != nil {
				log.Error("waitAndImportTSSAddress ImportPublicKey failed", "err", err)
				time.Sleep(time.Second * 3)
				continue
			}
		}

		log.Info("waitAndImportTSSAddress success", "address", addr.String())
		return

	}
}

func (b *btcWallet) loadMinPendingHeight() int32 {
	var height int32
	err := walletdb.View(b.db, func(tx walletdb.ReadTx) error {
		bucket := tx.ReadBucket([]byte(btcwalletMonitorBucket))
		if bucket == nil {
			return walletdb.ErrBucketNotFound
		}
		val := bucket.Get([]byte(minPendingHeightKey))
		if val == nil {
			return types.ErrNotFound
		}
		reply := &types.Int64{}
		if err := types.Decode(val, reply); err != nil {
			return err
		}
		height = int32(reply.GetData())
		return nil
	})
	if err != nil && !errors.Is(err, walletdb.ErrBucketNotFound) && !errors.Is(err, types.ErrNotFound) {
		log.Error("loadMinPendingHeight", "err", err)
	}
	return height
}

func (b *btcWallet) saveMinPendingHeight(height int32) {
	err := walletdb.Update(b.db, func(tx walletdb.ReadWriteTx) error {
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
	if err != nil && !errors.Is(err, walletdb.ErrBucketNotFound) {
		log.Error("saveMinPendingHeight", "err", err, "height", height)
	}
}

func (b *btcWallet) updateMinPendingHeightLocked() {
	minHeight := int32(0)
	for _, pending := range b.pendingTxs {
		if pending.blockHeight <= 0 {
			continue
		}
		if minHeight == 0 || pending.blockHeight < minHeight {
			minHeight = pending.blockHeight
		}
	}
	if b.monitorStartHeight < minHeight {
		b.monitorStartHeight = minHeight
		b.saveMinPendingHeight(minHeight)
	}
}

// rescanFromHeight 从指定高度开始重新扫描
func (b *btcWallet) rescanFromHeight(height int32) error {
	if height <= 0 {
		return nil
	}
	if b.tssAddress == nil {
		return fmt.Errorf("TSS address not ready")
	}
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
		return err
	case <-b.client.ctx.Done():
		return types.ErrChannelClosed
	}
}

// addPendingTx 添加待确认交易
func (b *btcWallet) addPendingTx(pending *pendingTx) {
	b.addPendingChan <- pending
}

// monitorTransactions 监听交易通知
func (b *btcWallet) monitorTransactions() {
	for {
		if b.client.tss.isDKGCompleted() {
			break
		}
		time.Sleep(time.Second * 3)
	}

	b.monitorStartHeight = b.loadMinPendingHeight()

	// 注册通知客户端
	client := b.Wallet.NtfnServer.TransactionNotifications()
	if b.monitorStartHeight > 0 {
		log.Debug("monitorTransactions resume from height", "height", b.monitorStartHeight)
		if err := b.rescanFromHeight(b.monitorStartHeight); err != nil {
			log.Error("monitorTransactions rescan", "height", b.monitorStartHeight, "err", err)
		}
	}
	interval := b.client.cfg.BtcBlockInterval/2 + 1
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	for {
		select {
		case <-b.client.ctx.Done():
			client.Done()
			return

		case ntfn := <-client.C:
			if ntfn == nil {
				continue
			}

			// 处理已确认交易
			for _, block := range ntfn.AttachedBlocks {
				for _, tx := range block.Transactions {
					b.handleTransaction(&tx, block.Height, *block.Hash)
				}
			}

			// 处理未确认交易（重置确认数）
			for _, tx := range ntfn.UnminedTransactions {
				b.handleUnminedTransaction(*tx.Hash)
			}

		case pending := <-b.addPendingChan:
			b.pendingTxs[pending.txHash] = pending
			log.Debug("addPendingTx", "txHash", pending.txHash.String(), "blockHeight", pending.blockHeight,
				"blockHash", pending.blockHash.String(), "txType", pending.txType)

		case <-ticker.C:
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

		log.Debug("handleUnminedTransaction reset confirmations", "txHash", txHash.String(),
			"oldConfirmations", oldConfirmations, "oldBlockHeight", oldBlockHeight, "type", pending.txType)
	}
}

// updateTransactionConfirmations 更新已存在交易的确认数
func (b *btcWallet) updateTransactionConfirmations() {
	bestBlock := b.client.getBestBlock()
	var readyHashes []chainhash.Hash

	for txHash, pending := range b.pendingTxs {

		txRes, err := b.Wallet.GetTransaction(txHash)
		if err != nil {
			log.Error("updateTransactionConfirmation getTransaction error", "txHash", txHash.String())
			if pending.blockHeight > 0 {
				pending.confirmations = bestBlock.Height - pending.blockHeight
			}
		} else {
			pending.confirmations = txRes.Confirmations
		}
		// 如果达到要求的确认数，发送通知
		if pending.confirmations >= b.requiredConfs {
			log.Debug("updateTransactionConfirmations ready for notification", "txHash", txHash.String(), "type", pending.txType,
				"confirmations", pending.confirmations, "required", b.requiredConfs)
			b.sendTransactionNotification(txHash, pending)
			readyHashes = append(readyHashes, txHash)
		}

	}

	for _, txHash := range readyHashes {
		delete(b.pendingTxs, txHash)
	}
	if len(readyHashes) > 0 {
		b.updateMinPendingHeightLocked()
	}
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

	if opData.action == transactionTypeWithdraw { // chain33 tx hash
		opData.payload = hex.EncodeToString([]byte(parts[2]))
	}

	return opData, nil
}

// analyzeTransaction 分析交易类型（优化版本）
// 返回: ("deposit"|"withdraw"|"", pendingTx)
func (b *btcWallet) analyzeTransaction(hash *chainhash.Hash, tx *wire.MsgTx) *pendingTx {
	info := &pendingTx{}

	// 检查输出：查找TSS地址、非TSS地址的输出和OP_RETURN
	var hasTssOutput bool
	var depositAmount, withdrawAmount btcutil.Amount
	var firstNonTssOutputAddress string
	var parsed *opReturnData
	var err error

	for i, output := range tx.TxOut {
		// 检查并解析 OP_RETURN 输出
		if output.PkScript[0] == txscript.OP_RETURN && parsed == nil && len(output.PkScript) > 2 {
			parsed, err = b.parseOpReturn(output.PkScript)
			if err != nil {
				log.Debug("analyzeTransaction parseOpReturn failed", "txHash", hash.String(),
					"outputIndex", i, "err", err)
			} else {

				info.opReturnData = *parsed
				log.Info("analyzeTransaction parseOpReturn success", "txHash", hash.String(),
					"protocol", parsed.protocol, "action", parsed.action, "payload", parsed.payload)
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
		info.chain33WithDrawHash = parsed.payload
		return info
	} else if hasTssOutput && !hasTssInput { // 充值交易特征：有TSS输出，无TSS输入
		info.depositAmount = depositAmount
		info.txType = transactionTypeDeposit
		info.chain33DepositAddress = parsed.payload
		return info
	}

	// 不符合充值或提现特征，可能是小额合并交易
	log.Debug("analyzeTransaction not deposit/withdraw", "txHash", hash.String(),
		"hasTssInput", hasTssInput, "hasTssOutput", hasTssOutput)
	return nil
}

// sendTransactionNotification 发送交易确认通知
func (b *btcWallet) sendTransactionNotification(txHash chainhash.Hash, pending *pendingTx) {
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
	if b.tssAddress == nil {
		return nil, nil, nil, fmt.Errorf("TSS address not ready")
	}

	// 解析目标地址
	toAddr, err := btcutil.DecodeAddress(req.toAddress, &b.chainParams)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid to address: %w", err)
	}

	// 获取手续费率
	feeRate := req.feeRate
	if feeRate == 0 {
		feeRate = b.feeRate
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
	tx, inputAmounts, lockedUTXOs, err := b.buildTransaction(outputs, feeRate, req.chain33WithDrawHash, true)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Debug("buildWithdrawTx", "to", req.toAddress, "amount", req.amount, "feeRate", feeRate,
		"inputs", len(tx.TxIn), "outputs", len(tx.TxOut), "lockedUTXOs", len(lockedUTXOs))

	return tx, inputAmounts, lockedUTXOs, nil
}

// buildTransaction 手动构建交易
// 返回: (交易, 输入金额列表, 选中的UTXO列表, 错误)
func (b *btcWallet) buildTransaction(outputs []*wire.TxOut, feeRate btcutil.Amount, chain33Hash []byte, receiverPaysFee bool) (*wire.MsgTx, []int64, []*UTXO, error) {
	// 计算输出总额
	var outputTotal btcutil.Amount
	for _, output := range outputs {
		outputTotal += btcutil.Amount(output.Value)
	}

	// 获取可用UTXO
	utxos, err := b.listUnspent()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("list unspent failed: %w", err)
	}

	if len(utxos) == 0 {
		return nil, nil, nil, fmt.Errorf("no available UTXO")
	}

	// 选择并锁定UTXO（防止并发双花）
	selectionFeeRate := feeRate
	if receiverPaysFee {
		selectionFeeRate = 0
	}
	selectedUTXOs, inputTotal, err := b.selectAndLockUTXOs(utxos, outputTotal, selectionFeeRate)
	if err != nil {
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
		log.Error("build tx op script failed", "err", err)
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
			b.releaseUTXOs(selectedUTXOs)
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
		b.releaseUTXOs(selectedUTXOs)
		return nil, nil, nil, fmt.Errorf("insufficient funds: need %d, have %d", outputTotal+fee, inputTotal)
	}

	log.Debug("buildTransaction success", "inputs", len(selectedUTXOs), "inputTotal", inputTotal,
		"outputTotal", outputTotal, "fee", fee, "change", change)

	return tx, inputAmounts, selectedUTXOs, nil
}

// listUnspent 获取可用UTXO
func (b *btcWallet) listUnspent() ([]*UTXO, error) {
	if b.tssAddress == nil {
		return nil, fmt.Errorf("TSS address not ready")
	}

	// 获取钱包中的未花费输出
	unspentOutputs, err := b.Wallet.ListUnspent(b.requiredConfs, math.MaxInt32, b.tssAddress.String())
	if err != nil {
		return nil, fmt.Errorf("list unspent from wallet failed: %w", err)
	}

	var utxos []*UTXO
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

		utxo := &UTXO{
			OutPoint: wire.OutPoint{
				Hash:  *txHash,
				Index: output.Vout,
			},
			Amount:   amount,
			PkScript: pkScript,
		}
		utxos = append(utxos, utxo)
	}

	// 按金额从大到小排序，便于选择
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Amount > utxos[j].Amount
	})

	log.Debug("listUnspent", "count", len(utxos), "totalAmount", b.calculateTotalAmount(utxos))
	return utxos, nil
}

// calculateTotalAmount 计算UTXO总金额
func (b *btcWallet) calculateTotalAmount(utxos []*UTXO) btcutil.Amount {
	var total btcutil.Amount
	for _, utxo := range utxos {
		total += utxo.Amount
	}
	return total
}

// selectUTXOs 选择UTXO（优化版本）
// 策略：
// - 如果最少数量 > 4，直接返回最少数量组合
// - 如果最少数量 ≤ 4，尝试优化选择，减少交易大小
func (b *btcWallet) selectUTXOs(utxos []*UTXO, targetAmount, feeRate btcutil.Amount) ([]*UTXO, btcutil.Amount, error) {
	if len(utxos) == 0 {
		return nil, 0, fmt.Errorf("no available UTXO")
	}

	// 第一步：找到最少数量的UTXO组合
	minResult, err := b.findMinimumInputCount(utxos, targetAmount, feeRate)
	if err != nil {
		return nil, 0, err
	}

	log.Debug("selectUTXOs minimum input count", "count", len(minResult.selected), "total", minResult.total, "target", targetAmount)

	// 如果最少数量 > 4，直接返回
	if len(minResult.selected) > 4 {
		log.Info("selectUTXOs using minimum count strategy",
			"selected", len(minResult.selected),
			"total", minResult.total)
		return minResult.selected, minResult.total, nil
	}

	// 第二步：尝试优化选择（数量 ≤ 4）
	optimizedResult, err := b.optimizeSelection(utxos, targetAmount, feeRate, len(minResult.selected))
	if err != nil {
		// 优化失败，使用最少数量策略
		log.Debug("selectUTXOs optimization failed, using minimum count",
			"err", err)
		return minResult.selected, minResult.total, nil
	}

	log.Info("selectUTXOs using optimized strategy", "selected", len(optimizedResult.selected),
		"total", optimizedResult.total, "saved", minResult.total-optimizedResult.total)

	return optimizedResult.selected, optimizedResult.total, nil
}

// selectionResult UTXO选择结果
type selectionResult struct {
	selected []*UTXO        // 选中的UTXO列表
	total    btcutil.Amount // 总金额
	waste    btcutil.Amount // 浪费金额（找零+手续费差异）
}

// findMinimumInputCount 找到最少数量的UTXO组合
// 策略：按金额从大到小排序，依次选择直到满足需求
func (b *btcWallet) findMinimumInputCount(utxos []*UTXO, targetAmount, feeRate btcutil.Amount) (*selectionResult, error) {
	// 按金额从大到小排序
	sorted := make([]*UTXO, len(utxos))
	copy(sorted, utxos)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Amount > sorted[j].Amount
	})

	var selected []*UTXO
	var total btcutil.Amount

	for _, utxo := range sorted {
		selected = append(selected, utxo)
		total += utxo.Amount

		// 估算手续费
		txSize := b.estimateTxSize(len(selected), 2, withdrawOpReturnScriptLen) // 目标+找零+OP_RETURN
		fee := btcutil.Amount(txSize) * feeRate

		if total >= targetAmount+fee {
			return &selectionResult{
				selected: selected,
				total:    total,
				waste:    total - targetAmount - fee,
			}, nil
		}
	}

	return nil, fmt.Errorf("insufficient funds: need %d, have %d", targetAmount, total)
}

// optimizeSelection 优化UTXO选择（数量 ≤ 4）
// 策略：在最少数量限制内，依次选择最小大于剩余所需的UTXO
func (b *btcWallet) optimizeSelection(utxos []*UTXO, targetAmount, feeRate btcutil.Amount, maxCount int) (*selectionResult, error) {
	// 按金额从小到大排序
	sorted := make([]*UTXO, len(utxos))
	copy(sorted, utxos)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Amount < sorted[j].Amount
	})

	var selected []*UTXO
	var total btcutil.Amount
	remaining := targetAmount

	for len(selected) < maxCount {
		// 估算当前手续费
		txSize := b.estimateTxSize(len(selected)+1, 2, withdrawOpReturnScriptLen)
		fee := btcutil.Amount(txSize) * feeRate
		needed := remaining + fee

		// 找到最小大于所需金额的UTXO
		utxo := b.findSmallestGreaterThan(sorted, selected, needed)
		if utxo == nil {
			// 没找到合适的，选择最大的
			utxo = b.findLargestUnselected(sorted, selected)
			if utxo == nil {
				break
			}
		}

		selected = append(selected, utxo)
		total += utxo.Amount
		remaining = targetAmount + fee - total

		// 检查是否满足需求
		if remaining <= 0 {
			return &selectionResult{
				selected: selected,
				total:    total,
				waste:    -remaining,
			}, nil
		}
	}

	// 未能在限制内满足需求
	return nil, fmt.Errorf("cannot satisfy with %d inputs", maxCount)
}

// findSmallestGreaterThan 找到最小大于指定金额的UTXO
func (b *btcWallet) findSmallestGreaterThan(utxos []*UTXO, selected []*UTXO, amount btcutil.Amount) *UTXO {
	// 创建已选择UTXO的map，用于快速查找
	selectedMap := make(map[wire.OutPoint]bool)
	for _, utxo := range selected {
		selectedMap[utxo.OutPoint] = true
	}

	// 查找最小大于amount的UTXO
	for _, utxo := range utxos {
		if selectedMap[utxo.OutPoint] {
			continue
		}
		if utxo.Amount >= amount {
			return utxo
		}
	}

	return nil
}

// findLargestUnselected 找到最大的未选择UTXO
func (b *btcWallet) findLargestUnselected(utxos []*UTXO, selected []*UTXO) *UTXO {
	// 创建已选择UTXO的map
	selectedMap := make(map[wire.OutPoint]bool)
	for _, utxo := range selected {
		selectedMap[utxo.OutPoint] = true
	}

	// 从后往前找（因为已按从小到大排序）
	for i := len(utxos) - 1; i >= 0; i-- {
		if !selectedMap[utxos[i].OutPoint] {
			return utxos[i]
		}
	}

	return nil
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
func (b *btcWallet) selectAndLockUTXOs(utxos []*UTXO, targetAmount, feeRate btcutil.Amount) ([]*UTXO, btcutil.Amount, error) {
	// 过滤已锁定的UTXO
	var availableUTXOs []*UTXO
	for _, utxo := range utxos {
		// 检查UTXO是否已被锁定
		locked := b.Wallet.LockedOutpoint(utxo.OutPoint)
		if !locked {
			availableUTXOs = append(availableUTXOs, utxo)
		} else {
			log.Debug("selectAndLockUTXOs skip locked UTXO",
				"outpoint", utxo.OutPoint.String(),
				"amount", utxo.Amount)
		}
	}

	if len(availableUTXOs) == 0 {
		return nil, 0, fmt.Errorf("no available unlocked UTXO")
	}

	log.Debug("selectAndLockUTXOs available UTXOs",
		"total", len(utxos),
		"available", len(availableUTXOs),
		"locked", len(utxos)-len(availableUTXOs))

	// 使用现有的选择策略
	selected, total, err := b.selectUTXOs(availableUTXOs, targetAmount, feeRate)
	if err != nil {
		return nil, 0, err
	}

	// 锁定选中的UTXO
	var lockedUTXOs []*UTXO
	for _, utxo := range selected {
		// 使用wallet.LeaseOutput锁定UTXO
		expiry, err := b.Wallet.LeaseOutput(utxoWithdrawLockID, utxo.OutPoint, utxoLeaseDuration)
		if err != nil {
			log.Error("selectAndLockUTXOs lease output failed",
				"outpoint", utxo.OutPoint.String(),
				"err", err)
			// 锁定失败，释放已锁定的UTXO
			b.releaseUTXOs(lockedUTXOs)
			return nil, 0, fmt.Errorf("lease output failed: %w", err)
		}

		lockedUTXOs = append(lockedUTXOs, utxo)
		log.Debug("selectAndLockUTXOs locked UTXO", "outpoint", utxo.OutPoint.String(), "amount", utxo.Amount, "expiry", expiry)
	}

	log.Debug("selectAndLockUTXOs success", "selected", len(lockedUTXOs), "totalAmount", total, "targetAmount", targetAmount, "feeRate", feeRate)

	return lockedUTXOs, total, nil
}

// releaseUTXOs 释放UTXO锁定
// 在交易构建失败或广播失败时调用，释放已锁定的UTXO
func (b *btcWallet) releaseUTXOs(utxos []*UTXO) {
	if len(utxos) == 0 {
		return
	}

	log.Debug("releaseUTXOs start", "count", len(utxos))

	for _, utxo := range utxos {
		err := b.Wallet.ReleaseOutput(utxoWithdrawLockID, utxo.OutPoint)
		if err != nil {
			log.Error("releaseUTXOs release output failed",
				"outpoint", utxo.OutPoint.String(),
				"err", err)
		} else {
			log.Debug("releaseUTXOs released UTXO",
				"outpoint", utxo.OutPoint.String(),
				"amount", utxo.Amount)
		}
	}

	log.Info("releaseUTXOs completed", "count", len(utxos))
}

// UTXO 结构
type UTXO struct {
	OutPoint wire.OutPoint
	Amount   btcutil.Amount
	PkScript []byte
}

// broadcastTransaction 广播交易
// lockedUTXOs: 已锁定的UTXO列表，广播失败时会自动释放
func (b *btcWallet) broadcastTransaction(tx *wire.MsgTx, toAddress string, lockedUTXOs []*UTXO) error {
	// 计算交易哈希
	txHash := tx.TxHash()

	log.Debug("BroadcastTransaction start", "txHash", txHash.String(), "toAddress", toAddress,
		"inputs", len(tx.TxIn), "outputs", len(tx.TxOut), "lockedUTXOs", len(lockedUTXOs))

	// 广播交易
	_, err := b.chainClient.SendRawTransaction(tx, false)
	if err != nil {
		log.Error("BroadcastTransaction failed", "txHash", txHash.String(), "toAddress", toAddress, "err", err)

		// 广播失败，释放已锁定的UTXO
		b.releaseUTXOs(lockedUTXOs)
		return fmt.Errorf("broadcast transaction failed: %w", err)
	}

	// 记录待确认交易

	// b.pendingTxs[txHash] = &pendingTx{
	// 	tx:              tx,
	// 	submitTime:      types.Now(),
	// 	confirmations:   0,
	// 	blockHeight:     -1,
	// 	txType:          "withdraw",
	// 	withdrawAddress: toAddress,
	// 	txHash:          txHash,
	// }

	log.Debug("broadcastTransaction success", "txHash", txHash.String())

	// 注意: 广播成功后，UTXO锁定会在交易确认后自动释放（由btcwallet管理）
	return nil
}

// buildTxExistenceProof 计算交易存在性证明（SPV）
// 输入: pendingTx（需要包含 tx 和 blockHash）
// 输出: lighttypes.BtcSpv
func (b *btcWallet) buildTxExistenceProof(pending *pendingTx) (*lighttypes.BtcSpv, error) {
	if pending == nil || pending.tx == nil {
		return nil, fmt.Errorf("pending tx data missing")
	}
	if pending.blockHeight <= 0 {
		return nil, fmt.Errorf("invalid pending block height")
	}
	txHashStr := pending.tx.TxHash().String()
	block, err := b.chainClient.GetBlock(&pending.blockHash)
	if err != nil {
		return nil, err
	}

	var txIndex uint32
	found := false
	txs := make([][]byte, 0, len(block.Transactions))
	for index, tx := range block.Transactions {
		hash := tx.TxHash()
		txs = append(txs, hash.CloneBytes())
		if hash == pending.txHash {
			txIndex = uint32(index)
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("tx not found in block")
	}

	proof := merkle.GetMerkleBranch(txs, txIndex)
	spv := &lighttypes.BtcSpv{
		TxHash:      txHashStr,
		Time:        block.Header.Timestamp.Unix(),
		Height:      uint64(pending.blockHeight),
		BlockHash:   pending.blockHash.String(),
		TxIndex:     txIndex,
		BranchProof: proof,
	}
	return spv, nil
}

// GetBalance 获取余额
func (b *btcWallet) GetBalance() (btcutil.Amount, error) {
	if b.tssAddress == nil {
		return 0, fmt.Errorf("TSS address not ready")
	}

	// 获取已确认余额
	balance, err := b.Wallet.CalculateBalance(b.requiredConfs)
	if err != nil {
		return 0, fmt.Errorf("calculate balance failed: %w", err)
	}

	return balance, nil
}

// GetDepositChannel 获取充值通知channel
func (b *btcWallet) GetDepositChannel() <-chan *pendingTx {
	return b.depositChan
}

// GetWithdrawChannel 获取提现通知channel
func (b *btcWallet) GetWithdrawChannel() <-chan *pendingTx {
	return b.withdrawChan
}
