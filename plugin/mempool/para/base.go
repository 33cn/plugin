package para

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

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
				mlog.Info("Receive msg from para mempool")
				tx := msg.GetData().(*types.Transaction)
				if tx.GetSignature().GetTy() == types.EncodeSignID(types.SECP256K1ETH, 2) {
					//检查NONCE
					if !strings.Contains(string(tx.Execer), "evm") {
						continue
					}
					ecli, err := ethclient.Dial("http://localhost:8546")
					if err != nil {
						log.Error("plugin SetQueueClient", "err", err)
						continue
					}
					nonce, err := ecli.NonceAt(context.Background(), common.HexToAddress(tx.From()), nil)
					if err != nil {
						log.Error("NonceAt SetQueueClient", "err", err)
						continue
					}
					if tx.GetNonce() < int64(nonce) {
						//return fmt.Errorf("nonce too low")
						msg.Reply(mem.client.NewMessage("rpc", types.EventTx, &types.Reply{IsOk: false, Msg: []byte("nonce too low.")}))
					}

					if tx.GetNonce() > int64(nonce) {
						//return &types.Reply{IsOk: false,Msg: []byte{"nonce too height."}} ,fmt.Errorf("nonce too height")
						msg.Reply(mem.client.NewMessage("rpc", types.EventTx, &types.Reply{IsOk: false, Msg: []byte("nonce too height.")}))
					}
				}

				reply, err = mem.mainGrpcCli.SendTransaction(context.Background(), tx)
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
