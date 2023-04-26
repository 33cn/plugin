package types

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

const (
	EthAddrLen = 40
	BTYAddrLen = 64
)

func Str2Byte(v string) []byte {
	var f fr.Element
	f.SetString(v)
	b := f.Bytes()
	return b[:]
}

func Byte2Str(v []byte) string {
	var f fr.Element
	f.SetBytes(v)
	return f.String()
}

func Byte2Uint64(v []byte) uint64 {
	return new(big.Int).SetBytes(v).Uint64()
}

// HexAddr2Decimal 16进制地址转10进制
func HexAddr2Decimal(addr string) (string, bool) {
	addr = FilterHexPrefix(addr)
	addrInt, ok := new(big.Int).SetString(strings.ToLower(addr), 16)
	if !ok {
		return "", false
	}
	return addrInt.String(), true
}

// DecimalAddr2Hex 10进制地址转16进制 需要传入地址期望长度 目前有两种地址格式 一种长度为40 另一种长度为64
func DecimalAddr2Hex(addr string, l int) (string, bool) {
	addrInt, ok := new(big.Int).SetString(strings.ToLower(addr), 10)
	if !ok {
		return "", false
	}
	// 会少前置0 需要补齐0
	return fmt.Sprintf("%0*s", l, hex.EncodeToString(addrInt.Bytes())), true
}

// DecodePacVal decode pac val with man+exp format
func DecodePacVal(p []byte, expBitWidth int) string {
	var v, expMask big.Int
	v.SetBytes(p)
	expMask.SetString(strings.Repeat("1", expBitWidth), 2)
	expV := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).And(&v, &expMask), nil)
	manV := big.NewInt(0).Rsh(&v, uint(expBitWidth))
	return Byte2Str(big.NewInt(0).Mul(manV, expV).Bytes())
}

func SplitNFTContent(contentHash string) (*big.Int, *big.Int, string, error) {
	hexContent := strings.ToLower(contentHash)
	if hexContent[0:2] == "0x" || hexContent[0:2] == "0X" {
		hexContent = hexContent[2:]
	}

	if len(hexContent) != 64 {
		return nil, nil, "", errors.Wrapf(types.ErrInvalidParam, "contentHash not 64 len, %s", hexContent)
	}
	part1, ok := big.NewInt(0).SetString(hexContent[:32], 16)
	if !ok {
		return nil, nil, "", errors.Wrapf(types.ErrInvalidParam, "contentHash.preHalf hex err, %s", hexContent[:32])
	}
	part2, ok := big.NewInt(0).SetString(hexContent[32:], 16)
	if !ok {
		return nil, nil, "", errors.Wrapf(types.ErrInvalidParam, "contentHash.postHalf hex err, %s", hexContent[32:])
	}
	return part1, part2, hexContent, nil
}

// ZkFindExponentPart 找到一个数的指数数量，最大31个0，也就是10^31,循环div 10， 找到最后一个余数非0
func ZkFindExponentPart(s string) int {
	//如果位数很少，直接返回0
	if len(s) <= 1 {
		return 0
	}

	count := 0
	for i := 0; i < len(s); i++ {
		if string(s[len(s)-1-i]) != "0" {
			break
		}
		count++
	}

	//最大不会超过（MaxExponentVal-1）,如果超过MaxExponentVal个0，只截取到MaxExponentVal-1个
	if count > MaxExponentVal-1 {
		return MaxExponentVal - 1
	}
	return count
}

// ZkTransferManExpPart 获取s的man和exp部分，exp部分只统计0的个数，man部分为尾部去掉0部分
func ZkTransferManExpPart(s string) (string, int) {
	exp := ZkFindExponentPart(s)
	return s[0 : len(s)-exp], exp
}

func GetOpChunkNum(opType uint32) (int, error) {
	switch opType {
	case TyDepositAction:
		return DepositChunks, nil
	case TyWithdrawAction:
		return WithdrawChunks, nil
	case TyContractToTreeAction:
		return Contract2TreeChunks, nil
	case TyContractToTreeNewAction:
		return Contract2TreeNewChunks, nil
	case TyTreeToContractAction:
		return Tree2ContractChunks, nil
	case TyTransferAction:
		return TransferChunks, nil
	case TyTransferToNewAction:
		return Transfer2NewChunks, nil
	case TyProxyExitAction:
		return ProxyExitChunks, nil
	case TyFullExitAction:
		return FullExitChunks, nil
	case TySetPubKeyAction:
		return SetPubKeyChunks, nil
	case TyFeeAction:
		return FeeChunks, nil
	case TyMintNFTAction:
		return MintNFTChunks, nil
	case TyWithdrawNFTAction:
		return WithdrawNFTChunks, nil
	case TyTransferNFTAction:
		return TransferNFTChunks, nil
	case TySwapAction:
		return SwapChunks, nil
	default:
		return 0, errors.Wrapf(types.ErrInvalidParam, "operation tx type=%d not support", opType)
	}
}

func FilterHexPrefix(s string) string {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return s[2:]
	}
	return s
}
