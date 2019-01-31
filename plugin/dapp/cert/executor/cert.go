// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cert/authority"
	ct "github.com/33cn/plugin/plugin/dapp/cert/types"
)

var clog = log.New("module", "execs.cert")
var driverName = ct.CertX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Cert{}))
}

// Init 初始化
func Init(name string, sub []byte) {
	driverName = name
	var cfg ct.Authority
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	err := authority.Author.Init(&cfg)
	if err != nil {
		clog.Error("error to initialize authority", err)
		return
	}
	drivers.Register(driverName, newCert, types.GetDappFork(driverName, "Enable"))
}

// GetName 获取cert执行器名
func GetName() string {
	return newCert().GetName()
}

// Cert cert执行器
type Cert struct {
	drivers.DriverBase
}

func newCert() drivers.Driver {
	c := &Cert{}
	c.SetChild(c)
	c.SetIsFree(true)
	return c
}

// GetDriverName 获取cert执行器名
func (c *Cert) GetDriverName() string {
	return driverName
}

// CheckTx cert执行器tx证书校验
func (c *Cert) CheckTx(tx *types.Transaction, index int) error {
	// 基类检查
	err := c.DriverBase.CheckTx(tx, index)
	if err != nil {
		return err
	}

	// auth模块关闭则返回
	if !authority.IsAuthEnable {
		clog.Error("Authority is not available. Please check the authority config or authority initialize error logs.")
		return ct.ErrInitializeAuthority
	}

	// 重启
	if authority.Author.HistoryCertCache.CurHeight == -1 {
		err := c.loadHistoryByPrefix()
		if err != nil {
			return err
		}
	}

	// 当前区块<上次证书变更区块，cert回滚
	if c.GetHeight() <= authority.Author.HistoryCertCache.CurHeight {
		err := c.loadHistoryByPrefix()
		if err != nil {
			return err
		}
	}

	// 当前区块>上次变更下一区块，下一区块不为-1，即非最新证书变更记录，用于cert回滚时判断是否到了下一变更记录
	nxtHeight := authority.Author.HistoryCertCache.NxtHeight
	if nxtHeight != -1 && c.GetHeight() > nxtHeight {
		err := c.loadHistoryByHeight()
		if err != nil {
			return err
		}
	}

	// auth校验
	return authority.Author.Validate(tx.GetSignature())
}

/**
根据前缀查找证书变更记录，cert回滚、重启、同步用到
*/
func (c *Cert) loadHistoryByPrefix() error {
	parm := &types.LocalDBList{
		Prefix:    []byte("LODB-cert-"),
		Key:       nil,
		Direction: 0,
		Count:     0,
	}
	result, err := c.DriverBase.GetAPI().LocalList(parm)
	if err != nil {
		return err
	}

	// 数据库没有变更记录，说明创世区块开始cert校验
	if len(result.Values) == 0 {
		authority.Author.HistoryCertCache.CurHeight = 0
		return nil
	}

	// 寻找当前高度使用的证书区间
	var historyData types.HistoryCertStore
	for _, v := range result.Values {
		err := types.Decode(v, &historyData)
		if err != nil {
			return err
		}
		if historyData.CurHeigth < c.GetHeight() && (historyData.NxtHeight >= c.GetHeight() || historyData.NxtHeight == -1) {
			return authority.Author.ReloadCert(&historyData)
		}
	}

	return ct.ErrGetHistoryCertData
}

/**
根据具体高度查找变更记录，cert回滚用到
*/
func (c *Cert) loadHistoryByHeight() error {
	key := calcCertHeightKey(c.GetHeight())
	parm := &types.LocalDBGet{Keys: [][]byte{key}}
	result, err := c.DriverBase.GetAPI().LocalGet(parm)
	if err != nil {
		return err
	}
	var historyData types.HistoryCertStore
	for _, v := range result.Values {
		err := types.Decode(v, &historyData)
		if err != nil {
			return err
		}
		if historyData.CurHeigth < c.GetHeight() && historyData.NxtHeight >= c.GetHeight() {
			return authority.Author.ReloadCert(&historyData)
		}
	}
	return ct.ErrGetHistoryCertData
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Cert) CheckReceiptExecOk() bool {
	return true
}
