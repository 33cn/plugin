package para

import (
	"context"
	"sync/atomic"
	"time"

	"sync"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
)

var retry_times = 3
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
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventTx:
				mlog.Info("Receive msg from para mempool")
				tx := msg.GetData().(*types.Transaction)
				for i := 0; i < retry_times; i++ {
					reply, err := mem.mainGrpcCli.SendTransaction(context.Background(), tx)
					if err == nil {
						msg.Reply(client.NewMessage(mem.key, types.EventReply, &types.Reply{IsOk: true, Msg: reply.GetMsg()}))
						break
					} else if err != nil && i != retry_times-1 {
						time.Sleep(time.Millisecond * 10)
						continue
					} else {
						msg.Reply(client.NewMessage(mem.key, types.EventReply, err))
					}
				}
			default:
				msg.Reply(client.NewMessage(mem.key, types.EventReply, types.ErrActionNotSupport))
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
