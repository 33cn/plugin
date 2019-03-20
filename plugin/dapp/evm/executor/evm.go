// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"math/big"

	"os"

	"reflect"

	"github.com/33cn/chain33/common/address"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/runtime"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/state"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

var (
	evmDebug = false

	// EvmAddress 本合约地址
	EvmAddress = address.ExecAddress(types.ExecName(evmtypes.ExecutorName))
)

var driverName = evmtypes.ExecutorName

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&EVMExecutor{}))
}

// Init 初始化本合约对象
func Init(name string, sub []byte) {
	driverName = name
	drivers.Register(driverName, newEVMDriver, types.GetDappFork(driverName, evmtypes.EVMEnable))
	EvmAddress = address.ExecAddress(types.ExecName(name))
	// 初始化硬分叉数据
	state.InitForkData()
}

// GetName 返回本合约名称
func GetName() string {
	return newEVMDriver().GetName()
}

func newEVMDriver() drivers.Driver {
	evm := NewEVMExecutor()
	evm.vmCfg.Debug = evmDebug
	return evm
}

// EVMExecutor EVM执行器结构
type EVMExecutor struct {
	drivers.DriverBase
	vmCfg    *runtime.Config
	mStateDB *state.MemoryStateDB
}

// NewEVMExecutor 新创建执行器对象
func NewEVMExecutor() *EVMExecutor {
	exec := &EVMExecutor{}

	exec.vmCfg = &runtime.Config{}
	exec.vmCfg.Tracer = runtime.NewJSONLogger(os.Stdout)

	exec.SetChild(exec)
	return exec
}

// GetFuncMap 获取方法列表
func (evm *EVMExecutor) GetFuncMap() map[string]reflect.Method {
	ety := types.LoadExecutorType(driverName)
	return ety.GetExecFuncMap()
}

// GetDriverName 获取本合约驱动名称
func (evm *EVMExecutor) GetDriverName() string {
	return evmtypes.ExecutorName
}

// ExecutorOrder 设置localdb的EnableRead
func (evm *EVMExecutor) ExecutorOrder() int64 {
	if types.IsFork(evm.GetHeight(), "ForkLocalDBAccess") {
		return drivers.ExecLocalSameTime
	}
	return evm.DriverBase.ExecutorOrder()
}

// Allow 允许哪些交易在本命执行器执行
func (evm *EVMExecutor) Allow(tx *types.Transaction, index int) error {
	err := evm.DriverBase.Allow(tx, index)
	if err == nil {
		return nil
	}
	//增加新的规则:
	//主链: user.evm.xxx  执行 evm 合约
	//平行链: user.p.guodun.user.evm.xxx 执行 evm 合约
	exec := types.GetParaExec(tx.Execer)
	if evm.AllowIsUserDot2(exec) {
		return nil
	}
	return types.ErrNotAllow
}

// IsFriend 是否允许对应的KEY
func (evm *EVMExecutor) IsFriend(myexec, writekey []byte, othertx *types.Transaction) bool {
	if othertx == nil {
		return false
	}
	exec := types.GetParaExec(othertx.Execer)
	if exec == nil || len(bytes.TrimSpace(exec)) == 0 {
		return false
	}
	if bytes.HasPrefix(exec, evmtypes.UserPrefix) || bytes.Equal(exec, evmtypes.ExecerEvm) {
		if bytes.HasPrefix(writekey, []byte("mavl-evm-")) {
			return true
		}
	}
	return false
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (evm *EVMExecutor) CheckReceiptExecOk() bool {
	return true
}

// 生成一个新的合约对象地址
func (evm *EVMExecutor) getNewAddr(txHash []byte) common.Address {
	return common.NewAddress(txHash)
}

// CheckTx 校验交易
func (evm *EVMExecutor) CheckTx(tx *types.Transaction, index int) error {
	return nil
}

// GetActionName 获取运行状态名
func (evm *EVMExecutor) GetActionName(tx *types.Transaction) string {
	if bytes.Equal(tx.Execer, []byte(types.ExecName(evmtypes.ExecutorName))) {
		return types.ExecName(evmtypes.ExecutorName)
	}
	return tx.ActionName()
}

// GetMStateDB 获取内部状态数据库
func (evm *EVMExecutor) GetMStateDB() *state.MemoryStateDB {
	return evm.mStateDB
}

// GetVMConfig 获取VM配置
func (evm *EVMExecutor) GetVMConfig() *runtime.Config {
	return evm.vmCfg
}

// NewEVMContext 构造一个新的EVM上下文对象
func (evm *EVMExecutor) NewEVMContext(msg *common.Message) runtime.Context {
	return runtime.Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(evm.GetAPI()),
		Origin:      msg.From(),
		Coinbase:    nil,
		BlockNumber: new(big.Int).SetInt64(evm.GetHeight()),
		Time:        new(big.Int).SetInt64(evm.GetBlockTime()),
		Difficulty:  new(big.Int).SetUint64(evm.GetDifficulty()),
		GasLimit:    msg.GasLimit(),
		GasPrice:    msg.GasPrice(),
	}
}
