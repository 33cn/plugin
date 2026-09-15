package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"

	wcom "github.com/33cn/chain33/wallet/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// legacyCBCEncrypt 复刻最初版 chain33 CBCEncrypterPrivkey：固定 IV=key[:16]，ciphertext-only。
func legacyCBCEncrypt(password, plain []byte) []byte {
	key := make([]byte, 32)
	copy(key, password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	iv := key[:block.BlockSize()]
	out := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, plain)
	return out
}

// randomIVCBCEncrypt 复刻 v0.69.1~v1.70 的 CBCEncrypterPrivkey：随机 IV，IV(16)+ciphertext。
func randomIVCBCEncrypt(password, plain []byte) []byte {
	key := make([]byte, 32)
	copy(key, password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return nil
	}
	out := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, plain)
	return append(iv, out...)
}

func assertHasPrivKeyMagic(t *testing.T, data []byte) {
	t.Helper()
	require.GreaterOrEqual(t, len(data), len(wcom.MagicPrivKey)+1)
	assert.True(t, bytes.HasPrefix(data, wcom.MagicPrivKey), "当前加密应带 C33K 魔数")
	assert.Equal(t, wcom.KdfVersion, data[len(wcom.MagicPrivKey)], "当前加密 version 应为 KdfVersion")
}

// TestDecryptDataWithPadingCompat 覆盖 chain33 三代 CBC 密文的解密：
//   - 第 3 代（v1.71.0+）：Magic+version+salt+IV+ciphertext，总长 %32==5
//   - 第 2 代（v0.69.1~v1.70）：IV(16)+ciphertext，总长 %32==16（含 96/160 明文）
//   - 第 1 代：ciphertext-only，总长 %32==0
func TestDecryptDataWithPadingCompat(t *testing.T) {
	password := make([]byte, 32)
	if _, err := rand.Read(password); err != nil {
		t.Fatal(err)
	}

	plains := [][]byte{
		[]byte("short"),
		bytes.Repeat([]byte{0xab}, 33),  // pad 到 64
		bytes.Repeat([]byte{0xcd}, 65),  // pad 到 96
		bytes.Repeat([]byte{0xef}, 129), // pad 到 160
	}

	for _, plain := range plains {
		padded := pKCS5Padding(plain, len(password))

		// 第 3 代：当前 CBCEncrypterPrivkey 输出
		newData := encryptDataWithPadding(password, plain)
		assertHasPrivKeyMagic(t, newData)
		assert.Equal(t, 5, len(newData)%32, "第3代总长应为 37+32k，%32==5")
		got, err := decryptDataWithPading(password, newData)
		assert.NoError(t, err)
		assert.Equal(t, plain, got, "第3代解密应还原原文")

		// 第 2 代：存量 IV+ciphertext（96/160 不能交给 chain33 解密）
		midData := randomIVCBCEncrypt(password, padded)
		assert.False(t, hasPrivKeyMagic(midData))
		assert.Equal(t, 16, len(midData)%32, "第2代长度应为 32k+16")
		gotMid, err := decryptDataWithPading(password, midData)
		assert.NoError(t, err)
		assert.Equal(t, plain, gotMid, "第2代解密应还原原文")

		// 第 1 代：存量 ciphertext-only
		oldData := legacyCBCEncrypt(password, padded)
		assert.False(t, hasPrivKeyMagic(oldData))
		assert.Equal(t, 0, len(oldData)%32, "第1代长度应为 32k")
		gotOld, err := decryptDataWithPading(password, oldData)
		assert.NoError(t, err)
		assert.Equal(t, plain, gotOld, "第1代解密应还原原文")
	}
}
