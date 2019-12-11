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

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	rpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	ctypes "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
)

//
// test auto generate tx and send to the node
//

var rootKey crypto.PrivKey

func init() {
	rand.Seed(time.Now().UnixNano())
	rootKey = HexToPrivkey("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
}

var rpcURL = flag.String("u", "http://localhost:8801", "rpc url")
var pnodes = flag.Bool("n", false, "only print node private keys")
var ini = flag.Bool("i", false, "send init tx")
var dpst = flag.String("d", "", "send deposit tx")
var maxacc = flag.Int("a", 1000, "max account")
var maxtx = flag.Int("t", 1000, "max txs")
var dw = flag.Int("w", 7, "deposit weight")
var rn = flag.Int("r", 3000, "sleep in Microsecond")
var conf = flag.String("c", "chain33.toml", "chain33 config file")

var gClient *rpc.JSONClient
var config *types.Chain33Config

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()

	config = types.NewChain33Config(types.ReadFile(*conf))

	privs := loadAccounts("./acc.dat", *maxacc)
	if *pnodes {
		return
	}

	client, err := rpc.NewJSONClient(*rpcURL)
	if err != nil {
		log.Fatal(err)
	}
	gClient = client

	if *ini {
		sendInitTxs(privs, "./init.dat")
		log.Println("@@@@@@@ send init txs", *ini)
	} else {
		run(privs)
	}
}

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

// Tx is alise types.Transaction
type Tx = types.Transaction

func run(privs []crypto.PrivKey) {
	ch := generateTxs(privs)
	i := 0
	for {
		tx := <-ch
		sendTx(tx)
		time.Sleep(time.Microsecond * time.Duration(*rn))
		i++
		log.Println(i, "... txs sent")
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
	err := c.Call("Chain33.SendTransaction", rpctypes.RawParm{Data: hex.EncodeToString(types.Encode(tx))}, nil)
	if err != nil {
		_, ok := err.(*json.InvalidUnmarshalError)
		if !ok {
			log.Println("@@@ rpc error: ", err, tx.From())
		}
	}
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

func newTx(priv crypto.PrivKey, amount int64, to string) *Tx {
	act := &ctypes.CoinsAction{Value: &ctypes.CoinsAction_Transfer{Transfer: &types.AssetsTransfer{Cointoken: "YCC", Amount: amount}}, Ty: ctypes.CoinsActionTransfer}
	payload := types.Encode(act)
	tx, err := types.CreateFormatTx(config, "coins", payload)
	if err != nil {
		panic(err)
	}
	tx.Fee *= 10
	tx.To = to
	tx.Sign(types.SECP256K1, priv)
	return tx
}

//HexToPrivkey ： convert hex string to private key
func HexToPrivkey(key string) crypto.PrivKey {
	cr, err := crypto.New(secp256k1.Name)
	if err != nil {
		panic(err)
	}
	bkey, err := common.FromHex(key)
	if err != nil {
		panic(err)
	}
	priv, err := cr.PrivKeyFromBytes(bkey)
	if err != nil {
		panic(err)
	}
	return priv
}

func generateInitTxs(n int, privs []crypto.PrivKey, ch chan *Tx, done chan struct{}) {
	for _, priv := range privs {
		select {
		case <-done:
			return
		default:
		}

		m := 1000 * types.Coin
		ch <- newTx(rootKey, m, address.PubKeyToAddress(priv.PubKey().Bytes()).String())
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

	c, _ := crypto.New(secp256k1.Name)
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

	n = 32
	c, _ := crypto.New(secp256k1.Name)
	for i := 0; i < ln; i++ {
		p := b[:n]
		priv, err := c.PrivKeyFromBytes(p)
		if err != nil {
			log.Fatal(err)
		}
		privs[i] = priv
		b = b[n:]
		if i%10000 == 0 {
			log.Println("load acc:", i)
		}
		log.Println("account: ", address.PubKeyToAddr(priv.PubKey().Bytes()))
	}
	log.Println("loadAccount: ", len(privs))
	return privs
}
