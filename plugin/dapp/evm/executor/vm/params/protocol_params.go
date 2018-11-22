// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

const (
	// MaxCodeSize 合约允许的最大字节数
	MaxCodeSize = 24576
	// CallCreateDepth  合约递归调用最大深度
	CallCreateDepth uint64 = 1024
	// StackLimit 栈允许的最大深度
	StackLimit uint64 = 1024

	// CreateDataGas 创建合约时，按字节计费
	CreateDataGas uint64 = 200
	// CallStipend 每次CALL调用之前，给予一定额度的免费Gas
	CallStipend uint64 = 2300
	// CallValueTransferGas  转账操作
	CallValueTransferGas uint64 = 9000
	// CallNewAccountGas  操作目标地址事先不存在
	CallNewAccountGas uint64 = 25000
	// QuadCoeffDiv  计算开辟内存花费时，在计算出的内存大小平方基础上除此值
	QuadCoeffDiv uint64 = 512
	// CopyGas 内存数据复制时，按字计费
	CopyGas uint64 = 3

	// Sha3Gas SHA3操作
	Sha3Gas uint64 = 30
	// Sha3WordGas SHA3操作的数据按字计费
	Sha3WordGas uint64 = 6
	// SstoreSetGas SSTORE 从零值地址到非零值地址存储
	SstoreSetGas uint64 = 20000
	// SstoreResetGas SSTORE 从非零值地址到非零值地址存储
	SstoreResetGas uint64 = 5000
	// SstoreClearGas SSTORE 从非零值地址到零值地址存储
	SstoreClearGas uint64 = 5000
	// SstoreRefundGas SSTORE 删除值时给予的奖励
	SstoreRefundGas uint64 = 15000
	// JumpdestGas JUMPDEST 指令
	JumpdestGas uint64 = 1
	// LogGas LOGN 操作计费
	LogGas uint64 = 375
	// LogDataGas  LOGN生成的数据，每个字节的计费价格
	LogDataGas uint64 = 8
	// LogTopicGas LOGN 生成日志时，使用N*此值计费
	LogTopicGas uint64 = 375
	// CreateGas CREATE 指令
	CreateGas uint64 = 32000
	// SuicideRefundGas  SUICIDE 操作时给予的奖励
	SuicideRefundGas uint64 = 24000
	// MemoryGas 开辟新内存时按字收费
	MemoryGas uint64 = 3

	// EcrecoverGas  ecrecover 指令
	EcrecoverGas uint64 = 3000
	// Sha256BaseGas SHA256 基础计费
	Sha256BaseGas uint64 = 60
	// Sha256PerWordGas SHA256 按字长计费 （总计费等于两者相加）
	Sha256PerWordGas uint64 = 12
	// Ripemd160BaseGas RIPEMD160 基础计费
	Ripemd160BaseGas uint64 = 600
	// Ripemd160PerWordGas RIPEMD160 按字长计费 （总计费等于两者相加）
	Ripemd160PerWordGas uint64 = 120
	// IdentityBaseGas  dataCopy 基础计费
	IdentityBaseGas uint64 = 15
	// IdentityPerWordGas dataCopy 按字长计费（总计费等于两者相加）
	IdentityPerWordGas uint64 = 3
	// ModExpQuadCoeffDiv 大整数取模运算时计算出的费用除此数
	ModExpQuadCoeffDiv uint64 = 20
	// Bn256AddGas Bn256Add 计费
	Bn256AddGas uint64 = 500
	// Bn256ScalarMulGas  Bn256ScalarMul 计费
	Bn256ScalarMulGas uint64 = 40000
	// Bn256PairingBaseGas bn256Pairing 基础计费
	Bn256PairingBaseGas uint64 = 100000
	// Bn256PairingPerPointGas  bn256Pairing 按point计费（总计费等于两者相加）
	Bn256PairingPerPointGas uint64 = 80000
)
