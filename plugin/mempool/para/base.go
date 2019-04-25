package para

import (
	"bytes"
	"context"
	"fmt"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
)

var mlog = log.New("module", "mempool.para")
var topic = "mempool"
var RETRY_TIMES = 3

//Mempool mempool 基础类
type Mempool struct {
	key         string
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
	go func() {
		client.Sub(mem.key)
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventTx:
				mlog.Info("Receive msg from para mempool")
				if bytes.HasPrefix(msg.GetData().(*types.Transaction).Execer, types.ParaKey) {
					tx := msg.GetData().(*types.Transaction)
					var reply *types.Reply
					reply, err := mem.mainGrpcCli.SendTransaction(context.Background(), tx)
					if err != nil {
						//进行重试
						for i := 0; i < RETRY_TIMES; i++ {
							reply, err = mem.mainGrpcCli.SendTransaction(context.Background(), tx)
							if err != nil {
								continue
							} else {
								break
							}
						}
						if err != nil {
							msg.Reply(client.NewMessage(mem.key, types.EventReply, &types.Reply{IsOk: false,
								Msg: []byte(fmt.Sprintf("Send transaction to main chain failed, %v", err))}))
							break
						}
					}
					msg.Reply(client.NewMessage(mem.key, types.EventReply, &types.Reply{IsOk: true, Msg: []byte(reply.GetMsg())}))
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
func (mem *Mempool) Close() {}
