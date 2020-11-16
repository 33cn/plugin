package types

import (
	"reflect"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

const (
	CheckNameExistsFunc        = "CheckContractNameExist"
	EstimateGasJvm             = "EstimateGasJvm"
	JvmDebug                   = "JvmDebug"
	JvmGetAbi                  = "JvmGetAbi"
	ConvertJSON2Abi            = "ConvertJson2Abi"
	Success                    = int(0)
	Exception_Fail             = int(1)
	JvmX                       = "jvm"
	UserJvmX                   = "user.jvm."
	CreateJvmContractStr       = "CreateJvmContract"
	CallJvmContractStr         = "CallJvmContract"
	UpdateJvmContractStr       = "UpdateJvmContract"
	QueryJvmContract           = "JavaContract"
	NameRegExp                 = "^[a-zA-Z0-9]+$"
	AccountOpFail              = false
	AccountOpSuccess           = true
	RetryNum                   = int(10)
	GRPCRecSize                = 5 * 30 * 1024 * 1024
	Coin_Precision       int64 = (1e4)
	MaxCodeSize                = 2 * 1024 * 1024
)

type JvmContratOpType int

const (
	CreateJvmContractAction = 1 + iota
	CallJvmContractAction
	UpdateJvmContractAction
)

// log for Jvm
const (
	// TyLogContractDataJvm 合约代码日志
	TyLogContractDataJvm = iota + 100
	// TyLogCallContractJvm 合约调用日志
	TyLogCallContractJvm
	// TyLogStateChangeItemJvm 合约状态变化的日志
	TyLogStateChangeItemJvm
	// TyLogCreateUserJvmContract 合约创建用户的日志
	TyLogCreateUserJvmContract
	// TyLogUpdateUserJvmContract 合约更新用户的日志
	TyLogUpdateUserJvmContract
	// TyLogLocalDataJvm 合约本地数据日志
	TyLogLocalDataJvm
)

// ContractLog 合约在日志，对应EVM中的Log指令，可以生成指定的日志信息
// 目前这些日志只是在合约执行完成时进行打印，没有其它用途
type ContractLog struct {
	// 合约地址
	Address address.Address
	// 对应交易哈希
	TxHash common.Hash
	// 日志序号
	Index int
	// 此合约提供的主题信息
	Topics []common.Hash
	// 日志数据
	Data []byte
}

// PrintLog 合约日志打印格式
func (log *ContractLog) PrintLog() {
	log15.Debug("!Contract Log!", "Contract address", log.Address.String(), "TxHash", log.TxHash.Bytes(), "Log Index", log.Index, "Log Topics", log.Topics)
}

var (
	actionName = map[string]int32{
		CreateJvmContractStr: CreateJvmContractAction,
		CallJvmContractStr:   CallJvmContractAction,
		UpdateJvmContractStr: UpdateJvmContractAction,
	}
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(JvmX))
	types.RegFork(JvmX, InitFork)

	types.RegExec(JvmX, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(JvmX, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(JvmX, NewType(cfg))
}

// JvmType EVM类型定义
type JvmType struct {
	types.ExecTypeBase
}

// NewType 新建EVM类型对象
func NewType(cfg *types.Chain33Config) *JvmType {
	c := &JvmType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取消息负载结构
func (jvm *JvmType) GetPayload() types.Message {
	return &JVMContractAction{}
}

// ActionName 获取ActionName
func (jvm JvmType) ActionName(tx *types.Transaction) string {
	// 这个需要通过合约交易目标地址来判断Action
	// 如果目标地址为空，或为jvm的固定合约地址，则为创建合约，否则为调用合约
	//if strings.EqualFold(tx.To, address.ExecAddress(types.ExecName(JvmX))) {
	//	return "createJvmContract"
	//}
	//return "callJvmContract"
	var action JVMContractAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return "unknown"
	}

	switch action.Value.(type) {
	case *JVMContractAction_CreateJvmContract:
		return "createJvmContract"
	case *JVMContractAction_CallJvmContract:
		return "callJvmContract"
	case *JVMContractAction_UpdateJvmContract:
		return "updateJvmContract"
	}

	return "unknown"
}

// GetTypeMap 获取类型映射
func (jvm *JvmType) GetTypeMap() map[string]int32 {
	return actionName
}

// GetRealToAddr 获取实际地址
func (jvm JvmType) GetRealToAddr(tx *types.Transaction) string {
	var action JVMContractAction
	err := types.Decode(tx.Payload, &action)
	if err == nil {
		return tx.To
	}
	return ""
}

// Amount 获取金额
func (jvm JvmType) Amount(tx *types.Transaction) (int64, error) {
	return 0, nil
}

// GetLogMap 获取日志类型映射
func (jvm *JvmType) GetLogMap() map[int64]*types.LogInfo {
	logInfo := map[int64]*types.LogInfo{
		TyLogContractDataJvm:       {Ty: reflect.TypeOf(LogJVMContractData{}), Name: "LogContractDataJvm"},
		TyLogCallContractJvm:       {Ty: reflect.TypeOf(ReceiptJVMContract{}), Name: "LogCallContractJvm"},
		TyLogStateChangeItemJvm:    {Ty: reflect.TypeOf(JVMStateChangeItem{}), Name: "LogStateChangeItemJvm"},
		TyLogCreateUserJvmContract: {Ty: reflect.TypeOf(ReceiptJVMContract{}), Name: "LogCreateUserJvmContract"},
		TyLogUpdateUserJvmContract: {Ty: reflect.TypeOf(ReceiptJVMContract{}), Name: "LogUpdateUserJvmContract"},
		TyLogLocalDataJvm:          {Ty: reflect.TypeOf(ReceiptLocalData{}), Name: "LogLocalDataJvm"},
	}
	return logInfo
}
