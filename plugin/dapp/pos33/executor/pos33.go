package executor

import (
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
	allw, err := p.getAllWeight()
	if err != nil {
		plog.Info("getAllWeight error", "error", err.Error())
	}
	if allw == nil {
		allw = &pt.Pos33AllWeight{NewWeight: make(map[int64]int64)}
	} else if allw.NewWeight == nil {
		allw.NewWeight = make(map[int64]int64)
	}
	height := p.GetHeight()
	if height > pt.Pos33SortitionSize {
		for h, v := range allw.NewWeight {
			if h+pt.Pos33SortitionSize*2-h%pt.Pos33SortitionSize <= height {
				allw.AllWeight += v
				delete(allw.NewWeight, h)
			}
		}
		allw.NewWeight[height] += int64(w)
	} else {
		allw.AllWeight = int64(w)
	}
	k := []byte(pt.KeyPos33AllWeight)
	v := types.Encode(allw)
	p.GetLocalDB().Set(k, v)
	return &types.KeyValue{Key: k, Value: v}
}

func (p *Pos33) getAllWeight() (*pt.Pos33AllWeight, error) {
	k := []byte(pt.KeyPos33AllWeight)
	val, err := p.GetLocalDB().Get(k)
	if err != nil {
		return nil, err
	}

	var w pt.Pos33AllWeight
	err = types.Decode(val, &w)
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (p *Pos33) addWeight(addr string, w int) *types.KeyValue {
	pw, err := p.getWeight(addr)
	if err != nil {
		plog.Info("getWeight error", "error", err.Error())
	}
	if pw == nil {
		pw = &pt.Pos33Weight{Weights: make(map[int64]int64)}
	} else if pw.Weights == nil {
		pw.Weights = make(map[int64]int64)
	}
	pw.Weights[p.GetHeight()] += int64(w)

	k := []byte(pt.KeyPos33WeightPrefix + addr)
	v := types.Encode(pw)
	p.GetLocalDB().Set(k, v)
	return &types.KeyValue{Key: k, Value: v}
}

func (p *Pos33) getWeight(addr string) (*pt.Pos33Weight, error) {
	val, err := p.GetLocalDB().Get([]byte(pt.KeyPos33WeightPrefix + addr))
	if err != nil {
		return nil, err
	}

	var w pt.Pos33Weight
	err = types.Decode(val, &w)
	if err != nil {
		return nil, err
	}

	return &w, nil
}
