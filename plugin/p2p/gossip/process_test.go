// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package gossip

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

type versionData struct {
	peerName string
	rawData  interface{}
	version  int32
}

func Test_processP2P(t *testing.T) {
	cfg := types.NewChain33Config(types.ReadFile("../../../chain33.toml"))
	q := queue.New("channel")
	q.SetConfig(cfg)
	go q.Start()
	p2p := newP2p(cfg, 12345, "testProcessP2p", q)
	p2p.subCfg.MinLtBlockSize = 0
	defer freeP2p(p2p)
	defer q.Close()
	node := p2p.node
	client := p2p.client
	pid := "testPid"
	sendChan := make(chan interface{}, 1)
	recvChan := make(chan *types.BroadCastData, 1)

	payload := []byte("testpayload")
	minerTx := &types.Transaction{Execer: []byte("coins"), Payload: payload, Fee: 14600, Expire: 200}
	tx := &types.Transaction{Execer: []byte("coins"), Payload: payload, Fee: 4600, Expire: 2}
	tx1 := &types.Transaction{Execer: []byte("coins"), Payload: payload, Fee: 460000000, Expire: 0}
	tx2 := &types.Transaction{Execer: []byte("coins"), Payload: payload, Fee: 100, Expire: 1}
	txGroup, _ := types.CreateTxGroup([]*types.Transaction{tx1, tx2}, cfg.GetMinTxFeeRate())
	gtx := txGroup.Tx()
	txList := append([]*types.Transaction{}, minerTx, tx, tx1, tx2)
	memTxList := append([]*types.Transaction{}, tx, gtx)

	block := &types.Block{
		TxHash: []byte("123"),
		Height: 10,
		Txs:    txList,
	}
	txHash := hex.EncodeToString(tx.Hash())
	blockHash := hex.EncodeToString(block.Hash(cfg))
	rootHash := merkle.CalcMerkleRoot(cfg, block.Height, txList)

	//mempool handler
	go func() {
		client := q.Client()
		client.Sub("mempool")
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventTxListByHash:
				query := msg.Data.(*types.ReqTxHashList)
				var txs []*types.Transaction
				if !query.IsShortHash {
					txs = memTxList[:1]
				} else {
					txs = memTxList
				}
				msg.Reply(client.NewMessage("p2p", types.EventTxListByHash, &types.ReplyTxList{Txs: txs}))
			}
		}
	}()

	//测试发送
	go func() {
		for data := range sendChan {
			verData, ok := data.(*versionData)
			assert.True(t, ok)
			sendData, doSend := node.processSendP2P(verData.rawData, verData.version, verData.peerName, "testIP:port")
			txHashFilter.Remove(txHash)
			blockHashFilter.Remove(blockHash)
			assert.True(t, doSend, "sendData:", verData.rawData)
			recvChan <- sendData
		}
	}()
	//测试接收
	go func() {
		for data := range recvChan {
			txHashFilter.Remove(txHash)
			blockHashFilter.Remove(blockHash)
			handled := node.processRecvP2P(data, pid, node.pubToPeer, "testIP:port")
			assert.True(t, handled)
		}
	}()

	go func() {
		p2pChan := node.pubsub.Sub("tx")
		for data := range p2pChan {
			if p2pTx, ok := data.(*types.P2PTx); ok {
				sendChan <- &versionData{rawData: p2pTx, version: lightBroadCastVersion}
			}
		}
	}()

	//data test
	subChan := node.pubsub.Sub(pid)
	//全数据广播
	sendChan <- &versionData{peerName: pid + "1", rawData: &types.P2PTx{Tx: tx, Route: &types.P2PRoute{}}, version: lightBroadCastVersion - 1}
	p2p.mgr.PubSub.Pub(client.NewMessage("p2p", types.EventTxBroadcast, tx), P2PTypeName)
	sendChan <- &versionData{peerName: pid + "1", rawData: &types.P2PBlock{Block: block}, version: lightBroadCastVersion - 1}
	//交易发送过滤
	txHashFilter.Add(hex.EncodeToString(tx1.Hash()), &types.P2PRoute{TTL: DefaultLtTxBroadCastTTL})
	p2p.mgr.PubSub.Pub(client.NewMessage("p2p", types.EventTxBroadcast, tx1), P2PTypeName)
	//交易短哈希广播
	sendChan <- &versionData{peerName: pid + "2", rawData: &types.P2PTx{Tx: tx, Route: &types.P2PRoute{TTL: DefaultLtTxBroadCastTTL}}, version: lightBroadCastVersion}
	recvWithTimeout(t, subChan, "case 1") //缺失交易，从对端获取
	//区块短哈希广播
	sendChan <- &versionData{peerName: pid + "2", rawData: &types.P2PBlock{Block: block}, version: lightBroadCastVersion}
	recvWithTimeout(t, subChan, "case 2")
	assert.True(t, ltBlockCache.Contains(blockHash))
	cpBlock := *ltBlockCache.Get(blockHash).(*types.Block)
	assert.True(t, bytes.Equal(rootHash, merkle.CalcMerkleRoot(cfg, cpBlock.Height, cpBlock.Txs)))

	//query tx
	sendChan <- &versionData{rawData: &types.P2PQueryData{Value: &types.P2PQueryData_TxReq{TxReq: &types.P2PTxReq{TxHash: tx.Hash()}}}}
	data := recvWithTimeout(t, subChan, "case 3")
	_, ok := data.(*types.P2PTx)
	assert.True(t, ok)
	sendChan <- &versionData{rawData: &types.P2PQueryData{Value: &types.P2PQueryData_BlockTxReq{BlockTxReq: &types.P2PBlockTxReq{
		BlockHash: blockHash,
		TxIndices: []int32{1, 2},
	}}}}
	data = recvWithTimeout(t, subChan, "case 4")
	rep, ok := data.(*types.P2PBlockTxReply)
	assert.True(t, ok)
	assert.Equal(t, 2, int(rep.TxIndices[1]))
	sendChan <- &versionData{rawData: &types.P2PQueryData{Value: &types.P2PQueryData_BlockTxReq{BlockTxReq: &types.P2PBlockTxReq{
		BlockHash: blockHash,
		TxIndices: nil,
	}}}}
	data = recvWithTimeout(t, subChan, "case 5")
	rep, ok = data.(*types.P2PBlockTxReply)
	assert.True(t, ok)
	assert.Nil(t, rep.TxIndices)

	//query reply
	sendChan <- &versionData{rawData: &types.P2PBlockTxReply{
		BlockHash: blockHash,
		TxIndices: []int32{1},
		Txs:       txList[1:2],
	}}
	rep1, ok := recvWithTimeout(t, subChan, "case 6").(*types.P2PQueryData)
	assert.True(t, ok)
	assert.Nil(t, rep1.GetBlockTxReq().GetTxIndices())
	sendChan <- &versionData{rawData: &types.P2PBlockTxReply{
		BlockHash: blockHash,
		Txs:       txList[0:],
	}}
	for ltBlockCache.Contains(blockHash) {
		time.Sleep(time.Millisecond)
	}
	//send tx with max ttl
	_, doSend := node.processSendP2P(&types.P2PTx{Tx: tx, Route: &types.P2PRoute{TTL: node.nodeInfo.cfg.MaxTTL + 1}}, lightBroadCastVersion, pid+"5", "testIP:port")
	assert.False(t, doSend)
}

// 等待接收channel数据，超时报错
func recvWithTimeout(t *testing.T, ch chan interface{}, testCase string) interface{} {
	select {
	case data := <-ch:
		return data
	case <-time.After(time.Second * 10):
		t.Error(testCase, "waitChanTimeout")
		t.FailNow()
	}
	return nil
}
