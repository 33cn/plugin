// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package signatory

import (
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	l "github.com/33cn/chain33/common/log/log15"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	tokenty "github.com/33cn/plugin/plugin/dapp/token/types"
)

var log = l.New("module", "signatory")

// Signatory 签名密钥
type Signatory struct {
	Privkey string
}

// Echo echo
func (*Signatory) Echo(in *string, out *interface{}) error {
	if in == nil {
		return types.ErrInvalidParam
	}
	*out = *in
	return nil
}

// TokenFinish token创建完成
type TokenFinish struct {
	OwnerAddr string `json:"ownerAddr"`
	Symbol    string `json:"symbol"`
	//	Fee       int64  `json:"fee"`
}

// SignApprove 完成签名
func (signatory *Signatory) SignApprove(in *TokenFinish, out *interface{}) error {
	if in == nil {
		return types.ErrInvalidParam
	}
	if len(in.OwnerAddr) == 0 || len(in.Symbol) == 0 {
		return types.ErrInvalidParam
	}
	v := &tokenty.TokenFinishCreate{Symbol: in.Symbol, Owner: in.OwnerAddr}
	finish := &tokenty.TokenAction{
		Ty:    tokenty.TokenActionFinishCreate,
		Value: &tokenty.TokenAction_TokenFinishCreate{TokenFinishCreate: v},
	}

	tx := &types.Transaction{
		Execer:  []byte("token"),
		Payload: types.Encode(finish),
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		To:      address.ExecAddress("token"),
	}

	var err error
	tx.Fee, err = tx.GetRealFee(types.GInt("MinFee"))
	if err != nil {
		log.Error("SignApprove", "calc fee failed", err)
		return err
	}
	err = signTx(tx, signatory.Privkey)
	if err != nil {
		return err
	}
	txHex := types.Encode(tx)
	*out = hex.EncodeToString(txHex)
	return nil
}

// SignTransfer 签名交易，in 输入要数据 签名out 签名之后数据
func (signatory *Signatory) SignTransfer(in *string, out *interface{}) error {
	if in == nil {
		return types.ErrInvalidParam
	}
	if len(*in) == 0 {
		return types.ErrInvalidParam
	}

	amount := 1 * types.Coin
	v := &types.AssetsTransfer{
		Amount: amount,
		Note:   []byte("transfer 1 bty by signatory-server"),
	}
	transfer := &cty.CoinsAction{
		Ty:    cty.CoinsActionTransfer,
		Value: &cty.CoinsAction_Transfer{Transfer: v},
	}

	tx := &types.Transaction{
		Execer:  []byte("coins"),
		Payload: types.Encode(transfer),
		To:      *in,
		Nonce:   rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
	}

	var err error
	tx.Fee, err = tx.GetRealFee(types.GInt("MinFee"))
	if err != nil {
		log.Error("SignTranfer", "calc fee failed", err)
		return err
	}
	err = signTx(tx, signatory.Privkey)
	if err != nil {
		log.Error("SignTranfer", "signTx failed", err)
		return err
	}
	txHex := types.Encode(tx)
	*out = hex.EncodeToString(txHex)
	return nil

}

func signTx(tx *types.Transaction, hexPrivKey string) error {
	c, _ := crypto.New(types.GetSignName("", types.SECP256K1))

	bytes, err := common.FromHex(hexPrivKey)
	if err != nil {
		log.Error("signTx", "err", err)
		return err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		log.Error("signTx", "err", err)
		return err
	}

	tx.Sign(int32(types.SECP256K1), privKey)
	return nil
}
