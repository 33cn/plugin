package executor

import (
	"fmt"

	log "github.com/inconshreveable/log15"
	uf "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
	drivers "gitlab.33.cn/chain33/chain33/system/dapp"
	"gitlab.33.cn/chain33/chain33/types"
)

var uflog = log.New("module", "execs.unfreeze")

var driverName = uf.UnfreezeX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Unfreeze{}))
}

func Init(name string) {
	drivers.Register(GetName(), newGame, 0)
}

type Unfreeze struct {
	drivers.DriverBase
}

func newUnfreeze() drivers.Driver {
	t := &Unfreeze{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

func GetName() string {
	return newUnfreeze().GetName()
}

func (u *Unfreeze) GetDriverName() string {
	return driverName
}

func (u *Unfreeze) GetPayloadValue() types.Message {
	return &uf.UnfreezeAction{}
}

func (u *Unfreeze) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create":    uf.UnfreezeActionCreate,
		"Withdraw":  uf.UnfreezeActionWithdraw,
		"Terminate": uf.UnfreezeActionTerminate,
	}
}
