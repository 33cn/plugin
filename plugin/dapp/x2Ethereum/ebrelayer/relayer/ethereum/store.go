package ethereum

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	dbm "github.com/33cn/chain33/common/db"
	chain33Types "github.com/33cn/chain33/types"
	ebTypes "github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/utils"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	eth2chain33TxHashPrefix  = "Eth2chain33TxHash"
	eth2chain33TxTotalAmount = []byte("Eth2chain33TxTotalAmount")

	chain33ToEthTxHashPrefix  = "chain33ToEthTxHash"
	chain33ToEthTxTotalAmount = []byte("chain33ToEthTxTotalAmount")

	bridgeRegistryAddrPrefix = []byte("x2EthBridgeRegistryAddr")

	chain33BridgeLogProcessedAt = []byte("chain33BridgeLogProcessedAt")
	bridgeBankLogProcessedAt    = []byte("bridgeBankLogProcessedAt")

	ethTxEventPrefix               = []byte("ethTxEventPrefix")
	lastBridgeBankHeightProcPrefix = []byte("lastBridgeBankHeight")

	bridgeBankTxProCnt = []byte("bridgeBankTxProCnt")
)

func ethTxEventKey4Height(height uint64, index uint32) []byte {
	return append(ethTxEventPrefix, []byte(fmt.Sprintf("%020d-%d", height, index))...)
}

func calcRelay2Chain33Txhash(txindex int64) []byte {
	return []byte(fmt.Sprintf("%s-%012d", eth2chain33TxHashPrefix, txindex))
}

func calcRelay2EthTxhash(txindex int64) []byte {
	return []byte(fmt.Sprintf("%s-%012d", chain33ToEthTxHashPrefix, txindex))
}

func (ethRelayer *EthereumRelayer) setBridgeRegistryAddr(bridgeRegistryAddr string) error {
	return ethRelayer.db.Set(bridgeRegistryAddrPrefix, []byte(bridgeRegistryAddr))
}

func (ethRelayer *EthereumRelayer) getBridgeRegistryAddr() (string, error) {
	addr, err := ethRelayer.db.Get(bridgeRegistryAddrPrefix)
	if nil != err {
		return "", err
	}
	return string(addr), nil
}

func (ethRelayer *EthereumRelayer) updateTotalTxAmount2chain33(total int64) error {
	totalTx := &chain33Types.Int64{
		Data: atomic.LoadInt64(&ethRelayer.totalTx4Eth2Chain33),
	}
	//更新成功见证的交易数
	return ethRelayer.db.Set(eth2chain33TxTotalAmount, chain33Types.Encode(totalTx))
}

func (ethRelayer *EthereumRelayer) setLastestRelay2Chain33Txhash(txhash string, txIndex int64) error {
	key := calcRelay2Chain33Txhash(txIndex)
	return ethRelayer.db.Set(key, []byte(txhash))
}

func (ethRelayer *EthereumRelayer) updateTotalTxAmount2Eth(total int64) error {
	totalTx := &chain33Types.Int64{
		Data: atomic.LoadInt64(&ethRelayer.totalTx4Chain33ToEth),
	}
	//更新成功见证的交易数
	return ethRelayer.db.Set(chain33ToEthTxTotalAmount, chain33Types.Encode(totalTx))
}

func (ethRelayer *EthereumRelayer) setLastestRelay2EthTxhash(txhash string, txIndex int64) error {
	key := calcRelay2EthTxhash(txIndex)
	return ethRelayer.db.Set(key, []byte(txhash))
}

func (ethRelayer *EthereumRelayer) queryTxhashes(prefix []byte) []string {
	return utils.QueryTxhashes(prefix, ethRelayer.db)
}

func (ethRelayer *EthereumRelayer) setHeight4chain33BridgeLogAt(height uint64) error {
	return ethRelayer.setLogProcHeight(chain33BridgeLogProcessedAt, height)
}

func (ethRelayer *EthereumRelayer) getHeight4chain33BridgeLogAt() uint64 {
	return ethRelayer.getLogProcHeight(chain33BridgeLogProcessedAt)
}

func (ethRelayer *EthereumRelayer) setHeight4BridgeBankLogAt(height uint64) error {
	return ethRelayer.setLogProcHeight(bridgeBankLogProcessedAt, height)
}

func (ethRelayer *EthereumRelayer) getHeight4BridgeBankLogAt() uint64 {
	return ethRelayer.getLogProcHeight(bridgeBankLogProcessedAt)
}

func (ethRelayer *EthereumRelayer) setLogProcHeight(key []byte, height uint64) error {
	data := &ebTypes.Uint64{
		Data: height,
	}
	return ethRelayer.db.Set(key, chain33Types.Encode(data))
}

func (ethRelayer *EthereumRelayer) getLogProcHeight(key []byte) uint64 {
	value, err := ethRelayer.db.Get(key)
	if nil != err {
		return 0
	}
	var height ebTypes.Uint64
	err = chain33Types.Decode(value, &height)
	if nil != err {
		return 0
	}
	return height.Data
}

//保存处理过的交易
func (ethRelayer *EthereumRelayer) setTxProcessed(txhash []byte) error {
	return ethRelayer.db.Set(txhash, []byte("1"))
}

//判断是否已经被处理，如果能够在数据库中找到该笔交易，则认为已经被处理
func (ethRelayer *EthereumRelayer) checkTxProcessed(txhash []byte) bool {
	_, err := ethRelayer.db.Get(txhash)
	if nil != err {
		return false
	}
	return true
}

func (ethRelayer *EthereumRelayer) setEthTxEvent(vLog types.Log) error {
	key := ethTxEventKey4Height(vLog.BlockNumber, uint32(vLog.TxIndex))
	value, err := json.Marshal(vLog)
	if nil != err {
		return err
	}
	return ethRelayer.db.Set(key, value)
}

func (ethRelayer *EthereumRelayer) getEthTxEvent(blockNumber uint64, txIndex uint32) (*types.Log, error) {
	key := ethTxEventKey4Height(blockNumber, txIndex)
	data, err := ethRelayer.db.Get(key)
	if nil != err {
		return nil, err
	}
	log := types.Log{}
	err = json.Unmarshal(data, &log)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (ethRelayer *EthereumRelayer) getNextValidEthTxEventLogs(height uint64, index uint32, fetchCnt int32) ([]*types.Log, error) {
	key := ethTxEventKey4Height(height, index)
	helper := dbm.NewListHelper(ethRelayer.db)
	datas := helper.List(ethTxEventPrefix, key, fetchCnt, dbm.ListASC)
	if nil == datas {
		return nil, nil
	}
	var logs []*types.Log
	for _, data := range datas {
		log := &types.Log{}
		err := json.Unmarshal(data, log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (ethRelayer *EthereumRelayer) setBridgeBankProcessedHeight(height uint64, index uint32) {
	bytes := chain33Types.Encode(&ebTypes.EventLogIndex{
		Height: height,
		Index:  index})
	_ = ethRelayer.db.Set(lastBridgeBankHeightProcPrefix, bytes)
}

func (ethRelayer *EthereumRelayer) getLastBridgeBankProcessedHeight() ebTypes.EventLogIndex {
	data, err := ethRelayer.db.Get(lastBridgeBankHeightProcPrefix)
	if nil != err {
		return ebTypes.EventLogIndex{}
	}
	logIndex := ebTypes.EventLogIndex{}
	_ = chain33Types.Decode(data, &logIndex)
	return logIndex
}

//构建一个引导查询使用的bridgeBankTx
func (ethRelayer *EthereumRelayer) initBridgeBankTx() {
	log, _ := ethRelayer.getEthTxEvent(0, 0)
	if nil != log {
		return
	}
	_ = ethRelayer.setEthTxEvent(types.Log{})
}
