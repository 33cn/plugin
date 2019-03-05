package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	// "math"
	"math/rand"
	"os"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	rpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"

	"github.com/33cn/chain33/system/dapp/coins/types"
	pb "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/consensus/pos33"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

//
// test auto generate tx and send to the node
//

func init() {
	rand.Seed(time.Now().UnixNano())
}

var rpcURL = flag.String("u", "http://localhost:8801", "rpc url")
var pnodes = flag.Bool("n", false, "only print node private keys")
var ini = flag.Bool("i", false, "send init tx")
var dpst = flag.String("d", "", "send deposit tx")
var maxacc = flag.Int("a", 1000, "max account")
var maxtx = flag.Int("t", 1000, "max txs")
var dw = flag.Int("w", 7, "deposit weight")

var gClient *rpc.JSONClient

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()
	privs := loadAccounts("./acc.dat", *maxacc)
	if *pnodes {
		return
	}

	client, err := rpc.NewJSONClient(*rpcURL)
	if err != nil {
		log.Fatal(err)
	}
	gClient = client

	if *dpst != "" {
		cpt, err := crypto.New("ed25519")
		if err != nil {
			panic(err)
		}
		kb, err := hex.DecodeString(*dpst)
		if err != nil {
			log.Fatal(err)
		}
		priv, err := cpt.PrivKeyFromBytes(kb)
		if err != nil {
			log.Fatal(err)
		}

		sendDepositTx(priv)
		return
	}
	if *ini {
		sendInitTxs(privs, "./init.dat")
		log.Println("@@@@@@@ send init txs", *ini)
	} else {
		run(privs)
	}
}

// NODEN is number of nodes
const NODEN = 10

var nodePrivs = make([]crypto.PrivKey, NODEN)

// Int is int64
type Int int64

// Marshal Int to []byte
func (i Int) Marshal() []byte {
	b := make([]byte, 16)
	n := binary.PutVarint(b, int64(i))
	return b[:n]
}

// Unmarshal []byte to Int
func (i *Int) Unmarshal(b []byte) (int, error) {
	a, n := binary.Varint(b)
	*i = Int(a)
	return n, nil
}

// UnmarshalInt is helper func
func UnmarshalInt(b []byte) (Int, int, error) {
	var i Int
	n, err := i.Unmarshal(b)
	return i, n, err
}

type pp struct {
	i int
	p crypto.PrivKey
}

// Tx is alise pb.Transaction
type Tx = pb.Transaction

func run(privs []crypto.PrivKey) {
	ch := generateTxs(privs)
	for {
		tx := <-ch
		sendTx(tx)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(50)))
	}
}

func generateTxs(privs []crypto.PrivKey) chan *Tx {
	N := 4
	l := len(privs) - 1
	ch := make(chan *Tx, N)
	f := func() {
		for {
			i := rand.Intn(len(privs))
			signer := privs[l-i]
			ch <- newTx(signer, 1, address.PubKeyToAddress(privs[i].PubKey().Bytes()).String())
		}
	}
	for i := 0; i < N; i++ {
		go f()
	}
	return ch
}

func sendTx(tx *Tx) {
	c := gClient
	err := c.Call("Chain33.SendTransaction", rpctypes.RawParm{Data: hex.EncodeToString(pb.Encode(tx))}, nil)
	if err != nil {
		_, ok := err.(*json.InvalidUnmarshalError)
		if !ok {
			log.Println("@@@ rpc error: ", err)
		}
	}
}

func sendDepositTx(p crypto.PrivKey) {
	act := &ty.Pos33Action{Value: &ty.Pos33Action_Deposit{Deposit: &ty.Pos33DepositAction{W: int64(*dw)}}, Ty: int32(ty.Pos33ActionDeposit)}
	tx := &Tx{
		Nonce:   rand.Int63(),
		Execer:  []byte("pos33"),
		Payload: pb.Encode(act),
		Fee:     1e7 * 17,
		To:      address.ExecAddress("pos33"),
		Expire:  time.Now().Unix() + 600,
	}
	tx.Sign(pb.ED25519, p)
	sendTx(tx)
}

func sendInitTxs(privs []crypto.PrivKey, filename string) {
	n := 0
	ch := make(chan *Tx, 16)
	done := make(chan struct{}, 1)
	go generateInitTxs(0, privs, ch, done)
	for {
		n++
		if n == len(privs) {
			close(done)
			break
		}
		tx, ok := <-ch
		if !ok {
			break
		}
		sendTx(tx)
	}
	log.Println("init txs finished:", n)
}

const pos33MinFee = 1e7

func newTx(priv crypto.PrivKey, amount int64, to string) *Tx {
	act := &types.CoinsAction{Value: &types.CoinsAction_Transfer{Transfer: &pb.AssetsTransfer{Cointoken: "YCC", Amount: amount}}, Ty: types.CoinsActionTransfer}
	tx := &Tx{
		Nonce:   rand.Int63(),
		Execer:  []byte(types.CoinsX),
		Fee:     pos33MinFee + rand.Int63n(pos33MinFee),
		To:      to, //
		Payload: pb.Encode(act),
		Expire:  time.Now().Unix() + 3600,
	}
	tx.Sign(2, pos33.RootPrivKey)
	return tx
}

func generateInitTxs(n int, privs []crypto.PrivKey, ch chan *Tx, done chan struct{}) {
	for _, priv := range privs {
		select {
		case <-done:
			return
		default:
		}

		m := 2000 * pb.Coin
		ch <- newTx(pos33.RootPrivKey, m, address.PubKeyToAddress(priv.PubKey().Bytes()).String())
	}
	log.Println(n, len(privs))
}

//
func generateAccounts(max int) []crypto.PrivKey {
	t := time.Now()
	var pks []crypto.PrivKey
	goN := 16
	ch := make(chan pp, goN)
	done := make(chan struct{}, 1)

	c, _ := crypto.New("ed25519")
	for i := 0; i < goN; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					priv, _ := c.GenKey()
					ch <- pp{i: 0, p: priv}
				}
			}
		}()
	}
	for {
		p := <-ch
		pks = append(pks, p.p)
		l := len(pks)
		if l%1000 == 0 {
			log.Println("generate acc:", len(pks))
		}
		if len(pks) == max {
			close(done)
			break
		}
	}
	log.Println(time.Now().Sub(t))
	return pks
}

func loadAccounts(filename string, max int) []crypto.PrivKey {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		privs := generateAccounts(max)

		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		var data []byte
		data = append(data, Int(len(privs)).Marshal()...)
		f.Write(data)
		for _, p := range privs {
			f.Write(p.Bytes())
		}
		log.Fatal("generate ", len(privs), " accounts, please restart")
		return privs
	}
	var l Int
	n, err := l.Unmarshal(b)
	if err != nil {
		log.Fatal(err)
	}
	ln := int(l)
	b = b[n:]

	if max < ln {
		ln = max
	}

	privs := make([]crypto.PrivKey, ln)

	n = 64
	c, _ := crypto.New("ed25519")
	for i := 0; i < ln; i++ {
		p := b[:n]
		priv, err := c.PrivKeyFromBytes(p)
		if err != nil {
			log.Fatal(err)
		}
		if i < NODEN {
			nodePrivs[i] = priv
			log.Println(i, hex.EncodeToString(priv.Bytes()))
		}
		privs[i] = priv
		b = b[n:]
		if i%10000 == 0 {
			log.Println("load acc:", i)
		}
	}
	log.Println("loadAccount: ", len(privs))
	return privs
}
