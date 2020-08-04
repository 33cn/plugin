// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"reflect"

	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var tlog = log.New("module", GameX)

func init() {
	// init executor type
	types.AllowUserExec = append(types.AllowUserExec, []byte(GameX))
	types.RegFork(GameX, InitFork)
	types.RegExec(GameX, InitExecutor)
}

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(GameX, "Enable", 0)
}

//InitExecutor ...
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(GameX, NewType(cfg))
}

//getRealExecName
//如果paraName == "", 那么自动用 types.ExecName("game")
//如果设置了paraName , 那么强制用paraName
//也就是说，我们可以构造其他平行链的交易
func getRealExecName(cfg *types.Chain33Config, paraName string) string {
	return cfg.ExecName(paraName + GameX)
}

// NewType  new type
func NewType(cfg *types.Chain33Config) *GameType {
	c := &GameType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GameType execType
type GameType struct {
	types.ExecTypeBase
}

// GetName 获取执行器名称
func (gt *GameType) GetName() string {
	return GameX
}

// GetLogMap get log
func (gt *GameType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogCreateGame: {Ty: reflect.TypeOf(ReceiptGame{}), Name: "LogCreateGame"},
		TyLogCancleGame: {Ty: reflect.TypeOf(ReceiptGame{}), Name: "LogCancleGame"},
		TyLogMatchGame:  {Ty: reflect.TypeOf(ReceiptGame{}), Name: "LogMatchGame"},
		TyLogCloseGame:  {Ty: reflect.TypeOf(ReceiptGame{}), Name: "LogCloseGame"},
	}
}

// GetPayload get payload
func (gt *GameType) GetPayload() types.Message {
	return &GameAction{}
}

// GetTypeMap get typeMap
func (gt *GameType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create": GameActionCreate,
		"Cancel": GameActionCancel,
		"Close":  GameActionClose,
		"Match":  GameActionMatch,
	}
}

// CreateTx  unused,just empty implementation
func (gt GameType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	tlog.Debug("Game.CreateTx", "action", action)
	cfg := gt.GetConfig()
	if action == ActionCreateGame {
		var param GamePreCreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return CreateRawGamePreCreateTx(cfg, &param)
	} else if action == ActionMatchGame {
		var param GamePreMatchTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return CreateRawGamePreMatchTx(cfg, &param)
	} else if action == ActionCancelGame {
		var param GamePreCancelTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return CreateRawGamePreCancelTx(cfg, &param)
	} else if action == ActionCloseGame {
		var param GamePreCloseTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return CreateRawGamePreCloseTx(cfg, &param)
	}
	return nil, types.ErrNotSupport
}

// CreateRawGamePreCreateTx  unused,just empty implementation
func CreateRawGamePreCreateTx(cfg *types.Chain33Config, parm *GamePreCreateTx) (*types.Transaction, error) {
	if parm == nil {
		tlog.Error("CreateRawGamePreCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	v := &GameCreate{
		Value:     parm.Amount,
		HashType:  parm.HashType,
		HashValue: parm.HashValue,
	}
	precreate := &GameAction{
		Ty:    GameActionCreate,
		Value: &GameAction_Create{v},
	}

	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(cfg, cfg.GetParaName())),
		Payload: types.Encode(precreate),
		Fee:     parm.Fee,
		To:      address.ExecAddress(getRealExecName(cfg, cfg.GetParaName())),
	}
	name := getRealExecName(cfg, cfg.GetParaName())
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawGamePreMatchTx  unused,just empty implementation
func CreateRawGamePreMatchTx(cfg *types.Chain33Config, parm *GamePreMatchTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}

	v := &GameMatch{
		GameId: parm.GameID,
		Guess:  parm.Guess,
	}
	game := &GameAction{
		Ty:    GameActionMatch,
		Value: &GameAction_Match{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(cfg, cfg.GetParaName())),
		Payload: types.Encode(game),
		Fee:     parm.Fee,
		To:      address.ExecAddress(getRealExecName(cfg, cfg.GetParaName())),
	}
	name := getRealExecName(cfg, cfg.GetParaName())
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawGamePreCancelTx  unused,just empty implementation
func CreateRawGamePreCancelTx(cfg *types.Chain33Config, parm *GamePreCancelTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := &GameCancel{
		GameId: parm.GameID,
	}
	cancel := &GameAction{
		Ty:    GameActionCancel,
		Value: &GameAction_Cancel{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(cfg, cfg.GetParaName())),
		Payload: types.Encode(cancel),
		Fee:     parm.Fee,
		To:      address.ExecAddress(getRealExecName(cfg, cfg.GetParaName())),
	}
	name := getRealExecName(cfg, cfg.GetParaName())
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawGamePreCloseTx  unused,just empty implementation
func CreateRawGamePreCloseTx(cfg *types.Chain33Config, parm *GamePreCloseTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := &GameClose{
		GameId: parm.GameID,
		Secret: parm.Secret,
	}
	close := &GameAction{
		Ty:    GameActionClose,
		Value: &GameAction_Close{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(cfg, cfg.GetParaName())),
		Payload: types.Encode(close),
		Fee:     parm.Fee,
		To:      address.ExecAddress(getRealExecName(cfg, cfg.GetParaName())),
	}
	name := getRealExecName(cfg, cfg.GetParaName())
	tx, err := types.FormatTx(cfg, name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
