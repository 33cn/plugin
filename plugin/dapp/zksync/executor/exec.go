package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
)

func (z *zksync) Exec_Deposit(payload *zt.ZkDeposit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.Deposit(payload)
}

func (z *zksync) Exec_ZkWithdraw(payload *zt.ZkWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.ZkWithdraw(payload)
}

func (z *zksync) Exec_ZkTransfer(payload *zt.ZkTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.ZkTransfer(payload, zt.TyTransferAction)
}

func (z *zksync) Exec_TransferToNew(payload *zt.ZkTransferToNew, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.TransferToNew(payload)
}

func (z *zksync) Exec_ProxyExit(payload *zt.ZkProxyExit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.ProxyExit(payload)
}

func (z *zksync) Exec_ContractToTree(payload *zt.ZkContractToTree, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.ContractToTree(payload)
}

func (z *zksync) Exec_TreeToContract(payload *zt.ZkTreeToContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.TreeToContract(payload)
}

func (z *zksync) Exec_SetPubKey(payload *zt.ZkSetPubKey, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.SetPubKey(payload)
}

func (z *zksync) Exec_FullExit(payload *zt.ZkFullExit, tx *types.Transaction, index int) (*types.Receipt, error) {
	return nil, errors.Wrapf(types.ErrNotAllow, "fullExit not allow currently")
}

func (z *zksync) Exec_Swap(payload *zt.ZkSwap, tx *types.Transaction, index int) (*types.Receipt, error) {
	//todo swap stub
	return nil, nil
}

func (z *zksync) Exec_SetVerifyKey(payload *zt.ZkVerifyKey, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setVerifyKey(payload)
}

func (z *zksync) Exec_CommitProof(payload *zt.ZkCommitProof, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	r, err := action.commitProof(payload)
	if err != nil {
		zlog.Error("CommitProof", "err", err)
	}
	return r, err
}

func (z *zksync) Exec_SetVerifier(payload *zt.ZkVerifier, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setVerifier(payload)
}

func (z *zksync) Exec_SetFee(payload *zt.ZkSetFee, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setFee(payload)
}

func (z *zksync) Exec_SetTokenSymbol(payload *zt.ZkTokenSymbol, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setTokenSymbol(payload)
}

func (z *zksync) Exec_MintNFT(payload *zt.ZkMintNFT, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.MintNFT(payload)
}

func (z *zksync) Exec_WithdrawNFT(payload *zt.ZkWithdrawNFT, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.withdrawNFT(payload)
}

func (z *zksync) Exec_TransferNFT(payload *zt.ZkTransferNFT, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	//系统设置exodus mode后，则不处理此类交易
	if err := isExodusMode(z.GetStateDB()); err != nil {
		return nil, err
	}
	return action.transferNFT(payload)
}

//Exec_SetExodusMode
func (z *zksync) Exec_SetExodusMode(payload *zt.ZkExodusMode, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.setExodusMode(payload)
}

//zksync作为2层链的数字资产的发行合约，需要支持以下3种类型的资产操作
//Exec_Transfer exec asset transfer process
func (z *zksync) Exec_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.AssetTransfer(payload, tx, index)
}

//Exec_Withdraw exec asset withdraw
func (z *zksync) Exec_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.AssetWithdraw(payload, tx, index)
}

//Exec_TransferToExec exec transfer asset，在平行链上payload里面的ExecName应该是title+Exec，command里面会自动加上，rpc需要注意添加
func (z *zksync) Exec_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(z, tx, index)
	return action.AssetTransferToExec(payload, tx, index)
}
