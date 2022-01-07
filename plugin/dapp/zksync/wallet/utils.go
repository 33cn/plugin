package wallet

import (
	"encoding/hex"
	"github.com/33cn/chain33/common/address"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func CreateRawTx(actionTy int32, tokenId int32, amount uint64, ethAddress string, toEthAddress string,
	chain33Addr string, cfg *rpctypes.ChainConfigInfo) (string, error) {
	action := new(zt.ZksyncAction)
	action.Ty = actionTy
	switch actionTy {
	case zt.TyDepositAction:
		action.Value = &zt.ZksyncAction_Deposit{
			Deposit: &zt.Deposit{
				ChainType:  "ETH",
				TokenId:    tokenId,
				Amount:     amount,
				EthAddress: ethAddress,
				Chain33Addr: chain33Addr,
			},
		}
	case zt.TyWithdrawAction:
		action.Value = &zt.ZksyncAction_Withdraw{
			Withdraw: &zt.Withdraw{
				ChainType: "ETH",
				TokenId:   tokenId,
				Amount:    amount,
				EthAddress: ethAddress,
			},
		}

	case zt.TyContractToLeafAction:
		action.Value = &zt.ZksyncAction_ContractToLeaf{
			ContractToLeaf: &zt.ContractToLeaf{
				ChainType: "ETH",
				TokenId:   tokenId,
				Amount:    amount,
				EthAddress: ethAddress,
			},
		}
	case zt.TyLeafToContractAction:
		action.Value = &zt.ZksyncAction_LeafToContract{
			LeafToContract: &zt.LeafToContract{
				ChainType: "ETH",
				TokenId:   tokenId,
				Amount:    amount,
				EthAddress: ethAddress,
			},
		}
	case zt.TyTransferAction:
		action.Value = &zt.ZksyncAction_Transfer{
			Transfer: &zt.Transfer{
				ChainType:     "ETH",
				TokenId:       tokenId,
				Amount:        amount,
				FromEthAddress: ethAddress,
				ToEthAddress:   toEthAddress,
			},
		}
	case zt.TyTransferToNewAction:
		action.Value = &zt.ZksyncAction_TransferToNew{
			TransferToNew: &zt.TransferToNew{
				ChainType:     "ETH",
				TokenId:       tokenId,
				Amount:        amount,
				FromEthAddress: ethAddress,
				ToEthAddress:  toEthAddress,
				ToChain33Address: chain33Addr,
			},
		}
	case zt.TyForceExitAction:
		action.Value = &zt.ZksyncAction_ForceQuit{
			ForceQuit: &zt.ForceQuit{
				ChainType:  "ETH",
				TokenId:    tokenId,
				EthAddress: ethAddress,
			},
		}
	case zt.TySetPubKeyAction:
		action.Value = &zt.ZksyncAction_SetPubKey{
			SetPubKey: &zt.SetPubKey{
				EthAddress: ethAddress,
			},
		}
	default:
		return "", types.ErrNotSupport
	}

	tx := &types.Transaction{Execer: []byte(zt.Zksync), Payload: types.Encode(action), To: address.ExecAddress(zt.Zksync)}

	tx, err := types.FormatTxExt(cfg.ChainID, false, cfg.MinTxFeeRate, zt.Zksync, tx)
	if err != nil {
		return "", err
	}
	txHex := types.Encode(tx)
	return hex.EncodeToString(txHex), nil
}