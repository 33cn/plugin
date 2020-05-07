package executor

import (
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	x2ethereumtypes "github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

//---------------- Ethereum --> Chain33(eth/erc20)-------------------//

// Eth2Chain33类型的交易是Ethereum侧锁定一定金额的eth或者erc20到合约中
// 然后relayer端订阅到该消息后向chain33发送该类型消息
// 本端在验证该类型的请求合理后铸币，并生成相同数额的token
func (x *x2ethereum) Exec_Eth2Chain33(payload *x2ethereumtypes.Eth2Chain33, tx *types.Transaction, index int) (*types.Receipt, error) {
	action, defaultCon := newAction(x, tx, int32(index))
	if payload.ValidatorAddress == "" {
		payload.ValidatorAddress = address.PubKeyToAddr(tx.Signature.Pubkey)
	}
	return action.procMsgEth2Chain33(payload, defaultCon)
}

// 将因ethereum端锁定的eth或者erc20而在chain33端生成的token返还
func (x *x2ethereum) Exec_WithdrawEth(payload *x2ethereumtypes.Eth2Chain33, tx *types.Transaction, index int) (*types.Receipt, error) {
	action, defaultCon := newAction(x, tx, int32(index))
	if payload.ValidatorAddress == "" {
		payload.ValidatorAddress = address.PubKeyToAddr(tx.Signature.Pubkey)
	}

	return action.procWithdrawEth(payload, defaultCon)
}

//---------------- Chain33(eth/erc20) --> Ethereum-------------------//

// WithdrawChain33类型的交易是Chain33侧将本端生成的token返还到Ethereum端
func (x *x2ethereum) Exec_WithdrawChain33(payload *x2ethereumtypes.Chain33ToEth, tx *types.Transaction, index int) (*types.Receipt, error) {
	action, defaultCon := newAction(x, tx, int32(index))
	return action.procMsgBurn(payload, defaultCon)
}

// Chain33ToEth类型的交易是Chain33侧在本端发出申请
// 在本端锁定一定数额的token，然后在ethereum端生成相同数额的token
func (x *x2ethereum) Exec_Chain33ToEth(payload *x2ethereumtypes.Chain33ToEth, tx *types.Transaction, index int) (*types.Receipt, error) {
	action, defaultCon := newAction(x, tx, int32(index))
	return action.procMsgLock(payload, defaultCon)
}

// 转账功能
func (x *x2ethereum) Exec_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action, defaultCon := newAction(x, tx, int32(index))
	return action.procMsgTransfer(payload, defaultCon)
}

//--------------------------合约管理员账户操作-------------------------//

// AddValidator是为了增加validator
func (x *x2ethereum) Exec_AddValidator(payload *x2ethereumtypes.MsgValidator, tx *types.Transaction, index int) (*types.Receipt, error) {
	err := checkTxSignBySpecificAddr(tx, x2ethereumtypes.X2ethereumAdmin)
	if err == nil {
		action, defaultCon := newAction(x, tx, int32(index))
		return action.procAddValidator(payload, defaultCon)
	}
	return nil, err
}

// RemoveValidator是为了移除某一个validator
func (x *x2ethereum) Exec_RemoveValidator(payload *x2ethereumtypes.MsgValidator, tx *types.Transaction, index int) (*types.Receipt, error) {
	err := checkTxSignBySpecificAddr(tx, x2ethereumtypes.X2ethereumAdmin)
	if err == nil {
		action, defaultCon := newAction(x, tx, int32(index))
		return action.procRemoveValidator(payload, defaultCon)
	}
	return nil, err
}

// ModifyPower是为了修改某个validator的power
func (x *x2ethereum) Exec_ModifyPower(payload *x2ethereumtypes.MsgValidator, tx *types.Transaction, index int) (*types.Receipt, error) {
	err := checkTxSignBySpecificAddr(tx, x2ethereumtypes.X2ethereumAdmin)
	if err == nil {
		action, defaultCon := newAction(x, tx, int32(index))
		return action.procModifyValidator(payload, defaultCon)
	}
	return nil, err
}

// SetConsensusThreshold是为了修改对validator所提供的claim达成共识的阈值
func (x *x2ethereum) Exec_SetConsensusThreshold(payload *x2ethereumtypes.MsgConsensusThreshold, tx *types.Transaction, index int) (*types.Receipt, error) {
	err := checkTxSignBySpecificAddr(tx, x2ethereumtypes.X2ethereumAdmin)
	if err == nil {
		action, _ := newAction(x, tx, int32(index))
		return action.procMsgSetConsensusThreshold(payload)
	}
	return nil, err
}

func checkTxSignBySpecificAddr(tx *types.Transaction, addr string) error {
	signAddr := address.PubKeyToAddr(tx.Signature.Pubkey)
	if signAddr != addr {
		return x2ethereumtypes.ErrInvalidAdminAddress
	}
	if tx.CheckSign() == false {
		return types.ErrSign
	}
	return nil
}
