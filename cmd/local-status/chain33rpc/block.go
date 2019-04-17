// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package chain33rpc 实现了向chain33 节点发送rpc请求
package chain33rpc

import (
	"fmt"
	"os"
	"time"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	lru "github.com/hashicorp/golang-lru"
)

var log = l.New("module", "chain33rpc")

// ts/height -> blockHeader
type chain33 struct {
	lastHeader *rpctypes.Header
	Host       string
	cache      *lru.Cache
}

var chain = chain33{
	lastHeader: &rpctypes.Header{Height: 0},
}

func init() {
	cache, err := lru.New(200)
	if err != nil {
		panic(err)
	}
	chain.cache = cache

}

func getLastHeader(cli *jsonclient.JSONClient) (*rpctypes.Header, error) {
	method := "Chain33.GetLastHeader"
	var res rpctypes.Header
	err := cli.Call(method, nil, &res)
	return &res, err
}

func getHeaders(cli *jsonclient.JSONClient, start, end int64) (*rpctypes.Headers, error) {
	method := "Chain33.GetHeaders"
	params := &types.ReqBlocks{Start: start, End: end, IsDetail: false}
	var res rpctypes.Headers
	err := cli.Call(method, params, &res)
	return &res, err
}

func getBlocks(cli *jsonclient.JSONClient, start, end int64) (*rpctypes.BlockDetails, error) {
	method := "Chain33.GetBlocks"
	params := &rpctypes.BlockParam{Start: start, End: end, Isdetail: true}
	var res rpctypes.BlockDetails
	err := cli.Call(method, params, &res)
	return &res, err
}

func syncLastHeader(host string) {
	rpcCli, err := jsonclient.NewJSONClient(host)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	last, err := getLastHeader(rpcCli)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	chain.lastHeader = last
}

func syncBlocks(host string, curHeight int64, blockChan chan interface{}) {
	if chain.cache.Contains(curHeight) {
		b, ok := chain.cache.Get(curHeight)
		if ok {
			blockChan <- b
			return
		}
	}

	rpcCli, err := jsonclient.NewJSONClient(host)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	last, err := getLastHeader(rpcCli)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	lastHeight := last.Height
	if curHeight >= lastHeight {
		return
	}
	if curHeight+10 < lastHeight {
		lastHeight = curHeight + 10
	}

	hs, err := getHeaders(rpcCli, curHeight, lastHeight)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	//fmt.Printf("%+v\n", hs)
	for _, h := range hs.Items { // TODO 下载另外的区块， 放到后台
		fmt.Printf("%+v\n", h)
		blocks, err := getBlocks(rpcCli, h.Height, h.Height+1)
		if err != nil {
			fmt.Fprintln(os.Stderr, err, "GetBlocks")
			return
		}
		fmt.Printf("%+v\n %+v\n", blocks.Items[0].Block, blocks.Items[0].Receipts)
		chain.cache.Add(h.Height, blocks.Items[0])
	}
	fmt.Fprintln(os.Stderr, err, chain.lastHeader.Height)

	if chain.cache.Contains(curHeight) {
		b, ok := chain.cache.Get(curHeight)
		if ok {
			blockChan <- b
			return
		}
	}

	fmt.Fprintln(os.Stderr, err, chain.lastHeader.Height)
}

//SyncBlock 同步区块
func SyncBlock(host string, headerChan chan int64, block chan interface{}) {
	chain.Host = host
	syncLastHeader(host)

	timeout := time.NewTicker(5 * time.Second)
	for {
		select {
		case h := <-headerChan:
			syncBlocks(host, h, block)
		case <-timeout.C:
			syncLastHeader(host)
		}
	}
}
