package executor

import (
	"errors"

	"github.com/33cn/chain33/system/dapp"
	manTy "github.com/33cn/chain33/system/dapp/manage/types"

	"github.com/33cn/chain33/types"
	x2eTy "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

//---------------- Ethereum(eth/erc20) --> Chain33-------------------//

// 在chain33上为ETH/ERC20铸币
func (x *x2ethereum) Exec_Eth2Chain33Lock(payload *x2eTy.Eth2Chain33, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}

	payload.ValidatorAddress = tx.From()

	return action.procEth2Chain33_lock(payload)
}

//----------------  Chain33(eth/erc20)------> Ethereum -------------------//
// 在chain33端将铸的币销毁，返还给eth
func (x *x2ethereum) Exec_Chain33ToEthBurn(payload *x2eTy.Chain33ToEth, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}
	return action.procChain33ToEth_burn(payload)
}

//---------------- Ethereum (bty) --> Chain33-------------------//
// 在eth端将铸的bty币销毁，返还给chain33
func (x *x2ethereum) Exec_Eth2Chain33Burn(payload *x2eTy.Eth2Chain33, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}

	payload.ValidatorAddress = tx.From()

	return action.procEth2Chain33_burn(payload)
}

//---------------- Chain33 --> Ethereum (bty) -------------------//
// 在 ethereum 上为 chain33 铸币
func (x *x2ethereum) Exec_Chain33ToEthLock(payload *x2eTy.Chain33ToEth, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}
	return action.procChain33ToEth_lock(payload)
}

// 转账功能
func (x *x2ethereum) Exec_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}
	return action.procMsgTransfer(payload)
}

func (x *x2ethereum) Exec_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}
	if !x2eTy.IsExecAddrMatch(payload.ExecName, tx.GetRealToAddr()) {
		return nil, types.ErrToAddrNotSameToExecAddr
	}
	return action.procMsgTransferToExec(payload)
}

func (x *x2ethereum) Exec_WithdrawFromExec(payload *types.AssetsWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}
	if dapp.IsDriverAddress(tx.GetRealToAddr(), x.GetHeight()) || x2eTy.IsExecAddrMatch(payload.ExecName, tx.GetRealToAddr()) {
		return action.procMsgWithDrawFromExec(payload)
	}
	return nil, errors.New("tx error")
}

//--------------------------合约管理员账户操作-------------------------//

// AddValidator是为了增加validator
func (x *x2ethereum) Exec_AddValidator(payload *x2eTy.MsgValidator, tx *types.Transaction, index int) (*types.Receipt, error) {
	confManager := types.ConfSub(x.GetAPI().GetConfig(), manTy.ManageX).GStrList("superManager")
	err := checkTxSignBySpecificAddr(tx, confManager)
	if err == nil {
		action := newAction(x, tx, int32(index))
		if action == nil {
			return nil, errors.New("Create Action Error")
		}
		return action.procAddValidator(payload)
	}
	return nil, err
}

// RemoveValidator是为了移除某一个validator
func (x *x2ethereum) Exec_RemoveValidator(payload *x2eTy.MsgValidator, tx *types.Transaction, index int) (*types.Receipt, error) {
	confManager := types.ConfSub(x.GetAPI().GetConfig(), manTy.ManageX).GStrList("superManager")
	err := checkTxSignBySpecificAddr(tx, confManager)
	if err == nil {
		action := newAction(x, tx, int32(index))
		if action == nil {
			return nil, errors.New("Create Action Error")
		}
		return action.procRemoveValidator(payload)
	}
	return nil, err
}

// ModifyPower是为了修改某个validator的power
func (x *x2ethereum) Exec_ModifyPower(payload *x2eTy.MsgValidator, tx *types.Transaction, index int) (*types.Receipt, error) {
	confManager := types.ConfSub(x.GetAPI().GetConfig(), manTy.ManageX).GStrList("superManager")
	err := checkTxSignBySpecificAddr(tx, confManager)
	if err == nil {
		action := newAction(x, tx, int32(index))
		if action == nil {
			return nil, errors.New("Create Action Error")
		}
		return action.procModifyValidator(payload)
	}
	return nil, err
}

// SetConsensusThreshold是为了修改对validator所提供的claim达成共识的阈值
func (x *x2ethereum) Exec_SetConsensusThreshold(payload *x2eTy.MsgConsensusThreshold, tx *types.Transaction, index int) (*types.Receipt, error) {
	confManager := types.ConfSub(x.GetAPI().GetConfig(), manTy.ManageX).GStrList("superManager")
	err := checkTxSignBySpecificAddr(tx, confManager)
	if err == nil {
		action := newAction(x, tx, int32(index))
		if action == nil {
			return nil, errors.New("Create Action Error")
		}
		return action.procMsgSetConsensusThreshold(payload)
	}
	return nil, err
}

func checkTxSignBySpecificAddr(tx *types.Transaction, addrs []string) error {
	signAddr := tx.From()
	var exist bool
	for _, addr := range addrs {
		if signAddr == addr {
			exist = true
			break
		}
	}

	if !exist {
		return x2eTy.ErrInvalidAdminAddress
	}

	if !tx.CheckSign(0) {
		return types.ErrSign
	}
	return nil
}
