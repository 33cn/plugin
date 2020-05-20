package executor

import (
	"errors"
	"github.com/33cn/chain33/system/dapp"
	manTy "github.com/33cn/chain33/system/dapp/manage/types"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	x2eTy "github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

//---------------- Ethereum(eth/erc20) --> Chain33-------------------//

// Eth2Chain33类型的交易是Ethereum侧锁定一定金额的eth或者erc20到合约中
// 然后relayer端订阅到该消息后向chain33发送该类型消息
// 本端在验证该类型的请求合理后铸币，并生成相同数额的token
func (x *x2ethereum) Exec_Eth2Chain33Lock(payload *x2eTy.Eth2Chain33, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}

	payload.ValidatorAddress = address.PubKeyToAddr(tx.Signature.Pubkey)

	return action.procEth2Chain33_lock(payload)
}

//----------------  Chain33(eth/erc20)------> Ethereum -------------------//
// WithdrawChain33类型的交易是将Eth端因Chain33端锁定所生成的token返还给Chain33端（Burn）
func (x *x2ethereum) Exec_Chain33ToEthBurn(payload *x2eTy.Chain33ToEth, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}
	return action.procChain33ToEth_burn(payload)
}

//---------------- Chain33(eth/erc20) --> Ethereum-------------------//

// 将因ethereum端锁定的eth或者erc20而在chain33端生成的token返还
func (x *x2ethereum) Exec_Eth2Chain33Burn(payload *x2eTy.Eth2Chain33, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(x, tx, int32(index))
	if action == nil {
		return nil, errors.New("Create Action Error")
	}

	payload.ValidatorAddress = address.PubKeyToAddr(tx.Signature.Pubkey)

	return action.procEth2Chain33_burn(payload)
}

// Chain33ToEth类型的交易是Chain33侧在本端发出申请
// 在本端锁定一定数额的token，然后在ethereum端生成相同数额的token
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
	signAddr := address.PubKeyToAddr(tx.Signature.Pubkey)
	var exist bool
	for _, addr := range addrs {
		if signAddr == addr {
			exist = true
			continue
		}
	}

	if !exist {
		return x2eTy.ErrInvalidAdminAddress
	}

	if tx.CheckSign() == false {
		return types.ErrSign
	}
	return nil
}
