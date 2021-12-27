package executor

import (
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
)

// Action action struct
type Action struct {
	statedb   dbm.KV
	txhash    []byte
	fromaddr  string
	toaddr    string
	blocktime int64
	height    int64
	execaddr  string
	localDB   dbm.KVDB
	index     int
	api       client.QueueProtocolAPI
}

//NewAction ...
func NewAction(e *zksync, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	toaddr := tx.GetTo()
	return &Action{
		statedb:   e.GetStateDB(),
		txhash:    hash,
		fromaddr:  fromaddr,
		toaddr:    toaddr,
		blocktime: e.GetBlockTime(),
		height:    e.GetHeight(),
		execaddr:  dapp.ExecAddress(string(tx.Execer)),
		localDB:   e.GetLocalDB(),
		index:     index,
		api:       e.GetAPI(),
	}
}

//GetIndex get index
func (a *Action) GetIndex() int64 {
	return a.height*types.MaxTxsPerBlock + int64(a.index)
}

func (a *Action) Deposit(payload *zt.Deposit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var err error

	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}

	leaf,err := GetLeafByEthAddress(a.localDB, payload.GetEthAddress())
	//leaf不存在就添加
	if leaf == nil {
		//添加之前先计算证明
		receipt, err := calProof(a.localDB, nil, 0, "")
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		receiptLog := &types.ReceiptLog{Ty: et.TyDepositLog, Log: types.Encode(receipt)}
		logs = append(logs, receiptLog)

		leaf , err =  AddNewLeaf(a.localDB, payload.GetEthAddress(), payload.GetChainType(), payload.GetTokenId(), int64(payload.GetAmount()))
		if err != nil {
			return nil, errors.Wrapf(err, "db.AddNewLeaf")
		}
		receipt, err = calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		receiptLog = &types.ReceiptLog{Ty: et.TyDepositLog, Log: types.Encode(receipt)}
		logs = append(logs, receiptLog)
	} else {
		receipt, err := calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		receiptLog := &types.ReceiptLog{Ty: et.TyDepositLog, Log: types.Encode(receipt)}
		logs = append(logs, receiptLog)

		leaf , err =  UpdateLeaf(a.localDB, leaf.GetAccountId(), payload.GetChainType(), payload.GetTokenId(), int64(payload.GetAmount()))
		if err != nil {
			return nil, errors.Wrapf(err, "db.AddNewLeaf")
		}
		receipt, err = calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		receiptLog = &types.ReceiptLog{Ty: et.TyDepositLog, Log: types.Encode(receipt)}
		logs = append(logs, receiptLog)
	}

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) Withdraw(payload *zt.Withdraw) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	leaf, err := GetLeafByAccountId(a.localDB, payload.GetAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if leaf == nil {
		return nil, errors.New("account not exist")
	}
	err = checkAmount(leaf, int64(payload.GetAmount()), payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	//更新之前先计算证明
	receipt, err := calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog := &types.ReceiptLog{Ty: et.TyDepositLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	err = UpdateAccountBalance(a.localDB, payload.GetAccountId(), payload.GetTokenId(), -int64(payload.GetAmount()), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateAccountBalance")
	}
	accountLog := &et.ReceiptAccount{
		AccountId: payload.GetAccountId(),
		Balance:   payload.GetAmount(),
		Index:     a.GetIndex(),
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyWithdrawLog, Log: types.Encode(accountLog)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func checkAmount(leaf *zt.Leaf, amount int64, tokenId int32, chainType string) error {
	if v, ok := leaf.GetChainBalanceMap()[chainType];ok{
		chainBalance := leaf.GetChainBalances()[v]
		if v, ok = chainBalance.GetTokenBalanceMap()[tokenId];ok {
			tokenBalance := chainBalance.GetTokenBalances()[v]
			if tokenBalance.GetBalance() >= amount {
				return nil
			} else {
				return errors.New("balance not enough")
			}
		}
	}
	//没找到也说明balance不够
	return errors.New("balance not enough")
}

func (a *Action) Transfer(payload *et.Transfer) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	err := checkAmount(payload.GetFromAccountId(), int64(payload.GetAmount()), a.localDB, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "checkAmount")
	}
	err = UpdateAccountBalance(a.localDB, payload.GetFromAccountId(), payload.GetTokenId(), -int64(payload.GetAmount()), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateAccountBalance")
	}
	fromAccountLog := &et.ReceiptAccount{
		AccountId: payload.GetFromAccountId(),
		Balance:   -payload.GetAmount(),
		Index:     a.GetIndex(),
	}
	fromReceiptLog := &types.ReceiptLog{Ty: et.TyTransferLog, Log: types.Encode(fromAccountLog)}
	logs = append(logs, fromReceiptLog)
	err = UpdateAccountBalance(a.localDB, payload.GetToAccountId(), payload.GetTokenId(), int64(payload.GetAmount()), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateAccountBalance")
	}

	toAccountLog := &et.ReceiptAccount{
		AccountId: payload.GetToAccountId(),
		Balance:   payload.GetAmount(),
		Index:     a.GetIndex(),
	}

	toReceiptLog := &types.ReceiptLog{Ty: et.TyWithdrawLog, Log: types.Encode(toAccountLog)}
	logs = append(logs, toReceiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) TransferToNew(payload *et.TransferToNew) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	err := checkAmount(payload.GetFromAccountId(), int64(payload.GetAmount()), a.localDB, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "checkAmount")
	}
	err = UpdateAccountBalance(a.localDB, payload.GetFromAccountId(), payload.GetTokenId(), -int64(payload.GetAmount()), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateAccountBalance")
	}
	fromAccountLog := &et.ReceiptAccount{
		AccountId: payload.GetFromAccountId(),
		Balance:   -payload.GetAmount(),
		Index:     a.GetIndex(),
	}
	fromReceiptLog := &types.ReceiptLog{Ty: et.TyTransferToNewLog, Log: types.Encode(fromAccountLog)}
	logs = append(logs, fromReceiptLog)
	toAccountId, err := CreateNewContractAccount(a.localDB, a.fromaddr)
	err = UpdateAccountBalance(a.localDB, toAccountId, payload.GetTokenId(), int64(payload.GetAmount()), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateAccountBalance")
	}
	toAccountLog := &et.ReceiptAccount{
		AccountId: toAccountId,
		Balance:   payload.GetAmount(),
		Index:     a.GetIndex(),
	}

	toReceiptLog := &types.ReceiptLog{Ty: et.TyTransferToNewLog, Log: types.Encode(toAccountLog)}
	logs = append(logs, toReceiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) ForceQuit(payload *et.ForceQuit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	account, err := getAccount(a.localDB, payload.GetAccountId())
	if err != nil {
		return nil, err
	}

	for _, v := range account.ChainBalances {
		if payload.GetChainType() == v.ChainType {
			for _, token := range v.GetTokenBalances() {
				if token.TokenId == payload.GetTokenId() {
					balance := token.GetBalance()
					token.Balance = 0
					accountLog := &et.ReceiptAccount{
						AccountId: payload.GetAccountId(),
						Balance:   uint64(balance),
						Index:     a.GetIndex(),
					}

					toReceiptLog := &types.ReceiptLog{Ty: et.TyForceExitLog, Log: types.Encode(accountLog)}
					logs = append(logs, toReceiptLog)
					receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
					return receipts, nil
				}
			}
			break
		}
	}

	return nil, errors.New("token not find")
}

func calProof(db dbm.KV, leaf *zt.Leaf, tokenId int32, chainType string) (*zt.ReceiptLeaf,error) {
	receipt := &zt.ReceiptLeaf{AccountId: leaf.GetAccountId()}
	if leaf == nil {
		//leaf之前不存在时，不会有token，直接返回leaf的子树即可
		leafProof, err := CalLeafProof(db, 0)
		if err != nil {
			return nil, errors.Wrapf(err, "CalLeafProof")
		}
		receipt.TreeProof = leafProof
		return receipt, nil
	}
	leafProof, err := CalLeafProof(db, leaf.GetAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "CalLeafProof")
	}
	receipt.TreeProof = leafProof
	if index,ok:= leaf.GetChainBalanceMap()[chainType];ok {
		tokenProof, err := CalTokenProof(db, leaf.GetChainBalances()[index], tokenId)
		if err != nil {
			return nil, errors.Wrapf(err, "CalLeafProof")
		}
		receipt.TokenProof = tokenProof
	}
	return receipt, nil
}



