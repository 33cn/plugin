package executor

import (
	"fmt"
	"math/big"
	"sort"
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
	newAddr, ok := zt.HexAddr2Decimal(payload.Chain33Addr)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer chain33Addr=%s", payload.Chain33Addr)
	}
	payload.Chain33Addr = newAddr

	newAddr, ok = zt.HexAddr2Decimal(payload.EthAddress)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer EthAddress=%s", payload.EthAddress)
	}
	payload.EthAddress = newAddr

	lastPriority, err := getLastEthPriorityQueueID(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "get eth last priority queue id")
	}
	//note:需要为string，不能为int64,因为db不支持ID=0的时候
	lastPriorityId, ok := big.NewInt(0).SetString(lastPriority.GetID(), 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, fmt.Sprintf("lastPriorityID=%s", lastPriority.GetID()))
	}
	if lastPriorityId.Int64()+1 != payload.L1PriorityId {
		return nil, errors.Wrapf(types.ErrNotAllow, "eth last priority queue id=%d,new=%d", lastPriorityId, payload.L1PriorityId)
	}

	leaf, err := GetLeafByChain33AndEthAddress(a.statedb, payload.GetChain33Addr(), payload.GetEthAddress())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}

	special := &zt.ZkDepositWitnessInfo{
		//accountId nil
		TokenID:      payload.TokenId,
		Amount:       payload.Amount,
		EthAddress:   payload.EthAddress,
		Layer2Addr:   payload.Chain33Addr,
		L1PriorityID: payload.L1PriorityId,
		BlockInfo:    &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}

	//leaf不存在就添加
	if leaf == nil {
		zklog.Debug("zksync deposit add new leaf")

		var accountID uint64
		lastAccountID, err := getLatestAccountID(a.statedb)
		if err != nil {
			return nil, errors.Wrapf(err, "getLatestAccountID")
		}
		if zt.SystemDefaultAcctId == lastAccountID {
			ethFeeAddr, chain33FeeAddr := getCfgFeeAddr(cfg)
			kvs4InitAccount, err := NewInitAccount(ethFeeAddr, chain33FeeAddr)
			if nil != err {
				return nil, err
			}
			kvs = append(kvs, kvs4InitAccount...)

			//第一笔存款不允许是SystemFeeAddr，简单处理
			if payload.EthAddress == ethFeeAddr && payload.Chain33Addr == chain33FeeAddr {
				return nil, errors.Wrapf(types.ErrInvalidParam, "first deposit with systemFeeAddr not allow")
			}
			//对于首次进行存款的用户，其账户ID从SystemNFTAccountId后开始进行连续分配
			accountID = zt.SystemTree2ContractAcctId + 1
		} else {
			accountID = uint64(lastAccountID) + 1
		}
		special.AccountID = accountID
		//accountID, tokenID uint64, amount, ethAddress,  chain33Addr string, statedb dbm.KV, leaf *zt.Leaf) ([]*types.KeyValue, *types.ReceiptLog, error)
		createKVS, l2Log, _, err := applyL2AccountCreate(accountID, special.TokenID, special.Amount, payload.EthAddress, payload.Chain33Addr, a.statedb, true)
		if nil != err {
			return nil, errors.Wrapf(err, "applyL2AccountCreate")
		}

		kvs = append(kvs, createKVS...)
		l2Log.Ty = int32(zt.TyDepositLog)
		logs = append(logs, l2Log)
	} else {
		updateKVs, l2Log, _, err := applyL2AccountUpdate(leaf.AccountId, special.TokenID, special.Amount, zt.Add, a.statedb, leaf, true)
		if nil != err {
			return nil, errors.Wrapf(err, "applyL2AccountUpdate")
		}
		special.AccountID = leaf.AccountId
		kvs = append(kvs, updateKVs...)
		l2Log.Ty = int32(zt.TyDepositLog)
		logs = append(logs, l2Log)
	}
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	//add priority part
	r := makeSetL1PriorityIdReceipt(lastPriorityId.Int64(), payload.L1PriorityId)
	mergeReceipt(receipts, r)

	//add deposit queue
	r, lastQueueId, err := setL2QueueData(a.statedb, []*zt.ZkOperation{{Ty: zt.TyDepositAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}}}})
	if err != nil {
		return nil, err
	}
	mergeReceipt(receipts, r)
	//note: deposit 只push queue一个operation，没有fee op， lastQueueId就对应deposit op,如果加入fee，就需要调整
	r = makeSetPriority2QueIdReceipt(payload.L1PriorityId, lastQueueId)
	mergeReceipt(receipts, r)
	return receipts, nil
}

//L2 queue id 从1开始编号，跟L1 priority 不同，后者为了和eth合约编号保持一致
func setL2QueueData(db dbm.KV, ops []*zt.ZkOperation) (*types.Receipt, int64, error) {
	receipts := &types.Receipt{Ty: types.ExecOk}
	//add deposit queue
	lastQueueId, err := GetL2LastQueueId(db)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "getL2LastQueueId")
	}
	newQueId := lastQueueId
	for _, o := range ops {
		newQueId += 1
		r := makeSetL2LastQueueIdReceipt(lastQueueId, newQueId)
		mergeReceipt(receipts, r)
		r = makeSetL2QueueIdReceipt(newQueId, o)
		mergeReceipt(receipts, r)
		lastQueueId = newQueId
	}

	return receipts, lastQueueId, nil
}

func saveKvs(db dbm.KV, kvs []*types.KeyValue) error {
	for _, kv := range kvs {
		err := db.Set(kv.Key, kv.Value)
		if err != nil {
			return errors.Wrapf(err, "saveKvs k=%s", string(kv.Key))
		}
	}
	return nil
}

func (a *Action) ZkWithdraw(payload *zt.ZkWithdraw) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}

	amountPlusFee, fee, err := GetAmountWithFee(a.statedb, zt.TyWithdrawAction, payload.Amount, payload.TokenId)
	if err != nil {
		return nil, err
	}
	//如果是系统feeId提取，则不收fee
	if payload.AccountId == zt.SystemFeeAccountId {
		amountPlusFee = payload.Amount
		fee = "0"
	}

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
	err = checkAmount(token, amountPlusFee)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	special := &zt.ZkWithdrawWitnessInfo{
		TokenID:   payload.TokenId,
		Amount:    payload.Amount,
		AccountID: payload.AccountId,
		//ethAddr nil
		EthAddress: leaf.EthAddress,
		Fee: &zt.ZkFee{
			Fee: fee,
		},
		BlockInfo: &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}

	balancekv, balancehistory, err := updateTokenBalance(payload.AccountId, special.TokenID, amountPlusFee, zt.Sub, a.statedb)
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
		EthAddress:    leaf.EthAddress,
		Chain33Addr:   leaf.Chain33Addr,
		TokenId:       payload.GetTokenId(),
		AccountId:     leaf.AccountId,
		BalanceBefore: balancehistory.before,
		BalanceAfter:  balancehistory.after,
	}

	receiptLog := &types.ReceiptLog{Ty: zt.TyWithdrawLog, Log: types.Encode(withdrawReceiptLog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	//在acctId=SystemFeeAccountId 时候未把kv设进fee，和下面的fee op处理冲突，这里需要把kv设进db
	err = saveKvs(a.statedb, receipts.KV)
	if err != nil {
		return nil, err
	}

	feeReceipt, feeQueue, err := a.MakeFeeLog(fee, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	//add  withdraw & fee queue
	var ops []*zt.ZkOperation
	ops = append(ops, &zt.ZkOperation{Ty: zt.TyWithdrawAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Withdraw{Withdraw: special}}})
	ops = append(ops, feeQueue)
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}

	return mergeReceipt(receipts, r), nil
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
	//保证精度是小数后8位，eth转换去掉10位
	if a.api.GetConfig().GetCoinPrecision() != types.DefaultCoinPrecision {
		return nil, errors.Wrapf(types.ErrInvalidParam, "coin precision is not defual=%d", types.DefaultCoinPrecision)
	}

	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	err = checkPackValue(payload.Amount, zt.PacAmountManBitWidth)
	if err != nil {
		return nil, errors.Wrapf(err, "checkPackVal")
	}

	token, err := GetTokenBySymbol(a.statedb, payload.TokenSymbol)
	if err != nil {
		return nil, err
	}

	//如果设置了toAccountId 直接转到id，否则根据ethAddr、layer2Addr创建新account
	if payload.GetToAccountId() > 0 {
		return a.contractToTreeAcctIdProc(payload, token)
	}

	return a.contractToTreeNewProc(payload, token)
}

//合约----> L2账户操作，
//SystemTree2ContractAcctId ----> 目的账户
//在合约上销毁等量的余额
func (a *Action) contractToTreeAcctIdProc(payload *zt.ZkContractToTree, token *zt.ZkTokenSymbol) (*types.Receipt, error) {
	tokenIdBigint, _ := new(big.Int).SetString(token.Id, 10)
	tokenId := tokenIdBigint.Uint64()

	////根据精度，转化到tree侧的amount
	//sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")
	//amountTree, err := TransferDecimalAmount(payload.Amount, sysDecimal, int(token.Decimal))
	//if err != nil {
	//	return nil, errors.Wrap(err, "getTreeSideAmount")
	//}
	//err = checkPackValue(amountTree, zt.PacAmountManBitWidth)
	//if err != nil {
	//	return nil, errors.Wrap(err, "checkPackValue")
	//}

	payload4transfer := &zt.ZkTransfer{
		TokenId:       tokenId,
		Amount:        payload.Amount,
		FromAccountId: zt.SystemTree2ContractAcctId,
		ToAccountId:   payload.ToAccountId,
	}
	receipts, err := a.l2TransferProc(payload4transfer, zt.TyContractToTreeAction, int(token.Decimal))
	if nil != err {
		return nil, err
	}

	amountPlusFee, _, err := GetAmountWithFee(a.statedb, zt.TyContractToTreeAction, payload.Amount, tokenId)
	if err != nil {
		return nil, err
	}
	//更新合约账户
	contractReceipt, err := a.UpdateContractAccount(amountPlusFee, payload.TokenSymbol, zt.Sub, payload.FromExec)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateContractAccount")
	}
	receipts = mergeReceipt(receipts, contractReceipt)

	return receipts, nil
}

func (a *Action) contractToTreeNewProc(payload *zt.ZkContractToTree, token *zt.ZkTokenSymbol) (*types.Receipt, error) {
	if len(payload.GetToEthAddr()) <= 0 || len(payload.GetToLayer2Addr()) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "accountId=%d,ethAddr=%s or layer2Addr=%s nil",
			payload.GetToAccountId(), payload.GetToEthAddr(), payload.GetToLayer2Addr())
	}
	//转换10进制
	newAddr, ok := zt.HexAddr2Decimal(payload.ToLayer2Addr)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer chain33Addr=%s", payload.ToLayer2Addr)
	}
	payload.ToLayer2Addr = newAddr

	newAddr, ok = zt.HexAddr2Decimal(payload.ToEthAddr)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer EthAddress=%s", payload.ToEthAddr)
	}
	payload.ToEthAddr = newAddr

	toLeaf, err := GetLeafByChain33AndEthAddress(a.statedb, payload.ToLayer2Addr, payload.ToEthAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}
	//toAccount存在，走accountId流程, 因为contract2New 电路不验证签名
	if toLeaf != nil {
		payload.ToAccountId = toLeaf.AccountId
		return a.contractToTreeAcctIdProc(payload, token)
	}

	tokenIdBigint, _ := new(big.Int).SetString(token.Id, 10)
	tokenId := tokenIdBigint.Uint64()
	amountPlusFee, fee, err := GetAmountWithFee(a.statedb, zt.TyContractToTreeAction, payload.Amount, tokenId)
	if nil != err {
		return nil, err
	}

	//更新合约账户
	contractReceipt, err := a.UpdateContractAccount(amountPlusFee, payload.TokenSymbol, zt.Sub, payload.FromExec)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateContractAccount")
	}

	//根据精度，转化到tree侧的amount
	sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")
	amountTree, amountPlusFeeTree, feeTree, err := GetTreeSideAmount(payload.Amount, amountPlusFee, fee, sysDecimal, int(token.Decimal))
	if err != nil {
		return nil, errors.Wrap(err, "getTreeSideAmount")
	}

	//sysAccount=3 -amountPlusFeeTree, toAccount +amountTree
	receiptsTransfer, toAccountId, err := a.transferToNewProcess(zt.SystemTree2ContractAcctId, payload.ToLayer2Addr, payload.ToEthAddr, amountPlusFeeTree, amountTree, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "transfer2NewProc")
	}

	feeReceipt, feeQueue, err := a.MakeFeeLog(feeTree, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}

	receipts := mergeReceipt(receiptsTransfer, contractReceipt)
	receipts = mergeReceipt(receipts, feeReceipt)

	//add  contract2treenew & fee queue
	special := &zt.ZkContractToTreeNewWitnessInfo{
		TokenID:     tokenId,
		Amount:      amountTree,
		ToAccountID: toAccountId,
		EthAddress:  payload.ToEthAddr,
		Layer2Addr:  payload.ToLayer2Addr,
		Fee:         &zt.ZkFee{Fee: feeTree},
		BlockInfo:   &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}
	var ops []*zt.ZkOperation
	ops = append(ops, &zt.ZkOperation{Ty: zt.TyContractToTreeNewAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Contract2TreeNew{Contract2TreeNew: special}}})
	ops = append(ops, feeQueue)
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}
	receipts = mergeReceipt(receipts, r)

	return receipts, nil
}

//L2 ---->  合约账户(树)
//操作１. FromAccountId -----> SystemTree2ContractAcctId，执行ZkTransfer
//操作2. UpdateContractAccount，在合约内部的铸币操作
func (a *Action) TreeToContract(payload *zt.ZkTreeToContract) (*types.Receipt, error) {
	err := checkParam(payload.Amount)
	if nil != err {
		return nil, err
	}
	//增加systemTree2ContractId 是为了验证签名，同时防止重放攻击，也可以和transfer重用电路
	if payload.ToAcctId != zt.SystemTree2ContractAcctId {
		return nil, errors.Wrapf(types.ErrInvalidParam, "toAcctId not systemId=%d", zt.SystemTree2ContractAcctId)
	}

	para := &zt.ZkTransfer{
		TokenId:       payload.TokenId,
		Amount:        payload.Amount,
		FromAccountId: payload.AccountId,
		ToAccountId:   zt.SystemTree2ContractAcctId,
		Signature:     payload.Signature,
	}
	receiptTranfer, err := a.ZkTransfer(para, zt.TyTreeToContractAction)
	if nil != err {
		return nil, err
	}

	//更新合约账户
	token, err := GetTokenByTokenId(a.statedb, strconv.Itoa(int(payload.GetTokenId())))
	if err != nil {
		return nil, err
	}
	//cfg decimal 系统启动时候做过检查，和0的个数一致，缺省1e8有8个0
	s := strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision()))
	sysDecimal := strings.Count(s, "0")
	contractAmount, err := TransferDecimalAmount(payload.GetAmount(), int(token.Decimal), sysDecimal)
	if err != nil {
		return nil, errors.Wrapf(err, "transfer2ContractAmount,tokenDecimal=%d,sysDecimal=%d", token.Decimal, sysDecimal)
	}
	accountReceipt, err := a.UpdateContractAccount(contractAmount, token.Symbol, zt.Add, payload.ToExec)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateContractAccount")
	}

	receipts := mergeReceipt(receiptTranfer, accountReceipt)
	return receipts, nil
}

func (a *Action) UpdateExecAccount(accountdb *account.DB, amount int64, option int32, execName string) (*types.Receipt, error) {
	//平行链的execName是通过title+execname注册的，比如参数execName=user.p.para.paracross,在user.p.para.的平行链上也是同步注册了的
	//所以这里exeName为paracross或者user.p.para.paracross都可以正确获取到address，但是如果是user.p.xx.paracross就会失败
	execAddr := dapp.ExecAddress(execName)
	if !dapp.IsDriverAddress(execAddr, a.height) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "execName=%s not driver", execName)
	}
	if option == zt.Sub {
		return accountdb.TransferWithdraw(a.fromaddr, execAddr, amount)
	}
	return accountdb.TransferToExec(a.fromaddr, execAddr, amount)
}

func (a *Action) UpdateContractAccount(amount, symbol string, option int32, execName string) (*types.Receipt, error) {
	//如果是exodus mode下，支持超级管理员提取剩余的资金，如果用户没有提取完的话
	if isSuperManager(a.api.GetConfig(), a.fromaddr) && nil != isExodusMode(a.statedb) {
		return nil, nil
	}

	accountdb, err := account.NewAccountDB(a.api.GetConfig(), zt.Zksync, symbol, a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "newZkSyncAccount")
	}

	shortChangeBigInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "UpdateContractAccount amount=%s", amount)
	}
	shortChange := shortChangeBigInt.Int64()
	var execReceipt types.Receipt
	if option == zt.Sub {
		if len(execName) > 0 {
			r, err := a.UpdateExecAccount(accountdb, shortChange, option, execName)
			if err != nil {
				return nil, errors.Wrapf(err, "withdraw from exec=%s,val=%d", execName, shortChange)
			}
			mergeReceipt(&execReceipt, r)
		}

		r, err := accountdb.Burn(a.fromaddr, shortChange)
		return mergeReceipt(&execReceipt, r), err
	}
	//deposit
	r, err := accountdb.Mint(a.fromaddr, shortChange)
	if err != nil {
		return nil, errors.Wrapf(err, "deposit val=%d", shortChange)
	}
	mergeReceipt(&execReceipt, r)
	if len(execName) > 0 {
		r, err = a.UpdateExecAccount(accountdb, shortChange, option, execName)
		return mergeReceipt(&execReceipt, r), err
	}
	return &execReceipt, nil
}

//trasfer, tree2contract, contract2tree 本质都是transfer，这里共用一个transferProc
func (a *Action) l2TransferProc(payload *zt.ZkTransfer, actionTy int32, decimal int) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	err = checkPackValue(payload.Amount, zt.PacAmountManBitWidth)
	if err != nil {
		return nil, errors.Wrapf(err, "checkPackVal")
	}

	if !checkIsNormalToken(payload.TokenId) {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenId=%d should less than system NFT base ID=%d", payload.TokenId, zt.SystemNFTTokenId)
	}

	amountPlusFee, fee, err := GetAmountWithFee(a.statedb, actionTy, payload.Amount, payload.TokenId)
	if err != nil {
		return nil, err
	}
	special := &zt.ZkTransferWitnessInfo{
		FromAccountID: payload.FromAccountId,
		ToAccountID:   payload.ToAccountId,
		TokenID:       payload.TokenId,
		Amount:        payload.Amount,
		Fee:           &zt.ZkFee{Fee: fee},
		BlockInfo:     &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}
	//transfer 和 tree2contract 重用电路的TyTransferAction
	operation := &zt.ZkOperation{Ty: zt.TyTransferAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Transfer{Transfer: special}}}

	if actionTy == zt.TyContractToTreeAction {
		sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")
		amountTree, amountPlusFeeTree, feeTree, err := GetTreeSideAmount(payload.Amount, amountPlusFee, fee, sysDecimal, decimal)
		if err != nil {
			return nil, errors.Wrap(err, "getTreeSideAmount")
		}
		amountPlusFee = amountPlusFeeTree
		fee = feeTree
		payload.Amount = amountTree
		//reset operation
		special2 := &zt.ZkContractToTreeWitnessInfo{
			AccountID: payload.ToAccountId,
			TokenID:   payload.TokenId,
			Amount:    payload.Amount,
			Fee:       &zt.ZkFee{Fee: fee},
			BlockInfo: &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
		}
		operation.Ty = zt.TyContractToTreeAction
		operation.Op = &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ContractToTree{ContractToTree: special2}}
	}

	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.GetFromAccountId())
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	fromToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(fromToken, amountPlusFee)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	//1.操作from 账户
	fromKVs, _, receiptFrom, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), payload.GetTokenId(), amountPlusFee, zt.Sub, a.statedb, fromLeaf, false)
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
		From: receiptFrom,
		To:   receiptTo,
	}
	l2Transferlog := &types.ReceiptLog{
		Ty:  zt.TyTransferLog,
		Log: types.Encode(transferLog),
	}
	if actionTy == zt.TyContractToTreeAction {
		l2Transferlog.Ty = zt.TyContractToTreeLog
	}
	if actionTy == zt.TyTreeToContractAction {
		l2Transferlog.Ty = zt.TyTreeToContractLog
	}
	logs = append(logs, l2Transferlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	//在acctId=SystemFeeAccountId 时候未把kv设进fee，和下面的fee op处理冲突，这里需要把kv设进db
	err = saveKvs(a.statedb, receipts.KV)
	if err != nil {
		return nil, err
	}

	//2.操作交易费账户
	feeReceipt, feeQueue, err := a.MakeFeeLog(fee, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	//add  transfer & fee queue
	var ops []*zt.ZkOperation
	ops = append(ops, operation)
	ops = append(ops, feeQueue)
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}
	receipts = mergeReceipt(receipts, r)

	return receipts, nil
}

func (a *Action) ZkTransfer(payload *zt.ZkTransfer, actionTy int32) (*types.Receipt, error) {
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

	//此处的decimal无用
	return a.l2TransferProc(payload, actionTy, 18)
}

func (a *Action) transferToNewProcess(accountIdFrom uint64, toChain33Address, toEthAddress, totalAmount, amount string, tokenID uint64) (*types.Receipt, uint64, error) {

	fromLeaf, err := GetLeafByAccountId(a.statedb, accountIdFrom)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, 0, errors.New("account not exist")
	}
	return a.transferToNewInnerProcess(fromLeaf, toChain33Address, toEthAddress, totalAmount, amount, tokenID)
}

//contract2tree 也支持tree上新创建账户id， 和transfer2new共用
func (a *Action) transferToNewInnerProcess(fromLeaf *zt.Leaf, toChain33Address, toEthAddress, totalAmount, amount string, tokenID uint64) (*types.Receipt, uint64, error) {
	var kvs []*types.KeyValue
	var logs []*types.ReceiptLog

	toLeaf, err := GetLeafByChain33AndEthAddress(a.statedb, toChain33Address, toEthAddress)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}
	if toLeaf != nil {
		return nil, 0, errors.New("to account already exist")
	}

	//1.操作from 账户
	fromKVs, _, receiptFrom, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), tokenID, totalAmount, zt.Sub, a.statedb, fromLeaf, false)
	if nil != err {
		return nil, 0, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)

	//1.操作to 账户
	lastAccountID, err := getLatestAccountID(a.statedb)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "getLatestAccountID")
	}
	if lastAccountID == zt.SystemDefaultAcctId {
		return nil, 0, errors.Wrapf(err, "getLatestAccountID")
	}
	accountIDNew := uint64(lastAccountID) + 1

	toKVs, _, receiptTo, err := applyL2AccountCreate(accountIDNew, tokenID, amount, toEthAddress, toChain33Address, a.statedb, false)
	if nil != err {
		return nil, 0, errors.Wrapf(err, "applyL2AccountCreate")
	}
	kvs = append(kvs, toKVs...)

	transferToNewLog := &zt.TransferReceipt4L2{
		From: receiptFrom,
		To:   receiptTo,
	}
	l2TransferLog := &types.ReceiptLog{
		Ty:  zt.TyTransferToNewLog,
		Log: types.Encode(transferToNewLog),
	}
	logs = append(logs, l2TransferLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	return receipts, accountIDNew, nil
}

func (a *Action) TransferToNew(payload *zt.ZkTransferToNew) (*types.Receipt, error) {
	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	amountPlusFee, fee, err := GetAmountWithFee(a.statedb, zt.TyTransferToNewAction, payload.Amount, payload.TokenId)
	if err != nil {
		return nil, err
	}

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

	//转换10进制
	newAddr, ok := zt.HexAddr2Decimal(payload.ToLayer2Address)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer toLayer2Addr=%s", payload.ToLayer2Address)
	}
	payload.ToLayer2Address = newAddr

	newAddr, ok = zt.HexAddr2Decimal(payload.ToEthAddress)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer ToEthAddress=%s", payload.ToEthAddress)
	}
	payload.ToEthAddress = newAddr

	fromToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(fromToken, amountPlusFee)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	receipts, toAccountId, err := a.transferToNewInnerProcess(fromLeaf, payload.GetToLayer2Address(), payload.GetToEthAddress(), amountPlusFee, payload.Amount, payload.GetTokenId())
	if nil != err {
		return nil, err
	}

	feeReceipt, feeQueue, err := a.MakeFeeLog(fee, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	//add  transfer2new & fee queue
	special := &zt.ZkTransferToNewWitnessInfo{
		FromAccountID: payload.GetFromAccountId(),
		TokenID:       payload.GetTokenId(),
		Amount:        payload.GetAmount(),
		ToAccountID:   toAccountId,
		EthAddress:    payload.ToEthAddress,
		Layer2Addr:    payload.ToLayer2Address,
		Fee:           &zt.ZkFee{Fee: fee},
		BlockInfo:     &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}
	var ops []*zt.ZkOperation
	ops = append(ops, &zt.ZkOperation{Ty: zt.TyTransferToNewAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TransferToNew{TransferToNew: special}}})
	ops = append(ops, feeQueue)
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}
	receipts = mergeReceipt(receipts, r)
	return receipts, nil
}

func (a *Action) ProxyExit(payload *zt.ZkProxyExit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	if payload.ProxyId == payload.TargetId {
		return nil, errors.Wrapf(types.ErrInvalidParam, "proxyId same as targetId")
	}

	feeInfo, err := GetFeeData(a.statedb, zt.TyProxyExitAction, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}
	fee := feeInfo.Fee

	targetLeaf, err := GetLeafByAccountId(a.statedb, payload.TargetId)
	if err != nil {
		return nil, errors.Wrapf(err, "GetLeafByAccountId")
	}
	//如果targetId已经设置过pubkey，说明用户地址设置没错，由用户自己提款
	if targetLeaf == nil || targetLeaf.PubKey != nil {
		return nil, errors.Wrapf(types.ErrNotAllow, "target account=%d not exist or pubkey existed", payload.TargetId)
	}

	targetToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.TargetId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	//token不存在时，不需要取
	if targetToken == nil {
		return nil, errors.New("token not find")
	}

	proxyLeaf, err := GetLeafByAccountId(a.statedb, payload.ProxyId)
	if err != nil {
		return nil, errors.Wrapf(err, "GetLeafByAccountId")
	}
	if proxyLeaf == nil {
		return nil, errors.New("proxy account not exist")
	}

	proxyToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.ProxyId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	//token不存在时，不需要取
	if proxyToken == nil {
		return nil, errors.New("token not find")
	}

	amountInt, _ := new(big.Int).SetString(proxyToken.Balance, 10)
	feeInt, _ := new(big.Int).SetString(fee, 10)
	//确保代理账户的余额足够支付手续费
	if amountInt.Cmp(feeInt) <= 0 {
		return nil, errors.New("no enough fee")
	}

	fromKVs, l2LogFrom, _, err := applyL2AccountUpdate(targetLeaf.GetAccountId(), payload.GetTokenId(), targetToken.Balance, zt.Sub, a.statedb, targetLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)
	l2LogFrom.Ty = zt.TyProxyExitLog
	logs = append(logs, l2LogFrom)

	proxyKVs, l2Logproxy, _, err := applyL2AccountUpdate(proxyLeaf.GetAccountId(), payload.GetTokenId(), fee, zt.Sub, a.statedb, targetLeaf, true)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, proxyKVs...)
	l2Logproxy.Ty = zt.TyProxyExitLog
	logs = append(logs, l2Logproxy)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	//在acctId=SystemFeeAccountId 时候未把kv设进fee，和下面的fee op处理冲突，这里需要把kv设进db
	err = saveKvs(a.statedb, receipts.KV)
	if err != nil {
		return nil, err
	}

	feeReceipt, feeQueue, err := a.MakeFeeLog(fee, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	//add  proxy exit & fee queue
	special := &zt.ZkProxyExitWitnessInfo{
		ProxyID:    payload.GetProxyId(),
		TargetID:   payload.TargetId,
		TokenID:    payload.TokenId,
		Amount:     targetToken.Balance,
		EthAddress: targetLeaf.GetEthAddress(),
		Fee:        &zt.ZkFee{Fee: fee},
		BlockInfo:  &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}
	var ops []*zt.ZkOperation
	ops = append(ops, &zt.ZkOperation{Ty: zt.TyProxyExitAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ProxyExit{ProxyExit: special}}})
	ops = append(ops, feeQueue)
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}
	receipts = mergeReceipt(receipts, r)

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
	if payload.PubKeyTy > zt.SuperProxyPubKey {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong proxy ty=%d", payload.PubKeyTy)
	}

	if payload.PubKeyTy == 0 {
		//已经设置过缺省公钥，不允许再设置
		if leaf.PubKey != nil {
			return nil, errors.Wrapf(types.ErrNotAllow, "pubKey existed already")
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
		PubKey:    payload.PubKey,
		PubKeyTy:  payload.PubKeyTy,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TySetPubKeyLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	//add  queue
	special := &zt.ZkSetPubKeyWitnessInfo{
		AccountID: payload.AccountId,
		PubKeyTy:  payload.PubKeyTy,
		PubKey:    payload.GetPubKey(),
		BlockInfo: &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}
	var ops []*zt.ZkOperation
	ops = append(ops, &zt.ZkOperation{Ty: zt.TySetPubKeyAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_SetPubKey{SetPubKey: special}}})
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}
	receipts = mergeReceipt(receipts, r)

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

//func (a *Action) FullExit(payload *zt.ZkFullExit) (*types.Receipt, error) {
//	var logs []*types.ReceiptLog
//	var kvs []*types.KeyValue
//
//	fee := zt.FeeMap[zt.TyFullExitAction]
//
//	//只有管理员能操作
//	cfg := a.api.GetConfig()
//	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, a.fromaddr) {
//		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
//	}
//
//	//fullexit last priority id 不能为空
//	lastPriority, err := getLastEthPriorityQueueID(a.statedb, 0)
//	if err != nil {
//		return nil, errors.Wrapf(err, "get eth last priority queue id")
//	}
//	lastId, ok := big.NewInt(0).SetString(lastPriority.GetID(), 10)
//	if !ok {
//		return nil, errors.Wrapf(types.ErrInvalidParam, fmt.Sprintf("getID =%s", lastPriority.GetID()))
//	}
//
//	if lastId.Int64()+1 != payload.GetEthPriorityQueueId() {
//		return nil, errors.Wrapf(types.ErrNotAllow, "eth last priority queue id=%s,new=%d", lastPriority.ID, payload.GetEthPriorityQueueId())
//	}
//
//	leaf, err := GetLeafByAccountId(a.statedb, payload.AccountId)
//	if err != nil {
//		return nil, errors.Wrapf(err, "calProof")
//	}
//
//	if leaf == nil {
//		return nil, errors.New("account not exist")
//	}
//
//	token, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.AccountId, payload.TokenId)
//	if err != nil {
//		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
//	}
//
//	//token不存在时，不需要取
//	if token == nil {
//		return nil, errors.New("token not find")
//	}
//
//	//加上手续费
//	amountInt, _ := new(big.Int).SetString(token.Balance, 10)
//	feeInt, _ := new(big.Int).SetString(fee, 10)
//	//存量不够手续费时，不能取
//	if amountInt.Cmp(feeInt) <= 0 {
//		return nil, errors.New("no enough fee")
//	}
//
//	fromKVs, l2LogFrom, _, err := applyL2AccountUpdate(leaf.GetAccountId(), payload.GetTokenId(), token.Balance, zt.Sub, a.statedb, leaf, true)
//	if nil != err {
//		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
//	}
//	kvs = append(kvs, fromKVs...)
//	l2LogFrom.Ty = zt.TyFullExitLog
//	logs = append(logs, l2LogFrom)
//
//	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
//	//add priority part
//	r := makeSetEthPriorityIdReceipt(0, lastId.Int64(), payload.EthPriorityQueueId)
//
//	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId, payload.Signature)
//	if err != nil {
//		return nil, errors.Wrapf(err, "MakeFeeLog")
//	}
//	receipts = mergeReceipt(receipts, feeReceipt)
//	return mergeReceipt(receipts, r), nil
//}

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
	if amount == "" || strings.HasPrefix(amount, "-") {
		return types.ErrAmount
	}
	v, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return errors.Wrapf(types.ErrInvalidParam, "decode amount=%s", amount)
	}
	if v.Cmp(big.NewInt(0)) == 0 {
		return errors.Wrapf(types.ErrInvalidParam, "amount=%s is 0", amount)
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

func getLastEthPriorityQueueID(db dbm.KV) (*zt.L1PriorityID, error) {
	key := getEthPriorityQueueKey()
	v, err := db.Get(key)
	//未找到返回-1
	if isNotFound(err) {
		return &zt.L1PriorityID{ID: "-1"}, nil
	}
	if err != nil {
		return nil, err
	}
	var id zt.L1PriorityID
	err = types.Decode(v, &id)
	if err != nil {
		zklog.Error("getLastEthPriorityQueueID.decode", "err", err)
		return nil, err
	}

	return &id, nil
}

func getLatestAccountID(db dbm.KV) (int64, error) {
	key := CalcLatestAccountIDKey()
	v, err := db.Get(key)
	if isNotFound(err) {
		return zt.SystemDefaultAcctId, nil
	}
	if err != nil {
		return zt.SystemDefaultAcctId, err
	}
	var id types.Int64
	err = types.Decode(v, &id)
	if err != nil {
		return zt.SystemDefaultAcctId, errors.Wrapf(err, "decode")
	}

	return id.Data, nil
}

func CalcNewAccountIDkv(accounID int64) *types.KeyValue {
	key := CalcLatestAccountIDKey()
	id := &types.Int64{
		Data: accounID,
	}
	value := types.Encode(id)
	kv := &types.KeyValue{
		Key:   key,
		Value: value,
	}
	return kv
}

func makeSetL1PriorityIdReceipt(prev, current int64) *types.Receipt {
	key := getEthPriorityQueueKey()
	log := &zt.ReceiptL1PriorityID{
		Prev:    prev,
		Current: current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&zt.L1PriorityID{ID: new(big.Int).SetInt64(current).String()})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  zt.TySetL1PriorityId,
				Log: types.Encode(log),
			},
		},
	}
}

func makeSetPriority2QueIdReceipt(priorityId, queueId int64) *types.Receipt {
	key := getL1PriorityId2QueueIdKey(priorityId)
	log := &zt.Priority2QueueId{
		PriorityId: priorityId,
		QueueId:    queueId,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&types.Int64{Data: queueId})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  zt.TyPriority2QueIdLog,
				Log: types.Encode(log),
			},
		},
	}
}

func makeSetL2LastQueueIdReceipt(prev, current int64) *types.Receipt {
	key := getL2LastQueueIdKey()
	log := &zt.ReceiptL2LastQueueID{
		Prev:    prev,
		Current: current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&types.Int64{Data: current})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  zt.TySetL2OpLastQueueIdLog,
				Log: types.Encode(log),
			},
		},
	}
}

func makeSetL2QueueIdReceipt(id int64, op *zt.ZkOperation) *types.Receipt {
	key := getL2QueueIdKey(id)
	log := &zt.ReceiptL2QueueIDData{
		Id: id,
		Op: op,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(op)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  zt.TySetL2OpQueueIdLog,
				Log: types.Encode(log),
			},
		},
	}
}

func makeSetL2FirstQueueIdReceipt(oldId, newId int64) *types.Receipt {
	key := getL2FirstQueueIdKey()
	log := &zt.ReceiptL2FirstQueueID{
		Prev:    oldId,
		Current: newId,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&types.Int64{Data: newId})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  zt.TySetL2OpFirstQueueIdLog,
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

func (a *Action) MakeFeeLog(amount string, tokenId uint64) (*types.Receipt, *zt.ZkOperation, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var err error

	//todo 手续费收款方accountId可配置
	leaf, err := GetLeafByAccountId(a.statedb, zt.SystemFeeAccountId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}

	if leaf == nil {
		return nil, nil, errors.New("account not exist")
	}

	toKVs, l2Log, _, err := applyL2AccountUpdate(leaf.GetAccountId(), tokenId, amount, zt.Add, a.statedb, leaf, true)
	if nil != err {
		return nil, nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	l2Log.Ty = zt.TyFeeLog
	logs = append(logs, l2Log)
	kvs = append(kvs, toKVs...)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	special := &zt.ZkFeeWitnessInfo{
		TokenID:   tokenId,
		Amount:    amount,
		AccountID: leaf.AccountId,
	}
	return receipts, &zt.ZkOperation{Ty: zt.TyFeeAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Fee{Fee: special}}}, nil
}

func (a *Action) setFee(payload *zt.ZkSetFee) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := a.api.GetConfig()

	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
	}
	amountInt, ok := new(big.Int).SetString(payload.Amount, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount=%s", payload.Amount)
	}
	if amountInt.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "amount=%s", payload.Amount)
	}
	//跟其他action手续费以二层各token精度一致不同，contract2tree action的手续费统一以合约的精度处理，这样和合约侧amount精度一致，简化了合约侧的精度处理
	if payload.ActionTy == zt.TyContractToTreeAction {
		token, err := GetTokenByTokenId(a.statedb, big.NewInt(int64(payload.TokenId)).String())
		if err != nil {
			return nil, errors.Wrapf(err, "getTokenId=%d", payload.TokenId)
		}
		sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")

		//比如token精度为6，sysDecimal=8, token的fee在sysDecimal下需要补2个0，也就是后缀需要至少有两个0，不然会丢失精度，在token精度大于sys精度时候没这问题
		if int(token.Decimal) < sysDecimal && amountInt.Cmp(big.NewInt(0)) > 0 && !strings.HasSuffix(payload.Amount, strings.Repeat("0", sysDecimal-int(token.Decimal))) {
			return nil, errors.Wrapf(types.ErrNotAllow, "contract2tree fee need at least with suffix=%s", strings.Repeat("0", sysDecimal-int(token.Decimal)))
		}
	}

	//fee 压缩格式检查
	err := checkPackValue(payload.Amount, zt.PacFeeManBitWidth)
	if err != nil {
		return nil, errors.Wrapf(err, "checkPackVal")
	}

	lastFee, err := getDbFeeData(a.statedb, payload.ActionTy, payload.TokenId)
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

func getDbFeeData(db dbm.KV, actionTy int32, tokenId uint64) (string, error) {
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

//该接口需要被zkrelayer使用
func GetFeeData(db dbm.KV, actionTy int32, tokenId uint64) (*zt.ZkFee, error) {
	//缺省输入的tokenId，如果swapFee有输入新tokenId，采用新的，在withdraw等action忽略swapFee tokenId
	feeInfo := &zt.ZkFee{
		TokenID: tokenId,
	}
	//从db读取
	fee, err := getDbFeeData(db, actionTy, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getDbFeeData")
	}
	feeInfo.Fee = fee

	return feeInfo, nil
}

func GetAmountWithFee(db dbm.KV, actionTy int32, amount string, tokenId uint64) (amountPlusFee, fee string, err error) {
	feeInfo, err := GetFeeData(db, actionTy, tokenId)
	if err != nil {
		return "", "", errors.Wrapf(err, "getFeeData")
	}

	//加上手续费
	amountInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", "", errors.Wrapf(types.ErrInvalidParam, "decode amount=%s", amount)
	}
	fromFeeInt, ok := new(big.Int).SetString(feeInfo.Fee, 10)
	if !ok {
		return "", "", errors.Wrapf(types.ErrInvalidParam, "fromFee=%s", feeInfo.Fee)
	}
	totalFromAmount := new(big.Int).Add(amountInt, fromFeeInt)

	return totalFromAmount.String(), fromFeeInt.String(), nil
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
	feeInfo, err := GetFeeData(a.statedb, zt.TyMintNFTAction, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}
	feeAmount := feeInfo.Fee

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
	if newNFTTokenId.Uint64()-1 <= zt.SystemNFTTokenId {
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

	feeReceipt, feeQueue, err := a.MakeFeeLog(feeAmount, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	//add  queue
	special := &zt.ZkMintNFTWitnessInfo{
		MintAcctID:     payload.GetFromAccountId(),
		RecipientID:    payload.RecipientId,
		ErcProtocol:    payload.ErcProtocol,
		ContentHash:    []string{contentPart1.String(), contentPart2.String()},
		NewNFTTokenID:  newNFTTokenId.Uint64(),
		CreateSerialID: serialId.Uint64(),
		Amount:         payload.Amount,
		Fee: &zt.ZkFee{
			Fee:     feeAmount,
			TokenID: feeTokenId,
		},
		BlockInfo: &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}
	var ops []*zt.ZkOperation
	ops = append(ops, &zt.ZkOperation{Ty: zt.TyMintNFTAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_MintNFT{MintNFT: special}}})
	ops = append(ops, feeQueue)
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}
	receipts = mergeReceipt(receipts, r)

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
	feeInfo, err := GetFeeData(a.statedb, zt.TyWithdrawNFTAction, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}
	feeAmount := feeInfo.Fee

	amountStr := big.NewInt(0).SetUint64(payload.Amount).String()

	nftStatus, err := getNFTById(a.statedb, payload.NFTTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getNFTById=%d", payload.NFTTokenId)
	}

	contentHashPart1, contentHashPart2, _, err := zt.SplitNFTContent(nftStatus.ContentHash)
	if err != nil {
		return nil, errors.Wrapf(err, "split content hash=%s", nftStatus.ContentHash)
	}

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

	feeReceipt, feeQueue, err := a.MakeFeeLog(feeAmount, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	//add  queue
	special := &zt.ZkWithdrawNFTWitnessInfo{
		FromAcctID:      payload.FromAccountId,
		NFTTokenID:      payload.NFTTokenId,
		WithdrawAmount:  payload.Amount,
		CreatorAcctID:   nftStatus.CreatorId,
		ErcProtocol:     nftStatus.ErcProtocol,
		ContentHash:     []string{contentHashPart1.String(), contentHashPart2.String()},
		CreatorSerialID: nftStatus.CreatorSerialId,
		InitMintAmount:  nftStatus.MintAmount,
		Signature:       payload.Signature,
		Fee:             feeInfo,
		BlockInfo:       &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}
	var ops []*zt.ZkOperation
	ops = append(ops, &zt.ZkOperation{Ty: zt.TyWithdrawNFTAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_WithdrawNFT{WithdrawNFT: special}}})
	ops = append(ops, feeQueue)
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}
	receipts = mergeReceipt(receipts, r)

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
	feeInfo, err := GetFeeData(a.statedb, zt.TyTransferNFTAction, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}
	feeAmount := feeInfo.Fee

	amountStr := big.NewInt(0).SetUint64(payload.Amount).String()

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
		From: receiptFrom,
		To:   receiptTo,
	}
	l2TransferNFTlog := &types.ReceiptLog{
		Ty:  zt.TyTransferNFTLog,
		Log: types.Encode(transferLog),
	}
	logs = append(logs, l2TransferNFTlog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, feeQueue, err := a.MakeFeeLog(feeAmount, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	//add  queue
	special := &zt.ZkTransferNFTWitnessInfo{
		FromAccountID: payload.FromAccountId,
		RecipientID:   payload.RecipientId,
		NFTTokenID:    payload.NFTTokenId,
		Amount:        payload.Amount,
		Fee:           feeInfo,
		BlockInfo:     &zt.OpBlockInfo{Height: a.height, TxIndex: int32(a.index)},
	}
	var ops []*zt.ZkOperation
	ops = append(ops, &zt.ZkOperation{Ty: zt.TyTransferNFTAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TransferNFT{TransferNFT: special}}})
	ops = append(ops, feeQueue)
	r, _, err := setL2QueueData(a.statedb, ops)
	if err != nil {
		return nil, err
	}
	receipts = mergeReceipt(receipts, r)

	return receipts, nil
}

func makeSetExodusModeReceipt(prev, current int64) *types.Receipt {
	key := getExodusModeKey()
	log := &zt.ReceiptExodusMode{
		Prev:    prev,
		Current: current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&types.Int64{Data: current})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  zt.TyLogSetExodusMode,
				Log: types.Encode(log),
			},
		},
	}
}

//在设置了invalidTx后，平行链从0开始同步到无效交易则设置系统为exodus mode，此模式意味着此链即将停用，资产需要退出到ETH
//目前此模式限制交易比较多，此模式开启后的后续所有跟L2 资产有关的交易(contract2tree例外)都视为无效交易，其中deposit,withdraw,proxyExit确实应该视为无效
//但是transfer，transfer2new, tree2contract实际上应该是允许的，因为只是在L2内部流转, 禁掉的影响是跟这几个操作相关连的交易会失败
//改进的一个方案是允许这几个操作,但是需要重新设一个截止标志，禁止这几个操作，也就是平行链同步完成后，由管理员设置，然后就只允许contract2tree流进资产
func isExodusMode(statedb dbm.KV) error {
	mode, err := getExodusMode(statedb)
	if err != nil {
		return err
	}
	if mode > zt.NormalMode {
		return errors.Wrapf(types.ErrNotAllow, "isExodusMode=%d", mode)
	}
	return nil
}

func getExodusMode(db dbm.KV) (int64, error) {
	data, err := db.Get(getExodusModeKey())
	if isNotFound(err) {
		//非exodus mode
		return zt.InitMode, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "db")
	}
	var k types.Int64
	err = types.Decode(data, &k)
	if err != nil {
		return 0, errors.Wrapf(err, "decode")
	}
	return k.Data, nil
}

//设置逃生舱模式,为保证顺序，管理员只允许在无效交易生效后，也就是逃生舱准备模式后设置清算模式
func (a *Action) setExodusMode(payload *zt.ZkExodusMode) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if payload.GetMode() < zt.NormalMode || payload.GetMode() > zt.ExodusFinalMode {
		return nil, errors.Wrapf(types.ErrInvalidParam, "mode=%d not between[%d:%d]", payload.GetMode(), zt.PauseMode, zt.ExodusFinalMode)
	}

	//当前mode
	mode, err := getExodusMode(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "getExoduxMode")
	}
	if payload.Mode == uint32(mode) {
		return nil, errors.Wrapf(types.ErrNotAllow, "current mode=%d,set mode=%d", mode, payload.Mode)
	}
	switch payload.Mode {
	case zt.NormalMode:
		//只有管理员可以设置normal mode,防止有可能verifier私钥被盗的场景
		if !isSuperManager(cfg, a.fromaddr) {
			return nil, errors.Wrapf(types.ErrNotAllow, "not manager")
		}
		if mode != zt.PauseMode {
			return nil, errors.Wrapf(types.ErrNotAllow, "current mode=%d", mode)
		}
		return makeSetExodusModeReceipt(mode, int64(payload.GetMode())), nil
	case zt.PauseMode:
		//允许设置暂停模式，在verifier校验L2上deposit和L1的queue不一致时，立即设置pause，校验L2和L1的deposit queue一致后再恢复
		//因为proof提交到L1可能会比较晚，防止proof被L1校验出错之前有大量deposit存入需要回滚
		//只有管理员可以设置
		if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, a.fromaddr) {
			return nil, errors.Wrapf(types.ErrNotAllow, "not manager or verifier")
		}
		if mode > zt.PauseMode {
			return nil, errors.Wrapf(types.ErrNotAllow, "current mode=%d", mode)
		}
		return makeSetExodusModeReceipt(mode, int64(payload.GetMode())), nil
	case zt.ExodusMode:
		//只有管理员可以设置,Pause模式比exodus模式更灵活，唯一区别是pause mode可以恢复，这样可以腾挪资金，eoxdus设置后就不允许恢复了
		if !isSuperManager(cfg, a.fromaddr) {
			return nil, errors.Wrapf(types.ErrNotAllow, "not manager")
		}
		//exodusMode时候，db mode只能更小
		if mode >= zt.ExodusMode {
			return nil, errors.Wrapf(types.ErrNotAllow, "current mode=%d,set mode=%d", mode, payload.Mode)
		}
		return makeSetExodusModeReceipt(mode, int64(payload.GetMode())), nil
	case zt.ExodusFinalMode:
		//只有管理员可以设置
		if !isSuperManager(cfg, a.fromaddr) {
			return nil, errors.Wrapf(types.ErrNotAllow, "not manager")
		}
		//1. 模式在exodus和pause下都可以进行rollback
		if mode != zt.ExodusMode && mode != zt.PauseMode {
			return nil, errors.Wrapf(types.ErrNotAllow, "current mode=%d", mode)
		}
		receipt, err := a.procExodusRollbackMode(payload)
		if err != nil {
			return nil, err
		}
		return mergeReceipt(receipt, makeSetExodusModeReceipt(mode, int64(payload.GetMode()))), nil
	}
	return nil, nil
}

func (a *Action) procExodusRollbackMode(payload *zt.ZkExodusMode) (*types.Receipt, error) {
	rollbackParam := payload.GetRollback()
	if rollbackParam.LastSuccessProofId < 1 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "rollbackMode lastSuccProofId=%d", rollbackParam.LastSuccessProofId)
	}
	proofQueueData, err := GetProofId2QueueId(a.statedb, rollbackParam.LastSuccessProofId)
	if err != nil {
		return nil, errors.Wrapf(err, "GetProofId2QueueId=%d", rollbackParam.LastSuccessProofId)
	}
	//1. 确保SystemTree2ContractAcctId的所有tokenId balance都为0才能设置回滚，意味着从contract已经全部token转回到tree了
	err = checkAccountBalanceNil(a.statedb, zt.SystemTree2ContractAcctId)
	if err != nil {
		return nil, errors.Wrapf(err, "checkAccBalanceNil id=%d", zt.SystemTree2ContractAcctId)
	}

	//2. rollback queue里面的deposit,withdraw等和L1有关系的操作
	//LastQueueId+1开始查找回滚
	ops, err := getRollbackOps(a.statedb, proofQueueData.GetLastQueueId()+1)
	if err != nil {
		return nil, errors.Wrapf(err, "getRollbackOps")
	}
	depositAcctIds, withdrawAcctIds, depositAccountMap, withdrawAccountMap, err := parseRollbackOps(ops)
	if err != nil {
		return nil, errors.Wrapf(err, "parseRollbackOps")
	}

	//deposit acctId 回滚数据，需要扣除
	depositRollbackAcctData, err := getDepositRollbackData(a.statedb, depositAcctIds, depositAccountMap, rollbackParam.KnownBalanceGap)
	if err != nil {
		return nil, errors.Wrapf(err, "getDepositRollbackData")
	}
	receipt := &types.Receipt{Ty: types.ExecOk}
	//acct+systemFeeAcct的rollback数据需要扣除余额
	for _, v := range depositRollbackAcctData {
		r, err := accountRollbackProc(a.statedb, v.AccountId, v.TokenId, v.NeedRollback, zt.Sub, zt.TyDepositRollbackLog)
		if err != nil {
			return nil, errors.Wrapf(err, "acctDepositRollbackProc acctId=%d,tokenId=%d,rollbackAmount=%s", v.AccountId, v.TokenId, v.NeedRollback)
		}
		mergeReceipt(receipt, r)
	}

	//withdraw acctId 需要增加余额
	for _, acctId := range withdrawAcctIds {
		for _, token := range withdrawAccountMap[acctId].Tokens {
			r, err := accountRollbackProc(a.statedb, acctId, token.TokenId, token.Balance, zt.Add, zt.TyWithdrawRollbackLog)
			if err != nil {
				return nil, errors.Wrapf(err, "acctWithdrawRollbackBalance acctId=%d,tokenId=%d,rollbackAmount=%s", acctId, token.TokenId, token.Balance)
			}
			mergeReceipt(receipt, r)
		}
	}
	return receipt, nil

}

//获取deposit操作需要回滚的数据，如果余额不够，尝试从systemFeeId扣除，如果fee也不够，记录gap从L1补充
func getDepositRollbackData(db dbm.KV, depositAcctIds []uint64, depositAccountMap map[uint64]*zt.HistoryLeaf, knownBalanceGap uint32) ([]*zt.ZkAcctRollbackInfo, error) {
	var depositRollbackAcctData []*zt.ZkAcctRollbackInfo
	tokensGap := make(map[uint64]string)
	var tokenIds []uint64 //记录顺序
	var totalAccountGapStr string
	//统计所有存款账户rollback相对本身账户余额的gap信息
	//1. 如果没有gap则从账户本身余额扣除
	for _, acctId := range depositAcctIds {
		for _, token := range depositAccountMap[acctId].Tokens {
			info, err := getAcctDepositRollbackInfo(db, acctId, token.TokenId, token.Balance)
			if err != nil {
				return nil, errors.Wrapf(err, "getAcctDepositRollbackInfo")
			}
			depositRollbackAcctData = append(depositRollbackAcctData, info)
			if len(info.Gap) > 0 {
				if _, ok := tokensGap[token.TokenId]; !ok {
					tokenIds = append(tokenIds, token.TokenId)
				}
				updateTokenGap(tokensGap, token.TokenId, info.Gap)
				totalAccountGapStr += fmt.Sprintf("acctId=%d,tokenId=%d,gap=%s,", info.AccountId, info.TokenId, info.Gap)
			}
		}
	}
	//2. 如果有gap则尝试从system fee账户扣除，如果fee账户仍不够，则提供累计gap信息到receipt，管理员存款到L1弥补后保证提款成功
	if len(tokensGap) > 0 {
		totalAccountGapStr = strings.TrimSuffix(totalAccountGapStr, ",")
		zklog.Info("getDepositRollbackData", "exist acct gap", totalAccountGapStr)
		sysFeeTokens, err := getAcctTokens(db, zt.SystemFeeAccountId)
		if err != nil {
			return nil, errors.Wrapf(err, "getSysFeeAcctTokens acctId=%d", zt.SystemFeeAccountId)
		}
		sysRollbackData, systemGapStr := checkSystemFeeAcctGap(sysFeeTokens, tokensGap)
		//说明系统rollback后也不够，需要L1层补充
		if len(systemGapStr) > 0 && knownBalanceGap != zt.ModeValYes {
			gapStr := fmt.Sprintf("deposit rollback balance gap,sum: %s(%s)", systemGapStr, totalAccountGapStr)
			zklog.Error("getDepositRollbackData", "system gap", gapStr)
			return nil, errors.Wrapf(types.ErrNotAllow, gapStr)
		}
		//收集SystemFeeAccountId回滚的数据
		depositRollbackAcctData = append(depositRollbackAcctData, sysRollbackData...)
	}
	return depositRollbackAcctData, nil
}

func getRollbackOps(db dbm.KV, reqStartQueueId int64) ([]*zt.ZkOperation, error) {
	lastQueueId, err := GetL2LastQueueId(db)
	if err != nil {
		return nil, errors.Wrapf(err, "GetL2LastQueueId")
	}
	//正常来说proof的lastQueueId是从statedb里面读取出来的，不可能小于lastQueueId，只能<=
	if reqStartQueueId > lastQueueId {
		return nil, errors.Wrapf(types.ErrNotAllow, "lastQueueId=%d less than reqStartQueueId=%d", lastQueueId, reqStartQueueId)
	}
	zklog.Info("procExodusRollbackMode", "proof's reqStartQueId", reqStartQueueId, "system.lastQueueId", lastQueueId)

	var ops []*zt.ZkOperation
	for i := reqStartQueueId; i <= lastQueueId; i++ {
		op, err := GetL2QueueIdOp(db, i)
		if err != nil {
			return nil, errors.Wrapf(err, "GetL2QueueIdOp queueId=%d", i)
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func getAcctTokens(db dbm.KV, accountId uint64) (map[uint64]string, error) {
	leaf, err := GetLeafByAccountId(db, accountId)
	if err != nil {
		return nil, errors.Wrapf(err, "get leaf id=%d", accountId)
	}

	sysTokens := make(map[uint64]string)
	if leaf != nil && len(leaf.TokenIds) > 0 {
		for _, tokenId := range leaf.TokenIds {
			balance, err := GetTokenByAccountIdAndTokenId(db, accountId, tokenId)
			if err != nil {
				return nil, errors.Wrapf(err, "acctId=%d,token=%d", accountId, tokenId)
			}
			if balance != nil {
				sysTokens[tokenId] = balance.Balance
			}
		}
	}
	return sysTokens, nil
}

func checkSystemFeeAcctGap(sysFeeTokens, tokensGap map[uint64]string) ([]*zt.ZkAcctRollbackInfo, string) {
	var systemGapResp string
	var systemGapData []*zt.ZkAcctRollbackInfo
	var tokenIds []uint64 //记录顺序
	for id, _ := range tokensGap {
		tokenIds = append(tokenIds, id)
	}
	sort.Slice(tokenIds, func(i, j int) bool { return tokenIds[i] < tokenIds[j] })
	for _, tokenId := range tokenIds {
		sysBalance, ok := sysFeeTokens[tokenId]
		gap := tokensGap[tokenId]
		if !ok {
			//feeAcct中不存在此token
			systemGapResp += fmt.Sprintf("tokenId=%d,gap=%s,", tokenId, gap)
		} else {

			sysV, _ := new(big.Int).SetString(sysBalance, 10)
			gapV, _ := new(big.Int).SetString(gap, 10)
			data := &zt.ZkAcctRollbackInfo{
				AccountId:    zt.SystemFeeAccountId,
				TokenId:      tokenId,
				Balance:      sysBalance,
				NeedRollback: gap,
			}
			if gapV.Cmp(sysV) > 0 {
				//余额不够，先扣除所有
				data.NeedRollback = data.Balance
				data.Gap = new(big.Int).Sub(gapV, sysV).String()
				systemGapData = append(systemGapData, data)
				//记录系统中扣除了gap后还差的缺口gap，如果不记录说明系统余额足够，不需要L1补充
				systemGapResp += fmt.Sprintf("tokenId=%d,gap=%s,", tokenId, data.Gap)
			} else {
				systemGapData = append(systemGapData, data)
			}
		}
	}

	return systemGapData, strings.TrimSuffix(systemGapResp, ",")
}

func checkAccountBalanceNil(db dbm.KV, accountId uint64) error {
	leaf, err := GetLeafByAccountId(db, accountId)
	if err != nil {
		return errors.Wrapf(err, "get leaf id=%d", accountId)
	}
	if leaf != nil && len(leaf.TokenIds) > 0 {
		for _, tokenId := range leaf.TokenIds {
			balance, err := GetTokenByAccountIdAndTokenId(db, accountId, tokenId)
			if err != nil {
				return errors.Wrapf(err, "acctId=%d,tokenId=%d", accountId, tokenId)
			}
			if balance != nil {
				v, ok := new(big.Int).SetString(balance.Balance, 10)
				if !ok {
					return errors.Wrapf(types.ErrInvalidParam, "acctId=%d,tokenId=%d,balance=%s", accountId, tokenId, balance.Balance)
				}
				if v.Cmp(big.NewInt(0)) != 0 {
					return errors.Wrapf(types.ErrNotAllow, "acctId=%d,tokenId=%d,balance=%s not 0", accountId, tokenId, balance.Balance)
				}
			}
		}
	}
	return nil
}

func getAcctDepositRollbackInfo(db dbm.KV, accountId, tokenId uint64, rollbackAmount string) (*zt.ZkAcctRollbackInfo, error) {
	token, err := GetTokenByAccountIdAndTokenId(db, accountId, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "tokenId=%d,acctId=%d", tokenId, accountId)
	}
	if token == nil {
		return nil, errors.Wrapf(types.ErrNotFound, "tokenId=%d,acctId=%d", tokenId, accountId)
	}
	acctRollbackInfo := &zt.ZkAcctRollbackInfo{
		AccountId:    accountId,
		TokenId:      tokenId,
		Balance:      token.Balance,
		NeedRollback: rollbackAmount, //记录需要回滚的value
	}
	balance, ok := new(big.Int).SetString(token.Balance, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "token balance=%s", token.GetBalance())
	}
	need, ok := new(big.Int).SetString(rollbackAmount, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "rollbackAmount=%s", rollbackAmount)
	}

	//token balance小于需要扣除的，从systemFeeAcctId里面扣
	if need.Cmp(balance) > 0 {
		//余额全部回滚掉
		acctRollbackInfo.NeedRollback = acctRollbackInfo.Balance
		acctRollbackInfo.Gap = new(big.Int).Sub(need, balance).String()
	}
	return acctRollbackInfo, nil

}

func accountRollbackProc(db dbm.KV, accountId, tokenId uint64, amount string, option, logTy int32) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	balancekv, balanceHistory, err := updateTokenBalance(accountId, tokenId, amount, option, db)
	if err != nil {
		return nil, errors.Wrapf(err, "updateTokenBalance")
	}
	kvs = append(kvs, balancekv)

	log := &zt.AccountTokenBalanceReceipt{
		TokenId:       tokenId,
		AccountId:     accountId,
		BalanceBefore: balanceHistory.before,
		BalanceAfter:  balanceHistory.after,
	}

	receiptLog := &types.ReceiptLog{Ty: logTy, Log: types.Encode(log)}
	logs = append(logs, receiptLog)
	return &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}, nil
}

func MakeSetTokenSymbolReceipt(id string, oldVal, newVal *zt.ZkTokenSymbol) *types.Receipt {
	var kvs []*types.KeyValue
	keyId := GetTokenSymbolKey(id)
	kvs = append(kvs, &types.KeyValue{Key: keyId, Value: types.Encode(newVal)})

	keySym := GetTokenSymbolIdKey(newVal.Symbol)
	kvs = append(kvs, &types.KeyValue{Key: keySym, Value: types.Encode(newVal)})

	//如果是更新了symbol，需要把旧的symbol对应的id更新为""，旧的symbol可以再次被别的tokenId使用，不然会混乱
	if oldVal != nil && oldVal.Symbol != newVal.Symbol {
		oldSymVal := *oldVal
		oldSymVal.Id = "" //置空
		kvs = append(kvs, &types.KeyValue{Key: GetTokenSymbolIdKey(oldVal.Symbol), Value: types.Encode(&oldSymVal)})
	}
	log := &zt.ReceiptSetTokenSymbol{
		Pre: oldVal,
		Cur: newVal,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: kvs,
		Logs: []*types.ReceiptLog{
			{Ty: zt.TyLogSetTokenSymbol, Log: types.Encode(log)},
		},
	}
}

//tokenId可以对应多个symbol，但一个symbol只能对应一个Id,比如Id=1,symbol=USTC,后改成USTD, USTC仍然会对应Id=1, 新的Id不能使用已存在的名字，防止重复混乱
func (a *Action) setTokenSymbol(payload *zt.ZkTokenSymbol) (*types.Receipt, error) {
	cfg := a.api.GetConfig()

	//只有管理员可以设置
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
	}
	if len(payload.Id) <= 0 || len(payload.Symbol) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "id=%s or symbol=%s nil", payload.Id, payload.Symbol)
	}
	idInt, ok := new(big.Int).SetString(payload.Id, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "id=%s", payload.Id)
	}

	if payload.Decimal > zt.MaxDecimalAllow || payload.Decimal < zt.MinDecimalAllow {
		return nil, errors.Wrapf(types.ErrInvalidParam, "Decimal=%d,max=%d,mini=%d", payload.Decimal, zt.MaxDecimalAllow, zt.MinDecimalAllow)
	}

	//首先检查symbol是否存在，symbol存在不允许修改
	token, err := GetTokenBySymbol(a.statedb, payload.Symbol)
	if err != nil && !isNotFound(errors.Cause(err)) {
		return nil, err
	}
	//id更换symbol后，之前的symbol对应的id会设为空，若symbol的id非空则不允许更换id，但是可以更换小数位
	if token != nil && token.Id != "" && token.Id != payload.Id {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenSymbol=%s existed id=%s", payload.Symbol, token.Id)
	}

	//id初始设置或者重设symbol
	lastSym, err := GetTokenByTokenId(a.statedb, payload.Id)
	if err != nil && !isNotFound(errors.Cause(err)) {
		return nil, err
	}
	//id已经存在，已有id修改symbol或decimal需要SystemTree2ContractAcctId的当前token balance为0的时候,说明没有转出到chain33资产或者已经全部转回来了
	if lastSym != nil && (lastSym.Symbol != payload.Symbol || lastSym.Decimal != payload.Decimal) {
		balance, err := GetTokenByAccountIdAndTokenIdInDB(a.statedb, zt.SystemTree2ContractAcctId, idInt.Uint64())
		if err != nil {
			return nil, errors.Wrapf(err, "get systemContractAcctId=%d token=%s", zt.SystemTree2ContractAcctId, token.Id)
		}
		if balance != nil {
			balanceInt, ok := new(big.Int).SetString(balance.Balance, 10)
			if !ok {
				return nil, errors.Wrapf(types.ErrInvalidParam, "systemContractAcctId=%d token=%s,balance=%s", zt.SystemTree2ContractAcctId, token.Id, balance.Balance)
			}
			//只有balance为0时候允许，说明已经从contract完全提取资产到tree
			if balanceInt.Cmp(big.NewInt(0)) != 0 {
				return nil, errors.Wrapf(types.ErrNotAllow, "systemContractAcctId=%d token=%s,balance=%s should be 0", zt.SystemTree2ContractAcctId, token.Id, balance.Balance)
			}
		}
		//balance=nil or =0  ok
	}
	return MakeSetTokenSymbolReceipt(payload.Id, lastSym, payload), nil
}

func GetTokenByTokenId(db dbm.KV, tokenId string) (*zt.ZkTokenSymbol, error) {
	key := GetTokenSymbolKey(tokenId)
	r, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "getSymbolByTokenId.getDb")
	}
	var symbol zt.ZkTokenSymbol
	err = types.Decode(r, &symbol)
	if err != nil {
		return nil, errors.Wrapf(err, "getTokenIdSymbol.decode")
	}
	return &symbol, nil
}

func GetTokenBySymbol(db dbm.KV, symbol string) (*zt.ZkTokenSymbol, error) {
	if len(symbol) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "symbol nil=%s", symbol)
	}
	key := GetTokenSymbolIdKey(symbol)
	r, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "getDb")
	}
	var token zt.ZkTokenSymbol
	err = types.Decode(r, &token)
	if err != nil {
		return nil, errors.Wrapf(err, "decode")
	}
	return &token, nil
}

func (a *Action) AssetTransfer(transfer *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	from := tx.From()

	cfg := a.api.GetConfig()
	acc, err := account.NewAccountDB(cfg, zt.Zksync, transfer.Cointoken, a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "newAccountDb")
	}
	//to 是 execs 合约地址
	if dapp.IsDriverAddress(tx.GetRealToAddr(), a.height) {
		return acc.TransferToExec(from, tx.GetRealToAddr(), transfer.Amount)
	}
	return acc.Transfer(from, tx.GetRealToAddr(), transfer.Amount)
}

func (a *Action) AssetWithdraw(withdraw *types.AssetsWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	acc, err := account.NewAccountDB(cfg, zt.Zksync, withdraw.Cointoken, a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "newAccountDb")
	}
	if dapp.IsDriverAddress(tx.GetRealToAddr(), a.height) || dapp.ExecAddress(withdraw.ExecName) == tx.GetRealToAddr() {
		return acc.TransferWithdraw(tx.From(), tx.GetRealToAddr(), withdraw.Amount)
	}
	return nil, types.ErrToAddrNotSameToExecAddr
}

func (a *Action) AssetTransferToExec(transfer *types.AssetsTransferToExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	from := tx.From()

	cfg := a.api.GetConfig()
	acc, err := account.NewAccountDB(cfg, zt.Zksync, transfer.Cointoken, a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "newAccountDb")
	}
	//to 是 execs 合约地址
	if dapp.IsDriverAddress(tx.GetRealToAddr(), a.height) || dapp.ExecAddress(transfer.ExecName) == tx.GetRealToAddr() {
		return acc.TransferToExec(from, tx.GetRealToAddr(), transfer.Amount)
	}
	return nil, types.ErrToAddrNotSameToExecAddr
}
