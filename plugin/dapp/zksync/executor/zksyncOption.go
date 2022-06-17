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
		if zt.InvalidAccountId != lastAccountID {
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
		createKVS, l2Log, err := applyL2AccountCreate(accountID, operationInfo.TokenID, operationInfo.Amount, payload.EthAddress, payload.Chain33Addr, a.statedb)
		if nil != err {
			return nil, errors.Wrapf(err, "applyL2AccountCreate")
		}

		kvs = append(kvs, createKVS...)
		l2Log.Ty = int32(zt.TyDepositLog)
		logs = append(logs, l2Log)
	} else {
		updateKVs, l2Log, err := applyL2AccountUpdate(leaf.AccountId, operationInfo.TokenID, operationInfo.Amount, zt.Add, a.statedb, leaf)
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

func getBranchByReceipt(receipt *zt.ZkReceiptLeaf, opInfo *zt.OperationInfo, ethAddr string, chain33Addr string,
	pubKey *zt.ZkPubKey, proxyPubKeys *zt.AccountProxyPubKeys, accountId, tokenId uint64, balance string) *zt.OperationMetaBranch {
	opInfo.Roots = append(opInfo.Roots, receipt.TreeProof.RootHash)
	treePath := &zt.SiblingPath{
		Path:   receipt.TreeProof.ProofSet,
		Helper: receipt.TreeProof.GetHelpers(),
	}
	accountW := &zt.AccountWitness{
		ID:           accountId,
		EthAddr:      ethAddr,
		Chain33Addr:  chain33Addr,
		PubKey:       pubKey,
		ProxyPubKeys: proxyPubKeys,
		Sibling:      treePath,
	}
	//token不存在生成默认TokenWitness
	if receipt.GetTokenProof() == nil {
		accountW.TokenTreeRoot = "0"
		return &zt.OperationMetaBranch{
			AccountWitness: accountW,
			TokenWitness: &zt.TokenWitness{
				ID:      tokenId,
				Balance: "0",
			},
		}
	}
	accountW.TokenTreeRoot = receipt.GetTokenProof().RootHash
	tokenPath := &zt.SiblingPath{
		Path:   receipt.TokenProof.ProofSet,
		Helper: receipt.TokenProof.GetHelpers(),
	}
	//如果设置balance为nil，则设为缺省0
	if len(balance) == 0 {
		balance = "0"

		if accountId == zt.SystemNFTAccountId && tokenId == zt.SystemNFTTokenId {
			balance = new(big.Int).SetUint64(zt.SystemNFTTokenId + 1).String()
		}
	}
	tokenW := &zt.TokenWitness{
		ID:      tokenId,
		Balance: balance,
		Sibling: tokenPath,
	}

	branch := &zt.OperationMetaBranch{
		AccountWitness: accountW,
		TokenWitness:   tokenW,
	}
	return branch
}

//涉及可能初始化的操作调用
func generateTreeUpdateInfo(stateDb dbm.KV, localDb dbm.KVDB, cfgEthFeeAddr, cfgChain33FeeAddr string) (*TreeUpdateInfo, error) {
	info, err := getTreeUpdateInfo(stateDb)
	if info != nil {
		return info, nil
	}
	//没查到就先初始化
	if err == types.ErrNotFound {
		updateMap := make(map[string][]byte)
		kvs, accountTable := NewAccountTree(localDb, cfgEthFeeAddr, cfgChain33FeeAddr)
		for _, kv := range kvs {
			updateMap[string(kv.GetKey())] = kv.GetValue()
		}
		return &TreeUpdateInfo{updateMap: updateMap, kvs: kvs, localKvs: make([]*types.KeyValue, 0), accountTable: accountTable}, nil
	} else {
		return nil, err
	}

}

func getTreeUpdateInfo(stateDb dbm.KV) (*TreeUpdateInfo, error) {
	updateMap := make(map[string][]byte)
	val, err := stateDb.Get(GetAccountTreeKey())
	//系统一定从deposit开始，在deposit里面初始化，非deposit操作如果获取不到返回错误
	if err != nil {
		return nil, err
	}
	var tree zt.AccountTree
	err = types.Decode(val, &tree)
	if err != nil {
		return nil, err
	}
	updateMap[string(GetAccountTreeKey())] = types.Encode(&tree)
	return &TreeUpdateInfo{updateMap: updateMap, kvs: make([]*types.KeyValue, 0), localKvs: make([]*types.KeyValue, 0)}, nil
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

	withdrawReceiptLog := &zt.AccountTokenBalanceReceiptLog{
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

	//因为合约balance需要/1e10，因此要先去掉精度
	amountInt, _ := new(big.Int).SetString(payload.Amount, 10)
	payload.Amount = new(big.Int).Mul(new(big.Int).Div(amountInt, big.NewInt(1e10)), big.NewInt(1e10)).String()

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

	l2BalanceLog := &zt.AccountTokenBalanceReceiptLog{}
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
	//因为合约balance需要/1e10，因此要先去掉精度
	amountInt, _ := new(big.Int).SetString(payload.Amount, 10)
	payload.Amount = new(big.Int).Mul(new(big.Int).Div(amountInt, big.NewInt(1e10)), big.NewInt(1e10)).String()

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

	l2BalanceLog := &zt.AccountTokenBalanceReceiptLog{}
	l2BalanceLog.EthAddress    = leaf.EthAddress
	l2BalanceLog.Chain33Addr   = leaf.Chain33Addr
	l2BalanceLog.TokenId       = payload.GetTokenId()
	l2BalanceLog.AccountId     = leaf.AccountId
	l2BalanceLog.BalanceBefore = balancehistory.before
	l2BalanceLog.BalanceAfter  = balancehistory.after
	l2Log := &types.ReceiptLog{Ty: zt.TyContractToTreeLog, Log: types.Encode(l2BalanceLog)}

	//更新合约账户
	accountKvs, l1Log, err := a.UpdateContractAccount(a.fromaddr, payload.GetAmount(), payload.GetTokenId(), zt.Add)
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
	change, _ := new(big.Int).SetString(amount, 10)
	//accountdb去除末尾10位小数
	shortChange := new(big.Int).Div(change, big.NewInt(1e10)).Int64()
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
	fromKVs, l2LogFrom, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), payload.GetTokenId(), totalAmount,zt.Sub, a.statedb, fromLeaf)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)
	l2LogFrom.Ty = zt.TyTransferLog
	logs = append(logs, l2LogFrom)

	//2.操作to 账户
	toLeaf, err := GetLeafByAccountId(a.statedb, payload.ToAccountId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if toLeaf == nil {
		return nil, errors.New("account not exist")
	}

	toKVs, l2LogTo, err := applyL2AccountUpdate(toLeaf.GetAccountId(), payload.GetTokenId(), payload.GetAmount(), zt.Add, a.statedb, toLeaf)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, toKVs...)
	l2LogTo.Ty = zt.TyTransferLog
	logs = append(logs, l2LogTo)
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
	fromKVs, l2LogFrom, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), payload.GetTokenId(), totalAmount,zt.Sub, a.statedb, fromLeaf)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)
	l2LogFrom.Ty = zt.TyTransferToNewLog
	logs = append(logs, l2LogFrom)


	//1.操作to 账户
	lastAccountID, err := getLatestAccountID(a.statedb)
	if lastAccountID == zt.InvalidAccountId {
		return nil, errors.Wrapf(err, "getLatestAccountID")
	}
	accountIDNew := uint64(lastAccountID) + 1

	toKVs, l2LogTo, err := applyL2AccountCreate(accountIDNew, operationInfo.TokenID, operationInfo.Amount, payload.GetToEthAddress(),  payload.GetToChain33Address(), a.statedb)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountCreate")
	}
	kvs = append(kvs, toKVs...)
	l2LogTo.Ty = zt.TyTransferToNewLog
	logs = append(logs, l2LogTo)

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

	fromKVs, l2LogFrom, err := applyL2AccountUpdate(leaf.GetAccountId(), payload.GetTokenId(), token.Balance, zt.Sub, a.statedb, leaf)
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

func calProof(statedb dbm.KV, info *TreeUpdateInfo, accountId uint64, tokenId uint64) (*zt.ZkReceiptLeaf, error) {
	receipt := &zt.ZkReceiptLeaf{}

	leaf, err := GetLeafByAccountId(statedb, accountId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	receipt.Leaf = leaf

	token, err := GetTokenByAccountIdAndTokenId(statedb, accountId, tokenId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	receipt.Token = token

	leafProof, err := CalLeafProof(statedb, leaf, info)
	if err != nil {
		return nil, errors.Wrapf(err, "CalLeafProof")
	}
	receipt.TreeProof = leafProof

	tokenProof, err := CalTokenProof(statedb, leaf, token, info)
	if err != nil {
		return nil, errors.Wrapf(err, "CalTokenProof")
	}
	receipt.TokenProof = tokenProof

	return receipt, nil
}

func (a *Action) SetPubKey(payload *zt.ZkSetPubKey) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}

	leaf, err := GetLeafByAccountId(a.statedb, payload.GetAccountId(), info)
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
		if zt.Byte2Str(hash.Sum(nil)) != leaf.Chain33Addr {
			return nil, errors.New("not your account")
		}
	}
	if payload.PubKeyTy > zt.SuperProxyPubKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong proxy ty=%d", payload.PubKeyTy)
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TySetPubKeyAction,
		TokenID:     leaf.TokenIds[0],
		Amount:      "0",
		SigData:     payload.Signature,
		AccountID:   payload.AccountId,
		SpecialInfo: new(zt.OperationSpecialInfo),
	}

	specialData := &zt.OperationSpecialData{
		PubKeyType: payload.PubKeyTy,
		PubKey:     payload.PubKey,
	}
	if payload.PubKeyTy == 0 {
		specialData.PubKey = payload.Signature.PubKey
	}
	operationInfo.SpecialInfo.SpecialDatas = append(operationInfo.SpecialInfo.SpecialDatas, specialData)

	if payload.PubKeyTy == 0 {
		kvs, localKvs, err = a.SetDefultPubKey(payload, info, leaf, operationInfo)
		if err != nil {
			return nil, errors.Wrapf(err, "setDefultPubKey")
		}
	} else {
		kvs, localKvs, err = a.SetProxyPubKey(payload, info, leaf, operationInfo)
		if err != nil {
			return nil, errors.Wrapf(err, "setDefultPubKey")
		}
	}

	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TySetPubKeyLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) SetDefultPubKey(payload *zt.ZkSetPubKey, info *TreeUpdateInfo, leaf *zt.Leaf, operationInfo *zt.OperationInfo) ([]*types.KeyValue, []*types.KeyValue, error) {

	//更新之前先计算证明
	receipt, err := calProof(a.statedb, info, payload.AccountId, leaf.TokenIds[0])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}
	before := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, nil, nil, operationInfo.AccountID, operationInfo.TokenID, receipt.Token.Balance)

	kvs, localKvs, err := UpdatePubKey(a.statedb, a.localDB, info, payload.GetPubKeyTy(), payload.GetPubKey(), payload.AccountId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.AccountId, leaf.TokenIds[0])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}
	after := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, payload.PubKey, nil, operationInfo.AccountID, operationInfo.TokenID, receipt.Token.Balance)
	rootHash := zt.Str2Byte(receipt.TreeProof.RootHash)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: rootHash,
	}
	kvs = append(kvs, kv)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)

	return kvs, localKvs, nil
}

//设置代理地址的公钥
func (a *Action) SetProxyPubKey(payload *zt.ZkSetPubKey, info *TreeUpdateInfo, leaf *zt.Leaf, operationInfo *zt.OperationInfo) ([]*types.KeyValue, []*types.KeyValue, error) {

	err := authVerification(payload.Signature.PubKey, leaf.PubKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "authVerification")
	}

	//更新之前先计算证明
	receipt, err := calProof(a.statedb, info, payload.AccountId, leaf.TokenIds[0])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}
	before := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, operationInfo.AccountID, operationInfo.TokenID, receipt.Token.Balance)

	kvs, localKvs, err := UpdatePubKey(a.statedb, a.localDB, info, payload.PubKeyTy, payload.GetPubKey(), payload.AccountId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.AccountId, leaf.TokenIds[0])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}
	after := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, receipt.Leaf.GetProxyPubKeys(), operationInfo.AccountID, operationInfo.TokenID, receipt.Token.Balance)
	rootHash := zt.Str2Byte(receipt.TreeProof.RootHash)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: rootHash,
	}
	kvs = append(kvs, kv)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)

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

	fromKVs, l2LogFrom, err := applyL2AccountUpdate(leaf.GetAccountId(), payload.GetTokenId(), token.Balance, zt.Sub, a.statedb, leaf)
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

	toKVs, l2Log, err := applyL2AccountUpdate(leaf.GetAccountId(), tokenId, amount,zt.Add, a.statedb, leaf)
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

	feeKVsFrom, l2feeLogFrom, err := applyL2AccountUpdate(creatorLeaf.GetAccountId(),feeTokenId, feeAmount, zt.Sub, a.statedb, creatorLeaf)
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

	kvsCreator, l2LogCreator, err := applyL2AccountUpdate(creatorLeaf.GetAccountId(), zt.SystemNFTTokenId, "1", zt.Add, a.statedb, creatorLeaf)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, kvsCreator...)
	l2LogCreator.Ty = zt.TyMintNFTLog
	logs = append(logs, l2LogCreator)
	systemNFToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, zt.SystemNFTTokenId)
	if err != nil {
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
	kvSystemNFTAcc, l2LogSystemNFTAcc, err := applyL2AccountUpdate(systemNFTLeaf.GetAccountId(), zt.SystemNFTTokenId, "1", zt.Add, a.statedb, systemNFTLeaf)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, kvSystemNFTAcc...)
	l2LogSystemNFTAcc.Ty = zt.TyMintNFTLog
	logs = append(logs, l2LogSystemNFTAcc)

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
	//operationInfo.SpecialInfo.SpecialDatas[0].TokenID = append(operationInfo.SpecialInfo.SpecialDatas[0].TokenID, newNFTTokenId.Uint64())
	serialId, ok := big.NewInt(0).SetString(creatorSerialId, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "creatorSerialId=%s nok", creatorSerialId)
	}
	//operationInfo.SpecialInfo.SpecialDatas[0].TokenID = append(operationInfo.SpecialInfo.SpecialDatas[0].TokenID, serialId.Uint64())

	//4. SystemNFTAccountId set new NFT id to balance by NFT contentHash
    //将系统用户名下account = SystemNFTAccountId　且　tokenID = creatorSerialId指定的token balance 设置为NFT contentHash
	newNFTTokenBalance, err := getNewNFTTokenBalance(payload.GetFromAccountId(), creatorSerialId, payload.ErcProtocol, payload.Amount, contentPart1.String(), contentPart2.String())
	if err != nil {
		return nil, errors.Wrapf(err, "getNewNFTToken balance")
	}
	//operationInfo.SpecialInfo.SpecialDatas[0].ContentHash = []string{contentPart1.String(), contentPart2.String()}
	kvSystemNFTAcc, l2LogSystemNFTAcc, err = applyL2AccountUpdate(systemNFTLeaf.GetAccountId(), newNFTTokenId.Uint64(), newNFTTokenBalance, zt.Add, a.statedb, systemNFTLeaf)
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
	kvsToAcc, l2LogToAcc, err := applyL2AccountUpdate(toLeaf.GetAccountId(), newNFTTokenId.Uint64(), big.NewInt(0).SetUint64(payload.Amount).String(), zt.Add, a.statedb, toLeaf)
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

func (a *Action) updateLeafRst(info *TreeUpdateInfo, opInfo *zt.OperationInfo, fromLeaf *zt.Leaf,
	tokenId uint64, amount string, option int32) (*zt.OperationPairBranch, []*types.KeyValue, []*types.KeyValue, error) {
	receipt, err := calProof(a.statedb, info, fromLeaf.GetAccountId(), tokenId)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "calProof")
	}
	before := getBranchByReceipt(receipt, opInfo, fromLeaf.EthAddress, fromLeaf.Chain33Addr, fromLeaf.PubKey, fromLeaf.ProxyPubKeys, fromLeaf.GetAccountId(), tokenId, receipt.GetToken().GetBalance())
	//更新fromLeaf
	fromKvs, fromLocal, err := UpdateLeaf(a.statedb, a.localDB, info, fromLeaf.GetAccountId(), tokenId, amount, option)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, fromLeaf.GetAccountId(), tokenId)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "calProof")
	}
	after := getBranchByReceipt(receipt, opInfo, fromLeaf.EthAddress, fromLeaf.Chain33Addr, fromLeaf.PubKey, fromLeaf.ProxyPubKeys, fromLeaf.GetAccountId(), tokenId, receipt.GetToken().GetBalance())
	return &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}, fromKvs, fromLocal, nil

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
	var localKvs []*types.KeyValue

	if !checkIsNFTToken(payload.NFTTokenId) {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenId=%d should big than system NFT base ID=%d", payload.NFTTokenId, zt.SystemNFTTokenId)
	}
	if payload.Amount <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong amount=%d", payload.Amount)
	}

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
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
	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.FromAccountId, info)
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
	feeToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, feeTokenId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(feeToken, feeAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	newBranch, fromKvs, fromLocal, err := a.updateLeafRst(info, operationInfo, fromLeaf, feeTokenId, feeAmount, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.fee")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	//2. from NFT id -amount
	fromLeaf, err = GetLeafByAccountId(a.statedb, payload.GetFromAccountId(), info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.2")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, fromLeaf, payload.NFTTokenId, amountStr, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.from.nftToken")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	//3.  校验SystemNFTAccountId's TokenId's balance same
	fromLeaf, err = GetLeafByAccountId(a.statedb, zt.SystemNFTAccountId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.NFTAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	//amount=0, just get proof
	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, fromLeaf, payload.NFTTokenId, "0", zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.from.nftToken")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	tokenBalance, err := getNewNFTTokenBalance(nftStatus.CreatorId,
		big.NewInt(0).SetUint64(nftStatus.CreatorSerialId).String(),
		nftStatus.ErcProtocol, nftStatus.MintAmount,
		contentHashPart1.String(), contentHashPart2.String())
	if err != nil {
		return nil, errors.Wrapf(err, "getNewNFTTokenBalance tokenId=%d", nftStatus.Id)
	}
	if newBranch.After.TokenWitness.Balance != tokenBalance {
		return nil, errors.Wrapf(types.ErrInvalidParam, "tokenId=%d,NFTAccount.balance=%s,calcBalance=%s", nftStatus.Id, newBranch.After.TokenWitness.Balance, tokenBalance)
	}

	//3.  校验NFT creator's eth addr same
	fromLeaf, err = GetLeafByAccountId(a.statedb, nftStatus.GetCreatorId(), info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.NFTAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	//amount=0, just get proof
	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, fromLeaf, zt.SystemNFTTokenId, "0", zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.from.nftToken")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)
	if fromLeaf.EthAddress != nftStatus.CreatorEthAddr {
		return nil, errors.Wrapf(types.ErrNotAllow, "creator eth Addr=%s, nft=%s", fromLeaf.EthAddress, nftStatus.CreatorEthAddr)
	}

	//accumulate NFT id burned amount
	nftStatus.BurnedAmount += payload.Amount
	kv := &types.KeyValue{
		Key:   GetNFTIdPrimaryKey(nftStatus.Id),
		Value: types.Encode(nftStatus),
	}
	kvs = append(kvs, kv)

	//end
	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyWithdrawNFTLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeAmount, info, feeTokenId, payload.Signature)
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
	var localKvs []*types.KeyValue

	if !checkIsNFTToken(payload.NFTTokenId) {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenId=%d should big than system NFT base ID=%d", payload.NFTTokenId, zt.SystemNFTTokenId)
	}
	if payload.Amount <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong amount=%d", payload.Amount)
	}

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
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
	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.FromAccountId, info)
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
	feeToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, feeTokenId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(feeToken, feeAmount)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	newBranch, fromKvs, fromLocal, err := a.updateLeafRst(info, operationInfo, fromLeaf, feeTokenId, feeAmount, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.fee")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	//2. from NFT id balance-amount
	fromLeaf, err = GetLeafByAccountId(a.statedb, payload.GetFromAccountId(), info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.2")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, fromLeaf, payload.NFTTokenId, amountStr, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.from.nftToken")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	//2. recipient NFT id balance+amount
	fromLeaf, err = GetLeafByAccountId(a.statedb, payload.GetRecipientId(), info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.2")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, fromLeaf, payload.NFTTokenId, amountStr, zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.from.nftToken")
	}

	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	//end
	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyTransferNFTLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeAmount, info, feeTokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}
