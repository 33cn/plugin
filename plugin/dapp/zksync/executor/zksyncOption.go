package executor

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/33cn/chain33/account"

	"github.com/33cn/chain33/common/log/log15"

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

func (a *Action) Deposit(payload *zt.ZkDeposit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue
	var err error

	err = checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}

	if !checkIsNormalToken(payload.TokenId) {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenId=%d should less than system NFT base ID=%d", payload.TokenId, zt.SystemNFTTokenId)
	}

	//只有管理员能操作,对于chainTitleId,当前链是平行链或主链，只接受id=ZkParaChainInnerTitleId设置
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, zt.ZkParaChainInnerTitleId, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager,from=%s", a.fromaddr)
	}

	//TODO set chainID
	lastPriority, err := getLastEthPriorityQueueID(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "get eth last priority queue id")
	}
	lastPriorityId, ok := big.NewInt(0).SetString(lastPriority.GetID(), 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "getID =%s", lastPriority.GetID())
	}
	if lastPriorityId.Int64()+1 != payload.GetEthPriorityQueueId() {
		return nil, errors.Wrapf(types.ErrNotAllow, "eth last priority queue id=%d,new=%d", lastPriorityId, payload.GetEthPriorityQueueId())
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

	ethFeeAddr, chain33FeeAddr := getCfgFeeAddr(cfg)
	info, err := generateTreeUpdateInfo(a.statedb, a.localDB, ethFeeAddr, chain33FeeAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "db.generateTreeUpdateInfo")
	}

	leaf, err := GetLeafByChain33AndEthAddress(a.statedb, payload.GetChain33Addr(), payload.GetEthAddress(), info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}

	tree, err := getAccountTree(a.statedb, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}

	special := &zt.ZkDepositWitnessInfo{
		//accountId nil
		TokenId:    payload.TokenId,
		Amount:     payload.Amount,
		EthAddress: payload.EthAddress,
		Layer2Addr: payload.Chain33Addr,
		Signature:  payload.Signature,
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight:   uint64(a.height),
		TxIndex:       uint32(a.index),
		TxType:        zt.TyDepositAction,
		EthPriorityId: payload.EthPriorityQueueId,
		SpecialInfo:   &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: special}},
	}

	//leaf不存在就添加
	if leaf == nil {
		newAccountId := tree.GetTotalIndex()
		//添加之前先计算证明
		receipt, err := calProof(a.statedb, info, newAccountId, payload.TokenId)
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}

		before := getBranchByReceipt(receipt, operationInfo, payload.EthAddress, payload.Chain33Addr, nil, nil, newAccountId, payload.TokenId, "0")

		kvs, localKvs, err = AddNewLeaf(a.statedb, a.localDB, info, payload.GetEthAddress(), payload.GetTokenId(), payload.GetAmount(), payload.GetChain33Addr())
		if err != nil {
			return nil, errors.Wrapf(err, "db.AddNewLeaf")
		}
		receipt, err = calProof(a.statedb, info, newAccountId, payload.TokenId)
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}

		after := getBranchByReceipt(receipt, operationInfo, payload.EthAddress, payload.Chain33Addr, nil, nil, newAccountId, payload.TokenId, receipt.Token.Balance)
		kv := &types.KeyValue{
			Key:   getHeightKey(a.height),
			Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
		}
		kvs = append(kvs, kv)

		branch := &zt.OperationPairBranch{
			Before: before,
			After:  after,
		}
		operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)
		zklog := &zt.ZkReceiptLog{
			OperationInfo: operationInfo,
			LocalKvs:      localKvs,
		}
		receiptLog := &types.ReceiptLog{Ty: zt.TyDepositLog, Log: types.Encode(zklog)}
		logs = append(logs, receiptLog)
	} else {
		accountId := leaf.GetAccountId()

		receipt, err := calProof(a.statedb, info, leaf.AccountId, payload.TokenId)
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}

		var balance string
		if receipt.Token == nil {
			balance = "0"
		} else {
			balance = receipt.Token.Balance
		}
		before := getBranchByReceipt(receipt, operationInfo, payload.EthAddress, payload.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, accountId, payload.TokenId, balance)

		kvs, localKvs, err = UpdateLeaf(a.statedb, a.localDB, info, leaf.GetAccountId(), payload.GetTokenId(), payload.GetAmount(), zt.Add)
		if err != nil {
			return nil, errors.Wrapf(err, "db.UpdateLeaf")
		}
		receipt, err = calProof(a.statedb, info, leaf.AccountId, payload.TokenId)
		if err != nil {
			return nil, errors.Wrapf(err, "calProof")
		}
		after := getBranchByReceipt(receipt, operationInfo, payload.EthAddress, payload.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, accountId, payload.TokenId, receipt.Token.Balance)
		kv := &types.KeyValue{
			Key:   getHeightKey(a.height),
			Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
		}
		kvs = append(kvs, kv)

		branch := &zt.OperationPairBranch{
			Before: before,
			After:  after,
		}
		operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)
		zklog := &zt.ZkReceiptLog{
			OperationInfo: operationInfo,
			LocalKvs:      localKvs,
		}
		receiptLog := &types.ReceiptLog{Ty: zt.TyDepositLog, Log: types.Encode(zklog)}
		logs = append(logs, receiptLog)
	}
	//存入1号账户的kv
	for _, kv := range info.kvs {
		if string(kv.GetKey()) != string(GetAccountTreeKey()) {
			kvs = append(kvs, kv)
		}
	}

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	//add priority part
	r := makeSetEthPriorityIdReceipt(lastPriorityId.Int64(), payload.EthPriorityQueueId)
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

func (a *Action) ZkWithdraw(payload *zt.ZkWithdraw) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue
	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}

	//加上手续费
	totalFromAmount, fromFee, err := getAmountWithFee(a.statedb, zt.TyWithdrawAction, payload.Amount, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getAmountWithFee")
	}

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}

	leaf, err := GetLeafByAccountId(a.statedb, payload.GetAccountId(), info)
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

	token, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.AccountId, payload.TokenId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(token, totalFromAmount.String())
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	special := &zt.ZkWithdrawWitnessInfo{
		TokenId:   payload.TokenId,
		Amount:    payload.Amount,
		AccountId: payload.AccountId,
		//ethAddr nil
		Signature: payload.Signature,
		Fee: &zt.ZkFee{
			FromFee: fromFee.String(),
		},
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyWithdrawAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Withdraw{Withdraw: special}},
	}

	//更新之前先计算证明
	receipt, err := calProof(a.statedb, info, payload.AccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	before := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, payload.AccountId, payload.TokenId, receipt.Token.Balance)

	kvs, localKvs, err = UpdateLeaf(a.statedb, a.localDB, info, leaf.GetAccountId(), payload.GetTokenId(), totalFromAmount.String(), zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//取款之后计算证明
	receipt, err = calProof(a.statedb, info, payload.AccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	after := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, payload.AccountId, payload.TokenId, receipt.Token.Balance)

	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
	}
	kvs = append(kvs, kv)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)
	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyWithdrawLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(fromFee.String(), info, payload.TokenId, payload.Signature)
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

//根据系统和token精度，计算合约转化为二层tree侧的amount，合约侧amount都是系统精度
func getTreeSideAmount(amount, totalAmount, fee string, sysDecimal, tokenDecimal int) (string, string, string, error) {
	amountT, err := transferDecimalAmount(amount, sysDecimal, tokenDecimal)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "transferDecimalAmount,amount=%s,tokenDecimal=%d,sysDecimal=%d", amount, tokenDecimal, sysDecimal)
	}
	totalAmountT, err := transferDecimalAmount(totalAmount, sysDecimal, tokenDecimal)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "transferDecimalAmount,amount=%s,tokenDecimal=%d,sysDecimal=%d", totalAmount, tokenDecimal, sysDecimal)
	}
	feeAmountT, err := transferDecimalAmount(fee, sysDecimal, tokenDecimal)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "transferDecimalAmount,amount=%s,tokenDecimal=%d,sysDecimal=%d", fee, tokenDecimal, sysDecimal)
	}
	err = checkPackValue(amountT, zt.PacAmountManBitWidth)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "checkPackVal amount=%s", amountT)
	}
	err = checkPackValue(feeAmountT, zt.PacFeeManBitWidth)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "checkPackVal fee=%s", feeAmountT)
	}
	return amountT, totalAmountT, feeAmountT, nil
}

func (a *Action) ContractToTreeAcctIdProc(payload *zt.ZkContractToTree, token *zt.ZkTokenSymbol) (*types.Receipt, error) {
	var logs []*types.ReceiptLog

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}

	tokenId, ok := new(big.Int).SetString(token.Id, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "tokenId=%s", token.Id)
	}

	//以合约侧精度获取amount和fee
	totalFromAmount, fromFee, err := getAmountWithFee(a.statedb, zt.TyContractToTreeAction, payload.Amount, tokenId.Uint64())
	if err != nil {
		return nil, errors.Wrapf(err, "getAmountWithFee")
	}

	//根据精度，转化到tree侧的amount
	sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")
	amountTree, totalFromAmountTree, feeTree, err := getTreeSideAmount(payload.Amount, totalFromAmount.String(), fromFee.String(), sysDecimal, int(token.Decimal))
	if err != nil {
		return nil, errors.Wrap(err, "getTreeSideAmount")
	}

	leaf, err := GetLeafByAccountId(a.statedb, payload.GetToAccountId(), info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if leaf == nil {
		return nil, errors.New("account:" + strconv.FormatUint(payload.ToAccountId, 10) + " not exist")
	}
	//当前允许转到别的ID
	//accountId 的地址需要和转入者对应，对转入者的保护，防止转到别人ID
	//err = authVerification(payload.GetSignature().PubKey, leaf.GetPubKey())
	//if err != nil {
	//	return nil, errors.Wrapf(err, "authVerification")
	//}

	special := &zt.ZkContractToTreeWitnessInfo{
		TokenId:   tokenId.Uint64(),
		Amount:    amountTree,
		AccountId: payload.ToAccountId,
		Signature: payload.Signature,
		Fee: &zt.ZkFee{
			FromFee: feeTree,
		},
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyContractToTreeAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ContractToTree{ContractToTree: special}},
	}

	//更新合约账户, 扣除 amount+fee
	contractReceipt, err := a.UpdateContractAccount(totalFromAmount.String(), payload.TokenSymbol, zt.Sub, payload.FromExec)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateChainAccount")
	}

	kvs, localKvs, err := a.transferProc(zt.SystemTree2ContractAcctId, payload.ToAccountId, tokenId.Uint64(), totalFromAmountTree, amountTree, info, operationInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "transferProc")
	}

	zksynclog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyContractToTreeLog, Log: types.Encode(zksynclog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeTree, info, tokenId.Uint64(), payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	return mergeReceipt(receipts, contractReceipt), nil
}

//contract2Tree 根据ethAddr、layer2Addr 自动创建新accountId
func (a *Action) contractToTreeNewProc(payload *zt.ZkContractToTree, token *zt.ZkTokenSymbol) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
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

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}
	toLeaf, err := GetLeafByChain33AndEthAddress(a.statedb, payload.ToLayer2Addr, payload.ToEthAddr, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}
	//toAccount存在，走accountId流程, 因为contract2New 电路不验证签名
	if toLeaf != nil {
		payload.ToAccountId = toLeaf.AccountId
		return a.ContractToTreeAcctIdProc(payload, token)
	}

	//accountId 的地址需要和转入者对应，对转入者的保护，防止转到别人ID
	//err = authVerification(payload.GetSignature().PubKey, leaf.GetPubKey())
	//if err != nil {
	//	return nil, errors.Wrapf(err, "authVerification")
	//}
	tokenId, ok := new(big.Int).SetString(token.Id, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "tokenId=%s", token.Id)
	}
	totalFromAmount, fromFee, err := getAmountWithFee(a.statedb, zt.TyContractToTreeAction, payload.Amount, tokenId.Uint64())
	if err != nil {
		return nil, errors.Wrapf(err, "getAmountWithFee")
	}

	//根据精度，转化到tree侧的amount
	sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")
	amountTree, totalFromAmountTree, feeTree, err := getTreeSideAmount(payload.Amount, totalFromAmount.String(), fromFee.String(), sysDecimal, int(token.Decimal))
	if err != nil {
		return nil, errors.Wrap(err, "getTreeSideAmount")
	}

	special := &zt.ZkContractToTreeNewWitnessInfo{
		TokenId: tokenId.Uint64(),
		Amount:  amountTree,
		//ToAccountId: nil,
		EthAddress: payload.ToEthAddr,
		Layer2Addr: payload.ToLayer2Addr,
		Signature:  payload.Signature,
		Fee: &zt.ZkFee{
			FromFee: feeTree,
		},
	}
	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		//这个action是为了电路里面做区分
		TxType:      zt.TyContractToTreeNewAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Contract2TreeNew{Contract2TreeNew: special}},
	}

	//更新合约账户
	contractReceipt, err := a.UpdateContractAccount(totalFromAmount.String(), payload.TokenSymbol, zt.Sub, payload.FromExec)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateContractAccount")
	}

	kvs, localKvs, err := a.transfer2NewProc(zt.SystemTree2ContractAcctId, tokenId.Uint64(), payload.ToEthAddr, payload.ToLayer2Addr, totalFromAmountTree, amountTree, info, operationInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "transfer2NewProc")
	}
	//设置new account id, 在重建proof时候使用
	special.ToAccountId = operationInfo.OperationBranches[len(operationInfo.OperationBranches)-1].After.AccountWitness.ID

	zksynclog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyContractToTreeLog, Log: types.Encode(zksynclog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeTree, info, tokenId.Uint64(), payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)

	return mergeReceipt(receipts, contractReceipt), nil
}

//ContractToTree 用户从L2转入到合约账户的资产转回到L2，amount为实际到账金额，fee额外从用户账户扣除
//也就是用户实际资产total>=amount+fee, 如果小于则提取不成功
//contract2tree的输入amount和fee精度统一是chain33合约精度，比如1e8，在二层再根据具体token精度做适配
func (a *Action) ContractToTree(payload *zt.ZkContractToTree) (*types.Receipt, error) {
	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}

	token, err := getTokenBySymbol(a.statedb, payload.TokenSymbol)
	if err != nil {
		return nil, err
	}
	if token != nil && token.Id == "" {
		return nil, errors.Wrapf(types.ErrInvalidParam, "symbol=%s id is invalid", payload.TokenSymbol)
	}

	//cfg decimal 系统启动时候做过检查，和0的个数一致，缺省1e8有8个0
	//如果设置了toAccountId 直接转到id，否则根据ethAddr、layer2Addr创建新account
	if payload.GetToAccountId() > 0 {
		return a.ContractToTreeAcctIdProc(payload, token)
	}

	return a.contractToTreeNewProc(payload, token)
}

//from向to小数对齐，如果from>to, 需要裁减掉差别部分，且差别部分需要全0，如果from<to,差别部分需要补0
func transferDecimalAmount(amount string, fromDecimal, toDecimal int) (string, error) {
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

//TreeToContract amount需要和签名的一致，签名需要电路验证，amount需要包括精度位
func (a *Action) TreeToContract(payload *zt.ZkTreeToContract) (*types.Receipt, error) {
	err := checkParam(payload.GetAmount())
	if err != nil {
		return nil, err
	}

	token, err := getTokenById(a.statedb, strconv.Itoa(int(payload.TokenId)))
	if err != nil {
		return nil, errors.Wrapf(err, "getTokenId=%d", payload.TokenId)
	}

	//cfg decimal 系统启动时候做过检查，和0的个数一致，缺省1e8有8个0
	s := strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision()))
	sysDecimal := strings.Count(s, "0")
	contractAmount, err := transferDecimalAmount(payload.GetAmount(), int(token.Decimal), sysDecimal)
	if err != nil {
		return nil, errors.Wrapf(err, "transfer2ContractAmount,tokenDecimal=%d,sysDecimal=%d", token.Decimal, sysDecimal)
	}

	//增加systemTree2ContractId 是为了验证签名，同时防止重放攻击，也可以和transfer重用电路
	if payload.ToAcctId != zt.SystemTree2ContractAcctId {
		return nil, errors.Wrapf(types.ErrInvalidParam, "toAcctId not systemId=%d", zt.SystemTree2ContractAcctId)
	}

	//本质上是向另一个账户转账，底层重用了transfer电路(节省电路规模)，
	//二层的转账，amount需要和签过名的payload一致，电路需要验证
	transfer := &zt.ZkTransfer{
		TokenId:       payload.TokenId,
		Amount:        payload.Amount,
		FromAccountId: payload.AccountId,
		ToAccountId:   zt.SystemTree2ContractAcctId,
		Signature:     payload.Signature,
	}
	receipt, err := a.transfer(transfer, zt.TyTreeToContractAction)
	if err != nil {
		return nil, errors.Wrapf(err, "transfer")
	}

	//更新合约账户,合约账户需要用符合chain33合约的精度的amount来更新
	contractReceipt, err := a.UpdateContractAccount(contractAmount, token.Symbol, zt.Add, payload.ToExec)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateContractAccount")
	}
	return mergeReceipt(receipt, contractReceipt), nil
}

func makeSetTokenSymbolReceipt(id string, oldVal, newVal *zt.ZkTokenSymbol) *types.Receipt {
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
//比如token1已经设置为USDT,token2再设置为USDT是不允许的，这样如果曾经设过的(包括错的)symbol都是不允许再设置的
//对于精度设置错了影响会更大，只有在SystemTree2ContractAcctId的当前token为0的时候才允许修改，也就意味着合约里面的资产按当前精度已经都提回到tree了，tree内的资产本身自带精度
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

	//首先检查symbol是否存在，不允许对已存在的symbol做修改，如果之前设置错了，也需要修改个新的名字设置
	token, err := getTokenBySymbol(a.statedb, payload.Symbol)
	if err != nil && !isNotFound(errors.Cause(err)) {
		return nil, err
	}
	//已有symbol未设置为invaild不允许更换id
	if token != nil && token.Id != "" && token.Id != payload.Id {
		return nil, errors.Wrapf(types.ErrNotAllow, "tokenSymbol=%s existed id=%s", payload.Symbol, token.Id)
	}

	//id初始设置或者重设symbol
	lastSym, err := getTokenById(a.statedb, payload.Id)
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
	return makeSetTokenSymbolReceipt(payload.Id, lastSym, payload), nil
}

func getTokenById(db dbm.KV, tokenId string) (*zt.ZkTokenSymbol, error) {
	key := GetTokenSymbolKey(tokenId)
	r, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "getDb")
	}
	var symbol zt.ZkTokenSymbol
	err = types.Decode(r, &symbol)
	if err != nil {
		return nil, errors.Wrapf(err, "decode")
	}
	return &symbol, nil
}
func getTokenBySymbol(db dbm.KV, symbol string) (*zt.ZkTokenSymbol, error) {
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

//在设置了invalidTx后，平行链从0开始同步到无效交易则设置系统为exodus mode，此模式意味着此链即将停用，资产需要退出到ETH
//目前此模式只限制deposit,withdraw,proxyExit,NFTWithdraw这几个跟L1有关的action，其余action正常执行，直到管理员设置ExodusClearMode清算模式
func isExodusMode(statedb dbm.KV) (bool, error) {
	_, err := getExodusMode(statedb)
	if isNotFound(errors.Cause(err)) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

//exodus 清算模式,管理员设置，禁止除contract2tree外一切L2交易，方便尽快收敛treeRoot
func isExodusClearMode(statedb dbm.KV) (bool, error) {
	mode, err := getExodusMode(statedb)
	if isNotFound(errors.Cause(err)) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return mode >= zt.ExodusClearMode, nil
}

func getExodusMode(db dbm.KV) (int64, error) {
	data, err := db.Get(getExodusModeKey())
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
	//notFoundErr也不允许设置
	if err != nil {
		return nil, err
	}
	if mode != zt.ExodusPrepareMode {
		return nil, errors.Wrapf(types.ErrNotAllow, "current mode=%d,not prepareMode=%d", mode, zt.ExodusPrepareMode)
	}

	return makeSetExodusModeReceipt(mode, int64(payload.GetMode())), nil
}

//newZkSyncAccount create new zkSync contract account
func newZkSyncAccount(cfg *types.Chain33Config, symbol string, db dbm.KV) (*account.DB, error) {
	return account.NewAccountDB(cfg, zt.Zksync, symbol, db)
}

func (a *Action) UpdateContractAccount(amount, symbol string, option int32, execName string) (*types.Receipt, error) {
	//如果是exodus mode下，支持超级管理员提取剩余的资金，如果用户没有提取完的话
	isExodus, err := isExodusMode(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "get isExodusMode")
	}
	if isSuperManager(a.api.GetConfig(), a.fromaddr) && isExodus {
		return nil, nil
	}

	accountdb, err := newZkSyncAccount(a.api.GetConfig(), symbol, a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "newZkSyncAccount")
	}
	shortChange, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "UpdateContractAccount amount=%s", amount)
	}
	var execReceipt types.Receipt
	if option == zt.Sub {
		if len(execName) > 0 {
			r, err := a.UpdateExecAccount(accountdb, shortChange.Int64(), option, execName)
			if err != nil {
				return nil, errors.Wrapf(err, "withdraw from exec=%s,val=%d", execName, shortChange)
			}
			mergeReceipt(&execReceipt, r)
		}

		r, err := accountdb.Burn(a.fromaddr, shortChange.Int64())
		return mergeReceipt(&execReceipt, r), err
	}
	//deposit
	r, err := accountdb.Mint(a.fromaddr, shortChange.Int64())
	if err != nil {
		return nil, errors.Wrapf(err, "deposit val=%d", shortChange)
	}
	mergeReceipt(&execReceipt, r)
	if len(execName) > 0 {
		r, err = a.UpdateExecAccount(accountdb, shortChange.Int64(), option, execName)
		return mergeReceipt(&execReceipt, r), err
	}
	return &execReceipt, nil
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

func (a *Action) transferProc(fromAcctId, toAcctId, tokenId uint64, fromAmount, toAmount string, treeInfo *TreeUpdateInfo, opInfo *zt.OperationInfo) ([]*types.KeyValue, []*types.KeyValue, error) {
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue

	fromLeaf, err := GetLeafByAccountId(a.statedb, fromAcctId, treeInfo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, nil, errors.New("account not exist")
	}

	fromToken, err := GetTokenByAccountIdAndTokenId(a.statedb, fromAcctId, tokenId, treeInfo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(fromToken, fromAmount)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.checkAmount")
	}

	//更新之前先计算证明
	receipt, err := calProof(a.statedb, treeInfo, fromAcctId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}

	before := getBranchByReceipt(receipt, opInfo, fromLeaf.EthAddress, fromLeaf.Chain33Addr, fromLeaf.PubKey, fromLeaf.ProxyPubKeys, fromAcctId, tokenId, receipt.Token.Balance)

	//更新fromLeaf
	fromKvs, fromLocal, err := UpdateLeaf(a.statedb, a.localDB, treeInfo, fromLeaf.GetAccountId(), tokenId, fromAmount, zt.Sub)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)
	//更新之后计算证明
	receipt, err = calProof(a.statedb, treeInfo, fromAcctId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}

	after := getBranchByReceipt(receipt, opInfo, fromLeaf.EthAddress, fromLeaf.Chain33Addr, fromLeaf.PubKey, fromLeaf.ProxyPubKeys, fromAcctId, tokenId, receipt.Token.Balance)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	opInfo.OperationBranches = append(opInfo.GetOperationBranches(), branch)

	toLeaf, err := GetLeafByAccountId(a.statedb, toAcctId, treeInfo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if toLeaf == nil {
		return nil, nil, errors.New("account not exist")
	}

	//更新之前先计算证明
	receipt, err = calProof(a.statedb, treeInfo, toAcctId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}
	var balance string
	if receipt.Token == nil {
		balance = "0"
	} else {
		balance = receipt.Token.Balance
	}
	before = getBranchByReceipt(receipt, opInfo, toLeaf.EthAddress, toLeaf.Chain33Addr, toLeaf.PubKey, toLeaf.ProxyPubKeys, toAcctId, tokenId, balance)

	//更新toLeaf
	tokvs, toLocal, err := UpdateLeaf(a.statedb, a.localDB, treeInfo, toLeaf.GetAccountId(), tokenId, toAmount, zt.Add)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	kvs = append(kvs, tokvs...)
	localKvs = append(localKvs, toLocal...)
	//更新之后计算证明
	receipt, err = calProof(a.statedb, treeInfo, toAcctId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}
	after = getBranchByReceipt(receipt, opInfo, toLeaf.EthAddress, toLeaf.Chain33Addr, toLeaf.PubKey, toLeaf.ProxyPubKeys, toAcctId, tokenId, receipt.Token.Balance)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
	}
	kvs = append(kvs, kv)

	branch = &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	opInfo.OperationBranches = append(opInfo.GetOperationBranches(), branch)

	return kvs, localKvs, nil
}

func (a *Action) ZkTransfer(payload *zt.ZkTransfer) (*types.Receipt, error) {
	return a.transfer(payload, zt.TyTransferAction)
}

func (a *Action) transfer(payload *zt.ZkTransfer, actionTy int32) (*types.Receipt, error) {
	var logs []*types.ReceiptLog

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

	//加上手续费, 按actionTy获取fee
	totalFromAmount, fromFee, err := getAmountWithFee(a.statedb, actionTy, payload.Amount, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getAmountWithFee")
	}

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}
	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.GetFromAccountId(), info)
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

	special := &zt.ZkTransferWitnessInfo{
		TokenId:       payload.TokenId,
		Amount:        payload.Amount,
		FromAccountId: payload.FromAccountId,
		ToAccountId:   payload.ToAccountId,
		Signature:     payload.Signature,
		Fee: &zt.ZkFee{
			FromFee: fromFee.String(),
		},
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyTransferAction, //这个类型给电路使用, tree2contract和transfer共用相同电路类型,减少电路规模
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Transfer{Transfer: special}},
	}

	kvs, localKvs, err := a.transferProc(payload.FromAccountId, payload.ToAccountId, payload.TokenId, totalFromAmount.String(), payload.Amount, info, operationInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "transferProc")
	}
	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyTransferLog, Log: types.Encode(zklog)}
	if actionTy == zt.TyTreeToContractAction {
		receiptLog.Ty = zt.TyTreeToContractLog
	}
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(fromFee.String(), info, payload.TokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

func (a *Action) transfer2NewProc(fromAcctId, tokenId uint64, toEthAddr, toLayer2Addr, fromAmount, toAmount string, treeInfo *TreeUpdateInfo, operationInfo *zt.OperationInfo) ([]*types.KeyValue, []*types.KeyValue, error) {
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue

	fromLeaf, err := GetLeafByAccountId(a.statedb, fromAcctId, treeInfo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, nil, errors.New("account not exist")
	}

	fromToken, err := GetTokenByAccountIdAndTokenId(a.statedb, fromAcctId, tokenId, treeInfo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(fromToken, fromAmount)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.checkAmount")
	}

	toLeaf, err := GetLeafByChain33AndEthAddress(a.statedb, toLayer2Addr, toEthAddr, treeInfo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.GetLeafByChain33AndEthAddress")
	}
	//不能直接走transfer流程，因为电路签名不一样
	if toLeaf != nil {
		return nil, nil, errors.Wrapf(types.ErrNotAllow, "toAccountId=%d existed", toLeaf.AccountId)
	}
	//更新之前先计算证明
	receipt, err := calProof(a.statedb, treeInfo, fromAcctId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}

	before := getBranchByReceipt(receipt, operationInfo, fromLeaf.EthAddress, fromLeaf.Chain33Addr, fromLeaf.PubKey, fromLeaf.ProxyPubKeys, fromAcctId, tokenId, receipt.Token.Balance)

	//更新fromLeaf
	fromkvs, fromLocal, err := UpdateLeaf(a.statedb, a.localDB, treeInfo, fromAcctId, tokenId, fromAmount, zt.Sub)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	kvs = append(kvs, fromkvs...)
	localKvs = append(localKvs, fromLocal...)
	//更新之后计算证明
	receipt, err = calProof(a.statedb, treeInfo, fromAcctId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}

	after := getBranchByReceipt(receipt, operationInfo, fromLeaf.EthAddress, fromLeaf.Chain33Addr, fromLeaf.PubKey, fromLeaf.ProxyPubKeys, fromAcctId, tokenId, receipt.Token.Balance)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)

	tree, err := getAccountTree(a.statedb, treeInfo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.getAccountTree")
	}
	newAcctId := tree.GetTotalIndex()
	//更新之前先计算证明
	receipt, err = calProof(a.statedb, treeInfo, newAcctId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}

	before = getBranchByReceipt(receipt, operationInfo, toEthAddr, toLayer2Addr, nil, nil, newAcctId, tokenId, "0")

	//新增toLeaf
	tokvs, toLocal, err := AddNewLeaf(a.statedb, a.localDB, treeInfo, toEthAddr, tokenId, toAmount, toLayer2Addr)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.AddNewLeaf")
	}
	kvs = append(kvs, tokvs...)
	localKvs = append(localKvs, toLocal...)
	//新增之后计算证明
	receipt, err = calProof(a.statedb, treeInfo, newAcctId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}

	after = getBranchByReceipt(receipt, operationInfo, toEthAddr, toLayer2Addr, nil, nil, newAcctId, tokenId, receipt.Token.Balance)
	branch = &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)

	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
	}
	kvs = append(kvs, kv)

	return kvs, localKvs, nil
}

func (a *Action) TransferToNew(payload *zt.ZkTransferToNew) (*types.Receipt, error) {
	var logs []*types.ReceiptLog

	err := checkParam(payload.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	err = checkPackValue(payload.Amount, zt.PacAmountManBitWidth)
	if err != nil {
		return nil, errors.Wrapf(err, "checkPackVal")
	}

	//加上手续费
	totalFromAmount, fromFee, err := getAmountWithFee(a.statedb, zt.TyTransferToNewAction, payload.Amount, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getAmountWithFee")
	}

	//转换10进制
	newAddr, ok := zt.HexAddr2Decimal(payload.ToChain33Address)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer chain33Addr=%s", payload.ToChain33Address)
	}
	payload.ToChain33Address = newAddr

	newAddr, ok = zt.HexAddr2Decimal(payload.ToEthAddress)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "transfer EthAddress=%s", payload.ToEthAddress)
	}
	payload.ToEthAddress = newAddr

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}
	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.GetFromAccountId(), info)
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

	special := &zt.ZkTransferToNewWitnessInfo{
		TokenId:       payload.TokenId,
		Amount:        payload.Amount,
		FromAccountId: payload.FromAccountId,
		//ToAccountId: nil
		EthAddress: payload.ToEthAddress,
		Layer2Addr: payload.ToChain33Address,
		Signature:  payload.Signature,
		Fee: &zt.ZkFee{
			FromFee: fromFee.String(),
		},
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyTransferToNewAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TransferToNew{TransferToNew: special}},
	}

	kvs, localKvs, err := a.transfer2NewProc(payload.FromAccountId, payload.TokenId, payload.ToEthAddress, payload.ToChain33Address, totalFromAmount.String(), payload.Amount, info, operationInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "transfer2NewProc")
	}
	//设置new AccountId 在重建proof时候使用
	special.ToAccountId = operationInfo.OperationBranches[len(operationInfo.OperationBranches)-1].After.AccountWitness.ID

	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyTransferToNewLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(fromFee.String(), info, payload.TokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
}

//ProxyExit 代理其他账户退出机制，在其他账户设置L2地址出错而不能设置公钥提款的场景，可由代理账户帮助退出回其eth地址
func (a *Action) ProxyExit(payload *zt.ZkProxyExit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue

	if payload.ProxyId == payload.TargetId {
		return nil, errors.Wrapf(types.ErrInvalidParam, "proxyId same as targetId")
	}

	feeInfo, err := getFeeData(a.statedb, zt.TyProxyExitAction, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}
	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}
	proxyLeaf, err := GetLeafByAccountId(a.statedb, payload.ProxyId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	if proxyLeaf == nil {
		return nil, errors.Wrapf(types.ErrNotAllow, "proxy account=%d not exist", payload.ProxyId)
	}

	proxyToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.ProxyId, payload.TokenId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "proxyId.GetTokenByAccountIdAndTokenId")
	}
	//token不存在时，退出
	if proxyToken == nil {
		return nil, errors.Wrapf(types.ErrNotAllow, "proxy account token not find")
	}
	err = authVerification(payload.Signature.PubKey, proxyLeaf.PubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}

	//get target Id
	targetLeaf, err := GetLeafByAccountId(a.statedb, payload.TargetId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "get targeId account=%d", payload.TargetId)
	}
	//如果targetId已经设置过pubkey，说明用户地址设置没错，由用户自己提款
	if targetLeaf == nil || targetLeaf.PubKey != nil {
		return nil, errors.Wrapf(types.ErrNotAllow, "target account=%d not exist or pubkey existed", payload.TargetId)
	}

	targetToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.TargetId, payload.TokenId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "targetId.GetTokenByAccountIdAndTokenId")
	}
	//token不存在时，退出
	if targetToken == nil || targetToken.Balance == "0" {
		return nil, errors.Wrapf(types.ErrNotAllow, "target account token not find or zero")
	}

	//加上手续费
	amountInt, ok := new(big.Int).SetString(proxyToken.Balance, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "balance=%s", proxyToken.Balance)
	}
	feeInt, ok := new(big.Int).SetString(feeInfo.FromFee, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "fromFee=%s", feeInfo.FromFee)
	}
	//存量不够手续费时，不能取
	if amountInt.Cmp(feeInt) <= 0 {
		return nil, errors.New("no enough fee")
	}

	specialInfo := &zt.ZkProxyExitWitnessInfo{
		ProxyID:   payload.ProxyId,
		TargetId:  payload.TargetId,
		TokenId:   payload.TokenId,
		Amount:    targetToken.Balance,
		Signature: payload.Signature,
		Fee:       feeInfo,
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyProxyExitAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ProxyExit{ProxyExit: specialInfo}},
	}

	//更新之前先计算证明
	receipt, err := calProof(a.statedb, info, payload.GetProxyId(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	before := getBranchByReceipt(receipt, operationInfo, proxyLeaf.EthAddress, proxyLeaf.Chain33Addr, proxyLeaf.PubKey, proxyLeaf.ProxyPubKeys, payload.ProxyId, payload.TokenId, receipt.Token.Balance)

	//更新fromLeaf
	proxyKvs, proxyLocalKvs, err := UpdateLeaf(a.statedb, a.localDB, info, proxyLeaf.GetAccountId(), payload.GetTokenId(), feeInfo.FromFee, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	kvs = append(kvs, proxyKvs...)
	localKvs = append(localKvs, proxyLocalKvs...)
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.GetProxyId(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}

	after := getBranchByReceipt(receipt, operationInfo, proxyLeaf.EthAddress, proxyLeaf.Chain33Addr, proxyLeaf.PubKey, proxyLeaf.ProxyPubKeys, proxyLeaf.AccountId, payload.TokenId, receipt.Token.Balance)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
	}
	kvs = append(kvs, kv)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)

	//2. process targetId
	receipt, err = calProof(a.statedb, info, payload.GetTargetId(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "tagetId calProof before")
	}
	targetBefore := getBranchByReceipt(receipt, operationInfo, targetLeaf.EthAddress, targetLeaf.Chain33Addr, targetLeaf.PubKey, targetLeaf.ProxyPubKeys, payload.TargetId, payload.TokenId, receipt.Token.Balance)

	//balance全部取出
	targetKvs, targetLocalKvs, err := UpdateLeaf(a.statedb, a.localDB, info, targetLeaf.GetAccountId(), payload.GetTokenId(), targetToken.Balance, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	kvs = append(kvs, targetKvs...)
	localKvs = append(localKvs, targetLocalKvs...)
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.GetTargetId(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "targetId calProof after")
	}

	targetAfter := getBranchByReceipt(receipt, operationInfo, targetLeaf.EthAddress, targetLeaf.Chain33Addr, targetLeaf.PubKey, targetLeaf.ProxyPubKeys, targetLeaf.AccountId, payload.TokenId, receipt.Token.Balance)
	kv = &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
	}
	kvs = append(kvs, kv)

	targetBranch := &zt.OperationPairBranch{
		Before: targetBefore,
		After:  targetAfter,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), targetBranch)

	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyProxyExitLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeInfo.FromFee, info, payload.TokenId, payload.Signature)
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
	payload.PubKey.X = zt.FilterHexPrefix(payload.PubKey.X)
	payload.PubKey.Y = zt.FilterHexPrefix(payload.PubKey.Y)

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

	special := &zt.ZkSetPubKeyWitnessInfo{
		AccountId: payload.AccountId,
		PubKeyTy:  payload.PubKeyTy,
		PubKey:    payload.PubKey,
		Signature: payload.Signature,
	}
	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TySetPubKeyAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_SetPubKey{SetPubKey: special}},
	}

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
	before := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, nil, nil, payload.AccountId, leaf.TokenIds[0], receipt.Token.Balance)

	kvs, localKvs, err := UpdatePubKey(a.statedb, a.localDB, info, payload.GetPubKeyTy(), payload.GetPubKey(), payload.AccountId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.AccountId, leaf.TokenIds[0])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}
	after := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, payload.PubKey, nil, payload.AccountId, leaf.TokenIds[0], receipt.Token.Balance)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
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
	before := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, payload.AccountId, leaf.TokenIds[0], receipt.Token.Balance)

	kvs, localKvs, err := UpdatePubKey(a.statedb, a.localDB, info, payload.PubKeyTy, payload.GetPubKey(), payload.AccountId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.AccountId, leaf.TokenIds[0])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "calProof")
	}
	after := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, receipt.Leaf.GetProxyPubKeys(), payload.AccountId, leaf.TokenIds[0], receipt.Token.Balance)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
	}
	kvs = append(kvs, kv)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)

	return kvs, localKvs, nil
}

//FullExit 为了避免L2不打包某些用户的从L2向L1的提款交易，fullExit是从L1发起的，强制L2打包提款的交易，L1会在证明中校验，当前没支持此类交易
func (a *Action) FullExit(payload *zt.ZkFullExit) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue

	//只有管理员能操作
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, zt.ZkParaChainInnerTitleId, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
	}

	//fullexit last priority id 不能为空
	lastPriority, err := getLastEthPriorityQueueID(a.statedb)
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

	info, err := getTreeUpdateInfo(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}
	leaf, err := GetLeafByAccountId(a.statedb, payload.AccountId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}

	if leaf == nil {
		return nil, errors.New("account not exist")
	}

	token, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.AccountId, payload.TokenId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}

	//token不存在时，不需要取
	if token == nil {
		return nil, errors.New("token not find")
	}

	feeInfo, err := getFeeData(a.statedb, zt.TyFullExitAction, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}
	//加上手续费
	amountInt, _ := new(big.Int).SetString(token.Balance, 10)
	makerFeeInt, _ := new(big.Int).SetString(feeInfo.FromFee, 10)
	exitAmount := "0"
	//存量不够手续费时，都算进手续费
	if amountInt.Cmp(makerFeeInt) <= 0 {
		//amount当手续费
		makerFeeInt.Set(amountInt)
	} else {
		exitAmount = new(big.Int).Sub(amountInt, makerFeeInt).String()
	}

	specialInfo := &zt.ZkFullExitWitnessInfo{
		AccountId: payload.AccountId,
		TokenId:   payload.TokenId,
		Amount:    exitAmount,
		Signature: payload.Signature,
		Fee: &zt.ZkFee{
			FromFee: makerFeeInt.String(),
		},
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight:   uint64(a.height),
		TxIndex:       uint32(a.index),
		TxType:        zt.TyFullExitAction,
		EthPriorityId: payload.EthPriorityQueueId,
		SpecialInfo:   &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_FullExit{FullExit: specialInfo}},
	}

	//更新之前先计算证明
	receipt, err := calProof(a.statedb, info, payload.GetAccountId(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	before := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, payload.AccountId, payload.TokenId, receipt.Token.Balance)

	//更新fromLeaf
	kvs, localKvs, err = UpdateLeaf(a.statedb, a.localDB, info, leaf.GetAccountId(), payload.GetTokenId(), token.Balance, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.GetAccountId(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}

	after := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, payload.AccountId, payload.TokenId, receipt.Token.Balance)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
	}
	kvs = append(kvs, kv)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)
	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyFullExitLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	//add priority part
	r := makeSetEthPriorityIdReceipt(lastId.Int64(), payload.EthPriorityQueueId)

	feeReceipt, err := a.MakeFeeLog(makerFeeInt.String(), info, payload.TokenId, payload.Signature)
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

func getLastEthPriorityQueueID(db dbm.KV) (*zt.EthPriorityQueueID, error) {
	key := getEthPriorityQueueKey()
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

func makeSetEthPriorityIdReceipt(prev, current int64) *types.Receipt {
	key := getEthPriorityQueueKey()
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

func (a *Action) MakeFeeLog(amount string, info *TreeUpdateInfo, tokenId uint64, sign *zt.ZkSignature) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue
	var err error

	//todo 手续费收款方accountId可配置
	leaf, err := GetLeafByAccountId(a.statedb, zt.SystemFeeAccountId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}

	if leaf == nil {
		return nil, errors.New("account not exist")
	}

	specialInfo := &zt.ZkFeeWitnessInfo{
		AccountId: zt.SystemFeeAccountId,
		TokenId:   tokenId,
		Amount:    amount,
		Signature: sign,
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		OpIndex:     1,
		TxType:      zt.TyFeeAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Fee{Fee: specialInfo}},
	}

	//leaf不存在就添加

	receipt, err := calProof(a.statedb, info, leaf.AccountId, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}

	var balance string
	if receipt.Token == nil {
		balance = "0"
	} else {
		balance = receipt.Token.Balance
	}
	before := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, leaf.GetAccountId(), tokenId, balance)

	kvs, localKvs, err = UpdateLeaf(a.statedb, a.localDB, info, leaf.GetAccountId(), tokenId, amount, zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	receipt, err = calProof(a.statedb, info, leaf.AccountId, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	after := getBranchByReceipt(receipt, operationInfo, leaf.EthAddress, leaf.Chain33Addr, leaf.PubKey, leaf.ProxyPubKeys, leaf.GetAccountId(), tokenId, receipt.Token.Balance)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: types.Encode(&types.TxHash{Hash: receipt.TreeProof.RootHash}),
	}
	kvs = append(kvs, kv)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)
	feelog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyFeeLog, Log: types.Encode(feelog)}
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

//电路会把大的amount以压缩格式表示，分为man主体和exp指数部分，大量的末尾的0用2的指数表示，最多31个0也就是5位
//剩余的主体部分不做压缩表示，最大不能超过可表示的最大值，不然电路无法表示
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

func (a *Action) setFee(payload *zt.ZkSetFee) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := a.api.GetConfig()

	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, zt.ZkParaChainInnerTitleId, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
	}

	//跟其他action手续费以二层各token精度一致不同，contract2tree action的手续费统一以合约的精度处理，这样和合约侧amount精度一致，简化了合约侧的精度处理
	if payload.ActionTy == zt.TyContractToTreeAction {
		token, err := getTokenById(a.statedb, big.NewInt(int64(payload.TokenId)).String())
		if err != nil {
			return nil, errors.Wrapf(err, "getTokenId=%d", payload.TokenId)
		}
		sysDecimal := strings.Count(strconv.Itoa(int(a.api.GetConfig().GetCoinPrecision())), "0")
		//比如token精度为6，sysDecimal=8, token的fee在sysDecimal下需要补2个0，也就是后缀需要至少有两个0，不然会丢失精度，在token精度大于sys精度时候没这问题
		suffix := strings.Repeat("0", sysDecimal-int(token.Decimal))
		if int(token.Decimal) < sysDecimal && !strings.HasSuffix(payload.Amount, suffix) {
			return nil, errors.Wrapf(types.ErrNotAllow, "contract2tree fee need at least with suffix=%s", suffix)
		}
	}
	//fee 压缩格式检查
	err := checkPackValue(payload.Amount, zt.PacFeeManBitWidth)
	if err != nil {
		return nil, errors.Wrapf(err, "checkPackVal")
	}

	lastFee, err := getDbFeeData(a.statedb, payload.ActionTy, payload.TokenId)
	if isNotFound(errors.Cause(err)) {
		lastFee = "0"
	} else if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
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
		return "", errors.Wrapf(err, "db fee")
	}
	return string(v), nil
}

func getFeeData(db dbm.KV, actionTy int32, tokenId uint64) (*zt.ZkFee, error) {
	//从db读取,强制配置，不配置报错，防止恶意提交小金额批量交易
	minFee, err := getDbFeeData(db, actionTy, tokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getDbFeeData,actionTy=%d,tokenId=%d", actionTy, tokenId)
	}
	return &zt.ZkFee{
		FromFee: minFee,
		ToFee:   "0",
		TokenId: tokenId,
	}, nil
}

func getAmountWithFee(db dbm.KV, actionTy int32, amount string, tokenId uint64) (*big.Int, *big.Int, error) {
	feeInfo, err := getFeeData(db, actionTy, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "getFeeData")
	}

	//加上手续费
	amountInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, nil, errors.Wrapf(types.ErrInvalidParam, "decode amount=%s", amount)
	}
	fromFeeInt, ok := new(big.Int).SetString(feeInfo.GetFromFee(), 10)
	if !ok {
		return nil, nil, errors.Wrapf(types.ErrInvalidParam, "fromFee=%s", feeInfo.GetFromFee())
	}
	totalFromAmount := new(big.Int).Add(amountInt, fromFeeInt)

	return totalFromAmount, fromFeeInt, nil
}

func (a *Action) MintNFT(payload *zt.ZkMintNFT) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue

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

	ethFeeAddr, chain33FeeAddr := getCfgFeeAddr(a.api.GetConfig())
	info, err := generateTreeUpdateInfo(a.statedb, a.localDB, ethFeeAddr, chain33FeeAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "db.generateTreeUpdateInfo")
	}

	//暂定0 后面从数据库读取 TODO
	feeTokenId := uint64(0)

	//如果tokenId来自setFee,需要系统预先设置过
	feeInfo, err := getFeeData(a.statedb, zt.TyMintNFTAction, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}

	specialInfo := &zt.ZkMintNFTWitnessInfo{
		MintAcctId:  payload.FromAccountId,
		RecipientId: payload.RecipientId,
		ErcProtocol: payload.ErcProtocol,
		ContentHash: []string{contentPart1.String(), contentPart2.String()},
		Amount:      new(big.Int).SetUint64(payload.Amount).String(),
		Signature:   payload.Signature,
		Fee:         feeInfo,
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyMintNFTAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_MintNFT{MintNFT: specialInfo}},
	}

	//1. calc fee
	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.GetFromAccountId(), info)
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
	err = checkAmount(feeToken, feeInfo.FromFee)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	newBranch, fromKvs, fromLocal, err := a.updateLeafRst(info, operationInfo, fromLeaf, feeTokenId, feeInfo.FromFee, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.fee")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	//2. creator SystemNFTTokenId balance+1 产生serialId
	fromLeaf, err = GetLeafByAccountId(a.statedb, payload.GetFromAccountId(), info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.2")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}
	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, fromLeaf, zt.SystemNFTTokenId, "1", zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.creator.nftToken")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)
	//serialId表示createor创建了多少nft,这里使用before的id
	creatorSerialId := newBranch.Before.TokenWitness.Balance
	creatorEthAddr := fromLeaf.EthAddress

	//3. SystemNFTAccountId's SystemNFTTokenId+1, 产生新的NFT的id
	fromLeaf, err = GetLeafByAccountId(a.statedb, zt.SystemNFTAccountId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.NFTAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}

	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, fromLeaf, zt.SystemNFTTokenId, "1", zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.NFTAccountId.nftToken")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	newNFTTokenId, ok := big.NewInt(0).SetString(newBranch.Before.TokenWitness.Balance, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "new NFT token balance=%s nok", newBranch.After.TokenWitness.Balance)
	}
	if newNFTTokenId.Uint64() <= zt.SystemNFTTokenId {
		return nil, errors.Wrapf(types.ErrNotAllow, "newNFTTokenId=%d should big than default %d", newNFTTokenId.Uint64(), zt.SystemNFTTokenId)
	}
	specialInfo.NewNFTTokenID = newNFTTokenId.Uint64()
	serialId, ok := big.NewInt(0).SetString(creatorSerialId, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "creatorSerialId=%s nok", creatorSerialId)
	}
	specialInfo.CreateSerialId = serialId.Uint64()

	//4. SystemNFTAccountId set new NFT id to balance by NFT contentHash
	fromLeaf, err = GetLeafByAccountId(a.statedb, zt.SystemNFTAccountId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId.NFTAccountId.NewNFT")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}

	newNFTTokenBalance, err := getNewNFTTokenBalance(payload.GetFromAccountId(), creatorSerialId, payload.ErcProtocol, payload.Amount, contentPart1.String(), contentPart2.String())
	if err != nil {
		return nil, errors.Wrapf(err, "getNewNFTToken balance")
	}

	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, fromLeaf, newNFTTokenId.Uint64(), newNFTTokenBalance, zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.NFTAccountId.nftToken")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

	//5. recipientAddr new NFT id balance+amount
	toLeaf, err := GetLeafByAccountId(a.statedb, payload.GetRecipientId(), info)
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
	newBranch, fromKvs, fromLocal, err = a.updateLeafRst(info, operationInfo, toLeaf, newNFTTokenId.Uint64(), big.NewInt(0).SetUint64(payload.Amount).String(), zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "updateLeafRst.NFTAccountId.nftToken")
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), newBranch)
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)

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

	//end
	zklog := &zt.ZkReceiptLog{
		OperationInfo: operationInfo,
		LocalKvs:      localKvs,
	}
	receiptLog := &types.ReceiptLog{Ty: zt.TyMintNFTLog, Log: types.Encode(zklog)}
	logs = append(logs, receiptLog)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}

	feeReceipt, err := a.MakeFeeLog(feeInfo.FromFee, info, feeTokenId, payload.Signature)
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

	feeInfo, err := getFeeData(a.statedb, zt.TyWithdrawNFTAction, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}
	amountStr := big.NewInt(0).SetUint64(payload.Amount).String()

	nftStatus, err := getNFTById(a.statedb, payload.NFTTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getNFTById=%d", payload.NFTTokenId)
	}

	contentHashPart1, contentHashPart2, _, err := zt.SplitNFTContent(nftStatus.ContentHash)
	if err != nil {
		return nil, errors.Wrapf(err, "split content hash=%s", nftStatus.ContentHash)
	}

	specialInfo := &zt.ZkWithdrawNFTWitnessInfo{
		FromAcctId:      payload.FromAccountId,
		NFTTokenID:      payload.NFTTokenId,
		WithdrawAmount:  new(big.Int).SetUint64(payload.Amount).String(),
		CreatorAcctId:   nftStatus.CreatorId,
		ErcProtocol:     nftStatus.ErcProtocol,
		ContentHash:     []string{contentHashPart1.String(), contentHashPart2.String()},
		CreatorSerialId: nftStatus.CreatorSerialId,
		InitMintAmount:  new(big.Int).SetUint64(nftStatus.MintAmount).String(),
		Signature:       payload.Signature,
		Fee:             feeInfo,
	}

	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyWithdrawNFTAction,
		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_WithdrawNFT{WithdrawNFT: specialInfo}},
	}

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
	err = checkAmount(feeToken, feeInfo.FromFee)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	newBranch, fromKvs, fromLocal, err := a.updateLeafRst(info, operationInfo, fromLeaf, feeTokenId, feeInfo.FromFee, zt.Sub)
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

	feeReceipt, err := a.MakeFeeLog(feeInfo.FromFee, info, feeTokenId, payload.Signature)
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

	feeInfo, err := getFeeData(a.statedb, zt.TyTransferNFTAction, feeTokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "getFeeData")
	}
	amountStr := big.NewInt(0).SetUint64(payload.Amount).String()

	special := &zt.ZkTransferNFTWitnessInfo{
		FromAccountId: payload.FromAccountId,
		RecipientId:   payload.RecipientId,
		NFTTokenId:    payload.NFTTokenId,
		Amount:        new(big.Int).SetUint64(payload.Amount).String(),
		Signature:     payload.Signature,
		Fee:           feeInfo,
	}
	operationInfo := &zt.OperationInfo{
		BlockHeight: uint64(a.height),
		TxIndex:     uint32(a.index),
		TxType:      zt.TyTransferNFTAction,

		SpecialInfo: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TransferNFT{TransferNFT: special}},
	}

	//nftStatus, err := getNFTById(a.statedb, payload.NFTTokenId)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "getNFTById=%d", payload.NFTTokenId)
	//}

	//speciaData := &zt.OperationSpecialData{
	//	RecipientID: payload.RecipientId,
	//	TokenID:     []uint64{feeTokenId, nftStatus.Id},
	//}
	//operationInfo.SpecialInfo.SpecialDatas = append(operationInfo.SpecialInfo.SpecialDatas, speciaData)

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
	err = checkAmount(feeToken, feeInfo.FromFee)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	newBranch, fromKvs, fromLocal, err := a.updateLeafRst(info, operationInfo, fromLeaf, feeTokenId, feeInfo.FromFee, zt.Sub)
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

	feeReceipt, err := a.MakeFeeLog(feeInfo.FromFee, info, feeTokenId, payload.Signature)
	if err != nil {
		return nil, errors.Wrapf(err, "MakeFeeLog")
	}
	receipts = mergeReceipt(receipts, feeReceipt)
	return receipts, nil
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

//从此invalidTx之后，系统进入退出状态(eth无法接收新的proof场景)，平行链需要从0开始重新同步交易，相当于回滚，无效交易之后的交易都失败(除contract2tree)
//退出状态不允许deposit,withdraw等的其他交易，只允许contract2Tree的退到二层的交易，
//退到二层后，统计各账户id的余额，根据最后的treeRootHash提交退出证明(不能走withdraw等流程退出）
func isInvalidTx(cfg *types.Chain33Config, txHash []byte) bool {
	invalidTx, _ := getCfgInvalidTx(cfg)
	return invalidTx == hex.EncodeToString(txHash)
}

//在此proofRootHash(包括此hash)后的所有历史roothash都会失效，直到替换此hash之后的新的proofhash收到
func isInvalidProof(cfg *types.Chain33Config, rootHash string) bool {
	_, invalidProof := getCfgInvalidTx(cfg)
	return invalidProof == rootHash
}
