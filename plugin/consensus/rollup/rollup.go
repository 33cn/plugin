package rollup

import (
	"context"
	"sync"
	"time"

	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"
	"github.com/golang/groupcache/lru"
)

const (
	minCommitTxCount = 128
)

var (
	rlog = log.New("module", "rollup")
)

func init() {
	consensus.RegCommitter("rollup", &RollUp{})
}

// RollUp roll up
type RollUp struct {
	nextBuildHeight int64
	nextBuildRound  int64
	cfg             Config
	op              *optimistic

	initDone        chan struct{}
	nextCommitRound int32

	currCommitter string

	commitCache *lru.Cache
	batchCache  sync.Map
	ctx         context.Context
	base        *consensus.BaseClient
	chainCfg    *types.Chain33Config
}

// Init init
func (r *RollUp) Init(base *consensus.BaseClient, chainCfg *types.Chain33Config, subCfg []byte) {

	if !chainCfg.IsPara() {
		return
	}

	r.chainCfg = chainCfg
	r.ctx = base.Context
	r.op = &optimistic{}
	r.initDone = make(chan struct{})

	go r.fetchRollupState()
	go r.startRollupRoutine()
}

func (r *RollUp) fetchRollupState() {

	r.initDone <- struct{}{}
}

func (r *RollUp) startRollupRoutine() {

	<-r.initDone

	go r.handleBuildBatch()
	go r.handleCommitBatch()
}

func (r *RollUp) handleBuildBatch() {

	ticker := time.NewTicker(time.Second * 10)
	var blocks []*types.Block
	for {

		select {
		case <-ticker.C:
			blocks = r.getNextBatchBlocks(r.nextBuildHeight)
		case <-r.ctx.Done():
			ticker.Stop()
			return
		}
		// 区块内未达到最低批量数量, 需要继续等待
		if blocks == nil {
			rlog.Debug("handleBuildBatch", "height", r.nextBuildHeight,
				"round", r.nextBuildRound, "msg", "wait more block")
			continue
		}

		batch := r.op.GetCommitBatch(blocks)
		r.batchCache.Store(batch.CommitRound, batch)
		r.nextBuildRound++
		r.nextBuildHeight += int64(len(blocks))
	}
}

// 提交共识
func (r *RollUp) handleCommitBatch() {

}

// 同步链上已提交的最新 blockHeight 和 commitRound, 维护batch缓存
func (r *RollUp) syncRollupState() {

}
