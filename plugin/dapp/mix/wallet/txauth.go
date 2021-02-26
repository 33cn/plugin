// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

type AuthorizeInput struct {
	//public
	TreeRootHash       string `tag:"public"`
	AuthorizeHash      string `tag:"public"`
	AuthorizeSpendHash string `tag:"public"`

	//secret
	ReceiverPubKey  string `tag:"secret"`
	ReturnPubKey    string `tag:"secret"`
	AuthorizePubKey string `tag:"secret"`
	AuthorizePriKey string `tag:"secret"`
	NoteRandom      string `tag:"secret"`

	Amount    string `tag:"secret"`
	SpendFlag string `tag:"secret"`
	NoteHash  string `tag:"secret"`

	//tree path info
	Path0 string `tag:"secret"`
	Path1 string `tag:"secret"`
	Path2 string `tag:"secret"`
	Path3 string `tag:"secret"`
	Path4 string `tag:"secret"`
	Path5 string `tag:"secret"`
	Path6 string `tag:"secret"`
	Path7 string `tag:"secret"`
	Path8 string `tag:"secret"`
	Path9 string `tag:"secret"`

	Helper0 string `tag:"secret"`
	Helper1 string `tag:"secret"`
	Helper2 string `tag:"secret"`
	Helper3 string `tag:"secret"`
	Helper4 string `tag:"secret"`
	Helper5 string `tag:"secret"`
	Helper6 string `tag:"secret"`
	Helper7 string `tag:"secret"`
	Helper8 string `tag:"secret"`
	Helper9 string `tag:"secret"`

	Valid0 string `tag:"secret"`
	Valid1 string `tag:"secret"`
	Valid2 string `tag:"secret"`
	Valid3 string `tag:"secret"`
	Valid4 string `tag:"secret"`
	Valid5 string `tag:"secret"`
	Valid6 string `tag:"secret"`
	Valid7 string `tag:"secret"`
	Valid8 string `tag:"secret"`
	Valid9 string `tag:"secret"`
}

func (p *mixPolicy) getAuthParms(req *mixTy.AuthTxReq) (*AuthorizeInput, error) {
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

	var input AuthorizeInput
	initTreePath(&input)

	input.NoteHash = note.NoteHash
	input.Amount = note.Secret.Amount
	input.ReceiverPubKey = note.Secret.ReceiverKey
	input.ReturnPubKey = note.Secret.ReturnKey
	input.AuthorizePubKey = note.Secret.AuthorizeKey
	input.NoteRandom = note.Secret.NoteRandom

	input.AuthorizePriKey = privacyKey.Privacy.PaymentKey.SpendKey
	input.AuthorizeHash = mixTy.Byte2Str(mimcHashString([]string{input.AuthorizePubKey, note.Secret.NoteRandom}))
	input.AuthorizeSpendHash = mixTy.Byte2Str(mimcHashString([]string{req.AuthorizeToAddr, note.Secret.Amount, note.Secret.NoteRandom}))

	//default auto to receiver
	input.SpendFlag = "1"
	if input.ReturnPubKey != "0" && input.ReturnPubKey == req.AuthorizeToAddr {
		//auth to returner
		input.SpendFlag = "0"
	}

	//get tree path
	treeProof, err := p.getTreeProof(note.NoteHash)
	if err != nil {
		return nil, errors.Wrapf(err, "getTreeProof for hash=%s", note.NoteHash)
	}
	input.TreeRootHash = treeProof.TreeRootHash
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

	proofInfo, err := getZkProofKeys(auth.ZkPath+mixTy.AuthCircuit, auth.ZkPath+mixTy.AuthPk, *input, req.Privacy)
	if err != nil {
		return nil, errors.Wrapf(err, "getZkProofKeys note=%s", auth.NoteHash)
	}
	//verify
	if err := p.verifyProofOnChain(mixTy.VerifyType_AUTHORIZE, proofInfo, auth.ZkPath+mixTy.AuthVk, req.Verify); err != nil {
		return nil, errors.Wrapf(err, "verifyProof fail for note=%s", auth.NoteHash)
	}

	return p.getAuthTx(strings.TrimSpace(req.Title+mixTy.MixX), proofInfo)
}

func (p *mixPolicy) getAuthTx(execName string, proof *mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixAuthorizeAction{}
	payload.Proof = proof

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
