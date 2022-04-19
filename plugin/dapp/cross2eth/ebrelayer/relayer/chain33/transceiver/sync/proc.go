package sync

import (
	"fmt"
	"math"
	"sync/atomic"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
)

// SeqType
const (
	SeqTypeAdd = int32(1)
	SeqTypeDel = int32(2)
)

var (
	syncLastHeight   = []byte("syncLastHeight:")
	evmTxLogPrefix   = []byte("evmTxLogPrefix:")
	lastSequences    = []byte("lastSequences:")
	seqOperationType = []string{"SeqTypeAdd", "SeqTypeDel"}
)

var evmTxLogsCh chan *types.EVMTxLogsInBlks
var resultCh chan error

func init() {
	evmTxLogsCh = make(chan *types.EVMTxLogsInBlks)
	resultCh = make(chan error)
}

func evmTxLogKey4Height(height int64) []byte {
	return append(evmTxLogPrefix, []byte(fmt.Sprintf("%012d", height))...)
}

// pushTxReceipts push block to backend
func pushTxReceipts(evmTxLogsInBlks *types.EVMTxLogsInBlks) error {
	evmTxLogsCh <- evmTxLogsInBlks
	err := <-resultCh
	return err
}

//EVMTxLogs ...
type EVMTxLogs struct {
	db     dbm.DB
	seqNum int64 //当前同步的序列号
	height int64 //当前区块高度
	quit   chan struct{}
}

//NewSyncTxReceipts ...
func NewSyncTxReceipts(db dbm.DB) *EVMTxLogs {
	sync := &EVMTxLogs{
		db: db,
	}
	sync.seqNum, _ = sync.loadBlockLastSequence()
	sync.height, _ = sync.LoadLastBlockHeight()
	sync.quit = make(chan struct{})
	sync.initSyncReceiptDataBase()

	return sync
}

//此处添加一个高度为0的空块，只是为了查找下一个比较方便，并不需要使用其信息
func (syncTx *EVMTxLogs) initSyncReceiptDataBase() {
	txLogs0, _ := syncTx.GetTxLogs(0)
	if nil != txLogs0 {
		return
	}
	logsPerBlock := &types.EVMTxLogPerBlk{
		Height: 0,
	}
	syncTx.setTxLogsPerBlock(logsPerBlock)
}

//Stop ...
func (syncTx *EVMTxLogs) Stop() {
	close(syncTx.quit)
}

// SaveAndSyncTxs2Relayer save block to db
func (syncTx *EVMTxLogs) SaveAndSyncTxs2Relayer() {
	for {
		select {
		case evmTxLogs := <-evmTxLogsCh:
			log.Info("to deal request", "seq", evmTxLogs.Logs4EVMPerBlk[0].SeqNum, "count", len(evmTxLogs.Logs4EVMPerBlk))
			syncTx.dealEVMTxLogs(evmTxLogs)
		case <-syncTx.quit:
			return
		}
	}
}

// 保存区块步骤
// 1. 记录 seqNumber ->  seq
// 2. 记录 lastseq
// 3. 更新高度
//
// 重启恢复
// 1. 看高度， 对应高度是已经完成的
// 2. 继续重新下一个高度即可。 重复写， 幂等
// 所以不需要恢复过程， 读出高度即可

// 处理输入流程
func (syncTx *EVMTxLogs) dealEVMTxLogs(evmTxLogsInBlks *types.EVMTxLogsInBlks) {
	count, start, evmTxLogsParsed := parseEvmTxLogsInBlks(evmTxLogsInBlks, syncTx.seqNum)
	txReceiptCount := len(evmTxLogsParsed)
	//重复注册推送接收保护，允许同一个中继服务在使用一段时间后，使用不同的推送名字重新进行注册，这样重复推送忽略就可以
	//需要进行ack，否则该节点的推送将会停止
	if 0 == txReceiptCount {
		resultCh <- nil
		return
	}

	var height int64
	for i := 0; i < txReceiptCount; i++ {
		txsPerBlock := evmTxLogsParsed[i]
		if txsPerBlock.AddDelType == SeqTypeAdd {
			syncTx.setTxLogsPerBlock(txsPerBlock)
			syncTx.setBlockLastSequence(txsPerBlock.SeqNum)
			syncTx.setBlockHeight(txsPerBlock.Height)
			height = txsPerBlock.Height
		} else {
			//删除分叉区块处理
			syncTx.delTxReceipts(txsPerBlock.Height)
			syncTx.setBlockLastSequence(txsPerBlock.SeqNum)
			height = txsPerBlock.Height - 1
			//删除区块不需要通知新的高度，因为这只会降低未处理区块的成熟度
			syncTx.setBlockHeight(height)
		}
	}
	//发送回复，确认接收成功
	resultCh <- nil
	resetTimer2KeepAlive()
	log.Debug("dealEVMTxLogs", "seqStart", start, "count", count, "maxBlockHeight", height, "syncTx.seqNum", syncTx.seqNum)
}

func (syncTx *EVMTxLogs) loadBlockLastSequence() (int64, error) {
	return utils.LoadInt64FromDB(lastSequences, syncTx.db)
}

//LoadLastBlockHeight ...
func (syncTx *EVMTxLogs) LoadLastBlockHeight() (int64, error) {
	return utils.LoadInt64FromDB(syncLastHeight, syncTx.db)
}

func (syncTx *EVMTxLogs) setBlockLastSequence(newSequence int64) {
	Sequencebytes := types.Encode(&types.Int64{Data: newSequence})
	if err := syncTx.db.SetSync(lastSequences, Sequencebytes); nil != err {
		panic("setBlockLastSequence failed due to cause:" + err.Error())
	}
	//同时更新内存中的seq
	syncTx.updateSequence(newSequence)
}

func (syncTx *EVMTxLogs) setBlockHeight(height int64) {
	bytes := types.Encode(&types.Int64{Data: height})
	_ = syncTx.db.SetSync(syncLastHeight, bytes)
	atomic.StoreInt64(&syncTx.height, height)
}

func (syncTx *EVMTxLogs) updateSequence(newSequence int64) {
	atomic.StoreInt64(&syncTx.seqNum, newSequence)
}

func (syncTx *EVMTxLogs) setTxLogsPerBlock(txLogs *types.EVMTxLogPerBlk) {
	key := evmTxLogKey4Height(txLogs.Height)
	value := types.Encode(txLogs)
	if err := syncTx.db.SetSync(key, value); nil != err {
		panic("setTxLogsPerBlock failed due to:" + err.Error())
	}
}

//GetTxReceipts ...
func (syncTx *EVMTxLogs) GetTxLogs(height int64) (*types.TxReceipts4SubscribePerBlk, error) {
	key := evmTxLogKey4Height(height)
	value, err := syncTx.db.Get(key)
	if err != nil {
		return nil, err
	}
	detail := &types.TxReceipts4SubscribePerBlk{}
	err = types.Decode(value, detail)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

//GetNextValidTxReceipts ...
func (syncTx *EVMTxLogs) GetNextValidEvmTxLogs(height int64) (*types.EVMTxLogPerBlk, error) {
	key := evmTxLogKey4Height(height)
	helper := dbm.NewListHelper(syncTx.db)
	evmTxLogs := helper.List(evmTxLogPrefix, key, 1, dbm.ListASC)
	if nil == evmTxLogs {
		return nil, nil
	}
	detail := &types.EVMTxLogPerBlk{}
	err := types.Decode(evmTxLogs[0], detail)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (syncTx *EVMTxLogs) delTxReceipts(height int64) {
	key := evmTxLogKey4Height(height)
	_ = syncTx.db.DeleteSync(key)
}

// 检查输入是否有问题, 并解析输入
func parseEvmTxLogsInBlks(evmTxLogs *types.EVMTxLogsInBlks, seqNumLast int64) (count int, start int64, txsWithReceipt []*types.EVMTxLogPerBlk) {
	count = len(evmTxLogs.Logs4EVMPerBlk)
	txsWithReceipt = make([]*types.EVMTxLogPerBlk, 0)
	start = math.MaxInt64
	for i := 0; i < count; i++ {
		if evmTxLogs.Logs4EVMPerBlk[i].AddDelType != SeqTypeAdd && evmTxLogs.Logs4EVMPerBlk[i].AddDelType != SeqTypeDel {
			log.Error("parseEvmTxLogsInBlks seq op not support", "seq", evmTxLogs.Logs4EVMPerBlk[i].SeqNum,
				"height", evmTxLogs.Logs4EVMPerBlk[i].Height, "seqOp", evmTxLogs.Logs4EVMPerBlk[i].AddDelType)
			continue
		}
		//过滤掉老的信息, 正常情况下，本次开始的的seq不能小于上次结束的seq
		if seqNumLast >= evmTxLogs.Logs4EVMPerBlk[i].SeqNum {
			log.Error("parseEvmTxLogsInBlks err: the tx and receipt pushed is old", "seqNumLast", seqNumLast,
				"evmTxLogs.Logs4EVMPerBlk[i].SeqNum", evmTxLogs.Logs4EVMPerBlk[i].SeqNum, "i", i)
			continue
		}
		txsWithReceipt = append(txsWithReceipt, evmTxLogs.Logs4EVMPerBlk[i])
		if evmTxLogs.Logs4EVMPerBlk[i].SeqNum < start {
			start = evmTxLogs.Logs4EVMPerBlk[i].SeqNum
		}
		log.Debug("parseEvmTxLogsInBlks get one block's tx with receipts", "seq", evmTxLogs.Logs4EVMPerBlk[i].SeqNum,
			"height", evmTxLogs.Logs4EVMPerBlk[i].Height, "seqOpType", seqOperationType[evmTxLogs.Logs4EVMPerBlk[i].AddDelType-1])

	}
	if 0 == len(txsWithReceipt) {
		log.Error("parseEvmTxLogsInBlks", "the valid number of tx receipt is", 0)
	}

	return
}
