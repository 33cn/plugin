// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
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

//计算隐私资产utxo的前缀, 和exec,token相关
func calcUtxoAssetPrefix(exec, token string) string {
	//只有coins资产的key不加exec前缀, 主要考虑是不加分叉兼容历史隐私交易
	if exec == "" || exec == types.DefaultCoinsExec {
		return token
	}
	return exec + ":" + token
}

//CalcPrivacyOutputKey 该key对应的是types.KeyOutput
//该kv会在store中设置
func CalcPrivacyOutputKey(exec, token string, amount int64, txhash string, outindex int) (key []byte) {
	return []byte(fmt.Sprintf(privacyOutputKeyPrefix+"-%s-%d-%s-%d", calcUtxoAssetPrefix(exec, token), amount, txhash, outindex))
}

func calcPrivacyKeyImageKey(exec, token string, keyimage []byte) []byte {
	return []byte(fmt.Sprintf(privacyKeyImagePrefix+"-%s-%s", calcUtxoAssetPrefix(exec, token), common.ToHex(keyimage)))
}

//CalcPrivacyUTXOkeyHeight 在本地数据库中设置一条可以找到对应amount的对应的utxo的global index
func CalcPrivacyUTXOkeyHeight(exec, token string, amount, height int64, txhash string, txindex, outindex int) (key []byte) {
	return []byte(fmt.Sprintf(privacyUTXOKEYPrefix+"-%s-%s-%d-%d-%s-%d-%d", exec, token, amount, height, txhash, txindex, outindex))
}

// CalcPrivacyUTXOkeyHeightPrefix get privacy utxo key by height and prefix
func CalcPrivacyUTXOkeyHeightPrefix(exec, token string, amount int64) (key []byte) {
	return []byte(fmt.Sprintf(privacyUTXOKEYPrefix+"-%s-%s-%d-", exec, token, amount))
}

//CalcprivacyKeyTokenAmountType 设置当前系统存在的token的amount的类型，如存在1,3,5,100...等等的类型,
func CalcprivacyKeyTokenAmountType(exec, token string) (key []byte) {
	return []byte(fmt.Sprintf(privacyAmountTypePrefix+"-%s-%s-", exec, token))
}

// CalcprivacyKeyTokenTypes get privacy token types key
func CalcprivacyKeyTokenTypes() (key []byte) {
	return []byte(privacyTokenTypesPrefix)
}

func calcExecLocalAssetKey(exec, symbol string) string {
	return exec + "-" + symbol
}
