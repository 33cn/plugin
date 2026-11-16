package neutrino

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/system/crypto/tss"
	"github.com/33cn/chain33/system/crypto/tss/gg18"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/walletdb"
)

const (
	moduleName = "dapp-lightclient-neutrino"
	// TSS pubsub topic - notification only
	tssSignNotifyTopic = "rgbx/tssSignNotify/1.0"

	// Database bucket and keys
	tssBucketName  = "rgbx-tss"
	dkgResultKey   = "dkg-result"
	dkgSessionName = "rgbx-btc-dkg"
)

type signTask struct {
	idx         int
	sigHash     []byte
	sessionName string
	signers     []string
	result      chan *signResult
}

type signResult struct {
	sig []byte
	err error
}

type tssService struct {
	client *neutrinoClient
	cfg    tssConfig

	// TSS related
	dkgResult    *tss.DKGResult
	tssPublicKey *btcec.PublicKey
	tssAddress   btcutil.Address
	pkScript     []byte
	dkgCompleted atomic.Bool

	// P2P channels
	subChan    chan *types.TopicData
	selfPeerId string
	signTaskCh chan *signTask
}

func newTssService(n *neutrinoClient) *tssService {
	t := &tssService{
		client:     n,
		subChan:    make(chan *types.TopicData, 100),
		signTaskCh: make(chan *signTask, 1024),
	}
	t.cfg = n.cfg.Tss
	return t
}

func (t *tssService) start() {
	log.Debug("tssService start")

	// Subscribe to TSS sign notification topic
	go t.subTopic(tssSignNotifyTopic)

	// Handle subscribed messages
	go t.handleSubMsg()

	// Dedicated signing worker.
	for i := 0; i < runtime.NumCPU(); i++ {
		go t.handleSignTask()
	}

	// Ensure DKG is completed
	go t.init()
}

func (t *tssService) handleSignTask() {
	for {
		select {
		case <-t.client.ctx.Done():
			return
		case task := <-t.signTaskCh:
			task.result <- t.signMsg(task.sigHash, task.sessionName, task.signers)
		}
	}
}

func (t *tssService) init() {
	// Wait for cross chain info to be available
	info := t.client.getCrossChainInfo()
	for info == nil {
		time.Sleep(3 * time.Second)
		log.Debug("ensureDKG getCrossChainInfo wait 3 seconds...")
		info = t.client.getCrossChainInfo()
	}

	if info.GetTssAddress() != "" {
		log.Debug("ensureDKG already exist, loading from database")
		err := t.loadDKGFromDB()
		if err != nil {
			panic("ensureDKG loadDKGFromDB error: " + err.Error())
		}
		t.dkgCompleted.Store(true)
		return
	}
	log.Info("init tssService starting new DKG process")

	// Perform DKG process with retry
	var dkgResult *tss.DKGResult
	var err error
	for {
		dkgResult, err = gg18.ProcessDKG(t.cfg.Peers, t.cfg.Threshold, t.cfg.Rank, dkgSessionName)
		if err == nil {
			break
		}
		log.Error("init tssService ProcessDKG retry", "err", err)
		time.Sleep(time.Minute)
	}

	t.dkgResult = dkgResult

	// Extract public key from DKG result (PubX, PubY coordinates)
	pubkey, err := tss.ParseBtcecPublicKey(dkgResult)
	if err != nil {
		log.Error("init tssService ParseBtcecPublicKey error", "err", err)
		return
	}
	t.tssPublicKey = pubkey

	// Generate Bitcoin address from public key
	err = t.generateTssAddress()
	if err != nil {
		log.Error("init tssService generateTssAddress error", "err", err)
		return
	}

	// Save DKG result to database with retry
	t.saveDKGToDB()

	// Commit DKG result to main chain with retry
	commitDKG := &rtypes.CommitDKG{
		AssetSymbol: rtypes.BTCSymbol,
		DkgAddress:  t.tssAddress.EncodeAddress(),
		PkScript:    t.pkScript,
	}
	t.client.submitMainchainTxUntilSuccess(rtypes.RgbxX, rtypes.NameCommitDKGAction, commitDKG)
	t.dkgCompleted.Store(true)
	for {
		peers, err := tss.FetchConnectedPeers(t.client.qclient, time.Second*3)
		if err == nil && len(peers) > 0 && peers[len(peers)-1].Self {
			t.selfPeerId = peers[len(peers)-1].Name
			break
		}
		log.Debug("init tssService waitForSelfPeerId FetchConnectedPeers retry", "err", err)
		time.Sleep(time.Second * 3)
	}

}

func (t *tssService) loadDKGFromDB() error {
	var dkgData []byte

	err := walletdb.View(t.client.neutrinoCfg.Database, func(tx walletdb.ReadTx) error {
		bucket := tx.ReadBucket([]byte(tssBucketName))
		if bucket == nil {
			return walletdb.ErrBucketNotFound
		}

		// Load DKG result
		dkgData = bucket.Get([]byte(dkgResultKey))
		if dkgData == nil {
			return types.ErrNotFound
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Decode DKG result
	var dkgResult tss.DKGResult
	err = types.Decode(dkgData, &dkgResult)
	if err != nil {
		return err
	}
	t.dkgResult = &dkgResult

	// Extract public key from DKG result coordinates
	pubKey, err := tss.ParseBtcecPublicKey(&dkgResult)
	if err != nil {
		log.Error("loadDKGFromDB ParseBtcecPublicKey error", "err", err)
		return err
	}
	t.tssPublicKey = pubKey

	// Generate Bitcoin address
	return t.generateTssAddress()
}

// saveDKGToDB saves DKG result to database with retry until success
func (t *tssService) saveDKGToDB() {
	// Encode DKG result
	dkgData := types.Encode(t.dkgResult)

	for {
		err := walletdb.Update(t.client.neutrinoCfg.Database, func(tx walletdb.ReadWriteTx) error {
			bucket, err := tx.CreateTopLevelBucket([]byte(tssBucketName))
			if err != nil {
				return err
			}

			// Save DKG result
			return bucket.Put([]byte(dkgResultKey), dkgData)
		})

		if err == nil {
			log.Debug("saveDKGToDB success")
			return
		}

		log.Error("saveDKGToDB retry", "err", err)
		time.Sleep(time.Second * 3)
	}
}

// generateBitcoinAddress generates Bitcoin address from TSS public key
func (t *tssService) generateTssAddress() error {
	if t.tssPublicKey == nil {
		return types.ErrInvalidParam
	}

	// Generate Bitcoin address (P2WPKH format)
	chainParams := &t.client.neutrinoCfg.ChainParams
	pubKeyHash := btcutil.Hash160(t.tssPublicKey.SerializeCompressed())
	addr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, chainParams)
	if err != nil {
		log.Error("generateTssAddress NewAddressWitnessPubKeyHash", "pubkey", hex.EncodeToString(t.tssPublicKey.SerializeCompressed()), "err", err)
		return err
	}

	t.tssAddress = addr
	t.pkScript, err = txscript.PayToAddrScript(addr)
	if err != nil {
		panic("generateTssAddress payToAddrScript error , address: " + addr.String() + " err: " + err.Error())
	}
	log.Info("generateTssAddress", "address", addr.String())

	return nil
}

func (t *tssService) waitForSufficientSigners() []string {
	var signers []string
	t.client.waitUntilDone("waitForSufficientSigners", func() bool {
		signers = tss.GetValidPeerCombination(t.client.qclient, t.cfg.Threshold, t.dkgResult.Bks)
		return len(signers) > 0
	}, time.Second*3)
	return signers
}

// processSignBtcTx processes a Bitcoin transaction using TSS protocol
// This is called by the main node to initiate TSS signing
func (t *tssService) processSignBtcTx(tx *wire.MsgTx, txType string, inputAmounts []int64, payload []byte) error {

	signers := t.waitForSufficientSigners()
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSizeStripped()))
	err := tx.SerializeNoWitness(buf)
	if err != nil {
		log.Error("processSignBtcTx SerializeNoWitness", "err", err)
		return err
	}
	notify := &lighttypes.TssSignNotify{
		InputAmounts: inputAmounts,
		TxType:       txType,
		Payload:      payload,
		BtcTxData:    buf.Bytes(),
		Signers:      signers,
	}
	// Publish notification to all TSS nodes
	t.pubMsg(tssSignNotifyTopic, types.Encode(notify))
	log.Debug("signMsg published notification", "txType", txType, "payload", hex.EncodeToString(payload), "signers", signers)
	return t.signBtcTx(tx, inputAmounts, signers)
}

func (t *tssService) signMsg(msg []byte, seesionName string, signers []string) *signResult {
	result := &signResult{}
	sigResult, err := gg18.ProcessSign(signers, msg, t.dkgResult, seesionName)
	if err != nil {
		log.Error("signMsg ProcessSign", "err", err)
		result.err = err
		return result
	}

	signature, err := gg18.AliceToBtcecSignature(sigResult)
	if err != nil {
		log.Error("signMsg AliceToBtcecSignature", "err", err)
		result.err = err
		return result
	}
	signatureBytes := signature.Serialize()
	log.Debug("signMsg success", "signature", hex.EncodeToString(signatureBytes))
	result.sig = signatureBytes
	return result
}

func (t *tssService) signBtcTx(tx *wire.MsgTx, inputAmounts []int64, signers []string) error {
	if len(tx.TxIn) != len(inputAmounts) {
		return fmt.Errorf("input count mismatch: tx=%d inputAmounts=%d", len(tx.TxIn), len(inputAmounts))
	}
	prevOutFetcher := txscript.NewMultiPrevOutFetcher(make(map[wire.OutPoint]*wire.TxOut, len(tx.TxIn)))
	for idx, txIn := range tx.TxIn {
		prevOutFetcher.AddPrevOut(txIn.PreviousOutPoint, &wire.TxOut{
			Value:    inputAmounts[idx],
			PkScript: t.pkScript,
		})
	}

	txSigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
	txHash := tx.TxID()
	pubKeyBytes := t.tssPublicKey.SerializeCompressed()
	sigTasks := make([]*signTask, 0, len(tx.TxIn))
	for idx := range tx.TxIn {
		// 计算签名哈希（使用预计算的脚本）
		sigHash, err := txscript.CalcWitnessSigHash(t.pkScript, txSigHashes, txscript.SigHashAll, tx, idx, inputAmounts[idx])
		if err != nil {
			return fmt.Errorf("calc sig hash failed for input %d: %w", idx, err)
		}

		sigTask := &signTask{
			idx:         idx,
			sigHash:     sigHash,
			sessionName: fmt.Sprintf("btctx-%s-%d", txHash, idx),
			signers:     signers,
			result:      make(chan *signResult, 1),
		}
		sigTasks = append(sigTasks, sigTask)
		t.signTaskCh <- sigTask
	}

	for _, sigTask := range sigTasks {
		result := <-sigTask.result
		if result.err != nil {
			return fmt.Errorf("signMsg failed for input %d: %w", sigTask.idx, result.err)
		}
		sigWithHashType := append(result.sig, byte(txscript.SigHashAll))
		tx.TxIn[sigTask.idx].Witness = wire.TxWitness{sigWithHashType, pubKeyBytes}
		log.Debug("signBtcTx applied signature to input", "idx", sigTask.idx)
	}

	return nil
}

func (t *tssService) parseTxFromNotify(notify *lighttypes.TssSignNotify) (*wire.MsgTx, []int64, error) {
	if notify == nil {
		return nil, nil, types.ErrInvalidParam
	}
	if len(notify.BtcTxData) == 0 {
		return nil, nil, fmt.Errorf("empty BtcTxData")
	}
	if len(notify.InputAmounts) == 0 {
		return nil, nil, fmt.Errorf("empty input amounts")
	}
	var tx wire.MsgTx
	if err := tx.DeserializeNoWitness(bytes.NewReader(notify.BtcTxData)); err != nil {
		return nil, nil, fmt.Errorf("deserialize tx failed: %w", err)
	}
	if len(tx.TxIn) != len(notify.InputAmounts) {
		return nil, nil, fmt.Errorf("input count mismatch: tx=%d inputAmounts=%d", len(tx.TxIn), len(notify.InputAmounts))
	}
	return &tx, notify.InputAmounts, nil
}

func (t *tssService) validateWithdrawTx(tx *wire.MsgTx, inputAmounts []int64, req *withdrawRequest) error {

	btcAddr, err := btcutil.DecodeAddress(req.toAddress, &t.client.neutrinoCfg.ChainParams)
	if err != nil {
		log.Error("validateWithdrawTx decode address", "err", err, "address", req.toAddress,
			"withdrawTxHash", hex.EncodeToString(req.chain33WithDrawHash))
		return fmt.Errorf("decode address failed")
	}
	btcAddrScript, err := txscript.PayToAddrScript(btcAddr)
	if err != nil {
		log.Error("validateWithdrawTx pay to addr script", "address", req.toAddress,
			"withdrawTxHash", hex.EncodeToString(req.chain33WithDrawHash), "err", err)
		return fmt.Errorf("pay to addr script failed")
	}
	var withdrawAmount, changeAmount int64
	for _, output := range tx.TxOut {
		if len(output.PkScript) > 0 && output.PkScript[0] == txscript.OP_RETURN {
			continue
		}
		if len(t.pkScript) > 0 && bytes.Equal(output.PkScript, t.pkScript) {
			changeAmount += output.Value
			continue
		}
		if !bytes.Equal(output.PkScript, btcAddrScript) {
			log.Error("validateWithdrawTx unexpected output script", "address", req.toAddress,
				"expected", hex.EncodeToString(btcAddrScript), "actual", hex.EncodeToString(output.PkScript),
				"withdrawTxHash", hex.EncodeToString(req.chain33WithDrawHash))
			return fmt.Errorf("unexpected output script")
		}
		withdrawAmount += output.Value
	}

	var totalInput int64
	for _, amount := range inputAmounts {
		totalInput += amount
	}
	var totalOutput int64
	for _, out := range tx.TxOut {
		totalOutput += out.Value
	}
	fee := totalInput - totalOutput
	expectedFee := int64(estimateBtcFee(tx, req.feeRate))
	// 控制手续费在合理范围内
	if fee > 2*expectedFee || fee < 0 {
		log.Error("validateWithdrawSignNotify invalid fee", "fee", fee, "expected", expectedFee,
			"withdrawTxHash", hex.EncodeToString(req.chain33WithDrawHash))
		return fmt.Errorf("invalid fee")
	}
	// 验证提现的关键是总支出不能超过提现金额，允许的最大磨损不能超过最小找零金额
	if totalInput-changeAmount > int64(req.amount)+minChangeAmount {
		log.Error("validateWithdrawSignNotify withdraw overflowed",
			"actualWithdraw", withdrawAmount, "changeAmount", changeAmount,
			"totalInput", totalInput, "expectWithdraw", int64(req.amount),
			"withdrawTxHash", hex.EncodeToString(req.chain33WithDrawHash))
		return fmt.Errorf("withdraw overflowed")
	}
	if withdrawAmount > int64(req.amount) || withdrawAmount < minChangeAmount {
		log.Error("validateWithdrawSignNotify invalid withdraw amount", "actualWithdraw", withdrawAmount,
			"expectWithdraw", int64(req.amount), "withdrawTxHash", hex.EncodeToString(req.chain33WithDrawHash))
		return fmt.Errorf("invalid withdraw amount")
	}
	return nil
}

func (t *tssService) checkNonOfficialWithdrawSign(chain33WithDrawHash []byte) (*rtypes.PendingTx, error) {

	txHash := hex.EncodeToString(chain33WithDrawHash)
	pendingTx, err := t.client.getRgbxPendingTxByHash(chain33WithDrawHash)
	if err != nil {
		log.Error("checkNonOfficialWithdrawSign getRgbxPendingTxByHash", "txHash", txHash, "err", err)
		return nil, err
	}
	if pendingTx.GetConfirmed() {
		return nil, fmt.Errorf("withdraw already confirmed")
	}
	return pendingTx, nil
}

func (t *tssService) checkStickyInput(chain33WithDrawHash []byte, tx *wire.MsgTx) error {
	if len(tx.TxIn) == 0 {
		return fmt.Errorf("withdraw tx has no inputs")
	}
	stickyOutPoint := tx.TxIn[len(tx.TxIn)-1].PreviousOutPoint.String()
	// 验证绑定的哈希是否一致
	expectedHash := t.client.getExpectedWithdrawHash(stickyOutPoint)
	if len(expectedHash) > 0 && !bytes.Equal(expectedHash, chain33WithDrawHash) {
		log.Error("checkStickyInput sticky input mismatch", "expected", hex.EncodeToString(expectedHash),
			"actual", hex.EncodeToString(chain33WithDrawHash), "stickyOutPoint", stickyOutPoint)
		return fmt.Errorf("invalid sticky input")
	}

	// 如果本地已记录过成功签名的绑定utxo，后续请求必须一致
	expectUTXO := t.client.getWithdrawStickyUTXO(chain33WithDrawHash)
	if expectUTXO != nil && expectUTXO.OutPoint.String() != stickyOutPoint {
		log.Error("checkStickyInput sticky input changed", "expected", expectUTXO.OutPoint.String(),
			"actual", stickyOutPoint, "chain33WithDrawHash", hex.EncodeToString(chain33WithDrawHash))
		return fmt.Errorf("invalid sticky input")
	}
	return nil
}

// handleSignNotify handles incoming TSS sign notifications
// All nodes (including main node) receive this and participate in signing
func (t *tssService) handleSignNotify(msg []byte) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("handleSignNotify panic", "err", r)
		}
	}()
	if !t.dkgCompleted.Load() {
		log.Error("handleSignNotify", "err", "DKG not completed")
		return
	}

	notify := &lighttypes.TssSignNotify{}
	err := types.Decode(msg, notify)
	if err != nil {
		log.Error("handleSignNotify Decode", "err", err)
		return
	}
	isSigner := false
	for _, signer := range notify.Signers {
		if signer == t.selfPeerId {
			isSigner = true
			break
		}
	}
	if !isSigner {
		log.Debug("handleSignNotify not signer", "signers", notify.Signers, "selfPeerId", t.selfPeerId, "type", notify.TxType)
		return
	}

	tx, inputAmounts, err := t.parseTxFromNotify(notify)
	if err != nil {
		log.Error("handleSignNotify parseTxFromNotify", "type", notify.TxType, "err", err)
		return
	}

	if notify.TxType == transactionTypeWithdraw {
		chain33WithDrawHash := notify.Payload
		if err = t.checkStickyInput(chain33WithDrawHash, tx); err != nil {
			log.Error("handleSignNotify checkStickyInput", "err", err,
				"withDrawHash", hex.EncodeToString(chain33WithDrawHash), "btcHash", tx.TxHash().String())
			return
		}
		pendingTx, err := t.checkNonOfficialWithdrawSign(chain33WithDrawHash)
		if err != nil {
			log.Error("handleSignNotify checkNonOfficialWithdrawSign", "type", notify.TxType,
				"withDrawHash", hex.EncodeToString(chain33WithDrawHash), "btcHash", tx.TxHash().String(), "err", err)
			return
		}
		req := pending2WithdrawRequest(pendingTx)
		err = t.validateWithdrawTx(tx, inputAmounts, req)
		if err != nil {
			log.Error("handleSignNotify validateWithdrawTx", "err", err,
				"withdrawHash", hex.EncodeToString(chain33WithDrawHash), "btcHash", tx.TxHash().String())
			return
		}
	}

	if err = t.signBtcTx(tx, inputAmounts, notify.Signers); err != nil {
		log.Error("handleSignNotify signBtcTx", "type", notify.TxType, "err", err)
		return
	}
	if notify.TxType == transactionTypeWithdraw {
		stickyUTXO := &UTXO{
			OutPoint: tx.TxIn[len(tx.TxIn)-1].PreviousOutPoint,
		}
		if err = t.client.setWithdrawStickyUTXO(notify.Payload, stickyUTXO); err != nil {
			log.Error("handleSignNotify setWithdrawStickyUTXO", "err", err,
				"withdrawHash", hex.EncodeToString(notify.Payload), "btcHash", tx.TxHash().String())
		}
	}
	log.Debug("handleSignNotify success", "txType", notify.TxType,
		"payload", hex.EncodeToString(notify.Payload), "btcHash", tx.TxHash().String())
}

// subTopic subscribes to a P2P topic
func (t *tssService) subTopic(topic string) {
	data := &types.SubTopic{Topic: topic, Module: moduleName}

	for {
		err := t.sendP2PMsg(types.EventSubTopic, data)
		if err == nil {
			log.Info("subTopic success", "topic", topic)
			break
		}
		log.Debug("subTopic", "topic", topic, "err", err)
		time.Sleep(time.Second)
	}
}

// pubMsg publishes a message to a P2P topic
func (t *tssService) pubMsg(topic string, msg []byte) {
	data := &types.PublishTopicMsg{Topic: topic, Msg: msg}
	tryCount := 0

	for {
		tryCount++
		err := t.sendP2PMsg(types.EventPubTopicMsg, data)
		if err == nil || tryCount >= 3 {
			break
		}
		log.Error("pubMsg", "topic", topic, "tryCount", tryCount, "err", err)
		time.Sleep(time.Second)
	}
}

func (t *tssService) sendP2PMsg(ty int64, data interface{}) error {
	msg := t.client.qclient.NewMessage("p2p", ty, data)
	err := t.client.qclient.Send(msg, true)
	if err != nil {
		return err
	}

	resp, err := t.client.qclient.WaitTimeout(msg, time.Second*5)
	if err != nil {
		return err
	}

	reply, ok := resp.GetData().(*types.Reply)
	if !ok {
		return types.ErrTypeAsset
	}

	if !reply.GetIsOk() {
		return types.ErrInvalidParam
	}

	return nil
}

// handleSubMsg handles subscribed messages from P2P network
func (t *tssService) handleSubMsg() {
	for {
		select {
		case <-t.client.ctx.Done():
			return

		case data := <-t.subChan:
			if data.Topic == tssSignNotifyTopic {
				t.handleSignNotify(data.GetData())
			}
		}
	}
}

// isDKGCompleted checks if DKG is completed using atomic operation
func (t *tssService) isDKGCompleted() bool {
	return t.dkgCompleted.Load()
}
