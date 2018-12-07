package types

import (
	"encoding/json"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
)

const (
	Action_CreateGame = "createGame"
	Action_MatchGame  = "matchGame"
	Action_CancelGame = "cancelGame"
	Action_CloseGame  = "closeGame"
)

func (game GameType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	tlog.Debug("Fingerguessing.CreateTx", "action", action)
	var tx *types.Transaction
	if action == Action_CreateGame {
		var param GamePreCreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return CreateRawGamePreCreateTx(&param)
	} else if action == Action_MatchGame {
		var param GamePreMatchTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return CreateRawGamePreMatchTx(&param)
	} else if action == Action_CancelGame {
		var param GamePreCancelTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return CreateRawGamePreCancelTx(&param)
	} else if action == Action_CloseGame {
		var param GamePreCloseTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			tlog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return CreateRawGamePreCloseTx(&param)
	} else {
		return nil, types.ErrNotSupport
	}

	return tx, nil
}

func CreateRawGamePreCreateTx(parm *GamePreCreateTx) (*types.Transaction, error) {
	if parm == nil {
		tlog.Error("CreateRawGamePreCreateTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}
	v := &GameCreate{
		Value:     parm.Amount,
		HashType:  parm.HashType,
		HashValue: parm.HashValue,
	}
	precreate := &FingerguessingAction{
		Ty:    GameActionCreate,
		Value: &FingerguessingAction_Create{v},
	}

	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(types.GetParaName())),
		Payload: types.Encode(precreate),
		Fee:     parm.Fee,
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	name := getRealExecName(types.GetParaName())
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func CreateRawGamePreMatchTx(parm *GamePreMatchTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}

	v := &GameMatch{
		GameId: parm.GameId,
		Guess:  parm.Guess,
	}
	game := &FingerguessingAction{
		Ty:    GameActionMatch,
		Value: &FingerguessingAction_Match{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(types.GetParaName())),
		Payload: types.Encode(game),
		Fee:     parm.Fee,
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	name := getRealExecName(types.GetParaName())
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func CreateRawGamePreCancelTx(parm *GamePreCancelTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := &GameCancel{
		GameId: parm.GameId,
	}
	cancel := &FingerguessingAction{
		Ty:    GameActionCancel,
		Value: &FingerguessingAction_Cancel{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(types.GetParaName())),
		Payload: types.Encode(cancel),
		Fee:     parm.Fee,
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	name := getRealExecName(types.GetParaName())
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

//CreateRawGamePreCloseTx
func CreateRawGamePreCloseTx(parm *GamePreCloseTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := &GameClose{
		GameId: parm.GameId,
		Secret: parm.Secret,
	}
	close := &FingerguessingAction{
		Ty:    GameActionClose,
		Value: &FingerguessingAction_Close{v},
	}
	tx := &types.Transaction{
		Execer:  []byte(getRealExecName(types.GetParaName())),
		Payload: types.Encode(close),
		Fee:     parm.Fee,
		To:      address.ExecAddress(getRealExecName(types.GetParaName())),
	}
	name := getRealExecName(types.GetParaName())
	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
