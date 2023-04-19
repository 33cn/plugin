package rollup

import (
	"context"
	"runtime"
	"time"

	"github.com/33cn/chain33/common/address"

	"github.com/33cn/chain33/queue"

	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"
)

const (
	minCommitTxCount  = 128
	minCommitBlkCount = 32
	// 两次提交最大间隔
	defaultMaxCommitInterval = 300 // seconds
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
	initFragIndex   int32
	cfg             Config

	initDone             chan struct{}
	client               queue.Client
	ctx                  context.Context
	cancel               context.CancelFunc
	base                 *consensus.BaseClient
	chainCfg             *types.Chain33Config
	subChan              chan *types.TopicData
	minBuildRoundInCache int64
	mainChainGrpc        types.Chain33Client
	val                  *validator
	cache                *commitCache
	cross                *crossTxHandler
	lastFeeRate          int64
}

// Init init
func (r *RollUp) Init(base *consensus.BaseClient, subCfg []byte) {

	chainCfg := base.GetAPI().GetConfig()
	if !chainCfg.IsPara() {
		return
	}

	types.MustDecode(subCfg, &r.cfg)

	if r.cfg.MaxCommitInterval < 60 {
		r.cfg.MaxCommitInterval = defaultMaxCommitInterval
	}

	if r.cfg.ReservedMainHeight <= 0 {
		r.cfg.ReservedMainHeight = defaultReservedMainHeight
	}

	if r.cfg.AuthAccount == "" && r.cfg.AuthKey == "" {
		rlog.Info("info", "addr", r.cfg.AuthAccount)
		panic("rollup must config authAccount")
	}

	r.chainCfg = chainCfg
	r.ctx, r.cancel = context.WithCancel(base.Context)
	r.initDone = make(chan struct{})
	r.subChan = make(chan *types.TopicData, 32)
	r.lastFeeRate = 100000
	r.cross = &crossTxHandler{}
	r.base = base
	r.client = base.GetQueueClient()
	r.val = &validator{}

	var err error
	r.mainChainGrpc, err = grpcclient.NewMainChainClient(chainCfg, "")
	if err != nil {
		panic("init main chain grpc client err:" + err.Error())
	}

	go r.initJob()
	go r.startRollupRoutine()
}

func (r *RollUp) getKeyFromWallet(addr string) string {

	for {

		time.Sleep(2 * time.Second)
		resp, err := r.base.GetAPI().ExecWalletFunc("wallet", "GetWalletStatus", &types.ReqNil{})
		if err != nil {
			rlog.Error("getKeyFromWallet", "GetWalletStatus err", err)
			continue
		}
		if !resp.(*types.WalletStatus).GetIsHasSeed() {
			rlog.Info("getKeyFromWallet wait wallet save seed...")
			continue
		}

		if resp.(*types.WalletStatus).GetIsWalletLock() {
			rlog.Info("getKeyFromWallet wait wallet unlock...")
			continue
		}

		resp, err = r.base.GetAPI().ExecWalletFunc("wallet", "DumpPrivkey", &types.ReqString{Data: addr})
		if err != nil {
			rlog.Info("getKeyFromWallet", "addr", addr, "wait import key err", err)
			continue
		}
		return resp.(*types.ReplyString).Data
	}

}

func (r *RollUp) initJob() {

	// 等待获取钱包私钥
	authKey := r.cfg.AuthKey

	if authKey == "" {
		var err error
		r.cfg.AddressID, err = address.GetAddressType(r.cfg.AuthAccount)
		if err != nil {
			panic("invalid address type for authAccount config, " + r.cfg.AuthAccount)
		}
		authKey = r.getKeyFromWallet(r.cfg.AuthAccount)
	}
	valPubs := r.getValidatorPubKeys()
	status := r.getRollupStatus()
	for len(valPubs.GetBlsPubs()) == 0 || status == nil {
		rlog.Warn("Init rollup wait 5 seconds...", "status", status, "valPubs", valPubs)
		time.Sleep(5 * time.Second)
		valPubs = r.getValidatorPubKeys()
		status = r.getRollupStatus()
	}
	// 等待同步
	for !r.isChainSync() {
		rlog.Info("Init rollup wait chain sync block...")
		time.Sleep(5 * time.Second)
	}

	rlog.Info("Init rollup validator start...")
	r.val.init(authKey, valPubs, status)
	rlog.Info("Init rollup validator stop...")
	r.nextBuildRound = status.CommitRound + 1
	r.initFragIndex = status.BlockFragIndex
	// 初始提交从高度1开始
	r.nextBuildHeight = status.CommitBlockHeight + 1
	// 最近一个区块被分段提交, 需要处理剩余数据, 构建高度不变
	if status.BlockFragIndex > 0 {
		r.nextBuildHeight = status.CommitBlockHeight
	}
	r.cache = newCommitCache(status.CommitRound)
	r.cross.init(r, status)
	r.trySubTopic(psValidatorSignTopic)
	r.initDone <- struct{}{}

}

func (r *RollUp) startRollupRoutine() {

	<-r.initDone

	if r.val.enable {
		go r.handleExit()
		go r.handleBuildBatch()
		go r.handleCommit()
		go r.syncRollupState()
		go r.cross.pullCrossTx()

		n := runtime.NumCPU()

		for i := 0; i < n; i++ {

			go r.handleSubMsg()
		}
	}
}

func (r *RollUp) handleExit() {

	for {

		select {
		case <-r.ctx.Done():
			return
		case <-r.val.exit:
			r.cancel()
			rlog.Info("rollup exit")
			return
		}
	}
}

// 同步链上已提交的最新 blockHeight 和 commitRound, 维护batch缓存
func (r *RollUp) syncRollupState() {

	ticker := time.NewTicker(5 * time.Second)

	for {

		select {
		case <-ticker.C:
			valPubs := r.getValidatorPubKeys()
			status := r.getRollupStatus()

			if len(valPubs.GetBlsPubs()) > 0 {
				r.val.updateValidators(valPubs)
			}

			if status != nil {
				r.val.updateRollupStatus(status)
				r.cache.cleanHistory(status.CommitRound)
			}

		case <-r.ctx.Done():
			ticker.Stop()
			return
		}

	}

}
