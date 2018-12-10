package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
)

var (
	ptylog = log.New("module", "execs.f3d")
)

var driverName = pt.F3DX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&f3d{}))
}

func Init(name string, sub []byte) {
	var cfg pt.Config
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	pt.SetConfig(&cfg)
	drivers.Register(GetName(), newf3d, types.GetDappFork(driverName, "Enable"))
}

type f3d struct {
	drivers.DriverBase
}

func newf3d() drivers.Driver {
	t := &f3d{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

func GetName() string {
	return newf3d().GetName()
}

func (u *f3d) GetDriverName() string {
	return driverName
}
