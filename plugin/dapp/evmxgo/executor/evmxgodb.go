package executor

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

type evmxgoDB struct {
	evmxgo evmxgotypes.Evmxgo
}

func newEvmxgoDB(mint *evmxgotypes.EvmxgoMint) *evmxgoDB {
	e := &evmxgoDB{}
	e.evmxgo.Symbol = mint.GetSymbol()
	return e
}

func (e *evmxgoDB) save(db dbm.KV, key []byte) {
	set := e.getKVSet(key)
	for i := 0; i < len(set); i++ {
		err := db.Set(set[i].GetKey(), set[i].Value)
		if err != nil {
			panic(err)
		}
	}
}

func (e *evmxgoDB) getLogs(ty int32, status int32) []*types.ReceiptLog {
	var log []*types.ReceiptLog
	value := types.Encode(&evmxgotypes.ReceiptEvmxgo{Symbol: e.evmxgo.Symbol})
	log = append(log, &types.ReceiptLog{Ty: ty, Log: value})

	return log
}

//key:mavl-create-token-addr-xxx or mavl-token-xxx <-----> value:token
func (e *evmxgoDB) getKVSet(key []byte) (kvset []*types.KeyValue) {
	value := types.Encode(&e.evmxgo)
	kvset = append(kvset, &types.KeyValue{Key: key, Value: value})
	return kvset
}

func loadEvmxgoDB(db dbm.KV, symbol string) (*evmxgoDB, error) {
	evmxgo, err := db.Get(calcEvmxgoKey(symbol))
	if err != nil {
		elog.Error("evmxgodb load ", "Can't get token form db for token", symbol)
		return nil, evmxgotypes.ErrEvmxgoSymbolNotExist
	}
	var e evmxgotypes.Evmxgo
	err = types.Decode(evmxgo, &e)
	if err != nil {
		elog.Error("evmxgodb load", "Can't decode token info", symbol)
		return nil, err
	}
	return &evmxgoDB{e}, nil
}

func safeAdd(balance, amount int64) (int64, error) {
	if balance+amount < amount || balance+amount > types.MaxTokenBalance {
		return balance, types.ErrAmount
	}
	return balance + amount, nil
}

func (e *evmxgoDB) mint(amount int64) ([]*types.KeyValue, []*types.ReceiptLog, error) {
	newTotal, err := safeAdd(e.evmxgo.Total, amount)
	if err != nil {
		return nil, nil, err
	}

	prevEvmxgo := e.evmxgo
	e.evmxgo.Total = newTotal

	kvs := e.getKVSet(calcEvmxgoKey(e.evmxgo.Symbol))
	logs := []*types.ReceiptLog{{Ty: evmxgotypes.TyLogEvmxgoMint, Log: types.Encode(&evmxgotypes.ReceiptEvmxgoAmount{Prev: &prevEvmxgo, Current: &e.evmxgo})}}
	return kvs, logs, nil
}

func (e *evmxgoDB) mintMap(amount int64) ([]*types.KeyValue, []*types.ReceiptLog, error) {
	newTotal, err := safeAdd(e.evmxgo.Total, amount)
	if err != nil {
		return nil, nil, err
	}

	prevEvmxgo := e.evmxgo
	e.evmxgo.Total = newTotal

	kvs := e.getKVSet(calcEvmxgoKey(e.evmxgo.Symbol))
	logs := []*types.ReceiptLog{{Ty: evmxgotypes.TyLogEvmxgoMintMap, Log: types.Encode(&evmxgotypes.ReceiptEvmxgoAmount{Prev: &prevEvmxgo, Current: &e.evmxgo})}}
	return kvs, logs, nil
}

func (e *evmxgoDB) burn(db dbm.KV, amount int64) ([]*types.KeyValue, []*types.ReceiptLog, error) {
	if e.evmxgo.Total < amount {
		return nil, nil, types.ErrNoBalance
	}
	prevToken := e.evmxgo
	e.evmxgo.Total -= amount

	kvs := e.getKVSet(calcEvmxgoKey(e.evmxgo.Symbol))
	logs := []*types.ReceiptLog{{Ty: evmxgotypes.TyLogEvmxgoBurn, Log: types.Encode(&evmxgotypes.ReceiptEvmxgoAmount{Prev: &prevToken, Current: &e.evmxgo})}}
	return kvs, logs, nil
}

func (e *evmxgoDB) burnMap(db dbm.KV, amount int64) ([]*types.KeyValue, []*types.ReceiptLog, error) {
	if e.evmxgo.Total < amount {
		return nil, nil, types.ErrNoBalance
	}
	prevToken := e.evmxgo
	e.evmxgo.Total -= amount

	kvs := e.getKVSet(calcEvmxgoKey(e.evmxgo.Symbol))
	logs := []*types.ReceiptLog{{Ty: evmxgotypes.TyLogEvmxgoBurnMap, Log: types.Encode(&evmxgotypes.ReceiptEvmxgoAmount{Prev: &prevToken, Current: &e.evmxgo})}}
	return kvs, logs, nil
}

type evmxgoAction struct {
	stateDB   dbm.KV
	txhash    []byte
	fromaddr  string
	blocktime int64
	height    int64
	api       client.QueueProtocolAPI
}

func newEvmxgoAction(e *evmxgo, tx *types.Transaction) *evmxgoAction {
	return &evmxgoAction{e.GetStateDB(), tx.Hash(),
		tx.From(), e.GetBlockTime(), e.GetHeight(), e.GetAPI()}
}

func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		elog.Info("evmxgodb", "get stateDB key", "not found manageKey", "key", manageKey)
		return getConfigKey(key, db)
	}
	return value, nil
}

func getConfigKey(key string, db dbm.KV) ([]byte, error) {
	configKey := types.ConfigKey(key)
	value, err := db.Get([]byte(configKey))
	if err != nil {
		elog.Info("evmxgodb", "get db key", "not found configKey", "key", configKey)
		return nil, err
	}
	return value, nil
}

func hasConfiged(v1, key string, db dbm.KV) (bool, error) {
	value, err := getManageKey(key, db)
	if err != nil {
		elog.Info("evmxgodb", "get db key", "not found", "key", key)
		return false, err
	}
	if value == nil {
		elog.Info("evmxgodb", "get db key", "  found nil value", "key", key)
		return false, nil
	}

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		elog.Error("evmxgodb", "get db key", err)
		return false, err // types.ErrBadConfigValue
	}

	for _, v := range item.GetArr().Value {
		if v == v1 {
			return true, nil
		}
	}

	return false, nil
}

func loadEvmxgoMintConfig(db dbm.KV, symbol string) (*evmxgotypes.EvmxgoMintConfig, error) {
	key := fmt.Sprintf(mintPrefix+"%s", symbol)
	return loadEvmxgoConfig(db, key)
}

func loadEvmxgoMintMapConfig(db dbm.KV, symbol string) (*evmxgotypes.EvmxgoMintConfig, error) {
	key := fmt.Sprintf(mintMapPrefix+"%s", symbol)
	return loadEvmxgoConfig(db, key)
}

func loadEvmxgoConfig(db dbm.KV, key string) (*evmxgotypes.EvmxgoMintConfig, error) {
	value, err := getManageKey(key, db)
	if err != nil {
		elog.Info("evmxgodb", "get db key", "not found", "key", key)
		return nil, err
	}
	if value == nil {
		elog.Info("evmxgodb", "get db key", "  found nil value", "key", key)
		return nil, nil
	}
	elog.Info("loadEvmxgoMintConfig", "value", string(value))

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		elog.Error("evmxgodb load loadEvmxgoMintConfig", "Can't decode ConfigItem", key)
		return nil, err // types.ErrBadConfigValue
	}

	configValue := item.GetArr().Value
	if len(configValue) <= 0 {
		return nil, evmxgotypes.ErrEvmxgoSymbolNotConfigValue
	}

	var e evmxgotypes.EvmxgoMintConfig
	err = json.Unmarshal([]byte(configValue[0]), &e)

	if err != nil {
		elog.Error("evmxgodb load", "Can't decode token info", key)
		return nil, err
	}
	return &e, nil
}

func calcTokenAssetsKey(addr string) []byte {
	return []byte(fmt.Sprintf(evmxgoAssetsPrefix+"%s", addr))
}

func getTokenAssetsKey(addr string, db dbm.KVDB) (*types.ReplyStrings, error) {
	key := calcTokenAssetsKey(addr)
	value, err := db.Get(key)
	if err != nil && err != types.ErrNotFound {
		elog.Error("evmxgodb", "GetTokenAssetsKey", err)
		return nil, err
	}
	var assets types.ReplyStrings
	if err == types.ErrNotFound {
		return &assets, nil
	}
	err = types.Decode(value, &assets)
	if err != nil {
		elog.Error("evmxgodb", "GetTokenAssetsKey", err)
		return nil, err
	}
	return &assets, nil
}

// AddTokenToAssets 添加个人资产列表
func AddTokenToAssets(addr string, db dbm.KVDB, symbol string) []*types.KeyValue {
	tokenAssets, err := getTokenAssetsKey(addr, db)
	if err != nil {
		return nil
	}
	if tokenAssets == nil {
		tokenAssets = &types.ReplyStrings{}
	}

	var found = false
	for _, sym := range tokenAssets.Datas {
		if sym == symbol {
			found = true
			break
		}
	}
	if !found {
		tokenAssets.Datas = append(tokenAssets.Datas, symbol)
	}
	var kv []*types.KeyValue
	kv = append(kv, &types.KeyValue{Key: calcTokenAssetsKey(addr), Value: types.Encode(tokenAssets)})
	return kv
}

// 铸币不可控， 也是麻烦。 2选1
// 1. 谁可以发起
// 2. 是否需要审核  这个会增加管理的成本
// 现在实现选择 1
func (action *evmxgoAction) mint(mint *evmxgotypes.EvmxgoMint, tx2lock *types.Transaction) (*types.Receipt, error) {
	if mint == nil {
		return nil, types.ErrInvalidParam
	}
	if mint.GetAmount() < 0 || mint.GetAmount() > types.MaxTokenBalance || mint.GetSymbol() == "" {
		return nil, types.ErrInvalidParam
	}
	cfg := action.api.GetConfig()
	if err := checkMintPara(mint, tx2lock, action.stateDB); nil != err {
		return nil, err
	}

	// evmxgo合约，配置symbol对应的实际地址，检验地址正确才能发币
	configSymbol, err := loadEvmxgoMintConfig(action.stateDB, mint.GetSymbol())
	if err != nil || configSymbol == nil {
		elog.Error("evmxgo mint ", "not config symbol", mint.GetSymbol(), "error", err)
		return nil, evmxgotypes.ErrEvmxgoSymbolNotAllowedMint
	}

	if mint.BridgeToken != configSymbol.Address {
		elog.Error("evmxgo mint ", "NotCorrectBridgeTokenAddress with address by manager", configSymbol.Address, "mint.BridgeToken", mint.BridgeToken)
		return nil, evmxgotypes.ErrNotCorrectBridgeTokenAddress
	}

	evmxgodb, err := loadEvmxgoDB(action.stateDB, mint.GetSymbol())
	if err != nil {
		if err != evmxgotypes.ErrEvmxgoSymbolNotExist {
			return nil, err
		}

		evmxgodb = newEvmxgoDB(mint)
	}

	kvs, logs, err := evmxgodb.mint(mint.Amount)
	if err != nil {
		elog.Error("evmxgo mint ", "symbol", mint.GetSymbol(), "error", err, "from", action.fromaddr)
		return nil, err
	}
	evmxgoAccount, err := account.NewAccountDB(cfg, "evmxgo", mint.GetSymbol(), action.stateDB)
	if err != nil {
		return nil, err
	}
	elog.Debug("mint", "evmxgo.Symbol", mint.Symbol, "evmxgo.Amount", mint.Amount)
	receipt, err := evmxgoAccount.Mint(mint.Recipient, mint.Amount)
	if err != nil {
		return nil, err
	}

	logs = append(logs, receipt.Logs...)
	kvs = append(kvs, receipt.KV...)

	return &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}, nil
}

func (action *evmxgoAction) burn(burn *evmxgotypes.EvmxgoBurn) (*types.Receipt, error) {
	if burn == nil {
		return nil, types.ErrInvalidParam
	}
	if burn.GetAmount() < 0 || burn.GetAmount() > types.MaxTokenBalance || burn.GetSymbol() == "" {
		return nil, types.ErrInvalidParam
	}

	evmxgodb, err := loadEvmxgoDB(action.stateDB, burn.GetSymbol())
	if err != nil {
		return nil, err
	}

	kvs, logs, err := evmxgodb.burn(action.stateDB, burn.Amount)
	if err != nil {
		elog.Error("evmxgo burn ", "symbol", burn.GetSymbol(), "error", err, "from", action.fromaddr)
		return nil, err
	}
	chain33cfg := action.api.GetConfig()
	evmxgoAccount, err := account.NewAccountDB(chain33cfg, "evmxgo", burn.GetSymbol(), action.stateDB)
	if err != nil {
		return nil, err
	}
	elog.Debug("evmxgo burn", "burn.Symbol", burn.Symbol, "burn.Amount", burn.Amount)
	receipt, err := evmxgoAccount.Burn(action.fromaddr, burn.Amount)
	if err != nil {
		return nil, err
	}

	logs = append(logs, receipt.Logs...)
	kvs = append(kvs, receipt.KV...)

	return &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}, nil
}

func (action *evmxgoAction) mintMap(mint *evmxgotypes.EvmxgoMintMap, tx *types.Transaction) (*types.Receipt, error) {
	evmxgodb, err := loadEvmxgoDB(action.stateDB, mint.GetSymbol())
	if err != nil {
		if err != evmxgotypes.ErrEvmxgoSymbolNotExist {
			return nil, err
		}
		evmxgodb = newEvmxgoDB(&evmxgotypes.EvmxgoMint{
			Symbol:      mint.Symbol,
			Amount:      mint.Amount,
			BridgeToken: mint.BridgeToken,
			Recipient:   mint.Recipient,
			Extra:       mint.Extra,
		})
	}

	kvs, logs, err := evmxgodb.mintMap(mint.Amount)
	if err != nil {
		elog.Error("evmxgo mint ", "symbol", mint.GetSymbol(), "error", err, "from", action.fromaddr)
		return nil, err
	}
	cfg := action.api.GetConfig()
	evmxgoAccount, err := account.NewAccountDB(cfg, "evmxgo", mint.GetSymbol(), action.stateDB)
	if err != nil {
		return nil, err
	}
	elog.Debug("mint", "evmxgo.Symbol", mint.Symbol, "evmxgo.Amount", mint.Amount)
	receipt, err := evmxgoAccount.Mint(mint.Recipient, mint.Amount)
	if err != nil {
		return nil, err
	}

	logs = append(logs, receipt.Logs...)
	kvs = append(kvs, receipt.KV...)

	return &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}, nil
}
func (action *evmxgoAction) burnMap(burn *evmxgotypes.EvmxgoBurnMap) (*types.Receipt, error) {
	if burn == nil {
		return nil, types.ErrInvalidParam
	}
	if burn.GetAmount() < 0 || burn.GetAmount() > types.MaxTokenBalance || burn.GetSymbol() == "" {
		return nil, types.ErrInvalidParam
	}

	evmxgodb, err := loadEvmxgoDB(action.stateDB, burn.GetSymbol())
	if err != nil {
		return nil, err
	}

	kvs, logs, err := evmxgodb.burnMap(action.stateDB, burn.Amount)
	if err != nil {
		elog.Error("evmxgo burn ", "symbol", burn.GetSymbol(), "error", err, "from", action.fromaddr)
		return nil, err
	}
	chain33cfg := action.api.GetConfig()
	evmxgoAccount, err := account.NewAccountDB(chain33cfg, "evmxgo", burn.GetSymbol(), action.stateDB)
	if err != nil {
		return nil, err
	}
	elog.Debug("evmxgo burn", "burn.Symbol", burn.Symbol, "burn.Amount", burn.Amount)
	receipt, err := evmxgoAccount.Burn(action.fromaddr, burn.Amount)
	if err != nil {
		return nil, err
	}

	logs = append(logs, receipt.Logs...)
	kvs = append(kvs, receipt.KV...)

	return &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}, nil
}
