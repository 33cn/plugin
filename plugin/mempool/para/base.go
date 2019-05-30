package para

import (
	"context"
	"sync/atomic"

	"sync"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
)

var mlog = log.New("module", "mempool.para")
var topic = "mempool"

//Mempool mempool 基础类
type Mempool struct {
	key         string
	wg          sync.WaitGroup
	client      queue.Client
	mainGrpcCli types.Chain33Client
	isclose     int32
}

//NewMempool 新建mempool 实例
func NewMempool(cfg *types.Mempool) *Mempool {
	pool := &Mempool{}
	pool.key = topic
	if types.IsPara() {
		grpcCli, err := grpcclient.NewMainChainClient("")
		if err != nil {
			panic(err)
		}
		pool.mainGrpcCli = grpcCli
	}

	return pool
}

//SetQueueClient 初始化mempool模块
func (mem *Mempool) SetQueueClient(client queue.Client) {
	mem.client = client
	mem.client.Sub(mem.key)
	mem.wg.Add(1)
	go func() {
		defer mem.wg.Done()
		for msg := range client.Recv() {
			var err error
			var reply interface{}
			switch msg.Ty {
			case types.EventTx:
				mlog.Info("Receive msg from para mempool")
				tx := msg.GetData().(*types.Transaction)
				reply, err = mem.mainGrpcCli.SendTransaction(context.Background(), tx)
			case types.EventGetProperFee:
				reply, err = mem.mainGrpcCli.GetProperFee(context.Background(), &types.ReqNil{})
			default:
				msg.Reply(client.NewMessage(mem.key, types.EventReply, types.ErrActionNotSupport))
			}
			if err != nil {
				msg.Reply(client.NewMessage(mem.key, types.EventReply, err))
			} else {
				msg.Reply(client.NewMessage(mem.key, types.EventReply, reply))
			}
		}
	}()
}

// Wait for ready
func (mem *Mempool) Wait() {}

// Close method
func (mem *Mempool) Close() {
	if !atomic.CompareAndSwapInt32(&mem.isclose, 0, 1) {
		return
	}
	if mem.client != nil {
		mem.client.Close()
	}
	// wait for cycle quit
	mlog.Info("para mempool module closing")
	mem.wg.Wait()
	mlog.Info("para mempool module closed")
}
