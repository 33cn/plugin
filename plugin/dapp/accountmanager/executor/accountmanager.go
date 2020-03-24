package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	accountmanagertypes "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

/*
 * 执行器相关定义
 * 重载基类相关接口
 */

var (
	//日志
	elog = log.New("module", "accountmanager.executor")
)

var driverName = accountmanagertypes.AccountmanagerX

// Init register dapp
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newAccountmanager, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// InitExecType Init Exec Type
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&accountmanager{}))
}

type accountmanager struct {
	drivers.DriverBase
}

func newAccountmanager() drivers.Driver {
	t := &accountmanager{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName get driver name
func GetName() string {
	return newAccountmanager().GetName()
}

func (a *accountmanager) GetDriverName() string {
	return driverName
}

//ExecutorOrder Exec 的时候 同时执行 ExecLocal
func (e *accountmanager) ExecutorOrder() int64 {
	return drivers.ExecLocalSameTime
}

// CheckTx 实现自定义检验交易接口，供框架调用
func (a *accountmanager) CheckTx(tx *types.Transaction, index int) error {
	// implement code
	return nil
}
