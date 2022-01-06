package wallet

import (
	"encoding/hex"
	"github.com/33cn/chain33/common/address"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func CreateRawTx(actionTy int32, tokenId int32, amount uint64, ethAddress string, accountId int32, toAccountId int32, toEthAddress string, cfg *rpctypes.ChainConfigInfo) (string, error) {
	var action *zt.ZksyncAction
	switch actionTy {
	case zt.TyDepositAction:
		action = &zt.ZksyncAction{
			Ty: actionTy,
			Value: &zt.ZksyncAction_Deposit{
				Deposit: &zt.Deposit{
					ChainType:  "ETH",
					TokenId:    tokenId,
					Amount:     amount,
					EthAddress: ethAddress,
				},
			},
		}
	case zt.TyWithdrawAction:
		action = &zt.ZksyncAction{
			Ty: actionTy,
			Value: &zt.ZksyncAction_Withdraw{
				Withdraw: &zt.Withdraw{
					ChainType: "ETH",
					TokenId:   tokenId,
					Amount:    amount,
					AccountId: accountId,
				},
			},
		}
	case zt.TyContractToLeafAction:
		action = &zt.ZksyncAction{
			Ty: actionTy,
			Value: &zt.ZksyncAction_ContractToLeaf{
				ContractToLeaf: &zt.ContractToLeaf{
					ChainType: "ETH",
					TokenId:   tokenId,
					Amount:    amount,
					AccountId: accountId,
				},
			},
		}
	case zt.TyLeafToContractAction:
		action = &zt.ZksyncAction{
			Ty: actionTy,
			Value: &zt.ZksyncAction_LeafToContract{
				LeafToContract: &zt.LeafToContract{
					ChainType: "ETH",
					TokenId:   tokenId,
					Amount:    amount,
					AccountId: accountId,
				},
			},
		}
	case zt.TyTransferAction:
		action = &zt.ZksyncAction{
			Ty: actionTy,
			Value: &zt.ZksyncAction_Transfer{
				Transfer: &zt.Transfer{
					ChainType:     "ETH",
					TokenId:       tokenId,
					Amount:        amount,
					FromAccountId: accountId,
					ToAccountId:   toAccountId,
				},
			},
		}
	case zt.TyTransferToNewAction:
		action = &zt.ZksyncAction{
			Ty: actionTy,
			Value: &zt.ZksyncAction_TransferToNew{
				TransferToNew: &zt.TransferToNew{
					ChainType:     "ETH",
					TokenId:       tokenId,
					Amount:        amount,
					FromAccountId: accountId,
					ToEthAddress:  toEthAddress,
				},
			},
		}
	case zt.TyForceExitAction:
		action = &zt.ZksyncAction{
			Ty: actionTy,
			Value: &zt.ZksyncAction_ForceQuit{
				ForceQuit: &zt.ForceQuit{
					ChainType:  "ETH",
					TokenId:    tokenId,
					EthAddress: ethAddress,
				},
			},
		}
	}

	sign := &zt.EddsaSigNature{
		Action: action,
	}

	tx := &types.Transaction{Execer: []byte(zt.Zksync), Payload: types.Encode(sign), To: address.ExecAddress(zt.Zksync)}

	tx, err := types.FormatTxExt(cfg.ChainID, false, cfg.MinTxFeeRate, zt.Zksync, tx)
	if err != nil {
		return "", err
	}
	txHex := types.Encode(tx)
	return hex.EncodeToString(txHex), nil
}

func Get()  {
	
}
