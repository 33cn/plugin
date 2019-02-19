package executor

import (
	"strconv"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

var plog = log.New("module", "exec.pos33")

const driverName = pt.Pos33X

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Pos33{}))
}

// Init to register pos33 driver
func Init(name string, sub []byte) {
	drivers.Register(newPos33().GetDriverName(), newPos33, 0)
}

// GetName for driver name
func GetName() string {
	return driverName
}

// Pos33 is the pos33 executor
type Pos33 struct {
	drivers.DriverBase
}

func newPos33() drivers.Driver {
	p := &Pos33{}
	p.SetChild(p)
	p.SetExecutorType(types.LoadExecutorType(driverName))
	return p
}

// GetDriverName return pos33 name
func (p *Pos33) GetDriverName() string {
	return driverName
}

func (p *Pos33) setAllWeight(w int) *types.KeyValue {
	k := []byte(pt.Pos33AllWeight)
	v := []byte(strconv.Itoa(w))
	p.GetLocalDB().Set(k, v)
	return &types.KeyValue{Key: k, Value: v}
}

func (p *Pos33) getAllWeight() int {
	k := []byte(pt.Pos33AllWeight)
	val, err := p.GetLocalDB().Get(k)
	if err != nil {
		return 0
	}

	w, err := strconv.Atoi(string(val))
	if err != nil {
		panic(err)
	}
	return w
}

func (p *Pos33) addWeight(addr string, w int) *types.KeyValue {
	w += p.getWeight(addr)
	k := []byte(pt.Pos33WeightPrefix + addr)
	v := []byte(strconv.Itoa(w))
	p.GetLocalDB().Set(k, v)
	return &types.KeyValue{Key: k, Value: v}
}

func (p *Pos33) getWeight(addr string) int {
	val, err := p.GetLocalDB().Get([]byte(pt.Pos33WeightPrefix + addr))
	if err != nil {
		return 0
	}

	w, err := strconv.Atoi(string(val))
	if err != nil {
		panic(err)
	}
	return w
}
