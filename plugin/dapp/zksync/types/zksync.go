package types

import (
	"reflect"

	"github.com/33cn/chain33/types"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TyNoopAction           = 0
	TyDepositAction        = 1  //eth存款
	TyWithdrawAction       = 2  //eth取款
	TyTransferAction       = 3  //转账
	TyTransferToNewAction  = 4  //向新地址转账
	TyForceExitAction      = 5  //强制退出
	TySetPubKeyAction      = 6  //设置公钥
	TyFullExitAction       = 7  //从L1完全退出
	TySwapAction           = 8  //交换
	TyContractToTreeAction = 9  //合约账户转入叶子
	TyTreeToContractAction = 10 //叶子账户转入合约
	TyFeeAction            = 11 //手续费

	//非电路action
	TySetVerifyKeyAction = 102 //设置电路验证key
	TyCommitProofAction  = 103 //提交zk proof
	TySetVerifierAction  = 104 //设置验证者

	NameNoopAction           = "Noop"
	NameDepositAction        = "Deposit"
	NameWithdrawAction       = "Withdraw"
	NameContractToTreeAction = "ContractToTree"
	NameTreeToContractAction = "TreeToContract"
	NameTransferAction       = "Transfer"
	NameTransferToNewAction  = "TransferToNew"
	NameForceExitAction      = "ForceExit"
	NameSetPubKeyAction      = "SetPubKey"
	NameFullExitAction       = "FullExit"
	NameSwapAction           = "Swap"
	NameFeeAction            = "Fee"

	NameSetVerifyKeyAction = "SetVerifyKey"
	NameCommitProofAction  = "CommitProof"
	NameSetVerifierAction  = "SetVerifier"
)

// log类型id值
const (
	TyNoopLog           = 100
	TyDepositLog        = 101 //存款
	TyWithdrawLog       = 102 //取款
	TyTransferLog       = 103 //转账
	TyTransferToNewLog  = 104 //向新地址转账
	TyForceExitLog      = 105 //强制退出
	TySetPubKeyLog      = 106 //设置公钥
	TyFullExitLog       = 107 //从L1完全退出
	TySwapLog           = 108 //交换
	TyContractToTreeLog = 109 //合约账户转入叶子
	TyTreeToContractLog = 110 //叶子账户转入合约
	TyFeeLog            = 111 //手续费

	TySetVerifyKeyLog       = 202 //设置电路验证key
	TyCommitProofLog        = 203 //提交zk proof
	TySetVerifierLog        = 204 //设置验证者
	TySetEthPriorityQueueId = 205 //设置 eth上 priority queue id;
)

const (
	ListDESC = int32(0)
	ListASC  = int32(1)
	ListSeek = int32(2)

	Add = int32(0)
	Sub = int32(1)
)

//Zksync 执行器名称定义
const Zksync = "zksync"
const ZkManagerKey = "manager"
const ZkMimcHashSeed = "seed"
const ZkVerifierKey = "verifier"

//msg宽度
const (
	TxTypeBitWidth      = 8   //1byte
	AccountBitWidth     = 32  //4byte
	TokenBitWidth       = 16  //2byte
	AmountBitWidth      = 128 //16byte
	AddrBitWidth        = 160 //20byte
	Chain33AddrBitWidth = 256 //20byte
	PubKeyBitWidth      = 256 //32byte
	FeeAmountBitWidth   = 72  //fee op凑满one chunk=128bit，最大10byte

	PacAmountManBitWidth = 35 //amount mantissa part, 比如12340000,只取1234部分，0000用exponent表示
	PacAmountExpBitWidth = 5  //amount exponent part
	PacFeeManBitWidth    = 11 //fee mantissa part
	PacFeeExpBitWidth    = 5  //fee exponent part
	MaxExponentVal       = 32 // 2**5 by exp bit width

	ChunkBitWidth = 128               //one chunk 16 bytes
	ChunkBytes    = ChunkBitWidth / 8 //16 bytes
)

const (
	//BN254Fp=254bit,254-2 bit
	MsgFirstWidth  = 252
	MsgSecondWidth = 252
	MsgThirdWidth  = 248
	MsgWidth       = 752 //94 byte

)

//不同type chunk数量
const (
	DepositChunks       = 5
	Contract2TreeChunks = 3
	Tree2ContractChunks = 3
	TransferChunks      = 2
	Transfer2NewChunks  = 5
	WithdrawChunks      = 3
	ForceExitChunks     = 3
	FullExitChunks      = 3
	SwapChunks          = 4
	NoopChunks          = 1
	ChangePubKeyChunks  = 5
	FeeChunks           = 1
)

var (

	//定义actionMap
	actionMap = map[string]int32{
		//NameNoopAction:           TyNoopAction,
		NameDepositAction:        TyDepositAction,
		NameWithdrawAction:       TyWithdrawAction,
		NameContractToTreeAction: TyContractToTreeAction,
		NameTreeToContractAction: TyTreeToContractAction,
		NameTransferAction:       TyTransferAction,
		NameTransferToNewAction:  TyTransferToNewAction,
		NameForceExitAction:      TyForceExitAction,
		NameSetPubKeyAction:      TySetPubKeyAction,
		NameFullExitAction:       TyFullExitAction,
		NameSwapAction:           TySwapAction,
		NameSetVerifyKeyAction:   TySetVerifyKeyAction,
		NameCommitProofAction:    TyCommitProofAction,
		NameSetVerifierAction:    TySetVerifierAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		//TyNoopLog:           {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyNoopLog"},
		TyDepositLog:            {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyDepositLog"},
		TyWithdrawLog:           {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyWithdrawLog"},
		TyContractToTreeLog:     {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyContractToTreeLog"},
		TyTreeToContractLog:     {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyTreeToContractLog"},
		TyTransferLog:           {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyTransferLog"},
		TyTransferToNewLog:      {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyTransferToNewLog"},
		TyForceExitLog:          {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyForceExitLog"},
		TySetPubKeyLog:          {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TySetPubKeyLog"},
		TyFullExitLog:           {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyFullExitLog"},
		TySwapLog:               {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TySwapLog"},
		TyFeeLog:                {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyFeeLog"},
		TySetVerifyKeyLog:       {Ty: reflect.TypeOf(ReceiptSetVerifyKey{}), Name: "TySetVerifyKey"},
		TyCommitProofLog:        {Ty: reflect.TypeOf(ReceiptCommitProof{}), Name: "TyCommitProof"},
		TySetVerifierLog:        {Ty: reflect.TypeOf(ReceiptSetVerifier{}), Name: "TySetVerifierLog"},
		TySetEthPriorityQueueId: {Ty: reflect.TypeOf(ReceiptEthPriorityQueueID{}), Name: "TySetEthPriorityQueueID"},
	}

	FeeMap = map[int64]string{
		TyWithdrawAction:      "1000000",
		TyTransferAction:      "100000",
		TyTransferToNewAction: "100000",
		TyForceExitAction:     "1000000",
		TyFullExitAction:      "1000000",
		TySwapAction:          "100000",
	}
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(Zksync))
	//注册合约启用高度
	types.RegFork(Zksync, InitFork)
	types.RegExec(Zksync, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(Zksync, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(Zksync, NewType(cfg))
}

//ZksyncType ...
type ZksyncType struct {
	types.ExecTypeBase
}

//NewType ...
func NewType(cfg *types.Chain33Config) *ZksyncType {
	c := &ZksyncType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (e *ZksyncType) GetPayload() types.Message {
	return &ZksyncAction{}
}

// GetTypeMap 获取合约action的id和name信息
func (e *ZksyncType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (e *ZksyncType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}
