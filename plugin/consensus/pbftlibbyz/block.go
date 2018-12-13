// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pbftlibbyz

import (
	"time"

	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	byzcli "github.com/33cn/libbyz-go/client"
	byzreplica "github.com/33cn/libbyz-go/replica"
)

func init() {
	drivers.Reg("pbftlibbyz", Newpbftlibbyz)
	drivers.QueryData.Register("pbftlibbyz", &Client{})
}

// Client pbftlibbyz implementation
type Client struct {
	*drivers.BaseClient
	isClient bool
}

// NewBlockstore create pbftlibbyz Client
func NewBlockstore(cfg *types.Consensus, isClient bool) *Client {
	c := drivers.NewBaseClient(cfg)
	client := &Client{BaseClient: c, isClient: isClient}
	c.SetChild(client)
	return client
}

// ProcEvent method
func (client *Client) ProcEvent(msg queue.Message) bool {
	return false
}

// ProposeAndReadReply method
func (client *Client) ProposeAndReadReply(block *types.Block) {
	byzcli.ByzRunClient()

	plog.Info("===============Get block from reply===========")
	lastBlock := client.GetCurrentBlock()
	err := client.WriteBlock(lastBlock.StateHash, block)

	if err != nil {
		plog.Error("********************err:", err)
		return
	}
	client.SetCurrentBlock(block)
}

// CheckBlock method
func (client *Client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	return nil
}

// SetQueueClient method
func (client *Client) SetQueueClient(c queue.Client) {
	plog.Info("Enter SetQueue method of pbftlibbyz consensus")
	client.InitClient(c, func() {

		client.InitBlock()
	})
	go client.EventLoop()
	//go client.readReply()
	go client.CreateBlock()
}

// CreateBlock method
func (client *Client) CreateBlock() {
	issleep := true
	// if !client.isPrimary {
	// 	return
	// }
	var config = "./config"
	var configPriv = "./priv_template"
	if client.isClient {
		byzcli.ByzInitClient(config, configPriv)
	} else {
		byzreplica.ByzInitReplica(config, configPriv)
		return
	}

	for {
		if issleep {
			time.Sleep(10 * time.Second)
		}
		plog.Info("=============start get tx===============")
		lastBlock := client.GetCurrentBlock()
		txs := client.RequestTx(int(types.GetP(lastBlock.Height+1).MaxTxNumber), nil)
		if len(txs) == 0 {
			issleep = true
			continue
		}
		issleep = false
		plog.Info("==================start create new block!=====================")
		//check dup
		//txs = client.CheckTxDup(txs)
		//fmt.Println(len(txs))

		var newblock types.Block
		newblock.ParentHash = lastBlock.Hash()
		newblock.Height = lastBlock.Height + 1
		newblock.Txs = txs
		newblock.TxHash = merkle.CalcMerkleRoot(newblock.Txs)
		newblock.BlockTime = types.Now().Unix()
		if lastBlock.BlockTime >= newblock.BlockTime {
			newblock.BlockTime = lastBlock.BlockTime + 1
		}
		client.ProposeAndReadReply(&newblock)
		//time.Sleep(time.Second)
		// client.readReply()
		plog.Info("===============readreply and writeblock done===============")
	}
}

// GetGenesisBlockTime get genesis blocktime
func (client *Client) GetGenesisBlockTime() int64 {
	return genesisBlockTime
}

// CreateGenesisTx get genesis tx
func (client *Client) CreateGenesisTx() (ret []*types.Transaction) {
	var tx types.Transaction
	tx.Execer = []byte("coins")
	tx.To = genesis
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{}
	g.Genesis.Amount = 1e8 * types.Coin
	tx.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx)
	return
}
