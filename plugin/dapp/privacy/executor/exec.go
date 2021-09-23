// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/account"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/privacy/types"
)

// Exec_Public2Privacy execute public to privacy
func (p *privacy) Exec_Public2Privacy(payload *ty.Public2Privacy, tx *types.Transaction, index int) (*types.Receipt, error) {

	accDB, err := p.createAccountDB(payload.GetAssetExec(), payload.GetTokenname())
	if err != nil {
		privacylog.Error("Exec_pub2priv_newAccountDB", "exec", payload.GetAssetExec(),
			"symbol", payload.GetTokenname(), "err", err)
		return nil, err
	}
	txhashstr := hex.EncodeToString(tx.Hash())
	from := tx.From()
	receipt, err := accDB.ExecWithdraw(address.ExecAddress(string(tx.Execer)), from, payload.Amount)
	if err != nil {
		privacylog.Error("PrivacyTrading Exec", "txhash", txhashstr, "ExecWithdraw error ", err)
		return nil, err
	}

	txhash := common.ToHex(tx.Hash())
	output := payload.GetOutput().GetKeyoutput()
	//因为只有包含当前交易的block被执行完成之后，同步到相应的钱包之后，
	//才能将相应的utxo作为input，进行支付，所以此处不需要进行将KV设置到
	//executor中的临时数据库中，只需要将kv返回给blockchain就行
	//即：一个块中产生的UTXO是不能够在同一个高度进行支付的
	for index, keyOutput := range output {
		key := CalcPrivacyOutputKey(payload.AssetExec, payload.Tokenname, keyOutput.Amount, txhash, index)
		value := types.Encode(keyOutput)
		receipt.KV = append(receipt.KV, &types.KeyValue{Key: key, Value: value})
	}
	privacylog.Debug("testkey", "output", payload.GetOutput().Keyoutput)
	receiptLogs := p.buildPrivacyReceiptLog(payload.GetAssetExec(), payload.GetTokenname(), payload.GetOutput())
	execlog := &types.ReceiptLog{Ty: ty.TyLogPrivacyOutput, Log: types.Encode(receiptLogs)}
	receipt.Logs = append(receipt.Logs, execlog)

	//////////////////debug code begin///////////////
	privacylog.Debug("PrivacyTrading Exec", "ActionPublic2Privacy txhash", txhashstr, "receipt is", receipt)
	//////////////////debug code end///////////////

	return receipt, nil
}

// Exec_Privacy2Privacy execute privacy to privacy transaction
func (p *privacy) Exec_Privacy2Privacy(payload *ty.Privacy2Privacy, tx *types.Transaction, index int) (*types.Receipt, error) {

	txhashstr := hex.EncodeToString(tx.Hash())
	receipt := &types.Receipt{KV: make([]*types.KeyValue, 0)}
	privacyInput := payload.Input
	for _, keyInput := range privacyInput.Keyinput {
		value := []byte{keyImageSpentAlready}
		key := calcPrivacyKeyImageKey(payload.AssetExec, payload.Tokenname, keyInput.KeyImage)
		stateDB := p.GetStateDB()
		stateDB.Set(key, value)
		receipt.KV = append(receipt.KV, &types.KeyValue{Key: key, Value: value})
	}

	execlog := &types.ReceiptLog{Ty: ty.TyLogPrivacyInput, Log: types.Encode(payload.GetInput())}
	receipt.Logs = append(receipt.Logs, execlog)

	txhash := common.ToHex(tx.Hash())
	output := payload.GetOutput().GetKeyoutput()
	for index, keyOutput := range output {
		key := CalcPrivacyOutputKey(payload.AssetExec, payload.Tokenname, keyOutput.Amount, txhash, index)
		value := types.Encode(keyOutput)
		receipt.KV = append(receipt.KV, &types.KeyValue{Key: key, Value: value})
	}
	receiptLogs := p.buildPrivacyReceiptLog(payload.GetAssetExec(), payload.GetTokenname(), payload.GetOutput())
	execlog = &types.ReceiptLog{Ty: ty.TyLogPrivacyOutput, Log: types.Encode(receiptLogs)}
	receipt.Logs = append(receipt.Logs, execlog)

	receipt.Ty = types.ExecOk

	//////////////////debug code begin///////////////
	privacylog.Debug("PrivacyTrading Exec", "ActionPrivacy2Privacy txhash", txhashstr, "receipt is", receipt)
	//////////////////debug code end///////////////
	return receipt, nil
}

// Exec_Privacy2Public execute privacy to public transaction
func (p *privacy) Exec_Privacy2Public(payload *ty.Privacy2Public, tx *types.Transaction, index int) (*types.Receipt, error) {
	accDB, err := p.createAccountDB(payload.GetAssetExec(), payload.GetTokenname())
	if err != nil {
		privacylog.Error("Exec_pub2priv_newAccountDB", "exec", payload.GetAssetExec(),
			"symbol", payload.GetTokenname(), "err", err)
		return nil, err
	}
	txhashstr := hex.EncodeToString(tx.Hash())
	receipt, err := accDB.ExecDeposit(payload.To, address.ExecAddress(string(tx.Execer)), payload.Amount)
	if err != nil {
		privacylog.Error("PrivacyTrading Exec", "ActionPrivacy2Public txhash", txhashstr, "ExecDeposit error ", err)
		return nil, err
	}
	privacyInput := payload.Input
	for _, keyInput := range privacyInput.Keyinput {
		value := []byte{keyImageSpentAlready}
		key := calcPrivacyKeyImageKey(payload.AssetExec, payload.Tokenname, keyInput.KeyImage)
		stateDB := p.GetStateDB()
		stateDB.Set(key, value)
		receipt.KV = append(receipt.KV, &types.KeyValue{Key: key, Value: value})
	}

	execlog := &types.ReceiptLog{Ty: ty.TyLogPrivacyInput, Log: types.Encode(payload.GetInput())}
	receipt.Logs = append(receipt.Logs, execlog)

	txhash := common.ToHex(tx.Hash())
	output := payload.GetOutput().GetKeyoutput()
	for index, keyOutput := range output {
		key := CalcPrivacyOutputKey(payload.AssetExec, payload.Tokenname, keyOutput.Amount, txhash, index)
		value := types.Encode(keyOutput)
		receipt.KV = append(receipt.KV, &types.KeyValue{Key: key, Value: value})
	}

	receiptLog := p.buildPrivacyReceiptLog(payload.GetAssetExec(), payload.GetTokenname(), payload.GetOutput())
	execlog = &types.ReceiptLog{Ty: ty.TyLogPrivacyOutput, Log: types.Encode(receiptLog)}
	receipt.Logs = append(receipt.Logs, execlog)

	receipt.Ty = types.ExecOk

	//////////////////debug code begin///////////////
	privacylog.Debug("PrivacyTrading Exec", "ActionPrivacy2Privacy txhash", txhashstr, "receipt is", receipt)
	//////////////////debug code end///////////////
	return receipt, nil
}

func (p *privacy) createAccountDB(exec, symbol string) (*account.DB, error) {
	cfg := p.GetAPI().GetConfig()
	if exec == "" || exec == cfg.GetCoinExec() {
		return p.GetCoinsAccount(), nil
	}
	return account.NewAccountDB(cfg, exec, symbol, p.GetStateDB())
}

func (p *privacy) buildPrivacyReceiptLog(assetExec, assetSymbol string, output *ty.PrivacyOutput) *ty.ReceiptPrivacyOutput {
	if assetExec == "" {
		assetExec = p.GetAPI().GetConfig().GetCoinExec()
	}
	receipt := &ty.ReceiptPrivacyOutput{
		AssetExec:   assetExec,
		AssetSymbol: assetSymbol,
		Keyoutput:   output.Keyoutput,
	}

	return receipt
}
