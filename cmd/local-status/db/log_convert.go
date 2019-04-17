package db

import (
	"github.com/33cn/chain33/types"
	"errors"
	"encoding/json"
	rpcTypes "github.com/33cn/chain33/rpc/types"
)

type Log2KV interface {
	Convert(logType int64, json string) (key []string, prev, current  []byte, err error)
}

func NewConvert(exec string, detail *rpcTypes.BlockDetail) Log2KV {
	if exec == "ticket" {
		return &ticketConvert{block: detail}
	} else if exec == "coins" {
		return &coinsConvert{block: detail}
	}
	return nil
}

func notSupport (logType int64, json []byte) (key []string, prev, current  []byte, err error) {
	return nil, nil, nil, errors.New("notSupport")
}

func convertFailed () (key []string, prev, current  []byte, err error) {
	return nil, nil, nil, errors.New("convertFailed")
}

func CommonConverts(ty int64, v []byte) (key []string, prev, current  []byte, err error) {
	if ty == types.TyLogFee {
		return LogFeeConvert(v)
	} else if ty == types.TyLogTransfer {
		return LogTransferConvert(v)
	} else if ty == types.TyLogDeposit {
		return LogDepositConvert(v)
	} else if ty == types.TyLogExecTransfer {
		return LogExecTransferConvert(v)
	} else if ty == types.TyLogExecWithdraw {
		return LogExecWithdrawConvert(v)
	} else if ty == types.TyLogExecDeposit {
		return LogExecDepositConvert(v)
	} else if ty == types.TyLogExecFrozen {
		return LogExecFrozenConvert(v)
	} else if ty == types.TyLogExecActive {
		return LogExecActiveConvert(v)
	} else if ty == types.TyLogGenesisTransfer {
		return LogGenesisTransferConvert(v)
	} else if ty == types.TyLogGenesisDeposit {
		return LogGenesisDepositConvert(v)
	} else if ty == types.TyLogMint {
		return LogMintConvert(v)
	} else if ty == types.TyLogBurn {
		return LogBurnConvert(v)
	}
	return notSupport(ty, v)
}


func LogFeeConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		key = []string{"coins-bty", "coins", l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

type account struct {
	Frozen int64 `json:"frozen"`
	Balance int64 `json:"balance"`
	Total int64 `json:"total"`
}

func accountConvert(acc *types.Account) account {
	return account{
		Frozen:  acc.Frozen,
		Balance: acc.Balance,
		Total:   acc.Frozen + acc.Balance,
	}
}

func LogTransferConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogDepositConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogExecTransferConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptExecAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.ExecAddr+":"+l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}


func LogExecWithdrawConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptExecAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.ExecAddr+":"+l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogExecDepositConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptExecAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.ExecAddr+":"+l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogExecFrozenConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptExecAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.ExecAddr+":"+l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogExecActiveConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptExecAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.ExecAddr+":"+l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogGenesisTransferConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogGenesisDepositConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptAccountTransfer
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogMintConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptAccountMint
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}

func LogBurnConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l types.ReceiptAccountMint
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		// TODO 如何设置提前的key ， 如 token
		key = []string{"coins-bty", "coins", l.Current.Addr}
		prev, _ = json.Marshal(accountConvert(l.Prev))
		current, _ = json.Marshal(accountConvert(l.Current))
		return
	}
	return nil, nil, nil, errors.New("failed")
}



