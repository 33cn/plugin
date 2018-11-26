// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"reflect"

	"github.com/33cn/chain33/types"
)

// PrivacyX privacy executor name
var PrivacyX = "privacy"

const (
	// InvalidAction invalid action type
	InvalidAction = 0
	//action type for privacy

	// ActionPublic2Privacy public to privacy action type
	ActionPublic2Privacy = iota + 100
	// ActionPrivacy2Privacy privacy to privacy action type
	ActionPrivacy2Privacy
	// ActionPrivacy2Public privacy to public action type
	ActionPrivacy2Public

	// log for privacy

	// TyLogPrivacyFee privacy fee log type
	TyLogPrivacyFee = 500
	// TyLogPrivacyInput privacy input type
	TyLogPrivacyInput = 501
	// TyLogPrivacyOutput privacy output type
	TyLogPrivacyOutput = 502

	//SignNameOnetimeED25519 privacy name of crypto
	SignNameOnetimeED25519 = "privacy.onetimeed25519"
	// SignNameRing signature name ring
	SignNameRing = "privacy.RingSignatue"
	// OnetimeED25519 one time ED25519
	OnetimeED25519 = 4
	// RingBaseonED25519 ring raseon ED25519
	RingBaseonED25519 = 5
	// PrivacyMaxCount max mix utxo cout
	PrivacyMaxCount = 16
	// PrivacyTxFee privacy tx fee
	PrivacyTxFee = types.Coin
)

// RescanUtxoFlag
const (
	UtxoFlagNoScan  int32 = 0
	UtxoFlagScaning int32 = 1
	UtxoFlagScanEnd int32 = 2
)

// RescanFlagMapint2string 常量字符串转换关系表
var RescanFlagMapint2string = map[int32]string{
	UtxoFlagNoScan:  "UtxoFlagNoScan",
	UtxoFlagScaning: "UtxoFlagScaning",
	UtxoFlagScanEnd: "UtxoFlagScanEnd",
}

var mapSignType2name = map[int]string{
	OnetimeED25519:    SignNameOnetimeED25519,
	RingBaseonED25519: SignNameRing,
}

var mapSignName2Type = map[string]int{
	SignNameOnetimeED25519: OnetimeED25519,
	SignNameRing:           RingBaseonED25519,
}

func init() {
	// init executor type
	types.AllowUserExec = append(types.AllowUserExec, []byte(PrivacyX))
	types.RegistorExecutor(PrivacyX, NewType())
	types.RegisterDappFork(PrivacyX, "Enable", 980000)
}

// PrivacyType declare PrivacyType class
type PrivacyType struct {
	types.ExecTypeBase
}

// NewType create PrivacyType object
func NewType() *PrivacyType {
	c := &PrivacyType{}
	c.SetChild(c)
	return c
}

// GetPayload get PrivacyType payload
func (t *PrivacyType) GetPayload() types.Message {
	return &PrivacyAction{}
}

// GetName get PrivacyType name
func (t *PrivacyType) GetName() string {
	return PrivacyX
}

// GetLogMap get PrivacyType log map
func (t *PrivacyType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogPrivacyFee:    {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogPrivacyFee"},
		TyLogPrivacyInput:  {Ty: reflect.TypeOf(PrivacyInput{}), Name: "LogPrivacyInput"},
		TyLogPrivacyOutput: {Ty: reflect.TypeOf(PrivacyOutput{}), Name: "LogPrivacyOutput"},
	}
}

// GetTypeMap get PrivacyType type map
func (t *PrivacyType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Public2Privacy":  ActionPublic2Privacy,
		"Privacy2Privacy": ActionPrivacy2Privacy,
		"Privacy2Public":  ActionPrivacy2Public,
	}
}

// ActionName get PrivacyType action name
func (t PrivacyType) ActionName(tx *types.Transaction) string {
	var action PrivacyAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return "unknow-privacy-err"
	}
	return action.GetActionName()
}

// TODO 暂时不修改实现， 先完成结构的重构

// CreateTx create transaction
func (t *PrivacyType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	var tx *types.Transaction
	return tx, nil
}

// Amount get amout
func (t *PrivacyType) Amount(tx *types.Transaction) (int64, error) {
	return 0, nil
}

// GetCryptoDriver get crypto driver
func (t *PrivacyType) GetCryptoDriver(ty int) (string, error) {
	if name, ok := mapSignType2name[ty]; ok {
		return name, nil
	}
	return "", types.ErrNotSupport
}

// GetCryptoType get crypto type
func (t *PrivacyType) GetCryptoType(name string) (int, error) {
	if ty, ok := mapSignName2Type[name]; ok {
		return ty, nil
	}
	return 0, types.ErrNotSupport
}

// GetInput get action input information
func (action *PrivacyAction) GetInput() *PrivacyInput {
	if action.GetTy() == ActionPrivacy2Privacy && action.GetPrivacy2Privacy() != nil {
		return action.GetPrivacy2Privacy().GetInput()

	} else if action.GetTy() == ActionPrivacy2Public && action.GetPrivacy2Public() != nil {
		return action.GetPrivacy2Public().GetInput()
	}
	return nil
}

// GetOutput get action output information
func (action *PrivacyAction) GetOutput() *PrivacyOutput {
	if action.GetTy() == ActionPublic2Privacy && action.GetPublic2Privacy() != nil {
		return action.GetPublic2Privacy().GetOutput()
	} else if action.GetTy() == ActionPrivacy2Privacy && action.GetPrivacy2Privacy() != nil {
		return action.GetPrivacy2Privacy().GetOutput()
	} else if action.GetTy() == ActionPrivacy2Public && action.GetPrivacy2Public() != nil {
		return action.GetPrivacy2Public().GetOutput()
	}
	return nil
}

// GetActionName get action name
func (action *PrivacyAction) GetActionName() string {
	if action.Ty == ActionPrivacy2Privacy && action.GetPrivacy2Privacy() != nil {
		return "Privacy2Privacy"
	} else if action.Ty == ActionPublic2Privacy && action.GetPublic2Privacy() != nil {
		return "Public2Privacy"
	} else if action.Ty == ActionPrivacy2Public && action.GetPrivacy2Public() != nil {
		return "Privacy2Public"
	}
	return "unknow-privacy"
}

// GetTokenName get action token name
func (action *PrivacyAction) GetTokenName() string {
	if action.GetTy() == ActionPublic2Privacy && action.GetPublic2Privacy() != nil {
		return action.GetPublic2Privacy().GetTokenname()
	} else if action.GetTy() == ActionPrivacy2Privacy && action.GetPrivacy2Privacy() != nil {
		return action.GetPrivacy2Privacy().GetTokenname()
	} else if action.GetTy() == ActionPrivacy2Public && action.GetPrivacy2Public() != nil {
		return action.GetPrivacy2Public().GetTokenname()
	}
	return ""
}
