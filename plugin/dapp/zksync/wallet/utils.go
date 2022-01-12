package wallet

import (
	"encoding/hex"
	"fmt"
	"github.com/33cn/chain33/common/address"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"math/big"
)

func CreateRawTx(actionTy int32, tokenId uint64, amount string, ethAddress string, toEthAddress string,
	chain33Addr string, accountId uint64, toAccountId uint64, cfg *rpctypes.ChainConfigInfo) (string, error) {
	action := new(zt.ZksyncAction)
	action.Ty = actionTy
	switch actionTy {
	case zt.TyDepositAction:
		action.Value = &zt.ZksyncAction_Deposit{
			Deposit: &zt.Deposit{
				TokenId:     tokenId,
				Amount:      amount,
				EthAddress:  ethAddress,
				Chain33Addr: chain33Addr,
			},
		}
	case zt.TyWithdrawAction:
		action.Value = &zt.ZksyncAction_Withdraw{
			Withdraw: &zt.Withdraw{
				TokenId:   tokenId,
				Amount:    amount,
				AccountId: accountId,
			},
		}

	case zt.TyContractToLeafAction:
		action.Value = &zt.ZksyncAction_ContractToLeaf{
			ContractToLeaf: &zt.ContractToLeaf{
				TokenId:   tokenId,
				Amount:    amount,
				AccountId: accountId,
			},
		}
	case zt.TyLeafToContractAction:
		action.Value = &zt.ZksyncAction_LeafToContract{
			LeafToContract: &zt.LeafToContract{
				TokenId:   tokenId,
				Amount:    amount,
				AccountId: accountId,
			},
		}
	case zt.TyTransferAction:
		action.Value = &zt.ZksyncAction_Transfer{
			Transfer: &zt.Transfer{
				TokenId:       tokenId,
				Amount:        amount,
				FromAccountId: accountId,
				ToAccountId:   toAccountId,
			},
		}
	case zt.TyTransferToNewAction:
		action.Value = &zt.ZksyncAction_TransferToNew{
			TransferToNew: &zt.TransferToNew{
				TokenId:          tokenId,
				Amount:           amount,
				FromAccountId:    accountId,
				ToEthAddress:     toEthAddress,
				ToChain33Address: chain33Addr,
			},
		}
	case zt.TyForceExitAction:
		action.Value = &zt.ZksyncAction_ForceQuit{
			ForceQuit: &zt.ForceQuit{
				TokenId:   tokenId,
				AccountId: accountId,
			},
		}
	case zt.TySetPubKeyAction:
		action.Value = &zt.ZksyncAction_SetPubKey{
			SetPubKey: &zt.SetPubKey{
				AccountId: accountId,
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

//11 => 00001011, 数组index0值为0，大端表示
func getBigEndBitsWithFixLen(val *big.Int, n uint64) []uint {
	l := val.BitLen()
	if n < uint64(l) {
		panic(fmt.Sprintf("n=%d less than len=%d", n, l))
	}

	//little-end array
	var bits []uint
	for i := 0; i < l; i++ {
		bits = append(bits, val.Bit(i))
	}
	for i := uint64(l); i < n; i++ {
		bits = append(bits, 0)
	}

	for i := uint64(0); i < n/2; i++ {
		bits[i], bits[n-1-i] = bits[n-1-i], bits[i]
	}
	return bits
}

//把bits以小端模式构建big.Int
func setBeBitsToVal(bits []uint) string {
	a := big.NewInt(0)
	for i := 0; i < len(bits); i++ {
		a.SetBit(a, i, bits[i])
	}
	return a.String()
}

func bitToByte(bits []uint) []byte {
	a := big.NewInt(0)
	for i := 0; i < len(bits); i++ {
		a.SetBit(a, i, bits[i])
	}
	return a.Bytes()
}

func GetDepositMsg(payload *zt.Deposit) *zt.Msg {
	var pubData []uint

	binaryData := make([]uint, 752)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyDepositAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)

	ethAddress, _ := new(big.Int).SetString(payload.EthAddress, 16)
	pubData = append(pubData, getBigEndBitsWithFixLen(ethAddress, zt.AddrBitWidth)...)

	chain33Address, _ := new(big.Int).SetString(payload.Chain33Addr, 16)
	pubData = append(pubData, getBigEndBitsWithFixLen(chain33Address, zt.AddrBitWidth)...)

	copy(binaryData, pubData)

	return &zt.Msg{
		First:  bitToByte(binaryData[:zt.MsgFirstWidth]),
		Second: bitToByte(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  bitToByte(binaryData[:zt.MsgFirstWidth]),
	}

}

func GetWithdrawMsg(payload *zt.Withdraw) *zt.Msg {
	var pubData []uint

	binaryData := make([]uint, 752)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyWithdrawAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.Msg{
		First:  bitToByte(binaryData[:zt.MsgFirstWidth]),
		Second: bitToByte(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  bitToByte(binaryData[:zt.MsgFirstWidth]),
	}

}

func GetLeafToContractMsg(payload *zt.LeafToContract) *zt.Msg {
	var pubData []uint

	binaryData := make([]uint, 752)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyLeafToContractAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.Msg{
		First:  bitToByte(binaryData[:zt.MsgFirstWidth]),
		Second: bitToByte(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  bitToByte(binaryData[:zt.MsgFirstWidth]),
	}

}

func GetContractToLeafMsg(payload *zt.ContractToLeaf) *zt.Msg {
	var pubData []uint

	binaryData := make([]uint, 752)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyContractToLeafAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.Msg{
		First:  bitToByte(binaryData[:zt.MsgFirstWidth]),
		Second: bitToByte(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  bitToByte(binaryData[:zt.MsgFirstWidth]),
	}

}

func GetTransferMsg(payload *zt.Transfer) *zt.Msg {
	var pubData []uint

	binaryData := make([]uint, 752)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyTransferAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.FromAccountId), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.ToAccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.Msg{
		First:  bitToByte(binaryData[:zt.MsgFirstWidth]),
		Second: bitToByte(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  bitToByte(binaryData[:zt.MsgFirstWidth]),
	}

}

func GetTransferToNewMsg(payload *zt.TransferToNew) *zt.Msg {
	var pubData []uint

	binaryData := make([]uint, 752)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyTransferToNewAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.FromAccountId), zt.AccountBitWidth)...)

	ethAddress, _ := new(big.Int).SetString(payload.ToEthAddress, 16)
	pubData = append(pubData, getBigEndBitsWithFixLen(ethAddress, zt.AddrBitWidth)...)

	chain33Address, _ := new(big.Int).SetString(payload.ToChain33Address, 16)
	pubData = append(pubData, getBigEndBitsWithFixLen(chain33Address, zt.AddrBitWidth)...)

	copy(binaryData, pubData)

	return &zt.Msg{
		First:  bitToByte(binaryData[:zt.MsgFirstWidth]),
		Second: bitToByte(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  bitToByte(binaryData[:zt.MsgFirstWidth]),
	}

}

func GetForceQuitMsg(payload *zt.ForceQuit) *zt.Msg {
	var pubData []uint

	binaryData := make([]uint, 752)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TySetPubKeyAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.Msg{
		First:  bitToByte(binaryData[:zt.MsgFirstWidth]),
		Second: bitToByte(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  bitToByte(binaryData[:zt.MsgFirstWidth]),
	}

}

func GetSetPubKeyMsg(payload *zt.SetPubKey) *zt.Msg {
	var pubData []uint

	binaryData := make([]uint, 752)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyDepositAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	pubKeyX, _ := new(big.Int).SetString(payload.PubKey.X, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(pubKeyX, zt.PubKeyBitWidth)...)
	pubKeyY, _ := new(big.Int).SetString(payload.PubKey.Y, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(pubKeyY, zt.PubKeyBitWidth)...)

	copy(binaryData, pubData)

	return &zt.Msg{
		First:  bitToByte(binaryData[:zt.MsgFirstWidth]),
		Second: bitToByte(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  bitToByte(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetMsgHash(msg *zt.Msg) []byte {
	hash := mimc.NewMiMC(mixTy.MimcHashSeed)
	hash.Write(msg.GetFirst())
	hash.Write(msg.GetSecond())
	hash.Write(msg.GetThird())
	return hash.Sum(nil)
}