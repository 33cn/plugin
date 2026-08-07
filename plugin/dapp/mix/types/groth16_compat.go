package types

import (
	"bytes"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	groth16bn254 "github.com/consensys/gnark/backend/groth16/bn254"
	"github.com/pkg/errors"
)

// ReadVerifyingKeyCompatible 兼容读取 groth16 VK 字节：
//
//   - 新格式（gnark v0.9.0）：bellman 主部分（[α]1,[β]1,[β]2,[γ]2,[δ]1,[δ]2,uint32(len(Kvk)),[Kvk]1）
//   - PublicAndCommitmentCommitted + CommitmentKey
//   - 旧格式（gnark v0.5.2）：仅 bellman 主部分，与 v0.9.0 的主部分逐字节相同（旧格式是新格式的前缀）
//
// 实现策略：先尝试 v0.9.0 的标准 ReadFrom（新格式），读旧格式会在解析完主部分后
// 因缺少 PublicAndCommitmentCommitted 而返回 EOF/错误，此时回退到旧格式手动解析
// （主部分 + 空 commitment）。
//
// 返回的 VK 已通过 Precompute 计算好 e、-[δ]2、-[γ]2，可直接用于 groth16.Verify。
// mix/zksync 电路不使用 gnark commitment，故旧格式 VK 的 CommitmentKey 保持零值、
// PublicAndCommitmentCommitted 为空，Verify 中空 commitment 路径可正常通过。
func ReadVerifyingKeyCompatible(buf []byte) (*groth16bn254.VerifyingKey, error) {
	// 先尝试新格式（v0.9.0），新格式必须走标准 ReadFrom，保证行为不变
	vk := &groth16bn254.VerifyingKey{}
	if _, err := vk.ReadFrom(bytes.NewReader(buf)); err == nil {
		return vk, nil
	}

	// 旧格式（v0.5.2）：仅 bellman 主部分，无 PublicAndCommitmentCommitted / CommitmentKey
	vkOld := &groth16bn254.VerifyingKey{}
	dec := bn254.NewDecoder(bytes.NewReader(buf))
	// [α]1,[β]1,[β]2,[γ]2,[δ]1,[δ]2
	if err := dec.Decode(&vkOld.G1.Alpha); err != nil {
		return nil, errors.Wrapf(err, "decode old vk G1.Alpha")
	}
	if err := dec.Decode(&vkOld.G1.Beta); err != nil {
		return nil, errors.Wrapf(err, "decode old vk G1.Beta")
	}
	if err := dec.Decode(&vkOld.G2.Beta); err != nil {
		return nil, errors.Wrapf(err, "decode old vk G2.Beta")
	}
	if err := dec.Decode(&vkOld.G2.Gamma); err != nil {
		return nil, errors.Wrapf(err, "decode old vk G2.Gamma")
	}
	if err := dec.Decode(&vkOld.G1.Delta); err != nil {
		return nil, errors.Wrapf(err, "decode old vk G1.Delta")
	}
	if err := dec.Decode(&vkOld.G2.Delta); err != nil {
		return nil, errors.Wrapf(err, "decode old vk G2.Delta")
	}
	// uint32(len(Kvk)),[Kvk]1
	if err := dec.Decode(&vkOld.G1.K); err != nil {
		return nil, errors.Wrapf(err, "decode old vk G1.K")
	}
	// 旧格式到 bellman 主部分即结束，必须完整消费
	if dec.BytesRead() != int64(len(buf)) {
		return nil, errors.Errorf("old vk trailing bytes: read=%d len=%d", dec.BytesRead(), len(buf))
	}
	// 重新计算 e、-[δ]2、-[γ]2（groth16.Verify 依赖，正常 ReadFrom 会做同样的事）
	if err := vkOld.Precompute(); err != nil {
		return nil, errors.Wrapf(err, "precompute old vk")
	}
	return vkOld, nil
}

// ReadProofCompatible 兼容读取 groth16 Proof 字节：
//
//   - 新格式（gnark v0.9.0）：Ar|Bs|Krs + Commitments + CommitmentPok
//   - 旧格式（gnark v0.5.2）：Ar|Bs|Krs，与 v0.9.0 的前三部分相同（旧格式是新格式的前缀）
//
// 先尝试 v0.9.0 的标准 ReadFrom（新格式），旧格式会因缺少 Commitments 字段而 EOF/失败，
// 此时回退到旧格式手动解析（前三部分，Commitments 为空、CommitmentPok 为零点）。
// mix/zksync 电路不使用 gnark commitment，空 commitment 的 proof 可被 groth16.Verify 正常验证。
func ReadProofCompatible(buf []byte) (*groth16bn254.Proof, error) {
	// 先尝试新格式（v0.9.0）
	proof := &groth16bn254.Proof{}
	if _, err := proof.ReadFrom(bytes.NewReader(buf)); err == nil {
		return proof, nil
	}

	// 旧格式（v0.5.2）：Ar|Bs|Krs
	proofOld := &groth16bn254.Proof{}
	dec := bn254.NewDecoder(bytes.NewReader(buf))
	if err := dec.Decode(&proofOld.Ar); err != nil {
		return nil, errors.Wrapf(err, "decode old proof Ar")
	}
	if err := dec.Decode(&proofOld.Bs); err != nil {
		return nil, errors.Wrapf(err, "decode old proof Bs")
	}
	if err := dec.Decode(&proofOld.Krs); err != nil {
		return nil, errors.Wrapf(err, "decode old proof Krs")
	}
	// 旧格式到 Ar|Bs|Krs 即结束，必须完整消费
	if dec.BytesRead() != int64(len(buf)) {
		return nil, errors.Errorf("old proof trailing bytes: read=%d len=%d", dec.BytesRead(), len(buf))
	}
	return proofOld, nil
}

// ReadProvingKeyCompatible 兼容读取 groth16 ProvingKey 字节：
//
//   - 新格式（gnark v0.9.0）：Domain + 主部分（G1.Alpha/Beta/Delta/A/B/Z/K、G2.Beta/Delta/B、
//     nbWires、NbInfinityA/B、InfinityA/B）+ uint32(len(CommitmentKeys)) + CommitmentKeys
//   - 旧格式（gnark v0.5.2）：Domain + 主部分，与 v0.9.0 的主部分逐字节相同（旧格式是新格式的前缀）
//
// 实现策略：先尝试 v0.9.0 的标准 ReadFrom（新格式），读旧格式会在解析完主部分后
// 因缺少 uint32(len(CommitmentKeys)) 而返回 EOF/错误，此时回退到旧格式手动解析
// （Domain + 主部分，CommitmentKeys 为空）。
//
// mix/zksync 电路不使用 gnark commitment（无 gnark:",commitment" 字段），旧格式 PK 的
// CommitmentKeys 为空，groth16.Prove 中 commitment 相关路径不会触发，可直接用于生成 proof。
func ReadProvingKeyCompatible(buf []byte) (*groth16bn254.ProvingKey, error) {
	// 先尝试新格式（v0.9.0），新格式必须走标准 ReadFrom，保证行为不变
	pk := &groth16bn254.ProvingKey{}
	if _, err := pk.ReadFrom(bytes.NewReader(buf)); err == nil {
		return pk, nil
	}

	// 旧格式（v0.5.2）：Domain + 主部分，无 uint32(len(CommitmentKeys)) / CommitmentKeys
	pkOld := &groth16bn254.ProvingKey{}
	n, err := pkOld.Domain.ReadFrom(bytes.NewReader(buf))
	if err != nil {
		return nil, errors.Wrapf(err, "decode old pk Domain")
	}
	dec := bn254.NewDecoder(bytes.NewReader(buf[n:]))

	// G1.Alpha/Beta/Delta/A/B/Z/K、G2.Beta/Delta/B、nbWires、NbInfinityA/B
	if err := dec.Decode(&pkOld.G1.Alpha); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G1.Alpha")
	}
	if err := dec.Decode(&pkOld.G1.Beta); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G1.Beta")
	}
	if err := dec.Decode(&pkOld.G1.Delta); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G1.Delta")
	}
	if err := dec.Decode(&pkOld.G1.A); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G1.A")
	}
	if err := dec.Decode(&pkOld.G1.B); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G1.B")
	}
	if err := dec.Decode(&pkOld.G1.Z); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G1.Z")
	}
	if err := dec.Decode(&pkOld.G1.K); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G1.K")
	}
	if err := dec.Decode(&pkOld.G2.Beta); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G2.Beta")
	}
	if err := dec.Decode(&pkOld.G2.Delta); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G2.Delta")
	}
	if err := dec.Decode(&pkOld.G2.B); err != nil {
		return nil, errors.Wrapf(err, "decode old pk G2.B")
	}
	var nbWires uint64
	if err := dec.Decode(&nbWires); err != nil {
		return nil, errors.Wrapf(err, "decode old pk nbWires")
	}
	if err := dec.Decode(&pkOld.NbInfinityA); err != nil {
		return nil, errors.Wrapf(err, "decode old pk NbInfinityA")
	}
	if err := dec.Decode(&pkOld.NbInfinityB); err != nil {
		return nil, errors.Wrapf(err, "decode old pk NbInfinityB")
	}
	pkOld.InfinityA = make([]bool, nbWires)
	pkOld.InfinityB = make([]bool, nbWires)
	if err := dec.Decode(&pkOld.InfinityA); err != nil {
		return nil, errors.Wrapf(err, "decode old pk InfinityA")
	}
	if err := dec.Decode(&pkOld.InfinityB); err != nil {
		return nil, errors.Wrapf(err, "decode old pk InfinityB")
	}
	// 旧格式到主部分即结束，必须完整消费（防尾随垃圾）
	if n+dec.BytesRead() != int64(len(buf)) {
		return nil, errors.Errorf("old pk trailing bytes: read=%d len=%d", n+dec.BytesRead(), len(buf))
	}
	return pkOld, nil
}
