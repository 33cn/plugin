/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package zksnark

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/pkg/errors"
)

func Verify(verifyKeyStr, proofStr, pubInputStr string) (bool, error) {
	vkBuf, err := mixTy.GetByteBuff(verifyKeyStr)
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.vk.GetByteBuff")
	}
	vk := groth16.NewVerifyingKey(ecc.BN254)
	if _, err := vk.ReadFrom(vkBuf); err != nil {
		return false, errors.Wrapf(err, "zkVerify.read.vk=%s", verifyKeyStr[:10])
	}

	// load proof
	proofBuf, err := mixTy.GetByteBuff(proofStr)
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.get.proof")
	}
	proof := groth16.NewProof(ecc.BN254)
	if _, err = proof.ReadFrom(proofBuf); err != nil {
		return false, errors.Wrapf(err, "zkVerify.read.proof=%s", proofStr[:10])
	}

	// decode pub input hex string
	pubBuf, err := mixTy.GetByteBuff(pubInputStr)
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.pub.GetByteBuff")
	}

	// verify proof
	//start := time.Now()
	err = groth16.ReadAndVerify(proof, vk, pubBuf)
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.verify")
	}
	return true, nil
}
