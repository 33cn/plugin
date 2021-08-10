// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"errors"
	"fmt"
	manager "github.com/33cn/chain33/system/dapp/manage/types"
	"strings"

	"github.com/33cn/chain33/account"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/runtime"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/state"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// Exec_EvmExec 创建或调用合约
func (evm *EVMExecutor) Exec_Exec(payload *evmtypes.EVMContractExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	evm.CheckInit()
	// 先转换消息

	msg, err := evm.GetMessageExec(payload, tx, index)
	if err != nil {
		return nil, err
	}

	return evm.innerExec(msg, tx.Hash(), index, evm.GetTxFee(tx, index), false)
}

// Exec_EvmExec 创建或调用合约
func (evm *EVMExecutor) Exec_Update(payload *evmtypes.EVMContractUpdate, tx *types.Transaction, index int) (*types.Receipt, error) {
	evm.CheckInit()
	// 先转换消息

	msg, err := evm.GetMessageUpdate(payload, tx, index)
	if err != nil {
		return nil, err
	}

	receipt, err := evm.innerExec(msg, tx.Hash(), index, evm.GetTxFee(tx, index), false)
	if err != nil {
		return nil, err
	}

	msgDestory := common.NewMessage(msg.From(), common.StringToAddress(payload.Addr), tx.Nonce, payload.Amount, msg.GasLimit(), msg.GasPrice(), payload.Code, nil, payload.GetAlias(), "")
	receiptDestory, err := evm.contractLifecycle(msgDestory, tx.Hash(), evmtypes.EvmDestroyAction)
	if err != nil {
		return nil, err
	}

	receipt.KV = append(receipt.KV, receiptDestory.KV...)
	receipt.Logs = append(receipt.Logs, receiptDestory.Logs...)

	return receipt, nil
}

// Exec_EvmDestroy 销毁合约
func (evm *EVMExecutor) Exec_Destroy(payload *evmtypes.EVMContractDestroy, tx *types.Transaction, index int) (*types.Receipt, error) {
	evm.CheckInit()
	// 先转换消息
	msg, err := evm.GetMessageLifecycle(payload.Addr, tx)
	if err != nil {
		return nil, err
	}

	return evm.contractLifecycle(msg, tx.Hash(), evmtypes.EvmDestroyAction)
}

// Exec_Freeze 冻结合约
func (evm *EVMExecutor) Exec_Freeze(payload *evmtypes.EVMContractFreeze, tx *types.Transaction, index int) (*types.Receipt, error) {
	evm.CheckInit()
	// 先转换消息
	msg, err := evm.GetMessageLifecycle(payload.Addr, tx)
	if err != nil {
		return nil, err
	}

	return evm.contractLifecycle(msg, tx.Hash(), evmtypes.EvmFreezeAction)
}

// Exec_EvmRelease 解冻合约
func (evm *EVMExecutor) Exec_Release(payload *evmtypes.EVMContractRelease, tx *types.Transaction, index int) (*types.Receipt, error) {
	evm.CheckInit()
	// 先转换消息
	msg, err := evm.GetMessageLifecycle(payload.Addr, tx)
	if err != nil {
		return nil, err
	}

	return evm.contractLifecycle(msg, tx.Hash(), evmtypes.EvmReleaseAction)
}

// 通用的EVM合约执行逻辑封装
// readOnly 是否只读调用，仅执行evm abi查询时为true
func (evm *EVMExecutor) innerExec(msg *common.Message, txHash []byte, index int, txFee int64, readOnly bool) (receipt *types.Receipt, err error) {
	var logs []*types.ReceiptLog
	var kvSet []*types.KeyValue

	// 获取当前区块的上下文信息构造EVM上下文
	context := evm.NewEVMContext(msg)
	cfg := evm.GetAPI().GetConfig()
	// 创建EVM运行时对象
	env := runtime.NewEVM(context, evm.mStateDB, *evm.vmCfg, cfg)
	isCreate := strings.Compare(msg.To().String(), EvmAddress) == 0 && len(msg.Data()) > 0
	isTransferOnly := strings.Compare(msg.To().String(), EvmAddress) == 0 && 0 == len(msg.Data())
	isUpdate := isCreate && ( len(msg.PreAddr()) != 0 )

	var (
		ret             []byte
		vmerr           error
		leftOverGas     uint64
		contractAddr    common.Address
		snapshot        int
		execName        string
		contractAddrStr string
	)

	if isTransferOnly {
		caller := msg.From()
		receiver := common.BytesToAddress(msg.Para())

		if !evm.mStateDB.CanTransfer(caller.String(), msg.Value()) {
			log.Error("innerExec", "Not enough balance to be transferred from", caller.String(), "amout", msg.Value())
			return nil, types.ErrNoBalance
		}
		env.StateDB.Snapshot()
		env.Transfer(env.StateDB, caller, receiver, msg.Value())
		curVer := evm.mStateDB.GetLastSnapshot()
		kvSet, logs = evm.mStateDB.GetChangedData(curVer.GetID())
		receipt = &types.Receipt{Ty: types.ExecOk, KV: kvSet, Logs: logs}
		return receipt, nil
	} else if isCreate {
		// 使用随机生成的地址作为合约地址（这个可以保证每次创建的合约地址不会重复，不存在冲突的情况）
		contractAddr = evm.createContractAddress(msg.From(), txHash)
		contractAddrStr = contractAddr.String()
		if !env.StateDB.Empty(contractAddrStr) {
			return receipt, model.ErrContractAddressCollision
		}
		// 只有新创建的合约才能生成合约名称
		execName = fmt.Sprintf("%s%s", cfg.ExecName(evmtypes.EvmPrefix), common.BytesToHash(txHash).Hex())
	} else {
		contractAddr = *msg.To()
		contractAddrStr = contractAddr.String()
		if !env.StateDB.Exist(contractAddrStr) {
			log.Error("innerExec", "Contract not exist for address", contractAddrStr)
			return receipt, model.ErrContractNotExist
		}
		log.Info("innerExec", "Contract exist for address", contractAddrStr)
	}

	// 状态机中设置当前交易状态
	evm.mStateDB.Prepare(common.BytesToHash(txHash), index)

	if isCreate {
		ret, snapshot, leftOverGas, vmerr = env.Create(runtime.AccountRef(msg.From()), contractAddr, msg.Data(), context.GasLimit, execName, msg.Alias(), msg.Value())
	} else {
		callPara := msg.Para()
		log.Debug("call contract ", "callPara", common.Bytes2Hex(callPara))
		ret, snapshot, leftOverGas, vmerr = env.Call(runtime.AccountRef(msg.From()), *msg.To(), callPara, context.GasLimit, msg.Value())
	}
	// 打印合约中生成的日志
	evm.mStateDB.PrintLogs()

	usedGas := msg.GasLimit() - leftOverGas
	logMsg := "call contract details:"
	if isCreate {
		logMsg = "create contract details:"
	}
	log.Debug(logMsg, "caller address", msg.From().String(), "contract address", contractAddrStr, "exec name", execName, "alias name", msg.Alias(), "usedGas", usedGas, "return data", common.Bytes2Hex(ret))
	curVer := evm.mStateDB.GetLastSnapshot()
	if vmerr != nil {
		log.Error("evm contract exec error", "error info", vmerr, "ret", string(ret))
		if ret != nil {
			vmerr = errors.New(fmt.Sprintf("%s,detail: %s", vmerr.Error(), string(ret)))
		} else {
			vmerr = errors.New(fmt.Sprintf("%s", vmerr.Error()))
		}

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

	// 没有任何数据变更
	if curVer == nil {
		return receipt, nil
	}
	// 从状态机中获取数据变更和变更日志
	kvSet, logs = evm.mStateDB.GetChangedData(curVer.GetID())
	contractReceipt := &evmtypes.ReceiptEVMContract{Caller: msg.From().String(), ContractName: execName, ContractAddr: contractAddrStr, UsedGas: usedGas, Ret: ret}
	//// 这里进行ABI调用结果格式化
	//if len(methodName) > 0 && len(msg.Para()) > 0 && cfg.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMABI) {
	//	jsonRet, err := abi.Unpack(ret, methodName, evm.mStateDB.GetAbi(msg.To().String()))
	//	if err != nil {
	//		// 这里出错不影响整体执行，只打印错误信息
	//		log.Error("unpack evm return error", "error", err)
	//	}
	//	contractReceipt.JsonRet = jsonRet
	//}
	logs = append(logs, &types.ReceiptLog{Ty: evmtypes.TyLogCallContract, Log: types.Encode(contractReceipt)})
	logs = append(logs, evm.mStateDB.GetReceiptLogs(contractAddrStr)...)

	if isCreate {
		var evmstat evmtypes.ReceiptEvmStatistic
		evmstat.Addr = contractAddrStr
		if isUpdate {
			evmstat.PreAddr = msg.PreAddr()
		}
		logs = append(logs,  &types.ReceiptLog{Ty: evmtypes.TyLogEVMStatisticDataInit, Log: types.Encode(&evmstat)})
	} else {
		var evmstat evmtypes.ReceiptEvmStatistic
		evmstat.Addr = contractAddrStr
		evmstat.CallTimes = 1
		evmstat.Caller = msg.From().String()
		evmstat.SuccseccTimes = 1
		logs = append(logs, &types.ReceiptLog{Ty: evmtypes.TyLogEVMStatisticData, Log: types.Encode(&evmstat)})
	}

	if cfg.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMKVHash) {
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
	state.ProcessFork(cfg, evm.GetHeight(), txHash, receipt)

	evm.collectEvmTxLog(txHash, contractReceipt, receipt)

	if isCreate {
		log.Info("innerExec", "Succeed to created new contract with name", msg.Alias(),
			"created contract address", contractAddrStr)
	}

	return receipt, nil
}

func isSuperManager(cfg *types.Chain33Config, addr string) bool {
	confManager := types.ConfSub(cfg, manager.ManageX)
	for _, m := range confManager.GStrList("superManager") {
		if addr == m {
			return true
		}
	}
	return false
}

func (evm *EVMExecutor)checkAccessPermission(cfg *types.Chain33Config, contractAddr, addr string) bool {
	contractAccount := evm.mStateDB.GetAccount(contractAddr)
	if contractAccount.GetCreator() == addr {
		return true
	}

	return isSuperManager(cfg, addr)
}

func (evm *EVMExecutor) contractLifecycle(msg *common.Message, txHash []byte, flag int) (receipt *types.Receipt, err error) {
	// 获取当前区块的上下文信息构造EVM上下文
	context := evm.NewEVMContext(msg)
	cfg := evm.GetAPI().GetConfig()
	// 创建EVM运行时对象
	env := runtime.NewEVM(context, evm.mStateDB, *evm.vmCfg, cfg)

	contractAddr := *msg.To()
	contractAddrStr := contractAddr.String()
	if !env.StateDB.Exist(contractAddrStr) {
		log.Error("contractLifecycle", "Contract not exist for address", contractAddrStr)
		return nil, model.ErrContractNotExist
	}

	if !evm.checkAccessPermission(cfg, contractAddrStr, msg.From().String()) {
		log.Error("contractLifecycle", "no permission for from address", msg.From().String())
		return nil, model.ErrPermission
	}

	switch flag {
	case evmtypes.EvmDestroyAction:
		env.Destroy(runtime.AccountRef(msg.From()), *msg.To())
		break
	case evmtypes.EvmFreezeAction:
		env.Freeze(*msg.To())
		break
	case evmtypes.EvmReleaseAction:
		env.Release(*msg.To())
		break
	default:
		return nil, model.ErrOperation
	}

	// 从状态机中获取数据变更和变更日志
	curVer := evm.mStateDB.GetLastSnapshot()
	kvSet, logs := evm.mStateDB.GetChangedData(curVer.GetID())
	contractReceipt := &evmtypes.ReceiptEVMContract{Caller: msg.From().String(), ContractName: "", ContractAddr: contractAddrStr, UsedGas: 0, Ret: nil}

	logs = append(logs, &types.ReceiptLog{Ty: evmtypes.TyLogCallContract, Log: types.Encode(contractReceipt)})
	logs = append(logs, evm.mStateDB.GetReceiptLogs(contractAddrStr)...)

	if cfg.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMKVHash) {
		// 将执行时生成的合约状态数据变更信息也计算哈希并保存
		hashKV := evm.calcKVHash(contractAddr, logs)
		if hashKV != nil {
			kvSet = append(kvSet, hashKV)
		}
	}

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kvSet, Logs: logs}

	// 替换导致分叉的执行数据信息
	state.ProcessFork(cfg, evm.GetHeight(), txHash, receipt)

	evm.collectEvmTxLog(txHash, contractReceipt, receipt)

	return receipt, nil
}

// CheckInit 检查是否初始化数据库
func (evm *EVMExecutor) CheckInit() {
	cfg := evm.GetAPI().GetConfig()
	if !cfg.IsPara() {
		//主链
		evm.mStateDB = state.NewMemoryStateDB(evm.GetStateDB(), evm.GetLocalDB(), evm.GetCoinsAccount(), evm.GetHeight(), evm.GetAPI())
		return
	}
	//平行链
	conf := types.ConfSub(cfg, evmtypes.ExecutorName)
	ethMapFromExecutor := conf.GStr("ethMapFromExecutor")
	ethMapFromSymbol := conf.GStr("ethMapFromSymbol")
	if "" == ethMapFromExecutor || "" == ethMapFromSymbol {
		panic("Both ethMapFromExecutor and ethMapFromSymbol should be configured, " + "ethMapFromExecutor=" + ethMapFromExecutor + ", ethMapFromSymbol=" + ethMapFromSymbol)
	}
	accountDB, _ := account.NewAccountDB(evm.GetAPI().GetConfig(), ethMapFromExecutor, ethMapFromSymbol, evm.GetStateDB())
	evm.mStateDB = state.NewMemoryStateDB(evm.GetStateDB(), evm.GetLocalDB(), accountDB, evm.GetHeight(), evm.GetAPI())
}

// GetMessageExec 目前的交易中，如果是coins交易，金额是放在payload的，但是合约不行，需要修改Transaction结构
func (evm *EVMExecutor) GetMessageExec(payload *evmtypes.EVMContractExec, tx *types.Transaction, index int) (msg *common.Message, err error) {
	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	from := getCaller(tx)
	to := getReceiver(payload.ContractAddr)
	if to == nil {
		return msg, types.ErrInvalidAddress
	}

	gasLimit := payload.GasLimit
	gasPrice := payload.GasPrice
	if gasLimit == 0 {
		gasLimit = uint64(evm.GetTxFee(tx, index))
	}
	if gasPrice == 0 {
		gasPrice = uint32(1)
	}

	// 合约的GasLimit即为调用者为本次合约调用准备支付的手续费
	msg = common.NewMessage(from, to, tx.Nonce, payload.Amount, gasLimit, gasPrice, payload.Code, payload.Para, payload.GetAlias(), "")
	return msg, err
}

// GetMessageUpdate 目前的交易中，如果是coins交易，金额是放在payload的，但是合约不行，需要修改Transaction结构
func (evm *EVMExecutor) GetMessageUpdate(payload *evmtypes.EVMContractUpdate, tx *types.Transaction, index int) (msg *common.Message, err error) {
	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	from := getCaller(tx)
	to := getReceiver(tx.To)
	if to == nil {
		return msg, types.ErrInvalidAddress
	}

	gasLimit := payload.GasLimit
	gasPrice := payload.GasPrice
	if gasLimit == 0 {
		gasLimit = uint64(evm.GetTxFee(tx, index))
	}
	if gasPrice == 0 {
		gasPrice = uint32(1)
	}

	// 合约的GasLimit即为调用者为本次合约调用准备支付的手续费
	msg = common.NewMessage(from, to, tx.Nonce, payload.Amount, gasLimit, gasPrice, payload.Code, nil, payload.GetAlias(), payload.Addr)
	return msg, err
}

// GetMessageLifecycle 目前的交易中，如果是coins交易，金额是放在payload的，但是合约不行，需要修改Transaction结构
func (evm *EVMExecutor) GetMessageLifecycle(addr string, tx *types.Transaction) (msg *common.Message, err error) {
	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	from := getCaller(tx)
	to := common.StringToAddress(addr)
	if to == nil {
		return msg, types.ErrInvalidAddress
	}

	// 合约的GasLimit即为调用者为本次合约调用准备支付的手续费
	msg = common.NewMessage(from, to, tx.Nonce, 0, 0, 0, nil, nil, "", "")
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

// GetTxFee 获取交易手续费，支持交易组
func (evm *EVMExecutor) GetTxFee(tx *types.Transaction, index int) int64 {
	fee := tx.Fee
	cfg := evm.GetAPI().GetConfig()
	if fee == 0 && cfg.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMTxGroup) {
		if tx.GroupCount >= 2 {
			txs, err := evm.GetTxGroup(index)
			if err != nil {
				log.Error("evm GetTxFee", "get tx group fail", err, "hash", hex.EncodeToString(tx.Hash()))
				return 0
			}
			fee = txs[0].Fee
		}
	}
	return fee
}

func getDataHashKey(addr common.Address) []byte {
	return []byte(fmt.Sprintf("mavl-%v-data-hash:%v", evmtypes.ExecutorName, addr))
}

// 从交易信息中获取交易发起人地址
func getCaller(tx *types.Transaction) common.Address {
	return *common.StringToAddress(tx.From())
}

// 从交易信息中获取交易目标地址，在创建合约交易中，此地址为空
func getReceiver(contractAddr string) *common.Address {
	if contractAddr == "" {
		return nil
	}
	return common.StringToAddress(contractAddr)
}
