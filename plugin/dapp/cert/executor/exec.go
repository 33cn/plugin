package executor

import (
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cert/authority"
	ct "github.com/33cn/plugin/plugin/dapp/cert/types"
)

func CertUserStoreKey(addr string) (key []byte) {
	return append([]byte("mavl-"+ct.CertX+"-"), address.FormatAddrKey(addr)...)
}

func isAdminAddr(addr string, db dbm.KV) bool {
	manageKey := types.ManageKey(ct.AdminKey)
	data, err := db.Get([]byte(manageKey))
	if err != nil {
		clog.Error("getSuperAddr", "error", err)
		return false
	}

	var item types.ConfigItem
	err = types.Decode(data, &item)
	if err != nil {
		clog.Error("isSuperAddr", "Decode", data)
		return false
	}

	for _, op := range item.GetArr().Value {
		if op == addr {
			return true
		}
	}

	return false
}

func (c *Cert) Exec_New(payload *ct.CertNew, tx *types.Transaction, index int) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	if !isAdminAddr(tx.From(), c.GetStateDB()) {
		clog.Error("Exec_New", "error", "Exec_New need admin address")
		return nil, ct.ErrPermissionDeny
	}

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (c *Cert) Exec_Update(payload *ct.CertUpdate, tx *types.Transaction, index int) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	if !isAdminAddr(tx.From(), c.GetStateDB()) {
		clog.Error("Exec_Update", "error", "Exec_Update need admin address")
		return nil, ct.ErrPermissionDeny
	}

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (c *Cert) Exec_Normal(payload *ct.CertNormal, tx *types.Transaction, index int) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	// 从proto中解码signature
	sn, err := authority.Author.GetSnFromByte(tx.Signature)
	if err != nil {
		clog.Error("Exec_Normal get sn from signature failed", "error", err)
		return nil, err
	}

	storekv := &types.KeyValue{Key: CertUserStoreKey(tx.From()), Value: sn}
	c.GetStateDB().Set(storekv.Key, storekv.Value)
	kv = append(kv, storekv)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (c *Cert) Query_CertValidSNByAddr(req *ct.ReqQueryValidCertSN) (types.Message, error) {
	sn, err := c.GetStateDB().Get(CertUserStoreKey(req.Addr))
	if err != nil {
		clog.Error("Query_CertValidSNByAddr", "error", err)
		return nil, err
	}

	return &ct.RepQueryValidCertSN{Sn: sn}, nil
}
