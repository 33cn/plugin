package executor

import (
	"time"

	dbm "github.com/33cn/chain33/common/db"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type btcChainContext struct {
	params *chaincfg.Params
}

func newBtcChainContext(params *chaincfg.Params) *btcChainContext {
	return &btcChainContext{params: params}
}

func (c *btcChainContext) ChainParams() *chaincfg.Params {
	return c.params
}

func (c *btcChainContext) BlocksPerRetarget() int32 {
	return int32(c.params.TargetTimespan / c.params.TargetTimePerBlock)
}

func (c *btcChainContext) MinRetargetTimespan() int64 {
	target := int64(c.params.TargetTimespan / time.Second)
	return target / c.params.RetargetAdjustmentFactor
}

func (c *btcChainContext) MaxRetargetTimespan() int64 {
	target := int64(c.params.TargetTimespan / time.Second)
	return target * c.params.RetargetAdjustmentFactor
}

func (c *btcChainContext) VerifyCheckpoint(_ int32, _ *chainhash.Hash) bool {
	return true
}

func (c *btcChainContext) FindPreviousCheckpoint() (blockchain.HeaderCtx, error) {
	return nil, nil
}

type btcHeaderContext struct {
	header   *ltypes.BtcHeader
	parent   blockchain.HeaderCtx
	localDB  dbm.KV
	ancestor map[uint64]blockchain.HeaderCtx
}

func newBtcHeaderContext(header *ltypes.BtcHeader, parent blockchain.HeaderCtx, localDB dbm.KV) *btcHeaderContext {
	ctx := &btcHeaderContext{
		header:   header,
		parent:   parent,
		localDB:  localDB,
		ancestor: make(map[uint64]blockchain.HeaderCtx),
	}
	ctx.ancestor[header.GetHeight()] = ctx
	return ctx
}

func (h *btcHeaderContext) Height() int32 {
	return int32(h.header.GetHeight())
}

func (h *btcHeaderContext) Bits() uint32 {
	return uint32(h.header.GetBits())
}

func (h *btcHeaderContext) Timestamp() int64 {
	return h.header.GetTime()
}

func (h *btcHeaderContext) Parent() blockchain.HeaderCtx {
	return h.parent
}

func (h *btcHeaderContext) RelativeAncestorCtx(distance int32) blockchain.HeaderCtx {
	if distance <= 0 {
		return h
	}

	curr := blockchain.HeaderCtx(h)
	for i := int32(0); i < distance && curr != nil; i++ {
		curr = curr.Parent()
	}
	if curr != nil {
		if ancestor, ok := curr.(*btcHeaderContext); ok {
			h.ancestor[ancestor.header.GetHeight()] = ancestor
		}
		return curr
	}

	targetHeight := int64(h.header.GetHeight()) - int64(distance)
	if targetHeight < 0 {
		return nil
	}
	if ancestor, ok := h.ancestor[uint64(targetHeight)]; ok {
		return ancestor
	}
	if h.localDB == nil {
		return nil
	}
	targetHeader, err := getBtcHeader(h.localDB, uint64(targetHeight))
	if err != nil {
		return nil
	}
	ancestor := newBtcHeaderContext(targetHeader, nil, h.localDB)
	h.ancestor[targetHeader.GetHeight()] = ancestor
	return ancestor
}

func toWireHeader(head *ltypes.BtcHeader) (*wire.BlockHeader, error) {
	preHash, err := chainhash.NewHashFromStr(head.GetPreviousHash())
	if err != nil {
		return nil, err
	}
	merkleRoot, err := chainhash.NewHashFromStr(head.GetMerkleRoot())
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
