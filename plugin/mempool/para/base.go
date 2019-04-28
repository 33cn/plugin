package para

import (
	"bytes"
	"context"
	"fmt"

	"sync"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
)

var mlog = log.New("module", "mempool.para")
var topic = "mempool"
var retry_times = 3

//Mempool mempool 基础类
type Mempool struct {
	key         string
	wg          sync.WaitGroup
	mainGrpcCli types.Chain33Client
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
	mem.wg.Add(1)
	go func() {
		client.Sub(mem.key)
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventTx:
				mlog.Info("Receive msg from para mempool")
				if bytes.HasPrefix(msg.GetData().(*types.Transaction).Execer, types.ParaKey) {
					tx := msg.GetData().(*types.Transaction)
					for i := 0; i < retry_times; i++ {
						reply, err := mem.mainGrpcCli.SendTransaction(context.Background(), tx)
						if err == nil {
							msg.Reply(client.NewMessage(mem.key, types.EventReply, &types.Reply{IsOk: true, Msg: reply.GetMsg()}))
							break
						} else if err != nil && i != retry_times-1 {
							continue
						} else {
							msg.Reply(client.NewMessage(mem.key, types.EventReply, &types.Reply{IsOk: false,
								Msg: []byte(fmt.Sprintf("Send transaction to main chain failed, %v", err))}))
						}
					}
				}
			default:
				msg.Reply(client.NewMessage(mem.key, types.EventReply, &types.Reply{IsOk: false,
					Msg: []byte(fmt.Sprintf("para %v doesn't handle message %v", mem.key, msg.Ty))}))
			}
		}
	}()
}

// Wait for ready
func (mem *Mempool) Wait() {}

// Close method
func (mem *Mempool) Close() {
	// wait for cycle quit
	mem.wg.Wait()
}
