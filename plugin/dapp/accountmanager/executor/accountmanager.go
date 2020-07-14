package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

/*
 * 执行器相关定义
 * 重载基类相关接口
 */

var (
	//日志
	elog = log.New("module", "accountmanager.executor")
)

var driverName = et.AccountmanagerX

// Init register dapp
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newAccountmanager, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// InitExecType Init Exec Type
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Accountmanager{}))
}

//Accountmanager ...
type Accountmanager struct {
	drivers.DriverBase
}

func newAccountmanager() drivers.Driver {
	t := &Accountmanager{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName get driver name
func GetName() string {
	return newAccountmanager().GetName()
}

//GetDriverName ...
func (a *Accountmanager) GetDriverName() string {
	return driverName
}

//ExecutorOrder Exec 的时候 同时执行 ExecLocal
func (a *Accountmanager) ExecutorOrder() int64 {
	return drivers.ExecLocalSameTime
}

// CheckTx 实现自定义检验交易接口，供框架调用
func (a *Accountmanager) CheckTx(tx *types.Transaction, index int) error {
	//发送交易的时候就检查payload,做严格的参数检查
	var ama et.AccountmanagerAction
	err := types.Decode(tx.GetPayload(), &ama)
	if err != nil {
		return err
	}
	switch ama.Ty {
	case et.TyRegisterAction:
		register := ama.GetRegister()
		if a.CheckAccountIDIsExist(register.GetAccountID()) {
			return et.ErrAccountIDExist
		}
	case et.TySuperviseAction:

	case et.TyApplyAction:

	case et.TyTransferAction:

	case et.TyResetAction:

	}
	return nil
}

//CheckAccountIDIsExist ...
func (a *Accountmanager) CheckAccountIDIsExist(accountID string) bool {
	_, err := findAccountByID(a.GetLocalDB(), accountID)
	return err != types.ErrNotFound
}
