package executor

import (
	"fmt"

	"github.com/33cn/plugin/plugin/crypto/paillier"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	ety "github.com/33cn/plugin/plugin/dapp/storage/types"
	"github.com/golang/protobuf/proto"
)

//StorageAction ...
type StorageAction struct {
	api       client.QueueProtocolAPI
	db        dbm.KV
	localdb   dbm.KV
	txhash    []byte
	fromaddr  string
	blocktime int64
	height    int64
	index     int
}

func newStorageAction(s *storage, tx *types.Transaction, index int) *StorageAction {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &StorageAction{s.GetAPI(), s.GetStateDB(), s.GetLocalDB(), hash, fromaddr,
		s.GetBlockTime(), s.GetHeight(), index}
}

//GetKVSet ...
func (s *StorageAction) GetKVSet(payload proto.Message) (kvset []*types.KeyValue) {
	kvset = append(kvset, &types.KeyValue{Key: Key(common.ToHex(s.txhash)), Value: types.Encode(payload)})
	return kvset
}

//ContentStorage ...
func (s *StorageAction) ContentStorage(payload *ety.ContentOnlyNotaryStorage) (*types.Receipt, error) {

	//TODO 这里可以加具体得文本内容限制，超过指定大小的数据不容许写到状态数据库中
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := s.api.GetConfig()
	if cfg.IsDappFork(s.height, ety.StorageX, ety.ForkStorageLocalDB) {
		key := payload.Key
		op := payload.Op
		if key == "" {
			key = common.ToHex(s.txhash)
		}
		payload.Key = key
		storage, err := QueryStorageFromLocalDB(s.localdb, key)
		if op == ety.OpCreate {
			if err != types.ErrNotFound {
				return nil, ety.ErrKeyExisted
			}
		} else {
			if err == nil && storage.Ty != ety.TyContentStorageAction {
				return nil, ety.ErrStorageType
			}
			if payload.GetContent() != nil {
				content := append(storage.GetContentStorage().Content, []byte(",")...)
				payload.Content = append(content, payload.Content...)
			}
			if payload.GetValue() != "" {
				value := storage.GetContentStorage().GetValue() + "," + payload.GetValue()
				payload.Value = value
			}

		}
		stg := &ety.Storage{Value: &ety.Storage_ContentStorage{ContentStorage: payload}, Ty: ety.TyContentStorageAction}
		log := &types.ReceiptLog{Ty: ety.TyContentStorageLog, Log: types.Encode(stg)}
		logs = append(logs, log)
	} else {
		log := &types.ReceiptLog{Ty: ety.TyContentStorageLog}
		logs = append(logs, log)
		kvs = s.GetKVSet(&ety.Storage{Value: &ety.Storage_ContentStorage{ContentStorage: payload}})
	}
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

//HashStorage ...
func (s *StorageAction) HashStorage(payload *ety.HashOnlyNotaryStorage) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := s.api.GetConfig()
	if cfg.IsDappFork(s.height, ety.StorageX, ety.ForkStorageLocalDB) {
		key := payload.Key
		if key == "" {
			key = common.ToHex(s.txhash)
		}
		_, err := QueryStorageFromLocalDB(s.localdb, key)
		if err != types.ErrNotFound {
			return nil, ety.ErrKeyExisted
		}
		payload.Key = key
		stg := &ety.Storage{Value: &ety.Storage_HashStorage{HashStorage: payload}, Ty: ety.TyHashStorageAction}
		log := &types.ReceiptLog{Ty: ety.TyHashStorageLog, Log: types.Encode(stg)}
		logs = append(logs, log)
	} else {
		log := &types.ReceiptLog{Ty: ety.TyHashStorageLog}
		logs = append(logs, log)
		kvs = s.GetKVSet(&ety.Storage{Value: &ety.Storage_HashStorage{HashStorage: payload}})
	}

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

//LinkStorage ...
func (s *StorageAction) LinkStorage(payload *ety.LinkNotaryStorage) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := s.api.GetConfig()
	if cfg.IsDappFork(s.height, ety.StorageX, ety.ForkStorageLocalDB) {
		key := payload.Key
		if key == "" {
			key = common.ToHex(s.txhash)
		}
		payload.Key = key
		_, err := QueryStorageFromLocalDB(s.localdb, key)
		if err != types.ErrNotFound {
			return nil, ety.ErrKeyExisted
		}
		stg := &ety.Storage{Value: &ety.Storage_LinkStorage{LinkStorage: payload}, Ty: ety.TyLinkStorageAction}
		log := &types.ReceiptLog{Ty: ety.TyLinkStorageLog, Log: types.Encode(stg)}
		logs = append(logs, log)
	} else {
		log := &types.ReceiptLog{Ty: ety.TyLinkStorageLog}
		logs = append(logs, log)
		kvs = s.GetKVSet(&ety.Storage{Value: &ety.Storage_LinkStorage{LinkStorage: payload}})
	}
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

//EncryptStorage ...
func (s *StorageAction) EncryptStorage(payload *ety.EncryptNotaryStorage) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := s.api.GetConfig()
	if cfg.IsDappFork(s.height, ety.StorageX, ety.ForkStorageLocalDB) {
		key := payload.Key
		if key == "" {
			key = common.ToHex(s.txhash)
		}
		payload.Key = key
		_, err := QueryStorageFromLocalDB(s.localdb, key)
		if err != types.ErrNotFound {
			return nil, ety.ErrKeyExisted
		}
		stg := &ety.Storage{Value: &ety.Storage_EncryptStorage{EncryptStorage: payload}, Ty: ety.TyEncryptStorageAction}
		log := &types.ReceiptLog{Ty: ety.TyEncryptStorageLog, Log: types.Encode(stg)}
		logs = append(logs, log)
	} else {
		log := &types.ReceiptLog{Ty: ety.TyEncryptStorageLog}
		logs = append(logs, log)
		kvs = s.GetKVSet(&ety.Storage{Value: &ety.Storage_EncryptStorage{EncryptStorage: payload}})
	}
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

//EncryptShareStorage ...
func (s *StorageAction) EncryptShareStorage(payload *ety.EncryptShareNotaryStorage) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := s.api.GetConfig()
	if cfg.IsDappFork(s.height, ety.StorageX, ety.ForkStorageLocalDB) {
		key := payload.Key
		if key == "" {
			key = common.ToHex(s.txhash)
		}
		payload.Key = key
		_, err := QueryStorageFromLocalDB(s.localdb, key)
		if err != types.ErrNotFound {
			return nil, ety.ErrKeyExisted
		}
		stg := &ety.Storage{Value: &ety.Storage_EncryptShareStorage{EncryptShareStorage: payload}, Ty: ety.TyEncryptStorageAction}
		log := &types.ReceiptLog{Ty: ety.TyEncryptShareStorageLog, Log: types.Encode(stg)}
		logs = append(logs, log)
	} else {
		log := &types.ReceiptLog{Ty: ety.TyEncryptShareStorageLog}
		logs = append(logs, log)
		kvs = s.GetKVSet(&ety.Storage{Value: &ety.Storage_EncryptShareStorage{EncryptShareStorage: payload}})
	}
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

//EncryptAdd ...
func (s *StorageAction) EncryptAdd(payload *ety.EncryptNotaryAdd) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := s.api.GetConfig()

	store, err := QueryStorage(s.db, s.localdb, payload.Key)
	if err != nil {
		return nil, fmt.Errorf("EncryptAdd.QueryStorage. err:%v", err)
	}

	cipherText := store.GetEncryptStorage().EncryptContent
	res, err := paillier.CiphertextAddBytes(cipherText, payload.EncryptAdd)
	if err != nil {
		return nil, fmt.Errorf("EncryptAdd.CiphertextAddBytes. err:%v", err)
	}

	store.GetEncryptStorage().EncryptContent = res

	newStore := &ety.EncryptNotaryStorage{
		ContentHash:    store.GetEncryptStorage().ContentHash,
		EncryptContent: res,
		Nonce:          store.GetEncryptStorage().Nonce,
		Key:            store.GetEncryptStorage().Key,
		Value:          store.GetEncryptStorage().Value,
	}

	if cfg.IsDappFork(s.height, ety.StorageX, ety.ForkStorageLocalDB) {
		stg := &ety.Storage{Value: &ety.Storage_EncryptStorage{EncryptStorage: newStore}, Ty: ety.TyEncryptStorageAction}
		log := &types.ReceiptLog{Ty: ety.TyEncryptAddLog, Log: types.Encode(stg)}
		logs = append(logs, log)
	} else {
		log := &types.ReceiptLog{Ty: ety.TyEncryptAddLog}
		logs = append(logs, log)
		kvs = append(kvs, &types.KeyValue{Key: Key(payload.Key), Value: types.Encode(newStore)})
	}

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

//QueryStorageByTxHash ...
func QueryStorageByTxHash(db dbm.KV, txhash string) (*ety.Storage, error) {
	data, err := db.Get(Key(txhash))
	if err != nil {
		elog.Debug("QueryStorage", "get", err)
		return nil, err
	}
	var storage ety.Storage
	//decode
	err = types.Decode(data, &storage)
	if err != nil {
		elog.Debug("QueryStorage", "decode", err)
		return nil, err
	}
	return &storage, nil
}

//QueryStorage ...
func QueryStorage(statedb, localdb dbm.KV, txHash string) (*ety.Storage, error) {
	if txHash == "" {
		return nil, fmt.Errorf("txhash can't equail nil")
	}
	//先去localdb中查询，如果没有，则再去状态数据库中查询
	storage, err := QueryStorageFromLocalDB(localdb, txHash)
	if err != nil {
		return QueryStorageByTxHash(statedb, txHash)
	}
	return storage, nil
}

//BatchQueryStorage ...
func BatchQueryStorage(statedb, localdb dbm.KV, in *ety.BatchQueryStorage) (types.Message, error) {
	if len(in.TxHashs) > 10 {
		return nil, fmt.Errorf("The number of batch queries is too large! the maximux is %d,but current num %d", 10, len(in.TxHashs))
	}
	var storage ety.BatchReplyStorage
	for _, txhash := range in.TxHashs {
		msg, err := QueryStorage(statedb, localdb, txhash)
		if err != nil {
			return msg, err
		}
		storage.Storages = append(storage.Storages, msg)
	}
	return &storage, nil
}

//QueryStorageFromLocalDB 因为table表不支持嵌套多种数据存储结构，改成手动KV存储
func QueryStorageFromLocalDB(localdb dbm.KV, key string) (*ety.Storage, error) {
	data, err := localdb.Get(getLocalDBKey(key))
	if err != nil {
		return nil, err
	}
	var storage ety.Storage
	err = types.Decode(data, &storage)
	if err != nil {
		return nil, err
	}
	return &storage, nil
}
