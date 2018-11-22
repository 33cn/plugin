// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common"
)

const (
	privacyOutputKeyPrefix  = "mavl-privacy-UTXO-tahi"
	privacyKeyImagePrefix   = "mavl-privacy-UTXO-keyimage"
	privacyUTXOKEYPrefix    = "LODB-privacy-UTXO-tahhi"
	privacyAmountTypePrefix = "LODB-privacy-UTXO-atype"
	privacyTokenTypesPrefix = "LODB-privacy-UTXO-token"
	keyImageSpentAlready    = 0x01
	invalidIndex            = -1
)

//CalcPrivacyOutputKey 该key对应的是types.KeyOutput
//该kv会在store中设置
func CalcPrivacyOutputKey(token string, amount int64, txhash string, outindex int) (key []byte) {
	return []byte(fmt.Sprintf(privacyOutputKeyPrefix+"-%s-%d-%s-%d", token, amount, txhash, outindex))
}

func calcPrivacyKeyImageKey(token string, keyimage []byte) []byte {
	return []byte(fmt.Sprintf(privacyKeyImagePrefix+"-%s-%s", token, common.ToHex(keyimage)))
}

//CalcPrivacyUTXOkeyHeight 在本地数据库中设置一条可以找到对应amount的对应的utxo的global index
func CalcPrivacyUTXOkeyHeight(token string, amount, height int64, txhash string, txindex, outindex int) (key []byte) {
	return []byte(fmt.Sprintf(privacyUTXOKEYPrefix+"-%s-%d-%d-%s-%d-%d", token, amount, height, txhash, txindex, outindex))
}

// CalcPrivacyUTXOkeyHeightPrefix get privacy utxo key by height and prefix
func CalcPrivacyUTXOkeyHeightPrefix(token string, amount int64) (key []byte) {
	return []byte(fmt.Sprintf(privacyUTXOKEYPrefix+"-%s-%d-", token, amount))
}

//CalcprivacyKeyTokenAmountType 设置当前系统存在的token的amount的类型，如存在1,3,5,100...等等的类型,
func CalcprivacyKeyTokenAmountType(token string) (key []byte) {
	return []byte(fmt.Sprintf(privacyAmountTypePrefix+"-%s-", token))
}

// CalcprivacyKeyTokenTypes get privacy token types key
func CalcprivacyKeyTokenTypes() (key []byte) {
	return []byte(privacyTokenTypesPrefix)
}
