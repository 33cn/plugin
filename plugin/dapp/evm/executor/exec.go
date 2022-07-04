// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/33cn/chain33/account"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
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
	msg, err := evm.GetMessage(tx, index, nil)
	if err != nil {
		return nil, err
	}

	cfg := evm.GetAPI().GetConfig()
	if !evmDebugInited {
		conf := types.ConfSub(cfg, evmtypes.ExecutorName)
		atomic.StoreInt32(&evm.vmCfg.Debug, int32(conf.GInt("evmDebugEnable")))
		evmDebugInited = true
	}

	receipt, err := evm.innerExec(msg, tx.Hash(), tx.GetSignature().GetTy(), index, msg.GasLimit(), false)
	return receipt, err
}

// 通用的EVM合约执行逻辑封装
// readOnly 是否只读调用，仅执行evm abi查询时为true
func (evm *EVMExecutor) innerExec(msg *common.Message, txHash []byte, sigType int32, index int, txFee uint64, readOnly bool) (receipt *types.Receipt, err error) {
	// 获取当前区块的上下文信息构造EVM上下文
	context := evm.NewEVMContext(msg, txHash)
	cfg := evm.GetAPI().GetConfig()
	// 创建EVM运行时对象
	env := runtime.NewEVM(context, evm.mStateDB, *evm.vmCfg, cfg)
	isCreate := strings.Compare(msg.To().String(), EvmAddress) == 0 && len(msg.Data()) > 0
	isTransferOnly := strings.Compare(msg.To().String(), EvmAddress) == 0 && 0 == len(msg.Data())
	log.Info("innerExec", "isCreate", isCreate, "isTransferOnly", isTransferOnly, "evmaddr", EvmAddress, "msg.From:", msg.From(), "msg.To", msg.To().String(),
		"data size:", len(msg.Data()), "readOnly", readOnly)
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
		kvSet, logs := evm.mStateDB.GetChangedData(curVer.GetID())
		receipt = &types.Receipt{Ty: types.ExecOk, KV: kvSet, Logs: logs}
		return receipt, nil
	} else if isCreate {
		if types.IsEthSignID(int32(sigType)) {
			// 通过ethsign 签名的兼容交易 采用from+nonce 创建合约地址
			contractAddr = evm.createEvmContractAddress(msg.From(), uint64(msg.Nonce()))
		} else {
			contractAddr = evm.createContractAddress(msg.From(), txHash)
		}

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
		var visiableOut []byte
		for i := 0; i < len(ret); i++ {
			//显示[32,126]之间的字符
			if ret[i] < 32 || ret[i] > 126 {
				continue
			}
			visiableOut = append(visiableOut, ret[i])
		}
		ret = visiableOut
		vmerr = fmt.Errorf("%s,detail: %s", vmerr.Error(), string(ret))
		log.Error("evm contract exec error", "error info", vmerr, "ret", string(ret))

		return receipt, vmerr
	}

	// 计算消耗了多少费用（实际消耗的费用）
	usedFee, overflow := common.SafeMul(usedGas, uint64(msg.GasPrice()))
	// 费用消耗溢出，执行失败
	if overflow || usedFee > txFee {
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
	kvSet, logs := evm.mStateDB.GetChangedData(curVer.GetID())
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

	if isCreate && !readOnly {
		log.Info("innerExec", "Succeed to created new contract with name", msg.Alias(),
			"created contract address", contractAddrStr)
	}
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

// GetMessage 目前的交易中，如果是coins交易，金额是放在payload的，但是合约不行，需要修改Transaction结构
func (evm *EVMExecutor) GetMessage(tx *types.Transaction, index int, fromPtr *common.Address) (msg *common.Message, err error) {
	var action evmtypes.EVMContractAction
	err = types.Decode(tx.Payload, &action)
	if err != nil {
		return msg, err
	}
	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	var from common.Address
	if fromPtr == nil {
		from = getCaller(tx)
	} else {
		from = *fromPtr
	}

	to := getReceiver(&action)
	if to == nil {
		return msg, types.ErrInvalidAddress
	}

	gasPrice := action.GasPrice
	//gasLimit 直接从交易费1:1转化而来，忽略action.GasLimit
	gasLimit := uint64(evm.GetTxFee(tx, index))
	//如果未设置交易费，则尝试读取免交易费联盟链模式下的gas设置
	if 0 == gasLimit {
		cfg := evm.GetAPI().GetConfig()
		conf := types.ConfSub(cfg, evmtypes.ExecutorName)
		gasLimit = uint64(conf.GInt("evmGasLimit"))
		if 0 == gasLimit {
			return nil, model.ErrNoGasConfigured
		}
		log.Info("GetMessage", "gasLimit is set to for permission blockchain", gasLimit)
	}

	if gasPrice == 0 {
		gasPrice = uint32(1)
	}
	log.Debug("GetMessage", "code size", len(action.Code), "data size:", len(action.Para))
	// 合约的GasLimit即为调用者为本次合约调用准备支付的手续费
	msg = common.NewMessage(from, to, tx.Nonce, action.Amount, gasLimit, gasPrice, action.Code, action.Para, action.GetAlias())
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
func getReceiver(action *evmtypes.EVMContractAction) *common.Address {
	if action.ContractAddr == "" {
		return nil
	}
	return common.StringToAddress(action.ContractAddr)
}
