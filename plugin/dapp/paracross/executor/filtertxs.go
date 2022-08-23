// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

//1,如果全部是paracross的，主链成功会ExecOk，如果有一个不成功，全部回退回PACK，后面检查TyLogErr，OK意味着全部成功
//2,如果是paracross+other， other的是PACK，如果有一个是OK，那意味着全部OK，如果全部是PACK，检查TyLogErr
//3,如果是全部other，全部是PACK
func checkReceiptExecOk(receipt *types.ReceiptData) bool {
	if receipt.Ty == types.ExecOk {
		return true
	}
	//如果主链allow 平行链tx 主链执行出错场景 比如paracross
	for _, log := range receipt.Logs {
		if log.Ty == types.TyLogErr {
			return false
		}
	}
	return true
}

//1. 如果涉及跨链合约，如果有超过两条平行链的交易被判定为失败，交易组会执行不成功,也不PACK。（这样的情况下，主链交易一定会执行不成功,最终也不会进到block里面）
//2. 跨链合约交易组，要么是paracross+user.p.xx.paracross组合，要么全是user.p.xx.paracross组合，后面是资产转移
//3. 如果交易组有一个ExecOk,主链上的交易都是ok的，可以全部打包
//4. 不论是否涉及跨链合约, 不同用途的tx打到一个group里面，如果主链交易有失败，平行链也不会执行，也需要排除掉
//5. 如果全部是ExecPack，有两种情况，一是交易组所有交易都是平行链交易，另一是主链有交易失败而打包了的交易，需要检查LogErr，如果有错，全部不打包
//经para filter之后， 交易组会存在如下几种tx：
// 1, 主链	paracross	+  	平行链  user.p.xx.paracross  跨链兑换合约
// 2, 主链   paracross	+  	平行链  user.p.xx.other 		混合交易组合
// 3, 主链   other  		+ 	平行链  user.p.xx.paracross 	混合交易组合约
// 4, 主链 	other 		+ 	平行链  user.p.xx.other 		混合交易组合
// 5, 主链+平行链 user.p.xx.paracross 交易组				混合跨链资产转移
// 6, 平行链	    user.p.xx.paracross + user.p.xx.other   混合平行链组合
// 7, 平行链     all user.p.xx.other  					混合平行链组合
///// 分叉以后只考虑平行链交易组全部是平行链tx，没有主链tx
//经para filter之后， 交易组会存在如下几种tx：
// 1, 主链+平行链 user.p.xx.paracross 交易组				混合跨链资产转移  paracross主链执行成功
// 2, 平行链	    user.p.xx.paracross + user.p.xx.other   混合平行链组合    paracross主链执行成功
// 3, 平行链     user.p.xx.other  交易组					混合平行链组合    other主链pack
func filterParaTxGroup(cfg *types.Chain33Config, tx *types.Transaction, allTxs []*types.TxDetail, index int, mainBlockHeight, forkHeight int64) ([]*types.Transaction, int) {
	var headIdx int

	for i := index; i >= 0; i-- {
		if bytes.Equal(tx.Header, allTxs[i].Tx.Hash()) {
			headIdx = i
			break
		}
	}

	endIdx := headIdx + int(tx.GroupCount)
	for i := headIdx; i < endIdx; i++ {
		//缺省是在forkHeight之前与更老版本一致，不检查平行链交易,但有些特殊平行链6.2.0版本升级上来无更老版本且要求blockhash不变，则需与6.2.0保持一致，不检查
		if cfg.IsPara() && mainBlockHeight < forkHeight && !types.Conf(cfg, pt.ParaPrefixConsSubConf).IsEnable(pt.ParaFilterIgnoreTxGroup) {
			if types.IsParaExecName(string(allTxs[i].Tx.Execer)) {
				continue
			}
		}

		if !checkReceiptExecOk(allTxs[i].Receipt) {
			clog.Error("filterParaTxGroup rmv tx group", "txhash", hex.EncodeToString(allTxs[i].Tx.Hash()))
			return nil, endIdx
		}
	}
	//全部是平行链交易 或平行链在主链执行成功的tx
	var retTxs []*types.Transaction
	for _, retTx := range allTxs[headIdx:endIdx] {
		retTxs = append(retTxs, retTx.Tx)
	}
	return retTxs, endIdx
}

//FilterTxsForPara include some main tx in tx group before ForkParacrossCommitTx
func FilterTxsForPara(cfg *types.Chain33Config, main *types.ParaTxDetail) []*types.Transaction {
	cfgPara := types.ConfSub(cfg, pt.ParaX)
	discardTxs := cfgPara.GStrList("discardTxs")
	discardTxsMap := make(map[string]bool)
	for _, v := range discardTxs {
		if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
			v = v[2:]
		}
		discardTxsMap[v] = true
	}
	var txs []*types.Transaction
	forkHeight := pt.GetDappForkHeight(cfg, pt.ForkCommitTx)
	for i := 0; i < len(main.TxDetails); i++ {
		tx := main.TxDetails[i].Tx
		if tx.GroupCount >= 2 {
			mainTxs, endIdx := filterParaTxGroup(cfg, tx, main.TxDetails, i, main.Header.Height, forkHeight)
			txs = append(txs, mainTxs...)
			i = endIdx - 1
			continue
		}
		//单独的paracross tx 如果主链执行失败也要排除, 6.2fork原因 没有排除 非user.p.xx.paracross的平行链交易
		if main.Header.Height >= forkHeight && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) && !checkReceiptExecOk(main.TxDetails[i].Receipt) {
			clog.Info("FilterTxsForPara rmv tx", "txhash", hex.EncodeToString(tx.Hash()))
			continue
		}

		if discardTxsMap[hex.EncodeToString(tx.Hash())] {
			clog.Info("FilterTxsForPara discard tx", "txhash", common.ToHex(tx.Hash()))
			continue
		}

		txs = append(txs, tx)
	}
	return txs
}

// FilterParaCrossTxs only all para chain cross txs like xx.paracross exec
func FilterParaCrossTxs(txs []*types.Transaction) []*types.Transaction {
	var paraCrossTxs []*types.Transaction
	for _, tx := range txs {
		if types.IsParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
			paraCrossTxs = append(paraCrossTxs, tx)
		}
	}
	return paraCrossTxs
}

// FilterParaCrossTxHashes only all para chain cross txs like xx.paracross exec
func FilterParaCrossTxHashes(txs []*types.Transaction) [][]byte {
	var txHashs [][]byte
	for _, tx := range txs {
		if types.IsParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
			txHashs = append(txHashs, tx.Hash())
		}
	}
	return txHashs
}

// FilterParaCrossAssetTxHashes 只过滤跨链资产转移的类型
func FilterParaCrossAssetTxHashes(txs []*types.Transaction) ([][]byte, error) {
	var txHashs [][]byte
	for _, tx := range txs {
		if types.IsParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
			var payload pt.ParacrossAction
			err := types.Decode(tx.Payload, &payload)
			if err != nil {
				clog.Error("FilterParaCrossAssetTxHashes decode tx", "txhash", hex.EncodeToString(tx.Hash()), "err", err.Error())
				return nil, err
			}
			if payload.Ty == pt.ParacrossActionAssetTransfer ||
				payload.Ty == pt.ParacrossActionAssetWithdraw ||
				payload.Ty >= pt.ParacrossActionCrossAssetTransfer {
				txHashs = append(txHashs, tx.Hash())
			}

		}
	}
	return txHashs, nil
}

//经para filter之后， 交易组会存在如下几种tx：
// 1, 主链	paracross	+  	平行链  user.p.xx.paracross  跨链兑换合约
// 2, 主链   paracross	+  	平行链  user.p.xx.other 		混合交易组合
// 3, 主链   other  		+ 	平行链  user.p.xx.paracross 	混合交易组合
// 4, 主链 	other 		+ 	平行链  user.p.xx.other 		混合交易组合
// 5, 主链+平行链 user.p.xx.paracross 交易组				混合跨链资产转移
// 6, 平行链	    user.p.xx.paracross + user.p.xx.other   混合平行链组合
// 7, 平行链     all user.p.xx.other  					混合平行链组合
// 这里只取跨链兑换和任何有user.p.xx.paracross的资产转移交易，资产兑换可能主链会需要查看平行链执行结果再对主链的paracross合约做后续处理
func crossTxGroupProc(title string, txs []*types.Transaction, index int) ([]*types.Transaction, int32) {
	var headIdx, endIdx int32

	for i := index; i >= 0; i-- {
		if bytes.Equal(txs[index].Header, txs[i].Hash()) {
			headIdx = int32(i)
			break
		}
	}
	//cross mix tx, contain main and para tx, main prefix with pt.paraX
	//最初设计是主链平行链跨链交换，都在paracross合约处理，平行链在主链共识结束后主链做unfreeze操作，但是那样出错时候回滚不好处理
	//目前只设计跨链转移场景，转移到平行链通过trade交换
	endIdx = headIdx + txs[index].GroupCount
	for i := headIdx; i < endIdx; i++ {
		if bytes.HasPrefix(txs[i].Execer, []byte(pt.ParaX)) {
			return txs[headIdx:endIdx], endIdx
		}
	}
	//cross asset transfer in tx group
	var transfers []*types.Transaction
	for i := headIdx; i < endIdx; i++ {
		if types.IsSpecificParaExecName(title, string(txs[i].Execer)) && bytes.HasSuffix(txs[i].Execer, []byte(pt.ParaX)) {
			transfers = append(transfers, txs[i])

		}
	}
	return transfers, endIdx

}

//FilterParaMainCrossTxHashes ForkParacrossCommitTx之前允许txgroup里面有main chain tx的跨链
func FilterParaMainCrossTxHashes(title string, txs []*types.Transaction) [][]byte {
	var crossTxHashs [][]byte
	//跨链tx 必须是paracross合约且user.p.打头， user.p.xx.的非paracross合约不是跨链
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		if tx.GroupCount > 1 {
			groupTxs, end := crossTxGroupProc(title, txs, i)
			for _, tx := range groupTxs {
				crossTxHashs = append(crossTxHashs, tx.Hash())

			}
			i = int(end) - 1
			continue
		}
		if types.IsSpecificParaExecName(title, string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
			crossTxHashs = append(crossTxHashs, tx.Hash())
		}
	}
	return crossTxHashs

}

//CalcTxHashsHash 计算几个txhash的hash值 作校验使用
func CalcTxHashsHash(txHashs [][]byte) []byte {
	if len(txHashs) == 0 {
		return nil
	}
	totalTxHash := &types.ReqHashes{}
	totalTxHash.Hashes = append(totalTxHash.Hashes, txHashs...)
	data := types.Encode(totalTxHash)
	return common.Sha256(data)
}
