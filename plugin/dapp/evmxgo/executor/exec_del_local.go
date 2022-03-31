package executor

import (
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
	"github.com/jinzhu/copier"
)

/*
 * 实现区块回退时本地执行的数据清除
 */

// ExecDelLocal localdb kv数据自动回滚接口
func (e *evmxgo) ExecDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kvs, err := e.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (e *evmxgo) ExecDelLocal_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := e.ExecDelLocalLocalTransWithdraw(tx, receiptData, index)
	if err != nil {
		return nil, err
	}
	if subCfg.SaveTokenTxList {
		action := evmxgotypes.EvmxgoAction{
			Ty: evmxgotypes.ActionTransfer,
			Value: &evmxgotypes.EvmxgoAction_Transfer{
				Transfer: payload,
			},
		}
		kvs, err := e.makeTokenTxKvs(tx, &action, receiptData, index, true)
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kvs...)
	}
	return set, nil
}

func (e *evmxgo) ExecDelLocal_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := e.ExecDelLocalLocalTransWithdraw(tx, receiptData, index)
	if err != nil {
		return nil, err
	}
	if subCfg.SaveTokenTxList {
		tokenAction := evmxgotypes.EvmxgoAction{
			Ty: evmxgotypes.ActionWithdraw,
			Value: &evmxgotypes.EvmxgoAction_Withdraw{
				Withdraw: payload,
			},
		}
		kvs, err := e.makeTokenTxKvs(tx, &tokenAction, receiptData, index, true)
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kvs...)
	}
	return set, nil
}

func (e *evmxgo) ExecDelLocal_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := e.ExecDelLocalLocalTransWithdraw(tx, receiptData, index)
	if err != nil {
		return nil, err
	}
	if subCfg.SaveTokenTxList {
		tokenAction := evmxgotypes.EvmxgoAction{
			Ty: evmxgotypes.EvmxgoActionTransferToExec,
			Value: &evmxgotypes.EvmxgoAction_TransferToExec{
				TransferToExec: payload,
			},
		}
		kvs, err := e.makeTokenTxKvs(tx, &tokenAction, receiptData, index, true)
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kvs...)
	}
	return set, nil
}

func resetMint(e *evmxgotypes.LocalEvmxgo, height, time, amount int64) *evmxgotypes.LocalEvmxgo {
	e.Total = e.Total - amount
	return e
}

func resetBurn(e *evmxgotypes.LocalEvmxgo, height, time, amount int64) *evmxgotypes.LocalEvmxgo {
	e.Total = e.Total + amount
	return e
}

func (e *evmxgo) ExecDelLocal_MintMap(payload *evmxgotypes.EvmxgoMintMap, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	pay := &evmxgotypes.EvmxgoMint{}
	_ = copier.Copy(pay, payload)
	return e.ExecDelLocal_Mint(pay, tx, receiptData, index)
}

func (e *evmxgo) ExecDelLocal_Mint(payload *evmxgotypes.EvmxgoMint, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	localToken, err := loadLocalToken(payload.Symbol, e.GetLocalDB())
	if err != nil {
		return nil, err
	}
	localToken = resetMint(localToken, e.GetHeight(), e.GetBlockTime(), payload.Amount)
	key := calcEvmxgoStatusKeyLocal(payload.Symbol)
	var set []*types.KeyValue
	set = append(set, &types.KeyValue{Key: key, Value: types.Encode(localToken)})

	table := NewLogsTable(e.GetLocalDB())
	txIndex := dapp.HeightIndexStr(e.GetHeight(), int64(index))
	err = table.Del([]byte(txIndex))
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

func (e *evmxgo) ExecDelLocal_BurnMap(payload *evmxgotypes.EvmxgoBurnMap, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	pay := &evmxgotypes.EvmxgoBurn{}
	_ = copier.Copy(pay, payload)
	return e.ExecDelLocal_Burn(pay, tx, receiptData, index)
}

func (e *evmxgo) ExecDelLocal_Burn(payload *evmxgotypes.EvmxgoBurn, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	localToken, err := loadLocalToken(payload.Symbol, e.GetLocalDB())
	if err != nil {
		return nil, err
	}
	localToken = resetBurn(localToken, e.GetHeight(), e.GetBlockTime(), payload.Amount)
	key := calcEvmxgoStatusKeyLocal(payload.Symbol)
	var set []*types.KeyValue
	set = append(set, &types.KeyValue{Key: key, Value: types.Encode(localToken)})

	table := NewLogsTable(e.GetLocalDB())
	txIndex := dapp.HeightIndexStr(e.GetHeight(), int64(index))
	err = table.Del([]byte(txIndex))
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
