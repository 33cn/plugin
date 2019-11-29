package pos33

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	driver "github.com/33cn/chain33/system/dapp"
	ct "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// Client is the pos33 consensus client
type Client struct {
	*drivers.BaseClient
	conf *subConfig
	n    *node

	tickLock   sync.Mutex
	ticketsMap map[string]*pt.Pos33Ticket
	privLock   sync.Mutex
	privmap    map[string]crypto.PrivKey
}

// Tx is ...
type Tx = types.Transaction

type genesisTicket struct {
	MinerAddr  string `json:"minerAddr"`
	ReturnAddr string `json:"returnAddr"`
	Count      int32  `json:"count"`
}

type subConfig struct {
	Genesis          []*genesisTicket `json:"genesis"`
	GenesisBlockTime int64            `json:"genesisBlockTime"`
	ListenAddr       string           `json:"Pos33ListenAddr,omitempty"`
	AdvertiseAddr    string           `json:"Pos33AdvertiseAddr,omitempty"`
	BootPeerAddr     string           `json:"Pos33BootPeerAddr,omitempty"`
	MaxTxs           int64            `json:"Pos33MaxTxs,omitempty"`
	BlockTime        int64            `json:"Pos33BlockTime,omitempty"`
	BlockTimeout     int64            `json:"Pos33BlockTimeout,omitempty"`
	MinFee           int64            `json:"Pos33MinFee,omitempty"`
	NodeAddr         string           `json:"nodeAddr,omitempty"`
}

// New create pos33 consensus client
func New(cfg *types.Consensus, sub []byte) queue.Module {
	c := drivers.NewBaseClient(cfg)
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}

	n := newNode(&subcfg)
	client := &Client{BaseClient: c, n: n, conf: &subcfg}
	client.n.Client = client
	c.SetChild(client)
	return client
}

// Close is close the client
func (client *Client) Close() {}

// ProcEvent do nothing?
func (client *Client) ProcEvent(msg *queue.Message) bool {
	return false
}

func (client *Client) newBlock(lastBlock *types.Block, txs []*types.Transaction, height int64) (*types.Block, error) {
	if lastBlock.Height+1 != height {
		plog.Error("newBlock height error", "lastHeight", lastBlock.Height, "height", height)
		return nil, fmt.Errorf("the last block too low")
	}

	ch := make(chan []*Tx, 1)
	go func() { ch <- client.RequestTx(int(client.conf.MaxTxs), nil) }()
	select {
	case <-time.After(time.Millisecond * 300):
	case ts := <-ch:
		txs = append(txs, ts...)
	}

	bt := time.Now().Unix()
	return &types.Block{
		ParentHash: lastBlock.Hash(client.GetAPI().GetConfig()),
		Height:     lastBlock.Height + 1,
		Txs:        txs,
		TxHash:     merkle.CalcMerkleRoot(txs),
		BlockTime:  bt,
	}, nil
}

// CheckBlock check block callback
func (client *Client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	return client.n.checkBlock(current.Block, parent)
}

func (client *Client) allWeight(height int64) int {
	msg, err := client.GetAPI().Query(pt.Pos33TicketX, "Pos33AllTicketCount", &pt.Pos33AllPos33TicketCount{Height: height})
	if err != nil {
		plog.Info("Pos33AllTicketCount error", "error", err)
		return 0
	}
	return int(msg.(*pt.ReplyPos33AllPos33TicketCount).Count)
}

func (client *Client) privFromBytes(privkey []byte) (crypto.PrivKey, error) {
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		return nil, err
	}
	return cr.PrivKeyFromBytes(privkey)
}

func getPrivMap(privs []crypto.PrivKey) map[string]crypto.PrivKey {
	list := make(map[string]crypto.PrivKey)
	for _, priv := range privs {
		addr := address.PubKeyToAddress(priv.PubKey().Bytes()).String()
		list[addr] = priv
	}
	return list
}

func (client *Client) setTicket(tlist *pt.ReplyPos33TicketList, privmap map[string]crypto.PrivKey) {
	client.ticketsMap = make(map[string]*pt.Pos33Ticket)
	if tlist == nil || privmap == nil {
		client.ticketsMap = nil
		client.privmap = nil
		return
	}
	for _, ticket := range tlist.Tickets {
		client.tickLock.Lock()
		client.ticketsMap[ticket.GetTicketId()] = ticket
		client.tickLock.Unlock()
		_, ok := privmap[ticket.MinerAddress]
		if !ok {
			delete(privmap, ticket.MinerAddress)
		}
	}

	client.privLock.Lock()
	client.privmap = privmap
	client.privLock.Unlock()
	plog.Debug("setTicket", "n", len(tlist.GetTickets()))
}

func (client *Client) flushTicket() error {
	//list accounts
	tickets, privs, err := client.getTickets()
	if err == types.ErrWalletIsLocked || err == pt.ErrNoPos33Ticket {
		plog.Error("flushTicket error", "err", err.Error())
		client.setTicket(nil, nil)
		return nil
	}
	if err != nil {
		plog.Error("flushTicket error", "err", err)
		return err
	}
	privMap := getPrivMap(privs)
	client.setTicket(&pt.ReplyPos33TicketList{Tickets: tickets}, privMap)
	return nil
}

func (client *Client) getTickets() ([]*pt.Pos33Ticket, []crypto.PrivKey, error) {
	resp, err := client.GetAPI().ExecWalletFunc("pos33", "WalletGetTickets", &types.ReqNil{})
	if err != nil {
		return nil, nil, err
	}
	reply := resp.(*pt.ReplyWalletPos33Tickets)
	var keys []crypto.PrivKey
	for i := 0; i < len(reply.Privkeys); i++ {
		priv, err := client.privFromBytes(reply.Privkeys[i])
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, priv)
	}
	plog.Info("getTickets", "ticket n", len(reply.Tickets), "nkey", len(keys))
	return reply.Tickets, keys, nil
}

func (client *Client) getAllWeight(height int64) int {
	preH := height - height%pt.Pos33SortitionSize
	if preH == height {
		preH -= pt.Pos33SortitionSize
	}
	key := []byte(pt.Pos33AllPos33TicketCountKeyPrefix + fmt.Sprintf("%d", preH))
	v, err := client.Get(key)
	if err != nil {
		plog.Error(err.Error())
		return 0
	}

	allw, err := strconv.Atoi(string(v))
	if err != nil {
		plog.Error(err.Error())
		return 0
	}

	return int(allw)
}

// AddBlock notice driver a new block incoming
func (client *Client) AddBlock(b *types.Block) error {
	client.n.addBlock(b)
	return nil
}

func (client *Client) nodeID() string {
	return client.conf.NodeAddr
}

func (client *Client) miningOK() bool {
	if !client.IsMining() || !(client.IsCaughtUp() || client.Cfg.ForceMining) {
		plog.Info("createblock.ismining is disable or client is caughtup is false")
		return false
	}
	ok := false
	client.tickLock.Lock()
	if len(client.ticketsMap) == 0 {
		plog.Info("your ticket count is 0, you MUST buy some ticket to start mining")
	} else {
		ok = true
	}
	client.tickLock.Unlock()
	return ok
}

// CreateBlock will start run
func (client *Client) CreateBlock() {
	for {
		if !client.miningOK() {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	client.n.runLoop()
}

//316190000 coins
func createTicket(cfg *types.Chain33Config, minerAddr, returnAddr string, count int32, height int64) (ret []*types.Transaction) {
	//给hotkey 10000 个币，作为miner的手续费
	tx1 := types.Transaction{}
	tx1.Execer = []byte("coins")
	tx1.To = minerAddr
	g := &ct.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{Amount: pt.GetPos33TicketMinerParam(cfg, height).Pos33TicketPrice}
	tx1.Payload = types.Encode(&ct.CoinsAction{Value: g, Ty: ct.CoinsActionGenesis})
	ret = append(ret, &tx1)

	// 发行并抵押
	tx2 := types.Transaction{}
	tx2.Execer = []byte("coins")
	tx2.To = driver.ExecAddress(pt.Pos33TicketX)
	g = &ct.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{Amount: int64(count) * pt.GetPos33TicketMinerParam(cfg, height).Pos33TicketPrice, ReturnAddress: returnAddr}
	tx2.Payload = types.Encode(&ct.CoinsAction{Value: g, Ty: ct.CoinsActionGenesis})
	ret = append(ret, &tx2)

	// 冻结资金并开启挖矿
	tx3 := types.Transaction{}
	tx3.Execer = []byte(pt.Pos33TicketX)
	tx3.To = driver.ExecAddress(pt.Pos33TicketX)
	gticket := &pt.Pos33TicketAction_Genesis{}
	gticket.Genesis = &pt.Pos33TicketGenesis{MinerAddress: minerAddr, ReturnAddress: returnAddr, Count: count}
	tx3.Payload = types.Encode(&pt.Pos33TicketAction{Value: gticket, Ty: pt.Pos33TicketActionGenesis})
	ret = append(ret, &tx3)
	return ret
}

// CreateGenesisTx ticket create genesis tx
func (client *Client) CreateGenesisTx() (ret []*types.Transaction) {
	// 预先发行maxcoin 到 genesis 账户
	tx0 := types.Transaction{}
	tx0.Execer = []byte("coins")
	tx0.To = client.Cfg.Genesis
	g := &ct.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{Amount: types.MaxCoin}
	tx0.Payload = types.Encode(&ct.CoinsAction{Value: g, Ty: ct.CoinsActionGenesis})
	ret = append(ret, &tx0)

	// 初始化挖矿
	cfg := client.GetAPI().GetConfig()
	for _, genesis := range client.conf.Genesis {
		tx1 := createTicket(cfg, genesis.MinerAddr, genesis.ReturnAddr, genesis.Count, 0)
		ret = append(ret, tx1...)
	}
	return ret
}

// CreateGenesisTx1 used generate the first txs
/*
func (client *Client) CreateGenesisTx1() (ret []*types.Transaction) {
	// the 1st tx for issue 10,000,000,000 YCC
	act := &ct.CoinsAction_Genesis{Genesis: &types.AssetsGenesis{Amount: types.MaxCoin, ReturnAddress: client.conf.GenesisAddr}}
	tx := &types.Transaction{
		Execer:  []byte("coins"),
		To:      client.conf.GenesisAddr, //address.GetExecAddress("pos33").String(),
		Payload: types.Encode(&ct.CoinsAction{Value: act, Ty: ct.CoinsActionGenesis}),
	}
	ret = append(ret, tx)

	tx = &types.Transaction{}
	tx.Execer = []byte("coins")
	tx.To = address.GetExecAddress("pos33").String()
	act = &ct.CoinsAction_Genesis{Genesis: &types.AssetsGenesis{Amount: 1001 * client.conf.Price, ReturnAddress: client.conf.GenesisAddr}}
	tx.Payload = types.Encode(&ct.CoinsAction{Value: act, Ty: ct.CoinsActionGenesis})
	ret = append(ret, tx)

	// the 2th tx for the genesis accout frozon margin,
	// so the second block must created by the genesis accout.
	tx = &types.Transaction{}
	tx.Execer = []byte("pos33")
	tx.To = address.GetExecAddress("pos33").String()
	dact := &pt.TicketOpen{W: 1000}
	tx.Payload = types.Encode(&pt.TicketAction{Value: &pt.TicketAction_Topen{: dact}, Ty: int32(pt.Pos33ActionDeposit)})
	tx.Sign(types.ED25519, RootPrivKey)
	if !tx.CheckSign() {
		panic("tx check error")
	}
	ret = append(ret, tx)
	return
}
*/

// write block to chain
func (client *Client) setBlock(b *types.Block) error {
	if b == nil {
		plog.Crit("block is nil")
		return nil
	}

	plog.Info("setBlock", "height", b.Height, "txCount", len(b.Txs))
	lastBlock, err := client.RequestBlock(b.Height - 1)
	if err != nil {
		return err
	}
	err = client.WriteBlock(lastBlock.StateHash, b)
	if err != nil {
		return err
	}
	return nil
}

func getMiner(b *types.Block) (*pt.Pos33Miner, error) {
	if b == nil {
		return nil, fmt.Errorf("b is nil")
	}
	tx := b.Txs[0]
	var pact pt.Pos33Miner
	err := types.Decode(tx.Payload, &pact)
	if err != nil {
		return nil, err
	}
	return &pact, nil
}

// Get used search block store db
func (client *Client) Get(key []byte) ([]byte, error) {
	query := &types.LocalDBGet{Keys: [][]byte{key}}
	msg := client.GetQueueClient().NewMessage("blockchain", types.EventLocalGet, query)
	client.GetQueueClient().Send(msg, true)
	resp, err := client.GetQueueClient().Wait(msg)

	if err != nil {
		plog.Error(err.Error()) //no happen for ever
		return nil, err
	}
	value := resp.GetData().(*types.LocalReplyValue).Values[0]
	if value == nil {
		return nil, types.ErrNotFound
	}
	return value, nil
}

func (client *Client) sendTx(tx *types.Transaction) error {
	qcli := client.GetQueueClient()
	if qcli == nil {
		panic("client not bind message queue.")
	}
	msg := qcli.NewMessage("mempool", types.EventTx, tx)
	err := qcli.Send(msg, true)
	if err != nil {
		return err
	}
	resp, err := qcli.Wait(msg)
	if err != nil {
		return err
	}
	r := resp.GetData().(*types.Reply)
	if r.IsOk {
		return nil
	}
	plog.Info("sendTx error:", "error", string(r.Msg))
	return fmt.Errorf(string(r.Msg))
}
