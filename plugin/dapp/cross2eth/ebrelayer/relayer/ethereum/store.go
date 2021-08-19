package ethereum

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"

	dbm "github.com/33cn/chain33/common/db"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	chain33ToEthStaticsPrefix      = "eth-chain33ToEthStatics"
	chain33ToEthTxTotalAmount      = []byte("eth-chain33ToEthTxTotalAmount")
	bridgeRegistryAddrPrefix       = []byte("eth-x2EthBridgeRegistryAddr")
	bridgeBankLogProcessedAt       = []byte("eth-bridgeBankLogProcessedAt")
	ethTxEventPrefix               = []byte("eth-ethTxEventPrefix")
	lastBridgeBankHeightProcPrefix = []byte("eth-lastBridgeBankHeight")
	ethTokenSymbol2AddrPrefix      = []byte("eth-ethTokenSymbol2AddrPrefix")
	ethTokenSymbol2LockAddrPrefix  = []byte("eth-ethTokenSymbol2LockAddrPrefix")
	ethLockTxUpdateTxIndex         = []byte("eth-ethLockTxUpdateTxIndex")
	ethBurnTxUpdateTxIndex         = []byte("eth-ethBurnTxUpdateTxIndex")
	multiSignAddressPrefix         = []byte("eth-multiSignAddress")
)

func ethTokenSymbol2AddrKey(symbol string) []byte {
	return append(ethTokenSymbol2AddrPrefix, []byte(fmt.Sprintf("-symbol-%s", symbol))...)
}

func ethTokenSymbol2LockAddrKey(symbol string) []byte {
	return append(ethTokenSymbol2LockAddrPrefix, []byte(fmt.Sprintf("-symbol-%s", symbol))...)
}

func ethTxEventKey4Height(height uint64, index uint32) []byte {
	return append(ethTxEventPrefix, []byte(fmt.Sprintf("%020d-%d", height, index))...)
}

func calcRelayFromChain33Key(claimType int32, txindex int64) []byte {
	return []byte(fmt.Sprintf("%s-%d-%012d", chain33ToEthStaticsPrefix, claimType, txindex))
}

func calcRelayFromChain33ListPrefix(claimType int32) []byte {
	return []byte(fmt.Sprintf("%s-%d-", chain33ToEthStaticsPrefix, claimType))
}

func (ethRelayer *Relayer4Ethereum) getStatics(claimType int32, txIndex int64) ([][]byte, error) {
	keyPrefix := calcRelayFromChain33ListPrefix(claimType)

	keyFrom := calcRelayFromChain33Key(claimType, txIndex)
	helper := dbm.NewListHelper(ethRelayer.db)
	datas := helper.List(keyPrefix, keyFrom, 20, dbm.ListASC)
	if nil == datas {
		return nil, errors.New("Not found")
	}

	return datas, nil
}

func (ethRelayer *Relayer4Ethereum) setEthLockTxUpdateTxIndex(txindex int64, claimType events.ClaimType) error {
	txIndexWrapper := &chain33Types.Int64{
		Data: txindex,
	}

	if events.ClaimTypeBurn == claimType {
		return ethRelayer.db.Set(ethBurnTxUpdateTxIndex, chain33Types.Encode(txIndexWrapper))
	}
	return ethRelayer.db.Set(ethLockTxUpdateTxIndex, chain33Types.Encode(txIndexWrapper))
}

func (ethRelayer *Relayer4Ethereum) getEthLockTxUpdateTxIndex(claimType events.ClaimType) int64 {
	var key []byte
	if events.ClaimTypeBurn == claimType {
		key = ethBurnTxUpdateTxIndex
	} else {
		key = ethLockTxUpdateTxIndex
	}
	data, err := ethRelayer.db.Get(key)
	if nil != err {
		return ebTypes.Invalid_Tx_Index
	}

	var txIndexWrapper chain33Types.Int64
	err = chain33Types.Decode(data, &txIndexWrapper)
	if nil != err {
		return ebTypes.Invalid_Tx_Index
	}
	return txIndexWrapper.Data
}

func (ethRelayer *Relayer4Ethereum) setBridgeRegistryAddr(bridgeRegistryAddr string) error {
	return ethRelayer.db.Set(bridgeRegistryAddrPrefix, []byte(bridgeRegistryAddr))
}

func (ethRelayer *Relayer4Ethereum) getBridgeRegistryAddr() (string, error) {
	addr, err := ethRelayer.db.Get(bridgeRegistryAddrPrefix)
	if nil != err {
		return "", err
	}
	return string(addr), nil
}

func (ethRelayer *Relayer4Ethereum) updateTotalTxAmount2chain33(total int64) error {
	totalTx := &chain33Types.Int64{
		Data: atomic.LoadInt64(&ethRelayer.totalTx4Eth2Chain33),
	}
	//更新成功见证的交易数
	return ethRelayer.db.Set(chain33ToEthTxTotalAmount, chain33Types.Encode(totalTx))
}

func (ethRelayer *Relayer4Ethereum) setLastestStatics(claimType int32, txIndex int64, data []byte) error {
	key := calcRelayFromChain33Key(claimType, txIndex)
	return ethRelayer.db.Set(key, data)
}

func (ethRelayer *Relayer4Ethereum) queryTxhashes(prefix []byte) []string {
	return utils.QueryTxhashes(prefix, ethRelayer.db)
}

func (ethRelayer *Relayer4Ethereum) setHeight4BridgeBankLogAt(height uint64) error {
	return ethRelayer.setLogProcHeight(bridgeBankLogProcessedAt, height)
}

func (ethRelayer *Relayer4Ethereum) getHeight4BridgeBankLogAt() uint64 {
	return ethRelayer.getLogProcHeight(bridgeBankLogProcessedAt)
}

func (ethRelayer *Relayer4Ethereum) setLogProcHeight(key []byte, height uint64) error {
	data := &ebTypes.Uint64{
		Data: height,
	}
	return ethRelayer.db.Set(key, chain33Types.Encode(data))
}

func (ethRelayer *Relayer4Ethereum) getLogProcHeight(key []byte) uint64 {
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
func (ethRelayer *Relayer4Ethereum) setTxProcessed(txhash []byte) error {
	return ethRelayer.db.Set(txhash, []byte("1"))
}

//判断是否已经被处理，如果能够在数据库中找到该笔交易，则认为已经被处理
func (ethRelayer *Relayer4Ethereum) checkTxProcessed(txhash []byte) bool {
	_, err := ethRelayer.db.Get(txhash)
	return nil == err
}

func (ethRelayer *Relayer4Ethereum) setEthTxEvent(vLog types.Log) error {
	key := ethTxEventKey4Height(vLog.BlockNumber, uint32(vLog.TxIndex))
	value, err := json.Marshal(vLog)
	if nil != err {
		return err
	}
	return ethRelayer.db.Set(key, value)
}

func (ethRelayer *Relayer4Ethereum) getEthTxEvent(blockNumber uint64, txIndex uint32) (*types.Log, error) {
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

func (ethRelayer *Relayer4Ethereum) getNextValidEthTxEventLogs(height uint64, index uint32, fetchCnt int32) ([]*types.Log, error) {
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

func (ethRelayer *Relayer4Ethereum) setBridgeBankProcessedHeight(height uint64, index uint32) {
	bytes := chain33Types.Encode(&ebTypes.EventLogIndex{
		Height: height,
		Index:  index})
	_ = ethRelayer.db.Set(lastBridgeBankHeightProcPrefix, bytes)
}

func (ethRelayer *Relayer4Ethereum) getLastBridgeBankProcessedHeight() ebTypes.EventLogIndex {
	data, err := ethRelayer.db.Get(lastBridgeBankHeightProcPrefix)
	if nil != err {
		return ebTypes.EventLogIndex{}
	}
	logIndex := ebTypes.EventLogIndex{}
	_ = chain33Types.Decode(data, &logIndex)
	return logIndex
}

//构建一个引导查询使用的bridgeBankTx
func (ethRelayer *Relayer4Ethereum) initBridgeBankTx() {
	log, _ := ethRelayer.getEthTxEvent(0, 0)
	if nil != log {
		return
	}
	_ = ethRelayer.setEthTxEvent(types.Log{})
}

func (ethRelayer *Relayer4Ethereum) SetTokenAddress(token2set ebTypes.TokenAddress) error {
	addr := common.HexToAddress(token2set.Address)
	bytes := chain33Types.Encode(&token2set)
	ethRelayer.rwLock.Lock()
	ethRelayer.symbol2Addr[token2set.Symbol] = addr
	ethRelayer.rwLock.Unlock()
	return ethRelayer.db.Set(ethTokenSymbol2AddrKey(token2set.Symbol), bytes)
}

func (ethRelayer *Relayer4Ethereum) SetLockedTokenAddress(token2set ebTypes.TokenAddress) error {
	addr := common.HexToAddress(token2set.Address)
	bytes := chain33Types.Encode(&token2set)
	ethRelayer.rwLock.Lock()
	ethRelayer.symbol2LockAddr[token2set.Symbol] = addr
	ethRelayer.rwLock.Unlock()
	return ethRelayer.db.Set(ethTokenSymbol2LockAddrKey(token2set.Symbol), bytes)
}

func (ethRelayer *Relayer4Ethereum) RestoreTokenAddress() error {
	ethRelayer.rwLock.Lock()
	defer ethRelayer.rwLock.Unlock()

	ethRelayer.symbol2LockAddr[ebTypes.SYMBOL_ETH] = common.HexToAddress(ebTypes.EthNilAddr)
	helper := dbm.NewListHelper(ethRelayer.db)

	datas := helper.List(ethTokenSymbol2AddrPrefix, nil, 100, dbm.ListASC)
	for _, data := range datas {
		var token2set ebTypes.TokenAddress
		err := chain33Types.Decode(data, &token2set)
		if nil != err {
			return err
		}
		relayerLog.Info("RestoreTokenAddress", "symbol", token2set.Symbol, "address", token2set.Address)
		ethRelayer.symbol2Addr[token2set.Symbol] = common.HexToAddress(token2set.Address)
	}

	datas = helper.List(ethTokenSymbol2LockAddrPrefix, nil, 100, dbm.ListASC)
	for _, data := range datas {
		var token2set ebTypes.TokenAddress
		err := chain33Types.Decode(data, &token2set)
		if nil != err {
			return err
		}
		relayerLog.Info("RestoreTokenAddress", "symbol", token2set.Symbol, "address", token2set.Address)
		ethRelayer.symbol2LockAddr[token2set.Symbol] = common.HexToAddress(token2set.Address)
	}
	return nil
}

func (ethRelayer *Relayer4Ethereum) ShowTokenAddress(token2show ebTypes.TokenAddress) (*ebTypes.TokenAddressArray, error) {
	res := &ebTypes.TokenAddressArray{}

	if len(token2show.Symbol) > 0 {
		//data, err := ethRelayer.db.Get(ethTokenSymbol2AddrKey(token2show.Symbol))
		//if err != nil {
		addr, err := ethRelayer.ShowTokenAddrBySymbol(token2show.Symbol)
		if err != nil {
			return nil, err
		}
		var token2set ebTypes.TokenAddress
		token2set.Address = addr
		token2set.Symbol = token2show.Symbol
		token2set.ChainName = token2show.ChainName
		res.TokenAddress = append(res.TokenAddress, &token2set)
		return res, nil
		//}
		//var token2set ebTypes.TokenAddress
		//err = chain33Types.Decode(data, &token2set)
		//if nil != err {
		//	return nil, err
		//}
		//res.TokenAddress = append(res.TokenAddress, &token2set)
		//return res, nil
	}
	helper := dbm.NewListHelper(ethRelayer.db)
	datas := helper.List(ethTokenSymbol2AddrPrefix, nil, 100, dbm.ListASC)
	if nil == datas {
		return nil, errors.New("Not found")
	}

	for _, data := range datas {
		var token2set ebTypes.TokenAddress
		err := chain33Types.Decode(data, &token2set)
		if nil != err {
			return nil, err
		}
		res.TokenAddress = append(res.TokenAddress, &token2set)
	}
	return res, nil
}

func (ethRelayer *Relayer4Ethereum) ShowETHLockTokenAddress(token2show ebTypes.TokenAddress) (*ebTypes.TokenAddressArray, error) {
	res := &ebTypes.TokenAddressArray{}

	if len(token2show.Symbol) > 0 {
		data, err := ethRelayer.db.Get(ethTokenSymbol2LockAddrKey(token2show.Symbol))
		if err != nil {
			return nil, err
		}
		var token2set ebTypes.TokenAddress
		err = chain33Types.Decode(data, &token2set)
		if nil != err {
			return nil, err
		}
		res.TokenAddress = append(res.TokenAddress, &token2set)
		return res, nil
	}
	helper := dbm.NewListHelper(ethRelayer.db)
	datas := helper.List(ethTokenSymbol2LockAddrPrefix, nil, 100, dbm.ListASC)
	if nil == datas {
		return nil, errors.New("Not found")
	}

	for _, data := range datas {
		var token2set ebTypes.TokenAddress
		err := chain33Types.Decode(data, &token2set)
		if nil != err {
			return nil, err
		}
		res.TokenAddress = append(res.TokenAddress, &token2set)
	}
	return res, nil
}

func (ethRelayer *Relayer4Ethereum) setMultiSignAddress(address string) {
	bytes := []byte(address)
	_ = ethRelayer.db.Set(multiSignAddressPrefix, bytes)
}

func (ethRelayer *Relayer4Ethereum) getMultiSignAddress() string {
	bytes, _ := ethRelayer.db.Get(multiSignAddressPrefix)
	if 0 == len(bytes) {
		return ""
	}
	return string(bytes)
}
