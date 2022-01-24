package ethereum

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethtxs"

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
	chain33ToEthTxTotalAmount      = "eth-chain33ToEthTxTotalAmount"
	bridgeRegistryAddrPrefix       = "eth-x2EthBridgeRegistryAddr"
	bridgeBankLogProcessedAt       = "eth-bridgeBankLogProcessedAt"
	ethTxEventPrefix               = "eth-ethTxEventPrefix"
	lastBridgeBankHeightProcPrefix = "eth-lastBridgeBankHeight"
	ethTokenSymbol2AddrPrefix      = "eth-ethTokenSymbol2AddrPrefix"
	ethTokenSymbol2LockAddrPrefix  = "eth-ethTokenSymbol2LockAddrPrefix"
	ethLockTxUpdateTxIndex         = "eth-ethLockTxUpdateTxIndex"
	ethBurnTxUpdateTxIndex         = "eth-ethBurnTxUpdateTxIndex"
	multiSignAddressPrefix         = "eth-multiSignAddress"
	withdrawParaKey                = "eth-withdrawPara"
	withdrawTokenPrefix            = "eth-withdrawToken"
	withdrawTokenListPrefix        = "eth-withdrawTokenList"
)

func ethTokenSymbol2AddrKey(chainName, symbol string) []byte {
	return []byte(fmt.Sprintf("%s-%s-symbol-%s", chainName, ethTokenSymbol2AddrPrefix, symbol))
}

func ethTokenSymbol2LockAddrKey(chainName, symbol string) []byte {
	return []byte(fmt.Sprintf("%s-%s-symbol-%s", chainName, ethTokenSymbol2LockAddrPrefix, symbol))
}

func ethTxEventKey4Height(chainName string, height uint64, index uint32) []byte {
	return []byte(fmt.Sprintf("%s-%s-%020d-%d", chainName, ethTxEventPrefix, height, index))
}

func calcRelayFromChain33Key(chainName string, claimType int32, txindex int64) []byte {
	return []byte(fmt.Sprintf("%s-%s-%d-%012d", chainName, chain33ToEthStaticsPrefix, claimType, txindex))
}

func calcRelayFromChain33ListPrefix(chainName string, claimType int32) []byte {
	return []byte(fmt.Sprintf("%s-%s-%d-", chainName, chain33ToEthStaticsPrefix, claimType))
}

func (ethRelayer *Relayer4Ethereum) getStatics(claimType int32, txIndex int64, count int32) ([][]byte, error) {
	keyPrefix := calcRelayFromChain33ListPrefix(ethRelayer.name, claimType)

	keyFrom := calcRelayFromChain33Key(ethRelayer.name, claimType, txIndex)
	helper := dbm.NewListHelper(ethRelayer.db)
	datas := helper.List(keyPrefix, keyFrom, count, dbm.ListASC)
	if nil == datas {
		return nil, errors.New("not found")
	}

	return datas, nil
}

func (ethRelayer *Relayer4Ethereum) setEthLockTxUpdateTxIndex(txindex int64, claimType events.ClaimType) error {
	txIndexWrapper := &chain33Types.Int64{
		Data: txindex,
	}

	if events.ClaimTypeBurn == claimType {
		key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, ethBurnTxUpdateTxIndex))
		return ethRelayer.db.Set(key, chain33Types.Encode(txIndexWrapper))
	}
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, ethLockTxUpdateTxIndex))
	return ethRelayer.db.Set(key, chain33Types.Encode(txIndexWrapper))
}

func (ethRelayer *Relayer4Ethereum) getEthLockTxUpdateTxIndex(claimType events.ClaimType) int64 {
	var key []byte
	if events.ClaimTypeBurn == claimType {
		key = []byte(fmt.Sprintf("%s-%s", ethRelayer.name, ethBurnTxUpdateTxIndex))
	} else {
		key = []byte(fmt.Sprintf("%s-%s", ethRelayer.name, ethLockTxUpdateTxIndex))
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
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, bridgeRegistryAddrPrefix))
	return ethRelayer.db.Set(key, []byte(bridgeRegistryAddr))
}

func (ethRelayer *Relayer4Ethereum) getBridgeRegistryAddr() (string, error) {
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, bridgeRegistryAddrPrefix))
	addr, err := ethRelayer.db.Get(key)
	if nil != err {
		return "", err
	}
	return string(addr), nil
}

func (ethRelayer *Relayer4Ethereum) updateTotalTxAmount2chain33(totalIndex int64) error {
	totalTx := &chain33Types.Int64{
		Data: totalIndex,
	}
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, chain33ToEthTxTotalAmount))
	//更新成功见证的交易数
	return ethRelayer.db.Set(key, chain33Types.Encode(totalTx))
}

func (ethRelayer *Relayer4Ethereum) getTotalTxAmount2Eth() int64 {
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, chain33ToEthTxTotalAmount))
	totalTx, _ := utils.LoadInt64FromDB(key, ethRelayer.db)
	return totalTx
}

func (ethRelayer *Relayer4Ethereum) setLastestStatics(claimType int32, txIndex int64, data []byte) error {
	key := calcRelayFromChain33Key(ethRelayer.name, claimType, txIndex)
	return ethRelayer.db.Set(key, data)
}

func (ethRelayer *Relayer4Ethereum) setHeight4BridgeBankLogAt(height uint64) error {
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, bridgeBankLogProcessedAt))
	return ethRelayer.setLogProcHeight(key, height)
}

func (ethRelayer *Relayer4Ethereum) getHeight4BridgeBankLogAt() uint64 {
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, bridgeBankLogProcessedAt))
	return ethRelayer.getLogProcHeight(key)
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
	key := ethTxEventKey4Height(ethRelayer.name, vLog.BlockNumber, uint32(vLog.TxIndex))
	value, err := json.Marshal(vLog)
	if nil != err {
		return err
	}
	return ethRelayer.db.Set(key, value)
}

func (ethRelayer *Relayer4Ethereum) getEthTxEvent(blockNumber uint64, txIndex uint32) (*types.Log, error) {
	key := ethTxEventKey4Height(ethRelayer.name, blockNumber, txIndex)
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
	key := ethTxEventKey4Height(ethRelayer.name, height, index)
	helper := dbm.NewListHelper(ethRelayer.db)
	prefix := []byte(fmt.Sprintf("%s-%s-", ethRelayer.name, ethTxEventPrefix))
	datas := helper.List(prefix, key, fetchCnt, dbm.ListASC)
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
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, lastBridgeBankHeightProcPrefix))
	_ = ethRelayer.db.Set(key, bytes)
}

func (ethRelayer *Relayer4Ethereum) getLastBridgeBankProcessedHeight() ebTypes.EventLogIndex {
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, lastBridgeBankHeightProcPrefix))
	data, err := ethRelayer.db.Get(key)
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

func (ethRelayer *Relayer4Ethereum) SetTokenAddress(token2set *ebTypes.TokenAddress) error {
	addr := common.HexToAddress(token2set.Address)
	bytes := chain33Types.Encode(token2set)
	ethRelayer.rwLock.Lock()
	ethRelayer.symbol2Addr[token2set.Symbol] = addr
	ethRelayer.rwLock.Unlock()
	return ethRelayer.db.Set(ethTokenSymbol2AddrKey(ethRelayer.name, token2set.Symbol), bytes)
}

func (ethRelayer *Relayer4Ethereum) SetLockedTokenAddress(token2set ebTypes.TokenAddress) error {
	bytes := chain33Types.Encode(&token2set)
	ethRelayer.rwLock.Lock()
	ethRelayer.symbol2LockAddr[token2set.Symbol] = token2set
	ethRelayer.rwLock.Unlock()
	return ethRelayer.db.Set(ethTokenSymbol2LockAddrKey(ethRelayer.name, token2set.Symbol), bytes)
}

func (ethRelayer *Relayer4Ethereum) GetLockedTokenAddress(symbol string) (*ebTypes.TokenAddress, error) {
	ethRelayer.rwLock.RLock()
	data, err := ethRelayer.db.Get(ethTokenSymbol2LockAddrKey(ethRelayer.name, symbol))
	ethRelayer.rwLock.RUnlock()
	if nil != err {
		return nil, err
	}
	var token2set ebTypes.TokenAddress
	if err := chain33Types.Decode(data, &token2set); nil != err {
		return nil, err
	}
	return &token2set, err
}

func (ethRelayer *Relayer4Ethereum) RestoreTokenAddress() error {
	ethRelayer.rwLock.Lock()
	defer ethRelayer.rwLock.Unlock()

	helper := dbm.NewListHelper(ethRelayer.db)

	prefix := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, ethTokenSymbol2AddrPrefix))
	datas := helper.List(prefix, nil, 100, dbm.ListASC)
	for _, data := range datas {
		var token2set ebTypes.TokenAddress
		err := chain33Types.Decode(data, &token2set)
		if nil != err {
			return err
		}
		relayerLog.Info("RestoreTokenAddress", "symbol", token2set.Symbol, "address", token2set.Address)
		ethRelayer.symbol2Addr[token2set.Symbol] = common.HexToAddress(token2set.Address)
	}

	prefix = []byte(fmt.Sprintf("%s-%s", ethRelayer.name, ethTokenSymbol2LockAddrPrefix))
	datas = helper.List(prefix, nil, 100, dbm.ListASC)
	for _, data := range datas {
		var tokenLocked ebTypes.TokenAddress
		err := chain33Types.Decode(data, &tokenLocked)
		if nil != err {
			return err
		}
		relayerLog.Info("RestoreTokenAddress", "symbol", tokenLocked.Symbol, "address", tokenLocked.Address)
		ethRelayer.symbol2LockAddr[tokenLocked.Symbol] = tokenLocked
	}
	return nil
}

func (ethRelayer *Relayer4Ethereum) ShowTokenAddress(token2show *ebTypes.TokenAddress) (*ebTypes.TokenAddressArray, error) {
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
	prefix := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, ethTokenSymbol2AddrPrefix))
	datas := helper.List(prefix, nil, 100, dbm.ListASC)
	if nil == datas {
		return nil, errors.New("not found")
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

func (ethRelayer *Relayer4Ethereum) ShowETHLockTokenAddress(token2show *ebTypes.TokenAddress) (*ebTypes.TokenAddressArray, error) {
	res := &ebTypes.TokenAddressArray{}

	if len(token2show.Symbol) > 0 {
		data, err := ethRelayer.db.Get(ethTokenSymbol2LockAddrKey(ethRelayer.name, token2show.Symbol))
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
	prefix := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, ethTokenSymbol2LockAddrPrefix))
	datas := helper.List(prefix, nil, 100, dbm.ListASC)
	if nil == datas {
		return nil, errors.New("not found")
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
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, multiSignAddressPrefix))
	_ = ethRelayer.db.Set(key, bytes)
}

func (ethRelayer *Relayer4Ethereum) getMultiSignAddress() string {
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, multiSignAddressPrefix))
	bytes, _ := ethRelayer.db.Get(key)
	if 0 == len(bytes) {
		return ""
	}
	return string(bytes)
}

func (ethRelayer *Relayer4Ethereum) setWithdrawFee(symbol2Para map[string]*ebTypes.WithdrawPara) error {
	withdrawSymbol2Fee := &ebTypes.WithdrawSymbol2Para{
		Symbol2Para: symbol2Para,
	}

	bytes := chain33Types.Encode(withdrawSymbol2Fee)
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, withdrawParaKey))
	return ethRelayer.db.Set(key, bytes)
}

func (ethRelayer *Relayer4Ethereum) restoreWithdrawFee() map[string]*ebTypes.WithdrawPara {
	key := []byte(fmt.Sprintf("%s-%s", ethRelayer.name, withdrawParaKey))
	bytes, _ := ethRelayer.db.Get(key)
	if 0 == len(bytes) {
		result := make(map[string]*ebTypes.WithdrawPara)
		return result
	}

	var withdrawSymbol2Para ebTypes.WithdrawSymbol2Para
	if err := chain33Types.Decode(bytes, &withdrawSymbol2Para); nil != err {
		result := make(map[string]*ebTypes.WithdrawPara)
		return result
	}

	return withdrawSymbol2Para.Symbol2Para
}

func (ethRelayer *Relayer4Ethereum) restoreWithdrawFeeInINt() map[string]*WithdrawFeeAndQuota {
	withdrawPara := ethRelayer.restoreWithdrawFee()
	res := make(map[string]*WithdrawFeeAndQuota)
	for symbol, para := range withdrawPara {
		feeInt, _ := big.NewInt(0).SetString(para.Fee, 10)
		amountPerDayInt, _ := big.NewInt(0).SetString(para.AmountPerDay, 10)
		res[symbol] = &WithdrawFeeAndQuota{
			Fee:          feeInt,
			AmountPerDay: amountPerDayInt,
		}
	}
	return res
}

func calcWithdrawKey(name, chain33Sender, symbol string, year, month, day int, nonce int64) []byte {
	return []byte(fmt.Sprintf("%s-%s-%s-%s-%d-%d-%d-%d", name, withdrawTokenPrefix, chain33Sender, symbol, year, month, day, nonce))
}

func calcWithdrawKeyPrefix(name, chain33Sender, symbol string, year, month, day int) []byte {
	return []byte(fmt.Sprintf("%s-%s-%s-%s-%d-%d-%d", name, withdrawTokenPrefix, chain33Sender, symbol, year, month, day))
}

func calcWithdrawListKey(name string, nonce int64) []byte {
	return []byte(fmt.Sprintf("%s-%s-%d", name, withdrawTokenListPrefix, nonce))
}

func (ethRelayer *Relayer4Ethereum) setWithdraw(withdrawTx *ebTypes.WithdrawTx) error {
	chain33Sender := withdrawTx.Chain33Sender
	symbol := withdrawTx.Symbol
	year := withdrawTx.Year
	month := withdrawTx.Month
	day := withdrawTx.Day

	key := calcWithdrawKey(ethRelayer.name, chain33Sender, symbol, int(year), int(month), int(day), withdrawTx.Nonce)
	bytes := chain33Types.Encode(withdrawTx)

	if err := ethRelayer.db.Set(key, bytes); nil != err {
		return err
	}

	//保存按照次序提币的交易，方便查询
	listKey := calcWithdrawListKey(ethRelayer.name, withdrawTx.Nonce)
	listData := key

	return ethRelayer.db.Set(listKey, listData)
}

func (ethRelayer *Relayer4Ethereum) setWithdrawStatics(withdrawTx *ebTypes.WithdrawTx, chain33Msg *events.Chain33Msg) error {
	txIndex := atomic.AddInt64(&ethRelayer.totalTxRelayFromChain33, 1)
	operationType := chain33Msg.ClaimType.String()
	statics := &ebTypes.Chain33ToEthereumStatics{
		EthTxstatus:      ebTypes.Tx_Status_Pending,
		Chain33Txhash:    common.Bytes2Hex(chain33Msg.TxHash),
		EthereumTxhash:   withdrawTx.TxHashOnEthereum,
		BurnLockWithdraw: int32(chain33Msg.ClaimType),
		Chain33Sender:    chain33Msg.Chain33Sender.String(),
		EthereumReceiver: chain33Msg.EthereumReceiver.String(),
		Symbol:           chain33Msg.Symbol,
		Amount:           chain33Msg.Amount.String(),
		Nonce:            chain33Msg.Nonce,
		TxIndex:          txIndex,
		OperationType:    operationType,
	}
	if withdrawTx.Status == int32(ethtxs.WDError) {
		statics.EthTxstatus = ebTypes.Tx_Status_Failed
	}
	relayerLog.Info("setWithdrawStatics::successful", "txIndex", txIndex, "Chain33Txhash", statics.Chain33Txhash, "EthereumTxhash", statics.EthereumTxhash, "type", operationType,
		"Symbol", chain33Msg.Symbol, "Amount", chain33Msg.Amount, "EthereumReceiver", statics.EthereumReceiver, "Chain33Sender", statics.Chain33Sender)

	data := chain33Types.Encode(statics)
	return ethRelayer.setLastestStatics(int32(chain33Msg.ClaimType), txIndex, data)
}

func (ethRelayer *Relayer4Ethereum) getWithdrawsWithinSameDay(withdrawTx *ebTypes.WithdrawTx) (*big.Int, error) {
	chain33Sender := withdrawTx.Chain33Sender
	symbol := withdrawTx.Symbol
	year := withdrawTx.Year
	month := withdrawTx.Month
	day := withdrawTx.Day

	prefix := calcWithdrawKeyPrefix(ethRelayer.name, chain33Sender, symbol, int(year), int(month), int(day))
	helper := dbm.NewListHelper(ethRelayer.db)
	datas := helper.List(prefix, nil, 100, dbm.ListASC)
	if nil == datas {
		return big.NewInt(0), nil
	}

	withdrawTotal := big.NewInt(0)
	for _, data := range datas {
		var info ebTypes.WithdrawTx
		err := chain33Types.Decode(data, &info)
		if nil != err {
			return big.NewInt(0), err
		}
		AmountInt, _ := big.NewInt(0).SetString(info.Amount, 0)
		withdrawTotal.Add(withdrawTotal, AmountInt)
	}
	return withdrawTotal, nil
}
