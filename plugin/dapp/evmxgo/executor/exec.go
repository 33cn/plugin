package executor

import (
	"errors"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/types"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (e *evmxgo) Exec_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	token := payload.GetCointoken()
	cfg := e.GetAPI().GetConfig()
	db, err := account.NewAccountDB(cfg, e.GetName(), token, e.GetStateDB())
	if err != nil {
		return nil, err
	}
	action := evmxgotypes.EvmxgoAction{
		Ty: evmxgotypes.ActionTransfer,
		Value: &evmxgotypes.EvmxgoAction_Transfer{
			Transfer: payload,
		},
	}
	return e.ExecTransWithdraw(db, tx, &action, index)
}

func (e *evmxgo) Exec_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	token := payload.GetCointoken()
	cfg := e.GetAPI().GetConfig()
	db, err := account.NewAccountDB(cfg, e.GetName(), token, e.GetStateDB())
	if err != nil {
		return nil, err
	}
	action := evmxgotypes.EvmxgoAction{
		Ty: evmxgotypes.ActionWithdraw,
		Value: &evmxgotypes.EvmxgoAction_Withdraw{
			Withdraw: payload,
		},
	}
	return e.ExecTransWithdraw(db, tx, &action, index)
}

func (e *evmxgo) Exec_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	token := payload.GetCointoken()
	cfg := e.GetAPI().GetConfig()
	db, err := account.NewAccountDB(cfg, e.GetName(), token, e.GetStateDB())
	if err != nil {
		return nil, err
	}
	action := evmxgotypes.EvmxgoAction{
		Ty: evmxgotypes.EvmxgoActionTransferToExec,
		Value: &evmxgotypes.EvmxgoAction_TransferToExec{
			TransferToExec: payload,
		},
	}
	return e.ExecTransWithdraw(db, tx, &action, index)
}

func (e *evmxgo) Exec_Mint(payload *evmxgotypes.EvmxgoMint, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newEvmxgoAction(e, "", tx)
	txGroup, err := e.GetTxGroup(index)
	if nil != err {
		return nil, err
	}
	if len(txGroup) < 2 || index == 0 {
		return nil, errors.New("Mint tx should be included in lock tx group")
	}
	txs := e.GetTxs()
	return action.mint(payload, txs[index-1])
}

func (e *evmxgo) Exec_Burn(payload *evmxgotypes.EvmxgoBurn, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newEvmxgoAction(e, "", tx)
	return action.burn(payload)
}
