package executor

import (
	"os"
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/plugin/plugin/dapp/jvm/executor/state"
	jvmTypes "github.com/33cn/plugin/plugin/dapp/jvm/types"
)

func (jvm *JVMExecutor) creatJarFileWithCode(contractName string) (string, string, *state.ContractAccount, error) {
	userJvmAddr := address.ExecAddress(contractName)
	contractAccount := jvm.mStateDB.GetAccount(userJvmAddr)
	if nil == contractAccount {
		return "", "", nil, jvmTypes.ErrContractNotExist
	}
	temp := strings.Split(contractName, ".")
	//just keep the last name
	contractName = temp[len(temp)-1]
	jarPath := "./" + contractName + ".jar"
	jarFileExist := true
	//判断jar文件是否存在
	_, err := os.Stat(jarPath)
	if err != nil && !os.IsExist(err) {
		jarFileExist = false
	}

	if !jarFileExist {
		javaClassfile, err := os.OpenFile(jarPath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Debug("jvm call", "os.OpenFile failed due to cause:", err.Error())
			return "", "", nil, err
		}
		code := contractAccount.Data.GetCode()
		writeLen, err := javaClassfile.Write(code)
		if writeLen != len(code) {
			return "", "", nil, jvmTypes.ErrWriteJavaClass
		}
		if closeErr := javaClassfile.Close(); nil != closeErr {
			log.Debug("jvm call", "javaClassfile.Close() failed due to cause:", closeErr.Error())
			return "", "", nil, closeErr
		}
	}
	return userJvmAddr, contractName, contractAccount, nil
}
