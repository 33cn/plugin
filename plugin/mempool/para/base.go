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

	isclose int32
}

//NewMempool 新建mempool 实例
func NewMempool(cfg *types.Mempool) *Mempool {
	pool := &Mempool{}
	pool.key = topic
	return pool
}

//SetQueueClient 初始化mempool模块
func (mem *Mempool) SetQueueClient(client queue.Client) {
	mem.client = client
	mem.client.Sub(mem.key)
	mem.setMainGrpcCli(client.GetConfig())
	mem.wg.Add(1)
	go func() {
		defer mem.wg.Done()
		for msg := range client.Recv() {
			var err error
			var reply interface{}
			switch msg.Ty {
			case types.EventTx:
				tx := msg.GetData().(*types.Transaction)
				reply, err = mem.mainGrpcCli.SendTransaction(context.Background(), tx)
				// 兼容rpc发送交易错误处理, 同常规mempool返回逻辑保持一致
				if err != nil {
					reply = &types.Reply{IsOk: false, Msg: []byte(err.Error())}
					err = nil
				}
			case types.EventAddDelayTx:
				dtx := msg.GetData().(*types.DelayTx)
				reply, err = mem.mainGrpcCli.SendDelayTransaction(context.Background(), dtx)
				// 兼容rpc接收返回时错误处理
				if err != nil {
					reply = &types.Reply{IsOk: false, Msg: []byte(err.Error())}
					err = nil
				}

			case types.EventGetProperFee:
				reply, err = mem.mainGrpcCli.GetProperFee(context.Background(), &types.ReqProperFee{})
			case types.EventGetMempoolSize:
				// 消息类型EventGetMempoolSize：获取mempool大小
				size := types.Conf(client.GetConfig(), "config.mempool").GInt("poolCacheSize")
				msg.Reply(mem.client.NewMessage("rpc", types.EventMempoolSize, &types.MempoolSize{Size: size}))
				continue
			default:
				msg.Reply(client.NewMessage(mem.key, types.EventReply, types.ErrActionNotSupport))
				continue
			}
			if err != nil {
				msg.Reply(client.NewMessage(mem.key, types.EventReply, err))
			} else {
				msg.Reply(client.NewMessage(mem.key, types.EventReply, reply))
			}
		}
	}()
}

func (mem *Mempool) setMainGrpcCli(cfg *types.Chain33Config) {
	if cfg != nil && cfg.IsPara() {
		grpcCli, err := grpcclient.NewMainChainClient(cfg, "")
		if err != nil {
			panic(err)
		}
		mem.mainGrpcCli = grpcCli
	}
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
