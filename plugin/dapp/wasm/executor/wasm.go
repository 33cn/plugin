package executor

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/system/dapp"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/perlin-network/life/exec"
)

var driverName = types2.WasmX
var log = log15.New("module", "execs."+types2.WasmX)

func Init(name string, cfg *types.Chain33Config, sub []byte) {
	if name != driverName {
		panic("system dapp can not be rename")
	}

	drivers.Register(cfg, name, newWasm, cfg.GetDappFork(name, "Enable"))
	initExecType()
}

func initExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Wasm{}))
}

type Wasm struct {
	drivers.DriverBase

	tx           *types.Transaction
	stateKVC     *dapp.KVCreator
	localCache   []*types2.LocalDataLog
	kvs          []*types.KeyValue
	receiptLogs  []*types.ReceiptLog
	customLogs   []string
	execAddr     string
	contractName string
	VMCache      map[string]*exec.VirtualMachine
	ENV          map[int]string
}

func newWasm() drivers.Driver {
	d := &Wasm{
		VMCache: make(map[string]*exec.VirtualMachine),
		ENV:     make(map[int]string),
	}
	d.SetChild(d)
	d.SetExecutorType(types.LoadExecutorType(driverName))
	return d
}

// GetName 获取执行器别名
func GetName() string {
	return newWasm().GetName()
}

func (w *Wasm) GetDriverName() string {
	return driverName
}

func (w *Wasm) ExecutorOrder() int64 {
	return drivers.ExecLocalSameTime
}
