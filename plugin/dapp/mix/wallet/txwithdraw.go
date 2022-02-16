// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"path/filepath"
	"strings"

	"github.com/consensys/gnark-crypto/ecc"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

func (p *mixPolicy) getWithdrawParams(exec, symbol, noteHash string) (*mixTy.WithdrawCircuit, error) {
	note, err := p.getNoteInfo(noteHash)
	if err != nil {
		return nil, err
	}
	if note.Status != mixTy.NoteStatus_VALID && note.Status != mixTy.NoteStatus_UNFROZEN {
		return nil, errors.Wrapf(types.ErrNotAllow, "wrong note status=%s", note.Status.String())
	}

	if note.Secret.AssetExec != exec || note.Secret.AssetSymbol != symbol {
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,sym=%s not equal req's exec=%s,symbol=%s",
			note.Secret.AssetExec, note.Secret.AssetSymbol, exec, symbol)
	}

	var input mixTy.WithdrawCircuit
	input.NullifierHash.Assign(note.Nullifier)
	input.NoteHash.Assign(note.NoteHash)
	input.AuthorizeSpendHash.Assign(note.AuthorizeSpendHash)

	input.Amount.Assign(note.Secret.Amount)
	input.ReceiverPubKey.Assign(note.Secret.ReceiverKey)
	input.ReturnPubKey.Assign(note.Secret.ReturnKey)
	input.AuthorizePubKey.Assign(note.Secret.AuthorizeKey)
	input.NoteRandom.Assign(note.Secret.NoteRandom)

	if len(note.AuthorizeSpendHash) > LENNULLKEY {
		input.AuthorizeFlag.Assign("1")
	} else {
		input.AuthorizeFlag.Assign("0")
	}

	//get spend privacy key
	privacyKey, err := p.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}
	input.SpendPriKey.Assign(privacyKey.Privacy.PaymentKey.SpendKey)
	if privacyKey.Privacy.PaymentKey.ReceiveKey == note.Secret.ReturnKey {
		input.SpendFlag.Assign("0")
	} else {
		input.SpendFlag.Assign("1")
	}

	//get tree path
	treeProof, err := p.getTreeProof(note.Secret.AssetExec, note.Secret.AssetSymbol, note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "getTreeProof for hash=%s", note.NoteHash)
	}
	input.TreeRootHash.Assign(treeProof.TreeRootHash)
	updateTreePath(&input, treeProof)

	return &input, nil

}

func (p *mixPolicy) createWithdrawTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	var withdraw mixTy.WithdrawTxReq
	err := types.Decode(req.Data, &withdraw)
	if err != nil {
		return nil, errors.Wrap(err, "decode req fail")
	}

	if len(req.AssetExec) == 0 || len(req.AssetSymbol) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "asset exec=%s or symbol=%s not filled", req.AssetExec, req.AssetSymbol)
	}

	if withdraw.TotalAmount <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "totalAmount=%d", withdraw.TotalAmount)
	}
	notes := strings.Split(withdraw.NoteHashs, ",")
	if len(notes) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "noteHashs=%s", withdraw.NoteHashs)
	}

	var proofs []*mixTy.ZkProofInfo

	var sum uint64
	for _, note := range notes {
		input, err := p.getWithdrawParams(req.AssetExec, req.AssetSymbol, note)
		if err != nil {
			return nil, errors.Wrapf(err, "getWithdrawParams note=%s", note)
		}
		proofInfo, err := getZkProofKeys(mixTy.VerifyType_WITHDRAW, withdraw.ZkPath, mixTy.WithdrawPk, input)
		if err != nil {
			return nil, errors.Wrapf(err, "getZkProofKeys note=%s", note)
		}
		//verify
		vkFile := filepath.Join(withdraw.ZkPath, mixTy.WithdrawVk)
		if err := p.verifyProofOnChain(mixTy.VerifyType_WITHDRAW, proofInfo, vkFile, req.VerifyOnChain); err != nil {
			return nil, errors.Wrapf(err, "verifyProof fail for note=%s", note)
		}

		v := input.Amount.GetWitnessValue(ecc.BN254)
		sum += v.Uint64()
		proofs = append(proofs, proofInfo)
	}

	//不设计找零操作，可以全部提取回来后再存入，提取的找零一定是本账户的，不利于隐私，而且提取操作功能不够单一
	if sum != withdraw.TotalAmount {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount not match req=%d,note.sum=%d", withdraw.TotalAmount, sum)
	}

	return p.getWithdrawTx(strings.TrimSpace(req.Title+mixTy.MixX), req.AssetExec, req.AssetSymbol, withdraw.TotalAmount, proofs)

}

func (p *mixPolicy) getWithdrawTx(execName, assetExec, assetSymbol string, amount uint64, proofs []*mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixWithdrawAction{}
	payload.AssetExec = assetExec
	payload.AssetSymbol = assetSymbol
	payload.Amount = amount
	payload.Proofs = proofs

	cfg := p.getWalletOperate().GetAPI().GetConfig()
	action := &mixTy.MixAction{
		Ty:    mixTy.MixActionWithdraw,
		Value: &mixTy.MixAction_Withdraw{Withdraw: payload},
	}

	tx := &types.Transaction{
		Execer:  []byte(execName),
		Payload: types.Encode(action),
		To:      address.ExecAddress(execName),
		Expire:  types.Now().Unix() + int64(300), //5 min
	}
	return types.FormatTx(cfg, execName, tx)
}
