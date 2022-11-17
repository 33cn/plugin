package executor

import (
	"fmt"
	"hash"
	"math/big"

	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func calcPubDataCommitHash(mimcHash hash.Hash, blockStart, blockEnd, chainTitleId uint64, oldRoot, newRoot string, pubDatas []string) string {
	mimcHash.Reset()

	var f fr.Element
	t := f.SetUint64(blockStart).Bytes()
	mimcHash.Write(t[:])

	t = f.SetUint64(blockEnd).Bytes()
	mimcHash.Write(t[:])

	t = f.SetString(oldRoot).Bytes()
	mimcHash.Write(t[:])

	t = f.SetString(newRoot).Bytes()
	mimcHash.Write(t[:])

	t = f.SetUint64(chainTitleId).Bytes()
	mimcHash.Write(t[:])

	for _, r := range pubDatas {
		t = f.SetString(r).Bytes()
		mimcHash.Write(t[:])
	}
	ret := mimcHash.Sum(nil)

	return f.SetBytes(ret).String()
}

func calcOnChainPubDataCommitHash(mimcHash hash.Hash, chainTitleId uint64, newRoot string, pubDatas []string) string {
	mimcHash.Reset()
	var f fr.Element

	t := f.SetString(newRoot).Bytes()
	mimcHash.Write(t[:])
	t = f.SetUint64(chainTitleId).Bytes()
	mimcHash.Write(t[:])

	sum := mimcHash.Sum(nil)

	for _, p := range pubDatas {
		mimcHash.Reset()
		t = f.SetString(p).Bytes()
		mimcHash.Write(sum)
		mimcHash.Write(t[:])
		sum = mimcHash.Sum(nil)
	}

	return f.SetBytes(sum).String()
}

func transferPubDataToOps(pubData []string) []*zt.ZkOperation {
	operations := make([]*zt.ZkOperation, 0)
	start := 0
	for start < len(pubData) {
		chunk := wallet.ChunkStringToByte(pubData[start])
		operationTy := getTyByChunk(chunk)
		chunkNum := getChunkNum(operationTy)
		if operationTy != zt.TyNoopAction {
			operation := getOperationByChunk(pubData[start:start+chunkNum], operationTy)
			//fmt.Println("transferPubDatasToOption.op=", operation)
			operations = append(operations, operation)
		}
		start = start + chunkNum
	}
	return operations
}

func getTyByChunk(chunk []byte) uint64 {
	return new(big.Int).SetBytes(chunk[:1]).Uint64()
}

func getChunkNum(opType uint64) int {
	switch opType {
	case zt.TyNoopAction:
		return zt.NoopChunks
	case zt.TyDepositAction:
		return zt.DepositChunks
	case zt.TyWithdrawAction:
		return zt.WithdrawChunks
	case zt.TyTransferAction:
		return zt.TransferChunks
	case zt.TyTransferToNewAction:
		return zt.Transfer2NewChunks
	case zt.TyProxyExitAction:
		return zt.ProxyExitChunks
	case zt.TySetPubKeyAction:
		return zt.SetPubKeyChunks
	case zt.TyFullExitAction:
		return zt.FullExitChunks
	case zt.TySwapAction:
		return zt.SwapChunks
	case zt.TyContractToTreeAction:
		return zt.Contract2TreeChunks
	case zt.TyContractToTreeNewAction:
		return zt.Contract2TreeNewChunks
	case zt.TyTreeToContractAction:
		return zt.Tree2ContractChunks
	case zt.TyFeeAction:
		return zt.FeeChunks
	case zt.TyMintNFTAction:
		return zt.MintNFTChunks
	case zt.TyWithdrawNFTAction:
		return zt.WithdrawNFTChunks
	case zt.TyTransferNFTAction:
		return zt.TransferNFTChunks

	default:
		panic(fmt.Sprintf("operation tx type=%d not support", opType))
	}

}

func getOperationByChunk(chunks []string, optionTy uint64) *zt.ZkOperation {
	totalChunk := make([]byte, 0)
	for _, chunk := range chunks {
		totalChunk = append(totalChunk, wallet.ChunkStringToByte(chunk)...)
	}
	switch optionTy {
	case zt.TyDepositAction:
		return getDepositOperationByChunk(totalChunk)
	case zt.TyWithdrawAction:
		return getWithDrawOperationByChunk(totalChunk)
	case zt.TyTransferAction:
		return getTransferOperationByChunk(totalChunk)
	case zt.TyTransferToNewAction:
		return getTransfer2NewOperationByChunk(totalChunk)
	case zt.TyProxyExitAction:
		return getProxyExitOperationByChunk(totalChunk)
	case zt.TySetPubKeyAction:
		return getSetPubKeyOperationByChunk(totalChunk)
	case zt.TyFullExitAction:
		return getFullExitOperationByChunk(totalChunk)
	case zt.TySwapAction:
		return getSwapOperationByChunk(totalChunk)
	case zt.TyContractToTreeAction:
		return getContract2TreeOptionByChunk(totalChunk)
	case zt.TyContractToTreeNewAction:
		return getContract2TreeNewOptionByChunk(totalChunk)
	case zt.TyTreeToContractAction:
		return getTree2ContractOperationByChunk(totalChunk)
	case zt.TyFeeAction:
		return getFeeOperationByChunk(totalChunk)
	case zt.TyMintNFTAction:
		return getMintNFTOperationByChunk(totalChunk)
	case zt.TyWithdrawNFTAction:
		return getWithdrawNFTOperationByChunk(totalChunk)
	case zt.TyTransferNFTAction:
		return getTransferNFTOperationByChunk(totalChunk)
	default:
		panic("operationTy not support")
	}
}

func getDepositOperationByChunk(chunk []byte) *zt.ZkOperation {
	deposit := &zt.ZkDepositWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	deposit.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	deposit.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	deposit.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	deposit.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/8
	deposit.Layer2Addr = zt.Byte2Str(chunk[start:end])

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: deposit}}
	return &zt.ZkOperation{Ty: zt.TyDepositAction, Op: special}
}

func getWithDrawOperationByChunk(chunk []byte) *zt.ZkOperation {
	withdraw := &zt.ZkWithdrawWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	withdraw.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	withdraw.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	withdraw.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	withdraw.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	withdraw.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Withdraw{Withdraw: withdraw}}
	return &zt.ZkOperation{Ty: zt.TyWithdrawAction, Op: special}
}

func getSwapOperationByChunk(chunk []byte) *zt.ZkOperation {
	leftOrder := &zt.ZkSwapOrderInfo{}
	rightOrder := &zt.ZkSwapOrderInfo{}
	operation := &zt.ZkSwapWitnessInfo{Left: leftOrder, Right: rightOrder, Fee: &zt.ZkFee{}}

	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	leftOrder.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	rightOrder.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	//1st token, left asset
	operation.LeftTokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	//2nd token, right asset
	operation.RightTokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	//1st amount, left asset amount
	operation.LeftDealAmount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	//2nd amount right asset amount
	operation.RightDealAmount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	//1st fee, left's fee
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Swap{Swap: operation}}
	return &zt.ZkOperation{Ty: zt.TySwapAction, Op: special}
}

func getContract2TreeOptionByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkContractToTreeWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ContractToTree{ContractToTree: operation}}
	return &zt.ZkOperation{Ty: zt.TyContractToTreeAction, Op: special}
}

func getContract2TreeNewOptionByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkContractToTreeNewWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.ToAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/8
	operation.Layer2Addr = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Contract2TreeNew{Contract2TreeNew: operation}}
	return &zt.ZkOperation{Ty: zt.TyContractToTreeAction, Op: special}
}

func getTree2ContractOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkTreeToContractWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TreeToContract{TreeToContract: operation}}
	return &zt.ZkOperation{Ty: zt.TyTreeToContractAction, Op: special}
}

func getTransferOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkTransferWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.FromAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Transfer{Transfer: operation}}
	return &zt.ZkOperation{Ty: zt.TyTransferAction, Op: special}
}

func getTransfer2NewOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkTransferToNewWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.FromAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/8
	operation.Layer2Addr = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TransferToNew{TransferToNew: operation}}
	return &zt.ZkOperation{Ty: zt.TyTransferToNewAction, Op: special}
}

func getSetPubKeyOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkSetPubKeyWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.PubKeyTy = zt.Byte2Uint64(chunk[start:end])
	pubkey := &zt.ZkPubKey{}
	start = end
	end = start + zt.PubKeyBitWidth/8
	pubkey.X = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.PubKeyBitWidth/8
	pubkey.Y = zt.Byte2Str(chunk[start:end])
	operation.PubKey = pubkey

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_SetPubKey{SetPubKey: operation}}
	return &zt.ZkOperation{Ty: zt.TySetPubKeyAction, Op: special}
}

func getProxyExitOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkProxyExitWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	//proxy id
	operation.ProxyID = zt.Byte2Uint64(chunk[start:end])
	start = zt.TxTypeBitWidth / 8
	end = start + zt.AccountBitWidth/8
	//toId
	operation.TargetID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ProxyExit{ProxyExit: operation}}
	return &zt.ZkOperation{Ty: zt.TyProxyExitAction, Op: special}
}

func getFullExitOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkFullExitWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_FullExit{FullExit: operation}}
	return &zt.ZkOperation{Ty: zt.TyFullExitAction, Op: special}
}

func getFeeOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkFeeWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Fee{Fee: operation}}
	return &zt.ZkOperation{Ty: zt.TyFeeAction, Op: special}
}

func getMintNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkMintNFTWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.MintAcctID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.RecipientID = zt.Byte2Uint64(chunk[start:end])
	//ERC 721/1155 protocol
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.ErcProtocol = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.Amount = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.ContentHash = append(operation.ContentHash, zt.Byte2Str(chunk[start:end]))
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.ContentHash = append(operation.ContentHash, zt.Byte2Str(chunk[start:end]))

	start = end
	end = start + zt.TokenBitWidth/8
	operation.Fee.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_MintNFT{MintNFT: operation}}
	return &zt.ZkOperation{Ty: zt.TyMintNFTAction, Op: special}
}

func getWithdrawNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkWithdrawNFTWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	//fromId
	operation.FromAcctID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	//original creator id
	operation.CreatorAcctID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.NFTTokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.CreatorSerialID = zt.Byte2Uint64(chunk[start:end])
	//ERC 721/1155 protocol
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.ErcProtocol = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.InitMintAmount = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.WithdrawAmount = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])

	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.ContentHash = append(operation.ContentHash, zt.Byte2Str(chunk[start:end]))
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.ContentHash = append(operation.ContentHash, zt.Byte2Str(chunk[start:end]))

	start = end
	end = start + zt.TokenBitWidth/8
	operation.Fee.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_WithdrawNFT{WithdrawNFT: operation}}
	return &zt.ZkOperation{Ty: zt.TyWithdrawNFTAction, Op: special}
}

func getTransferNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkTransferNFTWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.FromAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.RecipientID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.NFTTokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.Amount = zt.Byte2Uint64(chunk[start:end])

	start = end
	end = start + zt.TokenBitWidth/8
	operation.Fee.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TransferNFT{TransferNFT: operation}}
	return &zt.ZkOperation{Ty: zt.TyTransferNFTAction, Op: special}
}
