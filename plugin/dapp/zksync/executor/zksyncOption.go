package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
	"strconv"
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

func (a *Action)  Deposit(payload *zt.Deposit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var err error

	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}

	leaf, err := GetLeafByEthAddress(a.localDB, payload.GetEthAddress())
	//leaf不存在就添加
	if leaf == nil {
		//添加之前先计算证明
		receipt, err := calProof(a.localDB, nil, 0, "")
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		receiptLog := &types.ReceiptLog{Ty: zt.TyDepositLog, Log: types.Encode(receipt)}
		logs = append(logs, receiptLog)

		leaf, err = AddNewLeaf(a.localDB, payload.GetEthAddress(), payload.GetChainType(), payload.GetTokenId(), int64(payload.GetAmount()))
		if err != nil {
			return nil, errors.Wrapf(err, "db.AddNewLeaf")
		}
		receipt, err = calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		receiptLog = &types.ReceiptLog{Ty: zt.TyDepositLog, Log: types.Encode(receipt)}
		logs = append(logs, receiptLog)
	} else {
		receipt, err := calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		receiptLog := &types.ReceiptLog{Ty: zt.TyDepositLog, Log: types.Encode(receipt)}
		logs = append(logs, receiptLog)

		leaf, err = UpdateLeaf(a.localDB, leaf.GetAccountId(), payload.GetChainType(), payload.GetTokenId(), int64(payload.GetAmount()))
		if err != nil {
			return nil, errors.Wrapf(err, "db.AddNewLeaf")
		}
		receipt, err = calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		receiptLog = &types.ReceiptLog{Ty: zt.TyDepositLog, Log: types.Encode(receipt)}
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
	receiptLog := &types.ReceiptLog{Ty: zt.TyWithdrawLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	leaf, err = UpdateLeaf(a.localDB, payload.GetAccountId(), payload.GetChainType(), payload.GetTokenId(), -int64(payload.GetAmount()))
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//取款之后计算证明
	receipt, err = calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyWithdrawLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func checkAmount(leaf *zt.Leaf, amount int64, tokenId int32, chainType string) error {
	if v, ok := leaf.GetChainBalanceMap()[chainType]; ok {
		chainBalance := leaf.GetChainBalances()[v]
		if v, ok = chainBalance.GetTokenBalanceMap()[tokenId]; ok {
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

func (a *Action) ContractToLeaf(payload *zt.ContractToLeaf) (*types.Receipt, error) {
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
	receiptLog := &types.ReceiptLog{Ty: zt.TyContractToLeafLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	leaf, err = UpdateLeaf(a.localDB, payload.GetAccountId(), payload.GetChainType(), payload.GetTokenId(), int64(payload.GetAmount()))
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新合约账户
	err = a.UpdateContractAccount(a.fromaddr, -int64(payload.GetAmount()), payload.GetChainType(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateContractAccount")
	}
	//取款之后计算证明
	receipt, err = calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyContractToLeafLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) LeafToContract(payload *zt.LeafToContract) (*types.Receipt, error) {
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
	receiptLog := &types.ReceiptLog{Ty: zt.TyLeafToContractLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	leaf, err = UpdateLeaf(a.localDB, payload.GetAccountId(), payload.GetChainType(), payload.GetTokenId(), -int64(payload.GetAmount()))
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新合约账户
	err = a.UpdateContractAccount(a.fromaddr, int64(payload.GetAmount()), payload.GetChainType(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateContractAccount")
	}
	//取款之后计算证明
	receipt, err = calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyLeafToContractLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}


func (a *Action) UpdateContractAccount(addr string, amount int64, chainType string, tokenId int32) error {
	execAddr := address.ExecAddress(zt.Zksync)

	accountdb, _ := account.NewAccountDB(a.api.GetConfig(), zt.Zksync, chainType + strconv.Itoa(int(tokenId)), a.statedb)
	contractAccount := accountdb.LoadExecAccount(addr, execAddr)
	if contractAccount.Balance + amount < 0 {
		return errors.New("balance not enough")
	} else {
		contractAccount.Balance += amount
		accountdb.SaveExecAccount(execAddr, contractAccount)
	}
	return nil
}

func (a *Action) Transfer(payload *zt.Transfer) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	fromLeaf, err := GetLeafByAccountId(a.localDB, payload.GetFromAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	err = checkAmount(fromLeaf, int64(payload.GetAmount()), payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}
	toLeaf, err := GetLeafByAccountId(a.localDB, payload.GetToAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if toLeaf == nil {
		return nil, errors.New("account not exist")
	}
	//更新之前先计算证明
	receipt, err := calProof(a.localDB, fromLeaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyTransferLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	//更新fromLeaf
	fromLeaf, err = UpdateLeaf(a.localDB, payload.GetFromAccountId(), payload.GetChainType(), payload.GetTokenId(), -int64(payload.GetAmount()))
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.localDB, fromLeaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyTransferLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	//更新之前先计算证明
	receipt, err = calProof(a.localDB, toLeaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyTransferLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	//更新toLeaf
	toLeaf, err = UpdateLeaf(a.localDB, payload.GetToAccountId(), payload.GetChainType(), payload.GetTokenId(), int64(payload.GetAmount()))
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.localDB, toLeaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyTransferLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) TransferToNew(payload *zt.TransferToNew) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	fromLeaf, err := GetLeafByAccountId(a.localDB, payload.GetFromAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	err = checkAmount(fromLeaf, int64(payload.GetAmount()), payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}
	toLeaf, err := GetLeafByEthAddress(a.localDB, payload.GetToEthAddress())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if toLeaf != nil {
		return nil, errors.New("to account already exist")
	}
	//更新之前先计算证明
	receipt, err := calProof(a.localDB, fromLeaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyTransferToNewLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	//更新fromLeaf
	fromLeaf, err = UpdateLeaf(a.localDB, payload.GetFromAccountId(), payload.GetChainType(), payload.GetTokenId(), -int64(payload.GetAmount()))
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.localDB, fromLeaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyTransferToNewLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	//更新之前先计算证明
	receipt, err = calProof(a.localDB, nil, 0, "")
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyTransferToNewLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	//新增toLeaf
	toLeaf, err = AddNewLeaf(a.localDB, payload.GetToEthAddress(), payload.GetChainType(), payload.GetTokenId(), int64(payload.GetAmount()))
	if err != nil {
		return nil, errors.Wrapf(err, "db.AddNewLeaf")
	}
	//新增之后计算证明
	receipt, err = calProof(a.localDB, toLeaf, payload.GetTokenId(), payload.GetChainType())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	receiptLog = &types.ReceiptLog{Ty: zt.TyTransferToNewLog, Log: types.Encode(receipt)}
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) ForceQuit(payload *zt.ForceQuit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	leaf, err := GetLeafByEthAddress(a.localDB, payload.GetEthAddress())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}

	//首先找到token
	if idx, ok := leaf.ChainBalanceMap[payload.GetChainType()]; ok {
		chainBalance := leaf.GetChainBalances()[idx]
		if idx, ok = chainBalance.GetTokenBalanceMap()[payload.GetTokenId()]; ok {
			tokenBalance := chainBalance.GetTokenBalances()[idx]
			//更新之前先计算证明
			receipt, err := calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
			if err != nil {
				return nil, errors.Wrapf(err, "calProof")
			}
			receiptLog := &types.ReceiptLog{Ty: zt.TyForceExitLog, Log: types.Encode(receipt)}
			logs = append(logs, receiptLog)

			//更新fromLeaf
			leaf, err = UpdateLeaf(a.localDB, leaf.GetAccountId(), payload.GetChainType(), payload.GetTokenId(), -tokenBalance.GetBalance())
			if err != nil {
				return nil, errors.Wrapf(err, "db.UpdateLeaf")
			}
			//更新之后计算证明
			receipt, err = calProof(a.localDB, leaf, payload.GetTokenId(), payload.GetChainType())
			if err != nil {
				return nil, errors.Wrapf(err, "calProof")
			}
			receiptLog = &types.ReceiptLog{Ty: zt.TyForceExitLog, Log: types.Encode(receipt)}
			logs = append(logs, receiptLog)
			receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
			return receipts, nil
		}
	}

	//上面没返回就是没找到token
	return nil, errors.New("token not find")
}

func calProof(db dbm.KV, leaf *zt.Leaf, tokenId int32, chainType string) (*zt.ReceiptLeaf, error) {
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
	if index, ok := leaf.GetChainBalanceMap()[chainType]; ok {
		tokenProof, err := CalTokenProof(db, leaf.GetChainBalances()[index], tokenId)
		if err != nil {
			return nil, errors.Wrapf(err, "CalLeafProof")
		}
		receipt.TokenProof = tokenProof
	}
	return receipt, nil
}
