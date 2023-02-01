package executor

import (
	"math/big"
	"strings"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
)

func GetL2FirstQueueId(db dbm.KV) (int64, error) {
	key := getL2FirstQueueIdKey()
	r, err := db.Get(key)
	if isNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "getDb")
	}
	var id types.Int64
	err = types.Decode(r, &id)
	if err != nil {
		return 0, errors.Wrapf(err, "decode")
	}
	return id.Data, nil
}

//L2 queue id 从1开始编号，跟L1 priority 不同，后者为了和eth合约编号保持一致
func GetL2LastQueueId(db dbm.KV) (int64, error) {
	key := getL2LastQueueIdKey()
	r, err := db.Get(key)
	if isNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "getDb")
	}
	var id types.Int64
	err = types.Decode(r, &id)
	if err != nil {
		return 0, errors.Wrapf(err, "decode")
	}
	return id.Data, nil
}

func GetL2QueueIdOp(db dbm.KV, id int64) (*zt.ZkOperation, error) {
	key := getL2QueueIdKey(id)
	r, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "getDb")
	}
	var data zt.ZkOperation
	err = types.Decode(r, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode")
	}
	return &data, nil
}

//GetProofId2QueueId proof中的pubdata 对应的operation的start/end queueId
func GetProofId2QueueId(db dbm.KV, id uint64) (*zt.ProofId2QueueIdData, error) {
	key := getProofId2QueueIdKey(id)
	r, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "getDb")
	}
	var data zt.ProofId2QueueIdData
	err = types.Decode(r, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode")
	}
	return &data, nil
}

func GetPriority2QueueId(db dbm.KV, priorityId int64) (int64, error) {
	key := getL1PriorityId2QueueIdKey(priorityId)
	r, err := db.Get(key)
	if err != nil {
		return 0, errors.Wrapf(err, "getDb")
	}
	var data types.Int64
	err = types.Decode(r, &data)
	if err != nil {
		return 0, errors.Wrapf(err, "decode")
	}
	return data.Data, nil
}

func GetPriorityDepositData(db dbm.KV, priorityId int64) (*zt.ZkDepositWitnessInfo, error) {
	queueId, err := GetPriority2QueueId(db, priorityId)
	if err != nil {
		return nil, errors.Wrapf(err, "GetPriority2QueueId=%d", priorityId)
	}
	op, err := GetL2QueueIdOp(db, queueId)
	if err != nil {
		return nil, errors.Wrapf(err, "GetL2QueueIdOp id=%d", queueId)
	}
	if op.Ty != zt.TyDepositAction {
		return nil, errors.Wrapf(types.ErrInvalidParam, "priorityId=%d,to queueId=%d, queue op.ty=%d not deposit", priorityId, queueId, op.Ty)
	}
	return op.GetOp().GetDeposit(), nil
}

func checkOpSame(queueOp, pubDataOp *zt.ZkOperation) error {
	if queueOp == nil || pubDataOp == nil {
		return errors.Wrapf(types.ErrInvalidParam, "nil op: queueOp=%x,pubDataOp=%x ", queueOp, pubDataOp)
	}
	switch queueOp.Ty {
	case zt.TyDepositAction:
		return checkSameDeposit(queueOp, pubDataOp)
	case zt.TyWithdrawAction:
		return checkSameWithdraw(queueOp, pubDataOp)
	case zt.TyTransferAction:
		return checkSameTransfer(queueOp, pubDataOp)
	case zt.TyTransferToNewAction:
		return checkSameTransfer2New(queueOp, pubDataOp)
	case zt.TyProxyExitAction:
		return checkSameProxyExit(queueOp, pubDataOp)
	case zt.TySetPubKeyAction:
		return checkSameSetPubKey(queueOp, pubDataOp)
	case zt.TyContractToTreeAction:
		return checkSameContract2Tree(queueOp, pubDataOp)
	case zt.TyContractToTreeNewAction:
		return checkSameContract2TreeNew(queueOp, pubDataOp)
	case zt.TyTreeToContractAction:
		return checkSameTree2Contract(queueOp, pubDataOp)
	case zt.TyFeeAction:
		return checkSameFee(queueOp, pubDataOp)
	case zt.TyMintNFTAction:
		return checkSameMintNFT(queueOp, pubDataOp)
	case zt.TyWithdrawNFTAction:
		return checkSameWithdrawNFT(queueOp, pubDataOp)
	case zt.TyTransferNFTAction:
		return checkSameTransferNFT(queueOp, pubDataOp)
	default:
		return errors.Wrapf(types.ErrNotFound, "action=%d", queueOp.Ty)
	}
}

func checkSameDeposit(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetDeposit()
	p := pubDataOp.Op.GetDeposit()
	if q.AccountID != p.AccountID {
		return errors.Wrapf(types.ErrInvalidParam, "deposit acctId queue=%d, pub=%d", q.AccountID, p.AccountID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "deposit tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "deposit amount queue=%s, pub=%s", q.Amount, p.Amount)
	}
	if q.EthAddress != p.EthAddress {
		return errors.Wrapf(types.ErrInvalidParam, "deposit ethAddr queue=%s, pub=%s", q.EthAddress, p.EthAddress)
	}
	if q.Layer2Addr != p.Layer2Addr {
		return errors.Wrapf(types.ErrInvalidParam, "deposit layer2Addr queue=%s, pub=%s", q.Layer2Addr, p.Layer2Addr)
	}
	return nil
}

func checkSameWithdraw(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetWithdraw()
	p := pubDataOp.Op.GetWithdraw()
	if q.AccountID != p.AccountID {
		return errors.Wrapf(types.ErrInvalidParam, "withdraw acctId queue=%d, pub=%d", q.AccountID, p.AccountID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "withdraw tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "withdraw amount queue=%s, pub=%s", q.Amount, p.Amount)
	}
	if q.EthAddress != p.EthAddress {
		return errors.Wrapf(types.ErrInvalidParam, "withdraw ethAddr queue=%s, pub=%s", q.EthAddress, p.EthAddress)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "withdraw fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameTransfer(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetTransfer()
	p := pubDataOp.Op.GetTransfer()
	if q.FromAccountID != p.FromAccountID {
		return errors.Wrapf(types.ErrInvalidParam, "transfer from acctId queue=%d, pub=%d", q.FromAccountID, p.FromAccountID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "transfer tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "transfer amount queue=%s, pub=%s", q.Amount, p.Amount)
	}
	if q.ToAccountID != p.ToAccountID {
		return errors.Wrapf(types.ErrInvalidParam, "transfer to AcctId queue=%d, pub=%d", q.ToAccountID, p.ToAccountID)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "transfer fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameTransfer2New(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetTransferToNew()
	p := pubDataOp.Op.GetTransferToNew()
	if q.FromAccountID != p.FromAccountID {
		return errors.Wrapf(types.ErrInvalidParam, "transfer2new from acctId queue=%d, pub=%d", q.FromAccountID, p.FromAccountID)
	}
	if q.ToAccountID != p.ToAccountID {
		return errors.Wrapf(types.ErrInvalidParam, "transfer2new to AcctId queue=%d, pub=%d", q.ToAccountID, p.ToAccountID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "transfer2new tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "transfer2new amount queue=%s, pub=%s", q.Amount, p.Amount)
	}
	if q.EthAddress != p.EthAddress {
		return errors.Wrapf(types.ErrInvalidParam, "transfer2new ethAddr queue=%s, pub=%s", q.EthAddress, p.EthAddress)
	}
	if q.Layer2Addr != p.Layer2Addr {
		return errors.Wrapf(types.ErrInvalidParam, "transfer2new layer2Addr queue=%s, pub=%s", q.Layer2Addr, p.Layer2Addr)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "transfer2new fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameProxyExit(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetProxyExit()
	p := pubDataOp.Op.GetProxyExit()
	if q.ProxyID != p.ProxyID {
		return errors.Wrapf(types.ErrInvalidParam, "proxy proxy id queue=%d, pub=%d", q.ProxyID, p.ProxyID)
	}
	if q.TargetID != p.TargetID {
		return errors.Wrapf(types.ErrInvalidParam, "proxy target AcctId queue=%d, pub=%d", q.TargetID, p.TargetID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "proxy tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "proxy amount queue=%s, pub=%s", q.Amount, p.Amount)
	}
	if q.EthAddress != p.EthAddress {
		return errors.Wrapf(types.ErrInvalidParam, "proxy ethAddr queue=%s, pub=%s", q.EthAddress, p.EthAddress)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "proxy fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameSetPubKey(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetSetPubKey()
	p := pubDataOp.Op.GetSetPubKey()
	if q.AccountID != p.AccountID {
		return errors.Wrapf(types.ErrInvalidParam, "pubkey acct id queue=%d, pub=%d", q.AccountID, p.AccountID)
	}
	if q.PubKeyTy != p.PubKeyTy {
		return errors.Wrapf(types.ErrInvalidParam, "pubkey ty queue=%d, pub=%d", q.PubKeyTy, p.PubKeyTy)
	}
	if q.PubKey.X != p.PubKey.X {
		return errors.Wrapf(types.ErrInvalidParam, "pubkey x queue=%s, pub=%s", q.PubKey.X, p.PubKey.X)
	}
	if q.PubKey.Y != p.PubKey.Y {
		return errors.Wrapf(types.ErrInvalidParam, "pubkey y queue=%s, pub=%s", q.PubKey.Y, p.PubKey.Y)
	}

	return nil
}

func checkSameContract2Tree(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetContractToTree()
	p := pubDataOp.Op.GetContractToTree()
	if q.AccountID != p.AccountID {
		return errors.Wrapf(types.ErrInvalidParam, "contract2tree  acctId queue=%d, pub=%d", q.AccountID, p.AccountID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "contract2tree tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "contract2tree amount queue=%s, pub=%s", q.Amount, p.Amount)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "contract2tree fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameContract2TreeNew(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetContract2TreeNew()
	p := pubDataOp.Op.GetContract2TreeNew()
	if q.ToAccountID != p.ToAccountID {
		return errors.Wrapf(types.ErrInvalidParam, "contract2treeNew to acctId queue=%d, pub=%d", q.ToAccountID, p.ToAccountID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "contract2treeNew tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "contract2treeNew amount queue=%s, pub=%s", q.Amount, p.Amount)
	}
	if q.EthAddress != p.EthAddress {
		return errors.Wrapf(types.ErrInvalidParam, "contract2treeNew ethAddr queue=%s, pub=%s", q.EthAddress, p.EthAddress)
	}
	if q.Layer2Addr != p.Layer2Addr {
		return errors.Wrapf(types.ErrInvalidParam, "contract2treeNew layer2Addr queue=%s, pub=%s", q.Layer2Addr, p.Layer2Addr)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "contract2treeNew fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameTree2Contract(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetTreeToContract()
	p := pubDataOp.Op.GetTreeToContract()
	if q.AccountID != p.AccountID {
		return errors.Wrapf(types.ErrInvalidParam, "tree2contract  acctId queue=%d, pub=%d", q.AccountID, p.AccountID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "tree2contract tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "tree2contract amount queue=%s, pub=%s", q.Amount, p.Amount)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "tree2contract fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameFee(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetFee()
	p := pubDataOp.Op.GetFee()
	if q.AccountID != p.AccountID {
		return errors.Wrapf(types.ErrInvalidParam, "fee  acctId queue=%d, pub=%d", q.AccountID, p.AccountID)
	}
	if q.TokenID != p.TokenID {
		return errors.Wrapf(types.ErrInvalidParam, "fee tokenId queue=%d, pub=%d", q.TokenID, p.TokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "fee amount queue=%s, pub=%s", q.Amount, p.Amount)
	}

	return nil
}

func checkSameMintNFT(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetMintNFT()
	p := pubDataOp.Op.GetMintNFT()
	if q.MintAcctID != p.MintAcctID {
		return errors.Wrapf(types.ErrInvalidParam, "mintNFT mint acctId queue=%d, pub=%d", q.MintAcctID, p.MintAcctID)
	}
	if q.RecipientID != p.RecipientID {
		return errors.Wrapf(types.ErrInvalidParam, "mintNFT recv acctId queue=%d, pub=%d", q.RecipientID, p.RecipientID)
	}
	if q.ErcProtocol != p.ErcProtocol {
		return errors.Wrapf(types.ErrInvalidParam, "mintNFT protocal queue=%d, pub=%d", q.ErcProtocol, p.ErcProtocol)
	}
	for i, _ := range q.ContentHash {
		if q.ContentHash[i] != p.ContentHash[i] {
			return errors.Wrapf(types.ErrInvalidParam, "mintNFT contentHash i=%d queue=%s, pub=%s", i, q.ContentHash[i], p.ContentHash[i])
		}
	}

	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "mintNFT amount queue=%d, pub=%d", q.Amount, p.Amount)
	}
	if q.NewNFTTokenID != p.NewNFTTokenID {
		return errors.Wrapf(types.ErrInvalidParam, "mintNFT newTokenId queue=%d, pub=%d", q.NewNFTTokenID, p.NewNFTTokenID)
	}
	if q.CreateSerialID != p.CreateSerialID {
		return errors.Wrapf(types.ErrInvalidParam, "mintNFT serialId queue=%d, pub=%d", q.CreateSerialID, p.CreateSerialID)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "mintNFT fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameWithdrawNFT(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetWithdrawNFT()
	p := pubDataOp.Op.GetWithdrawNFT()
	if q.FromAcctID != p.FromAcctID {
		return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT from acctId queue=%d, pub=%d", q.FromAcctID, p.FromAcctID)
	}
	if q.NFTTokenID != p.NFTTokenID {
		return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT token queue=%d, pub=%d", q.NFTTokenID, p.NFTTokenID)
	}
	if q.WithdrawAmount != p.WithdrawAmount {
		return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT withdraw amount queue=%d, pub=%d", q.WithdrawAmount, p.WithdrawAmount)
	}
	if q.ErcProtocol != p.ErcProtocol {
		return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT protocal queue=%d, pub=%d", q.ErcProtocol, p.ErcProtocol)
	}
	for i, _ := range q.ContentHash {
		if q.ContentHash[i] != p.ContentHash[i] {
			return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT contentHash i=%d queue=%s, pub=%s", i, q.ContentHash[i], p.ContentHash[i])
		}
	}

	if q.CreatorAcctID != p.CreatorAcctID {
		return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT create id queue=%d, pub=%d", q.CreatorAcctID, p.CreatorAcctID)
	}
	if q.InitMintAmount != p.InitMintAmount {
		return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT init amount queue=%d, pub=%d", q.InitMintAmount, p.InitMintAmount)
	}
	if q.EthAddress != p.EthAddress {
		return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT ethAddr queue=%s, pub=%s", q.EthAddress, p.EthAddress)
	}
	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "withdrawNFT fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkSameTransferNFT(queueOp, pubDataOp *zt.ZkOperation) error {
	q := queueOp.Op.GetTransferNFT()
	p := pubDataOp.Op.GetTransferNFT()
	if q.FromAccountID != p.FromAccountID {
		return errors.Wrapf(types.ErrInvalidParam, "transferNFT from acctId queue=%d, pub=%d", q.FromAccountID, p.FromAccountID)
	}
	if q.RecipientID != p.RecipientID {
		return errors.Wrapf(types.ErrInvalidParam, "transferNFT recv acctId queue=%d, pub=%d", q.RecipientID, p.RecipientID)
	}
	if q.NFTTokenID != p.NFTTokenID {
		return errors.Wrapf(types.ErrInvalidParam, "transferNFT tokenId queue=%d, pub=%d", q.NFTTokenID, p.NFTTokenID)
	}
	if q.Amount != p.Amount {
		return errors.Wrapf(types.ErrInvalidParam, "mintNFT amount queue=%d, pub=%d", q.Amount, p.Amount)
	}

	if q.Fee.Fee != p.Fee.Fee {
		return errors.Wrapf(types.ErrInvalidParam, "contracmintNFTt2treeNew fee queue=%s, pub=%s", q.Fee.Fee, p.Fee.Fee)
	}
	return nil
}

func checkPackValue(amount string, manMaxBitWidth int64) error {
	amountInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return errors.Wrapf(types.ErrInvalidParam, "checkPackValue amount=%s", amount)
	}
	if amountInt.Cmp(big.NewInt(0)) == 0 {
		return nil
	}
	//exp部分默认最大是31，不需要检查
	man, _ := zt.ZkTransferManExpPart(amount)
	manV, ok := new(big.Int).SetString(man, 10)
	if !ok {
		return errors.Wrapf(types.ErrInvalidParam, "transferManExpPart,man=%s,amount=%s", manV, amount)
	}

	//最大mantissa部分的值 比如2^35, amount的非0部分的值不能超过此值，超过的话，可以分多次
	maxManV := new(big.Int).Exp(big.NewInt(2), big.NewInt(manMaxBitWidth), nil)
	//manv <= maxMan
	if maxManV.Cmp(manV) < 0 {
		return errors.Wrapf(types.ErrNotAllow, "amount's mant part=%s big than %s(2^%d)", man, maxManV.String(), manMaxBitWidth)
	}
	return nil
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
	amountInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", errors.Wrapf(types.ErrInvalidParam, "TransferDecimalAmount amount=%s", amount)
	}
	//防止产生amount="0"时候扩充到"0000"字串
	if amountInt.Cmp(big.NewInt(0)) == 0 {
		return "0", nil
	}
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
