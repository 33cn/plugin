// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"github.com/33cn/chain33/common"

	commondb "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/pkg/errors"
)

func (p *mixPolicy) execAutoLocalMix(tx *types.Transaction, receiptData *types.ReceiptData, index int, height int64) (*types.LocalDBSet, error) {
	set, err := p.execLocalMix(tx, receiptData, height, int64(index))
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = p.store.AddRollbackKV(tx, tx.Execer, set.KV)
	return dbSet, nil
}

func (p *mixPolicy) execLocalMix(tx *types.Transaction, receiptData *types.ReceiptData, height, index int64) (*types.LocalDBSet, error) {
	if receiptData.Ty != types.ExecOk {
		return nil, types.ErrInvalidParam
	}

	if !p.store.getPrivacyEnable() {
		return nil, nil
	}
	return p.processMixTx(tx, height, index)

}

func (p *mixPolicy) processMixTx(tx *types.Transaction, height, index int64) (*types.LocalDBSet, error) {

	var v mixTy.MixAction
	err := types.Decode(tx.Payload, &v)
	if err != nil {
		bizlog.Error("execLocalMix decode", "hash", tx.Hash(), "err", err)
		return nil, types.ErrInvalidParam
	}

	table := NewMixTable(commondb.NewKVDB(p.getWalletOperate().GetDBStore()))
	switch v.GetTy() {
	//deposit 匹配newcommits，属于自己的存到数据库
	case mixTy.MixActionDeposit:
		p.processDeposit(v.GetDeposit(), dapp.HeightIndexStr(height, index), table)

	//根据withdraw nullifier hash 更新数据状态为USED
	case mixTy.MixActionWithdraw:
		var nulls []string
		for _, m := range v.GetWithdraw().SpendCommits {
			out, err := mixTy.DecodePubInput(mixTy.VerifyType_WITHDRAW, m.PublicInput)
			if err != nil {
				bizlog.Error("processWithdraw decode", "pubInput", m.PublicInput)
				continue
			}
			nulls = append(nulls, out.(*mixTy.WithdrawPublicInput).NullifierHash)
		}
		p.processNullifiers(nulls, table)

	//nullifier hash更新为used， newcommit解密存储
	case mixTy.MixActionTransfer:
		p.processTransfer(v.GetTransfer(), dapp.HeightIndexStr(height, index), table)
	//查看本地authSpend hash是否hit, 是则更新为OPEN状态
	case mixTy.MixActionAuth:
		p.processAuth(v.GetAuthorize(), table)

	}

	kvs, err := table.Save()
	if err != nil {
		bizlog.Error("execLocalMix table save", "hash", tx.Hash(), "err", err)
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil

}

func (p *mixPolicy) processDeposit(deposit *mixTy.MixDepositAction, heightIndex string, table *table.Table) {
	for _, m := range deposit.NewCommits {
		data, err := mixTy.DecodePubInput(mixTy.VerifyType_DEPOSIT, m.PublicInput)
		if err != nil {
			bizlog.Error("processDeposit decode", "pubInput", m.PublicInput)
			continue
		}
		input := data.(*mixTy.DepositPublicInput)
		p.processSecretGroup(input.NoteHash, m.Group, heightIndex, table)

	}
}

func (p *mixPolicy) processTransfer(transfer *mixTy.MixTransferAction, heightIndex string, table *table.Table) {
	var nulls []string
	for _, m := range transfer.Input {
		data, err := mixTy.DecodePubInput(mixTy.VerifyType_TRANSFERINPUT, m.PublicInput)
		if err != nil {
			bizlog.Error("processTransfer.input decode", "pubInput", m.PublicInput)
			continue
		}
		input := data.(*mixTy.TransferInputPublicInput)
		nulls = append(nulls, input.NullifierHash)
	}
	p.processNullifiers(nulls, table)

	for _, m := range transfer.Output {
		data, err := mixTy.DecodePubInput(mixTy.VerifyType_TRANSFEROUTPUT, m.PublicInput)
		if err != nil {
			bizlog.Error("processTransfer.output decode", "pubInput", m.PublicInput)
			continue
		}
		input := data.(*mixTy.TransferOutputPublicInput)
		p.processSecretGroup(input.NoteHash, input.DhSecrets, heightIndex, table)
	}

}

func (p *mixPolicy) processAuth(auth *mixTy.MixAuthorizeAction, table *table.Table) {
	for _, m := range auth.AuthCommits {
		out, err := mixTy.DecodePubInput(mixTy.VerifyType_AUTHORIZE, m.PublicInput)
		if err != nil {
			bizlog.Error("processAuth decode", "pubInput", m.PublicInput)
			continue
		}
		input := out.(*mixTy.AuthorizePublicInput)
		updateAuthSpend(table, input.AuthorizeSpendHash)
	}
}

func (p *mixPolicy) processNullifiers(nulls []string, table *table.Table) {

	for _, n := range nulls {
		err := updateNullifier(table, n)
		if err != nil {
			bizlog.Error("processNullifiers", "nullifier", n, "err", err)
		}
	}

}

func updateNullifier(ldb *table.Table, nullifier string) error {
	xs, err := ldb.ListIndex("nullifier", []byte(nullifier), nil, 1, 0)
	if err != nil || len(xs) != 1 {
		bizlog.Error("updateNullifier update query List failed", "key", nullifier, "err", err, "len", len(xs))
		return nil
	}
	u, ok := xs[0].Data.(*mixTy.WalletDbMixInfo)
	if !ok {
		bizlog.Error("updateNullifier update decode failed", "data", xs[0].Data)
		return nil

	}
	u.Info.Status = mixTy.NoteStatus_INVALID
	return ldb.Update([]byte(u.TxIndex), u)
}

func updateAuthSpend(ldb *table.Table, authSpend string) error {
	xs, err := ldb.ListIndex("authSpendHash", []byte(authSpend), nil, 1, 0)
	if err != nil || len(xs) != 1 {
		bizlog.Error("updateAuthSpend update query List failed", "key", authSpend, "err", err, "len", len(xs))
		return nil
	}
	u, ok := xs[0].Data.(*mixTy.WalletDbMixInfo)
	if !ok {
		bizlog.Error("updateAuthSpend update decode failed", "data", xs[0].Data)
		return nil

	}
	u.Info.Status = mixTy.NoteStatus_VALID
	return ldb.Update([]byte(u.TxIndex), u)
}

func (e *mixPolicy) listMixInfos(req *mixTy.WalletMixIndexReq) (types.Message, error) {
	if req == nil {
		return nil, types.ErrInvalidParam
	}
	localDb := e.getWalletOperate().GetDBStore()
	query := NewMixTable(localDb).GetQuery(commondb.NewKVDB(localDb))
	var primary []byte

	indexName := ""
	if len(req.NoteHash) > 0 {
		indexName = "noteHash"
	} else if len(req.Nullifier) > 0 {
		indexName = "nullifier"
	} else if len(req.AuthSpendHash) > 0 {
		indexName = "authSpendHash"
	} else if len(req.Account) > 0 {
		indexName = "account"
	} else if req.Status > 0 {
		indexName = "status"
	}

	cur := &MixRow{
		WalletDbMixInfo: &mixTy.WalletDbMixInfo{Info: &mixTy.WalletIndexInfo{
			NoteHash:      req.NoteHash,
			Nullifier:     req.Nullifier,
			AuthSpendHash: req.AuthSpendHash,
			Account:       req.Account,
			Status:        mixTy.NoteStatus(req.Status),
		}},
	}

	prefix, err := cur.Get(indexName)
	if err != nil {
		bizlog.Error("listMixInfos Get", "indexName", indexName, "err", err)
		return nil, err
	}
	rows, err := query.ListIndex(indexName, prefix, primary, req.Count, req.Direction)
	if err != nil {
		bizlog.Error("listMixInfos query failed", "indexName", indexName, "prefix", prefix, "key", string(primary), "err", err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, types.ErrNotFound
	}
	var rep mixTy.WalletIndexResp
	for _, row := range rows {
		r, ok := row.Data.(*mixTy.WalletDbMixInfo)
		if !ok {
			bizlog.Error("listMixInfos", "err", "bad row type")
			return nil, types.ErrDecode
		}
		rep.Datas = append(rep.Datas, r.Info)
	}
	return &rep, nil
}

func (e *mixPolicy) execAutoDelLocal(tx *types.Transaction) (*types.LocalDBSet, error) {
	kvs, err := e.store.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (p *mixPolicy) addTable(info *mixTy.WalletIndexInfo, heightIndex string, table *table.Table) {
	r := &mixTy.WalletDbMixInfo{
		Info:    info,
		TxIndex: heightIndex + info.NoteHash,
	}
	err := table.Replace(r)
	if err != nil {
		bizlog.Error("addTable", "notehash", info.NoteHash, "err", err)
	}
}

func (p *mixPolicy) processSecretGroup(noteHash string, secretGroup *mixTy.DHSecretGroup, heightIndex string, table *table.Table) {
	if secretGroup == nil {
		bizlog.Info("noteHash secretGroup null", "noteHash", noteHash)
		return
	}

	privacyKeys, err := p.getWalletPrivacyKeys()
	if err != nil || privacyKeys == nil {
		bizlog.Error("processSecretGroup.getPrivacyPairs", "notehash", noteHash, "err", err)
		return
	}

	//可能自己账户里面既有spender,也有returner 或authorize,都要解一遍
	info, err := p.decodeSecret(noteHash, secretGroup.Payment, privacyKeys)
	if err != nil {
		bizlog.Error("processSecretGroup.spender", "err", err)
	}
	if info != nil {
		p.addTable(info, heightIndex, table)
	}

	info, err = p.decodeSecret(noteHash, secretGroup.Returner, privacyKeys)
	if err != nil {
		bizlog.Error("processSecretGroup.Returner", "err", err)
	}
	if info != nil {
		p.addTable(info, heightIndex, table)
	}

	info, err = p.decodeSecret(noteHash, secretGroup.Authorize, privacyKeys)
	if err != nil {
		bizlog.Error("processSecretGroup.Authorize", "err", err)
	}
	if info != nil {
		p.addTable(info, heightIndex, table)
	}
}

func (p *mixPolicy) decodeSecret(noteHash string, dhSecret *mixTy.DHSecret, privacyKeys []*mixTy.WalletAddrPrivacy) (*mixTy.WalletIndexInfo, error) {
	if dhSecret == nil {
		return nil, errors.Wrapf(types.ErrEmpty, "secret nil for notehash=%s", noteHash)
	}

	tempPubKey := &mixTy.PubKey{X: dhSecret.Epk.X, Y: dhSecret.Epk.Y}

	for _, key := range privacyKeys {
		cryptData, err := common.FromHex(dhSecret.Secret)
		if err != nil {
			return nil, errors.Wrapf(err, "decode for notehash=%s,crypt=%s", noteHash, dhSecret.Secret)
		}
		decryptData, err := decryptData(key.Privacy.ShareSecretKey.PrivKey, tempPubKey, cryptData)
		if err != nil {
			return nil, errors.Wrapf(err, "decrypt for notehash=%s,crypt=%s", noteHash, dhSecret.Secret)
		}
		var rawData mixTy.SecretData
		err = types.Decode(decryptData, &rawData)
		if err != nil {
			bizlog.Info("processSecret.DecrypterPrivkey", "err", err)
			continue
		}
		if rawData.PaymentPubKey == key.Privacy.PaymentKey.PayKey ||
			rawData.ReturnPubKey == key.Privacy.PaymentKey.PayKey ||
			rawData.AuthorizePubKey == key.Privacy.PaymentKey.PayKey {
			//decrypted, save database
			var info mixTy.WalletIndexInfo
			info.NoteHash = noteHash
			info.Nullifier = getFrString(mimcHashString([]string{rawData.NoteRandom}))
			//如果自己是spender,则记录有关spenderAuthHash,如果是returner，则记录returnerAuthHash
			//如果授权为spenderAuthHash，则spender更新本地为OPEN，returner侧仍为FROZEN，花费后，两端都变为USED
			//如果授权为returnerAuthHash，则returner更新本地为OPEN，spender侧仍为FROZEN，
			if rawData.PaymentPubKey == key.Privacy.PaymentKey.PayKey {
				info.AuthSpendHash = getFrString(mimcHashString([]string{rawData.PaymentPubKey, rawData.Amount, rawData.NoteRandom}))
			} else if rawData.ReturnPubKey == key.Privacy.PaymentKey.PayKey {
				info.IsReturner = true
				info.AuthSpendHash = getFrString(mimcHashString([]string{rawData.ReturnPubKey, rawData.Amount, rawData.NoteRandom}))
			}
			info.Status = mixTy.NoteStatus_VALID
			if len(rawData.AuthorizePubKey) > 0 {
				info.Status = mixTy.NoteStatus_FROZEN
			}
			//账户地址
			info.Account = key.Addr
			info.Secret = &rawData
			return &info, nil

		}
	}
	return nil, nil
}
