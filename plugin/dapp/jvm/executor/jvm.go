package executor

import (
	"bytes"
	"sync/atomic"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/jvm/executor/state"
	jvmTypes "github.com/33cn/plugin/plugin/dapp/jvm/types"
	lru "github.com/hashicorp/golang-lru"
)

type stopWithError struct {
	occurred bool
	info     error
}

// JVMExecutor 执行器结构
type JVMExecutor struct {
	drivers.DriverBase
	mStateDB      *state.MemoryStateDB
	tx            *types.Transaction
	txHash        string
	contract      string
	txIndex       int
	forceStopInfo stopWithError
	queryChan     chan QueryResult
}

type QueryResult struct {
	exceptionOccurred bool
	info              []string
}

var (
	log              = log15.New("module", "execs.jvm")
	jvmsCached       *lru.Cache
	jvmsCacheCreated = int32(0)
	jdkPath          string
)

func initExecType() {
	ety := types.LoadExecutorType(jvmTypes.JvmX)
	ety.InitFuncList(types.ListMethod(&JVMExecutor{}))
}

// Init register function
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	drivers.Register(cfg, GetName(), newJVM, cfg.GetDappFork(jvmTypes.JvmX, "Enable"))
	initExecType()

	conf := types.ConfSub(cfg, jvmTypes.JvmX)
	jdkPath = conf.GStr("jdkPath")
	if "" == jdkPath {
		panic("JDK path is not configured")
	}
	log.Info("jvm::Init", "JDK path is configured to:", jdkPath)
	initJvm(cfg)
}

func newJVM() drivers.Driver {
	return newJVMDriver()
}

// GetName get name for execname
func GetName() string {
	return newJVM().GetName()
}

func newJVMDriver() drivers.Driver {
	jvm := NewJVMExecutor()
	return jvm
}

// NewJVMExecutor new a jvm executor
func NewJVMExecutor() *JVMExecutor {
	exec := &JVMExecutor{}
	exec.SetChild(exec)
	exec.SetExecutorType(types.LoadExecutorType(jvmTypes.JvmX))
	atomic.LoadInt32(&jvmsCacheCreated)
	if int32(Bool_TRUE) != atomic.LoadInt32(&jvmsCacheCreated) {
		var err error
		jvmsCached, err = lru.New(1000)
		if nil != err {
			panic("Failed to new lru for caching jvms due to:" + err.Error())
		}
		atomic.StoreInt32(&jvmsCacheCreated, int32(Bool_TRUE))
	}
	return exec
}

// GetDriverName 获取driver 名称
func (jvm *JVMExecutor) GetDriverName() string {
	return jvmTypes.JvmX
}

// ExecutorOrder 设置localdb的EnableRead
func (jvm *JVMExecutor) ExecutorOrder() int64 {
	cfg := jvm.GetAPI().GetConfig()
	if cfg.IsFork(jvm.GetHeight(), "ForkLocalDBAccess") {
		return drivers.ExecLocalSameTime
	}
	return jvm.DriverBase.ExecutorOrder()
}

// Allow 允许哪些交易在本命执行器执行
func (jvm *JVMExecutor) Allow(tx *types.Transaction, index int) error {
	err := jvm.DriverBase.Allow(tx, index)
	if err == nil {
		return nil
	}
	//增加新的规则:
	//主链: user.jvm.xxx  执行 jvm用户自定义 合约
	//平行链: user.p.guodun.user.jvm.xxx 执行 jvm用户自定义合约
	exec := jvm.GetAPI().GetConfig().GetParaExec(tx.Execer)
	if jvm.AllowIsUserDot2(exec) {
		return nil
	}
	return types.ErrNotAllow
}

func (jvm *JVMExecutor) prepareExecContext(tx *types.Transaction, index int) {
	paraExector := string(jvm.GetAPI().GetConfig().GetParaExec(tx.Execer))
	if jvm.mStateDB == nil {
		log.Info("prepareExecContext", "executorName", paraExector)
		jvm.mStateDB = state.NewMemoryStateDB(paraExector, jvm.GetStateDB(), jvm.GetLocalDB(), jvm.GetCoinsAccount(), jvm.GetHeight())
	}

	jvm.tx = tx
	jvm.txHash = common.ToHex(tx.Hash())
	jvm.txIndex = index
}

func (jvm *JVMExecutor) prepareQueryContext(executorName []byte) {
	if jvm.mStateDB == nil {
		log.Info("prepareQueryContext", "executorName", string(jvm.GetAPI().GetConfig().GetParaExec(executorName)))
		jvm.mStateDB = state.NewMemoryStateDB(string(jvm.GetAPI().GetConfig().GetParaExec(executorName)), jvm.GetStateDB(), jvm.GetLocalDB(), jvm.GetCoinsAccount(), jvm.GetHeight())
	}
}

// GenerateExecReceipt generate exec receipt
func (jvm *JVMExecutor) GenerateExecReceipt(execName, caller, contractAddr string, opType jvmTypes.JvmContratOpType) (*types.Receipt, error) {
	curVer := jvm.mStateDB.GetLastSnapshot()

	// 打印合约中生成的日志
	jvm.mStateDB.PrintLogs()

	if curVer == nil {
		return nil, nil
	}
	// 从状态机中获取数据变更和变更日志
	data, logs := jvm.mStateDB.GetChangedData(curVer.GetID(), opType)
	contractReceipt := &jvmTypes.ReceiptJVMContract{Caller: caller, ContractName: execName, ContractAddr: contractAddr}

	runLog := &types.ReceiptLog{
		Ty:  jvmTypes.TyLogCallContractJvm,
		Log: types.Encode(contractReceipt)}
	if opType == jvmTypes.CreateJvmContractAction {
		runLog.Ty = jvmTypes.TyLogCreateUserJvmContract
	} else if opType == jvmTypes.UpdateJvmContractAction {
		runLog.Ty = jvmTypes.TyLogUpdateUserJvmContract
	}

	logs = append(logs, runLog)
	logs = append(logs, jvm.mStateDB.GetReceiptLogs(contractAddr)...)

	receipt := &types.Receipt{Ty: types.ExecOk, KV: data, Logs: logs}

	//jvm.collectJvmTxLog(jvm.tx, contractReceipt, receipt)

	return receipt, nil
}

//func (jvm *JVMExecutor) collectJvmTxLog(tx *types.Transaction, cr *jvmTypes.ReceiptJVMContract, receipt *types.Receipt) {
//	log.Debug("jvm collect begin")
//	log.Debug("Tx info", "txHash", common.ToHex(tx.Hash()), "height", jvm.GetHeight())
//	log.Debug("ReceiptJVMContract", "data", fmt.Sprintf("caller=%v, name=%v, addr=%v", cr.Caller, cr.ContractName, cr.ContractAddr))
//	log.Debug("receipt data", "type", receipt.Ty)
//	for _, kv := range receipt.KV {
//		log.Debug("KeyValue", "key", common.ToHex(kv.Key), "value", common.ToHex(kv.Value))
//	}
//	for _, kv := range receipt.Logs {
//		log.Debug("ReceiptLog", "Type", kv.Ty, "log", common.ToHex(kv.Log))
//	}
//	log.Debug("jvm collect end")
//}

// 检查合约地址是否存在，此操作不会改变任何状态，所以可以直接从statedb查询
func (jvm *JVMExecutor) checkContractNameExists(req *jvmTypes.CheckJVMContractNameReq) (types.Message, error) {
	contractName := req.JvmContractName
	if len(contractName) == 0 {
		return nil, jvmTypes.ErrNullContractName
	}

	if !bytes.Contains([]byte(contractName), []byte(jvmTypes.UserJvmX)) {
		contractName = jvmTypes.UserJvmX + contractName
	}
	exists := jvm.mStateDB.Exist(address.ExecAddress(jvm.GetAPI().GetConfig().ExecName(contractName)))
	ret := &jvmTypes.CheckJVMAddrResp{ExistAlready: exists}
	return ret, nil
}

func (jvm *JVMExecutor) GetContractAddr() string {
	if jvm.tx != nil {
		return address.GetExecAddress(string(jvm.tx.Execer)).String()
	}
	return address.GetExecAddress(jvm.contract).String()
}
