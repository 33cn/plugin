package types

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/witness"
	"github.com/stretchr/testify/require"
)

// TestReadWitnessCompatible 覆盖新旧两种 witness 字节格式的解析兼容：
// - 新格式（v0.9.0）：[uint32(nbPub)][uint32(nbSec)][uint32(len)][elems]，总长 %32==12
// - 旧格式（v0.5.2）：[uint32(n)][elems]，总长 %32==4
// 两者必须解析出相同的 public input 元素序列。
func TestReadWitnessCompatible(t *testing.T) {
	elems := []string{"12345", "67890", "11121314151617181920", "987654321"}

	newVec := make([]fr.Element, len(elems))
	for i, s := range elems {
		newVec[i].SetString(s)
	}

	// 新格式：witness.New + Fill + WriteTo
	w, err := witness.New(ecc.BN254.ScalarField())
	require.NoError(t, err)
	values := make([]any, len(elems))
	for i := range newVec {
		values[i] = newVec[i]
	}
	ch := make(chan any, len(values))
	for _, v := range values {
		ch <- v
	}
	close(ch)
	require.NoError(t, w.Fill(len(elems), 0, ch))
	var newBuf bytes.Buffer
	_, err = w.WriteTo(&newBuf)
	require.NoError(t, err)
	require.Equal(t, 12, newBuf.Len()%32, "新格式总长应为 12+32k")

	gotW, err := ReadWitnessCompatible(newBuf.Bytes())
	require.NoError(t, err, "新格式应能解析")
	gotVec := gotW.Vector().(fr.Vector)
	require.Len(t, gotVec, len(elems))
	for i, s := range elems {
		require.Equal(t, s, gotVec[i].String(), "新格式解析元素 %d 应一致", i)
	}

	// 旧格式：手动构造 [uint32(n)][n×32B]
	oldBuf := make([]byte, 4+len(elems)*32)
	binary.BigEndian.PutUint32(oldBuf[:4], uint32(len(elems)))
	for i, e := range newVec {
		b := e.Bytes()
		copy(oldBuf[4+i*32:4+(i+1)*32], b[:])
	}
	require.Equal(t, 4, len(oldBuf)%32, "旧格式总长应为 4+32k")

	gotW2, err := ReadWitnessCompatible(oldBuf)
	require.NoError(t, err, "旧格式应能解析")
	gotVec2 := gotW2.Vector().(fr.Vector)
	require.Len(t, gotVec2, len(elems))
	for i, s := range elems {
		require.Equal(t, s, gotVec2[i].String(), "旧格式解析元素 %d 应一致", i)
	}
}

// TestReadWitnessCompatibleInvalidLength 旧格式长度不匹配应报错
func TestReadWitnessCompatibleInvalidLength(t *testing.T) {
	// 声称 3 个元素但字节不足
	badBuf := make([]byte, 4+2*32)
	binary.BigEndian.PutUint32(badBuf[:4], 3)
	_, err := ReadWitnessCompatible(badBuf)
	require.Error(t, err, "长度不匹配应报错")
}
