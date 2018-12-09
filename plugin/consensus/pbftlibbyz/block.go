// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pbftlibbyz

// #cgo CFLAGS: -I./bft/gmp -I./bft/libbyz -I./bft/sfs/include/sfslite -O3 -fno-exceptions -DNDEBUG
// #cgo LDFLAGS: -L./bft/gmp -L./bft/libbyz -L./bft/sfs/lib/sfslite -lbyz -lsfscrypt -lasync -lgmp -lstdc++
// #include<stdio.h>
// #include<stdlib.h>
// #include<string.h>
// #include<signal.h>
// #include<unistd.h>
// #include<sys/param.h>
// #include"libbyz.h"
// int exec_command_cgo(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, int client, bool ro);
// void dump_handler();
// typedef int (*service)(Byz_req *inb, Byz_rep *outb, Byz_buffer *non_det, int client, bool ro);
import "C"
import (
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	"time"
	"unsafe"
)

const Simple_size int = 4096

var option = 0

func init() {
	drivers.Reg("pbftlibbyz", NewPbftlibbyz)
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

// Propose and set the block method
func (client *Client) ProposeAndReadReply(block *types.Block) {
	read_only := 0
	req := C.struct__Byz_buffer{}
	rep := C.struct__Byz_buffer{}

	C.Byz_alloc_request(&req, C.int(Simple_size))
	if req.size < C.int(Simple_size) {
		plog.Error("Request is too big") // ???
	}
	for i := 0; i < Simple_size; i++ {
		*(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(req.contents)) + uintptr(i))) = C.char(option)
	}
	if option != 2 {
		req.size = 8
	} else {
		req.size = C.int(Simple_size)
	}

	// invoke request and get reply
	C.Byz_invoke(&req, &rep, C.ulong(read_only))

	// check reply
	if !(((option == 2 || option == 0) && rep.size == 8) || (option == 1 && rep.size == C.int(Simple_size))) {
		plog.Error("Invalid reply")
	}

	// free reply
	C.Byz_free_reply(&rep)

	C.Byz_free_request(&req)

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
	var config string = "./bft/config"
	// fmt.Println(config)
	var config_priv string = "./bft/config_private/template"
	var port int = 0
	c_config := C.CString(config)
	c_config_priv := C.CString(config_priv)
	defer C.free(unsafe.Pointer(c_config))
	defer C.free(unsafe.Pointer(c_config_priv))

	if client.isClient {
		C.Byz_init_client(c_config, c_config_priv, C.short(port))
	} else {
		C.dump_handler()

		var mem_size int = 205 * 8192
		c_mem := (*C.char)(C.malloc(C.ulong(mem_size)))
		for i := 0; i < mem_size; i++ {
			*(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(c_mem)) + uintptr(i))) = C.char(0)
		}
		defer C.free(unsafe.Pointer(c_mem))

		C.Byz_init_replica(c_config, c_config_priv, c_mem, C.uint(mem_size), (C.service)(unsafe.Pointer(C.exec_command_cgo)), nil, 0)
		C.Byz_replica_run()
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
