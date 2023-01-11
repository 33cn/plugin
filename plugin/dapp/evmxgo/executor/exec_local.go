package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
	"github.com/jinzhu/copier"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (e *evmxgo) ExecLocal_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := e.ExecLocalTransWithdraw(tx, receiptData, index)
	if err != nil {
		return nil, err
	}
	// 添加个人资产列表
	kv := AddTokenToAssets(payload.To, e.GetLocalDB(), payload.Cointoken)
	if kv != nil {
		set.KV = append(set.KV, kv...)
	}
	if subCfg.SaveTokenTxList {
		evmxgoAction := evmxgotypes.EvmxgoAction{
			Ty: evmxgotypes.ActionTransfer,
			Value: &evmxgotypes.EvmxgoAction_Transfer{
				Transfer: payload,
			},
		}
		kvs, err := e.makeTokenTxKvs(tx, &evmxgoAction, receiptData, index, false)
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kvs...)
	}
	return set, nil
}

func (e *evmxgo) ExecLocal_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := e.ExecLocalTransWithdraw(tx, receiptData, index)
	if err != nil {
		return nil, err
	}
	// 添加个人资产列表
	kv := AddTokenToAssets(tx.From(), e.GetLocalDB(), payload.Cointoken)
	if kv != nil {
		set.KV = append(set.KV, kv...)
	}
	if subCfg.SaveTokenTxList {
		evmxgoAction := evmxgotypes.EvmxgoAction{
			Ty: evmxgotypes.ActionWithdraw,
			Value: &evmxgotypes.EvmxgoAction_Withdraw{
				Withdraw: payload,
			},
		}
		kvs, err := e.makeTokenTxKvs(tx, &evmxgoAction, receiptData, index, false)
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kvs...)
	}
	return set, nil
}

func (e *evmxgo) ExecLocal_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := e.ExecLocalTransWithdraw(tx, receiptData, index)
	if err != nil {
		return nil, err
	}
	if subCfg.SaveTokenTxList {
		evmxgoAction := evmxgotypes.EvmxgoAction{
			Ty: evmxgotypes.EvmxgoActionTransferToExec,
			Value: &evmxgotypes.EvmxgoAction_TransferToExec{
				TransferToExec: payload,
			},
		}
		kvs, err := e.makeTokenTxKvs(tx, &evmxgoAction, receiptData, index, false)
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kvs...)
	}
	return set, nil
}

func loadLocalToken(symbol string, db db.KVDB) (*evmxgotypes.LocalEvmxgo, error) {
	key := calcEvmxgoStatusKeyLocal(symbol)
	v, err := db.Get(key)
	if err != nil {
		return nil, evmxgotypes.ErrEvmxgoSymbolNotExist
	}
	var localToken evmxgotypes.LocalEvmxgo
	err = types.Decode(v, &localToken)
	if err != nil {
		return nil, err
	}
	return &localToken, nil
}

func newLocalEvmxgo(mint *evmxgotypes.EvmxgoMint) *evmxgotypes.LocalEvmxgo {
	e := evmxgotypes.LocalEvmxgo{}
	e.Symbol = mint.GetSymbol()
	return &e
}

func setMint(t *evmxgotypes.LocalEvmxgo, height, time, amount int64) *evmxgotypes.LocalEvmxgo {
	t.Total = t.Total + amount
	return t
}

func setBurn(t *evmxgotypes.LocalEvmxgo, height, time, amount int64) *evmxgotypes.LocalEvmxgo {
	t.Total = t.Total - amount
	return t
}

func (e *evmxgo) ExecLocal_MintMap(payload *evmxgotypes.EvmxgoMintMap, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	pay := &evmxgotypes.EvmxgoMint{}
	_ = copier.Copy(pay, payload)
	//return e.ExecLocal_Mint(pay, tx, receiptData, index)
	localToken, err := loadLocalToken(payload.Symbol, e.GetLocalDB())
	if err != nil && err != evmxgotypes.ErrEvmxgoSymbolNotExist {
		return nil, err
	}
	// evmxgo合约，只要配置了就可以铸币
	if err == evmxgotypes.ErrEvmxgoSymbolNotExist {
		configSynbol, err := loadEvmxgoMintMapConfig(e.GetStateDB(), payload.GetSymbol())
		if err != nil || configSynbol == nil {
			elog.Error("evmxgo mint ", "not config symbol", payload.GetSymbol(), "error", err)
			return nil, evmxgotypes.ErrEvmxgoSymbolNotAllowedMint
		}

		localToken = newLocalEvmxgo(pay)
		localToken.Introduction = configSynbol.Introduction
		localToken.Precision = configSynbol.Precision
	}

	localToken = setMint(localToken, e.GetHeight(), e.GetBlockTime(), payload.Amount)
	var set []*types.KeyValue
	key := calcEvmxgoStatusKeyLocal(payload.Symbol)
	set = append(set, &types.KeyValue{Key: key, Value: types.Encode(localToken)})

	table := NewLogsTable(e.GetLocalDB())
	txIndex := dapp.HeightIndexStr(e.GetHeight(), int64(index))
	err = table.Add(&evmxgotypes.LocalEvmxgoLogs{Symbol: payload.Symbol, TxIndex: txIndex, ActionType: evmxgotypes.EvmxgoActionMint, TxHash: "0x" + hex.EncodeToString(tx.Hash())})
	if err != nil {
		return nil, err
	}
	kv, err := table.Save()
	if err != nil {
		return nil, err
	}
	set = append(set, kv...)

	return &types.LocalDBSet{KV: set}, nil
}

func (e *evmxgo) ExecLocal_Mint(payload *evmxgotypes.EvmxgoMint, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	localToken, err := loadLocalToken(payload.Symbol, e.GetLocalDB())

	if err != nil && err != evmxgotypes.ErrEvmxgoSymbolNotExist {
		return nil, err
	}
	// evmxgo合约，只要配置了就可以铸币
	if err == evmxgotypes.ErrEvmxgoSymbolNotExist {
		configSynbol, err := loadEvmxgoMintConfig(e.GetStateDB(), payload.GetSymbol())
		if err != nil || configSynbol == nil {
			elog.Error("evmxgo mint ", "not config symbol", payload.GetSymbol(), "error", err)
			return nil, evmxgotypes.ErrEvmxgoSymbolNotAllowedMint
		}

		localToken = newLocalEvmxgo(payload)
		localToken.Introduction = configSynbol.Introduction
		localToken.Precision = configSynbol.Precision
	}

	localToken = setMint(localToken, e.GetHeight(), e.GetBlockTime(), payload.Amount)
	var set []*types.KeyValue
	key := calcEvmxgoStatusKeyLocal(payload.Symbol)
	set = append(set, &types.KeyValue{Key: key, Value: types.Encode(localToken)})

	table := NewLogsTable(e.GetLocalDB())
	txIndex := dapp.HeightIndexStr(e.GetHeight(), int64(index))
	err = table.Add(&evmxgotypes.LocalEvmxgoLogs{Symbol: payload.Symbol, TxIndex: txIndex, ActionType: evmxgotypes.EvmxgoActionMint, TxHash: "0x" + hex.EncodeToString(tx.Hash())})
	if err != nil {
		return nil, err
	}
	kv, err := table.Save()
	if err != nil {
		return nil, err
	}
	set = append(set, kv...)

	return &types.LocalDBSet{KV: set}, nil
}

func (e *evmxgo) ExecLocal_Burn(payload *evmxgotypes.EvmxgoBurn, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	localToken, err := loadLocalToken(payload.Symbol, e.GetLocalDB())
	if err != nil {
		return nil, err
	}
	localToken = setBurn(localToken, e.GetHeight(), e.GetBlockTime(), payload.Amount)
	var set []*types.KeyValue
	key := calcEvmxgoStatusKeyLocal(payload.Symbol)
	set = append(set, &types.KeyValue{Key: key, Value: types.Encode(localToken)})

	table := NewLogsTable(e.GetLocalDB())
	txIndex := dapp.HeightIndexStr(e.GetHeight(), int64(index))
	err = table.Add(&evmxgotypes.LocalEvmxgoLogs{Symbol: payload.Symbol, TxIndex: txIndex, ActionType: evmxgotypes.EvmxgoActionBurn, TxHash: "0x" + hex.EncodeToString(tx.Hash())})
	if err != nil {
		return nil, err
	}
	kv, err := table.Save()
	if err != nil {
		return nil, err
	}
	set = append(set, kv...)

	return &types.LocalDBSet{KV: set}, nil
}

func (e *evmxgo) ExecLocal_BurnMap(payload *evmxgotypes.EvmxgoBurnMap, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	pay := &evmxgotypes.EvmxgoBurn{}
	_ = copier.Copy(pay, payload)
	return e.ExecLocal_Burn(pay, tx, receiptData, index)
}

//当区块回滚时，框架支持自动回滚localdb kv，需要对exec-local返回的kv进行封装
func (e *evmxgo) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = e.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}
