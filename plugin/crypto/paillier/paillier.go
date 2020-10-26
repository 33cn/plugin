package paillier

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/33cn/chain33/common"
)

func CiphertextAdd(ciphertext1, ciphertext2 string) (string, error) {
	cipherbytes1, err := common.FromHex(ciphertext1)
	if err != nil {
		return "", fmt.Errorf("CiphertextAdd.FromHex. ciphertext1:%s, error:%v", ciphertext1, err)
	}

	cipherbytes2, err := common.FromHex(ciphertext2)
	if err != nil {
		return "", fmt.Errorf("CiphertextAdd.FromHex. ciphertext2:%s, error:%v", ciphertext2, err)
	}

	res, err := CiphertextAddBytes(cipherbytes1, cipherbytes2)
	if err != nil {
		return "", fmt.Errorf("CiphertextAdd.CiphertextAddBytes. error:%v", err)
	}

	return hex.EncodeToString(res), nil
}

func CiphertextAddBytes(cipherbytes1, cipherbytes2 []byte) ([]byte, error) {
	nlen1 := bytesToInt(cipherbytes1[0:2])
	if nlen1 >= len(cipherbytes1)-2 {
		return nil, fmt.Errorf("CiphertextAddBytes. error param length")
	}

	nBytes1 := make([]byte, nlen1)
	copy(nBytes1, cipherbytes1[2:2+nlen1])

	nlen2 := bytesToInt(cipherbytes2[0:2])
	if nlen2 >= len(cipherbytes2)-2 {
		return nil, fmt.Errorf("CiphertextAddBytes. error param length")
	}

	nBytes2 := make([]byte, nlen2)
	copy(nBytes2, cipherbytes2[2:2+nlen2])

	if !bytes.Equal(nBytes1, nBytes2) {
		return nil, fmt.Errorf("CiphertextAddBytes. error: param error nBytes1!=nBytes2")
	}

	data1 := make([]byte, len(cipherbytes1)-nlen1-2)
	copy(data1, cipherbytes1[2+nlen1:])

	data2 := make([]byte, len(cipherbytes2)-nlen2-2)
	copy(data2, cipherbytes2[2+nlen2:])

	cipher1 := new(big.Int).SetBytes(data1)
	cipher2 := new(big.Int).SetBytes(data2)

	n1 := new(big.Int).SetBytes(nBytes1)
	nsquare := new(big.Int).Mul(n1, n1)

	res := big.NewInt(0)
	res.Mul(cipher1, cipher2).Mod(res, nsquare)

	data := make([]byte, nlen1+2+len(res.Bytes()))
	copy(data[:nlen1+2], cipherbytes1[:nlen1+2])
	copy(data[nlen1+2:], res.Bytes())

	return data, nil
}

func bytesToInt(cipherbytes []byte) int {
	bytebuff := bytes.NewBuffer(cipherbytes)
	var data int16
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}
