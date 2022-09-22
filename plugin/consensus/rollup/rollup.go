package rollup

import (
	"context"
	"runtime"
	"time"

	"github.com/33cn/chain33/rpc/grpcclient"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"

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

	initDone chan struct{}

	ctx                  context.Context
	base                 *consensus.BaseClient
	chainCfg             *types.Chain33Config
	subChan              chan *types.TopicData
	minBuildRoundInCache int64
	mainChainGrpc        types.Chain33Client
	val                  *validator
	cache                *batchCache
	validatorUpdate      chan struct{}
}

// Init init
func (r *RollUp) Init(base *consensus.BaseClient, chainCfg *types.Chain33Config, subCfg []byte) {

	if !chainCfg.IsPara() {
		return
	}

	types.MustDecode(subCfg, r.cfg)

	r.chainCfg = chainCfg
	r.ctx = base.Context
	r.initDone = make(chan struct{})
	r.subChan = make(chan *types.TopicData, 32)

	var err error
	r.mainChainGrpc, err = grpcclient.NewMainChainClient(chainCfg, chainCfg.GetModuleConfig().RPC.MainChainGrpcAddr)
	if err != nil {
		panic("init main chain grpc client err:" + err.Error())
	}

	go r.initJob()
	go r.startRollupRoutine()
}

func (r *RollUp) initJob() {

	valPubs := r.getValidatorPubKeys()
	status := r.getRollupStatus()
	for len(valPubs) == 0 || status == nil {
		rlog.Error("initJob", "status", status, "valPubs", valPubs)
		time.Sleep(time.Second)
		valPubs = r.getValidatorPubKeys()
		status = r.getRollupStatus()
	}
	val := &validator{}
	val.init(r.cfg, valPubs, status)
	r.nextBuildRound = status.CommitRound + 1
	r.nextBuildHeight = status.CommitBlockHeight + 1
	r.cache = newCache(status.CommitRound)
	r.trySubTopic(psValidatorSignTopic)
	r.initDone <- struct{}{}
}

func (r *RollUp) startRollupRoutine() {

	<-r.initDone

	if r.val.enable {

		go r.handleBuildBatch()
		go r.handleCommitBatch()
		go r.syncRollupState()

		n := runtime.NumCPU()

		for i := 0; i < n; i++ {

			go r.handleSubMsg()
		}
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
		blkBatch := r.GetCommitBatch(blocks)
		batch := &rtypes.CommitBatch{
			ChainTitle:  r.chainCfg.GetTitle(),
			CommitRound: r.nextBuildRound,
			Batch:       blkBatch,
		}

		r.nextBuildRound++
		r.nextBuildHeight += int64(len(blocks))
		sign := r.val.sign(batch.GetCommitRound(), batch.GetBatch())

		r.cache.addCommitBatch(batch)
		r.cache.addValidatorSign(true, sign)
		r.tryPubMsg(psValidatorSignTopic, types.Encode(sign))
	}
}

// 同步链上已提交的最新 blockHeight 和 commitRound, 维护batch缓存
func (r *RollUp) syncRollupState() {

	ticker := time.NewTicker(time.Minute)

	for {

		select {
		case <-ticker.C:
			valPubs := r.getValidatorPubKeys()
			status := r.getRollupStatus()

			if len(valPubs) > 0 {
				r.val.updateValidators(valPubs)
			}

			if status != nil {
				r.val.updateRollupStatus(status)
				r.cache.remove(status.CommitRound)
			}

		case <-r.ctx.Done():
			ticker.Stop()
			return
		}

	}

}
