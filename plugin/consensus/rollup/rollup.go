package rollup

import (
	"context"
	"sync"
	"time"

	"github.com/33cn/chain33/rpc/grpcclient"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"

	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"
)

const (
	minCommitTxCount          = 128
	eachValidatorCommitRounds = 10
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
	nextCommitRound int64

	currCommitter string

	batchCache           sync.Map
	signMsgCache         sync.Map
	ctx                  context.Context
	base                 *consensus.BaseClient
	chainCfg             *types.Chain33Config
	subChan              chan *rolluptypes.ValidatorSignMsg
	minBuildRoundInCache int64
	mainChainGrpc        types.Chain33Client
	val                  *validator
	cache                *batchCache
}

// Init init
func (r *RollUp) Init(base *consensus.BaseClient, chainCfg *types.Chain33Config, subCfg []byte) {

	if !chainCfg.IsPara() {
		return
	}

	types.MustDecode(subCfg, r.cfg)

	r.chainCfg = chainCfg
	r.ctx = base.Context
	r.op = &optimistic{}
	r.initDone = make(chan struct{})
	r.subChan = make(chan *rolluptypes.ValidatorSignMsg, 32)

	var err error
	r.mainChainGrpc, err = grpcclient.NewMainChainClient(chainCfg, chainCfg.GetModuleConfig().RPC.MainChainGrpcAddr)
	if err != nil {
		panic(err)
	}

	go r.fetchRollupState()
	go r.startRollupRoutine()
}

func (r *RollUp) initJob() {

	r.initDone <- struct{}{}
}

func (r *RollUp) fetchRollupState() {

}

func (r *RollUp) startRollupRoutine() {

	<-r.initDone

	if r.val.enable {

		go r.handleBuildBatch()
		go r.handleCommitBatch()
		go r.syncRollupState()
	}
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
		r.nextBuildRound++
		r.nextBuildHeight += int64(len(blocks))
		msg, sign := r.val.sign(batch.GetCommitRound(), batch.GetBatch())

		r.cache.addCommitBatch(batch)
		r.cache.addSignMsg(msg, sign)
		r.tryPubMsg(psValidatorSignTopic, types.Encode(sign))
	}
}

// 同步链上已提交的最新 blockHeight 和 commitRound, 维护batch缓存
func (r *RollUp) syncRollupState() {

}
