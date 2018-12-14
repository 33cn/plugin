// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
)

//数据库存储格式key
const (
	//MultiSigPrefix statedb中账户和交易的存储格式
	MultiSigPrefix   = "mavl-multisig-"
	MultiSigTxPrefix = "mavl-multisig-tx-"

	//MultiSigLocalPrefix localdb中账户和交易的存储格式multisig account count记录账户个数
	MultiSigLocalPrefix = "LODB-multisig-"
	MultiSigAccCount    = "acccount"
	MultiSigAcc         = "account"
	MultiSigAllAcc      = "allacc"
	MultiSigTx          = "tx"
	MultiSigRecvAssets  = "assets"
	MultiSigAccCreate   = "create"
)

//statedb中账户和交易的存储格式
func calcMultiSigAccountKey(multiSigAccAddr string) (key []byte) {
	return []byte(fmt.Sprintf(MultiSigPrefix+"%s", multiSigAccAddr))
}

//存储格式："mavl-multisig-tx-accaddr-000000000000"
func calcMultiSigAccTxKey(multiSigAccAddr string, txid uint64) (key []byte) {
	txstr := fmt.Sprintf("%018d", txid)
	return []byte(fmt.Sprintf(MultiSigTxPrefix+"%s-%s", multiSigAccAddr, txstr))
}

//localdb中账户相关的存储格式

//记录创建的账户数量，key:Msac value：count。
func calcMultiSigAccCountKey() []byte {
	return []byte(fmt.Sprintf(MultiSigLocalPrefix+"%s", MultiSigAccCount))
}

//存储所有的MultiSig账户地址：按顺序存储方便以后分页查找: key:Ms:allacc:index,value:accaddr
func calcMultiSigAllAcc(accindex int64) (key []byte) {
	accstr := fmt.Sprintf("%018d", accindex)
	return []byte(fmt.Sprintf(MultiSigLocalPrefix+"%s-%s", MultiSigAllAcc, accstr))
}

//记录指定账号地址的信息value：MultiSig。key:Ms:acc
func calcMultiSigAcc(addr string) (key []byte) {
	return []byte(fmt.Sprintf(MultiSigLocalPrefix+"%s-%s", MultiSigAcc, addr))
}

//记录某个地址创建的所有多重签名账户。key:Ms:create:createAddr，value：[]string。
func calcMultiSigAccCreateAddr(createAddr string) (key []byte) {
	return []byte(fmt.Sprintf(MultiSigLocalPrefix+"%s:-%s", MultiSigAccCreate, createAddr))
}

//localdb中账户相关的存储格式

//记录指定账号地址的信息key:Ms:tx:addr:txid  value：MultiSigTx。
func calcMultiSigAccTx(addr string, txid uint64) (key []byte) {
	accstr := fmt.Sprintf("%018d", txid)

	return []byte(fmt.Sprintf(MultiSigLocalPrefix+"%s-%s-%s", MultiSigTx, addr, accstr))
}

//可以通过前缀查找获取指定账户上收到的所有资产数量
//MultiSig合约中账户收到指定资产的计数key:Ms:assets:addr:execname:symbol  value：AccountAssets。
//message AccountAssets {
//	string multiSigAddr = 1;
//	string execer 		= 2;
//	string symbol 		= 3;
//	int64  amount 		= 4;
func calcAddrRecvAmountKey(addr, execname, symbol string) []byte {
	return []byte(fmt.Sprintf(MultiSigLocalPrefix+"%s-%s-%s-%s", MultiSigRecvAssets, addr, execname, symbol))
}

// 前缀查找某个账户下的所有资产信息
func calcAddrRecvAmountPrefix(addr string) []byte {
	return []byte(fmt.Sprintf(MultiSigLocalPrefix+"%s-%s-", MultiSigRecvAssets, addr))
}
