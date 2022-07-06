package executor

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/33cn/chain33/common/log/log15"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/pkg/errors"
)

var (
	zklog = log15.New("module", "exec.zksync")
)

// Action action struct
type Action struct {
	statedb   dbm.KV
	txhash    []byte
	fromaddr  string
	blocktime int64
	height    int64
	execaddr  string
	localDB   dbm.KVDB
	index     int
	api       client.QueueProtocolAPI
}

//NewAction ...
func NewAction(z *zksync, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{
		statedb:   z.GetStateDB(),
		txhash:    hash,
		fromaddr:  fromaddr,
		blocktime: z.GetBlockTime(),
		height:    z.GetHeight(),
		execaddr:  dapp.ExecAddress(string(tx.Execer)),
		localDB:   z.GetLocalDB(),
		index:     index,
		api:       z.GetAPI(),
	}
}

//GetIndex get index
func (a *Action) GetIndex() int64 {
	return a.height*types.MaxTxsPerBlock + int64(a.index)
}

//TODO:HexAddr2Decimal 地址的转换在确认其必要性，最后在合约内部进行清理，
func (a *Action) Deposit(payload *zt.ZkDeposit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var err error

	err = checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}

	if !checkIsNormalToken(payload.TokenId) {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenId=%d should less than system NFT base ID=%d", payload.TokenId, zt.SystemNFTTokenId)
	}

	zklog.Info("start zksync deposit", "eth", payload.EthAddress, "chain33", payload.Chain33Addr)
	//只有管理员能操作
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
	}

	//转换10进制
	payload.Chain33Addr = zt.HexAddr2Decimal(payload.Chain33Addr)
	payload.EthAddress = zt.HexAddr2Decimal(payload.EthAddress)

	//TODO set chainID
	lastPriority, err := getLastEthPriorityQueueID(a.statedb, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "get eth last priority queue id")
	}
	lastPriorityId, ok := big.NewInt(0).SetString(lastPriority.GetID(), 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, fmt.Sprintf("getID =%s", lastPriority.GetID()))
	}
	if lastPriorityId.Int64()+1 != payload.GetEthPriorityQueueId() {
		return nil, errors.Wrapf(types.ErrNotAllow, "eth last priority queue id=%d,new=%d", lastPriorityId, payload.GetEthPriorityQueueId())
	}

	leaf, err := GetLeafByChain33AndEthAddress(a.statedb, payload.GetChain33Addr(), payload.GetEthAddress())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyDepositAction,
		TokenID:     payload.TokenId,
		Amount:      payload.Amount,
		SigData:     payload.Signature,
	}

	//leaf不存在就添加
	if leaf == nil {
		zklog.Info("zksync deposit add new leaf")

		var accountID uint64
		lastAccountID, _ := getLatestAccountID(a.statedb)
		if zt.InvalidAccountId == lastAccountID {
			ethFeeAddr, chain33FeeAddr := getCfgFeeAddr(cfg)
			kvs4InitAccount, err := NewInitAccount(ethFeeAddr, chain33FeeAddr)
			if nil != err {
				return nil, err
			}
			kvs = append(kvs, kvs4InitAccount...)

			//对于首次进行存款的用户，其账户ID从SystemNFTAccountId后开始进行连续分配
			accountID = zt.SystemNFTAccountId + 1
		} else {
			accountID = uint64(lastAccountID) + 1
		}

		//accountID, tokenID uint64, amount, ethAddress,  chain33Addr string, statedb dbm.KV, leaf *zt.Leaf) ([]*types.KeyValue, *types.ReceiptLog, error)
		createKVS, l2Log, _, err := applyL2AccountCreate(accountID, operationInfo.TokenID, operationInfo.Amount, payload.EthAddress, payload.Chain33Addr, a.statedb, true)
		if nil != err {
			return nil, errors.Wrapf(err, "applyL2AccountCreate")
		}

		kvs = append(kvs, createKVS...)
		l2Log.Ty = int32(zt.TyDepositLog)
		logs = append(logs, l2Log)
	} else {
		updateKVs, l2Log, _, err := applyL2AccountUpdate(leaf.AccountId, operationInfo.TokenID, operationInfo.Amount, zt.Add, a.statedb, leaf, true)
		if nil != err {
			return nil, errors.Wrapf(err, "applyL2AccountUpdate")
		}

		kvs = append(kvs, updateKVs...)
		l2Log.Ty = int32(zt.TyDepositLog)
		logs = append(logs, l2Log)
	}
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	//add priority part
	r := makeSetEthPriorityIdReceipt(0, lastPriorityId.Int64(), payload.EthPriorityQueueId)
	return mergeReceipt(receipts, r), nil
}

func (a *Action) Withdraw(payload *zt.ZkWithdraw) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	fee := zt.FeeMap[zt.TyWithdrawAction]
	//加上手续费
	amountInt, _ := new(big.Int).SetString(payload.Amount, 10)
	feeInt, _ := new(big.Int).SetString(fee, 10)
	totalAmount := new(big.Int).Add(amountInt, feeInt).String()

	leaf, err := GetLeafByAccountId(a.statedb, payload.GetAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if leaf == nil {
		return nil, errors.New("account not exist")
	}
	err = authVerification(payload.GetSignature().PubKey, leaf.GetPubKey())
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}

	token, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.AccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(token, totalAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyWithdrawAction,
		TokenID:     payload.TokenId,
		Amount:      payload.Amount,
		FeeAmount:   fee,
		SigData:     payload.Signature,
		AccountID:   payload.AccountId,
	}

	balancekv, balancehistory, err := updateTokenBalance(payload.AccountId, operationInfo.TokenID, totalAmount, zt.Sub, a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "updateTokenBalance")
	}
	kvs = append(kvs, balancekv)

	updateLeafKvs, err := updateLeafOpt(a.statedb, leaf, payload.GetTokenId(), zt.Sub)
	if nil != err {
		return nil, err
	}
	kvs = append(kvs, updateLeafKvs...)

	withdrawReceiptLog := &zt.AccountTokenBalanceReceipt{
		EthAddress: leaf.EthAddress,
		Chain33Addr: leaf.Chain33Addr,
		TokenId: payload.GetTokenId(),
		AccountId: leaf.AccountId,
		BalanceBefore: balancehistory.before,
		BalanceAfter: balancehistory.after,
	}

	receiptLog := &types.ReceiptLog{Ty: zt.TyWithdrawLog, Log: types.Encode(withdrawReceiptLog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

func checkAmount(token *zt.TokenBalance, amount string) error {
	if token != nil {
		balance, _ := new(big.Int).SetString(token.Balance, 10)
		need, _ := new(big.Int).SetString(amount, 10)
		if balance.Cmp(need) >= 0 {
			return nil
		} else {
			return errors.New("balance not enough")
		}
	}
	//token为nil
	return errors.New("balance not enough")
}

//合约　----> L2
func (a *Action) ContractToTree(payload *zt.ZkContractToTree) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	//因为chain33合约精度为1e8,而外部输入精度则为1e18, 单位为wei,需要统一转化为1e8
	amount_len := len(payload.Amount)
	if amount_len < 10 {
		return nil, errors.New("Too Little value to do operation TreeToContract")
	}

	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}

	leaf, err := GetLeafByAccountId(a.statedb, payload.GetAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if leaf == nil {
		return nil, errors.New("account:" + strconv.FormatUint(payload.AccountId, 10) + " not exist")
	}

	err = authVerification(payload.GetSignature().PubKey, leaf.GetPubKey())
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyContractToTreeAction,
		TokenID:     payload.TokenId,
		Amount:      payload.Amount,
		SigData:     payload.Signature,
		AccountID:   payload.AccountId,
	}

	balancekv, balancehistory, err := updateTokenBalance(leaf.AccountId, operationInfo.TokenID, operationInfo.Amount, zt.Add, a.statedb)
	if nil != err {
		return nil, err
	}
	kvs = append(kvs, balancekv)

	updateLeafKvs, err := updateLeafOpt(a.statedb, leaf, payload.GetTokenId(), zt.Add)
	if nil != err {
		return nil, err
	}
	kvs = append(kvs, updateLeafKvs...)

	l2BalanceLog := &zt.AccountTokenBalanceReceipt{}
	l2BalanceLog.EthAddress    = leaf.EthAddress
	l2BalanceLog.Chain33Addr   = leaf.Chain33Addr
	l2BalanceLog.TokenId       = payload.GetTokenId()
	l2BalanceLog.AccountId     = leaf.AccountId
	l2BalanceLog.BalanceBefore = balancehistory.before
	l2BalanceLog.BalanceAfter  = balancehistory.after
	l2Log := &types.ReceiptLog{Ty: zt.TyContractToTreeLog, Log: types.Encode(l2BalanceLog)}

	//更新合约账户
	accountKvs, l1Log, err := a.UpdateContractAccount(a.fromaddr, payload.GetAmount(), payload.GetTokenId(), zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateContractAccount")
	}
	kvs = append(kvs, accountKvs...)

	logs = append(logs, l1Log)
	logs = append(logs, l2Log)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

//L2 ---->  合约账户
func (a *Action) TreeToContract(payload *zt.ZkTreeToContract) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	//因为chain33合约精度为1e8,而外部输入精度则为1e18, 单位为wei,需要统一转化为1e8
	amount_len := len(payload.Amount)
	if amount_len < 10 {
		return nil, errors.New("Too Little value to do operation TreeToContract")
	}

	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	leaf, err := GetLeafByAccountId(a.statedb, payload.GetAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if leaf == nil {
		return nil, errors.New("account not exist")
	}
	err = authVerification(payload.Signature.PubKey, leaf.GetPubKey())
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}

	token, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.AccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(token, payload.GetAmount())
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyTreeToContractAction,
		TokenID:     payload.TokenId,
		Amount:      payload.Amount,
		SigData:     payload.Signature,
		AccountID:   payload.AccountId,
	}

	balancekv, balancehistory, err := updateTokenBalance(leaf.AccountId, operationInfo.TokenID, operationInfo.Amount, zt.Sub, a.statedb)
	if nil != err {
		return nil, err
	}
	kvs = append(kvs, balancekv)

	updateLeafKvs, err := updateLeafOpt(a.statedb, leaf, payload.GetTokenId(), zt.Sub)
	if nil != err {
		return nil, err
	}
	kvs = append(kvs, updateLeafKvs...)

	l2BalanceLog := &zt.AccountTokenBalanceReceipt{}
	l2BalanceLog.EthAddress    = leaf.EthAddress
	l2BalanceLog.Chain33Addr   = leaf.Chain33Addr
	l2BalanceLog.TokenId       = payload.GetTokenId()
	l2BalanceLog.AccountId     = leaf.AccountId
	l2BalanceLog.BalanceBefore = balancehistory.before
	l2BalanceLog.BalanceAfter  = balancehistory.after
	l2Log := &types.ReceiptLog{Ty: zt.TyContractToTreeLog, Log: types.Encode(l2BalanceLog)}

	//更新合约账户
	accountKvs, l1Log, err := a.UpdateContractAccount(a.fromaddr, payload.Amount, payload.GetTokenId(), zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateContractAccount")
	}
	kvs = append(kvs, accountKvs...)

	receiptLog := &types.ReceiptLog{Ty: zt.TyTreeToContractLog, Log: types.Encode(l2Log)}
	logs = append(logs, l1Log)
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) UpdateContractAccount(addr string, amount string, tokenId uint64, option int32) ([]*types.KeyValue, *types.ReceiptLog, error) {
	accountdb, _ := account.NewAccountDB(a.api.GetConfig(), zt.Zksync, strconv.Itoa(int(tokenId)), a.statedb)
	contractAccount := accountdb.LoadAccount(addr)
	//accountdb去除末尾10位小数
	amount2Contract := amount[ : len(amount) - 10]
	shortChangeBigInt, _ := new(big.Int).SetString(amount2Contract, 10)
	shortChange := shortChangeBigInt.Int64()
	accBefore := &types.Account{
		Balance:contractAccount.Balance,
		Addr: addr,
	}
	if option == zt.Sub {
		if contractAccount.Balance < shortChange {
			return nil, nil, errors.New("balance not enough")
		}
		contractAccount.Balance -= shortChange
	} else {
		contractAccount.Balance += shortChange
	}

	kvs := accountdb.GetKVSet(contractAccount)

	accAfter := &types.Account{
		Balance:contractAccount.Balance,
		Addr: addr,
	}

	receiptBalance := &types.ReceiptAccountTransfer{
		Prev:    accBefore,
		Current: accAfter,
	}
	log1 := &types.ReceiptLog{
		Ty:  int32(types.TyLogTransfer),
		Log: types.Encode(receiptBalance),
	}

	return kvs, log1, nil
}

func (a *Action) Transfer(payload *zt.ZkTransfer) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	if !checkIsNormalToken(payload.TokenId) {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenId=%d should less than system NFT base ID=%d", payload.TokenId, zt.SystemNFTTokenId)
	}

	fee := zt.FeeMap[zt.TyTransferAction]
	//加上手续费
	amountInt, _ := new(big.Int).SetString(payload.Amount, 10)
	feeInt, _ := new(big.Int).SetString(fee, 10)
	totalAmount := new(big.Int).Add(amountInt, feeInt).String()

	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.GetFromAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	err = authVerification(payload.Signature.PubKey, fromLeaf.PubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}
	fromToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(fromToken, totalAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	//1.操作from 账户
	fromKVs, _, receiptFrom, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), payload.GetTokenId(), totalAmount,zt.Sub, a.statedb, fromLeaf, false)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)

	//2.操作to 账户
	toLeaf, err := GetLeafByAccountId(a.statedb, payload.ToAccountId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if toLeaf == nil {
		return nil, errors.New("account not exist")
	}

	toKVs, _, receiptTo, err := applyL2AccountUpdate(toLeaf.GetAccountId(), payload.GetTokenId(), payload.GetAmount(), zt.Add, a.statedb, toLeaf, false)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, toKVs...)
	transferLog := &zt.TransferReceipt4L2{
		From:receiptFrom,
		To: receiptTo,
	}
	l2Transferlog := &types.ReceiptLog{
		Ty: zt.TyTransferLog,
		Log: types.Encode(transferLog),
	}
	logs = append(logs, l2Transferlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	//2.操作交易费账户
	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

func (a *Action) TransferToNew(payload *zt.ZkTransferToNew) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	fee := zt.FeeMap[zt.TyTransferToNewAction]
	//加上手续费
	amountInt, _ := new(big.Int).SetString(payload.Amount, 10)
	feeInt, _ := new(big.Int).SetString(fee, 10)
	totalAmount := new(big.Int).Add(amountInt, feeInt).String()

	//转换10进制
	payload.ToChain33Address = zt.HexAddr2Decimal(payload.ToChain33Address)
	payload.ToEthAddress = zt.HexAddr2Decimal(payload.ToEthAddress)

	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.GetFromAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	err = authVerification(payload.Signature.PubKey, fromLeaf.PubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}

	fromToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(fromToken, totalAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyTransferToNewAction,
		TokenID:     payload.TokenId,
		Amount:      payload.Amount,
		FeeAmount:   fee,
		SigData:     payload.Signature,
		AccountID:   payload.FromAccountId,
	}

	toLeaf, err := GetLeafByChain33AndEthAddress(a.statedb, payload.GetToChain33Address(), payload.GetToEthAddress())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}
	if toLeaf != nil {
		return nil, errors.New("to account already exist")
	}

	//1.操作from 账户
	fromKVs, _, receiptFrom, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), payload.GetTokenId(), totalAmount,zt.Sub, a.statedb, fromLeaf, false)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)



	//1.操作to 账户
	lastAccountID, err := getLatestAccountID(a.statedb)
	if lastAccountID == zt.InvalidAccountId {
		return nil, errors.Wrapf(err, "getLatestAccountID")
	}
	accountIDNew := uint64(lastAccountID) + 1

	toKVs, _, receiptTo, err := applyL2AccountCreate(accountIDNew, operationInfo.TokenID, operationInfo.Amount, payload.GetToEthAddress(),  payload.GetToChain33Address(), a.statedb, false)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountCreate")
	}
	kvs = append(kvs, toKVs...)


	transferToNewLog := &zt.TransferReceipt4L2{
		From:receiptFrom,
		To: receiptTo,
	}
	l2Transferlog := &types.ReceiptLog{
		Ty: zt.TyTransferToNewLog,
		Log: types.Encode(transferToNewLog),
	}
	logs = append(logs, l2Transferlog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

func (a *Action) ForceExit(payload *zt.ZkForceExit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	fee := zt.FeeMap[zt.TyForceExitAction]

	leaf, err := GetLeafByAccountId(a.statedb, payload.AccountId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	if leaf == nil {
		return nil, errors.New("account not exist")
	}

	token, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.AccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	//token不存在时，不需要取
	if token == nil {
		return nil, errors.New("token not find")
	}

	//加上手续费
	amountInt, _ := new(big.Int).SetString(token.Balance, 10)
	feeInt, _ := new(big.Int).SetString(fee, 10)
	//存量不够手续费时，不能取
	if amountInt.Cmp(feeInt) <= 0 {
		return nil, errors.New("no enough fee")
	}

	fromKVs, l2LogFrom, _, err := applyL2AccountUpdate(leaf.GetAccountId(), payload.GetTokenId(), token.Balance, zt.Sub, a.statedb, leaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)
	l2LogFrom.Ty = zt.TyForceExitLog
	logs = append(logs, l2LogFrom)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

func (a *Action) SetPubKey(payload *zt.ZkSetPubKey) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	leaf, err := GetLeafByAccountId(a.statedb, payload.GetAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByEthAddress")
	}
	if leaf == nil {
		return nil, errors.New("account not exist")
	}

	if payload.GetPubKey() == nil || len(payload.GetPubKey().X) <= 0 || len(payload.GetPubKey().Y) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "pubkey invalid")
	}

	if payload.PubKeyTy == 0 {
		//已经设置过缺省公钥，不允许再设置
		if leaf.PubKey != nil {
			return nil, errors.Wrapf(types.ErrNotAllow, "pubKey exited already")
		}

		//校验预存的地址是否和公钥匹配
		hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
		hash.Write(zt.Str2Byte(payload.PubKey.X))
		hash.Write(zt.Str2Byte(payload.PubKey.Y))
		calcChain33Addr := zt.Byte2Str(hash.Sum(nil))
		if calcChain33Addr != leaf.Chain33Addr {
            zklog.Error("SetPubKey", "leaf.Chain33Addr", leaf.Chain33Addr, "calcChain33Addr", calcChain33Addr)
			return nil, errors.New("not your account")
		}
	}
	if payload.PubKeyTy > zt.SuperProxyPubKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong proxy ty=%d", payload.PubKeyTy)
	}

	specialData := &zt.OperationSpecialData{
		PubKeyType: payload.PubKeyTy,
		PubKey:     payload.PubKey,
	}
	if payload.PubKeyTy == 0 {
		specialData.PubKey = payload.Signature.PubKey
	}

	if payload.PubKeyTy == 0 {
		kvs, _, err = a.SetDefultPubKey(payload)
		if err != nil {
			return nil, errors.Wrapf(err, "setDefultPubKey")
		}
	} else {
		kvs, _, err = a.SetProxyPubKey(payload, leaf)
		if err != nil {
			return nil, errors.Wrapf(err, "setDefultPubKey")
		}
	}

	zklog := &zt.SetPubKeyReceipt{
		AccountId: payload.AccountId,
		PubKey: payload.PubKey,
		PubKeyTy: payload.PubKeyTy,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TySetPubKeyLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) SetDefultPubKey(payload *zt.ZkSetPubKey) ([]*types.KeyValue, []*types.KeyValue, error) {

	kvs, localKvs, err := UpdatePubKey(a.statedb, a.localDB, payload.GetPubKeyTy(), payload.GetPubKey(), payload.AccountId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}

	return kvs, localKvs, nil
}

//设置代理地址的公钥
func (a *Action) SetProxyPubKey(payload *zt.ZkSetPubKey, leaf *zt.Leaf) ([]*types.KeyValue, []*types.KeyValue, error) {

	err := authVerification(payload.Signature.PubKey, leaf.PubKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "authVerification")
	}

	kvs, localKvs, err := UpdatePubKey(a.statedb, a.localDB, payload.PubKeyTy, payload.GetPubKey(), payload.AccountId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}

	return kvs, localKvs, nil
}

func (a *Action) FullExit(payload *zt.ZkFullExit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	fee := zt.FeeMap[zt.TyFullExitAction]

	//只有管理员能操作
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
	}

	//fullexit last priority id 不能为空
	lastPriority, err := getLastEthPriorityQueueID(a.statedb, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "get eth last priority queue id")
	}
	lastId, ok := big.NewInt(0).SetString(lastPriority.GetID(), 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, fmt.Sprintf("getID =%s", lastPriority.GetID()))
	}

	if lastId.Int64()+1 != payload.GetEthPriorityQueueId() {
		return nil, errors.Wrapf(types.ErrNotAllow, "eth last priority queue id=%s,new=%d", lastPriority.ID, payload.GetEthPriorityQueueId())
	}

	leaf, err := GetLeafByAccountId(a.statedb, payload.AccountId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}

	if leaf == nil {
		return nil, errors.New("account not exist")
	}

	token, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.AccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}

	//token不存在时，不需要取
	if token == nil {
		return nil, errors.New("token not find")
	}

	//加上手续费
	amountInt, _ := new(big.Int).SetString(token.Balance, 10)
	feeInt, _ := new(big.Int).SetString(fee, 10)
	//存量不够手续费时，不能取
	if amountInt.Cmp(feeInt) <= 0 {
		return nil, errors.New("no enough fee")
	}

	fromKVs, l2LogFrom, _, err := applyL2AccountUpdate(leaf.GetAccountId(), payload.GetTokenId(), token.Balance, zt.Sub, a.statedb, leaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)
	l2LogFrom.Ty = zt.TyFullExitLog
	logs = append(logs, l2LogFrom)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	//add priority part
	r := makeSetEthPriorityIdReceipt(0, lastId.Int64(), payload.EthPriorityQueueId)

	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return mergeReceipt(receipts, r), nil
}

//验证身份
func authVerification(signPubKey *zt.ZkPubKey, leafPubKey *zt.ZkPubKey) error {
	if signPubKey == nil || leafPubKey == nil {
		return errors.New("set your pubKey")
	}
	if signPubKey.GetX() != leafPubKey.GetX() || signPubKey.GetY() != leafPubKey.GetY() {
		return errors.New("not your account")
	}
	return nil
}

//检查参数
func checkParam(amount string) error {
	if amount == "" || amount == "0" || strings.HasPrefix(amount, "-") {
		return types.ErrAmount
	}
	return nil
}

//not NFT token
func checkIsNormalToken(id uint64) bool {
	return id < zt.SystemNFTTokenId
}

func checkIsNFTToken(id uint64) bool {
	return id > zt.SystemNFTTokenId
}

func getLastEthPriorityQueueID(db dbm.KV, chainID uint32) (*zt.EthPriorityQueueID, error) {
	key := getEthPriorityQueueKey(chainID)
	v, err := db.Get(key)
	//未找到返回-1
	if isNotFound(err) {
		return &zt.EthPriorityQueueID{ID: "-1"}, nil
	}
	if err != nil {
		return nil, err
	}
	var id zt.EthPriorityQueueID
	err = types.Decode(v, &id)
	if err != nil {
		zklog.Error("getLastEthPriorityQueueID.decode", "err", err)
		return nil, err
	}

	return &id, nil
}

func getLatestAccountID(db dbm.KV) (int64, error) {
	key := calcLatestAccountIDKey()
	v, err := db.Get(key)

	if err != nil {
		return zt.InvalidAccountId, err
	}
	var id types.Int64
	err = types.Decode(v, &id)
	if err != nil {
		zklog.Error("getLastEthPriorityQueueID.decode", "err", err)
		return zt.InvalidAccountId, err
	}

	return id.Data, nil
}

func calcNewAccountKV(accounID int64) (*types.KeyValue) {
	key := calcLatestAccountIDKey()
	id := &types.Int64{
		Data: accounID,
	}
	value := types.Encode(id)
	kv := &types.KeyValue{
		Key: key,
		Value: value,
	}
	return kv
}

func makeSetEthPriorityIdReceipt(chainId uint32, prev, current int64) *types.Receipt {
	key := getEthPriorityQueueKey(chainId)
	log := &zt.ReceiptEthPriorityQueueID{
		Prev:    prev,
		Current: current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&zt.EthPriorityQueueID{ID: big.NewInt(current).String()})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  zt.TySetEthPriorityQueueId,
				Log: types.Encode(log),
			},
		},
	}
}

func mergeReceipt(receipt1, receipt2 *types.Receipt) *types.Receipt {
	if receipt2 != nil {
		receipt1.KV = append(receipt1.KV, receipt2.KV...)
		receipt1.Logs = append(receipt1.Logs, receipt2.Logs...)
	}

	return receipt1
}

func (a *Action) MakeFeeLog(amount string, tokenId uint64, sign *zt.ZkSignature) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var err error

	//todo 手续费收款方accountId可配置
	leaf, err := GetLeafByAccountId(a.statedb, zt.SystemFeeAccountId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}

	if leaf == nil {
		return nil, errors.New("account not exist")
	}

	toKVs, l2Log, _, err := applyL2AccountUpdate(leaf.GetAccountId(), tokenId, amount,zt.Add, a.statedb, leaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	l2Log.Ty = zt.TyFeeLog
	logs = append(logs, l2Log)
	kvs = append(kvs, toKVs...)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) setFee(payload *zt.ZkSetFee) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
	}

	lastFee, err := getFeeData(a.statedb, payload.ActionTy, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData err")
	}
	kv := &types.KeyValue{
		Key:   getZkFeeKey(payload.ActionTy, payload.TokenId),
		Value: []byte(payload.Amount),
	}
	kvs = append(kvs, kv)
	setFeelog := &zt.ReceiptSetFee{
		TokenId:       payload.TokenId,
		ActionTy:      payload.ActionTy,
		PrevAmount:    lastFee,
		CurrentAmount: payload.Amount,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TySetFeeLog, Log: types.Encode(setFeelog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func getFeeData(db dbm.KV, actionTy int32, tokenId uint64) (string, error) {
	key := getZkFeeKey(actionTy, tokenId)
	v, err := db.Get(key)
	if err != nil {
		if isNotFound(err) {
			return "0", nil
		} else {
			return "", errors.Wrapf(err, "get db")
		}
	}

	return string(v), nil
}

func (a *Action) MintNFT(payload *zt.ZkMintNFT) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	if payload.Amount <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount=%d", payload.Amount)
	}
	if payload.ErcProtocol != zt.ZKERC721 && payload.ErcProtocol != zt.ZKERC1155 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong erc protocol=%d", payload.ErcProtocol)
	}

	if payload.ErcProtocol == zt.ZKERC721 && payload.Amount != 1 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "erc721 only allow 1 nft,got=%d", payload.Amount)
	}

	contentPart1, contentPart2, fullContent, err := zt.SplitNFTContent(payload.ContentHash)
	if err != nil {
		return nil, errors.Wrapf(err, "split content hash=%s", payload.ContentHash)
	}

	id, err := getNFTIdByHash(a.statedb, fullContent)
	if err != nil && !isNotFound(err) {
		return nil, errors.Wrapf(err, "getNFTIdByHash")
	}
	if id != nil {
		return nil, errors.Wrapf(types.ErrNotAllow, "contenthash existed in nft id=%d", id.Data)
	}

	//暂定0 后面从数据库读取 TODO
	feeTokenId := uint64(0)
	feeAmount := zt.FeeMap[zt.TyMintNFTAction]
	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyMintNFTAction,
		TokenID:     zt.SystemNFTTokenId,
		Amount:      big.NewInt(0).SetUint64(payload.GetAmount()).String(),
		FeeAmount:   feeAmount,
		SigData:     payload.Signature,
		AccountID:   payload.GetFromAccountId(),
		SpecialInfo: &zt.OperationSpecialInfo{},
	}
	speciaData := &zt.OperationSpecialData{
		AccountID:   payload.GetFromAccountId(),
		RecipientID: payload.RecipientId,
		TokenID:     []uint64{feeTokenId},
		Amount:      []string{big.NewInt(0).SetUint64(payload.ErcProtocol).String()},
	}
	operationInfo.SpecialInfo.SpecialDatas = append(operationInfo.SpecialInfo.SpecialDatas, speciaData)

	//1. calc fee,收取铸币的交易费
	creatorLeaf, err := GetLeafByAccountId(a.statedb, payload.GetFromAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if creatorLeaf == nil {
		return nil, errors.New("account not exist")
	}
	err = authVerification(payload.Signature.PubKey, creatorLeaf.PubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}
	feeToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(feeToken, feeAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	feeKVsFrom, l2feeLogFrom, _, err := applyL2AccountUpdate(creatorLeaf.GetAccountId(), feeTokenId, feeAmount, zt.Sub, a.statedb, creatorLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, feeKVsFrom...)
	l2feeLogFrom.Ty = zt.TyFeeLog
	logs = append(logs, l2feeLogFrom)


	//2. creator SystemNFTTokenId balance+1 产生serialId
	//创建者的NFT_TOKEN_ID余额代表创建nft的次数，同时将当前余额(即未计入当前创建次数)设置为serial_id,且将当前余额+1
	creatorLeaf, err = GetLeafByAccountId(a.statedb, payload.GetFromAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.2")
	}
	if creatorLeaf == nil {
		return nil, errors.New("account not exist")
	}

	kvsCreator, l2LogCreator, _, err := applyL2AccountUpdate(creatorLeaf.GetAccountId(), zt.SystemNFTTokenId, "1", zt.Add, a.statedb, creatorLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, kvsCreator...)
	l2LogCreator.Ty = zt.TyMintNFTLog
	logs = append(logs, l2LogCreator)
	for _, kv := range kvsCreator {
		//因为在接下来的处理中需要用到这些状态信息，所以需要先将其设置到状态中
		_ = a.statedb.Set(kv.Key, kv.Value)
	}
	systemNFToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, zt.SystemNFTTokenId)
	if err != nil || nil == systemNFToken {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	//serialId = createor创建nft的次数　- 1 ,
	timesCreate, _ := new(big.Int).SetString(systemNFToken.Balance, 10)
	creatorSerialId := new(big.Int).SetInt64(timesCreate.Int64() - 1).String()
	creatorEthAddr := creatorLeaf.EthAddress

	//3. SystemNFTAccountId's SystemNFTTokenId+1, 产生新的NFT的id
	systemNFTLeaf, err := GetLeafByAccountId(a.statedb, zt.SystemNFTAccountId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.NFTAccountId")
	}
	if systemNFTLeaf == nil {
		return nil, errors.New("account not exist")
	}
	kvSystemNFTAcc, l2LogSystemNFTAcc, _, err := applyL2AccountUpdate(systemNFTLeaf.GetAccountId(), zt.SystemNFTTokenId, "1", zt.Add, a.statedb, systemNFTLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, kvSystemNFTAcc...)
	l2LogSystemNFTAcc.Ty = zt.TyMintNFTLog
	logs = append(logs, l2LogSystemNFTAcc)

	for _, kv := range kvSystemNFTAcc {
		//因为在接下来的处理中需要用到这些状态信息，所以需要先将其设置到状态中
		_ = a.statedb.Set(kv.Key, kv.Value)
	}
	systemNFToken, err = GetTokenByAccountIdAndTokenId(a.statedb, systemNFTLeaf.GetAccountId(), zt.SystemNFTTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}

	newNFTTokenId, ok := big.NewInt(0).SetString(systemNFToken.Balance, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "new NFT token balance=%s nok", systemNFToken.Balance)
	}
	if newNFTTokenId.Uint64() - 1 <= zt.SystemNFTTokenId {
		return nil, errors.Wrapf(types.ErrNotAllow, "newNFTTokenId=%d should big than default %d", newNFTTokenId.Uint64(), zt.SystemNFTTokenId)
	}

	serialId, ok := big.NewInt(0).SetString(creatorSerialId, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "creatorSerialId=%s nok", creatorSerialId)
	}

	//4. SystemNFTAccountId set new NFT id to balance by NFT contentHash
    //将系统用户名下account = SystemNFTAccountId　且　tokenID = creatorSerialId指定的token balance 设置为NFT contentHash
	newNFTTokenBalance, err := getNewNFTTokenBalance(payload.GetFromAccountId(), creatorSerialId, payload.ErcProtocol, payload.Amount, contentPart1.String(), contentPart2.String())
	if err != nil {
		return nil, errors.Wrapf(err, "getNewNFTToken balance")
	}
	kvSystemNFTAcc, l2LogSystemNFTAcc, _, err = applyL2AccountUpdate(systemNFTLeaf.GetAccountId(), newNFTTokenId.Uint64(), newNFTTokenBalance, zt.Add, a.statedb, systemNFTLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, kvSystemNFTAcc...)
	l2LogSystemNFTAcc.Ty = zt.TyMintNFT2SystemLog
	logs = append(logs, l2LogSystemNFTAcc)

	//5. recipientAddr new NFT id balance+amount
	//将最新的nft铸造给recipientAddr,
	toLeaf, err := GetLeafByAccountId(a.statedb, payload.GetRecipientId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.recipientId")
	}
	if toLeaf == nil {
		return nil, errors.New("account not exist")
	}
	for _, i := range toLeaf.TokenIds {
		if i == newNFTTokenId.Uint64() {
			return nil, errors.Wrapf(types.ErrNotAllow, "recipient has the newNFTTokenId=%d", newNFTTokenId.Uint64())
		}
	}
	kvsToAcc, l2LogToAcc, _, err := applyL2AccountUpdate(toLeaf.GetAccountId(), newNFTTokenId.Uint64(), big.NewInt(0).SetUint64(payload.Amount).String(), zt.Add, a.statedb, toLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, kvsToAcc...)
	l2LogToAcc.Ty = zt.TyMintNFTLog
	logs = append(logs, l2LogToAcc)

	//set NFT token status
	nftStatus := &zt.ZkNFTTokenStatus{
		Id:              newNFTTokenId.Uint64(),
		CreatorId:       payload.GetFromAccountId(),
		CreatorEthAddr:  creatorEthAddr,
		CreatorSerialId: serialId.Uint64(),
		ErcProtocol:     payload.ErcProtocol,
		MintAmount:      payload.Amount,
		ContentHash:     fullContent,
	}
	kv := &types.KeyValue{
		Key:   GetNFTIdPrimaryKey(nftStatus.Id),
		Value: types.Encode(nftStatus),
	}
	kvs = append(kvs, kv)

	// content hash -> nft id
	kvId := &types.KeyValue{
		Key:   GetNFTHashPrimaryKey(nftStatus.ContentHash),
		Value: types.Encode(&types.Int64{Data: int64(nftStatus.Id)}),
	}
	kvs = append(kvs, kvId)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeAmount, feeTokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

//计数新NFT Id的balance 参数hash作为其balance，不可变
func getNewNFTTokenBalance(creatorId uint64, creatorSerialId string, protocol, amount uint64, contentHashPart1, contentHashPart2 string) (string, error) {
	hashFn := mimc.NewMiMC(zt.ZkMimcHashSeed)
	hashFn.Reset()
	hashFn.Write(zt.Str2Byte(big.NewInt(0).SetUint64(creatorId).String()))
	hashFn.Write(zt.Str2Byte(creatorSerialId))
	//nft protocol
	hashFn.Write(zt.Str2Byte(big.NewInt(0).SetUint64(protocol).String()))
	//mint amount
	hashFn.Write(zt.Str2Byte(big.NewInt(0).SetUint64(amount).String()))
	hashFn.Write(zt.Str2Byte(contentHashPart1))
	hashFn.Write(zt.Str2Byte(contentHashPart2))
	//只取后面16byte，和balance可表示的最大字节数一致
	return zt.Byte2Str(hashFn.Sum(nil)[16:]), nil
}

func (a *Action) withdrawNFT(payload *zt.ZkWithdrawNFT) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	if !checkIsNFTToken(payload.NFTTokenId) {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenId=%d should big than system NFT base ID=%d", payload.NFTTokenId, zt.SystemNFTTokenId)
	}
	if payload.Amount <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong amount=%d", payload.Amount)
	}

	//暂定0 后面从数据库读取 TODO
	feeTokenId := uint64(0)
	feeAmount := zt.FeeMap[zt.TyWithdrawNFTAction]

	amountStr := big.NewInt(0).SetUint64(payload.Amount).String()
	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyWithdrawNFTAction,
		TokenID:     payload.NFTTokenId,
		Amount:      amountStr,
		FeeAmount:   feeAmount,
		SigData:     payload.Signature,
		AccountID:   payload.FromAccountId,
		SpecialInfo: &zt.OperationSpecialInfo{},
	}

	nftStatus, err := getNFTById(a.statedb, payload.NFTTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getNFTById=%d", payload.NFTTokenId)
	}

	contentHashPart1, contentHashPart2, _, err := zt.SplitNFTContent(nftStatus.ContentHash)
	if err != nil {
		return nil, errors.Wrapf(err, "split content hash=%s", nftStatus.ContentHash)
	}

	speciaData := &zt.OperationSpecialData{
		AccountID:   nftStatus.CreatorId,
		ContentHash: []string{contentHashPart1.String(), contentHashPart2.String()},
		TokenID:     []uint64{feeTokenId, nftStatus.Id, nftStatus.CreatorSerialId},
		Amount:      []string{big.NewInt(0).SetUint64(nftStatus.ErcProtocol).String(), big.NewInt(0).SetUint64(nftStatus.MintAmount).String()},
	}
	operationInfo.SpecialInfo.SpecialDatas = append(operationInfo.SpecialInfo.SpecialDatas, speciaData)

	//1. calc fee
	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.FromAccountId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	err = authVerification(payload.Signature.PubKey, fromLeaf.PubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}
	feeToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(feeToken, feeAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	feeKVsFrom, l2feeLogFrom, _, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), feeTokenId, feeAmount, zt.Sub, a.statedb, fromLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, feeKVsFrom...)
	l2feeLogFrom.Ty = zt.TyFeeLog
	logs = append(logs, l2feeLogFrom)

	//2. from NFT id -amount
	fromLeaf, err = GetLeafByAccountId(a.statedb, payload.GetFromAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.2")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	withdrawKVsFrom, l2WithdrawLogFrom, _, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), payload.NFTTokenId, amountStr, zt.Sub, a.statedb, fromLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, withdrawKVsFrom...)
	l2WithdrawLogFrom.Ty = zt.TyWithdrawNFTLog
	logs = append(logs, l2WithdrawLogFrom)

	//accumulate NFT id burned amount
	nftStatus.BurnedAmount += payload.Amount
	kv := &types.KeyValue{
		Key:   GetNFTIdPrimaryKey(nftStatus.Id),
		Value: types.Encode(nftStatus),
	}
	kvs = append(kvs, kv)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeAmount, feeTokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

func getNFTById(db dbm.KV, id uint64) (*zt.ZkNFTTokenStatus, error) {
	if id <= zt.SystemNFTTokenId {
		return nil, errors.Wrapf(types.ErrInvalidParam, "nft id =%d should big than default %d", id, zt.SystemNFTTokenId)
	}

	var nft zt.ZkNFTTokenStatus
	val, err := db.Get(GetNFTIdPrimaryKey(id))
	if err != nil {
		return nil, err
	}

	err = types.Decode(val, &nft)
	if err != nil {
		return nil, err
	}
	return &nft, nil
}

func getNFTIdByHash(db dbm.KV, hash string) (*types.Int64, error) {

	var id types.Int64
	val, err := db.Get(GetNFTHashPrimaryKey(hash))
	if err != nil {
		return nil, err
	}

	err = types.Decode(val, &id)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (a *Action) transferNFT(payload *zt.ZkTransferNFT) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	if !checkIsNFTToken(payload.NFTTokenId) {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenId=%d should big than system NFT base ID=%d", payload.NFTTokenId, zt.SystemNFTTokenId)
	}
	if payload.Amount <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong amount=%d", payload.Amount)
	}

	//暂定0 后面从数据库读取 TODO
	feeTokenId := uint64(0)
	feeAmount := zt.FeeMap[zt.TyTransferNFTAction]

	amountStr := big.NewInt(0).SetUint64(payload.Amount).String()
	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyTransferNFTAction,
		TokenID:     payload.NFTTokenId,
		Amount:      amountStr,
		FeeAmount:   feeAmount,
		SigData:     payload.Signature,
		AccountID:   payload.FromAccountId,
		SpecialInfo: &zt.OperationSpecialInfo{},
	}

	nftStatus, err := getNFTById(a.statedb, payload.NFTTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getNFTById=%d", payload.NFTTokenId)
	}

	speciaData := &zt.OperationSpecialData{
		RecipientID: payload.RecipientId,
		TokenID:     []uint64{feeTokenId, nftStatus.Id},
	}
	operationInfo.SpecialInfo.SpecialDatas = append(operationInfo.SpecialInfo.SpecialDatas, speciaData)

	//1. calc fee
	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.FromAccountId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	err = authVerification(payload.Signature.PubKey, fromLeaf.PubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}
	feeToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(feeToken, feeAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	feeKVsFrom, l2feeLogFrom, _, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), feeTokenId, feeAmount, zt.Sub, a.statedb, fromLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, feeKVsFrom...)
	l2feeLogFrom.Ty = zt.TyFeeLog
	logs = append(logs, l2feeLogFrom)

	//2. from NFT id balance-amount
	transferKVsFrom, _, receiptFrom, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), payload.NFTTokenId, amountStr, zt.Sub, a.statedb, fromLeaf, false)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, transferKVsFrom...)

	//3. recipient NFT id balance+amount
	toLeaf, err := GetLeafByAccountId(a.statedb, payload.GetRecipientId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.2")
	}
	if toLeaf == nil {
		return nil, errors.New("account not exist")
	}
	toKVsFrom, _, receiptTo, err := applyL2AccountUpdate(toLeaf.GetAccountId(), payload.NFTTokenId, amountStr, zt.Add, a.statedb, fromLeaf, false)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, toKVsFrom...)

	transferLog := &zt.TransferReceipt4L2{
		From:receiptFrom,
		To: receiptTo,
	}
	l2TransferNFTlog := &types.ReceiptLog{
		Ty: zt.TyTransferNFTLog,
		Log: types.Encode(transferLog),
	}
	logs = append(logs, l2TransferNFTlog)


	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeAmount, feeTokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}
