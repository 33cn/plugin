package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
)

var (
	flog = log.New("module", "execs.f3d")
)

var driverName = pt.F3DX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&f3d{}))
}

func Init(name string, sub []byte) {
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

func (f *f3d) GetDriverName() string {
	return driverName
}

func (f *f3d) updateLocalDB() string {
	return driverName
}

// GetPayloadValue get payload value
func (f *f3d) GetPayloadValue() types.Message {
	return &pt.F3DAction{}
}

// GetTypeMap get TypeMap
//func (f *f3d) GetTypeMap() map[string]int32 {
//	return map[string]int32{
//
//	}
//}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (f *f3d) CheckReceiptExecOk() bool {
	return true
}
