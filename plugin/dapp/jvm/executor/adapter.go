package executor

//#cgo CFLAGS: -I../openjdk/header
//#cgo LDFLAGS: -L../openjdk -ljli
//#cgo LDFLAGS: -ldl -lpthread -lc
//#include <stdlib.h>
//#include <jli.h>
import "C"

import (
	"bytes"
	"errors"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/33cn/chain33/common"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/jvm/executor/state"
	jvmTypes "github.com/33cn/plugin/plugin/dapp/jvm/types"
	_ "github.com/ianlancetaylor/cgosymbolizer"
)

const (
	JLI_SUCCESS  = int(0)
	JLI_FAIL     = int(-1)
	TX_EXEC_JOB  = C.int(0)
	TX_QUERY_JOB = C.int(1)
	Bool_TRUE    = C.int(1)
	Bool_FALSE   = C.int(0)
)

var (
	jvm_init_alreay      = false
	consensusType        = ""
	Chain33LoaderJarPath = "." //路径信息不需要包含字符‘/’，C语言中拼接时，会添加
	//初始化random仅作为solo模式下的测试使用，没有其他用途
	randomStr           = "0x42f4eada40e876c476204dfb0749b2cda90020c68992dcacba6ea5a0fa75a371"
	lastBlockNum4Random = int64(0)
	lastHash4Random     = []byte{}
)

//调用java合约交易
func runJava(contract string, para []string, jvmHandleGo *JVMExecutor, jobType C.int) error {
	if TX_EXEC_JOB == jobType {
		height := jvmHandleGo.GetHeight()
		lastHash := jvmHandleGo.GetLastHash()
		//当共识类型为ticket，且产生新的区块时，需要重新获取random数据
		if consensusType == "ticket" && height != lastBlockNum4Random || !bytes.Equal(lastHash4Random, lastHash) {
			req := &chain33Types.ReqRandHash{
				ExecName: "ticket",
				BlockNum: jvmHandleGo.GetHeight(),
				Hash:     jvmHandleGo.GetLastHash(),
			}
			data, err := jvmHandleGo.GetExecutorAPI().GetRandNum(req)
			if nil != err {
				log.Error("getRandom", "GetRandom failed due to:", err.Error())
				return err
			}
			randomStr = common.ToHex(data)
			lastBlockNum4Random = height
			lastHash4Random = lastHash
		}
	}

	//构建jdk的输入参数
	tx2Exec := append([]string{contract}, para...)
	argc, argv := buildJavaArgument(tx2Exec)
	if TX_EXEC_JOB == jobType {
		//因为query的内在逻辑问题，参数的内存释放由jdk内部进行释放
		defer freeArgument(argc, argv)
	}

	var exception1DPtr *C.char
	exception := &exception1DPtr
	result := C.JLI_Exec_Contract(argc, argv, exception, jobType, (*C.char)(unsafe.Pointer(jvmHandleGo)))
	if int(result) != JLI_SUCCESS {
		exInfo := C.GoString(*exception)
		defer C.free(unsafe.Pointer(*exception))
		log.Debug("adapter::runJava", "java stopWithError", exInfo)
		return errors.New(exInfo)
	}
	return nil
}

func initJvm(chain33Config *chain33Types.Chain33Config) {
	if jvm_init_alreay {
		return
	}

	const_jdkPath := C.CString(jdkPath)
	defer C.free(unsafe.Pointer(const_jdkPath))

	const_jarPath := C.CString(Chain33LoaderJarPath)
	defer C.free(unsafe.Pointer(const_jarPath))

	result := C.JLI_Create_JVM(const_jdkPath, const_jarPath)
	if int(result) != JLI_SUCCESS {
		panic("Failed to init JLI_Init_JVM")
	}
	signal.Ignore(syscall.SIGPIPE)
	log.Info("JVM is created successfully")

	state.IsPara = chain33Config.IsPara()
	state.Title = chain33Config.GetTitle()
	consensusType = chain33Config.GetModuleConfig().Consensus.Name

	jvm_init_alreay = true
}

func buildJavaArgument(execPara []string) (C.int, **C.char) {
	argc := C.int(len(execPara))

	nil2dPtr := C.GetNil2dPtr()
	argv := (**C.char)(C.malloc(C.ulong(argc * C.GetPtrSize())))
	if argv == nil2dPtr {
		panic("Failed to malloc for argv")
	}
	//argv [argc]*C.char
	for i, para := range execPara {
		paraCstr := C.CString(para)
		C.SetPtr(argv, paraCstr, C.int(i))
	}
	return argc, argv
}

func freeArgument(argc C.int, argv **C.char) {
	C.FreeArgv(argc, argv)
}

//export SetQueryResult
func SetQueryResult(jvmgo *C.char, exceptionOccurred C.int, info **C.char, count, sizePtr C.int) C.int {
	exceptionOccur := false
	if Bool_TRUE == exceptionOccurred {
		exceptionOccur = true
	}
	var query []string
	tmpslice := (*[1 << 11]*C.char)(unsafe.Pointer(info))[:count:count]
	for i := 0; i < int(count); i++ {
		//ptr := (uintptr)(unsafe.Pointer(info)) + (uintptr)(int(sizePtr)*i)
		//infoGO := C.GoString(*(**C.char)(unsafe.Pointer(ptr)))
		infoGO := C.GoString(tmpslice[i])
		query = append(query, infoGO)
	}
	ForwardQueryResult(exceptionOccur, query, jvmgo)

	return 0
}

//用来保存txjvm或者是queryjvm中的env handle

//export BindTxQueryJVMEnvHandle
func BindTxQueryJVMEnvHandle(jvmGoHandle, envHandle *C.char) C.int {
	log.Debug("debug jvm panic: BindTxQueryJVMEnvHandle begin")
	defer log.Debug("debug jvm panic: BindTxQueryJVMEnvHandle end")
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	jvmExecutor := (*JVMExecutor)(unsafe.Pointer(jvmGoHandle))
	if !recordTxJVMEnv(jvmExecutor, envHandleUintptr) {
		return Bool_FALSE
	}
	return Bool_TRUE
}

/*
 * Account
 */
//export ExecFrozen
func ExecFrozen(from *C.char, amount C.long, envHandle *C.char) C.int {
	fromGoStr := C.GoString(from)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	if !execFrozen(fromGoStr, int64(amount), envHandleUintptr) {
		return Bool_FALSE
	}
	return Bool_TRUE
}

//export ExecActive
func ExecActive(from *C.char, amount C.long, envHandle *C.char) C.int {
	fromGoStr := C.GoString(from)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	if !execActive(fromGoStr, int64(amount), envHandleUintptr) {
		return Bool_FALSE
	}
	return Bool_TRUE
}

//export ExecTransfer
func ExecTransfer(from, to *C.char, amount C.long, envHandle *C.char) C.int {
	fromGoStr := C.GoString(from)
	toGoStr := C.GoString(to)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	if !execTransfer(fromGoStr, toGoStr, int64(amount), envHandleUintptr) {
		return Bool_FALSE
	}
	return Bool_TRUE
}

/*
 * blockchain misc
 */
//调用者负责释放返回指针内存
//export GetRandom
func GetRandom(envHandle *C.char) *C.char {
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	random, err := getRandom(envHandleUintptr)
	if nil != err {
		return nil
	}
	return C.CString(random)
}

//调用者负责释放返回指针内存
//export GetFrom
func GetFrom(envHandle *C.char) *C.char {
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	from := getFrom(envHandleUintptr)
	return C.CString(from)
}

//export GetCurrentHeight
func GetCurrentHeight(envHandle *C.char) C.long {
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	return C.long(getHeight(envHandleUintptr))
}

//export StopTransWithErrInfo
func StopTransWithErrInfo(errInfo *C.char, envHandle *C.char) C.int {
	errInfoStr := C.GoString(errInfo)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	stopTransWithErrInfo(errInfoStr, envHandleUintptr)

	return Bool_TRUE
}

/*
 * State DB
 */
//export SetState
func SetState(key *C.char, keySize C.int, value *C.char, valueSize C.int, envHandle *C.char) C.int {
	keySlice := C.GoBytes(unsafe.Pointer(key), keySize)
	valSlice := C.GoBytes(unsafe.Pointer(value), valueSize)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	if !stateDBSetState(keySlice, valSlice, envHandleUintptr) {
		return Bool_FALSE
	}
	return Bool_TRUE
}

//需要调用者释放内存
//export GetFromState
func GetFromState(key *C.char, keySize C.int, valueSize *C.int, envHandle *C.char) *C.char {
	keySlice := C.GoBytes(unsafe.Pointer(key), keySize)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	value := stateDBGetState(keySlice, envHandleUintptr)
	*valueSize = C.int(len(value))
	return (*C.char)(C.CBytes(value))
}

//export SetStateInStr
func SetStateInStr(key *C.char, value *C.char, envHandle *C.char) C.int {
	log.Debug("debug jvm panic: SetStateInStr begin")
	defer log.Debug("debug jvm panic: SetStateInStr end")
	keyStr := C.GoString(key)
	valueStr := C.GoString(value)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	if !stateDBSetState([]byte(keyStr), []byte(valueStr), envHandleUintptr) {
		return Bool_FALSE
	}
	return Bool_TRUE
}

//调用者负责释放返回指针内存
//export GetFromStateInStr
func GetFromStateInStr(key *C.char, size *C.int, envHandle *C.char) *C.char {
	log.Debug("debug jvm panic: GetFromStateInStr begin")
	defer log.Debug("debug jvm panic: GetFromStateInStr end")
	keyStr := C.GoString(key)
	if "" == keyStr {
		*size = C.int(0)
		return nil
	}
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	valueSlice := stateDBGetState([]byte(keyStr), envHandleUintptr)
	valSize := len(valueSlice)
	if 0 == valSize {
		*size = C.int(0)
		return nil
	}
	*size = C.int(valSize)
	return C.CString(string(valueSlice))
}

/*
 * Local DB
 */
//export SetLocal
func SetLocal(key *C.char, keySize C.int, value *C.char, valueSize C.int, envHandle *C.char) C.int {
	keySlice := C.GoBytes(unsafe.Pointer(key), keySize)
	valSlice := C.GoBytes(unsafe.Pointer(value), valueSize)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	if !setValue2Local(keySlice, valSlice, envHandleUintptr) {
		return Bool_FALSE
	}
	return Bool_TRUE
}

//export GetFromLocal
func GetFromLocal(key *C.char, keySize C.int, valueSize *C.int, envHandle *C.char) *C.char {
	keySlice := C.GoBytes(unsafe.Pointer(key), keySize)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	value := getValueFromLocal(keySlice, envHandleUintptr)
	*valueSize = C.int(len(value))
	return (*C.char)(C.CBytes(value))
}

//export SetLocalInStr
func SetLocalInStr(key *C.char, value *C.char, envHandle *C.char) C.int {
	log.Debug("debug jvm panic: SetLocalInStr begin")
	defer log.Debug("debug jvm panic: SetLocalInStr end")
	keyStr := C.GoString(key)
	valueStr := C.GoString(value)
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	if !setValue2Local([]byte(keyStr), []byte(valueStr), envHandleUintptr) {
		return Bool_FALSE
	}
	return Bool_TRUE
}

//调用者负责释放返回指针内存
//export GetFromLocalInStr
func GetFromLocalInStr(key *C.char, size *C.int, envHandle *C.char) *C.char {
	log.Debug("debug jvm panic: GetFromLocalInStr begin")
	defer log.Debug("debug jvm panic: GetFromLocalInStr end")
	keyStr := C.GoString(key)
	if "" == keyStr {
		*size = C.int(0)
		return nil
	}
	envHandleUintptr := uintptr(unsafe.Pointer(envHandle))
	valueSlice := getValueFromLocal([]byte(keyStr), envHandleUintptr)
	valSize := len(valueSlice)
	if 0 == valSize {
		*size = C.int(0)
		return nil
	}
	*size = C.int(valSize)
	return C.CString(string(valueSlice))
}

///////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

func recordTxJVMEnv(jvm *JVMExecutor, envHandle uintptr) bool {
	jvmsCached.Add(envHandle, jvm)
	_, ok := jvmsCached.Get(envHandle)
	return ok
}

func getJvmExector(envHandle uintptr) (*JVMExecutor, bool) {
	value, ok := jvmsCached.Get(envHandle)
	if !ok {
		log.Error("getJvmExector", "Failed to get JVMExecutor from lru cache with key", envHandle)
		return nil, false
	}

	jvmExecutor, ok := value.(*JVMExecutor)
	if !ok {
		log.Error("getJvmExector", "Failed to get JVMExecutor for query with key", envHandle)
		return nil, false
	}
	return jvmExecutor, true
}

/////////////////////////LocalDB interface//////////////////////////////////////////
func getValueFromLocal(key []byte, envHandle uintptr) []byte {
	log.Debug("Entering GetValueFromLocal", "key", string(key))
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		return nil
	}
	contractAddrgo := jvmExecutor.GetContractAddr()
	value := jvmExecutor.mStateDB.GetValueFromLocal(contractAddrgo, string(key), jvmExecutor.txHash)
	if 0 == len(value) {
		log.Debug("Entering Get GetValueFromLocal", "get null value for key", string(key))
		return nil
	}
	return value
}

func setValue2Local(key, value []byte, envHandle uintptr) bool {
	log.Debug("setValue2Local", "key", string(key), "value in string:", string(value),
		"value in slice:", value)
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		return false
	}
	contractAddrgo := jvmExecutor.GetContractAddr()
	return jvmExecutor.mStateDB.SetValue2Local(contractAddrgo, string(key), value, jvmExecutor.txHash)
}

func stateDBGetState(key []byte, envHandle uintptr) []byte {
	log.Debug("Entering StateDBGetState", "key", string(key))
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		log.Error("stateDBGetState", "Can't get jvmExecutor for key", string(key))
		return nil
	}
	contractAddrgo := jvmExecutor.GetContractAddr()
	value := jvmExecutor.mStateDB.GetState(contractAddrgo, string(key))
	if 0 == len(value) {
		log.Debug("StateDBGetState", "get null value for key", string(key))
		return nil
	}

	log.Debug("StateDBGetState Succeed to get value", "value in string", string(value), "value in slice", value)

	return value
}

func stateDBSetState(key, value []byte, envHandle uintptr) bool {
	log.Debug("StateDBSetStateCallback", "key", string(key), "value in string:", string(value),
		"value in slice:", value)
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		return false
	}
	contractAddrgo := jvmExecutor.GetContractAddr()
	return jvmExecutor.mStateDB.SetState(contractAddrgo, string(key), value)
}

////////////以下接口用于user.jvm.xxx合约内部转账/////////////////////////////
//必须要使用回传的envhandle获取jvm结构指针，否则存在java合约跨合约操作的安全性问题,
//比如在查询的时候，恶意发起数据库写（其中的账户操作就是）的操作，
func execFrozen(from string, amount int64, envHandle uintptr) bool {
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		return false
	}
	if nil == jvmExecutor || nil == jvmExecutor.mStateDB {
		log.Error("ExecFrozen failed due to nil handle", "pJvm", jvmExecutor,
			"pJvmMap[uint64(jvmIndex)].mStateDB", jvmExecutor.mStateDB)
		return jvmTypes.AccountOpFail
	}
	return jvmExecutor.mStateDB.ExecFrozen(jvmExecutor.tx, from, amount*jvmTypes.Coin_Precision)
}

// ExecActive 激活user.jvm.xxx合约addr上的部分余额
func execActive(from string, amount int64, envHandle uintptr) bool {
	log.Debug("Enter ExecActive", "from", from, "amount", amount)
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		log.Error("ExecActive", "Failed to getJvmExector with envHandle", envHandle)
		return jvmTypes.AccountOpFail
	}
	if nil == jvmExecutor || nil == jvmExecutor.mStateDB {
		log.Error("ExecActive failed due to nil handle", "pJvm", jvmExecutor,
			"pJvmMap[uint64(jvmIndex)].mStateDB", jvmExecutor.mStateDB)
		return jvmTypes.AccountOpFail
	}
	return jvmExecutor.mStateDB.ExecActive(jvmExecutor.tx, from, amount*jvmTypes.Coin_Precision)
}

// ExecTransfer transfer exec
func execTransfer(from, to string, amount int64, envHandle uintptr) bool {
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		return false
	}
	if nil == jvmExecutor || nil == jvmExecutor.mStateDB {
		log.Error("ExecTransfer failed due to nil handle", "pJvm", jvmExecutor,
			"pJvmMap[uint64(jvmIndex)].mStateDB", jvmExecutor.mStateDB)
		return jvmTypes.AccountOpFail
	}
	return jvmExecutor.mStateDB.ExecTransfer(jvmExecutor.tx, from, to, amount*jvmTypes.Coin_Precision)
}

// GetRandom 为jvm用户自定义合约提供随机数，该随机数是64位hash值,返回值为实际返回的长度
func getRandom(envHandle uintptr) (string, error) {
	return randomStr, nil
}

func getFrom(envHandle uintptr) string {
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		return ""
	}
	if nil == jvmExecutor || nil == jvmExecutor.tx {
		log.Error("GetFrom failed due to nil jvmExecutor or nil tx ", "pJvm", jvmExecutor)
		return ""
	}
	return jvmExecutor.tx.From()
}

func getHeight(envHandle uintptr) int64 {
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		return 0
	}
	return jvmExecutor.GetHeight()
}

func stopTransWithErrInfo(err string, envHandle uintptr) bool {
	jvmExecutor, ok := getJvmExector(envHandle)
	if !ok {
		return false
	}
	jvmExecutor.forceStopInfo.occurred = true
	jvmExecutor.forceStopInfo.info = errors.New(err)

	log.Info("StopTransWithErrInfo", "error info", err)

	return true
}

//forward the query result to the corresponding jvm
func ForwardQueryResult(exceptionOccurred bool, info []string, jvmgo *C.char) bool {
	queryResult := QueryResult{
		exceptionOccurred: exceptionOccurred,
		info:              info,
	}
	jvm := (*JVMExecutor)(unsafe.Pointer(jvmgo))
	jvm.queryChan <- queryResult
	log.Info("ForwardQueryResult get query result and forward it", "queryResult", queryResult)
	return true
}
