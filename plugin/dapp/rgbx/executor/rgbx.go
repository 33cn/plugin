package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	rgbxtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"sync"
)

/*
 * 执行器相关定义
 * 重载基类相关接口
 */

var (
	//日志
	elog        = log.New("module", "rgbx.executor")
	cfgInitOnce sync.Once
)

var driverName = rgbxtypes.RgbxX

type cfg struct {
	CommitAddress string `json:"commitAddress"`
}

var rgbxCfg = cfg{}

func initCfg(sub []byte) {

	cfgInitOnce.Do(func() {
		types.MustDecode(sub, &rgbxCfg)
	})
}

// Init register dapp
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	initCfg(sub)
	drivers.Register(cfg, GetName(), newRgbx, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// InitExecType Init Exec Type
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&rgbx{}))
}

type rgbx struct {
	drivers.DriverBase
}

func newRgbx() drivers.Driver {
	t := &rgbx{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName get driver name
func GetName() string {
	return newRgbx().GetName()
}

func (r *rgbx) GetDriverName() string {
	return driverName
}

func (r *rgbx) ExecutorOrder() int64 {
	return drivers.ExecLocalSameTime
}
