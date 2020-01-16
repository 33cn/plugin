package executor

import (
	"github.com/33cn/chain33/types"
	storagetypes "github.com/33cn/plugin/plugin/dapp/storage/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (s *storage) Exec_ContentStorage(payload *storagetypes.ContentOnlyNotaryStorage, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newStorageAction(s, tx, index)
	return action.ContentStorage(&storagetypes.Storage{Value: &storagetypes.Storage_ContentStorage{ContentStorage: payload}})
}

func (s *storage) Exec_HashStorage(payload *storagetypes.HashOnlyNotaryStorage, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newStorageAction(s, tx, index)
	return action.HashStorage(&storagetypes.Storage{Value: &storagetypes.Storage_HashStorage{HashStorage: payload}})
}

func (s *storage) Exec_LinkStorage(payload *storagetypes.LinkNotaryStorage, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newStorageAction(s, tx, index)
	return action.LinkStorage(&storagetypes.Storage{Value: &storagetypes.Storage_LinkStorage{LinkStorage: payload}})
}

func (s *storage) Exec_EncryptStorage(payload *storagetypes.EncryptNotaryStorage, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newStorageAction(s, tx, index)
	return action.EncryptStorage(&storagetypes.Storage{Value: &storagetypes.Storage_EncryptStorage{EncryptStorage: payload}})
}

func (s *storage) Exec_EncryptShareStorage(payload *storagetypes.EncryptShareNotaryStorage, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newStorageAction(s, tx, index)
	return action.EncryptShareStorage(&storagetypes.Storage{Value: &storagetypes.Storage_EncryptShareStorage{EncryptShareStorage: payload}})
}
