// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"fmt"

	"github.com/33cn/chain33/common/address"
)

const (
	prefix = "MixCoin-"
	// PrivacyDBVersion 隐私交易运行过程中，需要使用到钱包数据库存储的数据库版本信息的KEY值
	MixDBVersion = prefix + "DBVersion"
	// Privacy4Addr 存储隐私交易保存账户的隐私公钥对信息的KEY值
	// KEY值格式为  	Privacy4Addr-账号地址
	// VALUE值格式为 types.WalletAccountPrivacy， 存储隐私公钥对
	Mix4Addr = prefix + "Addr"
	//
	MixPrivacyEnable = prefix + "PrivacyEnable"
	//current rescan notes status
	MixRescanStatus = prefix + "RescanStatus"
	MixCommitHash   = prefix + "CommitHash"
	MixNullifier    = prefix + "Nullifier"
)

// calcPrivacyAddrKey 获取隐私账户私钥对保存在钱包中的索引串
func calcMixAddrKey(addr string) []byte {
	return []byte(fmt.Sprintf("%s-%s", Mix4Addr, address.FormatAddrKey(addr)))
}

func calcMixPrivacyEnable() []byte {
	return []byte(MixPrivacyEnable)
}

func calcRescanNoteStatus() []byte {
	return []byte(MixRescanStatus)
}
