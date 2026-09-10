package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// legacyCBCEncrypt 复刻旧版 chain33 CBCEncrypterPrivkey：固定 IV=key[:16]，ciphertext-only。
// 用于构造升级前格式的密文，验证存量兼容。
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

// TestDecryptDataWithPadingCompat 覆盖新旧两种 CBC 格式的解密兼容：
// - 新格式（chain33 随机 IV，IV(16)+ciphertext，总长 %32==16）
// - 旧格式（固定 IV=key[:16]，ciphertext-only，总长 %32==0）
// 覆盖 32/64/96 字节多种明文长度。
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
		// 新格式往返
		newData := encryptDataWithPadding(password, plain)
		assert.Equal(t, 16, len(newData)%32, "新格式长度应为 32k+16")
		got, err := decryptDataWithPading(password, newData)
		assert.NoError(t, err)
		assert.Equal(t, plain, got, "新格式解密应还原原文")

		// 旧格式往返
		padded := pKCS5Padding(plain, len(password))
		oldData := legacyCBCEncrypt(password, padded)
		assert.Equal(t, 0, len(oldData)%32, "旧格式长度应为 32k")
		got2, err := decryptDataWithPading(password, oldData)
		assert.NoError(t, err)
		assert.Equal(t, plain, got2, "旧格式解密应还原原文")
	}
}
