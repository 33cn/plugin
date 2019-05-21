// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"

	"github.com/33cn/chain33/types"
)

type parseUnfreezeCreate struct {
	StartTime      int64           `protobuf:"varint,1,opt,name=startTime,proto3" json:"startTime,omitempty"`
	AssetExec      string          `protobuf:"bytes,2,opt,name=assetExec,proto3" json:"assetExec,omitempty"`
	AssetSymbol    string          `protobuf:"bytes,3,opt,name=assetSymbol,proto3" json:"assetSymbol,omitempty"`
	TotalCount     int64           `protobuf:"varint,4,opt,name=totalCount,proto3" json:"totalCount,omitempty"`
	Beneficiary    string          `protobuf:"bytes,5,opt,name=beneficiary,proto3" json:"beneficiary,omitempty"`
	Means          string          `protobuf:"bytes,6,opt,name=means,proto3" json:"means,omitempty"`
	FixAmount      *FixAmount      `json:"fixAmount,omitempty"`
	LeftProportion *LeftProportion `json:"leftProportion,omitempty"`
}

// UnmarshalJSON 解析UnfreezeCreate
func (m *UnfreezeCreate) UnmarshalJSON(v []byte) error {
	var c parseUnfreezeCreate
	err := json.Unmarshal(v, &c)
	if err != nil {
		return err
	}
	if c.Means == FixAmountX && c.FixAmount != nil {
		m.MeansOpt = &UnfreezeCreate_FixAmount{FixAmount: c.FixAmount}
	} else if c.Means == LeftProportionX && c.LeftProportion != nil {
		m.MeansOpt = &UnfreezeCreate_LeftProportion{LeftProportion: c.LeftProportion}
	} else {
		return types.ErrInvalidParam
	}
	m.StartTime = c.StartTime
	m.AssetSymbol, m.AssetExec = c.AssetSymbol, c.AssetExec
	m.TotalCount, m.Beneficiary = c.TotalCount, c.Beneficiary
	m.Means = c.Means
	return nil
}
