package executor

import (
	"fmt"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	storagetypes "github.com/33cn/plugin/plugin/dapp/storage/types"
	"github.com/gogo/protobuf/proto"
)

type StorageAction struct {
	db        dbm.KV
	txhash    []byte
	fromaddr  string
	blocktime int64
	height    int64
	index     int
}

func newStorageAction(s *storage, tx *types.Transaction, index int) *StorageAction {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &StorageAction{s.GetStateDB(), hash, fromaddr,
		s.GetBlockTime(), s.GetHeight(), index}
}
func (s *StorageAction) GetKVSet(payload proto.Message) (kvset []*types.KeyValue) {
	kvset = append(kvset, &types.KeyValue{Key: Key(common.ToHex(s.txhash)), Value: types.Encode(payload)})
	return kvset
}

//TODO 这里得数据是否存储到状态数据库中？
func (s *StorageAction) storage(payload proto.Message) (*types.Receipt, error) {
	//TODO 这里可以加具体得文本内容限制，超过指定大小的数据不容许写到状态数据库中
	var logs []*types.ReceiptLog
	kv := s.GetKVSet(payload)
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func QueryStorageByTxHash(db dbm.KV, txhash string) (*storagetypes.Storage, error) {
	data, err := db.Get(Key(txhash))
	if err != nil {
		elog.Debug("QueryStorage", "get", err)
		return nil, err
	}
	var storage storagetypes.Storage
	//decode
	err = types.Decode(data, &storage)
	if err != nil {
		elog.Debug("QueryStorage", "decode", err)
		return nil, err
	}
	return &storage, nil
}
func QueryStorage(db dbm.KV, in *storagetypes.QueryStorage) (types.Message, error) {
	if in.TxHash == "" {
		return nil, fmt.Errorf("txhash can't equail nil")
	}
	return QueryStorageByTxHash(db, in.TxHash)
}
func BatchQueryStorage(db dbm.KV, in *storagetypes.BatchQueryStorage) (types.Message, error) {
	if len(in.TxHashs) > 10 {
		return nil, fmt.Errorf("The number of batch queries is too large! the maximux is %d,but current num %d", 10, len(in.TxHashs))
	}
	var storage storagetypes.BatchReplyStorage
	for _, txhash := range in.TxHashs {
		msg, err := QueryStorageByTxHash(db, txhash)
		if err != nil {
			return msg, err
		}
		storage.Storages = append(storage.Storages, msg)
	}
	return &storage, nil
}
