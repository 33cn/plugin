package executor

import (
	"encoding/hex"
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
	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, zt.ZkParaChainInnerTitleId, a.fromaddr) {
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

	special := &zt.ZkDepositWitnessInfo{
		//accountId nil
		TokenID:       payload.TokenId,
		Amount:        payload.Amount,
		EthAddress:    payload.EthAddress,
		Layer2Addr:    payload.Chain33Addr,
		Signature:     payload.Signature,
		EthPriorityID: payload.EthPriorityQueueId,
	}

	//leaf不存在就添加
	if leaf == nil {
		zklog.Info("zksync deposit add new leaf")

		var accountID uint64
		lastAccountID, _ := getLatestAccountID(a.statedb)
		if zt.SystemDefaultAcctId == lastAccountID {
			ethFeeAddr, chain33FeeAddr := getCfgFeeAddr(cfg)
			kvs4InitAccount, err := NewInitAccount(ethFeeAddr, chain33FeeAddr)
			if nil != err {
				return nil, err
			}
			kvs = append(kvs, kvs4InitAccount...)

			//对于首次进行存款的用户，其账户ID从SystemNFTAccountId后开始进行连续分配
			accountID = zt.SystemTree2ContractAcctId + 1
		} else {
			accountID = uint64(lastAccountID) + 1
		}

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

		kvs = append(kvs, updateKVs...)
		l2Log.Ty = int32(zt.TyDepositLog)
		logs = append(logs, l2Log)
	}
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	//add priority part
	r := makeSetEthPriorityIdReceipt(0, lastPriorityId.Int64(), payload.EthPriorityQueueId)
	return mergeReceipt(receipts, r), nil
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
		Signature: payload.Signature,
		Fee: &zt.ZkFee{
			Fee:     fee,
			TokenID: payload.TokenId,
		},
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

	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId)
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
	//保证精度是小数后8位，eth转换去掉10位
	if a.api.GetConfig().GetCoinPrecision() != types.DefaultCoinPrecision {
		return nil, errors.Wrapf(types.ErrInvalidParam, "coin precision is not defual=%d", types.DefaultCoinPrecision)
	}

	////因为chain33合约精度为1e8,而外部输入精度则为1e18, 单位为wei,需要统一转化为1e8
	//amount_len := len(payload.Amount)
	//if amount_len < 11 {
	//	return nil, errors.New("Too Little value to do operation TreeToContract")
	//}
	//payload.Amount = payload.Amount[:len(payload.Amount)-10] + zt.TenZeroStr

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

	amountPlusFee, fee, err := GetAmountWithFee(a.statedb, zt.TyContractToTreeAction, payload.Amount, tokenId)
	if nil != err {
		return nil, err
	}

	//根据精度，转化到tree侧的amount
	sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")
	amountTree, _, _, err := GetTreeSideAmount(payload.Amount, amountPlusFee, fee, sysDecimal, int(token.Decimal))
	if err != nil {
		return nil, errors.Wrap(err, "getTreeSideAmount")
	}

	payload4transfer := &zt.ZkTransfer{
		TokenId:       tokenId,
		Amount:        amountTree,
		FromAccountId: zt.SystemTree2ContractAcctId,
		ToAccountId:   payload.ToAccountId,
	}
	receipts, err := a.l2TransferProc(payload4transfer, zt.TyContractToTreeAction)
	if nil != err {
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

	receiptsTransfer, err := a.transferToNewProcess(zt.SystemTree2ContractAcctId, payload.ToLayer2Addr, payload.ToEthAddr, amountPlusFeeTree, amountTree, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "transfer2NewProc")
	}

	feeReceipt, err := a.MakeFeeLog(feeTree, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}

	receipts := mergeReceipt(receiptsTransfer, contractReceipt)
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

//L2 ---->  合约账户(树)
//操作１. FromAccountId -----> SystemTree2ContractAcctId，执行ZkTransfer
//操作2. UpdateContractAccount，在合约内部的铸币操作
func (a *Action) TreeToContract(payload *zt.ZkTreeToContract) (*types.Receipt, error) {
	err := checkParam(payload.Amount)
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
	receiptTranfer, err := a.ZkTransfer(para)
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

func (a *Action) l2TransferProc(payload *zt.ZkTransfer, actionTy int32) (*types.Receipt, error) {
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
	logs = append(logs, l2Transferlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	//2.操作交易费账户
	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

func (a *Action) ZkTransfer(payload *zt.ZkTransfer) (*types.Receipt, error) {
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

	return a.l2TransferProc(payload, zt.TyTransferAction)
}

func (a *Action) transferToNewProcess(accountIdFrom uint64, toChain33Address, toEthAddress, totalAmount, amount string, tokenID uint64) (*types.Receipt, error) {

	fromLeaf, err := GetLeafByAccountId(a.statedb, accountIdFrom)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	return a.transferToNewInnerProcess(fromLeaf, toChain33Address, toEthAddress, totalAmount, amount, tokenID)
}

func (a *Action) transferToNewInnerProcess(fromLeaf *zt.Leaf, toChain33Address, toEthAddress, totalAmount, amount string, tokenID uint64) (*types.Receipt, error) {
	var kvs []*types.KeyValue
	var logs []*types.ReceiptLog

	toLeaf, err := GetLeafByChain33AndEthAddress(a.statedb, toChain33Address, toEthAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}
	if toLeaf != nil {
		return nil, errors.New("to account already exist")
	}

	//1.操作from 账户
	fromKVs, _, receiptFrom, err := applyL2AccountUpdate(fromLeaf.GetAccountId(), tokenID, totalAmount, zt.Sub, a.statedb, fromLeaf, false)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountUpdate")
	}
	kvs = append(kvs, fromKVs...)

	//1.操作to 账户
	lastAccountID, err := getLatestAccountID(a.statedb)
	if lastAccountID == zt.SystemDefaultAcctId {
		return nil, errors.Wrapf(err, "getLatestAccountID")
	}
	accountIDNew := uint64(lastAccountID) + 1

	toKVs, _, receiptTo, err := applyL2AccountCreate(accountIDNew, tokenID, amount, toEthAddress, toChain33Address, a.statedb, false)
	if nil != err {
		return nil, errors.Wrapf(err, "applyL2AccountCreate")
	}
	kvs = append(kvs, toKVs...)

	transferToNewLog := &zt.TransferReceipt4L2{
		From: receiptFrom,
		To:   receiptTo,
	}
	l2Transferlog := &types.ReceiptLog{
		Ty:  zt.TyTransferToNewLog,
		Log: types.Encode(transferToNewLog),
	}
	logs = append(logs, l2Transferlog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	return receipts, nil
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
	newAddr, ok := zt.HexAddr2Decimal(payload.ToChain33Address)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer ToChain33Address=%s", payload.ToChain33Address)
	}
	payload.ToChain33Address = newAddr

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

	receipts, err := a.transferToNewInnerProcess(fromLeaf, payload.GetToChain33Address(), payload.GetToEthAddress(), amountPlusFee, payload.Amount, payload.GetTokenId())
	if nil != err {
		return nil, err
	}

	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
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
	l2LogFrom.Ty = zt.TyProxyExitLog
	logs = append(logs, l2Logproxy)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(fee, payload.TokenId)
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
		PubKey:    payload.PubKey,
		PubKeyTy:  payload.PubKeyTy,
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
	key := CalcLatestAccountIDKey()
	v, err := db.Get(key)

	if err != nil {
		return zt.SystemDefaultAcctId, err
	}
	var id types.Int64
	err = types.Decode(v, &id)
	if err != nil {
		zklog.Error("getLastEthPriorityQueueID.decode", "err", err)
		return zt.SystemDefaultAcctId, err
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

func checkPackValue(amount string, manMaxBitWidth int64) error {
	//exp部分默认最大是31，不需要检查
	man, _, err := zt.ZkTransferManExpPart(amount)
	if err != nil {
		return errors.Wrapf(err, "ZkTransferManExpPart,amount=%s", amount)
	}
	manV, ok := new(big.Int).SetString(man, 10)
	if !ok {
		return errors.Wrapf(types.ErrInvalidParam, "ZkTransferManExpPart,man=%s,amount=%s", manV, amount)
	}

	maxManV := new(big.Int).Exp(big.NewInt(2), big.NewInt(manMaxBitWidth), nil)
	//manv <= maxMan
	if maxManV.Cmp(manV) < 0 {
		return errors.Wrapf(types.ErrNotAllow, "fee amount's manV=%s big than 2^%d", man, manMaxBitWidth)
	}
	return nil
}

func (a *Action) MakeFeeLog(amount string, tokenId uint64) (*types.Receipt, error) {
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

	toKVs, l2Log, _, err := applyL2AccountUpdate(leaf.GetAccountId(), tokenId, amount, zt.Add, a.statedb, leaf, true)
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

	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, zt.ZkParaChainInnerTitleId, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
	}
	//跟其他action手续费以二层各token精度一致不同，contract2tree action的手续费统一以合约的精度处理，这样和合约侧amount精度一致，简化了合约侧的精度处理
	if payload.ActionTy == zt.TyContractToTreeAction {
		token, err := GetTokenByTokenId(a.statedb, big.NewInt(int64(payload.TokenId)).String())
		if err != nil {
			return nil, errors.Wrapf(err, "getTokenId=%d", payload.TokenId)
		}
		sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")
		//比如token精度为6，sysDecimal=8, token的fee在sysDecimal下需要补2个0，也就是后缀需要至少有两个0，不然会丢失精度，在token精度大于sys精度时候没这问题
		if int(token.Decimal) < sysDecimal && !strings.HasSuffix(payload.Amount, strings.Repeat("0", sysDecimal-int(token.Decimal))) {
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

	feeReceipt, err := a.MakeFeeLog(feeAmount, feeTokenId)
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

	//contentHashPart1, contentHashPart2, _, err := zt.SplitNFTContent(nftStatus.ContentHash)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "split content hash=%s", nftStatus.ContentHash)
	//}

	//specialInfo := &zt.ZkWithdrawNFTWitnessInfo{
	//	FromAcctId:      payload.FromAccountId,
	//	NFTTokenID:      payload.NFTTokenId,
	//	WithdrawAmount:  new(big.Int).SetUint64(payload.Amount).String(),
	//	CreatorAcctId:   nftStatus.CreatorId,
	//	ErcProtocol:     nftStatus.ErcProtocol,
	//	ContentHash:     []string{contentHashPart1.String(), contentHashPart2.String()},
	//	CreatorSerialId: nftStatus.CreatorSerialId,
	//	InitMintAmount:  new(big.Int).SetUint64(nftStatus.MintAmount).String(),
	//	Signature:       payload.Signature,
	//	Fee:             feeInfo,
	//}

	//operationInfo := &zt.OperationInfo{
	//	BlockHeight: uint64(a.height),
	//	TxIndex:     uint32(a.index),
	//	TxType:      zt.TyWithdrawNFTAction,
	//	SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_WithdrawNFT{WithdrawNFT: specialInfo}},
	//}

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

	feeReceipt, err := a.MakeFeeLog(feeAmount, feeTokenId)
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

	feeReceipt, err := a.MakeFeeLog(feeAmount, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

//从此invalidTx之后，系统进入退出状态(eth无法接收新的proof场景)，平行链需要从0开始重新同步交易，相当于回滚，无效交易之后的交易都失败(除contract2tree)
//退出状态不允许deposit,withdraw等的其他交易，只允许contract2Tree的退到二层的交易，
//退到二层后，统计各账户id的余额，根据最后的treeRootHash提交退出证明(不能走withdraw等流程退出）
func isInvalidTx(cfg *types.Chain33Config, txHash []byte) bool {
	invalidTx, _ := getCfgInvalidTx(cfg)
	return invalidTx == hex.EncodeToString(txHash)
}

func getCfgInvalidTx(cfg *types.Chain33Config) (string, string) {
	confManager := types.ConfSub(cfg, zt.Zksync)
	invalidTx := confManager.GStr(zt.ZkCfgInvalidTx)
	invalidProof := confManager.GStr(zt.ZkCfgInvalidProof)
	if (len(invalidTx) <= 0 && len(invalidProof) > 0) || (len(invalidTx) > 0 && len(invalidProof) <= 0) {
		panic(fmt.Sprintf("both invalidTx=%s and invalidProof=%s should filled", invalidTx, invalidProof))
	}
	if strings.HasPrefix(invalidTx, "0x") || strings.HasPrefix(invalidTx, "0X") {
		invalidTx = invalidTx[2:]
	}
	if strings.HasPrefix(invalidProof, "0x") || strings.HasPrefix(invalidProof, "0X") {
		invalidProof = invalidProof[2:]
	}

	return invalidTx, invalidProof
}

//在此proofRootHash(包括此hash)后的所有历史roothash都会失效，直到替换此hash之后的新的proofhash收到
func isInvalidProof(cfg *types.Chain33Config, rootHash string) bool {
	_, invalidProof := getCfgInvalidTx(cfg)
	return invalidProof == rootHash
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
	if mode > 0 {
		return errors.Wrapf(types.ErrNotAllow, "isExodusMode=%d", mode)
	}
	return nil
}

//exodus 清算模式,管理员设置，禁止除contract2tree外一切L2交易，方便尽快收敛treeRoot
func isExodusClearMode(statedb dbm.KV) error {
	mode, err := getExodusMode(statedb)
	if err != nil {
		return err
	}

	if mode >= zt.ExodusClearMode {
		return errors.Wrapf(types.ErrNotAllow, "current exodusClearStage")
	}
	return nil
}

func getExodusMode(db dbm.KV) (int64, error) {
	data, err := db.Get(getExodusModeKey())
	if isNotFound(err) {
		//非exodus mode
		return 0, nil
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

	//只有管理员可以设置
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "not manager")
	}

	if payload.GetMode() <= zt.ExodusPrepareMode {
		return nil, errors.Wrapf(types.ErrInvalidParam, "mode=%d should big than %d", payload.GetMode(), zt.ExodusPrepareMode)
	}

	//当前mode应该是prepareMode
	mode, err := getExodusMode(a.statedb)
	if err != nil {
		return nil, err
	}
	if mode != zt.ExodusPrepareMode {
		return nil, errors.Wrapf(types.ErrNotAllow, "current mode=%d,not prepareMode=%d", mode, zt.ExodusPrepareMode)
	}

	return makeSetExodusModeReceipt(mode, int64(payload.GetMode())), nil
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

	//首先检查symbol是否存在，symbol存在不允许修改
	token, err := GetTokenBySymbol(a.statedb, payload.Symbol)
	if err != nil && !isNotFound(errors.Cause(err)) {
		return nil, err
	}
	//已有symbol未设置为invaild不允许更换id
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

func GetSymbolByTokenId(db dbm.KV, tokenId string) (string, error) {
	key := GetTokenSymbolKey(tokenId)
	r, err := db.Get(key)
	if err != nil {
		return "", errors.Wrapf(err, "getSymbolByTokenId.getDb")
	}
	var symbol zt.ZkTokenSymbol
	err = types.Decode(r, &symbol)
	if err != nil {
		return "", errors.Wrapf(err, "getTokenIdSymbol.decode")
	}
	return symbol.Symbol, nil
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

//根据系统和token精度，计算合约转化为二层tree侧的amount，合约侧amount都是系统精度
func GetTreeSideAmount(amount, totalAmount, fee string, sysDecimal, tokenDecimal int) (amount4Tree, totalAmount4Tree, feeAmount4Tree string, err error) {
	amount4Tree, err = TransferDecimalAmount(amount, sysDecimal, tokenDecimal)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "transferDecimalAmount,amount=%s,tokenDecimal=%d,sysDecimal=%d", amount, tokenDecimal, sysDecimal)
	}
	totalAmount4Tree, err = TransferDecimalAmount(totalAmount, sysDecimal, tokenDecimal)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "transferDecimalAmount,amount=%s,tokenDecimal=%d,sysDecimal=%d", totalAmount, tokenDecimal, sysDecimal)
	}
	feeAmount4Tree, err = TransferDecimalAmount(fee, sysDecimal, tokenDecimal)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "transferDecimalAmount,amount=%s,tokenDecimal=%d,sysDecimal=%d", fee, tokenDecimal, sysDecimal)
	}
	err = checkPackValue(amount4Tree, zt.PacAmountManBitWidth)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "checkPackVal amount=%s", amount4Tree)
	}
	err = checkPackValue(feeAmount4Tree, zt.PacFeeManBitWidth)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "checkPackVal fee=%s", feeAmount4Tree)
	}
	return amount4Tree, totalAmount4Tree, feeAmount4Tree, nil
}

//from向to小数对齐，如果from>to, 需要裁减掉差别部分，且差别部分需要全0，如果from<to,差别部分需要补0
func TransferDecimalAmount(amount string, fromDecimal, toDecimal int) (string, error) {
	//from=tokenDecimal大于to=sysDecimal场景，需要裁减差别部分, 比如 1e18 > 1e8,裁减1e10
	if fromDecimal > toDecimal {
		diff := fromDecimal - toDecimal
		suffix := strings.Repeat("0", diff)
		if !strings.HasSuffix(amount, suffix) {
			return "", errors.Wrapf(types.ErrInvalidParam, "amount=%s not include suffix decimal=%d", amount, diff)
		}
		return amount[:len(amount)-diff], nil
	}
	//tokenDecimal <= 合约decimal场景，需要扩展，比如1e6 < 1e8,扩展"00"
	diff := toDecimal - fromDecimal
	suffix := strings.Repeat("0", diff)
	return amount + suffix, nil
}
