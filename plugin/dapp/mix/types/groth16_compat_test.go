//go:build !386

package types

import (
	"bytes"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
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

// newFormatVKProof 用当前 gnark v0.9.0 生成新格式 VK/proof 及公开 witness。
// 同时给出旧格式字节：新格式 = bellman 主部分 + PublicAndCommitmentCommitted(空,4B)
// + CommitmentKey(2×G2,128B)；proof = Ar|Bs|Krs + Commitments(空,4B) + CommitmentPok(32B)。
func newFormatVKProof(t *testing.T) (newVK, oldVK, newProof, oldProof []byte, pubW witness.Witness) {
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

	var vkBuf, proofBuf bytes.Buffer
	_, err = vk.WriteTo(&vkBuf)
	require.NoError(t, err)
	_, err = proof.WriteTo(&proofBuf)
	require.NoError(t, err)
	newVK = vkBuf.Bytes()
	newProof = proofBuf.Bytes()

	// 旧格式是前缀：截掉新格式追加的 commitment 相关尾部
	require.True(t, len(newVK) > 132, "new vk len=%d", len(newVK))
	require.True(t, len(newProof) > 36, "new proof len=%d", len(newProof))
	oldVK = newVK[:len(newVK)-132]
	oldProof = newProof[:len(newProof)-36]
	return
}

// TestReadGroth16CompatibleNewFormat 新格式 VK/proof 必须仍走标准 ReadFrom 且验证通过
func TestReadGroth16CompatibleNewFormat(t *testing.T) {
	newVK, _, newProof, _, pubW := newFormatVKProof(t)

	vk, err := ReadVerifyingKeyCompatible(newVK)
	require.NoError(t, err, "新格式 VK 应能被兼容 reader 读取")
	proof, err := ReadProofCompatible(newProof)
	require.NoError(t, err, "新格式 proof 应能被兼容 reader 读取")

	require.NoError(t, groth16.Verify(proof, vk, pubW), "新格式验证应通过")
}

// TestReadGroth16CompatibleOldFormat 旧格式（v0.5.2，bellman 主部分）VK/proof 应被兼容 reader
// 读取并验证通过
func TestReadGroth16CompatibleOldFormat(t *testing.T) {
	_, oldVK, _, oldProof, pubW := newFormatVKProof(t)

	vk, err := ReadVerifyingKeyCompatible(oldVK)
	require.NoError(t, err, "旧格式 VK 应能被兼容 reader 读取")
	proof, err := ReadProofCompatible(oldProof)
	require.NoError(t, err, "旧格式 proof 应能被兼容 reader 读取")

	require.NoError(t, groth16.Verify(proof, vk, pubW), "旧格式验证应通过")
}

// TestReadGroth16CompatibleTruncatedInvalid 截断一半的新格式数据不应被静默解析为旧格式
func TestReadGroth16CompatibleTruncatedInvalid(t *testing.T) {
	newVK, _, newProof, _, _ := newFormatVKProof(t)

	// 截取 bellman 主部分与完整新格式之间的长度，应因 trailing bytes 报错
	badVK := newVK[:len(newVK)-100]
	_, err := ReadVerifyingKeyCompatible(badVK)
	require.Error(t, err, "非完整旧/新格式 VK 应报错")

	badProof := newProof[:len(newProof)-20]
	_, err = ReadProofCompatible(badProof)
	require.Error(t, err, "非完整旧/新格式 proof 应报错")
}
