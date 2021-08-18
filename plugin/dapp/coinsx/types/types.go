// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
)

const (
	//CoinsActionConfig  Config transfer or manager addrs
	CoinsActionConfig = 20

	//TyCoinsxManagerStatusLog config manager status log
	TyCoinsxManagerStatusLog = 601
)

var (
	CoinsxX = "coinsx"
	// ExecerCoins execer coins
	ExecerCoins = []byte(CoinsxX)
	actionName  = map[string]int32{
		//Transfer..Genesis same as to coins, not redefine
		"Transfer":       cty.CoinsActionTransfer,
		"TransferToExec": cty.CoinsActionTransferToExec,
		"Withdraw":       cty.CoinsActionWithdraw,
		"Genesis":        cty.CoinsActionGenesis,

		//new add Config action to coinsx
		"Config": CoinsActionConfig,
	}

	clog = log.New("module", "execs.coinsx.types")
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, ExecerCoins)
	types.RegFork(CoinsxX, InitFork)
	types.RegExec(CoinsxX, InitExecutor)
}

// InitFork initials coins forks.
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(CoinsxX, "Enable", 0)
}

// InitExecutor registers coins.
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(CoinsxX, NewType(cfg))
}

// CoinsType defines exec type
type CoinsxType struct {
	types.ExecTypeBase
}

// NewType new coinstype
func NewType(cfg *types.Chain33Config) *CoinsxType {
	c := &CoinsxType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload  return payload
func (c *CoinsxType) GetPayload() types.Message {
	return &CoinsxAction{}
}

// GetName  return coins string
func (c *CoinsxType) GetName() string {
	return CoinsxX
}

// GetLogMap return log for map
func (c *CoinsxType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyCoinsxManagerStatusLog: {Ty: reflect.TypeOf(ReceiptManagerStatus{}), Name: "LogConfigManagerStatus"},
	}
}

// GetTypeMap return actionname for map
func (c *CoinsxType) GetTypeMap() map[string]int32 {
	return actionName
}

//DecodePayloadValue 为了性能考虑，coins 是最常用的合约，我们这里不用反射吗，做了特殊化的优化
func (c *CoinsxType) DecodePayloadValue(tx *types.Transaction) (string, reflect.Value, error) {
	name, value, err := c.decodePayloadValue(tx)
	return name, value, err
}

func (c *CoinsxType) decodePayloadValue(tx *types.Transaction) (string, reflect.Value, error) {
	var action CoinsxAction
	if tx.GetPayload() == nil {
		return "", reflect.ValueOf(nil), types.ErrActionNotSupport
	}
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return "", reflect.ValueOf(nil), err
	}
	var name string
	var value types.Message

	switch action.Ty {
	case cty.CoinsActionTransfer:
		name = "Transfer"
		value = action.GetTransfer()
	case cty.CoinsActionTransferToExec:
		name = "TransferToExec"
		value = action.GetTransferToExec()
	case cty.CoinsActionWithdraw:
		name = "Withdraw"
		value = action.GetWithdraw()
	case cty.CoinsActionGenesis:
		name = "Genesis"
		value = action.GetGenesis()

	case CoinsActionConfig:
		name = "Config"
		value = action.GetConfig()
	}

	if value == nil {
		return "", reflect.ValueOf(nil), types.ErrActionNotSupport
	}
	return name, reflect.ValueOf(value), nil
}

// RPC_Default_Process default process fo rpc
func (c *CoinsxType) RPC_Default_Process(action string, msg interface{}) (*types.Transaction, error) {
	var create *types.CreateTx
	if _, ok := msg.(*types.CreateTx); !ok {
		return nil, types.ErrInvalidParam
	}
	create = msg.(*types.CreateTx)
	if create.IsToken {
		return nil, types.ErrNotSupport
	}
	tx, err := c.AssertCreate(create)
	if err != nil {
		return nil, err
	}
	//to地址的问题,如果是主链交易，to地址就是直接是设置to
	types := c.GetConfig()
	if !types.IsPara() {
		tx.To = create.To
	}
	return tx, err
}

// GetAssets return asset list
func (c *CoinsxType) GetAssets(tx *types.Transaction) ([]*types.Asset, error) {
	assets, err := c.ExecTypeBase.GetAssets(tx)
	if err != nil || len(assets) == 0 {
		return nil, err
	}

	types := c.GetConfig()
	assets[0].Symbol = types.GetCoinSymbol()

	if assets[0].Symbol == "bty" {
		assets[0].Symbol = "BTY"
	}
	return assets, nil
}
