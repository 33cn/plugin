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

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log/log15"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/valnode/types"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

const fee = 1e6
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789-=_+=/<>!@#$%^&"

var r *rand.Rand

// TxHeightOffset needed
var TxHeightOffset int64

func main() {
	if len(os.Args) == 1 || os.Args[1] == "-h" {
		LoadHelp()
		return
	}
	fmt.Println("jrpc url:", os.Args[2]+":8801")
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	argsWithoutProg := os.Args[1:]
	switch argsWithoutProg[0] {
	case "-h": //使用帮助
		LoadHelp()
	case "perf":
		if len(argsWithoutProg) != 6 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		Perf(argsWithoutProg[1], argsWithoutProg[2], argsWithoutProg[3], argsWithoutProg[4], argsWithoutProg[5])
	case "put":
		if len(argsWithoutProg) != 3 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		Put(argsWithoutProg[1], argsWithoutProg[2], "")
	case "get":
		if len(argsWithoutProg) != 3 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		Get(argsWithoutProg[1], argsWithoutProg[2])
	case "valnode":
		if len(argsWithoutProg) != 4 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		ValNode(argsWithoutProg[1], argsWithoutProg[2], argsWithoutProg[3])
	case "perfOld":
		if len(argsWithoutProg) != 6 {
			fmt.Print(errors.New("参数错误").Error())
			return
		}
		PerfOld(argsWithoutProg[1], argsWithoutProg[2], argsWithoutProg[3], argsWithoutProg[4], argsWithoutProg[5])
	}
}

// LoadHelp ...
func LoadHelp() {
	fmt.Println("Available Commands:")
	fmt.Println("perf [host, size, num, interval, duration]                   : 写数据性能测试，interval单位为100毫秒，host形式为ip:port")
	fmt.Println("put  [ip, size]                                              : 写数据")
	fmt.Println("get  [ip, hash]                                              : 读数据")
	fmt.Println("valnode [ip, pubkey, power]                                  : 增加/删除/修改tendermint节点")
	fmt.Println("perfOld [ip, size, num, interval, duration]                  : 不推荐使用，写数据性能测试，interval单位为100毫秒")
}

// Perf 性能测试
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
					tx.Fee = rand.Int63()
					tx.Nonce = time.Now().UnixNano()
					tx.Expire = height + types.TxHeightFlag + types.LowAllowPackHeight
					tx.Payload = RandStringBytes(sizeInt)
					//交易签名
					tx.Sign(types.SECP256K1, priv)
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
					if strings.Contains(err.Error(), "ErrTxExpire") {
						continue
					}
					if strings.Contains(err.Error(), "ErrMemFull") {
						time.Sleep(time.Second)
						continue
					}

					log.Error("sendtx", "err", err)
					time.Sleep(time.Second)
					//conn.Close()
					//conn = newGrpcConn(ip)
					//gcli = types.NewChain33Client(conn)
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

var (
	log      = log15.New()
	execAddr = address.ExecAddress("user.write")
)

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
		tx := &types.Transaction{Execer: []byte("user.write")}
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
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
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

func genaddress() (string, crypto.PrivKey) {
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		panic(err)
	}
	privto, err := cr.GenKey()
	if err != nil {
		panic(err)
	}
	addrto := address.PubKeyToAddress(privto.PubKey().Bytes())
	fmt.Println("addr:", addrto.String())
	return addrto.String(), privto
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
