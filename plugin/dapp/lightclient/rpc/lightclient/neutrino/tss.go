package neutrino

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/crypto/secp256k1"
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
	tssBucketName = "rgbx-tss"
	dkgResultKey  = "dkg-result"
)

// tssSignNotify is the notification message for TSS signing
type tssSignNotify struct {
	TxHash    []byte
	Timestamp int64
}

type tssService struct {
	client      *neutrinoClient
	cfg         tssConfig
	commitTxKey crypto.PrivKey

	// TSS related
	dkgResult    *tss.DKGResult
	tssPublicKey *btcec.PublicKey
	btcAddress   btcutil.Address
	pkScript     []byte
	dkgCompleted atomic.Bool

	// P2P channels
	subChan chan *types.TopicData
}

func newTssService(n *neutrinoClient) *tssService {
	t := &tssService{
		client:  n,
		subChan: make(chan *types.TopicData, 100),
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

	// Ensure DKG is completed
	go t.ensureCommitDKG()
}

func (t *tssService) ensureCommitDKG() {
	// Wait for cross chain info to be available
	info := t.client.getCrossChainInfo()
	for info == nil {
		time.Sleep(3 * time.Second)
		log.Debug("ensureDKG getCrossChainInfo wait 3 seconds...")
		info = t.client.getCrossChainInfo()
	}

	if info.DepositAddress != "" {
		log.Debug("ensureDKG already exist, loading from database")
		err := t.loadDKGFromDB()
		if err != nil {
			log.Error("ensureDKG loadDKGFromDB error", "err", err)
		} else {
			t.dkgCompleted.Store(true)
		}
		return
	}

	t.commitTxKey = t.client.getCommitKey()

	// Try to load existing DKG result from database
	err := t.loadDKGFromDB()
	if err == nil && t.dkgResult != nil {
		log.Debug("ensureDKG loaded existing DKG from database")
		// Commit existing DKG result to main chain
		t.commitDKGResult()
		return
	}

	log.Info("ensureDKG starting new DKG process")

	// Perform DKG process with retry
	var dkgResult *tss.DKGResult
	for {
		dkgResult, err = gg18.ProcessDKG(t.cfg.Peers, t.cfg.Threshold, t.cfg.Rank)
		if err == nil {
			break
		}
		log.Error("ensureDKG ProcessDKG retry", "err", err)
		time.Sleep(time.Minute)
	}

	t.dkgResult = dkgResult

	// Extract public key from DKG result (PubX, PubY coordinates)
	pubKey, err := t.coordinatesToPublicKey(dkgResult.PubX, dkgResult.PubY)
	if err != nil {
		log.Error("ensureDKG coordinatesToPublicKey error", "err", err)
		return
	}
	t.tssPublicKey = pubKey

	// Generate Bitcoin address from public key
	err = t.generateBitcoinAddress()
	if err != nil {
		log.Error("ensureDKG generateBitcoinAddress error", "err", err)
		return
	}

	// Save DKG result to database with retry
	t.saveDKGToDB()

	// Mark DKG as completed
	t.dkgCompleted.Store(true)

	// Commit DKG result to main chain with retry
	t.commitDKGResult()
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
	pubKey, err := t.coordinatesToPublicKey(dkgResult.PubX, dkgResult.PubY)
	if err != nil {
		return err
	}
	t.tssPublicKey = pubKey

	// Generate Bitcoin address
	return t.generateBitcoinAddress()
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
func (t *tssService) generateBitcoinAddress() error {
	if t.tssPublicKey == nil {
		return types.ErrInvalidParam
	}

	// Generate Bitcoin address (P2WPKH format)
	chainParams := &t.client.neutrinoCfg.ChainParams
	pubKeyHash := btcutil.Hash160(t.tssPublicKey.SerializeCompressed())
	addr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, chainParams)
	if err != nil {
		return err
	}

	t.btcAddress = addr
	t.pkScript, err = txscript.PayToAddrScript(addr)
	if err != nil {
		panic("payToAddrScript error " + err.Error())
	}
	log.Info("generateBitcoinAddress", "address", t.btcAddress.String())

	return nil
}

// coordinatesToPublicKey converts PubX and PubY coordinates to btcec.PublicKey
func (t *tssService) coordinatesToPublicKey(pubX, pubY []byte) (*btcec.PublicKey, error) {
	// Create field values from coordinates
	var fieldX, fieldY btcec.FieldVal
	if fieldX.SetByteSlice(pubX) {
		return nil, types.ErrInvalidParam
	}
	if fieldY.SetByteSlice(pubY) {
		return nil, types.ErrInvalidParam
	}

	// Create public key from field values
	pubKey := btcec.NewPublicKey(&fieldX, &fieldY)

	// Verify the point is on the curve
	if !pubKey.IsOnCurve() {
		return nil, types.ErrInvalidParam
	}

	return pubKey, nil
}

// commitDKGResult commits DKG result to main chain with retry until success
func (t *tssService) commitDKGResult() {
	// Create CommitDKG transaction
	commitDKG := &rtypes.CommitDKG{
		AssetSymbol: "BTC",
		DkgAddress:  t.btcAddress.EncodeAddress(),
	}

	for {
		// Create transaction
		tx, err := t.client.createTx(rtypes.RgbxX, rtypes.NameCommitDKGAction, types.Encode(commitDKG))
		if err != nil {
			log.Error("commitDKGResult createTx retry", "err", err)
			time.Sleep(time.Second * 3)
			continue
		}

		// Set fee and sign
		tx.Fee, _ = tx.GetRealFee(t.client.getProperFeeRate())
		tx.Sign(types.EncodeSignID(secp256k1.ID, t.client.commitAddressType), t.commitTxKey)

		// Send transaction to main chain
		err = t.client.sendTx2MainChain(tx)
		if err != nil {
			log.Error("commitDKGResult sendTx retry", "txHash", hex.EncodeToString(tx.Hash()), "err", err)
			time.Sleep(time.Second * 3)
			continue
		}

		log.Debug("commitDKGResult success", "txHash", hex.EncodeToString(tx.Hash()), "btcAddress", t.btcAddress)
		break
	}
	t.dkgCompleted.Store(true)
}

// processSignBtcTx processes a Bitcoin transaction using TSS protocol
// This is called by the main node to initiate TSS signing
func (t *tssService) processSignBtcTx(tx *wire.MsgTx, txType string, inputAmounts []int64, payload []byte) error {
	if !t.dkgCompleted.Load() {
		log.Error("signMsg dkg not completed")
		return types.ErrNotSupport
	}

	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSizeStripped()))
	_ = tx.SerializeNoWitness(buf)
	notify := &lighttypes.TssSignNotify{
		InputAmounts: inputAmounts,
		TxType:       txType,
		Payload:      payload,
		BtcTxData:    buf.Bytes(),
	}
	// Publish notification to all TSS nodes
	t.pubMsg(tssSignNotifyTopic, types.Encode(notify))
	log.Debug("signMsg published notification", "txType", txType, "payload", hex.EncodeToString(payload))
	return t.signBtcTx(tx, inputAmounts)
}

func (t *tssService) signMsg(msg []byte) ([]byte, error) {
	sigResult, err := gg18.ProcessSign(t.cfg.Peers, msg, t.dkgResult)
	if err != nil {
		log.Error("signMsg ProcessSign", "err", err)
		return nil, err
	}

	// Extract R and S from signature result and serialize
	// sigResult contains R and S as big.Int coordinates
	signature := append(sigResult.R.Bytes(), sigResult.S.Bytes()...)

	log.Debug("signMsg success", "signature", hex.EncodeToString(signature))
	return signature, nil
}

func (t *tssService) signBtcTx(tx *wire.MsgTx, inputAmounts []int64) error {
	if len(tx.TxIn) != len(inputAmounts) {
		return fmt.Errorf("input count mismatch: tx=%d inputAmounts=%d", len(tx.TxIn), len(inputAmounts))
	}
	txSigHashes := txscript.NewTxSigHashes(tx, nil)
	for idx, in := range tx.TxIn {
		// 计算签名哈希（使用预计算的脚本）
		sigHash, err := txscript.CalcWitnessSigHash(t.pkScript, txSigHashes, txscript.SigHashAll, tx, idx, inputAmounts[idx])
		if err != nil {
			return fmt.Errorf("calc sig hash failed for input %d: %w", idx, err)
		}
		sig, err := t.signMsg(sigHash)
		if err != nil {
			return fmt.Errorf("signMsg failed for input %d: %w", idx, err)
		}
		pubKeyBytes := t.tssPublicKey.SerializeCompressed()
		sigWithHashType := append(sig, byte(txscript.SigHashAll))

		// 构建witness（P2WPKH格式）
		// witness = [signature + hashType, pubkey]
		in.Witness = wire.TxWitness{sigWithHashType, pubKeyBytes}
		log.Debug("signWithdrawTx applied signature to input", "idx", idx)
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

func (t *tssService) validateWithdrawSignNotify(notify *lighttypes.TssSignNotify) (*wire.MsgTx, []int64, error) {
	tx, inputAmounts, err := t.parseTxFromNotify(notify)
	if err != nil {
		return nil, nil, err
	}
	if notify.TxType != transactionTypeWithdraw {
		return tx, inputAmounts, nil
	}

	var withdrawAmount int64
	var withdrawAddress string
	for _, output := range tx.TxOut {
		if len(output.PkScript) > 0 && output.PkScript[0] == txscript.OP_RETURN {
			continue
		}
		if len(t.pkScript) > 0 && bytes.Equal(output.PkScript, t.pkScript) {
			continue
		}
		withdrawAmount += output.Value
		if withdrawAddress == "" {
			_, addrs, _, err := txscript.ExtractPkScriptAddrs(output.PkScript, &t.client.neutrinoCfg.ChainParams)
			if err != nil || len(addrs) == 0 {
				return nil, nil, fmt.Errorf("extract withdraw address failed: %w", err)
			}
			withdrawAddress = addrs[0].String()
		} else {
			_, addrs, _, err := txscript.ExtractPkScriptAddrs(output.PkScript, &t.client.neutrinoCfg.ChainParams)
			if err == nil && len(addrs) > 0 && addrs[0].String() != withdrawAddress {
				return nil, nil, fmt.Errorf("multiple withdraw addresses in tx")
			}
		}
	}
	if withdrawAddress == "" || withdrawAmount <= 0 {
		return nil, nil, fmt.Errorf("invalid withdraw output")
	}
	if len(notify.Payload) == 0 {
		return nil, nil, fmt.Errorf("empty withdraw asset hash")
	}
	// payload is the chain33 tx hash
	withdraw, err := t.client.getRgbxWithdrawAsset(notify.Payload)
	if err != nil {
		return nil, nil, err
	}
	if withdraw.GetDestinationAddr() != withdrawAddress {
		return nil, nil, fmt.Errorf("withdraw address mismatch")
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
	if fee < 0 {
		return nil, nil, fmt.Errorf("invalid fee calculation")
	}
	expectedAmount := withdraw.GetAmount() - fee
	if expectedAmount <= 0 {
		return nil, nil, fmt.Errorf("withdraw amount too small for fee")
	}
	if expectedAmount != withdrawAmount {
		return nil, nil, fmt.Errorf("withdraw amount mismatch")
	}
	return tx, inputAmounts, nil
}

// handleSignNotify handles incoming TSS sign notifications
// All nodes (including main node) receive this and participate in signing
func (t *tssService) handleSignNotify(msg []byte) {
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

	tx, inputAmounts, err := t.validateWithdrawSignNotify(notify)
	if err != nil {
		log.Error("handleSignNotify validate", "err", err)
		return
	}

	if err := t.signBtcTx(tx, inputAmounts); err != nil {
		log.Error("handleSignNotify signBtcTxData", "err", err)
		return
	}
	log.Debug("handleSignNotify success", "txType", notify.TxType)
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

// getBitcoinAddress returns the Bitcoin address generated from TSS public key
func (t *tssService) getBitcoinAddress() btcutil.Address {
	return t.btcAddress
}

// isDKGCompleted checks if DKG is completed using atomic operation
func (t *tssService) isDKGCompleted() bool {
	return t.dkgCompleted.Load()
}
