package spot

import (
	"encoding/json"
	"reflect"

	dbm "github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func LoadSpotFeeAccountConfig(db dbm.KV) (*et.DexAccount, error) {
	key := string(spotFeeAccountKey)
	value, err := getManageKey(key, db)
	if err != nil {
		elog.Info("loadSpotFeeAccountConfig", "get db key", "not found", "key", key)
		return nil, err
	}
	if value == nil {
		elog.Info("loadSpotFeeAccountConfig", "get db key", "  found nil value", "key", key)
		return nil, nil
	}
	elog.Info("loadSpotFeeAccountConfig", "value", string(value))

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		elog.Error("loadSpotFeeAccountConfig", "Can't decode ConfigItem due to", err.Error())
		return nil, err // types.ErrBadConfigValue
	}

	configValue := item.GetArr().Value
	if len(configValue) <= 0 {
		return nil, et.ErrSpotFeeConfig
	}

	var e et.DexAccount
	err = json.Unmarshal([]byte(configValue[0]), &e)

	if err != nil {
		elog.Error("loadSpotFeeAccountConfig load", "Can't decode token info due to:", err.Error())
		return nil, err
	}
	return &e, nil
}

func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		elog.Info("getManageKey", "get db key", "not found manageKey", "key", manageKey)
		return getConfigKey(key, db)
	}
	return value, nil
}

func getConfigKey(key string, db dbm.KV) ([]byte, error) {
	configKey := types.ConfigKey(key)
	value, err := db.Get([]byte(configKey))
	if err != nil {
		elog.Info("getManageKey", "get db key", "not found configKey", "key", configKey)
		return nil, err
	}
	return value, nil
}

// config

func ParseConfig(cfg *types.Chain33Config, height int64) (*et.Econfig, error) {
	banks, err := ParseStrings(cfg, "banks", height)
	if err != nil || len(banks) == 0 {
		return nil, err
	}
	coins, err := ParseCoins(cfg, "coins", height)
	if err != nil {
		return nil, err
	}
	exchanges, err := ParseSymbols(cfg, "exchanges", height)
	if err != nil {
		return nil, err
	}
	return &et.Econfig{
		Banks:     banks,
		Coins:     coins,
		Exchanges: exchanges,
	}, nil
}

func ParseStrings(cfg *types.Chain33Config, tradeKey string, height int64) (ret []string, err error) {
	val, err := cfg.MG(et.MverPrefix+"."+tradeKey, height)
	if err != nil {
		return nil, err
	}

	datas, ok := val.([]interface{})
	if !ok {
		elog.Error("invalid val", "val", val, "key", tradeKey)
		return nil, et.ErrCfgFmt
	}

	for _, v := range datas {
		one, ok := v.(string)
		if !ok {
			elog.Error("invalid one", "one", one, "key", tradeKey)
			return nil, et.ErrCfgFmt
		}
		ret = append(ret, one)
	}
	return
}

func ParseCoins(cfg *types.Chain33Config, tradeKey string, height int64) (coins []et.CoinCfg, err error) {
	coins = make([]et.CoinCfg, 0)

	val, err := cfg.MG(et.MverPrefix+"."+tradeKey, height)
	if err != nil {
		return nil, err
	}

	datas, ok := val.([]interface{})
	if !ok {
		elog.Error("invalid coins", "val", val, "type", reflect.TypeOf(val))
		return nil, et.ErrCfgFmt
	}

	for _, e := range datas {
		v, ok := e.(map[string]interface{})
		if !ok {
			elog.Error("invalid coins one", "one", v, "key", tradeKey)
			return nil, et.ErrCfgFmt
		}

		coin := et.CoinCfg{
			Coin:   v["coin"].(string),
			Execer: v["execer"].(string),
			Name:   v["name"].(string),
		}
		coins = append(coins, coin)
	}
	return
}

func ParseSymbols(cfg *types.Chain33Config, tradeKey string, height int64) (symbols map[string]*et.Trade, err error) {
	symbols = make(map[string]*et.Trade)

	val, err := cfg.MG(et.MverPrefix+"."+tradeKey, height)
	if err != nil {
		return nil, err
	}

	datas, ok := val.([]interface{})
	if !ok {
		elog.Error("invalid Symbols", "val", val, "type", reflect.TypeOf(val))
		return nil, et.ErrCfgFmt
	}

	for _, e := range datas {
		v, ok := e.(map[string]interface{})
		if !ok {
			elog.Error("invalid Symbols one", "one", v, "key", tradeKey)
			return nil, et.ErrCfgFmt
		}

		symbol := v["symbol"].(string)
		symbols[symbol] = &et.Trade{
			Symbol:       symbol,
			PriceDigits:  int32(formatInterface(v["priceDigits"])),
			AmountDigits: int32(formatInterface(v["amountDigits"])),
			Taker:        int32(formatInterface(v["taker"])),
			Maker:        int32(formatInterface(v["maker"])),
			MinFee:       formatInterface(v["minFee"]),
		}
	}
	return
}

func formatInterface(data interface{}) int64 {
	switch data := data.(type) {
	case int64:
		return data
	case int32:
		return int64(data)
	case int:
		return int64(data)
	default:
		return 0
	}
}
