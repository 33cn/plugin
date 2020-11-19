// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	// ParaX paracross exec name
	MixX = "mix"
	glog = log.New("module", MixX)
)

func init() {
	// init executor type
	types.AllowUserExec = append(types.AllowUserExec, []byte(MixX))
	types.RegFork(MixX, InitFork)
	types.RegExec(MixX, InitExecutor)

}

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(MixX, "Enable", 0)

}

//InitExecutor ...
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(MixX, NewType(cfg))
}

// GetExecName get para exec name
func GetExecName(cfg *types.Chain33Config) string {
	return cfg.ExecName(MixX)
}

// ParacrossType base paracross type
type MixType struct {
	types.ExecTypeBase
}

// NewType get paracross type
func NewType(cfg *types.Chain33Config) *MixType {
	c := &MixType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetName 获取执行器名称
func (p *MixType) GetName() string {
	return MixX
}

// GetLogMap get receipt log map
func (p *MixType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogMixConfigVk:             {Ty: reflect.TypeOf(ZkVerifyKeys{}), Name: "LogMixConfigVk"},
		TyLogMixConfigAuth:           {Ty: reflect.TypeOf(AuthPubKeys{}), Name: "LogMixConfigAuthPubKey"},
		TyLogCurrentCommitTreeLeaves: {Ty: reflect.TypeOf(CommitTreeLeaves{}), Name: "TyLogCommitTreeLeaves"},
		TyLogCurrentCommitTreeRoots:  {Ty: reflect.TypeOf(CommitTreeRoots{}), Name: "TyLogCommitTreeRoots"},
	}
}

// GetTypeMap get action type
func (p *MixType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Config":    MixActionConfig,
		"Deposit":   MixActionDeposit,
		"Withdraw":  MixActionWithdraw,
		"Transfer":  MixActionTransfer,
		"Authorize": MixActionAuth,
	}
}

// GetPayload mix get action payload
func (p *MixType) GetPayload() types.Message {
	return &MixAction{}
}

//
//func DecodeDepositInput(input string) (*DepositPublicInput, error) {
//	var v DepositPublicInput
//	data, err := hex.DecodeString(input)
//	if err != nil {
//		return nil, errors.Wrapf(err, "decode string=%s", input)
//	}
//	err = json.Unmarshal(data, &v)
//	if err != nil {
//		return nil, errors.Wrapf(err, "unmarshal string=%s", input)
//	}
//
//	return &v, nil
//}
//
//func DecodeWithdrawInput(input string) (*WithdrawPublicInput, error) {
//	var v WithdrawPublicInput
//	data, err := hex.DecodeString(input)
//	if err != nil {
//		return nil, errors.Wrapf(err, "decode string=%s", input)
//	}
//	err = json.Unmarshal(data, &v)
//	if err != nil {
//		return nil, errors.Wrapf(err, "unmarshal string=%s", input)
//	}
//
//	return &v, nil
//}
//
//
//func DecodeTransferInput(input string) (*TransferInputPublicInput, error) {
//	var v TransferInputPublicInput
//	data, err := hex.DecodeString(input)
//	if err != nil {
//		return nil, errors.Wrapf(err, "decode string=%s", input)
//	}
//	err = json.Unmarshal(data, &v)
//	if err != nil {
//		return nil, errors.Wrapf(err, "unmarshal string=%s", input)
//	}
//
//	return &v, nil
//}
//
//func DecodeTransferOut(input string) (*TransferOutputPublicInput, error) {
//	var v TransferOutputPublicInput
//	data, err := hex.DecodeString(input)
//	if err != nil {
//		return nil, errors.Wrapf(err, "decode string=%s", input)
//	}
//	err = json.Unmarshal(data, &v)
//	if err != nil {
//		return nil, errors.Wrapf(err, "unmarshal string=%s", input)
//	}
//
//	return &v, nil
//}
//
//func DecodeAuthorizeInput(input string) (*AuthorizePublicInput, error) {
//	var v AuthorizePublicInput
//	data, err := hex.DecodeString(input)
//	if err != nil {
//		return nil, errors.Wrapf(err, "decode string=%s", input)
//	}
//	err = json.Unmarshal(data, &v)
//	if err != nil {
//		return nil, errors.Wrapf(err, "unmarshal string=%s", input)
//	}
//
//	return &v, nil
//}

func DecodePubInput(ty VerifyType, input string) (interface{}, error) {
	data, err := hex.DecodeString(input)
	if err != nil {
		return nil, errors.Wrapf(err, "decode string=%s", input)
	}
	switch ty {
	case VerifyType_DEPOSIT:
		var v DepositPublicInput
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal string=%s", input)
		}
		return &v, nil
	case VerifyType_WITHDRAW:
		var v WithdrawPublicInput
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal string=%s", input)
		}
		return &v, nil
	case VerifyType_TRANSFERINPUT:
		var v TransferInputPublicInput
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal string=%s", input)
		}
		return &v, nil
	case VerifyType_TRANSFEROUTPUT:
		var v TransferOutputPublicInput
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal string=%s", input)
		}
		return &v, nil
	case VerifyType_AUTHORIZE:
		var v AuthorizePublicInput
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal string=%s", input)
		}
		return &v, nil
	}
	return nil, types.ErrInvalidParam
}
