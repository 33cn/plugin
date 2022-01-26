package types

import (
	"encoding/json"
	tlog "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"reflect"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TyNoopAction           = iota + 200
	TyDepositAction        //eth存款
	TyWithdrawAction       //eth取款
	TyContractToLeafAction //合约账户转入叶子
	TyLeafToContractAction //叶子账户转入合约
	TyTransferAction       //转账
	TyTransferToNewAction  //向新地址转账
	TyForceExitAction      //强制退出
	TySetPubKeyAction      //设置公钥

	NameNoopAction           = "Noop"
	NameDepositAction        = "Deposit"
	NameWithdrawAction       = "Withdraw"
	NameContractToLeafAction = "ContractToLeaf"
	NameLeafToContractAction = "LeafToContract"
	NameTransferAction       = "Transfer"
	NameTransferToNewAction  = "TransferToNew"
	NameForceExitAction      = "ForceExit"
	NameSetPubKeyAction      = "SetPubKey"
)

// log类型id值
const (
	TyNoopLog           = iota + 200
	TyDepositLog        //存款
	TyWithdrawLog       //取款
	TyContractToLeafLog //合约账户转入叶子
	TyLeafToContractLog //叶子账户转入合约
	TyTransferLog       //转账
	TyTransferToNewLog  //向新地址转账
	TyForceExitLog      //强制退出
	TySetPubKeyLog      //设置公钥
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
)

var (

	//定义actionMap
	actionMap = map[string]int32{
		//NameNoopAction:           TyNoopAction,
		NameDepositAction:        TyDepositAction,
		NameWithdrawAction:       TyWithdrawAction,
		NameContractToLeafAction: TyContractToLeafAction,
		NameLeafToContractAction: TyLeafToContractAction,
		NameTransferAction:       TyTransferAction,
		NameTransferToNewAction:  TyTransferToNewAction,
		NameForceExitAction:      TyForceExitAction,
		NameSetPubKeyAction:      TySetPubKeyAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		//TyNoopLog:           {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyNoopLog"},
		TyDepositLog:        {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyDepositLog"},
		TyWithdrawLog:       {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyWithdrawLog"},
		TyContractToLeafLog: {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyContractToLeafLog"},
		TyLeafToContractLog: {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyLeafToContractLog"},
		TyTransferLog:       {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyTransferLog"},
		TyTransferToNewLog:  {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyTransferToNewLog"},
		TyForceExitLog:      {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TyForceExitLog"},
		TySetPubKeyLog:      {Ty: reflect.TypeOf(ZkReceiptLeaf{}), Name: "TySetPubKeyLog"},
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

// CreateTx paracross create tx by different action
func (e *ZksyncType) CreateTx(action string, msg json.RawMessage) (*types.Transaction, error) {
	data := new(ZksyncAction)
	b, err := msg.MarshalJSON()
	if err != nil {
		tlog.Error(action + " MarshalJSON  error")
		return nil, err
	}
	err = types.JSONToPB(b, data)
	if err != nil {
		tlog.Error(action + " jsontopb  error")
		return nil, err
	}
	return e.CreateTransaction(action, data)
}
