package crypto

import (
	"crypto/cipher"
	"crypto/des"
)

type DES struct {
	key []byte
	//iv的长度必须等于block块的大小
	iv []byte
}

func NewDES(key, iv []byte) *DES {
	return &DES{key: key, iv: iv}
}
func (d *DES) Encrypt(origData []byte) ([]byte, error) {
	block, err := des.NewCipher(d.key)
	if err != nil {
		return nil, err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, d.iv[:block.BlockSize()])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 密钥key长度固定8字节
func (d *DES) Decrypt(crypted []byte) ([]byte, error) {
	block, err := des.NewCipher(d.key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, d.iv[:block.BlockSize()])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}

type TripleDES struct {
	key []byte
	//iv的长度必须等于block块的大小
	iv []byte
}

func NewTripleDES(key, iv []byte) *TripleDES {
	return &TripleDES{key: key, iv: iv}
}

// 3DES加密 24字节
func (d *TripleDES) Encrypt(origData []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(d.key)
	if err != nil {
		return nil, err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, d.iv[:block.BlockSize()])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 3DES解密
func (d *TripleDES) Decrypt(crypted []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(d.key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, d.iv[:block.BlockSize()])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}
