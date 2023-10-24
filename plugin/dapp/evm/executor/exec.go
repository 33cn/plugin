// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/33cn/chain33/system/crypto/secp256k1eth"
	"math"
	"strings"
	"sync/atomic"

	"github.com/33cn/chain33/common/address"

	"github.com/33cn/chain33/account"
	"github.com/ethereum/go-ethereum/params"

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
	//nonce check
	err = evm.checkEvmNonce(msg, tx.GetSignature().GetTy())
	if err != nil {
		return nil, err
	}

	receipt, err := evm.innerExec(msg, tx.Hash(), tx.GetSignature().GetTy(), index, msg.GasLimit(), false)
	return receipt, err
}

// 通用的EVM合约执行逻辑封装
// readOnly 是否只读调用，仅执行evm abi查询时为true
func (evm *EVMExecutor) innerExec(msg *common.Message, txHash []byte, sigType int32, index int, txFee uint64, readOnly bool) (receipt *types.Receipt, err error) {

	cfg := evm.GetAPI().GetConfig()
	// 获取当前区块的上下文信息构造EVM上下文
	context := evm.NewEVMContext(msg, txHash)
	execAddr := evm.getEvmExecAddress()
	// 创建EVM运行时对象
	env := runtime.NewEVM(context, evm.mStateDB, *evm.vmCfg, cfg)
	isCreate := strings.Compare(msg.To().String(), execAddr) == 0 && len(msg.Data()) > 0
	isTransferOnly := strings.Compare(msg.To().String(), execAddr) == 0 && 0 == len(msg.Data())
	//coins转账，para数据作为备注交易
	isTransferNote := strings.Compare(msg.To().String(), execAddr) != 0 && !env.StateDB.Exist(msg.To().String()) && len(msg.Para()) > 0 && msg.Value() != 0
	var gas uint64
	if evm.GetAPI().GetConfig().IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkIntrinsicGas) {
		//加上固有消费的gas
		gas, err = intrinsicGas(msg, isCreate, true)
		if err != nil {
			return nil, err
		}
	}
	log.Info("innerExec", "isCreate", isCreate, "isTransferOnly", isTransferOnly, "isTransferNote:", isTransferNote, "evm-execaddr", execAddr, "msg.From:", msg.From(), "msg.To", msg.To().String(),

		"data size:", len(msg.Data()), "para size:", len(msg.Para()), "readOnly:", readOnly, "intrinsicGas:", gas, "value:", msg.Value(), "nonce:", msg.Nonce(), "gas:", msg.GasLimit())
	if msg.GasLimit() < gas {
		return nil, fmt.Errorf("%w: have %d, want %d", model.ErrIntrinsicGas, msg.GasLimit(), gas)
	}

	context.GasLimit = msg.GasLimit() - gas
	var (
		ret             []byte
		vmerr           error
		leftOverGas     uint64
		contractAddr    common.Address
		snapshot        int
		execName        string
		contractAddrStr string
	)

	if isTransferOnly || isTransferNote {
		caller := msg.From()
		var receiver common.Address
		if isTransferNote { //payload 数据作为备注信息，evm 不执行
			receiver = common.BytesToAddress(msg.To().Bytes())
		} else {
			receiver = common.BytesToAddress(msg.Para())
		}

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
		if types.IsEthSignID(sigType) {
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
	}

	//      evm
	// 状态机中设置当前交易状态
	evm.mStateDB.Prepare(common.BytesToHash(txHash), index)
	if isCreate {
		ret, snapshot, leftOverGas, vmerr = env.Create(runtime.AccountRef(msg.From()), contractAddr, msg.Data(), context.GasLimit, execName, msg.Alias(), msg.Value())
	} else {
		callPara := msg.Para()
		//log.Debug("call contract ", "callPara", common.Bytes2Hex(callPara))
		//设置eth 签名交易标签，如果msg.Value 不为0，则在evm 合约执行中从精度1e8转换为1e18
		env.SetEthTxFlag(types.IsEthSignID(sigType))
		ret, snapshot, leftOverGas, vmerr = env.Call(runtime.AccountRef(msg.From()), *msg.To(), callPara, context.GasLimit, msg.Value())
	}
	// 打印合约中生成的日志
	evm.mStateDB.PrintLogs()
	usedGas := msg.GasLimit() - leftOverGas
	logMsg := "call contract details:"
	if isCreate {
		logMsg = "create contract details:"
	}
	log.Info(logMsg, "caller address", msg.From().String(), "contract address", contractAddrStr, "exec name", execName, "alias name", msg.Alias(), "usedGas", usedGas, "leftOverGas:", leftOverGas,
		"msg.GasLimit:", msg.GasLimit())
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

		vmerr = fmt.Errorf("%s,detail: %s:", vmerr.Error(), string(ret))
		log.Error("innerExec evm contract exec error", "error info", vmerr, "string ret", string(ret), "hex ret:", common.Bytes2Hex(ret), "leftOverGas:", leftOverGas, "usedGas:", usedGas)
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
		log.Error("innerExec evm contract exec error", "overflow", overflow, "usedFee:", usedFee, "txFee:", txFee)
		return receipt, model.ErrOutOfGas
	}

	// 没有任何数据变更
	if curVer == nil {
		return receipt, nil
	}
	// 从状态机中获取数据变更和变更日志
	kvSet, logs := evm.mStateDB.GetChangedData(curVer.GetID())
	contractReceipt := &evmtypes.ReceiptEVMContract{Caller: msg.From().String(), ContractName: execName, ContractAddr: contractAddrStr, UsedGas: usedGas, Ret: ret}
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
			"created contract address", contractAddrStr, "isethtx", types.IsEthSignID(sigType))
	}

	return receipt, nil
}

// intrinsicGas 计算固定gas消费
func intrinsicGas(msg *common.Message, isContractCreation bool, isEIP2028 bool) (uint64, error) {
	var data []byte
	if isContractCreation {
		data = msg.Data()
	} else {
		data = msg.Para()
	}
	// Set the starting gas for the raw transaction
	var gas uint64
	if isContractCreation {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		nonZeroGas := params.TxDataNonZeroGasFrontier
		if isEIP2028 {
			nonZeroGas = params.TxDataNonZeroGasEIP2028
		}
		if (math.MaxUint64-gas)/nonZeroGas < nz {
			return 0, model.ErrGasUintOverflow
		}
		gas += nz * nonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, model.ErrGasUintOverflow
		}
		gas += z * params.TxDataZeroGas
	}

	return gas, nil
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

	mixAddressFork := evm.GetAPI().GetConfig().IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMMixAddress)
	to := getReceiver(&action, mixAddressFork)
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
	cfg := evm.GetAPI().GetConfig()
	conf := types.ConfSub(cfg, evmtypes.ExecutorName)
	if !conf.IsEnable("debugEvmTxLog") { //避免过多evm交易导致节点log 刷屏
		return
	}
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

// 获取evm 执行器地址
func (evm *EVMExecutor) getEvmExecAddress() string {

	isFork := evm.GetAPI().GetConfig().IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMAddressInit)
	if isFork && address.IsEthAddress(evmExecAddress) {
		return evmExecFormatAddress
	}

	return evmExecAddress
}

func (evm *EVMExecutor) checkEvmNonce(msg *common.Message, sigType int32) error {
	if types.IsEthSignID(sigType) && evm.GetAPI().GetConfig().IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEvmExecNonceV2) {
		nonceLocalKey := secp256k1eth.CaculCoinsEvmAccountKey(msg.From().String())
		evmNonce := &types.EvmAccountNonce{}
		nonceV, err := evm.GetLocalDB().Get(nonceLocalKey)
		if err == nil {
			_ = types.Decode(nonceV, evmNonce)

		}
		log.Info("EVMExecutor", "from", msg.From(), "checkEvmNonce localdb nonce:", evmNonce.Nonce, "tx.Nonce:", msg.Nonce())
		if msg.Nonce() < evmNonce.GetNonce() {
			return types.ErrLowNonce
		} else if msg.Nonce() > evmNonce.GetNonce() {
			return errors.New("nonce too high")
		}
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
func getReceiver(action *evmtypes.EVMContractAction, mixAddressFork bool) *common.Address {
	if action.ContractAddr == "" {
		return nil
	}
	if mixAddressFork {
		return common.StringToAddress(action.ContractAddr)
	}
	return common.StringToAddressLegacy(action.ContractAddr)
}
