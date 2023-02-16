// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/plugin/plugin/crypto/bls"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
)

// GetExecAddr 获取执行器地址
func GetExecAddr(exec string) (string, error) {
	if ok := types.IsAllowExecName([]byte(exec), []byte(exec)); !ok {
		return "", types.ErrExecNameNotAllow
	}

	addrResult := address.ExecAddress(exec)
	result := addrResult
	return result, nil
}

func getRealExecName(paraName string, name string) string {
	if strings.HasPrefix(name, pt.ParaPrefix) {
		return name
	}
	return paraName + name
}

func getBlsPubFromSecp256Key(key string) (string, error) {

	d1 := secp256k1.Driver{}
	if key == "" {
		return "", types.ErrInvalidParam
	}
	privByte, err := common.FromHex(key)
	if err != nil {
		return "", err
	}
	priv, err := d1.PrivKeyFromBytes(privByte[:])
	if err != nil {
		return "", err
	}
	_, blsPriv := bls.MustPrivKeyFromBytes(priv.Bytes())
	return hex.EncodeToString(blsPriv.PubKey().Bytes()), nil
}
