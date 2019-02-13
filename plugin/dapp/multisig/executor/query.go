// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

//Query_MultiSigAccCount 获取多重签名账户的数量，用于分批获取多重签名账户地址
//返回ReplyMultiSigAccounts
func (m *MultiSig) Query_MultiSigAccCount(in *types.ReqNil) (types.Message, error) {
	db := m.GetLocalDB()
	count, err := getMultiSigAccCount(db)
	if err != nil {
		return nil, err
	}

	return &types.Int64{Data: count}, nil
}

//Query_MultiSigAccounts 获取指定区间的多重签名账户
//输入：
//message ReqMultiSigAccs {
//	int64	start	= 1;
//	int64	end		= 2;
//输出：
//message ReplyMultiSigAccs {
//    repeated string address = 1;
func (m *MultiSig) Query_MultiSigAccounts(in *mty.ReqMultiSigAccs) (types.Message, error) {
	accountAddrs := &mty.ReplyMultiSigAccs{}

	if in.Start > in.End || in.Start < 0 {
		return nil, types.ErrInvalidParam
	}

	db := m.GetLocalDB()
	totalcount, err := getMultiSigAccCount(db)
	if err != nil {
		return nil, err
	}
	if totalcount == 0 {
		return accountAddrs, nil
	}
	if in.End >= totalcount {
		return nil, types.ErrInvalidParam
	}
	for index := in.Start; index <= in.End; index++ {
		addr, err := getMultiSigAccList(db, index)
		if err == nil {
			accountAddrs.Address = append(accountAddrs.Address, addr)
		}
	}
	return accountAddrs, nil
}

//Query_MultiSigAccountInfo 获取指定多重签名账号的状态信息
//输入：
//message ReqMultiSigAccountInfo {
//	string MultiSigAccAddr = 1;
//返回：
//message MultiSig {
//    string 							createAddr        	= 1;
//    string 							multiSigAddr      	= 2;
//    repeated Owner           			owners				= 3;
//    repeated DailyLimit          		dailyLimits   		= 4;
//    uint64           					txCount				= 5;
//	  uint64           					requiredWeight		= 6;
func (m *MultiSig) Query_MultiSigAccountInfo(in *mty.ReqMultiSigAccInfo) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	db := m.GetLocalDB()
	addr := in.MultiSigAccAddr

	if err := address.CheckMultiSignAddress(addr); err != nil {
		return nil, types.ErrInvalidAddress
	}
	multiSigAcc, err := getMultiSigAccount(db, addr)
	if err != nil {
		return nil, err
	}
	if multiSigAcc == nil {
		multiSigAcc = &mty.MultiSig{}
	}
	return multiSigAcc, nil
}

//Query_MultiSigAccTxCount 获取指定多重签名账号下的tx交易数量
//输入：
//message ReqMultiSigAccountInfo {
//	string MultiSigAccAddr = 1;
//返回：
//uint64
func (m *MultiSig) Query_MultiSigAccTxCount(in *mty.ReqMultiSigAccInfo) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	db := m.GetLocalDB()
	addr := in.MultiSigAccAddr

	if err := address.CheckMultiSignAddress(addr); err != nil {
		return nil, types.ErrInvalidAddress
	}

	multiSigAcc, err := getMultiSigAccount(db, addr)
	if err != nil {
		return nil, err
	}
	if multiSigAcc == nil {
		return nil, mty.ErrAccountHasExist
	}
	return &mty.Uint64{Data: multiSigAcc.TxCount}, nil
}

//Query_MultiSigTxids 获取txids通过设置的过滤条件和区间，pending, executed
//输入：
//message ReqMultiSigTxids {
//  string multisigaddr = 1;
//	uint64 fromtxid = 2;
//	uint64 totxid = 3;
//	bool   pending = 4;
//	bool   executed	= 5;
// 返回:
//message ReplyMultiSigTxids {
//  string 			multisigaddr = 1;
//	repeated uint64	txids		 = 2;
func (m *MultiSig) Query_MultiSigTxids(in *mty.ReqMultiSigTxids) (types.Message, error) {
	if in == nil || in.FromTxId > in.ToTxId || in.FromTxId < 0 {
		return nil, types.ErrInvalidParam
	}

	db := m.GetLocalDB()
	addr := in.MultiSigAddr

	if err := address.CheckMultiSignAddress(addr); err != nil {
		return nil, types.ErrInvalidAddress
	}
	multiSigAcc, err := getMultiSigAccount(db, addr)
	if err != nil {
		return nil, err
	}
	if multiSigAcc == nil || multiSigAcc.TxCount <= in.ToTxId {
		return nil, types.ErrInvalidParam
	}

	multiSigTxids := &mty.ReplyMultiSigTxids{}
	multiSigTxids.MultiSigAddr = addr
	for txid := in.FromTxId; txid <= in.ToTxId; txid++ {
		multiSigTx, err := getMultiSigTx(db, addr, txid)
		if err != nil || multiSigTx == nil {
			multisiglog.Error("Query_MultiSigTxids:getMultiSigTx", "addr", addr, "txid", txid, "err", err)
			continue
		}
		findTxid := txid
		//查找Pending/Executed的交易txid
		if in.Pending && !multiSigTx.Executed || in.Executed && multiSigTx.Executed {
			multiSigTxids.Txids = append(multiSigTxids.Txids, findTxid)
		}
	}
	return multiSigTxids, nil

}

//Query_MultiSigTxInfo 获取txid交易的信息，以及参与确认的owner信息
//输入:
//message ReqMultiSigTxInfo {
//  string multisigaddr = 1;
//	uint64 txid = 2;
//返回:
//message ReplyMultiSigTxInfo {
//    MultiSigTransaction multisigtxinfo = 1;
//    repeated Owner confirmowners = 3;
func (m *MultiSig) Query_MultiSigTxInfo(in *mty.ReqMultiSigTxInfo) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	db := m.GetLocalDB()
	addr := in.MultiSigAddr
	txid := in.TxId

	if err := address.CheckMultiSignAddress(addr); err != nil {
		return nil, types.ErrInvalidAddress
	}

	multiSigTx, err := getMultiSigTx(db, addr, txid)
	if err != nil {
		return nil, err
	}
	if multiSigTx == nil {
		multiSigTx = &mty.MultiSigTx{}
	} else { //由于代码中使用hex.EncodeToString()接口转换的，没有加0x，为了方便上层统一处理再次返回时增加0x即可
		multiSigTx.TxHash = "0x" + multiSigTx.TxHash
	}
	return multiSigTx, nil
}

//Query_MultiSigTxConfirmedWeight 获取txid交易已经确认的权重之和
//输入:
//message ReqMultiSigTxInfo {
//  string multisigaddr = 1;
//	uint64 txid = 2;
//返回:
//message Int64
func (m *MultiSig) Query_MultiSigTxConfirmedWeight(in *mty.ReqMultiSigTxInfo) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	db := m.GetLocalDB()
	addr := in.MultiSigAddr
	txid := in.TxId

	if err := address.CheckMultiSignAddress(addr); err != nil {
		return nil, types.ErrInvalidAddress
	}

	multiSigTx, err := getMultiSigTx(db, addr, txid)
	if err != nil {
		return nil, err
	}
	if multiSigTx == nil {
		return nil, mty.ErrTxidNotExist
	}
	var totalWeight uint64
	for _, owner := range multiSigTx.ConfirmedOwner {
		totalWeight += owner.Weight
	}

	return &mty.Uint64{Data: totalWeight}, nil
}

//Query_MultiSigAccUnSpentToday  获取指定资产当日还能使用的免多重签名的余额
//输入:
//message ReqMultiSigAccUnSpentToday {
//	string multiSigAddr = 1;
//	string execer 		= 2;
//	string symbol 		= 3;
//返回:
//message ReplyMultiSigAccUnSpentToday {
//	uint64 	amount = 1;
func (m *MultiSig) Query_MultiSigAccUnSpentToday(in *mty.ReqAccAssets) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	db := m.GetLocalDB()
	addr := in.MultiSigAddr
	isAll := in.IsAll

	if err := address.CheckMultiSignAddress(addr); err != nil {
		return nil, types.ErrInvalidAddress
	}
	multiSigAcc, err := getMultiSigAccount(db, addr)
	if err != nil {
		return nil, err
	}

	replyUnSpentAssets := &mty.ReplyUnSpentAssets{}
	if multiSigAcc == nil {
		return replyUnSpentAssets, nil
	}
	if isAll {
		for _, dailyLimit := range multiSigAcc.DailyLimits {
			var unSpentAssets mty.UnSpentAssets
			assets := &mty.Assets{
				Execer: dailyLimit.Execer,
				Symbol: dailyLimit.Symbol,
			}
			unSpentAssets.Assets = assets
			unSpentAssets.Amount = 0
			if dailyLimit.DailyLimit > dailyLimit.SpentToday {
				unSpentAssets.Amount = dailyLimit.DailyLimit - dailyLimit.SpentToday
			}
			replyUnSpentAssets.UnSpentAssets = append(replyUnSpentAssets.UnSpentAssets, &unSpentAssets)
		}
	} else {
		//assets资产合法性校验
		err := mty.IsAssetsInvalid(in.Assets.Execer, in.Assets.Symbol)
		if err != nil {
			return nil, err
		}

		for _, dailyLimit := range multiSigAcc.DailyLimits {
			var unSpentAssets mty.UnSpentAssets

			if dailyLimit.Execer == in.Assets.Execer && dailyLimit.Symbol == in.Assets.Symbol {
				assets := &mty.Assets{
					Execer: dailyLimit.Execer,
					Symbol: dailyLimit.Symbol,
				}
				unSpentAssets.Assets = assets
				unSpentAssets.Amount = 0
				if dailyLimit.DailyLimit > dailyLimit.SpentToday {
					unSpentAssets.Amount = dailyLimit.DailyLimit - dailyLimit.SpentToday
				}
				replyUnSpentAssets.UnSpentAssets = append(replyUnSpentAssets.UnSpentAssets, &unSpentAssets)
				break
			}
		}
	}
	return replyUnSpentAssets, nil
}

//Query_MultiSigAccAssets  获取多重签名账户上的所有资产，或者指定资产
//输入:
//message ReqAccAssets {
//	string multiSigAddr = 1;
//	Assets assets 		= 2;
//	bool   isAll 		= 3;
//返回:
//message MultiSigAccAssets {
//	Assets 		assets 		= 1;
//	int64   	recvAmount 	= 2;
//   Account 	account 	= 3;
func (m *MultiSig) Query_MultiSigAccAssets(in *mty.ReqAccAssets) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	//多重签名地址或者普通地址
	if err := address.CheckMultiSignAddress(in.MultiSigAddr); err != nil {
		if err = address.CheckAddress(in.MultiSigAddr); err != nil {
			return nil, types.ErrInvalidAddress
		}
	}

	replyAccAssets := &mty.ReplyAccAssets{}
	//获取账户上的所有资产数据
	if in.IsAll {
		values, err := getMultiSigAccAllAssets(m.GetLocalDB(), in.MultiSigAddr)
		if err != nil {
			return nil, err
		}
		if len(values) != 0 {
			for _, value := range values {
				reciver := mty.AccountAssets{}
				err = types.Decode(value, &reciver)
				if err != nil {
					continue
				}
				accAssets := &mty.AccAssets{}
				account, err := m.getMultiSigAccAssets(reciver.MultiSigAddr, reciver.Assets)
				if err != nil {
					multisiglog.Error("Query_MultiSigAccAssets:getMultiSigAccAssets", "MultiSigAddr", reciver.MultiSigAddr, "err", err)
				}
				accAssets.Account = account
				accAssets.Assets = reciver.Assets
				accAssets.RecvAmount = reciver.Amount

				replyAccAssets.AccAssets = append(replyAccAssets.AccAssets, accAssets)
			}
		}
	} else { //获取账户上的指定资产数据
		accAssets := &mty.AccAssets{}
		//assets资产合法性校验
		err := mty.IsAssetsInvalid(in.Assets.Execer, in.Assets.Symbol)
		if err != nil {
			return nil, err
		}
		account, err := m.getMultiSigAccAssets(in.MultiSigAddr, in.Assets)
		if err != nil {
			multisiglog.Error("Query_MultiSigAccAssets:getMultiSigAccAssets", "MultiSigAddr", in.MultiSigAddr, "err", err)
		}
		accAssets.Account = account
		accAssets.Assets = in.Assets

		amount, err := getAddrReciver(m.GetLocalDB(), in.MultiSigAddr, in.Assets.Execer, in.Assets.Symbol)
		if err != nil {
			multisiglog.Error("Query_MultiSigAccAssets:getAddrReciver", "MultiSigAddr", in.MultiSigAddr, "err", err)
		}
		accAssets.RecvAmount = amount

		replyAccAssets.AccAssets = append(replyAccAssets.AccAssets, accAssets)
	}

	return replyAccAssets, nil
}

//Query_MultiSigAccAllAddress 获取指定地址创建的所有多重签名账户
//输入:
//createaddr
//返回:
//[]string
func (m *MultiSig) Query_MultiSigAccAllAddress(in *mty.ReqMultiSigAccInfo) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	if err := address.CheckAddress(in.MultiSigAccAddr); err != nil {
		return nil, types.ErrInvalidAddress
	}
	return getMultiSigAccAllAddress(m.GetLocalDB(), in.MultiSigAccAddr)
}
