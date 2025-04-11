package executor

import (
	"math/big"
	"time"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/difficulty"
	ty "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

const (
	BtcTargetTimespan           = time.Hour * 24 * 14 / time.Second // 14 days
	BtcTargetTimePerBlock       = time.Minute * 10 / time.Second    // 10 minutes
	BtcRetargetAdjustmentFactor = 4                                 // 25% less, 400% more
	BtcBlocksPerRetarget        = BtcTargetTimespan / BtcTargetTimePerBlock
)

// refer to btcd's blockchain's calcNextRequiredDifficulty function
// calculates the required difficulty for the block
// after the passed previous block node based on the difficulty retarget rules.
func calcNextRequiredDifficulty(currTimestamp, currBits int64, currHeight uint64, kdb dbm.KV) int64 {

	// Return the previous block's difficulty requirements if this block
	// is not at a difficulty retarget interval.
	if (currHeight+1)%uint64(BtcBlocksPerRetarget) != 0 {

		// For the main network (or any unrecognized networks), simply
		// return the previous block's difficulty requirements.
		return currBits
	}

	timeSpan := BtcTargetTimespan

	// powLimit is the highest proof of work value a Bitcoin block
	// can have for the regression test network.  It is the value 2^255 - 1.
	bigOne := big.NewInt(1)
	powLimit := new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	minRetargetTimespan := int64(timeSpan / BtcRetargetAdjustmentFactor)
	maxRetargetTimespan := int64(timeSpan * BtcRetargetAdjustmentFactor)

	// Get the block node at the previous retarget (targetTimespan days
	// worth of blocks)
	prevHead, err := getBtcHeader(kdb, currHeight-uint64(BtcBlocksPerRetarget)-1)
	if err != nil {
		return 0
	}

	// Limit the amount of adjustment that can occur to the previous
	// difficulty.
	adjustedTimespan := currTimestamp - prevHead.Time
	if adjustedTimespan < minRetargetTimespan {
		adjustedTimespan = minRetargetTimespan
	} else if adjustedTimespan > maxRetargetTimespan {
		adjustedTimespan = maxRetargetTimespan
	}

	// Calculate new target difficulty as:
	//  currentDifficulty * (adjustedTimespan / targetTimespan)
	// The result uses integer division which means it will be slightly
	// rounded down.  Bitcoind also uses integer division to calculate this
	// result.
	oldTarget := difficulty.CompactToBig(uint32(currBits))
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	newTarget.Div(newTarget, big.NewInt(int64(timeSpan)))

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(powLimit) > 0 {
		newTarget.Set(powLimit)
	}

	newTargetBits := difficulty.BigToCompact(newTarget)

	return int64(newTargetBits)
}

func toWireHeader(head *ty.BtcHeader) (*wire.BlockHeader, error) {
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
