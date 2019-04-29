package pbft

import (
	"os"
	"time"

	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pbft/types"
	"github.com/golang/protobuf/proto"
)

func init() {
	drivers.Reg("pbft", NewPbftNode)
	drivers.QueryData.Register("pbft", &PbftNode{})
}

// PbftNode 是一个PBFT节点端(可以是客户端，参与共识)
type PbftNode struct {
	*drivers.BaseClient
	requestChan chan *pt.Request
	dataChan    chan *pt.BlockData
	isClient    bool
	address     string
}

// NewBlockstore 用于初始化新的Blockstore
func NewBlockstore(cfg *types.Consensus, requestChan chan *pt.Request, dataChan chan *pt.BlockData, isClient bool, address string) *PbftNode {
	c := drivers.NewBaseClient(cfg)
	client := &PbftNode{BaseClient: c, requestChan: requestChan, dataChan: dataChan, isClient: isClient, address: address}
	c.SetChild(client)
	return client
}

// ProcEvent 返回false
func (client *PbftNode) ProcEvent(msg *queue.Message) bool {
	return false
}

// CheckBlock 用于验证区块
func (client *PbftNode) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	return nil
}

// GetGenesisBlockTime 用于获取创世块时间戳
func (client *PbftNode) GetGenesisBlockTime() int64 {
	return genesisBlockTime
}

// CreateGenesisTx 用于产生创世交易
func (client *PbftNode) CreateGenesisTx() (ret []*types.Transaction) {
	var tx types.Transaction
	tx.Execer = []byte(cty.CoinsX)
	tx.To = genesis
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{}
	g.Genesis.Amount = 1e8 * types.Coin
	tx.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx)
	return
}

// SetQueueClient 用于初始化节点队列
func (client *PbftNode) SetQueueClient(c queue.Client) {
	plog.Info("Enter SetQueue method of typesft consensus")
	client.InitClient(c, func() {
		client.InitBlock()
	})
	go client.EventLoop()
	//go client.readReply()
	go client.CreateBlock()
}

/*
// Close 用于关闭这个节点，TODO
func (client *PbftClient) Close() {
	client.stopC <- struct{}{}
	plog.Info("consensus raft closed")
}
*/

// Propose 用于客户端发请求
func (client *PbftNode) Propose(block *types.Block) {
	op := &pt.BlockData{Value: block}
	req := ToRequestClient(op, time.Now().String(), client.address)
	client.requestChan <- req
}

// 用于客户端读请求
func (client *PbftNode) readReply() {

	data := <-client.dataChan
	if data == nil {
		plog.Error("block is nil")
		return
	}
	plog.Info("===========Client Get Reply ===========")
	/*
		flag := client.IsCaughtUp()
		if flag {
			return
		}
	*/
	lastBlock := client.GetCurrentBlock()
	err := client.WriteBlock(lastBlock.StateHash, data.Value)
	if err != nil {
		plog.Error("Block Write Error", err)
	}
	err = WriteSnap(data.Value, "typesft.log")
	if err != nil {
		plog.Error("Snap Write Error", err)
	}
	client.SetCurrentBlock(data.Value)
	plog.Info("===========Readreply and Writeblock done===========")

}

// CreateBlock 用于出块
func (client *PbftNode) CreateBlock() {
	issleep := true

	if !client.isClient {
		return
	}

	for {
		if issleep {
			time.Sleep(10 * time.Second)
		}
		plog.Info("===========Start get Txs===========")
		/*
			flag := client.IsCaughtUp()
			if flag {
				return
			}
		*/
		lastBlock := client.GetCurrentBlock()
		txs := client.RequestTx(int(types.GetP(lastBlock.Height+1).MaxTxNumber), nil)
		if len(txs) == 0 {
			issleep = true
			continue
		}
		issleep = false
		plog.Info("===========Start Create New Block!===========")
		//check dup
		//txs = client.CheckTxDup(txs)
		//fmt.Println(len(txs))

		var newblock types.Block
		newblock.StateHash = lastBlock.StateHash
		newblock.ParentHash = lastBlock.Hash()
		newblock.Height = lastBlock.Height + 1
		newblock.Txs = txs
		newblock.TxHash = merkle.CalcMerkleRoot(newblock.Txs)
		newblock.BlockTime = time.Now().Unix()
		if lastBlock.BlockTime >= newblock.BlockTime {
			newblock.BlockTime = lastBlock.BlockTime + 1
		}
		client.Propose(&newblock)
		client.readReply()
	}
}

// MesToByte 用于对protobuf类消息编码
func MesToByte(mes proto.Message) ([]byte, error) {
	return proto.Marshal(mes)
}

// WriteSnap write block to file
func WriteSnap(block proto.Message, snapdir string) error {
	blockbyte, err := proto.Marshal(block)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(snapdir, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	_, err = f.Write(blockbyte)
	f.Close()
	if err != nil {
		return err
	}
	return nil
}

// WriteLog write log to file
func WriteLog(data []*pt.Request, snapdir string) error {
	f, err := os.OpenFile(snapdir, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666) //os.ModeAppend
	if err != nil {
		return err
	}
	for _, REQ := range data {
		databyte, err := MesToByte(REQ)
		if err != nil {
			return err
		}
		_, err = f.Write(databyte)
		if err != nil {
			return err
		}
	}

	f.Close()

	return nil
}
