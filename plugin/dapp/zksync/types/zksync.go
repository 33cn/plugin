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
	TyContractToTreeAction = 3  //合约账户转入叶子
	TyTreeToContractAction = 4  //叶子账户转入合约
	TyTransferAction       = 5  //转账
	TyTransferToNewAction  = 6  //向新地址转账
	TyForceExitAction      = 7  //强制退出
	TySetPubKeyAction      = 8  //设置公钥
	TyFullExitAction       = 9  //从L1完全退出
	TySwapAction           = 10 //交换

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

	NameSetVerifyKeyAction = "SetVerifyKey"
	NameCommitProofAction  = "CommitProof"
	NameSetVerifierAction  = "SetVerifier"
)

// log类型id值
const (
	TyNoopLog           = 0
	TyDepositLog        = 1  //存款
	TyWithdrawLog       = 2  //取款
	TyContractToTreeLog = 3  //合约账户转入叶子
	TyTreeToContractLog = 4  //叶子账户转入合约
	TyTransferLog       = 5  //转账
	TyTransferToNewLog  = 6  //向新地址转账
	TyForceExitLog      = 7  //强制退出
	TySetPubKeyLog      = 8  //设置公钥
	TyFullExitLog       = 9  //从L1完全退出
	TySwapLog           = 10 //交换

	TySetVerifyKeyLog = 102 //设置电路验证key
	TyCommitProofLog  = 103 //提交zk proof
	TySetVerifierLog  = 104 //设置验证者
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
	TxTypeBitWidth  = 8   //1byte
	AccountBitWidth = 32  //4byte
	TokenBitWidth   = 16  //2byte
	AmountBitWidth  = 128 //16byte
	AddrBitWidth    = 160 //20byte
	PubKeyBitWidth  = 256 //32byte
)

const (
	MsgFirstWidth  = 252
	MsgSecondWidth = 252
	MsgThirdWidth  = 248
	MsgWidth       = 752 //32byte
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
		TyDepositLog:        {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyDepositLog"},
		TyWithdrawLog:       {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyWithdrawLog"},
		TyContractToTreeLog: {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyContractToTreeLog"},
		TyTreeToContractLog: {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyTreeToContractLog"},
		TyTransferLog:       {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyTransferLog"},
		TyTransferToNewLog:  {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyTransferToNewLog"},
		TyForceExitLog:      {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyForceExitLog"},
		TySetPubKeyLog:      {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TySetPubKeyLog"},
		TyFullExitLog:       {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TyFullExitLog"},
		TySwapLog:           {Ty: reflect.TypeOf(ZkReceiptLog{}), Name: "TySwapLog"},
		TySetVerifyKeyLog:   {Ty: reflect.TypeOf(ReceiptSetVerifyKey{}), Name: "TySetVerifyKey"},
		TyCommitProofLog:    {Ty: reflect.TypeOf(ReceiptCommitProof{}), Name: "TyCommitProof"},
		TySetVerifierLog:    {Ty: reflect.TypeOf(ReceiptSetVerifier{}), Name: "TySetVerifierLog"},
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
