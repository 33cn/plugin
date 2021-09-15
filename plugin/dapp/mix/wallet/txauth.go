// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"path/filepath"
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

func (p *mixPolicy) getAuthParms(req *mixTy.AuthTxReq) (*mixTy.AuthorizeCircuit, error) {
	note, err := p.getNoteInfo(req.NoteHash)
	if err != nil {
		return nil, err
	}
	if note.Status != mixTy.NoteStatus_FROZEN {
		return nil, errors.Wrapf(types.ErrNotAllow, "wrong note status=%s", note.Status.String())
	}
	if note.Secret.ReceiverKey != req.AuthorizeToAddr && note.Secret.ReturnKey != req.AuthorizeToAddr {
		return nil, errors.Wrapf(types.ErrInvalidParam, "note no match addr to AuthorizeToAddr=%s", req.AuthorizeToAddr)
	}

	//get spend privacy key
	privacyKey, err := p.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}

	if privacyKey.Privacy.PaymentKey.ReceiveKey != note.Secret.AuthorizeKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "auth pubkey from note=%s, from privacyKey=%s,for account =%s",
			note.Secret.AuthorizeKey, privacyKey.Privacy.PaymentKey.ReceiveKey, note.Account)
	}

	var input mixTy.AuthorizeCircuit

	input.NoteHash.Assign(note.NoteHash)
	input.Amount.Assign(note.Secret.Amount)
	input.ReceiverPubKey.Assign(note.Secret.ReceiverKey)
	input.ReturnPubKey.Assign(note.Secret.ReturnKey)
	input.AuthorizePubKey.Assign(note.Secret.AuthorizeKey)
	input.NoteRandom.Assign(note.Secret.NoteRandom)

	input.AuthorizePriKey.Assign(privacyKey.Privacy.PaymentKey.SpendKey)
	input.AuthorizeHash.Assign(mixTy.Byte2Str(mimcHashString([]string{note.Secret.AuthorizeKey, note.Secret.NoteRandom})))
	input.AuthorizeSpendHash.Assign(mixTy.Byte2Str(mimcHashString([]string{req.AuthorizeToAddr, note.Secret.Amount, note.Secret.NoteRandom})))

	//default auto to receiver
	if note.Secret.ReturnKey != "0" && note.Secret.ReturnKey == req.AuthorizeToAddr {
		//auth to returner
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

func (p *mixPolicy) createAuthTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	var auth mixTy.AuthTxReq
	err := types.Decode(req.Data, &auth)
	if err != nil {
		return nil, errors.Wrap(err, "decode req fail")
	}
	input, err := p.getAuthParms(&auth)
	if err != nil {
		return nil, err
	}

	if len(req.AssetExec) == 0 || len(req.AssetSymbol) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "asset exec=%s or symbol=%s not filled", req.AssetExec, req.AssetSymbol)
	}

	proofInfo, err := getZkProofKeys(mixTy.VerifyType_AUTHORIZE, auth.ZkPath, mixTy.AuthPk, input)
	if err != nil {
		return nil, errors.Wrapf(err, "getZkProofKeys note=%s", auth.NoteHash)
	}
	//verify
	vkFile := filepath.Join(auth.ZkPath, mixTy.AuthVk)
	if err := p.verifyProofOnChain(mixTy.VerifyType_AUTHORIZE, proofInfo, vkFile, req.VerifyOnChain); err != nil {
		return nil, errors.Wrapf(err, "verifyProof fail for note=%s", auth.NoteHash)
	}

	return p.getAuthTx(strings.TrimSpace(req.Title+mixTy.MixX), req.AssetExec, req.AssetSymbol, proofInfo)
}

func (p *mixPolicy) getAuthTx(execName string, exec, symbol string, proof *mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixAuthorizeAction{}
	payload.ProofInfo = proof
	payload.AssetExec = exec
	payload.AssetSymbol = symbol

	cfg := p.getWalletOperate().GetAPI().GetConfig()
	action := &mixTy.MixAction{
		Ty:    mixTy.MixActionAuth,
		Value: &mixTy.MixAction_Authorize{Authorize: payload},
	}

	tx := &types.Transaction{
		Execer:  []byte(execName),
		Payload: types.Encode(action),
		To:      address.ExecAddress(execName),
		Expire:  types.Now().Unix() + int64(300), //5 min
	}
	return types.FormatTx(cfg, execName, tx)
}
