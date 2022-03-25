package chain33

import (
	"errors"
	"fmt"

	dbm "github.com/33cn/chain33/common/db"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
)

//key ...
var (
	lastSyncHeightPrefix               = []byte("chain33-lastSyncHeight:")
	eth2Chain33BurnLockTxStaticsPrefix = "chain33-eth2chain33BurnLockStatics"
	eth2Chain33BurnLockTxFinished      = "chain33-eth2Chain33BurnLockTxFinished"
	relayEthBurnLockTxTotalAmount      = []byte("chain33-relayEthBurnLockTxTotalAmount")
	chain33BurnTxUpdateTxIndex         = []byte("chain33-chain33BurnTxUpdateTxIndx")
	chain33LockTxUpdateTxIndex         = []byte("chain33-chain33LockTxUpdateTxIndex")
	bridgeRegistryAddrOnChain33        = []byte("chain33-x2EthBridgeRegistryAddrOnChain33")
	tokenSymbol2AddrPrefix             = []byte("chain33-chain33TokenSymbol2AddrPrefix")
	multiSignAddressPrefix             = []byte("chain33-multiSignAddress")
	symbol2Ethchain                    = []byte("chain33-symbol2Ethchain")
	txIsRelayedUnconfirm               = []byte("chain33-txIsRelayedUnconfirm")
	chain33TxRelayedAlready            = []byte("chain33-txRelayedAlready")
	fdTx2EthTotalAmount                = []byte("chain33-fdTx2EthTotalAmount")
	ethTxRelayAlreadyPrefix            = []byte("chain33-ethTxRelayAlready")
)

func ethTxRelayAlreadyKey(chain33Txhash string) []byte {
	return append(ethTxRelayAlreadyPrefix, []byte(fmt.Sprintf("-txHash-%s", chain33Txhash))...)
}

func chain33TxIsRelayedUnconfirmKey(txHash string) []byte {
	return append(txIsRelayedUnconfirm, []byte(fmt.Sprintf("-txHash-%s", txHash))...)
}

func chain33TxRelayedAlreadyKey(txHash string) []byte {
	return append(chain33TxRelayedAlready, []byte(fmt.Sprintf("-txHash-%s", txHash))...)
}

func tokenSymbol2AddrKey(symbol string) []byte {
	return append(tokenSymbol2AddrPrefix, []byte(fmt.Sprintf("-symbol-%s", symbol))...)
}

func calcRelayFromEthStaticsKey(txindex int64, claimType int32) []byte {
	return []byte(fmt.Sprintf("%s-%d-%012d", eth2Chain33BurnLockTxStaticsPrefix, claimType, txindex))
}

//未完成，处在pending状态
func calcRelayFromEthStaticsList(claimType int32) []byte {
	return []byte(fmt.Sprintf("%s-%d-", eth2Chain33BurnLockTxStaticsPrefix, claimType))
}

func calcFromEthFinishedStaticsKey(txindex int64, claimType int32) []byte {
	return []byte(fmt.Sprintf("%s-%d-%012d", eth2Chain33BurnLockTxFinished, claimType, txindex))
}

func calcFromEthFinishedStaticsList(claimType int32) []byte {
	return []byte(fmt.Sprintf("%s-%d-", eth2Chain33BurnLockTxFinished, claimType))
}

func (chain33Relayer *Relayer4Chain33) updateFdTx2EthTotalAmount(index int64) error {
	totalTx := &chain33Types.Int64{
		Data: index,
	}
	//更新成功见证的交易数
	return chain33Relayer.db.SetSync(fdTx2EthTotalAmount, chain33Types.Encode(totalTx))
}

func (chain33Relayer *Relayer4Chain33) getFdTx2EthTotalAmount() int64 {
	totalTx, _ := utils.LoadInt64FromDB(fdTx2EthTotalAmount, chain33Relayer.db)
	return totalTx
}

func (chain33Relayer *Relayer4Chain33) getAllTxsUnconfirm() (txInfos []*ebTypes.TxRelayConfirm4Chain33, err error) {
	helper := dbm.NewListHelper(chain33Relayer.db)
	datas := helper.List(txIsRelayedUnconfirm, nil, 0, dbm.ListASC)
	cnt := len(datas)
	if 0 == cnt {
		return nil, nil
	}

	txInfos = make([]*ebTypes.TxRelayConfirm4Chain33, cnt)
	for i, data := range datas {
		txInfo := &ebTypes.TxRelayConfirm4Chain33{}
		if err := chain33Types.Decode(data, txInfo); nil != err {
			return nil, err
		}

		txInfos[i] = txInfo
	}
	return
}

func (chain33Relayer *Relayer4Chain33) resetKeyChain33TxRelayedAlready(txHash string) error {
	key := chain33TxIsRelayedUnconfirmKey(txHash)
	data, err := chain33Relayer.db.Get(key)
	if nil != err {
		relayerLog.Info("resetKeyTxRelayedAlready", "No data for tx", txHash)
		return err
	}
	_ = chain33Relayer.db.DeleteSync(key)
	setkey := chain33TxRelayedAlreadyKey(txHash)

	return chain33Relayer.db.SetSync(setkey, data)
}

func (chain33Relayer *Relayer4Chain33) setChain33TxIsRelayedUnconfirm(txHash string, index int64, txRelayConfirm4Chain33 *ebTypes.TxRelayConfirm4Chain33) error {
	key := chain33TxIsRelayedUnconfirmKey(txHash)
	data := chain33Types.Encode(txRelayConfirm4Chain33)
	relayerLog.Info("setChain33TxIsRelayedUnconfirm", "TxHash", txHash, "index", index, "ForwardTimes", txRelayConfirm4Chain33.FdTimes)
	return chain33Relayer.db.SetSync(key, data)
}

func (chain33Relayer *Relayer4Chain33) setEthTxRelayAlreadyInfo(ethTxhash string, relayTxDetail *ebTypes.RelayTxDetail) error {
	key := ethTxRelayAlreadyKey(ethTxhash)
	data := chain33Types.Encode(relayTxDetail)
	return chain33Relayer.db.SetSync(key, data)
}

func (chain33Relayer *Relayer4Chain33) getEthTxRelayAlreadyInfo(ethTxhash string) (*ebTypes.RelayTxDetail, error) {
	key := ethTxRelayAlreadyKey(ethTxhash)
	data, err := chain33Relayer.db.Get(key)
	if nil != err {
		return nil, err
	}
	var relayTxDetail ebTypes.RelayTxDetail
	err = chain33Types.Decode(data, &relayTxDetail)
	return &relayTxDetail, err
}

func (chain33Relayer *Relayer4Chain33) updateTotalTxAmount2Eth(txIndex int64) error {
	totalTx := &chain33Types.Int64{
		Data: txIndex,
	}
	//更新成功见证的交易数
	return chain33Relayer.db.SetSync(relayEthBurnLockTxTotalAmount, chain33Types.Encode(totalTx))
}

func (chain33Relayer *Relayer4Chain33) getTotalTxAmount() int64 {
	totalTx, _ := utils.LoadInt64FromDB(relayEthBurnLockTxTotalAmount, chain33Relayer.db)
	return totalTx
}

func (chain33Relayer *Relayer4Chain33) setLastestRelay2Chain33TxStatics(txIndex int64, claimType int32, data []byte) error {
	key := calcRelayFromEthStaticsKey(txIndex, claimType)
	return chain33Relayer.db.SetSync(key, data)
}

func (chain33Relayer *Relayer4Chain33) getStatics(claimType int32, txIndex int64, count int32) ([][]byte, error) {
	//第一步：获取处在pending状态的
	keyPrefix := calcRelayFromEthStaticsList(claimType)
	keyFrom := calcRelayFromEthStaticsKey(txIndex, claimType)
	helper := dbm.NewListHelper(chain33Relayer.db)
	datas := helper.List(keyPrefix, keyFrom, count, dbm.ListASC)
	if nil == datas {
		return nil, errors.New("Not found")
	}

	return datas, nil
}

func (chain33Relayer *Relayer4Chain33) setChain33UpdateTxIndex(txindex int64, claimType events.ClaimType) error {
	txIndexWrapper := &chain33Types.Int64{
		Data: txindex,
	}

	if events.ClaimTypeBurn == claimType {
		return chain33Relayer.db.SetSync(chain33BurnTxUpdateTxIndex, chain33Types.Encode(txIndexWrapper))
	}
	return chain33Relayer.db.SetSync(chain33LockTxUpdateTxIndex, chain33Types.Encode(txIndexWrapper))
}

func (chain33Relayer *Relayer4Chain33) getChain33UpdateTxIndex(claimType events.ClaimType) int64 {
	var key []byte
	if events.ClaimTypeBurn == claimType {
		key = chain33BurnTxUpdateTxIndex
	} else {
		key = chain33LockTxUpdateTxIndex
	}
	data, err := chain33Relayer.db.Get(key)
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

//获取上次同步到app的高度
func (chain33Relayer *Relayer4Chain33) loadLastSyncHeight() int64 {
	height, err := utils.LoadInt64FromDB(lastSyncHeightPrefix, chain33Relayer.db)
	if nil != err && err != chain33Types.ErrHeightNotExist {
		relayerLog.Error("loadLastSyncHeight", "err:", err.Error())
		return 0
	}
	return height
}

func (chain33Relayer *Relayer4Chain33) setLastSyncHeight(syncHeight int64) {
	bytes := chain33Types.Encode(&chain33Types.Int64{Data: syncHeight})
	_ = chain33Relayer.db.SetSync(lastSyncHeightPrefix, bytes)
}

func (chain33Relayer *Relayer4Chain33) setBridgeRegistryAddr(bridgeRegistryAddr string) error {
	return chain33Relayer.db.SetSync(bridgeRegistryAddrOnChain33, []byte(bridgeRegistryAddr))
}

func (chain33Relayer *Relayer4Chain33) getBridgeRegistryAddr() (string, error) {
	addr, err := chain33Relayer.db.Get(bridgeRegistryAddrOnChain33)
	if nil != err {
		return "", err
	}
	return string(addr), nil
}

func (chain33Relayer *Relayer4Chain33) SetTokenAddress(token2set *ebTypes.TokenAddress) error {
	bytes := chain33Types.Encode(token2set)
	chain33Relayer.rwLock.Lock()
	chain33Relayer.symbol2Addr[token2set.Symbol] = token2set.Address
	chain33Relayer.rwLock.Unlock()
	return chain33Relayer.db.SetSync(tokenSymbol2AddrKey(token2set.Symbol), bytes)
}

func (chain33Relayer *Relayer4Chain33) RestoreTokenAddress() error {
	chain33Relayer.rwLock.Lock()
	defer chain33Relayer.rwLock.Unlock()
	chain33Relayer.symbol2Addr[ebTypes.SYMBOL_BTY] = ebTypes.BTYAddrChain33

	helper := dbm.NewListHelper(chain33Relayer.db)
	datas := helper.List(tokenSymbol2AddrPrefix, nil, 100, dbm.ListASC)
	if nil == datas {
		return nil
	}

	for _, data := range datas {
		var token2set ebTypes.TokenAddress
		err := chain33Types.Decode(data, &token2set)
		if nil != err {
			return err
		}
		relayerLog.Info("RestoreTokenAddress", "symbol", token2set.Symbol, "address", token2set.Address)
		chain33Relayer.symbol2Addr[token2set.Symbol] = token2set.Address
	}
	return nil
}

func (chain33Relayer *Relayer4Chain33) ShowTokenAddress(token2show *ebTypes.TokenAddress) (*ebTypes.TokenAddressArray, error) {
	res := &ebTypes.TokenAddressArray{}

	if len(token2show.Symbol) > 0 {
		data, err := chain33Relayer.db.Get(tokenSymbol2AddrKey(token2show.Symbol))
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
	helper := dbm.NewListHelper(chain33Relayer.db)
	datas := helper.List(tokenSymbol2AddrPrefix, nil, 100, dbm.ListASC)
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

func (chain33Relayer *Relayer4Chain33) setMultiSignAddress(address string) {
	bytes := []byte(address)
	_ = chain33Relayer.db.SetSync(multiSignAddressPrefix, bytes)
}

func (chain33Relayer *Relayer4Chain33) getMultiSignAddress() string {
	bytes, _ := chain33Relayer.db.Get(multiSignAddressPrefix)
	if 0 == len(bytes) {
		return ""
	}
	return string(bytes)
}

func (chain33Relayer *Relayer4Chain33) storeSymbol2chainName(symbol2Name map[string]string) {
	Symbol2EthChain := &ebTypes.Symbol2EthChain{
		Symbol2Name: symbol2Name,
	}
	data := chain33Types.Encode(Symbol2EthChain)
	_ = chain33Relayer.db.SetSync(symbol2Ethchain, data)
}

func (chain33Relayer *Relayer4Chain33) restoreSymbol2chainName() map[string]string {
	data, _ := chain33Relayer.db.Get(symbol2Ethchain)
	if 0 == len(data) {
		return make(map[string]string)
	}

	symbol2EthChain := &ebTypes.Symbol2EthChain{}
	if err := chain33Types.Decode(data, symbol2EthChain); nil != err {
		return make(map[string]string)
	}
	return symbol2EthChain.Symbol2Name
}

//判断是否已经被处理，如果能够在数据库中找到该笔交易，则认为已经被处理
func (chain33Relayer *Relayer4Chain33) checkTxProcessed(txhash string) bool {
	key1 := chain33TxIsRelayedUnconfirmKey(txhash)
	data, err := chain33Relayer.db.Get(key1)
	if 0 != len(data) && nil == err {
		return true
	}

	key2 := chain33TxRelayedAlreadyKey(txhash)
	data, err = chain33Relayer.db.Get(key2)
	if 0 != len(data) && nil == err {
		return true
	}

	return false
}
