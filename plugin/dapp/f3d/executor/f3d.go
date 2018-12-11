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

func (f *f3d) updateLocalDB(r *pt.ReceiptF3D) (kvs []*types.KeyValue) {
	switch r.Action {
	case pt.F3dActionStart:
		start := &pt.F3DStartRound{
			Round: r.Round,
		}
		kvs = append(kvs, &types.KeyValue{Key: calcF3dStartRound(r.Round), Value: types.Encode(start)})
	case pt.F3dActionBuy:
		addrInfo := &pt.AddrInfo{
			Addr:     r.Addr,
			Round:    r.Round,
			BuyCount: r.BuyCount,
		}
		kvs = append(kvs, &types.KeyValue{Key: calcF3dAddrRound(r.Round, r.Addr), Value: types.Encode(addrInfo)})
		buyRecord := &pt.F3DBuyRecord{
			Round: r.Round,
			Addr:  r.Addr,
			Index: r.Index,
		}
		kvs = append(kvs, &types.KeyValue{Key: calcF3dBuyRound(r.Round, r.Addr, r.Index), Value: types.Encode(buyRecord)})
	case pt.F3dActionDraw:
		draw := &pt.F3DDrawRound{
			Round: r.Round,
		}
		kvs = append(kvs, &types.KeyValue{Key: calcF3dDrawRound(r.Round), Value: types.Encode(draw)})
	}
	return kvs
}

func (f *f3d) rollbackLocalDB(r *pt.ReceiptF3D) (kvs []*types.KeyValue) {
	switch r.Action {
	case pt.F3dActionStart:
		kvs = append(kvs, &types.KeyValue{Key: calcF3dStartRound(r.Round), Value: nil})
	case pt.F3dActionBuy:
		if r.BuyCount <= 1 {
			kvs = append(kvs, &types.KeyValue{Key: calcF3dAddrRound(r.Round, r.Addr), Value: nil})
		} else {
			addrInfo := &pt.AddrInfo{
				Addr:     r.Addr,
				Round:    r.Round,
				BuyCount: r.BuyCount - 1,
			}
			kvs = append(kvs, &types.KeyValue{Key: calcF3dAddrRound(r.Round, r.Addr), Value: types.Encode(addrInfo)})
		}
		kvs = append(kvs, &types.KeyValue{Key: calcF3dBuyRound(r.Round, r.Addr, r.Index), Value: nil})
	case pt.F3dActionDraw:
		kvs = append(kvs, &types.KeyValue{Key: calcF3dDrawRound(r.Round), Value: nil})
	}
	return kvs
	return kvs
}

// GetPayloadValue get payload value
func (f *f3d) GetPayloadValue() types.Message {
	return &pt.F3DAction{}
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (f *f3d) CheckReceiptExecOk() bool {
	return true
}
