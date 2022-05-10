package wallet

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

func CreateRawTx(actionTy int32, tokenId uint64, amount string, ethAddress string, toEthAddress string,
	chain33Addr string, accountId uint64, toAccountId uint64) ([]byte, error) {
	var payload []byte
	switch actionTy {
	case zt.TyWithdrawAction:
		withdraw := &zt.ZkWithdraw{
			TokenId:   tokenId,
			Amount:    amount,
			AccountId: accountId,
		}
		payload = types.MustPBToJSON(withdraw)
	case zt.TyContractToTreeAction:
		contractToLeaf := &zt.ZkContractToTree{
			TokenId:   tokenId,
			Amount:    amount,
			AccountId: accountId,
		}
		payload = types.MustPBToJSON(contractToLeaf)
	case zt.TyTreeToContractAction:
		leafToContract := &zt.ZkTreeToContract{
			TokenId:   tokenId,
			Amount:    amount,
			AccountId: accountId,
		}
		payload = types.MustPBToJSON(leafToContract)
	case zt.TyTransferAction:
		transfer := &zt.ZkTransfer{
			TokenId:       tokenId,
			Amount:        amount,
			FromAccountId: accountId,
			ToAccountId:   toAccountId,
		}
		payload = types.MustPBToJSON(transfer)
	case zt.TyTransferToNewAction:
		transferToNew := &zt.ZkTransferToNew{
			TokenId:          tokenId,
			Amount:           amount,
			FromAccountId:    accountId,
			ToEthAddress:     toEthAddress,
			ToChain33Address: chain33Addr,
		}
		payload = types.MustPBToJSON(transferToNew)
	case zt.TyForceExitAction:
		forceExit := &zt.ZkForceExit{
			TokenId:   tokenId,
			AccountId: accountId,
		}
		payload = types.MustPBToJSON(forceExit)
	//case zt.TySetPubKeyAction:
	//	setPubKey := &zt.ZkSetPubKey{
	//		AccountId: accountId,
	//	}
	//	payload = types.MustPBToJSON(setPubKey)

	case zt.TySetVerifierAction:
		verifier := &zt.ZkVerifier{
			Verifiers: strings.Split(chain33Addr, ","),
		}
		payload = types.MustPBToJSON(verifier)

	default:
		return nil, types.ErrNotSupport
	}

	return payload, nil
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
	return new(fr.Element).SetBigInt(a).String()
}

func StringToByte(s string) []byte {
	byteArray := new(fr.Element).SetString(s).Bytes()
	return byteArray[:]
}

func ChunkStringToByte(s string) []byte {
	f := new(fr.Element).SetString(s)
	chunk := f.Bytes()
	//bits := Byte2Bit(chunk[22:])
	//for i := 0; i < len(bits)/2; i++ {
	//	bits[i], bits[len(bits) - 1 - i] = bits[len(bits) - 1 - i], bits[i]
	//}
	return chunk[32-zt.ChunkBytes:]
}

func Byte2Bit(data []byte) []uint {
	bits := make([]uint, 0)
	for _, v := range data {
		for i := 0; i < 8; i++ {
			bits = append(bits, uint(v>>(7-i)&1))
		}
	}
	return bits
}

func Bit2Byte(bits []uint) []byte {
	data := make([]byte, 0)
	for i := 0; i < len(bits)/8; i++ {
		num := uint(0)
		for j, v := range bits[8*i : 8*(i+1)] {
			num = num + (v << uint(7-j))
		}
		data = append(data, byte(num))
	}
	return data
}

func GetDepositMsg(payload *zt.ZkDeposit) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyDepositAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)

	ethAddress, _ := new(big.Int).SetString(strings.ToLower(payload.EthAddress), 16)
	pubData = append(pubData, getBigEndBitsWithFixLen(ethAddress, zt.AddrBitWidth)...)

	chain33Address, _ := new(big.Int).SetString(payload.Chain33Addr, 16)
	pubData = append(pubData, getBigEndBitsWithFixLen(chain33Address, zt.HashBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetWithdrawMsg(payload *zt.ZkWithdraw) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyWithdrawAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetTreeToContractMsg(payload *zt.ZkTreeToContract) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyTreeToContractAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetContractToTreeMsg(payload *zt.ZkContractToTree) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyContractToTreeAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetTransferMsg(payload *zt.ZkTransfer) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyTransferAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.FromAccountId), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.ToAccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetTransferToNewMsg(payload *zt.ZkTransferToNew) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyTransferToNewAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	amount, _ := new(big.Int).SetString(payload.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.FromAccountId), zt.AccountBitWidth)...)

	ethAddress, _ := new(big.Int).SetString(strings.ToLower(payload.ToEthAddress), 16)

	pubData = append(pubData, getBigEndBitsWithFixLen(ethAddress, zt.AddrBitWidth)...)

	chain33Address, _ := new(big.Int).SetString(payload.ToChain33Address, 16)
	pubData = append(pubData, getBigEndBitsWithFixLen(chain33Address, zt.HashBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetForceExitMsg(payload *zt.ZkForceExit) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyForceExitAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetSetPubKeyMsg(payload *zt.ZkSetPubKey) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TySetPubKeyAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.PubKeyTy), zt.TxTypeBitWidth)...)

	pubKeyX, _ := new(big.Int).SetString(payload.PubKey.X, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(pubKeyX, zt.PubKeyBitWidth)...)
	pubKeyY, _ := new(big.Int).SetString(payload.PubKey.Y, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(pubKeyY, zt.PubKeyBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetFullExitMsg(payload *zt.ZkFullExit) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyFullExitAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.TokenId), zt.TokenBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.AccountId), zt.AccountBitWidth)...)

	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetMintNFTMsg(payload *zt.ZkMintNFT) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	part1, part2, _, err := zt.SplitNFTContent(payload.ContentHash)
	if err != nil {
		fmt.Println(fmt.Sprintf("split content hash=%s wrong", payload.ContentHash))
		panic(err)
	}

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyMintNFTAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.FromAccountId), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.RecipientId), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.ErcProtocol), zt.TxTypeBitWidth)...)
	//nft amount 需要和其他token amount 宽度一致
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.Amount), zt.NFTAmountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(part1, zt.HashBitWidth/2)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(part2, zt.HashBitWidth/2)...)
	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetTransferNFTMsg(payload *zt.ZkTransferNFT) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyTransferNFTAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.FromAccountId), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.RecipientId), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.NFTTokenId), zt.TokenBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.Amount), zt.NFTAmountBitWidth)...)
	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetWithdrawNFTMsg(payload *zt.ZkWithdrawNFT) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TyWithdrawNFTAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.FromAccountId), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.NFTTokenId), zt.TokenBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(payload.Amount), zt.NFTAmountBitWidth)...)
	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}

}

func GetMsgHash(msg *zt.ZkMsg) []byte {
	hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	hash.Write(StringToByte(msg.GetFirst()))
	hash.Write(StringToByte(msg.GetSecond()))
	hash.Write(StringToByte(msg.GetThird()))
	return hash.Sum(nil)
}
