package pos33

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	pb "github.com/33cn/chain33/types"

	"github.com/33cn/chain33/system/dapp/coins/types"
)

// *****************
// FOR TEST
// *****************

const nodeN = 10

var nodePrivs = make([]crypto.PrivKey, nodeN)

// Int is int64 for marshal
type Int int64

// Marshal to []byte
func (i Int) Marshal() []byte {
	b := make([]byte, 16)
	n := binary.PutVarint(b, int64(i))
	return b[:n]
}

// Unmarshal from []byte to int
func (i *Int) Unmarshal(b []byte) (int, error) {
	a, n := binary.Varint(b)
	*i = Int(a)
	return n, nil
}

// UnmarshalInt from []byte to int, and return length
func UnmarshalInt(b []byte) (Int, int, error) {
	var i Int
	n, err := i.Unmarshal(b)
	return i, n, err
}

type pp struct {
	i int
	p crypto.PrivKey
}

// Tx is alias Transaction
type Tx = pb.Transaction

func runTest(cli *Client, initTx bool, n, m int, hch <-chan int64) {
	privs := loadAccounts("acc.dat", n)
	if initTx {
		initRun(cli, privs, "init.dat")
		time.Sleep(time.Second * 10)
	}
	run(cli, privs, m, hch)
}

func run(cli *Client, privs []crypto.PrivKey, n int, hch <-chan int64) {
	ch := generateTxs(privs, n, hch)
	var tmCh = make(chan int, 1)
	sleepM := 100

	go func() {
		for range time.Tick(time.Second) {
			m := rand.Intn(n)
			if m < 10 {
				m = 10
			}
			tmCh <- m
		}
	}()

	for {
		tx := <-ch
		select {
		case sleepM = <-tmCh:
		default:
		}
		m := sleepM
		sendTx(cli, tx, m)
		// if tx.Expire-pb.TxHeightFlag > 4000 {
		// 	break
		// }
	}
}

func generateTxs(privs []crypto.PrivKey, n int, hch <-chan int64) chan *Tx {
	log.Println("generateTxs:", n)
	l := len(privs) - 1
	ch := make(chan *Tx, 16)
	f := func(chh <-chan int64) {
		height := int64(0)
		for {
			i := rand.Intn(len(privs))
			signer := privs[l-i]
			select {
			case height = <-chh:
			default:
			}
			ch <- newTx(signer, 1, address.PubKeyToAddress(privs[i].PubKey().Bytes()).String(), height)
		}
	}
	N := 16
	hchs := make([]chan int64, N)
	for i := 0; i < N; i++ {
		hchs[i] = make(chan int64, 1)
	}
	go func() {
		for height := range hch {
			for i := 0; i < N; i++ {
				select {
				case hchs[i] <- height:
				default:
				}
			}
		}
	}()

	for i := 0; i < N; i++ {
		go f(hchs[i])
	}
	return ch
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
		for i := 0; i < m; i++ {
			l, err := binary.ReadVarint(r)
			if err != nil {
				log.Println(err)
				break
			}
			data := make([]byte, l)
			_, err = r.Read(data)
			if err != nil {
				log.Println(err)
				break
			}
			var tx Tx
			err = pb.Decode(data, &tx)
			if err != nil {
				log.Println(err)
				break
			}
			ch <- &tx
			if i%10000 == 0 {
				log.Println("readTxs:", i)
			}
		}
		close(ch)
	}(n)
	return nil
}

func sendTx(cli *Client, tx *Tx, n int) {
	// log.Println("send tx:", tx.To, string(tx.Execer))
	err := cli.sendTx(tx)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Microsecond * time.Duration(1000000/n))
}

func initRun(cli *Client, privs []crypto.PrivKey, filename string) {
	ch := make(chan *Tx, 16)
	// err := readTxs(ch, filename, len(privs))
	// if err != nil {
	// 	os.Remove(filename)
	// 	writeInitTxs(privs, filename)
	// 	return
	// }
	n := 0
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
		sendTx(cli, tx, 1000)
	}
	log.Println("init txs finished:", n)
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
		writeTx(tx, f)
		n++
		if n == len(privs) {
			close(done)
			break
		}
	}
}

func newTx(priv crypto.PrivKey, amount int64, to string, height int64) *Tx {
	act := &types.CoinsAction{Value: &types.CoinsAction_Transfer{Transfer: &pb.AssetsTransfer{Cointoken: "YCC", Amount: amount}}, Ty: types.CoinsActionTransfer}
	tx := &Tx{
		Nonce:   rand.Int63(),
		Execer:  pb.ExecerCoins,
		Fee:     pb.MinFee + rand.Int63n(pb.MinFee),
		To:      to, //
		Payload: pb.Encode(act),
		Expire:  height + pb.TxHeightFlag,
	}
	tx.Sign(2, RootPrivKey)
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
		ch <- newTx(RootPrivKey, m, address.PubKeyToAddress(priv.PubKey().Bytes()).String(), 0)
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
		if i < nodeN {
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
