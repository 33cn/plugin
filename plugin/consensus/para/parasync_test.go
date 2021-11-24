// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"github.com/33cn/chain33/common/crypto"
	drivers "github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"

	"encoding/hex"
	"sync/atomic"

	"github.com/33cn/chain33/queue"
	typesmocks "github.com/33cn/chain33/types/mocks"
	"github.com/33cn/plugin/plugin/dapp/paracross/testnode"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	//TestPrivateKey 测试私钥
	TestPrivateKey = "6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
	//TestBlockTime 测试时间搓
	TestBlockTime = 1514533390
	//TestMaxCacheCount 测试本地DB最大缓冲数
	TestMaxCacheCount = 100
	//TestLoopCount   测试轮数
	TestMaxLoopCount = 3
)

var (
	//testLoopCountAtom 设置queue返回的Message是正例还是反例
	testLoopCountAtom int32
	//actionReturnIndexAtom 对应getNextAction的每一步返回顺序
	actionReturnIndexAtom int32
)

//测试初始化
func initTestSyncBlock() {
	//println("initSyncBlock")
}

//新建一个para测试实例并初始化一些参数
func createParaTestInstance(t *testing.T, q queue.Queue) *client {
	para := new(client)
	para.subCfg = new(subConfig)

	baseCli := drivers.NewBaseClient(&types.Consensus{Name: "name"})
	para.BaseClient = baseCli

	para.InitClient(q.Client(), initTestSyncBlock)

	//生成rpc Client
	grpcClient := &typesmocks.Chain33Client{}
	para.grpcClient = grpcClient

	//生成私钥
	pk, err := hex.DecodeString(TestPrivateKey)
	assert.Nil(t, err)
	secp, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)
	para.minerPrivateKey = priKey

	//实例化BlockSyncClient
	para.blockSyncClient = &blockSyncClient{
		paraClient:      para,
		notifyChan:      make(chan bool),
		quitChan:        make(chan struct{}),
		maxCacheCount:   TestMaxCacheCount,
		maxSyncErrCount: 100,
	}

	para.commitMsgClient = &commitMsgClient{
		paraClient: para,
	}
	return para
}

//生成创世区块测试数据
func makeGenesisBlockInputTestData() *types.Block {
	newBlock := &types.Block{}
	newBlock.Height = 0
	newBlock.BlockTime = TestBlockTime
	newBlock.ParentHash = zeroHash[:]
	newBlock.MainHash = []byte("genesisHash")
	newBlock.MainHeight = 0

	return newBlock
}

//生成创世区块响应测试数据
func makeGenesisBlockReplyTestData(testLoopCount int32) interface{} {
	switch testLoopCount {
	case 0:
		return &types.BlockDetail{}
	default:
		return errors.New("error")
	}
}

//生成getNextAction不同返回情况下的同步测试数据
//index 对应getNextAction的每一步返回顺序，按return的先后顺序索引
//testLoopCount 测试轮数
func makeSyncReplyTestData(index int32, testLoopCount int32) (
	interface{}, //*types.Block, //GetLastBlock reply
	interface{}, //*types.LocalReplyValue, //GetLastLocalHeight reply
	interface{}, //*types.LocalReplyValue, //GetLocalBlockByHeight reply
	interface{}, //*types.BlockDetail, //writeBlock reply
	interface{}, //*types.BlockDetails, //rollbackBlock  reply
	interface{}) { //*types.Reply) { //rollbackBlock  reply

	detail := &types.BlockDetail{Block: &types.Block{}}
	details := &types.BlockDetails{Items: []*types.BlockDetail{detail}}

	err := errors.New("error")

	switch index {
	case 1:
		return err, err, err, err, err, err
	case 2:
		return &types.Block{},
			err, err, err, err, err
	case 3:
		return &types.Block{},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 0})}},
			err, err, err, err
	case 4:
		return &types.Block{Height: 2},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 1})}},
			err, err,
			details,
			&types.Reply{IsOk: testLoopCount == 0}
	case 5:
		return &types.Block{Height: 2},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 2})}},
			err, err, err, err
	case 6:
		localBlock := &pt.ParaLocalDbBlock{MainHash: []byte("hash1"), Height: 2}
		return &types.Block{Height: 2, MainHash: []byte("hash1")},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 2})}},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(localBlock)}},
			err, err, err

	case 7:
		localBlock := &pt.ParaLocalDbBlock{MainHash: []byte("hash2"), Height: 2}
		return &types.Block{Height: 2, MainHash: []byte("hash1")},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 2})}},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(localBlock)}},
			err,
			details,
			&types.Reply{IsOk: testLoopCount == 0}

	case 8:
		return &types.Block{Height: 2, MainHash: []byte("hash1")},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 3})}},
			err, err, err, err
	case 9:
		localBlock := &pt.ParaLocalDbBlock{ParentMainHash: []byte("hash2"), Height: 3}
		return &types.Block{Height: 2, MainHash: []byte("hash1")},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 3})}},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(localBlock)}},
			err,
			details,
			&types.Reply{IsOk: testLoopCount == 0}
	case 10:
		localBlock := &pt.ParaLocalDbBlock{ParentMainHash: []byte("hash1"), Height: 3}
		return &types.Block{Height: 2, MainHash: []byte("hash1")},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 3})}},
			&types.LocalReplyValue{Values: [][]byte{types.Encode(localBlock)}},
			&types.BlockDetail{},
			err, err
	default:
		return err, err, err, err, err, err
	}
}

//生成清理功能Get Reply测试数据
func makeCleanDataGetReplyTestData(clearLocalDBCallCount int32, testLoopCount int32) interface{} {
	switch clearLocalDBCallCount {
	case 1: //testinitFirstLocalHeightIfNeed会调用到
		switch testLoopCount {
		case 0:
			return &types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 1})}}
		default:
			return &types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: -1})}}
		}

	case 2: //testclearLocalOldBlocks会调用到
		switch testLoopCount {
		case 0:
			return &types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 1 + 2*TestMaxCacheCount})}}
		default:
			return &types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 1 + 2*TestMaxCacheCount - 50})}}
		}
	case 3: //testclearLocalOldBlocks会调用到
		switch testLoopCount {
		case 0:
			return &types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 1})}}
		case 1:
			return &types.LocalReplyValue{Values: [][]byte{types.Encode(&types.Int64{Data: 1})}}
		default: //2
			return errors.New("error")
		}
	default:
		return errors.New("error")
	}
}

//生成清理功能Set Reply测试数据
func makeCleanDataSetReplyTestData(testLoopCount int32) interface{} {
	reply := &types.Reply{}
	reply.IsOk = testLoopCount == 0

	return reply
}

//mock queue Message 返回
func mockMessageReply(q queue.Queue) {

	blockChainKey := "blockchain"
	cli := q.Client()
	cli.Sub(blockChainKey)
	//记录消息Call次数,用于loop退出；quitEndCount通过事先统计得出
	//quitCount := int32(0)
	//quitEndCount := int32(111) //TODO: Need a nice loop quit way
	//用于处理数据同步情况下EventGetValueByKey消息的多重返回
	useLocalReply := false
	usrLocalReplyStart := true
	//用于处理数据清理情况下EventGetValueByKey消息的多重返回
	clearLocalDBCallCount := int32(0)

	for msg := range cli.Recv() {

		testLoopCount := atomic.LoadInt32(&testLoopCountAtom)
		getActionReturnIndex := atomic.LoadInt32(&actionReturnIndexAtom)

		switch {
		case getActionReturnIndex > 0:
			//mock数据同步处理消息返回,testsyncBlocksIfNeed会调用到
			lastBlockReply,
				lastLocalReply,
				localReply,
				writeBlockReply,
				getBlocksReply,
				rollBlockReply := makeSyncReplyTestData(getActionReturnIndex, testLoopCount)

			switch msg.Ty {
			case types.EventGetLastBlock:
				//quitCount++

				msg.Reply(cli.NewMessage(blockChainKey, types.EventBlock, lastBlockReply))

			case types.EventAddParaChainBlockDetail:
				//quitCount++

				msg.Reply(cli.NewMessage(blockChainKey, types.EventReply, writeBlockReply))

			case types.EventDelParaChainBlockDetail:
				//quitCount++

				msg.Reply(cli.NewMessage(blockChainKey, types.EventReply, rollBlockReply))

			case types.EventGetValueByKey:
				//quitCount++

				switch {
				case getActionReturnIndex > 4:
					if usrLocalReplyStart {
						usrLocalReplyStart = false
						useLocalReply = false
					} else {
						useLocalReply = !useLocalReply
					}
				default:
					useLocalReply = false
					usrLocalReplyStart = true

				}

				if !useLocalReply {
					msg.Reply(cli.NewMessage(blockChainKey, types.EventLocalReplyValue, lastLocalReply))
				} else {
					msg.Reply(cli.NewMessage(blockChainKey, types.EventLocalReplyValue, localReply))
				}

			case types.EventGetBlocks:
				//quitCount++

				msg.Reply(cli.NewMessage(blockChainKey, types.EventBlocks, getBlocksReply))
			default:
				//nothing
			}
		default:
			switch msg.Ty {
			case types.EventAddParaChainBlockDetail: //mock创世区块创建消息返回,testCreateGenesisBlock
				//quitCount++

				reply := makeGenesisBlockReplyTestData(testLoopCount)
				msg.Reply(cli.NewMessage(blockChainKey, types.EventReply, reply))

			case types.EventGetValueByKey: //mock数据清理消息返回
				//quitCount++

				clearLocalDBCallCount++
				reply := makeCleanDataGetReplyTestData(clearLocalDBCallCount, testLoopCount)
				msg.Reply(cli.NewMessage(blockChainKey, types.EventLocalReplyValue, reply))
				if clearLocalDBCallCount == 3 {
					//正例测试完成，初始化等待互例测试
					clearLocalDBCallCount = 0
				}

			case types.EventSetValueByKey: //mock数据清理消息返回,testclearLocalOldBlocks会调用到
				//quitCount++

				reply := makeCleanDataSetReplyTestData(testLoopCount)

				msg.Reply(cli.NewMessage(blockChainKey, types.EventReply, reply))
			default:
				//nothing
			}
		}

		//println(quitCount)
		//if quitCount == quitEndCount {
		//	break
		//}
	}
}

//测试创世区块写入
func testCreateGenesisBlock(t *testing.T, para *client, testLoopCount int32) {
	genesisBlock := makeGenesisBlockInputTestData()
	err := para.blockSyncClient.createGenesisBlock(genesisBlock)

	switch testLoopCount {
	case 0:
		assert.Nil(t, err)
	default:
		assert.Error(t, err)
	}

}

//测试清理localdb缓存数据
func testClearLocalOldBlocks(t *testing.T, para *client, testLoopCount int32) {
	err := para.blockSyncClient.clearLocalOldBlocks()

	switch testLoopCount {
	case 0:
		assert.Nil(t, err)
	case 1:
		assert.Equal(t, true, err == nil)
	default: //2
		assert.Error(t, err)
	}
}

//测试初始化开始高度
func testInitFirstLocalHeightIfNeed(t *testing.T, para *client, testLoopCount int32) {
	err := para.blockSyncClient.initFirstLocalHeightIfNeed()

	switch testLoopCount {
	case 0:
		assert.Nil(t, err)
	default:
		assert.Error(t, err)
	}
}

//测试一次区块同步操作
func testSyncBlocksIfNeed(t *testing.T, para *client, testLoopCount int32) {
	errorCount := int32(0)
	for i := int32(1); i <= 10; i++ {
		atomic.StoreInt32(&actionReturnIndexAtom, i)
		isSynced, err := para.blockSyncClient.syncBlocksIfNeed()
		if err != nil {
			errorCount++
		}
		assert.Equalf(t, isSynced, i == 3 || i == 6, "i=%d", i)
	}

	switch testLoopCount {
	case 0:
		assert.Equal(t, true, errorCount == 4)
	default:
		assert.Equal(t, true, errorCount == 7)
	}

	atomic.StoreInt32(&actionReturnIndexAtom, 0)
}

//测试SyncHasCaughtUp
func testSyncHasCaughtUp(t *testing.T, para *client, testLoopCount int32) {
	oldValue := para.blockSyncClient.syncHasCaughtUp()
	para.blockSyncClient.setSyncCaughtUp(true)
	isSyncHasCaughtUp := para.blockSyncClient.syncHasCaughtUp()
	para.blockSyncClient.setSyncCaughtUp(oldValue)

	assert.Equal(t, true, isSyncHasCaughtUp)
}

//测试getBlockSyncState
func testGetBlockSyncState(t *testing.T, para *client, testLoopCount int32) {
	oldValue := para.blockSyncClient.getBlockSyncState()
	para.blockSyncClient.setBlockSyncState(blockSyncStateFinished)
	syncState := para.blockSyncClient.getBlockSyncState()
	para.blockSyncClient.setBlockSyncState(oldValue)

	assert.Equal(t, true, syncState == blockSyncStateFinished)
}

//执行所有函数测试
func execTest(t *testing.T, para *client, testLoopCount int32) {
	atomic.StoreInt32(&actionReturnIndexAtom, 0)
	atomic.StoreInt32(&testLoopCountAtom, testLoopCount)

	testCreateGenesisBlock(t, para, testLoopCount)
	testSyncBlocksIfNeed(t, para, testLoopCount)
	testInitFirstLocalHeightIfNeed(t, para, testLoopCount)
	testClearLocalOldBlocks(t, para, testLoopCount)

	testSyncHasCaughtUp(t, para, testLoopCount)
	testGetBlockSyncState(t, para, testLoopCount)
}

//测试入口
func TestSyncBlocks(t *testing.T) {
	cfg := types.NewChain33Config(testnode.DefaultConfig)
	q := queue.New("channel")
	q.SetConfig(cfg)
	defer q.Close()
	para := createParaTestInstance(t, q)
	go q.Start()
	go mockMessageReply(q)
	//测试分多轮测试，每一轮测模拟不同的测试数据输入,包括正常数据和异常数据
	for i := int32(0); i <= TestMaxLoopCount-1; i++ {
		execTest(t, para, i)
	}

}
