// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"math/big"
	"strings"
	"time"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/difficulty"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/golang/protobuf/proto"
)

type btcStore struct {
	db dbm.KVDB
}

func newBtcStore(db dbm.KVDB) *btcStore {
	return &btcStore{db: db}
}

func (b *btcStore) getBtcHeadHeightFromDb(key []byte) (int64, error) {
	val, err := b.db.Get(key)
	if err != nil {
		return -1, err
	}

	height, err := decodeHeight(val)
	if err != nil {
		return -1, err
	}

	return height, nil
}

func (b *btcStore) getLastBtcHeadHeight() (int64, error) {
	key := relayBTCHeaderLastHeight
	return b.getBtcHeadHeightFromDb(key)
}

func (b *btcStore) getBtcHeadByHeight(height int64) (*ty.BtcHeader, error) {
	var head ty.BtcHeader
	key := calcBtcHeaderKeyHeight(height)
	val, err := b.db.Get(key)
	if err != nil {
		return nil, err
	}
	err = types.Decode(val, &head)
	if err != nil {
		return nil, err
	}

	return &head, nil
}

func (b *btcStore) getLastBtcHead() (*ty.BtcHeader, error) {
	height, err := b.getLastBtcHeadHeight()
	if err != nil {
		return nil, err
	}

	head, err := b.getBtcHeadByHeight(height)
	if err != nil {
		return nil, err
	}

	return head, nil
}

func (b *btcStore) saveBlockHead(head *ty.BtcHeader) ([]*types.KeyValue, error) {
	var kv []*types.KeyValue
	var key []byte

	val, err := proto.Marshal(head)
	if err != nil {
		relaylog.Error("saveBlockHead", "height", head.Height, "hash", head.Hash)
		return nil, err

	}

	// hash:header
	key = calcBtcHeaderKeyHash(head.Hash)
	kv = append(kv, &types.KeyValue{Key: key, Value: val})
	// height:header
	key = calcBtcHeaderKeyHeight(int64(head.Height))
	kv = append(kv, &types.KeyValue{Key: key, Value: val})

	// prefix-height:height
	key = calcBtcHeaderKeyHeightList(int64(head.Height))
	heightBytes := types.Encode(&types.Int64{Data: int64(head.Height)})
	kv = append(kv, &types.KeyValue{Key: key, Value: heightBytes})

	return kv, nil
}

func (b *btcStore) saveBlockLastHead(head *ty.ReceiptRelayRcvBTCHeaders) ([]*types.KeyValue, error) {
	var kv []*types.KeyValue

	heightBytes := types.Encode(&types.Int64{Data: int64(head.NewHeight)})
	key := relayBTCHeaderLastHeight
	kv = append(kv, &types.KeyValue{Key: key, Value: heightBytes})

	heightBytes = types.Encode(&types.Int64{Data: int64(head.NewBaseHeight)})
	key = relayBTCHeaderBaseHeight
	kv = append(kv, &types.KeyValue{Key: key, Value: heightBytes})

	return kv, nil
}

func (b *btcStore) delBlockHead(head *ty.BtcHeader) ([]*types.KeyValue, error) {
	var kv []*types.KeyValue

	key := calcBtcHeaderKeyHash(head.Hash)
	kv = append(kv, &types.KeyValue{Key: key, Value: nil})
	// height:header
	key = calcBtcHeaderKeyHeight(int64(head.Height))
	kv = append(kv, &types.KeyValue{Key: key, Value: nil})

	// prefix-height:height
	key = calcBtcHeaderKeyHeightList(int64(head.Height))
	kv = append(kv, &types.KeyValue{Key: key, Value: nil})

	return kv, nil
}

func (b *btcStore) delBlockLastHead(head *ty.ReceiptRelayRcvBTCHeaders) ([]*types.KeyValue, error) {
	var kv []*types.KeyValue
	var key []byte

	heightBytes := types.Encode(&types.Int64{Data: int64(head.LastHeight)})
	key = relayBTCHeaderLastHeight
	kv = append(kv, &types.KeyValue{Key: key, Value: heightBytes})

	heightBytes = types.Encode(&types.Int64{Data: int64(head.LastBaseHeight)})
	key = relayBTCHeaderBaseHeight
	kv = append(kv, &types.KeyValue{Key: key, Value: heightBytes})

	return kv, nil
}

func decodeHeight(heightBytes []byte) (int64, error) {
	var height types.Int64
	err := types.Decode(heightBytes, &height)
	if err != nil {
		return -1, err
	}
	return height.Data, nil
}

func (b *btcStore) getBtcCurHeight(req *ty.ReqRelayQryBTCHeadHeight) (types.Message, error) {

	height, err := b.getLastBtcHeadHeight()
	if err == types.ErrNotFound {
		height = -1
	} else if err != nil {
		return nil, err
	}

	key := relayBTCHeaderBaseHeight
	baseHeight, err := b.getBtcHeadHeightFromDb(key)
	if err == types.ErrNotFound {
		baseHeight = -1
	} else if err != nil {
		return nil, err
	}
	var replay ty.ReplayRelayQryBTCHeadHeight
	replay.CurHeight = height
	replay.BaseHeight = baseHeight
	return &replay, nil
}

func (b *btcStore) getBtcHeadByHash(hash string) (*ty.BtcHeader, error) {
	value, err := b.db.Get(calcBtcHeaderKeyHash(hash))
	if err != nil {
		return nil, err
	}

	var header ty.BtcHeader
	if err = types.Decode(value, &header); err != nil {
		return nil, err
	}

	return &header, nil
}

func (b *btcStore) getMerkleRootFromHeader(blockhash string) (string, error) {
	header, err := b.getBtcHeadByHash(blockhash)
	if err != nil {
		return "", err
	}

	return header.MerkleRoot, nil

}

func (b *btcStore) verifyBtcTx(cfg *types.Chain33Config, height int64, verify *ty.RelayVerify, order *ty.RelayOrder) error {
	isFork := cfg.IsDappFork(height, ty.RelayX, ty.ForkRelayVerifyBtcTx)

	var rawHash []byte
	var err error
	var head *ty.BtcHeader
	if isFork {
		// ForkRelayVerifyBtcTx 之后，交易哈希必须根据 rawTx 内容重算，
		// 订单要求的收款地址和金额必须真实存在于 rawTx 的输出中
		rawHash, err = verifyBtcTxContent(verify.GetTx(), verify.GetSpv(), order)
		if err != nil {
			return err
		}

		// 区块头由 relayd 同步且经过难度校验，其高度与时间可信
		head, err = b.getBtcHeadByHash(verify.GetSpv().GetBlockHash())
		if err != nil {
			return err
		}
		if verify.GetTx().GetBlockHeight() != head.Height ||
			(verify.GetSpv().GetHeight() != 0 && verify.GetSpv().GetHeight() != head.Height) {
			relaylog.Error("verifyTx", "tx block height", verify.GetTx().GetBlockHeight(), "spv height",
				verify.GetSpv().GetHeight(), "real header height", head.Height)
			return ty.ErrRelayBtcTxHeightErr
		}
	} else {
		var foundtx bool
		for _, outtx := range verify.GetTx().GetVout() {
			if outtx.Address == order.XAddr && outtx.Value >= order.XAmount {
				foundtx = true
			}
		}

		if !foundtx {
			return ty.ErrRelayVerifyAddrNotFound
		}
	}

	acceptTime := time.Unix(order.AcceptTime, 0)
	confirmTime := time.Unix(order.ConfirmTime, 0)
	var txTime time.Time
	if isFork {
		// 分叉后交易时间取自 SPV 证明所在区块头的时间，避免伪造 Tx.Time 重放旧交易
		txTime = time.Unix(head.Time, 0)
	} else {
		txTime = time.Unix(verify.GetTx().Time, 0)
	}

	if txTime.Sub(acceptTime) < 0 || confirmTime.Sub(txTime) < 0 {
		relaylog.Error("verifyTx", "tx time not correct to accept", txTime.Sub(acceptTime), "to confirm time", confirmTime.Sub(txTime))
		return ty.ErrRelayBtcTxTimeErr
	}

	lastHeight, err := b.getLastBtcHeadHeight()
	if err != nil {
		return err
	}

	// 确认数基于 SPV 证明实际所在区块的高度计算，区块头由 relayd 同步且经过难度校验，其高度可信
	txBlockHeight := verify.GetTx().GetBlockHeight()
	if isFork {
		txBlockHeight = head.Height
	}

	if txBlockHeight+uint64(order.XBlockWaits) > uint64(lastHeight) {
		return ty.ErrRelayWaitBlocksErr
	}

	if !isFork {
		rawHash, err = btcHashStrRevers(verify.GetTx().GetHash())
		if err != nil {
			return err
		}
	}

	sibs := verify.GetSpv().GetBranchProof()

	verifyRoot := merkle.GetMerkleRootFromBranch(sibs, rawHash, verify.GetSpv().GetTxIndex())

	var merkleRootStr string
	if isFork {
		merkleRootStr = head.MerkleRoot
	} else {
		merkleRootStr, err = b.getMerkleRootFromHeader(verify.GetSpv().GetBlockHash())
		if err != nil {
			return err
		}
	}
	realMerkleRoot, err := btcHashStrRevers(merkleRootStr)
	if err != nil {
		return err
	}

	rst := bytes.Equal(realMerkleRoot, verifyRoot)
	if !rst {
		return ty.ErrRelayVerify
	}

	return nil

}

// verifyBtcTxContent 根据 rawTx 重算 btc 交易哈希，校验与自报的 hash 及 SPV 证明的哈希一致，
// 并校验订单要求的收款地址和金额真实存在于 rawTx 的输出中，返回重算的交易哈希
func verifyBtcTxContent(btcTx *ty.BtcTransaction, spv *ty.BtcSpv, order *ty.RelayOrder) ([]byte, error) {
	rawTx := btcTx.GetRawTx()
	if rawTx == "" {
		relaylog.Error("verifyBtcTxContent", "empty rawTx of hash", btcTx.GetHash())
		return nil, ty.ErrRelayBtcTxHashErr
	}

	msgTx, err := decodeRawTx(rawTx)
	if err != nil {
		relaylog.Error("verifyBtcTxContent", "decode rawTx err", err)
		return nil, ty.ErrRelayBtcTxHashErr
	}

	// txid 使用不含 witness 的序列化计算，与区块 merkle 树一致(segwit 交易的 txid != wtxid)
	rawHash := txidFromMsgTx(msgTx)

	claimHash, err := btcHashStrRevers(btcTx.GetHash())
	if err != nil {
		relaylog.Error("verifyBtcTxContent", "decode claimed hash err", err)
		return nil, ty.ErrRelayBtcTxHashErr
	}

	if !bytes.Equal(rawHash, claimHash) {
		relaylog.Error("verifyBtcTxContent", "recomputed hash", common.ToHex(reverse(rawHash)), "not match claimed hash", btcTx.GetHash())
		return nil, ty.ErrRelayBtcTxHashErr
	}

	// SPV 证明中的交易哈希也必须与重算结果一致
	if spv.GetHash() != "" && spv.GetHash() != btcTx.GetHash() {
		relaylog.Error("verifyBtcTxContent", "spv hash", spv.GetHash(), "not match tx hash", btcTx.GetHash())
		return nil, ty.ErrRelayBtcTxHashErr
	}

	// 订单收款地址可能属于任一 btc 网络(集成测试为 simnet)，
	// 用能成功解码订单地址的网络参数解析输出地址，
	// 否则硬编码 MainNet 会把 simnet 输出地址重编码成 mainnet 地址，永远匹配不上
	params := addrNetParams(order.XAddr)
	var foundtx bool
	for _, out := range msgTx.TxOut {
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, params)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if addr.EncodeAddress() == order.XAddr && uint64(out.Value) >= order.XAmount {
				foundtx = true
			}
		}
	}

	if !foundtx {
		return nil, ty.ErrRelayVerifyAddrNotFound
	}

	return rawHash, nil
}

func (b *btcStore) verifyCmdBtcTx(verify *ty.RelayVerifyCli) error {
	rawhash, err := getRawTxHash(verify.RawTx)
	if err != nil {
		return err
	}
	sibs, err := getSiblingHash(verify.MerkBranch)
	if err != nil {
		return err
	}

	verifymerkleroot := merkle.GetMerkleRootFromBranch(sibs, rawhash, verify.TxIndex)
	str, err := b.getMerkleRootFromHeader(verify.BlockHash)
	if err != nil {
		return err
	}
	realmerkleroot, err := btcHashStrRevers(str)
	if err != nil {
		return err
	}

	rst := bytes.Equal(realmerkleroot, verifymerkleroot)
	if !rst {
		return ty.ErrRelayVerify
	}

	return nil
}

// addrNetParams 返回能成功解码 addr 的 btc 网络参数，全部失败时回退 MainNet
func addrNetParams(addr string) *chaincfg.Params {
	for _, p := range []*chaincfg.Params{
		&chaincfg.MainNetParams,
		&chaincfg.TestNet3Params,
		&chaincfg.RegressionNetParams,
		&chaincfg.SimNetParams,
	} {
		if _, err := btcutil.DecodeAddress(addr, p); err == nil {
			return p
		}
	}
	return &chaincfg.MainNetParams
}

// decodeRawTx 将 hex 编码的原始 btc 交易反序列化为 MsgTx
func decodeRawTx(rawTx string) (*wire.MsgTx, error) {
	data, err := common.FromHex(rawTx)
	if err != nil {
		return nil, err
	}
	msgTx := wire.NewMsgTx(wire.TxVersion)
	if err := msgTx.Deserialize(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return msgTx, nil
}

// txidFromMsgTx 计算 btc 交易的 txid：对不含 witness 的序列化做 double-sha256，
// 与区块 merkle 树使用的 txid 一致(segwit 交易的 txid != wtxid)
func txidFromMsgTx(msgTx *wire.MsgTx) []byte {
	var buf bytes.Buffer
	// SerializeNoWitness 写入 bytes.Buffer 不会失败
	_ = msgTx.SerializeNoWitness(&buf)
	return common.Sha2Sum(buf.Bytes())
}

func getRawTxHash(rawtx string) ([]byte, error) {
	msgTx, err := decodeRawTx(rawtx)
	if err != nil {
		return nil, err
	}
	return txidFromMsgTx(msgTx), nil
}

func getSiblingHash(sibling string) ([][]byte, error) {
	var err error
	sibsarr := strings.Split(sibling, "-")

	sibs := make([][]byte, len(sibsarr))
	for i, val := range sibsarr {
		sibs[i], err = btcHashStrRevers(val)
		if err != nil {
			return nil, err
		}

	}
	return sibs[:][:], nil
}

func btcHashStrRevers(str string) ([]byte, error) {
	data, err := common.FromHex(str)
	if err != nil {
		return nil, err
	}
	merkle := reverse(data)
	return merkle, nil
}

func reverse(h []byte) []byte {
	for i := 0; i < common.Sha256Len/2; i++ {
		h[i], h[common.Sha256Len-1-i] = h[common.Sha256Len-1-i], h[i]
	}
	return h
}

func (b *btcStore) getHeadHeightList(req *ty.ReqRelayBtcHeaderHeightList) (types.Message, error) {
	prefix := []byte(relayBTCHeaderHeightList)
	key := calcBtcHeaderKeyHeightList(req.ReqHeight)

	values, err := b.db.List(prefix, key, req.Counts, req.Direction)
	if err != nil {
		values, err = b.db.List(prefix, nil, req.Counts, req.Direction)
		if err != nil {
			return nil, err
		}
	}

	var replay ty.ReplyRelayBtcHeadHeightList
	heightGot := make(map[int64]bool)
	for _, heightByte := range values {
		height, _ := decodeHeight(heightByte)
		if !heightGot[height] {
			replay.Heights = append(replay.Heights, height)
			heightGot[height] = true
		}
	}

	return &replay, nil

}

func btcWireHeader(head *ty.BtcHeader) (*wire.BlockHeader, error) {
	preHash, err := chainhash.NewHashFromStr(head.PreviousHash)
	if err != nil {
		return nil, err
	}
	merkleRoot, err := chainhash.NewHashFromStr(head.MerkleRoot)
	if err != nil {
		return nil, err
	}

	h := &wire.BlockHeader{}
	h.Version = int32(head.Version)
	h.PrevBlock = *preHash
	h.MerkleRoot = *merkleRoot
	h.Bits = uint32(head.Bits)
	h.Nonce = uint32(head.Nonce)
	h.Timestamp = time.Unix(head.Time, 0)

	return h, nil
}

func verifyBlockHeader(head *ty.BtcHeader, preHead *ty.RelayLastRcvBtcHeader, localDb dbm.KVDB) error {
	if head == nil {
		return types.ErrInvalidParam
	}

	if preHead != nil && preHead.Header != nil && (preHead.Header.Hash != head.PreviousHash || preHead.Header.Height+1 != head.Height) && !head.IsReset {

		return ty.ErrRelayBtcHeadSequenceErr
	}

	//real BTC block not change the bits before height<30000, not match with the calculation result
	if !head.IsReset && head.Height > 30000 {
		newBits, err := calcNextRequiredDifficulty(preHead.Header, localDb)
		if err != nil && err != types.ErrNotFound {
			return err
		}

		if newBits != 0 && newBits != head.Bits {
			return ty.ErrRelayBtcHeadNewBitsErr
		}
	}

	btcHeader, err := btcWireHeader(head)
	if err != nil {
		return err
	}
	hash := btcHeader.BlockHash()

	if hash.String() != head.Hash {
		return ty.ErrRelayBtcHeadHashErr
	}

	target := difficulty.CompactToBig(uint32(head.Bits))

	// The block hash must be less than the claimed target.
	hashNum := difficulty.HashToBig(hash[:])
	if hashNum.Cmp(target) > 0 {
		return ty.ErrRelayBtcHeadBitsErr
	}

	return nil
}

// refer to btcd's blockchain's calcNextRequiredDifficulty() function
// calcNextRequiredDifficulty calculates the required difficulty for the block
// after the passed previous block node based on the difficulty retarget rules.
func calcNextRequiredDifficulty(preHead *ty.BtcHeader, localDb dbm.KVDB) (int64, error) {
	if preHead == nil {
		return 0, nil
	}

	// Genesis block.
	targetTimespan := time.Hour * 24 * 14  // 14 days
	TargetTimePerBlock := time.Minute * 10 // 10 minutes
	retargetAdjustmentFactor := int64(4)   // 25% less, 400% more
	timeSpan := int64(targetTimespan / time.Second)
	timeBlock := int64(TargetTimePerBlock / time.Second)

	// powLimit is the highest proof of work value a Bitcoin block
	// can have for the regression test network.  It is the value 2^255 - 1.
	bigOne := big.NewInt(1)
	powLimit := new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)
	blocksPerRetarget := uint64(timeSpan / timeBlock)
	minRetargetTimespan := timeSpan / retargetAdjustmentFactor
	maxRetargetTimespan := timeSpan * retargetAdjustmentFactor

	// Return the previous block's difficulty requirements if this block
	// is not at a difficulty retarget interval.
	if (preHead.Height+1)%blocksPerRetarget != 0 {
		// For networks that support it, allow special reduction of the
		// required difficulty once too much time has elapsed without
		// mining a block.

		// For the main network (or any unrecognized networks), simply
		// return the previous block's difficulty requirements.
		return preHead.Bits, nil
	}

	// Get the block node at the previous retarget (targetTimespan days
	// worth of blocks).
	btc := newBtcStore(localDb)
	firstHead, err := btc.getBtcHeadByHeight(int64(preHead.Height - (blocksPerRetarget - 1)))
	if err != nil {
		return 0, err
	}

	// Limit the amount of adjustment that can occur to the previous
	// difficulty.
	actualTimespan := preHead.Time - firstHead.Time
	adjustedTimespan := actualTimespan
	if actualTimespan < minRetargetTimespan {
		adjustedTimespan = minRetargetTimespan
	} else if actualTimespan > maxRetargetTimespan {
		adjustedTimespan = maxRetargetTimespan
	}

	// Calculate new target difficulty as:
	//  currentDifficulty * (adjustedTimespan / targetTimespan)
	// The result uses integer division which means it will be slightly
	// rounded down.  Bitcoind also uses integer division to calculate this
	// result.
	oldTarget := difficulty.CompactToBig(uint32(preHead.Bits))
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	newTarget.Div(newTarget, big.NewInt(timeSpan))

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(powLimit) > 0 {
		newTarget.Set(powLimit)
	}

	newTargetBits := difficulty.BigToCompact(newTarget)

	return int64(newTargetBits), nil
}
