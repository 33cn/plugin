// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"encoding/json"
	"reflect"

	"github.com/33cn/chain33/common/address"

	"github.com/33cn/chain33/common/crypto"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var tlog = log.New("module", Pos33X)

func init() {
	// init executor type
	types.AllowDepositExec = append(types.AllowDepositExec, []byte(Pos33X))
	types.AllowUserExec = append(types.AllowUserExec, []byte(Pos33X))
	types.RegistorExecutor(Pos33X, NewType())
	types.RegisterDappFork(Pos33X, "Enable", 0)
}

// NewType  new type
func NewType() *Pos33Type {
	c := &Pos33Type{}
	c.SetChild(c)
	return c
}

// Pos33Type execType
type Pos33Type struct {
	types.ExecTypeBase
}

// GetName 获取执行器名称
func (pt *Pos33Type) GetName() string {
	return Pos33X
}

// GetLogMap get log
func (pt *Pos33Type) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogDeposit:  {Ty: reflect.TypeOf(ReceiptPos33{}), Name: "LogDeposit"},
		TyLogWithdraw: {Ty: reflect.TypeOf(ReceiptPos33{}), Name: "LogWithdraw"},
		TyLogDelegate: {Ty: reflect.TypeOf(ReceiptPos33{}), Name: "LogDelegate"},
		TyLogReword:   {Ty: reflect.TypeOf(ReceiptPos33{}), Name: "LogReword"},
		TyLogPunish:   {Ty: reflect.TypeOf(ReceiptPos33{}), Name: "LogPunish"},
		TyLogElecte:   {Ty: reflect.TypeOf(ReceiptPos33{}), Name: "LogElecte"},
	}
}

// GetPayload get payload
func (pt *Pos33Type) GetPayload() types.Message {
	return &Pos33Action{}
}

// GetTypeMap get typeMap
func (pt *Pos33Type) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Deposit":  Pos33ActionDeposit,
		"Withdraw": Pos33ActionWithdraw,
		"Delegate": Pos33ActionDelegate,
		"Reword":   Pos33ActionReword,
		"Punish":   Pos33ActionPunish,
		"electe":   Pos33ActionElecte,
	}
}

// Verify is verify vote msg
func (v *Pos33Vote) Verify() bool {
	s := v.Sig
	v.Sig = nil
	b := crypto.Sha256(types.Encode(v))
	v.Sig = s
	return types.CheckSign(b, "", s)
}

// Sign is sign vote msg
func (v *Pos33Vote) Sign(priv crypto.PrivKey) {
	v.Sig = nil
	b := crypto.Sha256(types.Encode(v))
	sig := priv.Sign(b)
	v.Sig = &types.Signature{Ty: types.ED25519, Pubkey: priv.PubKey().Bytes(), Signature: sig.Bytes()}
}

// ToString is rands to string
func (m *Pos33Rands) ToString() string {
	s := ""
	for _, r := range m.Rands {
		s += hex.EncodeToString(r.RandHash) + " "
	}
	return s
}

// ToString is reword to string
func (act *Pos33RewordAction) ToString() string {
	b, err := json.MarshalIndent(act, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (m *Pos33Rands) Len() int { return len(m.Rands) }
func (m *Pos33Rands) Less(i, j int) bool {
	return string(m.Rands[i].RandHash) < string(m.Rands[j].RandHash)
}
func (m *Pos33Rands) Swap(i, j int) { m.Rands[i], m.Rands[j] = m.Rands[j], m.Rands[i] }

// Public is get Pos33Rand public in singature helper func
func (m *Pos33Rand) Public() string {
	return string(m.Sig.Pubkey)
}

// Address is get Pos33Rand address in singature helper func
func (m *Pos33Rand) Address() string {
	return address.PubKeyToAddress([]byte(m.Public())).String()
}
