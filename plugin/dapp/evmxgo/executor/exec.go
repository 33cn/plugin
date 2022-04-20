package executor

import (
	"errors"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
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
	action := newEvmxgoAction(e, tx)
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

func (e *evmxgo) Exec_MintMap(mint *evmxgotypes.EvmxgoMintMap, tx *types.Transaction, index int) (*types.Receipt, error) {
	if mint == nil {
		return nil, types.ErrInvalidParam
	}
	if mint.GetAmount() < 0 || mint.GetAmount() > types.MaxTokenBalance || mint.GetSymbol() == "" {
		return nil, types.ErrInvalidParam
	}

	if len(mint.Recipient) == 0 {
		return nil, types.ErrInvalidParam
	}

	//TODO check address
	err := address.CheckAddress(mint.Recipient, -1)
	if err != nil {
		return nil, err
	}
	mintConfig, err := loadEvmxgoMintMapConfig(e.GetStateDB(), mint.GetSymbol())
	if err != nil {
		return nil, err
	}
	// evmxgo合约，配置symbol对应的实际地址，检验地址正确才能发币
	if tx.From() != mintConfig.Address {
		elog.Error("evmxgo mint address error", "GetSymbol:", mint.GetSymbol(), "from:", tx.From(), "configSymbol.Address: ", mintConfig.Address)
		return nil, evmxgotypes.ErrEvmxgoSymbolNotAllowedMint
	}

	action := newEvmxgoAction(e, tx)
	return action.mintMap(mint, tx)
}

func (e *evmxgo) Exec_Burn(payload *evmxgotypes.EvmxgoBurn, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newEvmxgoAction(e, tx)
	return action.burn(payload)
}

func (e *evmxgo) Exec_BurnMap(payload *evmxgotypes.EvmxgoBurnMap, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newEvmxgoAction(e, tx)
	return action.burnMap(payload)
}
