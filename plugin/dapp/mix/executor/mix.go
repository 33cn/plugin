// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

var (
	mlog       = log.New("module", "execs.mix")
	driverName = mixTy.MixX
)

// Mix exec
type Mix struct {
	drivers.DriverBase
}

//Init paracross exec register
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newMix, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
	setPrefix()
}

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Mix{}))
}

//GetName return paracross name
func GetName() string {
	return newMix().GetName()
}

func newMix() drivers.Driver {
	c := &Mix{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetDriverName return paracross driver name
func (m *Mix) GetDriverName() string {
	return mixTy.MixX
}

// CheckTx check transaction
func (m *Mix) CheckTx(tx *types.Transaction, index int) error {
	action := new(mixTy.MixAction)
	if err := types.Decode(tx.Payload, action); err != nil {
		mlog.Error("CheckTx decode", "err", err)
		return err
	}
	if action.Ty != mixTy.MixActionTransfer {
		// mix隐私交易，只私对私需要特殊签名验证
		return m.DriverBase.CheckTx(tx, index)
	}

	_, _, err := MixTransferInfoVerify(m.GetAPI().GetConfig(), m.GetStateDB(), action.GetTransfer())
	if err != nil {
		mlog.Error("checkTx", "err", err, "txhash", common.ToHex(tx.Hash()))
		return err
	}
	return nil

}
