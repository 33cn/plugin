package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

/*
 * 执行器相关定义
 * 重载基类相关接口
 */

const (
	mintPrefix            = "evmxgo-mint-"
	mintMapPrefix         = "evmxgo-mint-map-"
	evmxgoAssetsPrefix    = "LODB-evmxgo-assets:"
	bridgevmxgoAddrPrefix = "bridgevmxgo-contract-addr"
)

var (
	//日志
	elog = log.New("module", "evmxgo.executor")
)

var driverName = evmxgotypes.EvmxgoX

type subConfig struct {
	SaveTokenTxList bool `json:"saveTokenTxList"`
}

var subCfg subConfig

// Init register dapp
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	if sub != nil {
		types.MustDecode(sub, &subCfg)
	}
	drivers.Register(cfg, GetName(), newEvmxgo, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// InitExecType Init Exec Type
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&evmxgo{}))
}

type evmxgo struct {
	drivers.DriverBase
}

func newEvmxgo() drivers.Driver {
	t := &evmxgo{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName get driver name
func GetName() string {
	return newEvmxgo().GetName()
}

func (e *evmxgo) GetDriverName() string {
	return driverName
}

// CheckTx 实现自定义检验交易接口，供框架调用
func (e *evmxgo) CheckTx(tx *types.Transaction, index int) error {
	// implement code
	return nil
}

func (e *evmxgo) getTokens(req *evmxgotypes.ReqEvmxgos) (types.Message, error) {
	replyTokens := &evmxgotypes.ReplyEvmxgos{}
	tokens, err := e.listTokenKeys(req)
	if err != nil {
		return nil, err
	}
	elog.Error("token Query GetTokens", "get count", len(tokens))

	for _, t1 := range tokens {
		// delete impl by set nil
		if len(t1) == 0 {
			continue
		}

		var evmxgoValue evmxgotypes.LocalEvmxgo
		err = types.Decode(t1, &evmxgoValue)
		if err == nil {
			replyTokens.Tokens = append(replyTokens.Tokens, &evmxgoValue)
		}
	}

	//tokenlog.Info("token Query", "replyTokens", replyTokens)
	return replyTokens, nil
}

func (e *evmxgo) listTokenKeys(req *evmxgotypes.ReqEvmxgos) ([][]byte, error) {
	querydb := e.GetLocalDB()
	if req.QueryAll {
		keys, err := querydb.List(calcEvmxgoKeyLocal(), nil, 0, 0)
		if err != nil && err != types.ErrNotFound {
			return nil, err
		}
		if len(keys) == 0 {
			return nil, types.ErrNotFound
		}
		elog.Debug("token Query GetTokens", "get count", len(keys))
		return keys, nil
	}
	var keys [][]byte
	for _, token := range req.Tokens {
		keys1, err := querydb.List(calcEvmxgoStatusKeyLocal(token), nil, 0, 0)
		if err != nil && err != types.ErrNotFound {
			return nil, err
		}
		keys = append(keys, keys1...)

		elog.Debug("token Query GetTokens", "get count", len(keys))
	}
	if len(keys) == 0 {
		return nil, types.ErrNotFound
	}
	return keys, nil
}

func (e *evmxgo) makeTokenTxKvs(tx *types.Transaction, action *evmxgotypes.EvmxgoAction, receipt *types.ReceiptData, index int, isDel bool) ([]*types.KeyValue, error) {
	var kvs []*types.KeyValue
	var symbol string
	if action.Ty == evmxgotypes.ActionTransfer {
		symbol = action.GetTransfer().Cointoken
	} else if action.Ty == evmxgotypes.ActionWithdraw {
		symbol = action.GetWithdraw().Cointoken
	} else if action.Ty == evmxgotypes.EvmxgoActionTransferToExec {
		symbol = action.GetTransferToExec().Cointoken
	} else {
		return kvs, nil
	}

	kvs, err := tokenTxKvs(tx, symbol, e.GetHeight(), int64(index), isDel)
	return kvs, err
}

func (e *evmxgo) getTokenInfo(symbol string) (types.Message, error) {
	if symbol == "" {
		return nil, types.ErrInvalidParam
	}
	key := calcEvmxgoStatusKeyLocal(symbol)
	values, err := e.GetLocalDB().List(key, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 || values[0] == nil || len(values[0]) == 0 {
		return nil, types.ErrNotFound
	}
	var tokenInfo evmxgotypes.LocalEvmxgo
	err = types.Decode(values[0], &tokenInfo)
	if err != nil {
		return &tokenInfo, err
	}
	return &tokenInfo, nil
}

func (e *evmxgo) getBalance(req *types.ReqBalance) ([]*types.Account, error) {
	cfg := e.GetAPI().GetConfig()
	accountdb, err := account.NewAccountDB(cfg, evmxgotypes.EvmxgoX, req.GetAssetSymbol(), nil)
	if err != nil {
		return nil, err
	}

	switch req.GetExecer() {
	case cfg.ExecName(evmxgotypes.EvmxgoX):
		queryAddrs := req.GetAddresses()

		accounts, err := accountdb.LoadAccounts(e.GetAPI(), queryAddrs)
		if err != nil {
			log.Error("GetTokenBalance", "err", err.Error(), "token symbol", req.GetAssetSymbol(), "address", queryAddrs)
			return nil, err
		}
		return accounts, nil

	default:
		execaddress := address.ExecAddress(req.GetExecer())
		addrs := req.GetAddresses()
		var accounts []*types.Account
		for _, addr := range addrs {
			acc, err := accountdb.LoadExecAccountQueue(e.GetAPI(), addr, execaddress)
			if err != nil {
				log.Error("GetTokenBalance for exector", "err", err.Error(), "token symbol", req.GetAssetSymbol(),
					"address", addr)
				continue
			}
			accounts = append(accounts, acc)
		}

		return accounts, nil
	}
}
