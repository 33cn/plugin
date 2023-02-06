// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/33cn/plugin/plugin/crypto/bls"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log/log15"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/system/crypto/none"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/valnode/types"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

const fee = 1e6
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789-=_+=/<>!@#$%^&"

var r *rand.Rand
var signType string

var (
	log      = log15.New()
	testExec = "none"
	execAddr = address.ExecAddress(testExec)
)

// TxHeightOffset needed
var TxHeightOffset int64

func main() {
	if len(os.Args) == 1 || os.Args[1] == "-h" {
		LoadHelp()
		return
	}
	fmt.Println("grpc url:", os.Args[2])

	// 指定设置交易执行器为平行链
	if strings.Contains(os.Args[0], "para") {
		testExec = "user.p.para.none"
		execAddr = address.ExecAddress(testExec)
	}
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	args := os.Args[1:]
	// 最后一个参数指定签名类型, 支持 bls
	signType = args[len(args)-1]

	switch args[0] {
	case "-h": //使用帮助
		LoadHelp()
	case "perf":
		if len(args) < 6 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		Perf(args[1], args[2], args[3], args[4], args[5])
	case "perfV2":
		if len(args) != 5 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		PerfV2(args[1], args[2], args[3], args[4])
	case "put":
		if len(args) != 3 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		Put(args[1], args[2], "")
	case "get":
		if len(args) != 3 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		Get(args[1], args[2])
	case "valnode":
		if len(args) != 4 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		ValNode(args[1], args[2], args[3])
	case "perfOld":
		if len(args) != 6 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		PerfOld(args[1], args[2], args[3], args[4], args[5])
	}
}

// LoadHelp ...
func LoadHelp() {
	fmt.Println("Available Commands:")
	fmt.Println("perf    [host, size, num, interval, duration]                : 写数据性能测试，interval单位为100毫秒，host形式为ip:port")
	fmt.Println("perfV2  [host, size, interval, duration]                     : 写数据性能测试，interval单位为秒,host形式为ip:port")
	fmt.Println("put     [ip, size]                                           : 写数据")
	fmt.Println("get     [ip, hash]                                           : 读数据")
	fmt.Println("valnode [ip, pubkey, power]                                  : 增加/删除/修改tendermint节点")
	fmt.Println("perfOld [ip, size, num, interval, duration]                  : 不推荐使用，写数据性能测试，interval单位为100毫秒")
}

// Perf 性能测试
// host grpc地址, localhost:8802
// txsize 存证交易字节大小
// num 单次循环总发送交易数量
// sleepinterval 单次循环后协程等待间隔秒
// totalduration  总持续次数
func Perf(host, txsize, num, sleepinterval, totalduration string) {
	var numThread int
	numInt, err := strconv.Atoi(num)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	sleep, err := strconv.Atoi(sleepinterval)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	durInt, err := strconv.Atoi(totalduration)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	sizeInt, _ := strconv.Atoi(txsize)
	if numInt < 10 {
		numThread = 1
	} else if numInt > 100 {
		numThread = 10
	} else {
		numThread = numInt / 10
	}
	numThread = runtime.NumCPU()
	ch := make(chan struct{}, numThread)
	chSend := make(chan struct{}, numThread*2)
	txChan := make(chan *types.Transaction, numInt)
	//payload := RandStringBytes(sizeInt)
	var blockHeight int64
	total := int64(0)
	success := int64(0)

	go func() {
		ch <- struct{}{}
		conn := newGrpcConn(host)
		defer conn.Close()
		gcli := types.NewChain33Client(conn)
		for {
			height, err := getHeight(gcli)
			if err != nil {
				//conn.Close()
				log.Error("getHeight", "err", err)
				//conn = newGrpcConn(ip)
				//gcli = types.NewChain33Client(conn)
				time.Sleep(time.Second)
			} else {
				atomic.StoreInt64(&blockHeight, height)
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
	<-ch

	for i := 0; i < numThread; i++ {
		go func() {
			_, priv := genaddress()
			for sec := 0; durInt == 0 || sec < durInt; sec++ {
				height := atomic.LoadInt64(&blockHeight)
				for txs := 0; txs < numInt/numThread; txs++ {
					//构造存证交易
					tx := txPool.Get().(*types.Transaction)
					tx.To = execAddr
					tx.Fee = 1e6
					tx.Nonce = time.Now().UnixNano()
					tx.Expire = height + types.TxHeightFlag + types.LowAllowPackHeight
					tx.Payload = RandStringBytes(sizeInt)
					//交易签名
					tx.Sign(int32(getSignID()), priv)
					txChan <- tx
				}
				if sleep > 0 {
					time.Sleep(100 * time.Millisecond * time.Duration(sleep))
				}
			}
			ch <- struct{}{}
		}()
	}

	for i := 0; i < numThread*2; i++ {
		go func() {
			conn := newGrpcConn(host)
			defer conn.Close()
			gcli := types.NewChain33Client(conn)

			for tx := range txChan {
				//发送交易
				_, err := gcli.SendTransaction(context.Background(), tx, grpc.UseCompressor("gzip"))

				txPool.Put(tx)
				atomic.AddInt64(&total, 1)
				if err != nil {
					log.Error("sendtx", "err", err)
					time.Sleep(time.Second)

				} else {
					atomic.AddInt64(&success, 1)
				}
			}
			chSend <- struct{}{}
		}()
	}

	for j := 0; j < numThread; j++ {
		<-ch
	}
	close(txChan)
	for k := 0; k < numThread*2; k++ {
		<-chSend
	}
	//打印发送的交易总数
	log.Info("sendtx total tx", "total", total)
	//打印成功发送的交易总数
	log.Info("sendtx success tx", "success", success)
}

// PerfV2
func PerfV2(host, txsize, sleepinterval, duration string) {
	durInt, _ := strconv.Atoi(duration)
	sizeInt, _ := strconv.Atoi(txsize)
	sleep, _ := strconv.Atoi(sleepinterval)
	numCPU := runtime.NumCPU()
	numThread := numCPU
	numSend := numCPU * 2
	ch := make(chan struct{}, numThread)
	chSend := make(chan struct{}, numSend)
	numInt := 10000
	batchNum := 100
	txChan := make(chan *types.Transaction, numInt)
	var blockHeight int64
	total := int64(0)
	success := int64(0)
	start := time.Now()

	go func() {
		ch <- struct{}{}
		conn := newGrpcConn(host)
		defer conn.Close()
		gcli := types.NewChain33Client(conn)
		for {
			height, err := getHeight(gcli)
			if err != nil {
				log.Error("getHeight", "err", err)
				time.Sleep(time.Second)
			} else {
				atomic.StoreInt64(&blockHeight, height)
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
	<-ch

	for i := 0; i < numThread; i++ {
		go func() {
			ticker := time.NewTicker(time.Duration(durInt) * time.Second)
			defer ticker.Stop()

			_, priv := genaddress()
			beg := time.Now()
		OuterLoop:
			for {
				select {
				case <-ticker.C:
					log.Info("thread duration", "cost", time.Since(beg))
					break OuterLoop
				default:
					//txHeight := atomic.LoadInt64(&blockHeight) + types.LowAllowPackHeight
					for txs := 0; txs < batchNum; txs++ {
						//构造存证交易
						tx := &types.Transaction{Execer: []byte("user.write")}
						tx.To = execAddr
						tx.Fee = 1e6
						tx.Nonce = time.Now().UnixNano()
						//tx.Expire = types.TxHeightFlag + txHeight
						tx.Expire = 0
						tx.Payload = RandStringBytes(sizeInt)
						//交易签名
						//tx.Sign(types.SECP256K1, priv)
						tx.Signature = &types.Signature{Ty: none.ID, Pubkey: priv.PubKey().Bytes()}
						txChan <- tx
					}
				}
			}
			ch <- struct{}{}
		}()
	}

	for i := 0; i < numSend; i++ {
		go func() {
			conn := newGrpcConn(host)
			defer conn.Close()
			gcli := types.NewChain33Client(conn)
			txs := &types.Transactions{Txs: make([]*types.Transaction, 0, batchNum)}
			retryTxs := make([]*types.Transaction, 0, batchNum*2)

			for tx := range txChan {
				txs.Txs = append(txs.Txs, tx)
				if len(retryTxs) > 0 {
					txs.Txs = append(txs.Txs, retryTxs...)
					retryTxs = retryTxs[:0]
				}

				if len(txs.Txs) >= batchNum {
					reps, err := gcli.SendTransactions(context.Background(), txs)
					if err != nil {
						log.Error("sendtxs", "err", err)
						return
					}
					atomic.AddInt64(&total, int64(len(txs.Txs)))

					// retry failed txs
					for index, reply := range reps.GetReplyList() {
						if reply.IsOk {
							continue
						}
						if string(reply.GetMsg()) == types.ErrChannelClosed.Error() {
							return
						}
						if string(reply.GetMsg()) == types.ErrMemFull.Error() ||
							string(reply.GetMsg()) == types.ErrManyTx.Error() {
							retryTxs = append(retryTxs, txs.Txs[index])
						}
					}
					atomic.AddInt64(&success, int64(len(txs.Txs)-len(retryTxs)))
					if len(retryTxs) > 0 {
						time.Sleep(time.Second * time.Duration(sleep))
					}
					txs.Txs = txs.Txs[:0]
				}
			}
			chSend <- struct{}{}
		}()
	}

	for j := 0; j < numThread; j++ {
		<-ch
	}
	close(txChan)
	for k := 0; k < numSend; k++ {
		<-chSend
	}
	log.Info("sendtx duration", "cost", time.Since(start))
	//打印发送的交易总数
	log.Info("sendtx total tx", "total", total)
	//打印成功发送的交易总数
	log.Info("sendtx success tx", "success", success)
}

func getHeight(gcli types.Chain33Client) (int64, error) {
	header, err := gcli.GetLastHeader(context.Background(), &types.ReqNil{})
	if err != nil {
		log.Error("getHeight", "err", err)
		return 0, err
	}
	return header.Height, nil
}

var txPool = sync.Pool{
	New: func() interface{} {
		tx := &types.Transaction{Execer: []byte(testExec)}
		return tx
	},
}

func newGrpcConn(host string) *grpc.ClientConn {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	for err != nil {
		log.Error("grpc dial", "err", err)
		time.Sleep(time.Millisecond * 100)
		conn, err = grpc.Dial(host, grpc.WithInsecure())
	}
	return conn
}

// PerfOld ...
func PerfOld(ip, size, num, interval, duration string) {
	var numThread int
	numInt, err := strconv.Atoi(num)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	intervalInt, err := strconv.Atoi(interval)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	durInt, err := strconv.Atoi(duration)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if numInt < 10 {
		numThread = 1
	} else if numInt > 100 {
		numThread = 10
	} else {
		numThread = numInt / 10
	}
	maxTxPerAcc := 50
	ch := make(chan struct{}, numThread)
	for i := 0; i < numThread; i++ {
		go func() {
			txCount := 0
			_, priv := genaddress()
			for sec := 0; durInt == 0 || sec < durInt; {
				setTxHeight(ip)
				for txs := 0; txs < numInt/numThread; txs++ {
					if txCount >= maxTxPerAcc {
						_, priv = genaddress()
						txCount = 0
					}
					Put(ip, size, common.ToHex(priv.Bytes()))
					txCount++
				}
				time.Sleep(100 * time.Millisecond * time.Duration(intervalInt))
				sec += intervalInt
			}
			ch <- struct{}{}
		}()
	}
	for j := 0; j < numThread; j++ {
		<-ch
	}
}

// Put ...
func Put(ip string, size string, privkey string) {
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	url := "http://" + ip + ":8801"
	if privkey == "" {
		_, priv := genaddress()
		privkey = common.ToHex(priv.Bytes())
	}
	payload := RandStringBytes(sizeInt)
	//fmt.Println("payload:", common.ToHex(payload))

	tx := &types.Transaction{Execer: []byte("user.write"), Payload: payload, Fee: 1e6}
	tx.To = address.ExecAddress("user.write")
	tx.Expire = TxHeightOffset + types.TxHeightFlag
	tx.Sign(types.SECP256K1, getprivkey(privkey))
	poststr := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"data":"%v"}]}`,
		common.ToHex(types.Encode(tx)))

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(poststr))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("returned JSON: %s\n", string(b))
}

// Get ...
func Get(ip string, hash string) {
	url := "http://" + ip + ":8801"
	fmt.Println("transaction hash:", hash)

	poststr := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"%s"}]}`, hash)
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(poststr))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("returned JSON: %s\n", string(b))
}

func setTxHeight(ip string) {
	url := "http://" + ip + ":8801"
	poststr := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"Chain33.GetLastHeader","params":[]}`)
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(poststr))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Printf("returned JSON: %s\n", string(b))
	msg := &RespMsg{}
	err = json.Unmarshal(b, msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	TxHeightOffset = msg.Result.Height
	fmt.Println("TxHeightOffset:", TxHeightOffset)
}

// RespMsg ...
type RespMsg struct {
	ID     int64           `json:"id"`
	Result rpctypes.Header `json:"result"`
	Err    string          `json:"error"`
}

func getprivkey(key string) crypto.PrivKey {
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
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

func getSignID() int {
	if signType == "bls" {
		return bls.ID
	}
	return types.SECP256K1
}

func genaddress() (string, crypto.PrivKey) {

	cr, err := crypto.Load(types.GetSignName("", getSignID()), -1)
	if err != nil {
		panic(err)
	}
	privto, err := cr.GenKey()
	if err != nil {
		panic(err)
	}
	addrto := address.PubKeyToAddr(address.DefaultID, privto.PubKey().Bytes())
	fmt.Println("addr:", addrto)
	return addrto, privto
}

// RandStringBytes ...
func RandStringBytes(n int) []byte {
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

// ValNode ...
func ValNode(ip, pubkey, power string) {
	url := "http://" + ip + ":8801"

	fmt.Println(pubkey, ":", power)
	pubkeybyte, err := hex.DecodeString(pubkey)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	powerInt, err := strconv.Atoi(power)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	_, priv := genaddress()
	privkey := common.ToHex(priv.Bytes())
	nput := &ty.ValNodeAction_Node{Node: &ty.ValNode{PubKey: pubkeybyte, Power: int64(powerInt)}}
	action := &ty.ValNodeAction{Value: nput, Ty: ty.ValNodeActionUpdate}
	tx := &types.Transaction{Execer: []byte("valnode"), Payload: types.Encode(action), Fee: fee}
	tx.To = address.ExecAddress("valnode")
	tx.Nonce = r.Int63()
	tx.Sign(types.SECP256K1, getprivkey(privkey))

	poststr := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"data":"%v"}]}`,
		common.ToHex(types.Encode(tx)))

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(poststr))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("returned JSON: %s\n", string(b))
}
