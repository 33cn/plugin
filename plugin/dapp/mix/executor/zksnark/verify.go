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
	"github.com/consensys/gnark/backend/groth16"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/pkg/errors"
)

func Verify(verifyKeyStr, proofStr, pubInputStr string) (bool, error) {
	vkBuf, err := mixTy.GetByteBuff(verifyKeyStr)
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.vk.GetByteBuff")
	}
	// 兼容读取新旧格式 VK（gnark v0.5.2 旧格式 / v0.9.0 新格式）
	vk, err := mixTy.ReadVerifyingKeyCompatible(vkBuf.Bytes())
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.read.vk=%s", safePrefix(verifyKeyStr, 10))
	}

	// load proof
	proofBuf, err := mixTy.GetByteBuff(proofStr)
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.get.proof")
	}
	// 兼容读取新旧格式 Proof
	proof, err := mixTy.ReadProofCompatible(proofBuf.Bytes())
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.read.proof=%s", safePrefix(proofStr, 10))
	}

	// decode pub input hex string
	pubBuf, err := mixTy.GetByteBuff(pubInputStr)
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.pub.GetByteBuff")
	}

	// 兼容读取新旧 witness 格式，保持存量 pubInput 可验证
	pubW, err := mixTy.ReadWitnessCompatible(pubBuf.Bytes())
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.pub.witness")
	}
	err = groth16.Verify(proof, vk, pubW)
	if err != nil {
		return false, errors.Wrapf(err, "zkVerify.verify")
	}
	return true, nil
}

// safePrefix 返回字符串前 n 字节，长度不足时返回整个字符串。
// verifyKeyStr/proofStr 来自链上交易数据（攻击者可控），直接 s[:10] 会对短输入 panic。
func safePrefix(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
