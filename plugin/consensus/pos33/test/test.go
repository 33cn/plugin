package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
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
	"github.com/33cn/plugin/plugin/consensus/pos33"

	"github.com/33cn/chain33/system/dapp/coins/types"
	pb "github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

//
// test auto generate tx and send to the node
//

func init() {
	rand.Seed(time.Now().UnixNano())
}

var rpcUrl = flag.String("u", "http://localhost:8801", "rpc url")
var pnodes = flag.Bool("n", false, "only print node private keys")
var ini = flag.Bool("i", false, "send init tx")
var dpst = flag.String("d", "", "send deposit tx")
var maxacc = flag.Int("a", 1000, "max account")
var maxtx = flag.Int("t", 1000, "max txs")
var dw = flag.Int("w", 7, "deposit weight")

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()
	privs := loadAccounts("./acc.dat", *maxacc)
	if *pnodes {
		return
	}
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
		run(privs, "./txs.dat", *maxtx)
	}
}

const NODE_N = 20

var nodePrivs []crypto.PrivKey = make([]crypto.PrivKey, NODE_N)

type Int int64

func (i Int) Marshal() []byte {
	b := make([]byte, 16)
	n := binary.PutVarint(b, int64(i))
	return b[:n]
}

func (i *Int) Unmarshal(b []byte) (int, error) {
	a, n := binary.Varint(b)
	*i = Int(a)
	return n, nil
}

func UnmarshalInt(b []byte) (Int, int, error) {
	var i Int
	n, err := i.Unmarshal(b)
	return i, n, err
}

type pp struct {
	i int
	p crypto.PrivKey
}

type Tx = pb.Transaction

func run(privs []crypto.PrivKey, filename string, n int) {
	ch := make(chan *Tx, 16)
	err := readTxs(ch, filename, n)
	if err != nil {
		os.Remove(filename)
		generateTxs(privs, filename, n)
	} else {
		for {
			tx := <-ch
			sendTx(tx)
		}
	}
}

func generateTxs(privs []crypto.PrivKey, filename string, n int) {
	log.Println("generateTxs:", filename, n)
	l := len(privs) - 1
	ch := make(chan *Tx, 16)
	f := func() {
		for {
			i := rand.Intn(len(privs))
			signer := privs[l-i]
			act := &types.CoinsAction{Value: &types.CoinsAction_Transfer{Transfer: &pb.AssetsTransfer{Cointoken: "YCC", Amount: 2}}, Ty: types.CoinsActionTransfer}
			tx := &Tx{
				Nonce:   rand.Int63(),
				Execer:  types.ExecerCoins,
				Fee:     1e7 * 17,
				To:      address.PubKeyToAddress(privs[i].PubKey().Bytes()).String(),
				Payload: pb.Encode(act),
			}
			tx.Sign(pb.ED25519, signer)
			ch <- tx
		}
	}

	for i := 0; i < 16; i++ {
		go f()
	}

	fr, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer fr.Close()

	ch2 := make(chan *Tx, 16)
	g := func() {
		for {
			tx := <-ch2
			//writeTx(tx, fr)
			sendTx(tx)
		}
	}

	for i := 0; i < 16; i++ {
		go g()
	}

	t1 := time.Now()
	// n0 := 0
	// n := n0
	m := 0
	for {
		ch2 <- <-ch
		m++
		if m%10000 == 0 {
			log.Println("generate tx:", m)
		}
		if m >= n {
			break
		}
	}
	time.Sleep(time.Second)
	log.Println("generate tx:", n, time.Now().Sub(t1))
}

func writeTx(tx *Tx, w io.Writer) {
	data := pb.Encode(tx)
	l := len(data)
	_, err := w.Write(Int(l).Marshal())
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

type freader struct {
	*os.File
}

func (r *freader) ReadByte() (byte, error) {
	b := make([]byte, 1)
	_, err := r.File.Read(b)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	return b[0], nil
}

func readTxs(ch chan *Tx, filename string, n int) error {
	log.Println("readTxs:", filename, n)
	f, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Println(err)
		return err
	}
	r := &freader{f}
	go func(m int) {
		defer f.Close()
		for {
			l, err := binary.ReadVarint(r)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
				break
			}
			data := make([]byte, l)
			_, err = r.Read(data)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
				break
			}
			var tx Tx
			err = pb.Decode(data, &tx)
			if err != nil {
				log.Fatal(err)
			}
			ch <- &tx
			m--
			if m <= 0 {
				break
			}
		}
		close(ch)
	}(n)
	return nil
}

func sendTx(tx *Tx) {
	c, err := rpc.NewJSONClient(*rpcUrl)
	if err != nil {
		log.Fatal(err)
	}
	err = c.Call("Chain33.SendTransaction", rpctypes.RawParm{Data: hex.EncodeToString(pb.Encode(tx))}, nil)
	if err != nil {
		_, ok := err.(*json.InvalidUnmarshalError)
		if !ok {
			log.Println("@@@ rpc error: ", err)
		}
	}
}

func sendDepositTx(p crypto.PrivKey) {
	// for _, p := range nodePrivs {
	act := &ty.Pos33Action{Value: &ty.Pos33Action_Deposit{Deposit: &ty.Pos33DepositAction{W: int64(*dw)}}, Ty: int32(ty.Pos33ActionDeposit)}
	tx := &Tx{Nonce: rand.Int63(), Execer: []byte("pos33"), Payload: pb.Encode(act), Fee: 1e7 * 17, To: address.ExecAddress("pos33")}
	tx.Sign(pb.ED25519, p)
	sendTx(tx)
	// }
}

func sendInitTxs(privs []crypto.PrivKey, filename string) {
	writeInitTxs(privs, filename)
	/*
			err := readTxs(ch, filename, len(privs))
			if err != nil {
				os.Remove(filename)
				return
			}
		for {
			tx, ok := <-ch
			if !ok {
				break
			}
			sendTx(tx)
		}
	*/
}

func writeInitTxs(privs []crypto.PrivKey, filename string) {
	log.Println("writeInitTxs:", filename, len(privs))
	ch := make(chan *Tx, 16)
	done := make(chan struct{}, 1)
	l := len(privs) / 10
	for i := 0; i < 10; i++ {
		// log.Println(l*i, l*i+l-1)
		go generateInitTxs(i, privs[l*i:l*i+l], ch, done)
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	n := 0
	for {
		tx := <-ch
		//writeTx(tx, f)
		sendTx(tx)
		n++
		if n == len(privs) {
			close(done)
			break
		}
	}
}

func generateInitTxs(n int, privs []crypto.PrivKey, ch chan *Tx, done chan struct{}) {
	for _, priv := range privs {
		select {
		case <-done:
			return
		default:
		}

		n := 1000000 * pb.Coin
		// if i < NODE_N {
		// 	n = 2 * types.Pos33Miner
		// 	// log.Printf("nodePrivs[%d] = %s\n", i, hex.EncodeToString(priv.Bytes()))
		// }
		act := &types.CoinsAction{Value: &types.CoinsAction_Transfer{Transfer: &pb.AssetsTransfer{Cointoken: "YCC", Amount: n}}, Ty: types.CoinsActionTransfer}
		tx := &Tx{
			Nonce:   rand.Int63(),
			Execer:  types.ExecerCoins,
			Fee:     1e7 * 17,
			To:      address.PubKeyToAddress(priv.PubKey().Bytes()).String(),
			Payload: pb.Encode(act),
		}
		tx.Sign(2, pos33.RootPrivKey)
		ch <- tx
		// time.Sleep(time.Millisecond * 50)
	}
	log.Println(n, len(privs))
	// close(ch)
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
		if i < NODE_N {
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
