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

func Init(name string, sub []byte) {
	drivers.Register(GetName(), newUnfreeze, 0)
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

func (u *Unfreeze) saveUnfreezeCreate(res *uf.ReceiptUnfreeze) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = []byte(fmt.Sprintf("mavl-unfreeze-"+"%s-"+"%s-"+"%s", res.Cur.Initiator, res.Cur.Beneficiary, res.Cur.AssetSymbol))
	kv.Value = []byte(res.Cur.UnfreezeID)
	kvs = append(kvs, kv)
	return kvs
}

func (u *Unfreeze) rollbackUnfreezeCreate(res *uf.ReceiptUnfreeze) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = []byte(fmt.Sprintf("mavl-unfreeze-"+"%s-"+"%s-"+"%s", res.Cur.Initiator, res.Cur.Beneficiary, res.Cur.AssetSymbol))
	kv.Value = []byte(res.Cur.UnfreezeID)
	kvs = append(kvs, kv)
	return kvs
}
