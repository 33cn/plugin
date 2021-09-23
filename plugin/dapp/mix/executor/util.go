// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/zksnark"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"

	"github.com/pkg/errors"
)

//考虑vk平滑切换的场景，允许有两个vk存在
func zkProofVerify(db dbm.KV, proof *mixTy.ZkProofInfo, ty mixTy.VerifyType) error {
	keys, err := getVerifyKeys(db, int32(ty))
	if err != nil {
		return err
	}

	var pass bool
	for _, verifyKey := range keys.Data {
		ok, err := zksnark.Verify(verifyKey.Value, proof.Proof, proof.PublicInput)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		pass = true
		break
	}
	if !pass {
		return errors.Wrap(mixTy.ErrZkVerifyFail, "verify")
	}

	return nil
}
