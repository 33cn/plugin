package spot

import (
	"fmt"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

var (
	// mavl-zkspot-dex-   资金帐号
	// mavl-zkspot-spot-  现货帐号
	// 先都用现货帐号
	spotDexName       = "spot"
	spotFeeAccountKey = []byte("zkspot-spotfeeaccount") // mavl-manager-{here}
)

func LoadSpotAccount(addr string, id uint64, statedb dbm.KV, p et.DBprefix) (*DexAccount, error) {
	return newAccountRepo(spotDexName, statedb, p).LoadAccount(addr, id)
}

func newAccountRepo(dexName string, statedb dbm.KV, p et.DBprefix) *accountRepo {
	return &accountRepo{
		statedb:  statedb,
		dbprefix: p,
		dexName:  dexName,
	}
}

type accountRepo struct {
	dexName  string
	statedb  dbm.KV
	dbprefix et.DBprefix
}

func (repo *accountRepo) genAccountKey(addr string, id uint64) []byte {
	return []byte(fmt.Sprintf("%s:%s:%d", repo.dbprefix.GetStatedbPrefix(), repo.dexName, id))
}

func (repo *accountRepo) LoadSpotAccount(addr string, id uint64) (*DexAccount, error) {
	return repo.LoadAccount(addr, id)
}

func (repo *accountRepo) LoadAccount(addr string, accID uint64) (*DexAccount, error) {
	var acc et.DexAccount
	key := repo.genAccountKey(addr, accID)
	v, err := repo.statedb.Get(key)
	if err == types.ErrNotFound {
		acc2 := emptyAccount(repo.dexName, accID, addr)
		return NewDexAccount(acc2, repo), nil
	}

	err = types.Decode(v, &acc)
	if err != nil {
		return nil, err
	}

	return NewDexAccount(&acc, repo), nil
}

type DexAccount struct {
	acc *et.DexAccount
	db  *accountRepo
}

func emptyAccount(ty string, id uint64, addr string) *et.DexAccount {
	return &et.DexAccount{
		Id:      id,
		Addr:    addr,
		DexName: ty,
	}
}

func NewDexAccount(acc *et.DexAccount, db *accountRepo) *DexAccount {
	return &DexAccount{acc: acc, db: db}
}

func (acc *DexAccount) findTokenIndex(tid uint64) int {
	for i, token := range acc.acc.Balance {
		if token.Id == tid {
			return i
		}
	}
	return -1
}

func (acc *DexAccount) newToken(tid uint64, amount uint64) int {
	acc.acc.Balance = append(acc.acc.Balance, &et.DexAccountBalance{
		Id:      tid,
		Balance: amount,
	})
	return len(acc.acc.Balance) - 1
}

func (acc *DexAccount) GetBalance(tid uint64) uint64 {
	return acc.getBalance(tid)
}

func (acc *DexAccount) getBalance(tid uint64) uint64 {
	idx := acc.findTokenIndex(tid)
	if idx == -1 {
		return 0
	}
	return acc.acc.Balance[idx].Balance
}

func (acc *DexAccount) GetFrozen(tid uint64) uint64 {
	return acc.getFrozen(tid)
}

func (acc *DexAccount) getFrozen(tid uint64) uint64 {
	idx := acc.findTokenIndex(tid)
	if idx == -1 {
		return 0
	}
	return acc.acc.Balance[idx].Frozen
}

func (acc *DexAccount) doMint(tid uint64, amount uint64) error {
	idx := acc.findTokenIndex(tid)
	if idx == -1 {
		acc.acc.Balance = append(acc.acc.Balance, &et.DexAccountBalance{
			Id:      tid,
			Balance: amount,
		})
	} else {
		acc.acc.Balance[idx].Balance += amount
	}
	return nil
}

func (acc *DexAccount) doBurn(tid uint64, amount uint64) error {
	idx := acc.findTokenIndex(tid)
	if idx == -1 {
		return et.ErrDexNotEnough
	}

	if acc.acc.Balance[idx].Balance < amount {
		return et.ErrDexNotEnough
	}

	acc.acc.Balance[idx].Balance -= amount
	return nil
}

func (acc *DexAccount) doFrozen(token uint64, amount uint64) error {
	idx := acc.findTokenIndex(token)
	if idx < 0 {
		return et.ErrDexNotEnough
	}
	if acc.acc.Balance[idx].Balance < amount {
		return et.ErrDexNotEnough
	}
	acc.acc.Balance[idx].Balance -= amount
	acc.acc.Balance[idx].Frozen += amount
	return nil
}

func (acc *DexAccount) doActive(token uint64, amount uint64) error {
	idx := acc.findTokenIndex(token)
	if idx < 0 {
		return et.ErrDexNotEnough
	}
	if acc.acc.Balance[idx].Frozen < amount {
		return et.ErrDexNotEnough
	}
	acc.acc.Balance[idx].Balance += amount
	acc.acc.Balance[idx].Frozen -= amount
	return nil
}

func dupAccount(acc *et.DexAccount) *et.DexAccount {
	copyAcc := et.DexAccount{
		Id:      acc.Id,
		Addr:    acc.Addr,
		DexName: acc.DexName,
	}
	var bs []*et.DexAccountBalance
	for _, b := range acc.Balance {
		copyB := et.DexAccountBalance{
			Id:      b.Id,
			Balance: b.Balance,
			Frozen:  b.Frozen,
		}
		bs = append(bs, &copyB)
	}
	copyAcc.Balance = bs
	return &copyAcc
}

// GetKVSet account to statdb kv
func (acc *DexAccount) GetKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(acc.acc)
	key := acc.db.genAccountKey(acc.acc.Addr, acc.acc.Id)
	acc.db.statedb.Set(key, value)

	kvset = make([]*types.KeyValue, 1)
	kvset[0] = &types.KeyValue{
		Key:   key,
		Value: value,
	}
	return kvset
}

func (acc *DexAccount) Frozen(token uint64, amount uint64) (*types.Receipt, error) {
	return acc.updateWithFunc(token, amount, acc.doFrozen, et.TyDexAccountFrozen)
}

func (acc *DexAccount) Active(token uint64, amount uint64) (*types.Receipt, error) {
	return acc.updateWithFunc(token, amount, acc.doActive, et.TyDexAccountActive)
}

func (acc *DexAccount) Mint(token uint64, amount uint64) (*types.Receipt, error) {
	return acc.updateWithFunc(token, amount, acc.doMint, et.TyDexAccountMint)
}

func (acc *DexAccount) Burn(token uint64, amount uint64) (*types.Receipt, error) {
	return acc.updateWithFunc(token, amount, acc.doBurn, et.TyDexAccountBurn)
}

type updateDexAccount func(token uint64, amount uint64) error
type updateDexAccount2 func(accTo *DexAccount, token uint64, amount uint64) error

func (acc *DexAccount) updateWithFunc(token uint64, amount uint64, f updateDexAccount, logType int32) (*types.Receipt, error) {
	copyAcc := dupAccount(acc.acc)
	err := f(token, amount)
	if err != nil {
		return nil, err
	}
	receiptlog := et.ReceiptDexAccount{
		Prev:    copyAcc,
		Current: acc.acc,
	}

	return acc.genReceipt(logType, acc, &receiptlog), nil
}

func (acc *DexAccount) genReceipt(ty int32, acc1 *DexAccount, r *et.ReceiptDexAccount) *types.Receipt {
	log1 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(r),
	}
	kv := acc.GetKVSet()
	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: []*types.ReceiptLog{log1},
	}
}

// two account operator

// 撮合 包含 1个交换, 和两个手续费
// 币的源头是是从balance/frozen 中转 看balance 的中值是否为frozen
// 币的目的一般到 balance即可, 如果有到frozen的 提供额外的函数或参数
/*
func (acc *dexAccount) Swap(accTo *dexAccount, got, gave *et.DexAccountBalance) error {
	err := acc.Tranfer(accTo, gave)
	if err != nil {
		return err
	}
	return acc.Withdraw(accTo, got)
}
*/

func (acc *DexAccount) doTranfer(accTo *DexAccount, token uint64, balance uint64) error {
	idx := acc.findTokenIndex(token)
	if idx < 0 {
		return et.ErrDexNotEnough
	}
	idxTo := accTo.findTokenIndex(token)
	if idxTo < 0 {
		idxTo = accTo.newToken(token, 0)
	}

	if acc.acc.Balance[idx].Balance < balance {
		return et.ErrDexNotEnough
	}

	acc.acc.Balance[idx].Balance -= balance
	accTo.acc.Balance[idxTo].Balance += balance

	return nil
}

func (acc *DexAccount) doFrozenTranfer(accTo *DexAccount, token uint64, amount uint64) error {
	idx := acc.findTokenIndex(token)
	if idx < 0 {
		return et.ErrDexNotEnough
	}
	idxTo := accTo.findTokenIndex(token)
	if idxTo < 0 {
		idxTo = accTo.newToken(token, 0)
	}

	if acc.acc.Balance[idx].Frozen < amount {
		return et.ErrDexNotEnough
	}

	acc.acc.Balance[idx].Frozen -= amount
	accTo.acc.Balance[idxTo].Balance += amount
	return nil
}

func (acc *DexAccount) FrozenTranfer(accTo *DexAccount, token uint64, amount uint64) (*types.Receipt, error) {
	return acc.updateWithFunc2(accTo, token, amount, acc.doFrozenTranfer, et.TyDexAccountTransferFrozen)
}

func (acc *DexAccount) Tranfer(accTo *DexAccount, token uint64, amount uint64) (*types.Receipt, error) {
	return acc.updateWithFunc2(accTo, token, amount, acc.doTranfer, et.TyDexAccountTransfer)
}

func (acc *DexAccount) updateWithFunc2(accTo *DexAccount, token uint64, amount uint64, f updateDexAccount2, logType int32) (*types.Receipt, error) {
	copyAcc := dupAccount(acc.acc)
	copyAccTo := dupAccount(accTo.acc)
	err := f(accTo, token, amount)
	if err != nil {
		return nil, err
	}
	receiptlog := et.ReceiptDexAccount{
		Prev:    copyAcc,
		Current: acc.acc,
	}
	receiptlog2 := et.ReceiptDexAccount{
		Prev:    copyAccTo,
		Current: accTo.acc,
	}

	return acc.genReceipt2(logType, accTo, &receiptlog, &receiptlog2), nil
}

func (acc *DexAccount) genReceipt2(ty int32, acc2 *DexAccount, r, r2 *et.ReceiptDexAccount) *types.Receipt {
	log1 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(r),
	}
	log2 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(r2),
	}

	kv := acc.GetKVSet()
	kv2 := acc2.GetKVSet()
	kv = append(kv, kv2...)
	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: []*types.ReceiptLog{log1, log2},
	}
}
