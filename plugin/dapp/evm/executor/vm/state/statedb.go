// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// MemoryStateDB 内存状态数据库，保存在区块操作时内部的数据变更操作
// 本数据库不会直接写文件，只会暂存变更记录
// 在区块打包完成后，这里的缓存变更数据会被清空（通过区块打包分别写入blockchain和statedb数据库）
// 在交易执行过程中，本数据库会暂存并变更，在交易执行结束后，会返回变更的数据集，返回给blockchain
// 执行器的Exec阶段会返回：交易收据、合约账户（包含合约地址、合约代码、合约存储信息）
// 执行器的ExecLocal阶段会返回：合约创建人和合约的关联信息
type MemoryStateDB struct {
	// StateDB 状态DB，从执行器框架传入
	StateDB db.KV

	// LocalDB 本地DB，从执行器框架传入
	LocalDB db.KVDB

	// CoinsAccount Coins账户操作对象，从执行器框架传入
	CoinsAccount *account.DB

	// 缓存账户对象
	accounts map[string]*ContractAccount

	// 合约执行过程中退回的资金
	refund uint64

	// 存储makeLogN指令对应的日志数据
	logs    map[common.Hash][]*model.ContractLog
	logSize uint

	// 版本号，用于标识数据变更版本
	snapshots  []*Snapshot
	currentVer *Snapshot
	versionID  int

	// 存储sha3指令对应的数据，仅用于debug日志
	preimages map[common.Hash][]byte

	// 当前临时交易哈希和交易序号
	txHash  common.Hash
	txIndex int

	// 当前区块高度
	blockHeight int64

	// 用户保存合约账户的状态数据或合约代码数据有没有发生变更
	stateDirty map[string]interface{}
	dataDirty  map[string]interface{}
}

// NewMemoryStateDB 基于执行器框架的三个DB构建内存状态机对象
// 此对象的生命周期对应一个区块，在同一个区块内的多个交易执行时共享同一个DB对象
// 开始执行下一个区块时（执行器框架调用setEnv设置的区块高度发生变更时），会重新创建此DB对象
func NewMemoryStateDB(StateDB db.KV, LocalDB db.KVDB, CoinsAccount *account.DB, blockHeight int64) *MemoryStateDB {
	mdb := &MemoryStateDB{
		StateDB:      StateDB,
		LocalDB:      LocalDB,
		CoinsAccount: CoinsAccount,
		accounts:     make(map[string]*ContractAccount),
		logs:         make(map[common.Hash][]*model.ContractLog),
		preimages:    make(map[common.Hash][]byte),
		stateDirty:   make(map[string]interface{}),
		dataDirty:    make(map[string]interface{}),
		blockHeight:  blockHeight,
		refund:       0,
		txIndex:      0,
	}
	return mdb
}

// Prepare 每一个交易执行之前调用此方法，设置此交易的上下文信息
// 目前的上下文中包含交易哈希以及交易在区块中的序号
func (mdb *MemoryStateDB) Prepare(txHash common.Hash, txIndex int) {
	mdb.txHash = txHash
	mdb.txIndex = txIndex
}

// CreateAccount 创建一个新的合约账户对象
func (mdb *MemoryStateDB) CreateAccount(addr, creator string, execName, alias string) {
	acc := mdb.GetAccount(addr)
	if acc == nil {
		// 这种情况下为新增合约账户
		acc := NewContractAccount(addr, mdb)
		acc.SetCreator(creator)
		acc.SetExecName(execName)
		acc.SetAliasName(alias)
		mdb.accounts[addr] = acc
		mdb.addChange(createAccountChange{baseChange: baseChange{}, account: addr})
	}
}

func (mdb *MemoryStateDB) addChange(entry DataChange) {
	if mdb.currentVer != nil {
		mdb.currentVer.append(entry)
	}
}

// SubBalance 从外部账户地址扣钱（钱其实是打到合约账户中的）
func (mdb *MemoryStateDB) SubBalance(addr, caddr string, value uint64) {
	res := mdb.Transfer(addr, caddr, value)
	log15.Debug("transfer result", "from", addr, "to", caddr, "amount", value, "result", res)
}

// AddBalance 向外部账户地址打钱（钱其实是外部账户之前打到合约账户中的）
func (mdb *MemoryStateDB) AddBalance(addr, caddr string, value uint64) {
	res := mdb.Transfer(caddr, addr, value)
	log15.Debug("transfer result", "from", addr, "to", caddr, "amount", value, "result", res)
}

// GetBalance 这里需要区分对待，如果是合约账户，则查看合约账户所有者地址在此合约下的余额；
// 如果是外部账户，则直接返回外部账户的余额
func (mdb *MemoryStateDB) GetBalance(addr string) uint64 {
	if mdb.CoinsAccount == nil {
		return 0
	}
	isExec := mdb.Exist(addr)
	var ac *types.Account
	if isExec {
		if types.IsDappFork(mdb.GetBlockHeight(), "evm", evmtypes.ForkEVMFrozen) {
			ac = mdb.CoinsAccount.LoadExecAccount(addr, addr)
		} else {
			contract := mdb.GetAccount(addr)
			if contract == nil {
				return 0
			}
			creator := contract.GetCreator()
			if len(creator) == 0 {
				return 0
			}
			ac = mdb.CoinsAccount.LoadExecAccount(creator, addr)
		}
	} else {
		ac = mdb.CoinsAccount.LoadAccount(addr)
	}
	if ac != nil {
		return uint64(ac.Balance)
	}
	return 0
}

// GetNonce 目前chain33中没有保留账户的nonce信息，这里临时添加到合约账户中；
// 所以，目前只有合约对象有nonce值
func (mdb *MemoryStateDB) GetNonce(addr string) uint64 {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		return acc.GetNonce()
	}
	return 0
}

// SetNonce 设置nonce值
func (mdb *MemoryStateDB) SetNonce(addr string, nonce uint64) {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		acc.SetNonce(nonce)
	}
}

// GetCodeHash 获取代码哈希
func (mdb *MemoryStateDB) GetCodeHash(addr string) common.Hash {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		return common.BytesToHash(acc.Data.GetCodeHash())
	}
	return common.Hash{}
}

// GetCode 获取代码内容
func (mdb *MemoryStateDB) GetCode(addr string) []byte {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		return acc.Data.GetCode()
	}
	return nil
}

// SetCode 设置代码内容
func (mdb *MemoryStateDB) SetCode(addr string, code []byte) {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		mdb.dataDirty[addr] = true
		acc.SetCode(code)
	}
}

// SetAbi 设置ABI内容
func (mdb *MemoryStateDB) SetAbi(addr, abi string) {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		mdb.dataDirty[addr] = true
		acc.SetAbi(abi)
	}
}

// GetAbi 获取ABI
func (mdb *MemoryStateDB) GetAbi(addr string) string {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		return acc.Data.GetAbi()
	}
	return ""
}

// GetCodeSize 获取合约代码自身的大小
// 对应 EXTCODESIZE 操作码
func (mdb *MemoryStateDB) GetCodeSize(addr string) int {
	code := mdb.GetCode(addr)
	if code != nil {
		return len(code)
	}
	return 0
}

// AddRefund 合约自杀或SSTORE指令时，返还Gas
func (mdb *MemoryStateDB) AddRefund(gas uint64) {
	mdb.addChange(refundChange{baseChange: baseChange{}, prev: mdb.refund})
	mdb.refund += gas
}

// GetRefund 获取奖励
func (mdb *MemoryStateDB) GetRefund() uint64 {
	return mdb.refund
}

// GetAccount 从缓存中获取或加载合约账户
func (mdb *MemoryStateDB) GetAccount(addr string) *ContractAccount {
	if acc, ok := mdb.accounts[addr]; ok {
		return acc
	}
	// 需要加载合约对象，根据是否存在合约代码来判断是否有合约对象
	contract := NewContractAccount(addr, mdb)
	contract.LoadContract(mdb.StateDB)
	if contract.Empty() {
		return nil
	}
	mdb.accounts[addr] = contract
	return contract
}

// GetState SLOAD 指令加载合约状态数据
func (mdb *MemoryStateDB) GetState(addr string, key common.Hash) common.Hash {
	// 先从合约缓存中获取
	acc := mdb.GetAccount(addr)
	if acc != nil {
		return acc.GetState(key)
	}
	return common.Hash{}
}

// SetState SSTORE 指令修改合约状态数据
func (mdb *MemoryStateDB) SetState(addr string, key common.Hash, value common.Hash) {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		acc.SetState(key, value)
		// 新的分叉中状态数据变更不需要单独进行标识
		if !types.IsDappFork(mdb.blockHeight, "evm", evmtypes.ForkEVMState) {
			mdb.stateDirty[addr] = true
		}
	}
}

// TransferStateData 转换合约状态数据存储
func (mdb *MemoryStateDB) TransferStateData(addr string) {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		acc.TransferState()
	}
}

// UpdateState 表示合约地址的状态数据发生了变更，需要进行更新
func (mdb *MemoryStateDB) UpdateState(addr string) {
	mdb.stateDirty[addr] = true
}

// Suicide SELFDESTRUCT 合约对象自杀
// 合约自杀后，合约对象依然存在，只是无法被调用，也无法恢复
func (mdb *MemoryStateDB) Suicide(addr string) bool {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		mdb.addChange(suicideChange{
			baseChange: baseChange{},
			account:    addr,
			prev:       acc.State.GetSuicided(),
		})
		mdb.stateDirty[addr] = true
		return acc.Suicide()
	}
	return false
}

// HasSuicided 判断此合约对象是否已经自杀
// 自杀的合约对象是不允许调用的
func (mdb *MemoryStateDB) HasSuicided(addr string) bool {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		return acc.HasSuicided()
	}
	return false
}

// Exist 判断合约对象是否存在
func (mdb *MemoryStateDB) Exist(addr string) bool {
	return mdb.GetAccount(addr) != nil
}

// Empty 判断合约对象是否为空
func (mdb *MemoryStateDB) Empty(addr string) bool {
	acc := mdb.GetAccount(addr)

	// 如果包含合约代码，则不为空
	if acc != nil && !acc.Empty() {
		return false
	}

	// 账户有余额，也不为空
	if mdb.GetBalance(addr) != 0 {
		return false
	}
	return true
}

// RevertToSnapshot 将数据状态回滚到指定快照版本（中间的版本数据将会被删除）
func (mdb *MemoryStateDB) RevertToSnapshot(version int) {
	if version >= len(mdb.snapshots) {
		return
	}

	ver := mdb.snapshots[version]

	// 如果版本号不对，回滚失败
	if ver == nil || ver.id != version {
		log15.Crit(fmt.Errorf("Snapshot id %v cannot be reverted", version).Error())
		return
	}

	// 从最近版本开始回滚
	for index := len(mdb.snapshots) - 1; index >= version; index-- {
		mdb.snapshots[index].revert()
	}

	// 只保留回滚版本之前的版本数据
	mdb.snapshots = mdb.snapshots[:version]
	mdb.versionID = version
	if version == 0 {
		mdb.currentVer = nil
	} else {
		mdb.currentVer = mdb.snapshots[version-1]
	}

}

// Snapshot 对当前的数据状态打快照，并生成快照版本号，方便后面回滚数据
func (mdb *MemoryStateDB) Snapshot() int {
	id := mdb.versionID
	mdb.versionID++
	mdb.currentVer = &Snapshot{id: id, statedb: mdb}
	mdb.snapshots = append(mdb.snapshots, mdb.currentVer)
	return id
}

// GetLastSnapshot 获取最后一次成功的快照版本号
func (mdb *MemoryStateDB) GetLastSnapshot() *Snapshot {
	if mdb.versionID == 0 {
		return nil
	}
	return mdb.snapshots[mdb.versionID-1]
}

// GetReceiptLogs 获取合约对象的变更日志
func (mdb *MemoryStateDB) GetReceiptLogs(addr string) (logs []*types.ReceiptLog) {
	acc := mdb.GetAccount(addr)
	if acc != nil {
		if mdb.stateDirty[addr] != nil {
			stateLog := acc.BuildStateLog()
			if stateLog != nil {
				logs = append(logs, stateLog)
			}
		}

		if mdb.dataDirty[addr] != nil {
			logs = append(logs, acc.BuildDataLog())
		}
		return
	}
	return
}

// GetChangedData 获取本次操作所引起的状态数据变更
// 因为目前执行器每次执行都是一个新的MemoryStateDB，所以，所有的快照都是从0开始的，
// 这里获取的应该是从0到目前快照的所有变更；
// 另外，因为合约内部会调用其它合约，也会产生数据变更，所以这里返回的数据，不止是一个合约的数据。
func (mdb *MemoryStateDB) GetChangedData(version int) (kvSet []*types.KeyValue, logs []*types.ReceiptLog) {
	if version < 0 {
		return
	}

	for _, snapshot := range mdb.snapshots {
		kv, log := snapshot.getData()
		if kv != nil {
			kvSet = append(kvSet, kv...)
		}

		if log != nil {
			logs = append(logs, log...)
		}
	}
	return
}

// CanTransfer 借助coins执行器进行转账相关操作
func (mdb *MemoryStateDB) CanTransfer(sender, recipient string, amount uint64) bool {

	log15.Debug("check CanTransfer", "sender", sender, "recipient", recipient, "amount", amount)

	tType, errInfo := mdb.checkTransfer(sender, recipient, amount)

	if errInfo != nil {
		log15.Error("check transfer error", "sender", sender, "recipient", recipient, "amount", amount, "err info", errInfo)
		return false
	}

	value := int64(amount)
	if value < 0 {
		return false
	}

	switch tType {
	case NoNeed:
		return true
	case ToExec:
		// 无论其它账户还是创建者向合约地址转账，都需要检查其当前合约账户活动余额是否充足
		accFrom := mdb.CoinsAccount.LoadExecAccount(sender, recipient)
		b := accFrom.GetBalance() - value
		if b < 0 {
			log15.Error("check transfer error", "error info", types.ErrNoBalance)
			return false
		}
		return true
	case FromExec:
		return mdb.checkExecAccount(sender, value)
	default:
		return false
	}
}
func (mdb *MemoryStateDB) checkExecAccount(execAddr string, value int64) bool {
	var err error
	defer func() {
		if err != nil {
			log15.Error("checkExecAccount error", "error info", err)
		}
	}()
	// 如果是合约地址，则需要判断创建者在本合约中的余额是否充足
	if !types.CheckAmount(value) {
		err = types.ErrAmount
		return false
	}
	contract := mdb.GetAccount(execAddr)
	if contract == nil {
		err = model.ErrAddrNotExists
		return false
	}
	creator := contract.GetCreator()
	if len(creator) == 0 {
		err = model.ErrNoCreator
		return false
	}

	var accFrom *types.Account
	if types.IsDappFork(mdb.GetBlockHeight(), "evm", evmtypes.ForkEVMFrozen) {
		// 分叉后，需要检查合约地址下的金额是否足够
		accFrom = mdb.CoinsAccount.LoadExecAccount(execAddr, execAddr)
	} else {
		accFrom = mdb.CoinsAccount.LoadExecAccount(creator, execAddr)
	}
	balance := accFrom.GetBalance()
	remain := balance - value
	if remain < 0 {
		err = types.ErrNoBalance
		return false
	}
	return true
}

// TransferType 定义转账类型
type TransferType int

const (
	_ TransferType = iota
	// NoNeed 无需转账
	NoNeed
	// ToExec 向合约转账
	ToExec
	// FromExec 从合约转入
	FromExec
	// Error 处理出错
	Error
)

func (mdb *MemoryStateDB) checkTransfer(sender, recipient string, amount uint64) (tType TransferType, err error) {
	if amount == 0 {
		return NoNeed, nil
	}
	if mdb.CoinsAccount == nil {
		log15.Error("no coinsaccount exists", "sender", sender, "recipient", recipient, "amount", amount)
		return Error, model.ErrNoCoinsAccount
	}

	// 首先需要检查转账双方的信息，是属于合约账户还是外部账户
	execSender := mdb.Exist(sender)
	execRecipient := mdb.Exist(recipient)

	if execRecipient && execSender {
		// 双方均为合约账户，不支持
		err = model.ErrTransferBetweenContracts
		tType = Error
	} else if execSender {
		// 从合约账户到外部账户转账 （这里调用外部账户从合约账户取钱接口）
		tType = FromExec
		err = nil
	} else if execRecipient {
		// 从外部账户到合约账户转账
		tType = ToExec
		err = nil
	} else {
		// 双方都是外部账户，不支持
		err = model.ErrTransferBetweenEOA
		tType = Error
	}

	return tType, err
}

// Transfer 借助coins执行器进行转账相关操作
// 只支持 合约账户到合约账户，其它情况不支持
func (mdb *MemoryStateDB) Transfer(sender, recipient string, amount uint64) bool {
	log15.Debug("transfer from contract to external(contract)", "sender", sender, "recipient", recipient, "amount", amount)

	tType, errInfo := mdb.checkTransfer(sender, recipient, amount)

	if errInfo != nil {
		log15.Error("transfer error", "sender", sender, "recipient", recipient, "amount", amount, "err info", errInfo)
		return false
	}

	var (
		ret *types.Receipt
		err error
	)

	value := int64(amount)
	if value < 0 {
		return false
	}

	switch tType {
	case NoNeed:
		return true
	case ToExec:
		ret, err = mdb.transfer2Contract(sender, recipient, value)
	case FromExec:
		ret, err = mdb.transfer2External(sender, recipient, value)
	default:
		return false
	}

	// 这种情况下转账失败并不进行处理，也不会从sender账户扣款，打印日志即可
	if err != nil {
		log15.Error("transfer error", "sender", sender, "recipient", recipient, "amount", amount, "err info", err)
		return false
	}
	if ret != nil {
		mdb.addChange(transferChange{
			baseChange: baseChange{},
			amount:     value,
			data:       ret.KV,
			logs:       ret.Logs,
		})
	}
	return true
}

// 因为chain33的限制，在执行器中转账只能在以下几个方向进行：
// A账户的X合约 <-> B账户的X合约；
// 其它情况不支持，所以要想实现EVM合约与账户之间的转账需要经过中转处理，比如A要向B创建的X合约转账，则执行以下流程：
// A -> A:X -> B:X；  (其中第一步需要外部手工执行)
// 本方法封装第二步转账逻辑;
func (mdb *MemoryStateDB) transfer2Contract(sender, recipient string, amount int64) (ret *types.Receipt, err error) {
	// 首先获取合约的创建者信息
	contract := mdb.GetAccount(recipient)
	if contract == nil {
		return nil, model.ErrAddrNotExists
	}
	creator := contract.GetCreator()
	if len(creator) == 0 {
		return nil, model.ErrNoCreator
	}
	execAddr := recipient

	ret = &types.Receipt{}

	if types.IsDappFork(mdb.GetBlockHeight(), "evm", evmtypes.ForkEVMFrozen) {
		// 用户向合约转账时，将钱转到合约地址下execAddr:execAddr
		rs, err := mdb.CoinsAccount.ExecTransfer(sender, execAddr, execAddr, amount)
		if err != nil {
			return nil, err
		}

		ret.KV = append(ret.KV, rs.KV...)
		ret.Logs = append(ret.Logs, rs.Logs...)
	} else {
		if strings.Compare(sender, creator) != 0 {
			// 用户向合约转账时，首先将钱转到创建者合约地址下
			rs, err := mdb.CoinsAccount.ExecTransfer(sender, creator, execAddr, amount)
			if err != nil {
				return nil, err
			}

			ret.KV = append(ret.KV, rs.KV...)
			ret.Logs = append(ret.Logs, rs.Logs...)
		}
	}

	return ret, nil
}

// chain33转账限制请参考方法 Transfer2Contract ；
// 本方法封装从合约账户到外部账户的转账逻辑；
func (mdb *MemoryStateDB) transfer2External(sender, recipient string, amount int64) (ret *types.Receipt, err error) {
	// 首先获取合约的创建者信息
	contract := mdb.GetAccount(sender)
	if contract == nil {
		return nil, model.ErrAddrNotExists
	}
	creator := contract.GetCreator()
	if len(creator) == 0 {
		return nil, model.ErrNoCreator
	}

	execAddr := sender

	if types.IsDappFork(mdb.GetBlockHeight(), "evm", evmtypes.ForkEVMFrozen) {
		// 合约向用户地址转账时，从合约地址下的钱中转出到用户合约地址
		ret, err = mdb.CoinsAccount.ExecTransfer(execAddr, recipient, execAddr, amount)
		if err != nil {
			return nil, err
		}
	} else {
		// 第一步先从创建者的合约账户到接受者的合约账户
		// 如果是自己调用自己创建的合约，这一步也可以省略
		if strings.Compare(creator, recipient) != 0 {
			ret, err = mdb.CoinsAccount.ExecTransfer(creator, recipient, execAddr, amount)
			if err != nil {
				return nil, err
			}
		}
	}
	return ret, nil
}

func (mdb *MemoryStateDB) mergeResult(one, two *types.Receipt) (ret *types.Receipt) {
	ret = one
	if ret == nil {
		ret = two
	} else if two != nil {
		ret.KV = append(ret.KV, two.KV...)
		ret.Logs = append(ret.Logs, two.Logs...)
	}
	return
}

// AddLog LOG0-4 指令对应的具体操作
// 生成对应的日志信息，目前这些生成的日志信息会在合约执行后打印到日志文件中
func (mdb *MemoryStateDB) AddLog(log *model.ContractLog) {
	mdb.addChange(addLogChange{txhash: mdb.txHash})
	log.TxHash = mdb.txHash
	log.Index = int(mdb.logSize)
	mdb.logs[mdb.txHash] = append(mdb.logs[mdb.txHash], log)
	mdb.logSize++
}

// AddPreimage 存储sha3指令对应的数据
func (mdb *MemoryStateDB) AddPreimage(hash common.Hash, data []byte) {
	// 目前只用于打印日志
	if _, ok := mdb.preimages[hash]; !ok {
		mdb.addChange(addPreimageChange{hash: hash})
		pi := make([]byte, len(data))
		copy(pi, data)
		mdb.preimages[hash] = pi
	}
}

// PrintLogs 本合约执行完毕之后打印合约生成的日志（如果有）
// 这里不保证当前区块可以打包成功，只是在执行区块中的交易时，如果交易执行成功，就会打印合约日志
func (mdb *MemoryStateDB) PrintLogs() {
	items := mdb.logs[mdb.txHash]
	for _, item := range items {
		item.PrintLog()
	}
}

// WritePreimages 打印本区块内生成的preimages日志
func (mdb *MemoryStateDB) WritePreimages(number int64) {
	for k, v := range mdb.preimages {
		log15.Debug("Contract preimages ", "key:", k.Str(), "value:", common.Bytes2Hex(v), "block height:", number)
	}
}

// ResetDatas 测试用，清空版本数据
func (mdb *MemoryStateDB) ResetDatas() {
	mdb.currentVer = nil
	mdb.snapshots = mdb.snapshots[:0]
}

// GetBlockHeight 返回当前区块高度
func (mdb *MemoryStateDB) GetBlockHeight() int64 {
	return mdb.blockHeight
}
