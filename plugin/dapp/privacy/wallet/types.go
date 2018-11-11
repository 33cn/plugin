// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	privacy "github.com/33cn/plugin/plugin/dapp/privacy/crypto"
	privacytypes "github.com/33cn/plugin/plugin/dapp/privacy/types"
)

type addrAndprivacy struct {
	PrivacyKeyPair *privacy.Privacy
	Addr           *string
}

// buildInputInfo 构建隐私交易输入的参数结构
type buildInputInfo struct {
	tokenname string
	sender    string
	amount    int64
	mixcount  int32
}

// txOutputInfo 存储当前钱包地址下被选中的UTXO信息
type txOutputInfo struct {
	amount           int64
	utxoGlobalIndex  *privacytypes.UTXOGlobalIndex
	txPublicKeyR     []byte
	onetimePublicKey []byte
}

type walletUTXO struct {
	height  int64
	outinfo *txOutputInfo
}

type walletUTXOs struct {
	utxos []*walletUTXO
}
