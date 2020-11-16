package state

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	jvmTypes "github.com/33cn/plugin/plugin/dapp/jvm/types"
)

var (
	IsPara bool
	Title  string
)

// MemoryStateDB 内存状态数据库，保存在区块操作时内部的数据变更操作
// 本数据库不会直接写文件，只会暂存变更记录
// 在区块打包完成后，这里的缓存变更数据会被清空（通过区块打包分别写入blockchain和statedb数据库）
// 在交易执行过程中，本数据库会暂存并变更，在交易执行结束后，会返回变更的数据集，返回给blockchain
// 执行器的Exec阶段会返回：交易收据、合约账户（包含合约地址、合约代码、合约存储信息）
// 执行器的ExecLocal阶段会返回：合约创建人和合约的关联信息
type MemoryStateDB struct {
	//current executor's name, could be Jvm, user.p.xxx.user.Jvm.xxx or usr.Jvm.xxx
	ExecutorName string
	// 状态DB，从执行器框架传入
	StateDB db.KV

	// 本地DB，从执行器框架传入
	LocalDB db.KVDB

	// Coins账户操作对象，从执行器框架传入
	CoinsAccount *account.DB

	// 缓存账户对象
	accounts map[string]*ContractAccount

	// 存储makeLogN指令对应的日志数据
	logs    map[common.Hash][]*jvmTypes.ContractLog
	logSize uint

	// 版本号，用于标识数据变更版本
	snapshots  []*Snapshot
	currentVer *Snapshot
	versionID  int

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
func NewMemoryStateDB(executorName string, StateDB db.KV, LocalDB db.KVDB, CoinsAccount *account.DB, blockHeight int64) *MemoryStateDB {
	mdb := &MemoryStateDB{
		ExecutorName: executorName,
		StateDB:      StateDB,
		LocalDB:      LocalDB,
		CoinsAccount: CoinsAccount,
		accounts:     make(map[string]*ContractAccount),
		logs:         make(map[common.Hash][]*jvmTypes.ContractLog),
		logSize:      0,
		versionID:    0,
		stateDirty:   make(map[string]interface{}),
		dataDirty:    make(map[string]interface{}),
		blockHeight:  blockHeight,
		txIndex:      0,
	}
	return mdb
}

// CreateAccount 创建一个新的合约账户对象
func (m *MemoryStateDB) CreateAccount(addr, creator string, name string) {
	acc := m.GetAccount(addr)
	if acc == nil {
		// 这种情况下为新增合约账户
		acc := NewContractAccount(addr, m)
		acc.SetCreator(creator)
		acc.SetExecName(name)
		m.accounts[addr] = acc
		m.addChange(createAccountChange{baseChange: baseChange{}, account: addr})
	}
}

func (m *MemoryStateDB) addChange(entry DataChange) {
	if m.currentVer != nil {
		m.currentVer.append(entry)
	}
}

// GetBalance 这里需要区分对待，如果是合约账户，则查看合约账户所有者地址在此合约下的余额；
// 如果是外部账户，则直接返回外部账户的余额
func (m *MemoryStateDB) GetBalance(addr string) uint64 {
	if m.CoinsAccount == nil {
		return 0
	}
	isExec := m.Exist(addr)
	var ac *types.Account
	if isExec {
		contract := m.GetAccount(addr)
		if contract == nil {
			return 0
		}
		creator := contract.GetCreator()
		if len(creator) == 0 {
			return 0
		}
		ac = m.CoinsAccount.LoadExecAccount(creator, addr)
	} else {
		ac = m.CoinsAccount.LoadAccount(addr)
	}
	if ac != nil {
		return uint64(ac.Balance)
	}
	return 0
}

// GetCode 获取合约的code
func (m *MemoryStateDB) GetCode(addr string) []byte {
	acc := m.GetAccount(addr)
	if acc != nil {
		return acc.Data.GetCode()
	}
	return nil
}

// GetName 获取设备的名称
func (m *MemoryStateDB) GetName(addr string) string {
	acc := m.GetAccount(addr)
	if acc != nil {
		return acc.Data.GetName()
	}
	return ""
}

// SetCodeAndAbi 设置code和abi 数据
func (m *MemoryStateDB) SetCodeAndAbi(addr string, code []byte, abi []byte) {
	acc := m.GetAccount(addr)
	if acc != nil {
		m.dataDirty[addr] = true
		acc.SetCodeAndAbi(code, abi)
	}
}

// GetAccount 从缓存中获取或加载合约账户
func (m *MemoryStateDB) GetAccount(addr string) *ContractAccount {
	if acc, ok := m.accounts[addr]; ok {
		return acc
	}
	// 需要加载合约对象，根据是否存在合约代码来判断是否有合约对象
	contract := NewContractAccount(addr, m)
	contract.LoadContract(m.StateDB)
	if contract.Empty() {
		log15.Error("StateDBGetState GetAccount:contract.Empty()")
		return nil
	}
	m.accounts[addr] = contract
	return contract

}

// List 根据前缀查询本地数据库
func (m *MemoryStateDB) List(prefix []byte, maxCountOnce int32) [][]byte {
	count := m.LocalDB.PrefixCount(prefix)
	log15.Debug("PrefixCount", "prefix", string(prefix), "count", count)

	values, err := m.LocalDB.List(prefix, nil, maxCountOnce, 0)
	if err != nil {
		return nil
	}
	return values
}

// GetValueFromLocal 从本地数据库获取值
func (m *MemoryStateDB) GetValueFromLocal(addr, key, txhash string) []byte {
	// 先从合约缓存中获取
	acc := m.GetAccount(addr)
	if acc == nil {
		return nil
	}
	localkey := []byte(acc.GetLocalDataKey(addr, key))

	//交易执行过程中的查询,优先在本地缓存中进行查询,而普通的查询则直接在历史数据库中进行查询
	if "" != txhash {
		value, _ := getLocalValue(localkey, txhash)
		if value != nil {
			return value
		}
	}
	//如果在本地缓存中没有找到，再继续在历史数据库中进行查询
	value, err := m.LocalDB.Get(localkey)
	if err != nil {
		log15.Debug("GetValueFromLocal failed", "key", key, "err", err)
		return nil
	}
	return value
}

// SetValue2Local 设置数据存储到本地
func (m *MemoryStateDB) SetValue2Local(addr, key string, value []byte, txHash string) bool {
	acc := m.GetAccount(addr)
	if acc == nil {
		return jvmTypes.AccountOpFail
	}

	if nil != acc.SetValue2Local(key, value, txHash) {
		return jvmTypes.AccountOpFail
	}
	return jvmTypes.AccountOpSuccess
}

// GetState SLOAD 指令加载合约状态数据
func (m *MemoryStateDB) GetState(addr string, key string) []byte {
	// 先从合约缓存中获取
	acc := m.GetAccount(addr)
	if acc != nil {
		return acc.GetState(key)
	}
	log15.Error("StateDBGetState jvm query via GetState acc is nil:")
	return nil
}

// SetState SSTORE 指令修改合约状态数据
func (m *MemoryStateDB) SetState(addr, key string, value []byte) bool {
	acc := m.GetAccount(addr)
	if acc != nil {
		if nil != acc.SetState(key, value) {
			return jvmTypes.AccountOpFail
		}
		return jvmTypes.AccountOpSuccess
	}
	return jvmTypes.AccountOpFail
}

// Exist 判断合约对象是否存在
func (m *MemoryStateDB) Exist(addr string) bool {
	return m.GetAccount(addr) != nil
}

// Empty 判断合约对象是否为空
func (m *MemoryStateDB) Empty(addr string) bool {
	acc := m.GetAccount(addr)

	// 如果包含合约代码，则不为空
	if acc != nil && !acc.Empty() {
		return false
	}

	// 账户有余额，也不为空
	if m.GetBalance(addr) != 0 {
		return false
	}
	return true
}

// Snapshot 对当前的数据状态打快照，并生成快照版本号，方便后面回滚数据
func (m *MemoryStateDB) Snapshot() int {
	id := m.versionID
	m.versionID++
	m.currentVer = &Snapshot{id: id, statedb: m}
	m.snapshots = append(m.snapshots, m.currentVer)
	return id
}

// GetLastSnapshot 获取最后一次成功的快照版本号
func (m *MemoryStateDB) GetLastSnapshot() *Snapshot {
	if m.versionID == 0 {
		return nil
	}
	return m.snapshots[m.versionID-1]
}

// GetReceiptLogs 获取合约对象的变更日志
func (m *MemoryStateDB) GetReceiptLogs(addr string) (logs []*types.ReceiptLog) {
	acc := m.GetAccount(addr)
	if acc != nil {
		if m.dataDirty[addr] != nil {
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
func (m *MemoryStateDB) GetChangedData(version int, opType jvmTypes.JvmContratOpType) (kvSet []*types.KeyValue, logs []*types.ReceiptLog) {
	if version < 0 {
		return
	}

	for _, snapshot := range m.snapshots {
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

// PrintLogs 本合约执行完毕之后打印合约生成的日志（如果有）
// 这里不保证当前区块可以打包成功，只是在执行区块中的交易时，如果交易执行成功，就会打印合约日志
func (m *MemoryStateDB) PrintLogs() {
	items := m.logs[m.txHash]
	for _, item := range items {
		item.PrintLog()
	}
}

// SetCurrentExecutorName 设置当前执行器的名称
func (m *MemoryStateDB) SetCurrentExecutorName(executorName string) {
	m.ExecutorName = executorName
}

// ExecFrozen exec frozen information
func (m *MemoryStateDB) ExecFrozen(tx *types.Transaction, addr string, amount int64) bool {
	if nil == tx {
		log15.Error("ExecFrozen get nil tx")
		return jvmTypes.AccountOpFail
	}

	execaddr := address.ExecAddress(string(tx.Execer))
	ret, err := m.CoinsAccount.ExecFrozen(addr, execaddr, amount)
	if err != nil {
		log15.Error("ExecFrozen error", "addr", addr, "execaddr", execaddr, "amount", amount, "err info", err)
		return jvmTypes.AccountOpFail
	}

	m.addChange(balanceChange{
		baseChange: baseChange{},
		amount:     amount,
		data:       ret.KV,
		logs:       ret.Logs,
	})

	return jvmTypes.AccountOpSuccess
}

// ExecActive active exec
func (m *MemoryStateDB) ExecActive(tx *types.Transaction, addr string, amount int64) bool {
	if nil == tx {
		log15.Error("ExecActive get nil tx")
		return jvmTypes.AccountOpFail
	}
	execaddr := address.ExecAddress(string(tx.Execer))
	ret, err := m.CoinsAccount.ExecActive(addr, execaddr, amount)
	if err != nil {
		log15.Error("ExecActive error", "addr", addr, "execaddr", execaddr, "amount", amount, "err info", err)
		return jvmTypes.AccountOpFail
	}

	m.addChange(balanceChange{
		baseChange: baseChange{},
		amount:     amount,
		data:       ret.KV,
		logs:       ret.Logs,
	})

	return jvmTypes.AccountOpSuccess

}

// ExecTransfer transfer exec
func (m *MemoryStateDB) ExecTransfer(tx *types.Transaction, from, to string, amount int64) bool {
	if nil == tx {
		log15.Error("ExecTransfer get nil tx")
		return jvmTypes.AccountOpFail
	}

	execaddr := address.ExecAddress(string(tx.Execer))
	ret, err := m.CoinsAccount.ExecTransfer(from, to, execaddr, amount)
	if err != nil {
		log15.Error("ExecFrozen error", "from", from, "to", to, "execaddr", execaddr, "amount", amount, "err info", err)
		return jvmTypes.AccountOpFail
	}

	m.addChange(balanceChange{
		baseChange: baseChange{},
		amount:     amount,
		data:       ret.KV,
		logs:       ret.Logs,
	})

	return jvmTypes.AccountOpSuccess
}

// ExecTransferFrozen transfer frozen exec
func (m *MemoryStateDB) ExecTransferFrozen(tx *types.Transaction, from, to string, amount int64) bool {
	if nil == tx {
		log15.Error("ExecTransferFrozen get nil tx")
		return jvmTypes.AccountOpFail
	}
	execaddr := address.ExecAddress(string(tx.Execer))
	ret, err := m.CoinsAccount.ExecTransferFrozen(from, to, execaddr, amount)
	if err != nil {
		log15.Error("ExecFrozen error", "from", from, "to", to, "execaddr", execaddr, "amount", amount, "err info", err)
		return jvmTypes.AccountOpFail
	}

	m.addChange(balanceChange{
		baseChange: baseChange{},
		amount:     amount,
		data:       ret.KV,
		logs:       ret.Logs,
	})

	return jvmTypes.AccountOpSuccess
}
