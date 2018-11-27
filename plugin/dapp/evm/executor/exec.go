// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"strings"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/runtime"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/state"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// Exec 本合约执行逻辑
func (evm *EVMExecutor) Exec(tx *types.Transaction, index int) (*types.Receipt, error) {
	evm.CheckInit()
	// 先转换消息
	msg, err := evm.GetMessage(tx)
	if err != nil {
		return nil, err
	}
	return evm.innerExec(msg, tx.Hash(), index, tx.Fee, false)
}

// 通用的EVM合约执行逻辑封装
// readOnly 是否只读调用，仅执行evm abi查询时为true
func (evm *EVMExecutor) innerExec(msg *common.Message, txHash []byte, index int, txFee int64, readOnly bool) (receipt *types.Receipt, err error) {
	// 获取当前区块的上下文信息构造EVM上下文
	context := evm.NewEVMContext(msg)

	// 创建EVM运行时对象
	env := runtime.NewEVM(context, evm.mStateDB, *evm.vmCfg)
	isCreate := strings.Compare(msg.To().String(), EvmAddress) == 0
	var (
		ret          []byte
		vmerr        error
		leftOverGas  uint64
		contractAddr common.Address
		snapshot     int
		execName     string
		methodName   string
	)

	// 为了方便计费，即使合约为新生成，也将地址的初始化放到外面操作
	if isCreate {
		// 使用随机生成的地址作为合约地址（这个可以保证每次创建的合约地址不会重复，不存在冲突的情况）
		contractAddr = evm.getNewAddr(txHash)
		if !env.StateDB.Empty(contractAddr.String()) {
			return receipt, model.ErrContractAddressCollision
		}
		// 只有新创建的合约才能生成合约名称
		execName = fmt.Sprintf("%s%s", types.ExecName(evmtypes.EvmPrefix), common.BytesToHash(txHash).Hex())
	} else {
		contractAddr = *msg.To()
	}

	// 状态机中设置当前交易状态
	evm.mStateDB.Prepare(common.BytesToHash(txHash), index)

	if isCreate {
		// 如果携带ABI数据，则对数据合法性进行检查
		if len(msg.ABI()) > 0 && types.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMABI) {
			_, err = abi.JSON(strings.NewReader(msg.ABI()))
			if err != nil {
				return receipt, err
			}
		}
		ret, snapshot, leftOverGas, vmerr = env.Create(runtime.AccountRef(msg.From()), contractAddr, msg.Data(), context.GasLimit, execName, msg.Alias(), msg.ABI())
	} else {
		inData := msg.Data()
		// 在这里进行ABI和十六进制的调用参数转换
		if len(msg.ABI()) > 0 && types.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMABI) {
			funcName, packData, err := abi.Pack(msg.ABI(), evm.mStateDB.GetAbi(msg.To().String()), readOnly)
			if err != nil {
				return receipt, err
			}
			inData = packData
			methodName = funcName
		}
		ret, snapshot, leftOverGas, vmerr = env.Call(runtime.AccountRef(msg.From()), *msg.To(), inData, context.GasLimit, msg.Value())
	}

	log.Debug("call(create) contract ", "input", common.Bytes2Hex(msg.Data()))
	usedGas := msg.GasLimit() - leftOverGas
	logMsg := "call contract details:"
	if isCreate {
		logMsg = "create contract details:"
	}
	log.Debug(logMsg, "caller address", msg.From().String(), "contract address", contractAddr.String(), "exec name", execName, "alias name", msg.Alias(), "usedGas", usedGas, "return data", common.Bytes2Hex(ret))

	curVer := evm.mStateDB.GetLastSnapshot()

	if vmerr != nil {
		log.Error("evm contract exec error", "error info", vmerr)
		return receipt, vmerr
	}

	// 计算消耗了多少费用（实际消耗的费用）
	usedFee, overflow := common.SafeMul(usedGas, uint64(msg.GasPrice()))
	// 费用消耗溢出，执行失败
	if overflow || usedFee > uint64(txFee) {
		// 如果操作没有回滚，则在这里处理
		if curVer != nil && snapshot >= curVer.GetID() && curVer.GetID() > -1 {
			evm.mStateDB.RevertToSnapshot(snapshot)
		}
		return receipt, model.ErrOutOfGas
	}

	// 打印合约中生成的日志
	evm.mStateDB.PrintLogs()

	// 没有任何数据变更
	if curVer == nil {
		return receipt, nil
	}

	// 从状态机中获取数据变更和变更日志
	kvSet, logs := evm.mStateDB.GetChangedData(curVer.GetID())
	contractReceipt := &evmtypes.ReceiptEVMContract{Caller: msg.From().String(), ContractName: execName, ContractAddr: contractAddr.String(), UsedGas: usedGas, Ret: ret}
	// 这里进行ABI调用结果格式化
	if len(methodName) > 0 && len(msg.ABI()) > 0 && types.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMABI) {
		jsonRet, err := abi.Unpack(ret, methodName, evm.mStateDB.GetAbi(msg.To().String()))
		if err != nil {
			// 这里出错不影响整体执行，只打印错误信息
			log.Error("unpack evm return error", "error", err)
		}
		contractReceipt.JsonRet = jsonRet
	}
	logs = append(logs, &types.ReceiptLog{Ty: evmtypes.TyLogCallContract, Log: types.Encode(contractReceipt)})
	logs = append(logs, evm.mStateDB.GetReceiptLogs(contractAddr.String())...)

	if types.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMKVHash) {
		// 将执行时生成的合约状态数据变更信息也计算哈希并保存
		hashKV := evm.calcKVHash(contractAddr, logs)
		if hashKV != nil {
			kvSet = append(kvSet, hashKV)
		}
	}

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kvSet, Logs: logs}

	// 返回之前，把本次交易在区块中生成的合约日志集中打印出来
	if evm.mStateDB != nil {
		evm.mStateDB.WritePreimages(evm.GetHeight())
	}

	// 替换导致分叉的执行数据信息
	state.ProcessFork(evm.GetHeight(), txHash, receipt)

	evm.collectEvmTxLog(txHash, contractReceipt, receipt)

	return receipt, nil
}

// CheckInit 检查是否初始化数据库
func (evm *EVMExecutor) CheckInit() {
	if evm.mStateDB == nil {
		evm.mStateDB = state.NewMemoryStateDB(evm.GetStateDB(), evm.GetLocalDB(), evm.GetCoinsAccount(), evm.GetHeight())
	}
}

// GetMessage 目前的交易中，如果是coins交易，金额是放在payload的，但是合约不行，需要修改Transaction结构
func (evm *EVMExecutor) GetMessage(tx *types.Transaction) (msg *common.Message, err error) {
	var action evmtypes.EVMContractAction
	err = types.Decode(tx.Payload, &action)
	if err != nil {
		return msg, err
	}
	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	from := getCaller(tx)
	to := getReceiver(tx)
	if to == nil {
		return msg, types.ErrInvalidAddress
	}

	gasLimit := action.GasLimit
	gasPrice := action.GasPrice
	if gasLimit == 0 {
		gasLimit = uint64(tx.Fee)
	}
	if gasPrice == 0 {
		gasPrice = uint32(1)
	}

	// 合约的GasLimit即为调用者为本次合约调用准备支付的手续费
	msg = common.NewMessage(from, to, tx.Nonce, action.Amount, gasLimit, gasPrice, action.Code, action.GetAlias(), action.Abi)
	return msg, err
}

func (evm *EVMExecutor) collectEvmTxLog(txHash []byte, cr *evmtypes.ReceiptEVMContract, receipt *types.Receipt) {
	log.Debug("evm collect begin")
	log.Debug("Tx info", "txHash", common.Bytes2Hex(txHash), "height", evm.GetHeight())
	log.Debug("ReceiptEVMContract", "data", fmt.Sprintf("caller=%v, name=%v, addr=%v, usedGas=%v, ret=%v", cr.Caller, cr.ContractName, cr.ContractAddr, cr.UsedGas, common.Bytes2Hex(cr.Ret)))
	log.Debug("receipt data", "type", receipt.Ty)
	for _, kv := range receipt.KV {
		log.Debug("KeyValue", "key", common.Bytes2Hex(kv.Key), "value", common.Bytes2Hex(kv.Value))
	}
	for _, kv := range receipt.Logs {
		log.Debug("ReceiptLog", "Type", kv.Ty, "log", common.Bytes2Hex(kv.Log))
	}
	log.Debug("evm collect end")
}

func (evm *EVMExecutor) calcKVHash(addr common.Address, logs []*types.ReceiptLog) (kv *types.KeyValue) {
	hashes := []byte{}
	// 使用合约状态变更的数据生成哈希，保存为执行KV
	for _, logItem := range logs {
		if evmtypes.TyLogEVMStateChangeItem == logItem.Ty {
			data := logItem.Log
			hashes = append(hashes, common.ToHash(data).Bytes()...)
		}
	}

	if len(hashes) > 0 {
		hash := common.ToHash(hashes)
		return &types.KeyValue{Key: getDataHashKey(addr), Value: hash.Bytes()}
	}
	return nil
}

func getDataHashKey(addr common.Address) []byte {
	return []byte(fmt.Sprintf("mavl-%v-data-hash:%v", evmtypes.ExecutorName, addr))
}

// 从交易信息中获取交易发起人地址
func getCaller(tx *types.Transaction) common.Address {
	return *common.StringToAddress(tx.From())
}

// 从交易信息中获取交易目标地址，在创建合约交易中，此地址为空
func getReceiver(tx *types.Transaction) *common.Address {
	if tx.To == "" {
		return nil
	}
	return common.StringToAddress(tx.To)
}
