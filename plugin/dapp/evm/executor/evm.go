// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"fmt"
	log "github.com/33cn/chain33/common/log/log15"
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
	evmDebugInited = false
	// 执行器地址, 格式由evm配置的地址驱动指定
	evmExecAddress       = ""
	evmExecFormatAddress = ""
	driverName           = evmtypes.ExecutorName
	elog                 = log.New("module", "evm.executor")
)

type subConfig struct {
	// AddressDriver address driver name, support btc/eth
	AddressDriver string `json:"addressDriver"`
	// PreCompileAddr key: preContractorAddress  value: real contract information
	PreCompile   runtime.TokenContract `json:"preCompile,omitempty"`
	UpgradeUrl   string                `json:"upgradeUrl,omitempty"`
	NonceUpGrade bool                  `json:"nonceUpGrade,omitempty"`
}

func initEvmSubConfig(sub []byte, evmEnableHeight int64) {
	var subCfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subCfg)
	}
	addressType, err := address.GetDriverType(subCfg.AddressDriver)

	if err != nil && subCfg.AddressDriver != "" {
		panic("GetDriverType:" + err.Error())
	}

	// get default if not config
	if subCfg.AddressDriver == "" {
		addressType = address.GetDefaultAddressID()
	}
	// 加载, 确保在evm使能高度前, eth地址驱动已使能
	driver, err := address.LoadDriver(addressType, evmEnableHeight)
	if err != nil {
		panic(fmt.Sprintf("address driver must enable before %d", evmEnableHeight))
	}
	common.InitEvmAddressDriver(driver)

	runtime.CustomizePrecompiledContracts[common.HexToAddress(runtime.TokenPrecompileAddr)] = runtime.NewTokenPrecompile(&runtime.TokenContract{SuperManager: subCfg.PreCompile.SuperManager})

}

// Init 初始化本合约对象
func Init(name string, cfg *types.Chain33Config, sub []byte) {

	enableHeight := cfg.GetDappFork(driverName, evmtypes.EVMEnable)
	initEvmSubConfig(sub, enableHeight)
	driverName = name
	drivers.Register(cfg, driverName, newEVMDriver, enableHeight)
	// 格式化为配置地址格式
	evmExecAddress = common.StringToAddress(address.ExecAddress(cfg.ExecName(name))).String()
	evmExecFormatAddress = address.ToLower(evmExecAddress)
	log.Info("evmInit", "execAddr", evmExecAddress, "formatAddr", evmExecFormatAddress)
	// 初始化硬分叉数据
	state.InitForkData()
	state.InitEvmCheckData()
	InitExecType()
}

// InitExecType Init Exec Type
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&EVMExecutor{}))
}

// GetName 返回本合约名称
func GetName() string {
	return newEVMDriver().GetName()
}

func newEVMDriver() drivers.Driver {
	evm := NewEVMExecutor()
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
	//exec.vmCfg.Tracer = runtime.NewJSONLogger(os.Stdout)
	exec.vmCfg.Tracer = runtime.NewMarkdownLogger(
		&runtime.LogConfig{
			DisableMemory:     false,
			DisableStack:      false,
			DisableStorage:    false,
			DisableReturnData: false,
			Debug:             true,
			Limit:             0,
		},
		os.Stdout,
	)

	exec.SetChild(exec)
	exec.SetExecutorType(types.LoadExecutorType(driverName))
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
	cfg := evm.GetAPI().GetConfig()
	if cfg.IsFork(evm.GetHeight(), "ForkLocalDBAccess") {
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
	cfg := evm.GetAPI().GetConfig()
	exec := cfg.GetParaExec(tx.Execer)
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
	cfg := evm.GetAPI().GetConfig()
	exec := cfg.GetParaExec(othertx.Execer)
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
	cfg := evm.GetAPI().GetConfig()
	return common.NewAddress(cfg, txHash)
}

// createContractAddress creates an ethereum address given the bytes and the nonce
func (evm *EVMExecutor) createContractAddress(b common.Address, txHash []byte) common.Address {
	return common.NewContractAddress(b, txHash)
}

// createContractAddress creates an ethereum address given the bytes and the nonce
func (evm *EVMExecutor) createEvmContractAddress(b common.Address, nonce uint64) common.Address {
	return common.NewEvmContractAddress(b, nonce)
}

// CheckTx 校验交易
func (evm *EVMExecutor) CheckTx(tx *types.Transaction, index int) error {
	if evm.GetAPI().GetConfig().IsPara() {
		return nil
	}

	if tx == nil {
		return fmt.Errorf("tx empty")
	}

	return state.ProcessCheck(evm.GetMainHeight(), tx.Hash())
}

// GetActionName 获取运行状态名
func (evm *EVMExecutor) GetActionName(tx *types.Transaction) string {
	cfg := evm.GetAPI().GetConfig()
	if bytes.Equal(tx.Execer, []byte(cfg.ExecName(evmtypes.ExecutorName))) {
		return cfg.ExecName(evmtypes.ExecutorName)
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
func (evm *EVMExecutor) NewEVMContext(msg *common.Message, txHash []byte) runtime.Context {
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
		TxHash:      txHash,
	}
}
