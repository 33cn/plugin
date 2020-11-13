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
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/consensys/gnark/backend"
	groth16_bn256 "github.com/consensys/gnark/backend/bn256/groth16"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gurvy"

	"github.com/pkg/errors"
)

func getByteBuff(input string) (*bytes.Buffer, error) {
	var buffInput bytes.Buffer
	res, err := hex.DecodeString(input)
	if err != nil {
		return nil, errors.Wrapf(err, "getByteBuff to %s", input)
	}
	buffInput.Write(res)
	return &buffInput, nil

}

func deserializeInput(input string) (map[string]interface{}, error) {
	buff, err := getByteBuff(input)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(buff)
	toRead := make(map[string]interface{})
	if err := decoder.Decode(&toRead); err != nil {
		return nil, errors.Wrapf(err, "deserializeInput %s", input)
	}

	return toRead, nil
}

func Verify(verifyKeyStr, proofStr, pubInputStr string) (bool, error) {
	curveID := gurvy.BN256

	output, err := getByteBuff(verifyKeyStr)
	if err != nil {
		return false, errors.Wrapf(err, "zk.verify")
	}
	var vk groth16_bn256.VerifyingKey
	if err := gob.Deserialize(output, &vk, curveID); err != nil {
		return false, errors.Wrapf(err, "zk.verify.Deserize.VK=%s", verifyKeyStr[:10])
	}

	// parse input file
	assigns, err := deserializeInput(pubInputStr)
	if err != nil {
		return false, err
	}
	r1csInput := backend.NewAssignment()
	for k, v := range assigns {
		r1csInput.Assign(backend.Public, k, v)
	}

	// load proof
	output, err = getByteBuff(proofStr)
	if err != nil {
		return false, errors.Wrapf(err, "proof")
	}
	var proof groth16_bn256.Proof
	if err := gob.Deserialize(output, &proof, curveID); err != nil {
		return false, errors.Wrapf(err, "zk.verify.deserial.proof=%s", proofStr[:10])
	}

	// verify proof
	//start := time.Now()
	result, err := groth16_bn256.Verify(&proof, &vk, r1csInput)
	if err != nil {
		return false, errors.Wrapf(err, "zk.Verify")
	}
	return result, nil
}
