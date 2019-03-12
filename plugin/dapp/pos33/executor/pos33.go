package executor

import (
	"fmt"
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
	k := []byte(pt.KeyPos33AllWeight)
	v := []byte(strconv.Itoa(w))
	p.GetLocalDB().Set(k, v)
	return &types.KeyValue{Key: k, Value: v}
}

// GetAllWeight get all weight deposit ycc
func (p *Pos33) GetAllWeight() int {
	return p.getAllWeight()
}

func (p *Pos33) getAllWeight() int {
	k := []byte(pt.KeyPos33AllWeight)
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
	k := []byte(pt.KeyPos33WeightPrefix + addr)
	v := []byte(strconv.Itoa(w))
	p.GetLocalDB().Set(k, v)
	return &types.KeyValue{Key: k, Value: v}
}

// GetWeight return addr depoist weight of vote
func (p *Pos33) GetWeight(addr string) int {
	return p.getWeight(addr)
}

// GetWeight get all weight deposit ycc by addr
func (p *Pos33) getWeight(addr string) int {
	val, err := p.GetLocalDB().Get([]byte(pt.KeyPos33WeightPrefix + addr))
	if err != nil {
		return 0
	}

	w, err := strconv.Atoi(string(val))
	if err != nil {
		panic(err)
	}
	return w
}

// func (p *Pos33) getCommittee(key string) (*pt.Pos33Committee, error) {
// 	val, err := p.GetLocalDB().Get([]byte(key))
// 	if err != nil {
// 		return nil, err
// 	}

// 	var comm pt.Pos33Committee
// 	err = types.Decode(val, &comm)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &comm, nil
// }

// func (p *Pos33) setCommittee(key string, comm *pt.Pos33Committee) *types.KeyValue {
// 	value := types.Encode(comm)
// 	return p.GetLocalDB().Set([]byte(key), value)
// }

func (p *Pos33) getBlockSeed(height int64) ([]byte, error) {
	val, err := p.GetLocalDB().Get([]byte(fmt.Sprintf("%s%d", pt.KeyPos33RewordPrefix, height)))
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (p *Pos33) getElecteLocal(height int64) (*pt.Pos33ElecteLocal, error) {
	val, err := p.GetLocalDB().Get([]byte(fmt.Sprintf("%s%d", pt.KeyPos33ElectePrefix, height)))
	if err != nil {
		return nil, err
	}

	var e pt.Pos33ElecteLocal
	err = types.Decode(val, &e)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

// func (p *Pos33) setElecteLocal(e *pt.Pos33ElecteLocal) *types.KeyValue {
// 	value := types.Encode(e)
// 	return &types.KeyValue{[]byte(keyPos33Electe), value}
// }
