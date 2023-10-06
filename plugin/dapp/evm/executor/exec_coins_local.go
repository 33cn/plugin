package executor

import (
	"fmt"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
)

// calcAddrKey store information on the receiving address
func calcAddrKey(addr string) []byte {
	return []byte(fmt.Sprintf("LODB-coins-Addr:%s", address.FormatAddrKey(addr)))
}

func geAddrReciverKV(addr string, reciverAmount int64) *types.KeyValue {
	reciver := &types.Int64{Data: reciverAmount}
	amountbytes := types.Encode(reciver)
	kv := &types.KeyValue{Key: calcAddrKey(addr), Value: amountbytes}
	return kv
}

func getAddrReciver(db dbm.KVDB, addr string) (int64, error) {
	reciver := types.Int64{}
	addrReciver, err := db.Get(calcAddrKey(addr))
	if err != nil && err != types.ErrNotFound {
		return 0, err
	}
	if len(addrReciver) == 0 {
		return 0, nil
	}
	err = types.Decode(addrReciver, &reciver)
	if err != nil {
		return 0, err
	}
	return reciver.Data, nil
}

func setAddrReciver(db dbm.KVDB, addr string, reciverAmount int64) error {
	kv := geAddrReciverKV(addr, reciverAmount)
	return db.Set(kv.Key, kv.Value)
}

func updateAddrReciver(cachedb dbm.KVDB, addr string, amount int64, isadd bool) (*types.KeyValue, error) {

	recv, err := getAddrReciver(cachedb, addr)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if isadd {
		recv += amount
	} else {
		recv -= amount
	}
	err = setAddrReciver(cachedb, addr, recv)
	if err != nil {
		return nil, err
	}
	//keyvalue
	return geAddrReciverKV(addr, recv), nil
}
