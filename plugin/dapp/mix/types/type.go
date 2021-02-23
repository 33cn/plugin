// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/consensys/gurvy/bn256/fr"
	"github.com/consensys/gurvy/bn256/twistededwards"
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
		TyLogMixConfigAuth:           {Ty: reflect.TypeOf(AuthKeys{}), Name: "LogMixConfigAuthPubKey"},
		TyLogCurrentCommitTreeLeaves: {Ty: reflect.TypeOf(CommitTreeLeaves{}), Name: "LogCommitTreeLeaves"},
		TyLogCurrentCommitTreeRoots:  {Ty: reflect.TypeOf(CommitTreeRoots{}), Name: "LogCommitTreeRoots"},
		TyLogMixConfigPaymentKey:     {Ty: reflect.TypeOf(PaymentKey{}), Name: "LogConfigReceivingKey"},
		TyLogNulliferSet:             {Ty: reflect.TypeOf(ExistValue{}), Name: "LogNullifierSet"},
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

func MulCurvePointG(val interface{}) *twistededwards.Point {
	v := fr.FromInterface(val)
	var point twistededwards.Point
	ed := twistededwards.GetEdwardsCurve()
	point.ScalarMul(&ed.Base, *v.FromMont())
	return &point
}

func MulCurvePointH(val string) *twistededwards.Point {
	v := fr.FromInterface(val)

	var pointV, pointH twistededwards.Point
	pointH.X.SetString(PointHX)
	pointH.Y.SetString(PointHY)

	pointV.ScalarMul(&pointH, *v.FromMont())
	return &pointV
}

func GetCurveSum(points ...*twistededwards.Point) *twistededwards.Point {

	//Add之前需初始化pointSum,不能空值，不然会等于0
	pointSum := twistededwards.NewPoint(points[0].X, points[0].Y)
	for _, a := range points[1:] {
		pointSum.Add(&pointSum, a)
	}

	return &pointSum
}

//A=B+C
func CheckSumEqual(points ...*twistededwards.Point) bool {
	if len(points) < 2 {
		return false
	}
	//Add之前需初始化pointSum,不能空值，不然会等于0
	pointSum := twistededwards.NewPoint(points[1].X, points[1].Y)
	for _, a := range points[2:] {
		pointSum.Add(&pointSum, a)
	}

	if pointSum.X.Equal(&points[0].X) && pointSum.Y.Equal(&points[0].Y) {
		return true
	}
	return false

}

func GetByteBuff(input string) (*bytes.Buffer, error) {
	var buffInput bytes.Buffer
	res, err := hex.DecodeString(input)
	if err != nil {
		return nil, errors.Wrapf(err, "getByteBuff to %s", input)
	}
	buffInput.Write(res)
	return &buffInput, nil

}

func Str2Byte(v string) []byte {
	var fr fr.Element
	fr.SetString(v)
	return fr.Bytes()
}
func Byte2Str(v []byte) string {
	var f fr.Element
	f.SetBytes(v)
	return f.String()
}

func GetFrRandom() string {
	var f fr.Element
	return f.SetRandom().String()
}
