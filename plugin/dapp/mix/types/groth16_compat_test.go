//go:build !386

package types

import (
	"bytes"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	groth16bn254 "github.com/consensys/gnark/backend/groth16/bn254"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	r1csbuilder "github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/stretchr/testify/require"
)

// compatCircuit 是用于生成新格式（gnark v0.9.0）VK/proof 的最小电路，
// 不使用 gnark commitment（无 gnark:",commitment" 字段），与 mix/zksync 电路一致。
type compatCircuit struct {
	X frontend.Variable `gnark:",public"`
	Y frontend.Variable `gnark:",public"`
	Z frontend.Variable
}

func (c *compatCircuit) Define(api frontend.API) error {
	res := api.Mul(c.X, c.Y)
	api.AssertIsEqual(res, c.Z)
	return nil
}

// newFormatVKProof 用当前 gnark v0.9.0 生成新格式 VK/proof/PK 及公开 witness。
// 同时给出旧格式字节（旧格式是新格式的前缀）：
//
//	VK     = bellman 主部分 + PublicAndCommitmentCommitted(空,4B) + CommitmentKey(2×G2,128B)
//	proof  = Ar|Bs|Krs + Commitments(空,4B) + CommitmentPok(32B)
//	PK     = Domain + 主部分 + uint32(len(CommitmentKeys))(无 commitment 电路为 4B)
func newFormatVKProof(t *testing.T) (newVK, oldVK, newProof, oldProof, newPK, oldPK []byte, pubW witness.Witness) {
	t.Helper()
	var circuit compatCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1csbuilder.NewBuilder, &circuit)
	require.NoError(t, err)

	circuit.X = 3
	circuit.Y = 5
	circuit.Z = 15

	w, err := frontend.NewWitness(&circuit, ecc.BN254.ScalarField())
	require.NoError(t, err)
	pubW, err = w.Public()
	require.NoError(t, err)

	pk, vk, err := groth16.Setup(ccs)
	require.NoError(t, err)

	proof, err := groth16.Prove(ccs, pk, w)
	require.NoError(t, err)

	var vkBuf, proofBuf, pkBuf bytes.Buffer
	_, err = vk.WriteTo(&vkBuf)
	require.NoError(t, err)
	_, err = proof.WriteTo(&proofBuf)
	require.NoError(t, err)
	_, err = pk.WriteTo(&pkBuf)
	require.NoError(t, err)
	newVK = vkBuf.Bytes()
	newProof = proofBuf.Bytes()
	newPK = pkBuf.Bytes()

	// 旧格式是前缀：截掉新格式追加的 commitment 相关尾部
	require.True(t, len(newVK) > 132, "new vk len=%d", len(newVK))
	require.True(t, len(newProof) > 36, "new proof len=%d", len(newProof))
	require.True(t, len(newPK) > 4, "new pk len=%d", len(newPK))
	oldVK = newVK[:len(newVK)-132]
	oldProof = newProof[:len(newProof)-36]
	oldPK = newPK[:len(newPK)-4]
	return
}

// TestReadGroth16CompatibleNewFormat 新格式 VK/proof 必须仍走标准 ReadFrom 且验证通过
func TestReadGroth16CompatibleNewFormat(t *testing.T) {
	newVK, _, newProof, _, _, _, pubW := newFormatVKProof(t)

	vk, err := ReadVerifyingKeyCompatible(newVK)
	require.NoError(t, err, "新格式 VK 应能被兼容 reader 读取")
	proof, err := ReadProofCompatible(newProof)
	require.NoError(t, err, "新格式 proof 应能被兼容 reader 读取")

	require.NoError(t, groth16.Verify(proof, vk, pubW), "新格式验证应通过")
}

// TestReadGroth16CompatibleOldFormat 旧格式（v0.5.2，bellman 主部分）VK/proof 应被兼容 reader
// 读取并验证通过
func TestReadGroth16CompatibleOldFormat(t *testing.T) {
	_, oldVK, _, oldProof, _, _, pubW := newFormatVKProof(t)

	vk, err := ReadVerifyingKeyCompatible(oldVK)
	require.NoError(t, err, "旧格式 VK 应能被兼容 reader 读取")
	proof, err := ReadProofCompatible(oldProof)
	require.NoError(t, err, "旧格式 proof 应能被兼容 reader 读取")

	require.NoError(t, groth16.Verify(proof, vk, pubW), "旧格式验证应通过")
}

// TestReadGroth16CompatibleTruncatedInvalid 截断一半的新格式数据不应被静默解析为旧格式
func TestReadGroth16CompatibleTruncatedInvalid(t *testing.T) {
	newVK, _, newProof, _, newPK, _, _ := newFormatVKProof(t)

	// 截取 bellman 主部分与完整新格式之间的长度，应因 trailing bytes 报错
	badVK := newVK[:len(newVK)-100]
	_, err := ReadVerifyingKeyCompatible(badVK)
	require.Error(t, err, "非完整旧/新格式 VK 应报错")

	badProof := newProof[:len(newProof)-20]
	_, err = ReadProofCompatible(badProof)
	require.Error(t, err, "非完整旧/新格式 proof 应报错")

	// PK：截掉新格式追加的 uint32(len(CommitmentKeys)) 的一部分（剩余 2 字节），
	// 标准 ReadFrom 会因 EOF 失败，旧格式解析应因 trailing bytes 报错
	badPK := newPK[:len(newPK)-2]
	_, err = ReadProvingKeyCompatible(badPK)
	require.Error(t, err, "非完整旧/新格式 PK 应报错")

	badPKMain := newPK[:len(newPK)-100]
	_, err = ReadProvingKeyCompatible(badPKMain)
	require.Error(t, err, "截断主部分的 PK 应报错")
}

// TestReadGroth16CompatibleNewPK 新格式（v0.9.0）PK 应被兼容 reader 读取，生成 proof 后新 VK 验证通过
func TestReadGroth16CompatibleNewPK(t *testing.T) {
	newVK, _, _, _, newPK, _, pubW := newFormatVKProof(t)

	pk, err := ReadProvingKeyCompatible(newPK)
	require.NoError(t, err, "新格式 PK 应能被兼容 reader 读取")

	proof := proveWithReadBackPK(t, pk)

	vk, err := ReadVerifyingKeyCompatible(newVK)
	require.NoError(t, err)
	require.NoError(t, groth16.Verify(proof, vk, pubW), "新格式 PK 生成的 proof 验证应通过")
}

// TestReadGroth16CompatibleOldPK 旧格式（v0.5.2）PK 应被兼容 reader 读取，生成 proof 后
// 旧 VK（截尾）验证通过 —— 模拟"旧 PK 生成 proof + 链上旧 VK"的无缝替换场景
func TestReadGroth16CompatibleOldPK(t *testing.T) {
	_, oldVK, _, _, _, oldPK, pubW := newFormatVKProof(t)

	pk, err := ReadProvingKeyCompatible(oldPK)
	require.NoError(t, err, "旧格式 PK 应能被兼容 reader 读取")

	proof := proveWithReadBackPK(t, pk)

	vk, err := ReadVerifyingKeyCompatible(oldVK)
	require.NoError(t, err)
	require.NoError(t, groth16.Verify(proof, vk, pubW), "旧 PK 生成 proof + 旧 VK 验证应通过")
}

// proveWithReadBackPK 用兼容 reader 读回的 PK 对 compatCircuit 重新生成 proof
func proveWithReadBackPK(t *testing.T, pk *groth16bn254.ProvingKey) groth16.Proof {
	t.Helper()
	var circuit compatCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1csbuilder.NewBuilder, &circuit)
	require.NoError(t, err)

	circuit.X = 3
	circuit.Y = 5
	circuit.Z = 15

	w, err := frontend.NewWitness(&circuit, ecc.BN254.ScalarField())
	require.NoError(t, err)

	proof, err := groth16.Prove(ccs, pk, w)
	require.NoError(t, err, "读回 PK 应能正常生成 proof")
	return proof
}
