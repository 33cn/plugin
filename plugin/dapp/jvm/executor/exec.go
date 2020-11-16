package executor

import "C"
import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/jvm/executor/contract"
	jvmTypes "github.com/33cn/plugin/plugin/dapp/jvm/types"
)

// Exec_CreateJvmContract 创建合约
func (jvm *JVMExecutor) Exec_CreateJvmContract(createJvmContract *jvmTypes.CreateJvmContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	jvm.prepareExecContext(tx, index)
	// 使用随机生成的地址作为合约地址（这个可以保证每次创建的合约地址不会重复，不存在冲突的情况）

	log.Debug("Exec_CreateJvmContract", "createJvmContract.Name", createJvmContract.Name)
	contractAddr := address.GetExecAddress(createJvmContract.Name)
	contractAddrInStr := contractAddr.String()
	log.Debug("Exec_CreateJvmContract", "new created jvm jvmContract addr =", contractAddrInStr)
	if !jvm.mStateDB.Empty(contractAddrInStr) {
		return nil, jvmTypes.ErrContractAddressCollisionJvm
	}

	codeSize := len(createJvmContract.GetCode())
	if codeSize > jvmTypes.MaxCodeSize {
		return nil, jvmTypes.ErrMaxCodeSizeExceededJvm
	}

	if 0 == codeSize {
		return nil, jvmTypes.ErrNUllJvmContract
	}

	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	from := address.PubKeyToAddress(jvm.tx.GetSignature().GetPubkey())
	_ = jvm.mStateDB.Snapshot()

	// 创建新的合约对象，包含双方地址以及合约代码，可用Gas信息
	code, err := common.FromHex(createJvmContract.Code)
	if nil != err {
		return nil, jvmTypes.ErrJvmCodeString
	}
	jvmContract := contract.NewContract(contract.AccountRef(*from), contract.AccountRef(*contractAddr), 0)
	jvmContract.SetCallCode(*contractAddr, common.BytesToHash(common.Sha256(code)), code)

	// 创建一个新的账户对象（合约账户）
	jvm.mStateDB.CreateAccount(contractAddrInStr, jvmContract.CallerAddress.String(), createJvmContract.Name)
	jvm.mStateDB.SetCodeAndAbi(contractAddrInStr, code, nil)

	receipt, err := jvm.GenerateExecReceipt(
		createJvmContract.Name,
		jvmContract.CallerAddress.String(),
		contractAddrInStr,
		jvmTypes.CreateJvmContractAction)
	log.Debug("jvm create", "receipt", receipt, "err info", err)

	return receipt, err
}

// Exec_CallJvmContract 调用合约
func (jvm *JVMExecutor) Exec_CallJvmContract(callJvmContract *jvmTypes.CallJvmContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	jvm.prepareExecContext(tx, index)
	//因为在真正地执行user.jvm.xxx合约前，还需要通过Jvm合约平台获取其合约字节码，
	//所以需要先将其合约名字设置为Jvm
	jvm.mStateDB.SetCurrentExecutorName(jvmTypes.JvmX)

	log.Debug("jvm call", "Para CallJvmContract", callJvmContract,
		"string(tx.Execer)", string(tx.Execer))

	userJvmAddr, contractName, contractAccount, err := jvm.creatJarFileWithCode(string(tx.Execer))
	if nil != err {
		return nil, err
	}

	//将当前合约执行名字修改为user.jvm.xxx
	jvm.mStateDB.SetCurrentExecutorName(string(jvm.GetAPI().GetConfig().GetParaExec(tx.Execer)))
	_ = jvm.mStateDB.Snapshot()

	//1st step: create tx para
	caller := tx.From()
	actionData := callJvmContract.ActionData
	log.Debug("jvm call para", "from", caller,
		"ContractName", string(tx.Execer),
		"ActionName", callJvmContract.ActionData[0],
		"ActionData", callJvmContract.ActionData)
	//2nd step: just call contract
	//在此处将gojvm指针传递到c实现的jvm中，进行回调的时候用来区分是获取数据时，使用执行db还是查询db
	errinfo := runJava(contractName, actionData, jvm, TX_EXEC_JOB)
	//合约执行失败，有2种可能
	//1.余额不足等原因被合约强制退出本次交易
	//2.java合约本身的代码问题，抛出异常
	if errinfo != nil || jvm.forceStopInfo.occurred {
		var exeErr error
		if errinfo != nil {
			exeErr = errinfo
			log.Error("call jvm contract", "failed to call contract due to stopWithError", exeErr.Error())
		} else {
			exeErr = jvm.forceStopInfo.info
			log.Error("call jvm contract", "failed to call contract due to stop with error", exeErr.Error())
		}

		return nil, exeErr
	}

	receipt, _ := jvm.GenerateExecReceipt(
		contractAccount.GetExecName(),
		caller,
		userJvmAddr,
		jvmTypes.CallJvmContractAction)
	log.Debug("jvm call", "receipt", receipt)
	log.Debug("jvm call succeed", "tx hash", jvm.txHash)

	return receipt, nil
}

// Exec_UpdateJvmContract 创建合约
func (jvm *JVMExecutor) Exec_UpdateJvmContract(updateJvmContract *jvmTypes.UpdateJvmContract, tx *types.Transaction, index int) (*types.Receipt, error) {
	jvm.prepareExecContext(tx, index)

	// 使用随机生成的地址作为合约地址（这个可以保证每次创建的合约地址不会重复，不存在冲突的情况）
	contractAddr := address.GetExecAddress(updateJvmContract.Name)
	contractAddrInStr := contractAddr.String()
	if !jvm.mStateDB.Exist(contractAddrInStr) {
		return nil, jvmTypes.ErrContractNotExist
	}
	//只有创建合约的人可以更新合约
	manager := jvm.mStateDB.GetAccount(contractAddrInStr).GetCreator()
	if tx.From() != manager {
		log.Error("update jvmContract", "tx from:", tx.From(), "manager:", manager)
		return nil, jvmTypes.ErrNoPermission
	}
	log.Debug("jvm update", "updated jvm jvmContract addr =", contractAddrInStr)

	codeSize := len(updateJvmContract.GetCode())
	if codeSize > jvmTypes.MaxCodeSize {
		return nil, jvmTypes.ErrMaxCodeSizeExceededJvm
	}

	if 0 == codeSize {
		return nil, jvmTypes.ErrNUllJvmContract
	}

	// 此处暂时不考虑消息发送签名的处理，chain33在mempool中对签名做了检查
	from := address.PubKeyToAddress(jvm.tx.GetSignature().GetPubkey())
	_ = jvm.mStateDB.Snapshot()
	// 更新合约对象，包含双方地址以及合约代码，可用Gas信息
	code, err := common.FromHex(updateJvmContract.Code)
	if nil != err {
		return nil, jvmTypes.ErrJvmCodeString
	}
	jvmContract := contract.NewContract(contract.AccountRef(*from), contract.AccountRef(*contractAddr), 0)
	jvmContract.SetCallCode(*contractAddr, common.BytesToHash(common.Sha256(code)), code)
	jvm.mStateDB.SetCodeAndAbi(contractAddrInStr, code, nil)

	receipt, err := jvm.GenerateExecReceipt(
		updateJvmContract.Name,
		jvmContract.CallerAddress.String(),
		contractAddrInStr,
		jvmTypes.UpdateJvmContractAction)
	log.Debug("jvm create", "receipt", receipt, "err info", err)

	return receipt, err
}
