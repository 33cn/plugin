// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"strconv"
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

type WithdrawInput struct {
	//public
	TreeRootHash       string `tag:"public"`
	AuthorizeSpendHash string `tag:"public"`
	NullifierHash      string `tag:"public"`
	Amount             string `tag:"public"`

	//secret
	ReceiverPubKey  string `tag:"secret"`
	ReturnPubKey    string `tag:"secret"`
	AuthorizePubKey string `tag:"secret"`
	NoteRandom      string `tag:"secret"`
	SpendPriKey     string `tag:"secret"`
	SpendFlag       string `tag:"secret"`
	AuthorizeFlag   string `tag:"secret"`

	//tree path info
	NoteHash string `tag:"secret"`
	Path0    string `tag:"secret"`
	Path1    string `tag:"secret"`
	Path2    string `tag:"secret"`
	Path3    string `tag:"secret"`
	Path4    string `tag:"secret"`
	Path5    string `tag:"secret"`
	Path6    string `tag:"secret"`
	Path7    string `tag:"secret"`
	Path8    string `tag:"secret"`
	Path9    string `tag:"secret"`

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

func (p *mixPolicy) getWithdrawParams(noteHash string) (*WithdrawInput, error) {
	note, err := p.getNoteInfo(noteHash)
	if err != nil {
		return nil, err
	}
	if note.Status != mixTy.NoteStatus_VALID && note.Status != mixTy.NoteStatus_UNFROZEN {
		return nil, errors.Wrapf(types.ErrNotAllow, "wrong note status=%s", note.Status.String())
	}

	var input WithdrawInput
	initTreePath(&input)
	input.NullifierHash = note.Nullifier
	input.NoteHash = note.NoteHash
	input.AuthorizeSpendHash = note.AuthorizeSpendHash

	input.Amount = note.Secret.Amount
	input.ReceiverPubKey = note.Secret.ReceiverKey
	input.ReturnPubKey = note.Secret.ReturnKey
	input.AuthorizePubKey = note.Secret.AuthorizeKey
	input.NoteRandom = note.Secret.NoteRandom

	input.AuthorizeFlag = "0"
	if len(input.AuthorizeSpendHash) > LENNULLKEY {
		input.AuthorizeFlag = "1"
	}

	//get spend privacy key
	privacyKey, err := p.getAccountPrivacyKey(note.Account)
	if err != nil {
		return nil, errors.Wrapf(err, "getAccountPrivacyKey addr=%s", note.Account)
	}
	input.SpendPriKey = privacyKey.Privacy.PaymentKey.SpendKey
	input.SpendFlag = "1"
	if privacyKey.Privacy.PaymentKey.ReceiveKey == input.ReturnPubKey {
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

func (p *mixPolicy) createWithdrawTx(req *mixTy.CreateRawTxReq) (*types.Transaction, error) {
	var withdraw mixTy.WithdrawTxReq
	err := types.Decode(req.Data, &withdraw)
	if err != nil {
		return nil, errors.Wrap(err, "decode req fail")
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
		input, err := p.getWithdrawParams(note)
		if err != nil {
			return nil, errors.Wrapf(err, "getWithdrawParams note=%s", note)
		}

		proofInfo, err := getZkProofKeys(withdraw.ZkPath+mixTy.WithdrawCircuit, withdraw.ZkPath+mixTy.WithdrawPk, *input)
		if err != nil {
			return nil, errors.Wrapf(err, "getZkProofKeys note=%s", note)
		}
		//verify
		if err := p.verifyProofOnChain(mixTy.VerifyType_WITHDRAW, proofInfo, withdraw.ZkPath+mixTy.WithdrawVk); err != nil {
			return nil, errors.Wrapf(err, "verifyProof fail for note=%s", note)
		}

		v, err := strconv.Atoi(input.Amount)
		if err != nil {
			return nil, errors.Wrapf(err, "atoi fail for note=%s,amount=%s", note, input.Amount)
		}
		sum += uint64(v)
		proofs = append(proofs, proofInfo)
	}

	//不设计找零操作，可以全部提取回来后再存入，提取的找零一定是本账户的，不利于隐私，而且提取操作功能不够单一
	if sum != withdraw.TotalAmount {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount not match req=%d,note.sum=%d", withdraw.TotalAmount, sum)
	}

	return p.getWithdrawTx(strings.TrimSpace(req.Title+mixTy.MixX), withdraw.TotalAmount, proofs)

}

func (p *mixPolicy) getWithdrawTx(execName string, amount uint64, proofs []*mixTy.ZkProofInfo) (*types.Transaction, error) {
	payload := &mixTy.MixWithdrawAction{}
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
