// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cert/authority"
	ct "github.com/33cn/plugin/plugin/dapp/cert/types"
)

func calcCertHeightKey(height int64) []byte {
	return []byte(fmt.Sprintf("LODB-cert-%d", height))
}

// ExecLocal_New 启用证书交易执行
func (c *Cert) ExecLocal_New(payload *ct.CertNew, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if !authority.IsAuthEnable {
		clog.Error("Authority is not available. Please check the authority config or authority initialize error logs.")
		return nil, ct.ErrInitializeAuthority
	}
	var set types.LocalDBSet

	historityCertdata := &types.HistoryCertStore{}
	authority.Author.HistoryCertCache.CurHeight = c.GetHeight()
	authority.Author.HistoryCertCache.ToHistoryCertStore(historityCertdata)
	key := calcCertHeightKey(c.GetHeight())
	set.KV = append(set.KV, &types.KeyValue{
		Key:   key,
		Value: types.Encode(historityCertdata),
	})

	// 构造非证书历史数据
	noneCertdata := &types.HistoryCertStore{}
	noneCertdata.NxtHeight = historityCertdata.CurHeigth
	noneCertdata.CurHeigth = 0
	set.KV = append(set.KV, &types.KeyValue{
		Key:   calcCertHeightKey(0),
		Value: types.Encode(noneCertdata),
	})

	return &set, nil
}

// ExecLocal_Update 更新证书交易执行
func (c *Cert) ExecLocal_Update(payload *ct.CertUpdate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if !authority.IsAuthEnable {
		clog.Error("Authority is not available. Please check the authority config or authority initialize error logs.")
		return nil, ct.ErrInitializeAuthority
	}
	var set types.LocalDBSet

	// 写入上一纪录的next-height
	key := calcCertHeightKey(authority.Author.HistoryCertCache.CurHeight)
	historityCertdata := &types.HistoryCertStore{}
	authority.Author.HistoryCertCache.NxtHeight = c.GetHeight()
	authority.Author.HistoryCertCache.ToHistoryCertStore(historityCertdata)
	set.KV = append(set.KV, &types.KeyValue{
		Key:   key,
		Value: types.Encode(historityCertdata),
	})

	// 证书更新
	historityCertdata = &types.HistoryCertStore{}
	err := authority.Author.ReloadCertByHeght(c.GetHeight())
	if err != nil {
		return nil, err
	}

	authority.Author.HistoryCertCache.ToHistoryCertStore(historityCertdata)
	setKey := calcCertHeightKey(c.GetHeight())
	set.KV = append(set.KV, &types.KeyValue{
		Key:   setKey,
		Value: types.Encode(historityCertdata),
	})
	return &set, nil
}

// ExecLocal_Normal 非证书变更交易执行
func (c *Cert) ExecLocal_Normal(payload *ct.CertNormal, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if !authority.IsAuthEnable {
		clog.Error("Authority is not available. Please check the authority config or authority initialize error logs.")
		return nil, ct.ErrInitializeAuthority
	}
	var set types.LocalDBSet

	return &set, nil
}
