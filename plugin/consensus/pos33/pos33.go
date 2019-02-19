package pos33

import (
	"bytes"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/ed25519"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	ced25519 "github.com/33cn/chain33/system/crypto/ed25519"
	"github.com/33cn/chain33/system/dapp/coins/types"
	pb "github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
	// "github.com/33cn/chain33/util"
)

const rootSeed = "YCC-ROOT"
const pos33MinFee = 1e7

// RootPrivKey is the root private key for ycc
var RootPrivKey crypto.PrivKey

// RootAddr is the root account address for ycc
var RootAddr string

var myCrypto crypto.Crypto

func genKeyFromSeed(seed []byte) (crypto.PrivKey, error) {
	_, priv, err := ed25519.GenerateKey(bytes.NewReader(crypto.Sha256(seed)))
	if err != nil {
		return nil, err
	}
	return ced25519.PrivKeyEd25519(*priv), nil
}

func init() {
	drivers.Reg("pos33", New)
	cpt, err := crypto.New("ed25519")
	if err != nil {
		panic(err)
	}
	myCrypto = cpt
	priv, err := genKeyFromSeed([]byte(rootSeed))
	if err != nil {
		panic(err)
	}
	RootPrivKey = priv
	RootAddr = address.PubKeyToAddress(RootPrivKey.PubKey().Bytes()).String()
	rand.Seed(time.Now().Unix())
}

// Client is the pos33 consensus client
type Client struct {
	*drivers.BaseClient
	n    *node
	priv crypto.PrivKey
}

type subConfig struct {
	Pos33SecretSeed    string `json:"Pos33SecretSeed,omitempty"`
	Pos33ListenAddr    string `json:"Pos33ListenAddr,omitempty"`
	Pos33AdvertiseAddr string `json:"Pos33AdvertiseAddr,omitempty"`
	Pos33PeerSeed      string `json:"Pos33PeerSeed,omitempty"`
	Pos33Test          bool   `json:"Pos33Test,omitempty"`
	Pos33TestInit      bool   `json:"Pos33TestInit,omitempty"`
	Pos33TestMaxAccs   int64  `json:"Pos33TestMaxAccs,omitempty"`
	Pos33TestMaxTxs    int64  `json:"Pos33TestMaxTxs,omitempty"`
	Pos33MaxTxs        int64  `json:"Pos33MaxTxs,omitempty"`
	Pos33BlockTime     int64  `json:"Pos33BlockTime,omitempty"`
	Pos33BlockTimeout  int64  `json:"Pos33BlockTimeout,omitempty"`
	Pos33MinFee        int64  `json:"Pos33MinFee,omitempty"`
}

// New create pos33 consensus client
func New(cfg *pb.Consensus, sub []byte) queue.Module {
	c := drivers.NewBaseClient(cfg)
	var subcfg subConfig
	if sub != nil {
		pb.MustDecode(sub, &subcfg)
	}

	n := newNode(&subcfg)
	client := &Client{BaseClient: c, n: n}
	client.n.Client = client
	client.Cfg.Genesis = RootAddr
	c.SetChild(client)
	return client
}

// Close is close the client
func (client *Client) Close() {}

// ProcEvent do nothing?
func (client *Client) ProcEvent(msg queue.Message) bool {
	return false
}

func (client *Client) newBlock(txs []*pb.Transaction, height int64, null bool, txcont int) (*pb.Block, error) {
	lastBlock, err := client.RequestLastBlock()
	if err != nil {
		plog.Error("newBlock error", "error", err)
		return nil, err
	}
	if lastBlock.Height+1 != height {
		plog.Error("newBlock height error", "lastHeight", lastBlock.Height, "height", height)
		return nil, errors.New("the last block too low")
	}
	if !null {
		ch := make(chan []*Tx, 1)
		go func() { ch <- client.RequestTx(txcont, nil) }()
		select {
		case <-time.After(time.Millisecond * 300):
		case ts := <-ch:
			txs = append(txs, ts...)
		}
	}

	bt := time.Now().UnixNano() / 1000000
	return &pb.Block{
		ParentHash: lastBlock.Hash(),
		Height:     lastBlock.Height + 1,
		Txs:        txs,
		TxHash:     merkle.CalcMerkleRoot(txs),
		BlockTime:  bt,
		Difficulty: lastBlock.Difficulty + 1,
	}, nil
}

// CheckBlock check block callback
func (client *Client) CheckBlock(parent *pb.Block, current *pb.BlockDetail) error {
	b := current.GetBlock()
	if client.checkBlock(b) {
		return nil
	}
	return errors.New("check block error")
}

// AddBlock notice driver a new block incoming
func (client *Client) AddBlock(b *pb.Block) {
	client.n.addBlock(b)
}

// CreateBlock will start run
func (client *Client) CreateBlock() {
	for {
		if !client.IsMining() || !(client.IsCaughtUp() || client.Cfg.ForceMining) {
			plog.Info("createblock.ismining is disable or client is caughtup is false")
			time.Sleep(time.Second)
			continue
		}
		if client.getWeight(client.n.pub) == 0 {
			plog.Info("if do consensus, must deposit 1,000,000 YCC")
			time.Sleep(time.Second)
			continue
		}
		break
	}
	client.n.run()
}

// CreateGenesisTx used generate the first txs
func (client *Client) CreateGenesisTx() (ret []*pb.Transaction) {
	// the 1st tx for issue YCC
	act0 := &types.CoinsAction_Genesis{Genesis: &pb.AssetsGenesis{Amount: 1e8 * pb.Coin * 3}}
	tx := &pb.Transaction{
		Execer:  []byte("coins"),
		To:      RootAddr,
		Payload: pb.Encode(&types.CoinsAction{Value: act0, Ty: types.CoinsActionGenesis}),
	}
	ret = append(ret, tx)

	// plog.Error("@@@@@@@ genersis tx0 to: ", tx.To)

	// the 2th tx for the genesis accout frozon margin,
	// so the second block must created by the genesis accout.
	tx = &pb.Transaction{}
	tx.Execer = []byte("pos33")
	tx.To = address.GetExecAddress("pos33").String()
	dact := &ty.Pos33DepositAction{W: 100}
	tx.Payload = pb.Encode(&ty.Pos33Action{Value: &ty.Pos33Action_Deposit{Deposit: dact}, Ty: int32(ty.Pos33ActionDeposit)})
	tx.Sign(pb.ED25519, RootPrivKey)
	ret = append(ret, tx)
	// plog.Error("@@@@@@@ genersis tx1 to: ", tx.To)
	return
}

// write block to chain
func (client *Client) setBlock(b *pb.Block) error {
	if b == nil {
		plog.Crit("block is nil")
		return nil
	}
	plog.Info("setBlock", "height", b.Height)
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

func (client *Client) allWeight() int {
	k := []byte(ty.Pos33AllWeight)
	v, err := client.Get(k)
	if err != nil {
		plog.Error(err.Error())
		return 0
	}
	w, err := strconv.Atoi(string(v))
	if err != nil {
		plog.Error(err.Error())
		return 0
	}
	return w
}

func (client *Client) getWeight(pub string) int {
	addr := address.PubKeyToAddress([]byte(pub)).String()
	v, err := client.Get([]byte(ty.Pos33WeightPrefix + addr))
	if err != nil {
		plog.Error(err.Error())
		return 0
	}
	w, err := strconv.Atoi(string(v))
	if err != nil {
		plog.Error(err.Error())
		return 0
	}
	return w
}

// TODO:?
func (client *Client) checkBlock(b *pb.Block) bool {
	plog.Info("checkBlock", "b.Height=", b.Height)

	lastB, err := client.RequestBlock(b.Height - 1)
	if err != nil {
		plog.Error(err.Error())
		return false
	}

	if string(b.ParentHash) != string(lastB.Hash()) {
		plog.Error("parent hash error")
		return false
	}

	if len(b.Txs) < 1 {
		plog.Error("len(txs) == 0")
		return false
	}

	// if !client.n.checkBlock(b) {
	// 	return false
	// }
	/*
		act, err := getBlockReword(b)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		llb, err := client.RequestBlock(b.Height - 2)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		seed, err := getBlockSeed(llb)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
			err = client.checkReword(seed, act)
			if err != nil {
				plog.Error(err.Error())
				return false
			}
	*/

	// TODO: check punish

	return true
}

/*
func (client *Client) checkPunish(lastBlock *pb.Block, act *ty.Pos33PunishAction) error {
	for who, p := range act.Punishs {
		if p.Vote1 == nil || p.Vote2 == nil {
			return errors.New("punish votes error")
		}

		if p.Vote1.BlockHeight != lastBlock.Height || p.Vote2.BlockHeight != lastBlock.Height {
			return errors.New("punish votes height error")
		}
		if string(p.Vote1.BlockHash) == string(p.Vote2.BlockHash) { // only punish NOT same votes
			return errors.New("punish votes block hash error")
		}
		if !p.Vote1.Verify() || !p.Vote2.Verify() {
			return errors.New("punish votes verify error")
		}
		if who != address.PubKeyToAddress(p.Vote1.BlockSig.Pubkey).String() ||
			who != address.PubKeyToAddress(p.Vote2.BlockSig.Pubkey).String() {
			return errors.New("punish votes address error")
		}
	}

	return nil
}
*/

func getBlockReword(b *pb.Block) (*ty.Pos33RewordAction, error) {
	tx := b.Txs[0]
	var pact ty.Pos33Action
	err := pb.Decode(tx.Payload, &pact)
	if err != nil {
		return nil, err
	}
	act := pact.GetReword()
	return act, nil
}

/*
func (client *Client) checkReword(seed []byte, act *ty.Pos33RewordAction) error {
	votes := act.Votes
	if len(votes) == 0 {
		return nil
	}

	allw := client.allWeight()
	sumw := 0
	for _, v := range votes {
		if !v.Verify() {
			return errors.New("vote verify error")
		}
		pub := string(v.BlockSig.Pubkey)
		w := client.getWeight(pub)
		_, err := checkRands(pub, allw, w, seed, v.Rands, v.BlockHeight)
		if err != nil {
			return err
		}
		sumw += len(v.Rands.Rands)
	}
	if allw >= ty.Pos33Members {
		if sumw*3 < ty.Pos33Members*2 {
			return errors.New("vote weight error 0")
		}
		if sumw > ty.Pos33Members {
			return errors.New("vote weight error 1")
		}
	} else {
		if sumw > allw {
			return errors.New("vote weight error 2")
		}
		if sumw*3 <= allw*2 {
			return errors.New("vote weight error 3")
		}
	}
	return nil
}
*/

// Get used search block store db
func (client *Client) Get(key []byte) ([]byte, error) {
	query := &pb.LocalDBGet{Keys: [][]byte{key}}
	msg := client.GetQueueClient().NewMessage("blockchain", pb.EventLocalGet, query)
	client.GetQueueClient().Send(msg, true)
	resp, err := client.GetQueueClient().Wait(msg)

	if err != nil {
		plog.Error(err.Error()) //no happen for ever
	}
	value := resp.GetData().(*pb.LocalReplyValue).Values[0]
	if value == nil {
		return nil, pb.ErrNotFound
	}
	return value, nil
}

func (client *Client) sendTx(tx *pb.Transaction) error {
	qcli := client.GetQueueClient()
	if qcli == nil {
		panic("client not bind message queue.")
	}
	msg := qcli.NewMessage("mempool", pb.EventTx, tx)
	err := qcli.Send(msg, false)
	if err != nil {
		return err
	}
	//plog.Info("sendTx", "N", N)

	return nil
	// resp, err := qcli.Wait(msg)
	// if err != nil {
	// 	return err
	// }
	// r := resp.GetData().(*pb.Reply)
	// if r.IsOk {
	// 	return nil
	// } else {
	// 	return errors.New(string(r.Msg))
	// }
}
