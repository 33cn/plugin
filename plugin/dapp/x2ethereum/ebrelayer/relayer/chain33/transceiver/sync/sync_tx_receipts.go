// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//Package sync ...
package sync

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/33cn/chain33/blockchain"
	dbm "github.com/33cn/chain33/common/db"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/jsonclient"
	"github.com/33cn/chain33/types"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	"github.com/rs/cors"
)

var (
	log            = l.New("module", "sync.tx_receipts")
	syncTxReceipts *TxReceipts
)

//StartSyncTxReceipt ...
func StartSyncTxReceipt(cfg *relayerTypes.SyncTxReceiptConfig, db dbm.DB) *TxReceipts {
	log.Debug("StartSyncTxReceipt, load config", "para:", cfg)
	log.Debug("TxReceipts started ")

	bindOrResumePush(cfg)
	syncTxReceipts = NewSyncTxReceipts(db)
	go syncTxReceipts.SaveAndSyncTxs2Relayer()
	go startHTTPService(cfg.PushBind, "*")
	return syncTxReceipts
}

//func StopSyncTxReceipt() {
//	syncTxReceipts.Stop()
//}

func startHTTPService(url string, clientHost string) {
	listen, err := net.Listen("tcp", url)
	if err != nil {
		panic(err)
	}
	var handler http.Handler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			//fmt.Println(r.URL, r.Header, r.Body)
			beg := types.Now()
			defer func() {
				log.Info("handler", "cost", types.Since(beg))
			}()

			client := strings.Split(r.RemoteAddr, ":")[0]
			if !checkClient(client, clientHost) {
				log.Error("HandlerFunc", "client", r.RemoteAddr, "expect", clientHost)
				_, _ = w.Write([]byte(`{"errcode":"-1","result":null,"msg":"reject"}`))
				// unbind 逻辑有问题， 需要的外部处理
				//  切换外部服务时， 可能换 name
				// 收到一个不是client 的请求，很有可能是以前注册过的， 取消掉
				//unbind(client)
				return
			}

			if r.URL.Path == "/" {
				w.Header().Set("Content-type", "application/json")
				w.WriteHeader(200)
				if len(r.Header["Content-Encoding"]) >= 1 && r.Header["Content-Encoding"][0] == "gzip" {
					gr, err := gzip.NewReader(r.Body)
					if err != nil {
						log.Debug("Error while reader serving JSON request: %v", err)
						return
					}
					//body := make([]byte, r.ContentLength)
					body, err := ioutil.ReadAll(gr)
					//n, err := r.Body.Read(body)
					if err != nil {
						log.Debug("Error while serving JSON request: %v", err)
						return
					}

					err = handleRequest(body)
					if err == nil {
						_, _ = w.Write([]byte("OK"))
					} else {
						_, _ = w.Write([]byte(err.Error()))
					}
				}
			}
		})

	co := cors.New(cors.Options{})
	handler = co.Handler(handler)

	_ = http.Serve(listen, handler)
}

func handleRequest(body []byte) error {
	beg := types.Now()
	defer func() {
		log.Info("handleRequest", "cost", types.Since(beg))
	}()

	var req types.TxReceipts4Subscribe
	err := types.Decode(body, &req)
	if err != nil {
		log.Error("handleRequest", "DecodeBlockSeqErr", err)
		return err
	}
	err = pushTxReceipts(&req)
	return err
}

func checkClient(addr string, expectClient string) bool {
	if expectClient == "0.0.0.0" || expectClient == "*" {
		return true
	}
	return addr == expectClient
}

//向chain33节点的注册推送交易回执，AddSubscribeTxReceipt具有2种功能：
//首次注册功能，如果没有进行过注册，则进行首次注册
//如果已经注册，则继续推送
func bindOrResumePush(cfg *relayerTypes.SyncTxReceiptConfig) {
	contract := make(map[string]bool)
	contract["x2ethereum"] = true
	params := types.PushSubscribeReq{
		Name:          cfg.PushName,
		URL:           cfg.PushHost,
		Encode:        "proto",
		LastSequence:  cfg.StartSyncSequence,
		LastHeight:    cfg.StartSyncHeight,
		LastBlockHash: cfg.StartSyncHash,
		Type:          int32(blockchain.PushTxReceipt),
		Contract:      contract,
	}
	var res types.ReplySubscribePush
	ctx := jsonclient.NewRPCCtx(cfg.Chain33Host, "Chain33.AddPushSubscribe", params, &res)
	_, err := ctx.RunResult()
	if err != nil {
		fmt.Println("Failed to AddSubscribeTxReceipt to  rpc addr:", cfg.Chain33Host, "ReplySubTxReceipt", res)
		panic("bindOrResumePush client failed due to:" + err.Error())
	}
	if !res.IsOk {
		fmt.Println("Failed to AddSubscribeTxReceipt to  rpc addr:", cfg.Chain33Host, "ReplySubTxReceipt", res)
		panic("bindOrResumePush client failed due to:" + res.Msg)
	}
	log.Info("bindOrResumePush", "Succeed to AddSubscribeTxReceipt for rpc address:", cfg.Chain33Host)
	fmt.Println("Succeed to AddPushSubscribe")
}
