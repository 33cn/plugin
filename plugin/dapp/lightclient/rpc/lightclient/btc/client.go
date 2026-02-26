package btc

import (
	"context"
	"encoding/json"

	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"

	"github.com/33cn/chain33/client"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
)

func init() {
	lightclient.Register("btc", newClient)
}

type btcLight struct {
	ctx context.Context
	cli BtcClient
	cfg config
	api client.QueueProtocolAPI
}

func newClient() lightclient.Lighter {
	return &btcLight{}
}

func (b *btcLight) Init(ctx context.Context, q queue.Queue, cfg *lightclient.Config) error {

	b.ctx = ctx
	b.api, _ = client.New(q.Client(), nil)
	subCfg, _ := json.Marshal(cfg.Btc)
	types.MustDecode(subCfg, &b.cfg)

	return nil
}

func (b *btcLight) Start() {

}
