package chain33

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"

	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	btcec_secp256k1 "github.com/btcsuite/btcd/btcec"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/33cn/chain33/common"
	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	chain33Crypto "github.com/33cn/chain33/common/crypto"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4chain33/generated"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common/math"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	ethSecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/golang/protobuf/proto"
)

//DeployPara ...
type DeployPara4Chain33 struct {
	Deployer       address.Address
	Operator       address.Address
	InitValidators []address.Address
	InitPowers     []*big.Int
}

type DeployResult struct {
	Address address.Address
	TxHash  string
}

type X2EthDeployResult struct {
	BridgeRegistry *DeployResult
	BridgeBank     *DeployResult
	EthereumBridge *DeployResult
	Valset         *DeployResult
	Oracle         *DeployResult
}

var chain33txLog = log.New("module", "chain33_txs")
var chainID int32

func setChainID(chainID4Chain33 int32) {
	chainID = chainID4Chain33
}

func createEvmTx(privateKey chain33Crypto.PrivKey, action proto.Message, execer, to string, fee int64) string {
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(action), Fee: fee, To: to, ChainID: chainID}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()

	tx.Sign(types.SECP256K1, privateKey)
	txData := types.Encode(tx)
	dataStr := common.ToHex(txData)
	return dataStr
}

func relayEvmTx2Chain33(privateKey chain33Crypto.PrivKey, claim *ebrelayerTypes.EthBridgeClaim, parameter, oracleAddr, chainName string, rpcURLs []string) (string, error) {
	note := fmt.Sprintf("relay with type:%s, chain33-receiver:%s, ethereum-sender:%s, symbol:%s, amout:%s, ethTxHash:%s",
		events.ClaimType(claim.ClaimType).String(), claim.Chain33Receiver, claim.EthereumSender, claim.Symbol, claim.Amount, claim.EthTxHash)
	_, packData, err := evmAbi.Pack(parameter, generated.OracleABI, false)
	if nil != err {
		chain33txLog.Info("relayEvmTx2Chain33", "Failed to do abi.Pack due to:", err.Error())
		return "", ebrelayerTypes.ErrPack
	}

	action := evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: note, Para: packData, ContractAddr: oracleAddr}

	//TODO: 交易费超大问题需要调查，hezhengjun on 20210420
	feeInt64 := int64(1 * 1e6)
	wholeEvm := getExecerName(chainName)
	toAddr := address.ExecAddress(wholeEvm)
	//name表示发给哪个执行器
	data := createEvmTx(privateKey, &action, wholeEvm, toAddr, feeInt64)
	params := rpctypes.RawParm{
		Token: "BTY",
		Data:  data,
	}

	// 存在发送交易成功, 但是由于chain33节点崩溃, 导致交易没有打包, 所以向多个节点发送交易, 提高可靠性
	var txHash string
	bExecuted := false
	for _, rpcURL := range rpcURLs {
		var txhash string
		ctx := jsonclient.NewRPCCtx(rpcURL, "Chain33.SendTransaction", params, &txhash)
		_, err = ctx.RunResult()

		// 如果成功 记录这笔哈希
		if err == nil {
			bExecuted = true
			txHash = txhash
		}
	}
	if !bExecuted {
		return txHash, err
	}
	return txHash, nil
}

func getExecerName(name string) string {
	var ret string
	names := strings.Split(name, ".")
	for _, v := range names {
		if v != "" {
			ret = ret + v + "."
		}
	}
	ret += "evm"
	return ret
}

func queryTxsByHashesRes(arg interface{}) (interface{}, error) {
	var receipt *rpctypes.ReceiptDataResult
	for _, v := range arg.(*rpctypes.TransactionDetails).Txs {
		if v == nil {
			continue
		}
		receipt = v.Receipt
		if nil != receipt {
			return receipt.Ty, nil
		}
	}
	return nil, nil
}

func GetTxStatusByHashesRpc(txhex, rpcLaddr string) int32 {
	hashesArr := strings.Split(txhex, " ")
	params2 := rpctypes.ReqHashes{
		Hashes: hashesArr,
	}

	var res rpctypes.TransactionDetails
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.GetTxByHashes", params2, &res)
	ctx.SetResultCb(queryTxsByHashesRes)
	result, err := ctx.RunResult()
	if err != nil || result == nil {
		return ebrelayerTypes.Invalid_Chain33Tx_Status
	}
	return result.(int32)
}

func getTxByHashesRpc(txhex, rpcLaddr string) (string, error) {
	hashesArr := strings.Split(txhex, " ")
	params2 := rpctypes.ReqHashes{
		Hashes: hashesArr,
	}

	var res rpctypes.TransactionDetails
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.GetTxByHashes", params2, &res)
	ctx.SetResultCb(queryTxsByHashesRes)
	result, err := ctx.RunResult()
	if err != nil || result == nil {
		return "", err
	}
	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getContractAddr(caller, txhex string) address.Address {
	return *address.BytesToBtcAddress(address.NormalVer,
		address.ExecPubKey(caller+ethcommon.Bytes2Hex(common.HexToHash(txhex).Bytes())))
}

func deploySingleContract(code []byte, abi, constructorPara, contractName, paraChainName, deployer, rpcLaddr string) (string, error) {
	note := "deploy " + contractName
	exector := paraChainName + "evm"
	to := address.ExecAddress(exector)

	var action evmtypes.EVMContractAction
	action = evmtypes.EVMContractAction{Amount: 0, Code: code, GasLimit: 0, GasPrice: 0, Note: note, Alias: contractName, ContractAddr: to}
	if constructorPara != "" {
		packData, err := evmAbi.PackContructorPara(constructorPara, abi)
		if err != nil {
			return "", errors.New(contractName + " " + constructorPara + " Pack Contructor Para error:" + err.Error())
		}
		action.Code = append(action.Code, packData...)
	}

	data, err := createSignedEvmTx(&action, exector, deployer, rpcLaddr, to)
	if err != nil {
		return "", errors.New(contractName + " create contract error:" + err.Error())
	}

	txhex, err := sendTransactionRpc(data, rpcLaddr)
	if err != nil {
		return "", errors.New(contractName + " send transaction error:" + err.Error())
	}
	chain33txLog.Info("deploySingleContract", "Deploy contract for", contractName, " with tx hash:", txhex)
	return txhex, nil
}

func createSignedEvmTx(action proto.Message, execer, caller, rpcLaddr, to string) (string, error) {
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(action), Fee: int64(1e8), To: to, ChainID: chainID}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	txHex := types.Encode(tx)
	rawTx := hex.EncodeToString(txHex)

	unsignedTx := &types.ReqSignRawTx{
		Addr:   caller,
		TxHex:  rawTx,
		Fee:    tx.Fee,
		Expire: "120s",
	}

	var res string
	client, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		chain33txLog.Error("createSignedEvmTx", "jsonclient.NewJSONClient", err.Error())
		return "", err
	}
	err = client.Call("Chain33.SignRawTx", unsignedTx, &res)
	if err != nil {
		chain33txLog.Error("createSignedEvmTx", "Chain33.SignRawTx", err.Error())
		return "", err
	}

	return res, nil
}

func sendTransactionRpc(data, rpcLaddr string) (string, error) {
	params := rpctypes.RawParm{
		Data: data,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.SendTransaction", params, nil)
	var txhex string
	rpc, err := jsonclient.NewJSONClient(ctx.Addr)
	if err != nil {
		return "", err
	}

	err = rpc.Call(ctx.Method, ctx.Params, &txhex)
	if err != nil {
		return "", err
	}

	return txhex, nil
}

func sendTx2Evm(parameter []byte, rpcURL, evmContractAddr, chainName, caller string) (string, error) {
	note := fmt.Sprintf("sendTx2Evm by caller:%s", caller)

	action := evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: note, Para: parameter, ContractAddr: evmContractAddr}
	wholeEvm := chainName + "evm"
	toAddr := address.ExecAddress(wholeEvm)
	data, err := createSignedEvmTx(&action, wholeEvm, caller, rpcURL, toAddr)
	if err != nil {
		return "", errors.New(toAddr + " createSignedEvmTx error:" + err.Error())
	}

	txhex, err := sendTransactionRpc(data, rpcURL)
	if err != nil {
		return "", errors.New(toAddr + " send transaction error:" + err.Error())
	}
	return txhex, nil
}

func approve(privateKey chain33Crypto.PrivKey, contractAddr, spender, chainName, rpcURL string, amount int64) (string, error) {
	note := fmt.Sprintf("approve for spender:%s, amount:%d", spender, amount)

	//approve(address spender, uint256 amount)
	parameter := fmt.Sprintf("approve(%s, %d)", spender, amount)
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeTokenABI, false)
	if nil != err {
		chain33txLog.Info("approve", "Failed to do abi.Pack due to:", err.Error())
		return "", err
	}
	return sendEvmTx(privateKey, contractAddr, chainName, rpcURL, note, packData, 0)
}

func burn(privateKey chain33Crypto.PrivKey, contractAddr, ethereumReceiver, ethereumTokenAddress, chainName, rpcURL string, amount int64) (string, error) {
	//    function burnBridgeTokens(
	//        bytes memory _ethereumReceiver,
	//        address _ethereumTokenAddress,
	//        uint256 _amount
	//    )
	parameter := fmt.Sprintf("burnBridgeTokens(%s, %s, %d)", ethereumReceiver, ethereumTokenAddress, amount)
	note := parameter
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeBankABI, false)
	if nil != err {
		chain33txLog.Info("burn", "Failed to do abi.Pack due to:", err.Error())
		return "", err
	}

	return sendEvmTx(privateKey, contractAddr, chainName, rpcURL, note, packData, 0)
}

func lockBty(privateKey chain33Crypto.PrivKey, contractAddr, ethereumReceiver, chainName, rpcURL string, amount int64) (string, error) {
	//function lock(
	//	bytes memory _recipient,
	//	address _token,
	//	uint256 _amount
	//)
	parameter := fmt.Sprintf("lock(%s, %s, %d)", ethereumReceiver, ebrelayerTypes.BTYAddrChain33, amount)
	note := parameter
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeBankABI, false)
	if nil != err {
		chain33txLog.Info("setOracle", "Failed to do abi.Pack due to:", err.Error())
		return "", ebrelayerTypes.ErrPack
	}
	return sendEvmTx(privateKey, contractAddr, chainName, rpcURL, note, packData, amount)
}

func sendEvmTx(privateKey chain33Crypto.PrivKey, contractAddr, chainName, rpcURL, note string, parameter []byte, amount int64) (string, error) {
	action := evmtypes.EVMContractAction{Amount: uint64(amount), GasLimit: 0, GasPrice: 0, Note: note, Para: parameter, ContractAddr: contractAddr}

	feeInt64 := int64(1e7)
	wholeEvm := getExecerName(chainName)
	toAddr := address.ExecAddress(wholeEvm)
	//name表示发给哪个执行器
	data := createEvmTx(privateKey, &action, wholeEvm, toAddr, feeInt64)
	params := rpctypes.RawParm{
		Token: "",
		Data:  data,
	}
	var txhash string

	ctx := jsonclient.NewRPCCtx(rpcURL, "Chain33.SendTransaction", params, &txhash)
	_, err := ctx.RunResult()
	return txhash, err
}

func burnAsync(ownerPrivateKeyStr, tokenAddrstr, ethereumReceiver string, amount int64, bridgeBankAddr string, chainName, rpcURL string) (string, error) {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(ownerPrivateKeyStr)
	if nil != err {
		return "", err
	}
	ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return "", err
	}

	approveTxHash, err := approve(ownerPrivateKey, tokenAddrstr, bridgeBankAddr, chainName, rpcURL, amount)
	if err != nil {
		chain33txLog.Error("BurnAsync", "failed to send approve tx due to:", err.Error())
		return "", err
	}
	chain33txLog.Debug("BurnAsync", "approve with tx hash", approveTxHash)

	//privateKey chain33Crypto.PrivKey, contractAddr, ethereumReceiver, ethereumTokenAddress, chainName, rpcURL string, amount int6
	burnTxHash, err := burn(ownerPrivateKey, bridgeBankAddr, ethereumReceiver, tokenAddrstr, chainName, rpcURL, amount)
	if err != nil {
		chain33txLog.Error("BurnAsync", "failed to send burn tx due to:", err.Error())
		return "", err
	}
	chain33txLog.Debug("BurnAsync", "burn with tx hash", burnTxHash)

	return burnTxHash, err
}

func lockAsync(ownerPrivateKeyStr, ethereumReceiver string, amount int64, bridgeBankAddr string, chainName, rpcURL string) (string, error) {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(ownerPrivateKeyStr)
	if nil != err {
		return "", err
	}
	ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return "", err
	}

	//privateKey chain33Crypto.PrivKey, contractAddr, ethereumReceiver, ethereumTokenAddress, chainName, rpcURL string, amount int6
	lockBtyTxHash, err := lockBty(ownerPrivateKey, bridgeBankAddr, ethereumReceiver, chainName, rpcURL, amount)
	if err != nil {
		chain33txLog.Error("lockBty", "failed to send approve tx due to:", err.Error())
		return "", err
	}
	chain33txLog.Debug("lockBty", "lockBty with tx hash", lockBtyTxHash)

	return "", err
}

func setupMultiSign(ownerPrivateKeyStr, contractAddr, chainName, rpcURL string, owners []string) (string, error) {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(ownerPrivateKeyStr)
	if nil != err {
		return "", err
	}
	ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return "", err
	}
	//function setup(
	//	address[] calldata _owners,
	//	uint256 _threshold,
	//	address to,
	//	bytes calldata data,
	//	address fallbackHandler,
	//	address paymentToken,
	//	uint256 payment,
	//	address payable paymentReceiver
	//)
	parameter := "setup(["
	parameter += fmt.Sprintf("%s", owners[0])
	for _, owner := range owners[1:] {
		parameter += fmt.Sprintf(",%s", owner)
	}
	parameter += "], "
	parameter += fmt.Sprintf("%d, %s, 0102, %s, %s, 0, %s)", len(owners), ebrelayerTypes.BTYAddrChain33, ebrelayerTypes.BTYAddrChain33, ebrelayerTypes.BTYAddrChain33, ebrelayerTypes.BTYAddrChain33)
	note := parameter
	_, packData, err := evmAbi.Pack(parameter, generated.GnosisSafeABI, false)
	if nil != err {
		chain33txLog.Info("burn", "Failed to do abi.Pack due to:", err.Error())
		return "", err
	}

	return sendEvmTx(ownerPrivateKey, contractAddr, chainName, rpcURL, note, packData, 0)
}

func safeTransfer(ownerPrivateKeyStr, mulSign, chainName, rpcURL, receiver, token string, privateKeys []string, amount float64) (string, error) {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(ownerPrivateKeyStr)
	if nil != err {
		return "", err
	}
	ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return "", err
	}
	//function execTransaction(
	//	address to,
	//	uint256 value,
	//	bytes memory data,
	//	Enum.Operation operation,
	//	uint256 safeTxGas,
	//	uint256 baseGas,
	//	uint256 gasPrice,
	//	address gasToken,
	//	address payable refundReceiver,
	//	bytes memory signatures
	//)

	//对于平台币转账，这个data只是个占位符，没有作用
	dataStr := "0x"
	safeTxGas := int64(10 * 10000)
	baseGas := 0
	gasPrice := 0
	valueStr := utils.ToWei(amount, 8).String()
	//如果是bty转账,则直接将to地址设置为receiver,而如果是ERC20转账，则需要将其设置为token地址
	to := receiver
	//如果是erc20转账，则需要构建data数据
	if "" != token {
		parameter := fmt.Sprintf("transfer(%s, %s)", receiver, utils.ToWei(amount, 8).String())
		_, data, err := evmAbi.Pack(parameter, erc20.ERC20ABI, false)
		if err != nil {
			return "", err
		}
		chain33txLog.Info("safeTransfer", "evmAbi.Pack with parameter", parameter,
			"data", common.ToHex(data))
		//对于其他erc20资产，直接将其设置为0
		valueStr = "0"
		to = token
		dataStr = common.ToHex(data)
	}

	//获取nonce
	nonce := getMulSignNonce(mulSign, rpcURL)
	//构造getTransactionHash参数
	//function getTransactionHash(
	//	address to,
	//	uint256 value,
	//	bytes memory data,
	//	Enum.Operation operation,
	//	uint256 safeTxGas,
	//	uint256 baseGas,
	//	uint256 gasPrice,
	//	address gasToken,
	//	address refundReceiver,
	//	uint256 _nonce
	//)
	parameter2getHash := fmt.Sprintf("getTransactionHash(%s, %s, %s, 0, %d, %d, %d, %s, %s, %d)", to, valueStr, dataStr,
		safeTxGas, baseGas, gasPrice,
		ebrelayerTypes.NilAddrChain33, ebrelayerTypes.NilAddrChain33, nonce)

	chain33txLog.Info("safeTransfer", "parameter2getHash", parameter2getHash)
	result := query(mulSign, parameter2getHash, mulSign, rpcURL, generated.GnosisSafeABI)
	if nil == result {
		return "", ebrelayerTypes.ErrGetTransactionHash
	}
	contentHashArray := result.([32]byte)
	contentHash := contentHashArray[:]
	chain33txLog.Info("safeTransfer", "contentHash", common.ToHex(contentHash))
	var sigs []byte
	for i, privateKey := range privateKeys {
		chain33txLog.Info("safeTransfer", "index", i, "privateKey", privateKey)
		var driver secp256k1.Driver
		privateKeySli, err := chain33Common.FromHex(privateKey)
		if nil != err {
			return "", err
		}
		ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
		if nil != err {
			return "", err
		}
		temp, _ := btcec_secp256k1.PrivKeyFromBytes(btcec_secp256k1.S256(), ownerPrivateKey.Bytes())
		privateKey4Chain33_ecdsa := temp.ToECDSA()

		sig, err := ethSecp256k1.Sign(contentHash, math.PaddedBigBytes(privateKey4Chain33_ecdsa.D, 32))
		if nil != err {
			chain33txLog.Error("safeTransfer", "Failed to do ethSecp256k1.Sign to:", err.Error())
			return "", err
		}

		sig[64] += 27
		chain33txLog.Info("safeTransfer", "single signature", common.ToHex(sig))
		sigs = append(sigs, sig...)
	}

	//构造execTransaction参数
	parameter2Exec := fmt.Sprintf("execTransaction(%s, %s, %s, 0, %d, %d, %d, %s, %s, %s)", to, valueStr, dataStr,
		safeTxGas, baseGas, gasPrice,
		ebrelayerTypes.NilAddrChain33, ebrelayerTypes.NilAddrChain33, common.ToHex(sigs))
	note := parameter2Exec
	chain33txLog.Info("safeTransfer", "parameter2Exec", parameter2Exec)
	_, packData, err := evmAbi.Pack(parameter2Exec, generated.GnosisSafeABI, false)
	if nil != err {
		chain33txLog.Error("safeTransfer", "Failed to do abi.Pack due to:", err.Error())
		return "", err
	}

	return sendEvmTx(ownerPrivateKey, mulSign, chainName, rpcURL, note, packData, 0)
}

func recoverContractAddrFromRegistry(bridgeRegistry, rpcLaddr string) (oracle, bridgeBank string) {
	parameter := fmt.Sprint("oracle()")

	result := query(bridgeRegistry, parameter, bridgeRegistry, rpcLaddr, generated.BridgeRegistryABI)
	if nil == result {
		return "", ""
	}
	oracle = result.(string)

	parameter = fmt.Sprint("bridgeBank()")
	result = query(bridgeRegistry, parameter, bridgeRegistry, rpcLaddr, generated.BridgeRegistryABI)
	if nil == result {
		return "", ""
	}
	bridgeBank = result.(string)
	return
}

func getLockedTokenAddress(bridgeBank, symbol, rpcLaddr string) string {
	parameter := fmt.Sprintf("getLockedTokenAddress(%s)", symbol)

	result := query(bridgeBank, parameter, bridgeBank, rpcLaddr, generated.BridgeBankABI)
	if nil == result {
		return ""
	}
	return result.(string)
}

func getBridgeToken2address(bridgeBank, symbol, rpcLaddr string) string {
	parameter := fmt.Sprintf("getToken2address(%s)", symbol)

	result := query(bridgeBank, parameter, bridgeBank, rpcLaddr, generated.BridgeBankABI)
	if nil == result {
		return ""
	}
	return result.(string)
}

func getMulSignNonce(mulsign, rpcLaddr string) int64 {
	parameter := fmt.Sprintf("nonce()")

	result := query(mulsign, parameter, mulsign, rpcLaddr, generated.GnosisSafeABI)
	if nil == result {
		return 0
	}
	nonce := result.(*big.Int)
	return nonce.Int64()
}

func Query(contractAddr, input, caller, rpcLaddr, abiData string) interface{} {
	return query(contractAddr, input, caller, rpcLaddr, abiData)
}

func query(contractAddr, input, caller, rpcLaddr, abiData string) interface{} {
	methodName, packedinput, err := evmAbi.Pack(input, abiData, true)
	if err != nil {
		chain33txLog.Debug("query", "Failed to do para pack due to", err.Error())
		return nil
	}

	var req = evmtypes.EvmQueryReq{Address: contractAddr, Input: common.ToHex(packedinput), Caller: caller}
	var resp evmtypes.EvmQueryResp
	query := sendQuery(rpcLaddr, "Query", &req, &resp)

	if !query {
		return nil
	}
	_, err = json.MarshalIndent(&resp, "", "  ")
	if err != nil {
		fmt.Println(resp.String())
		return nil
	}

	data, err := common.FromHex(resp.RawData)
	if nil != err {
		fmt.Println("common.FromHex failed due to:", err.Error())
	}

	outputs, err := evmAbi.Unpack(data, methodName, abiData)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "unpack evm return error", err)
	}

	if len(outputs) < 1 {
		fmt.Println("outputs len = ", len(outputs))
		return nil
	}
	chain33txLog.Debug("query", "outputs", outputs)

	return outputs[0].Value
}

func sendQuery(rpcAddr, funcName string, request types.Message, result proto.Message) bool {
	params := rpctypes.Query4Jrpc{
		Execer:   "evm",
		FuncName: funcName,
		Payload:  types.MustPBToJSON(request),
	}

	jsonrpc, err := jsonclient.NewJSONClient(rpcAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	err = jsonrpc.Call("Chain33.Query", params, result)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	return true
}

func withdrawAsync(ownerPrivateKeyStr, tokenAddrstr, ethereumReceiver string, amount int64, bridgeBankAddr string, chainName, rpcURL string) (string, error) {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(ownerPrivateKeyStr)
	if nil != err {
		return "", err
	}
	ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return "", err
	}

	approveTxHash, err := approve(ownerPrivateKey, tokenAddrstr, bridgeBankAddr, chainName, rpcURL, amount)
	if err != nil {
		chain33txLog.Error("withdrawAsync", "failed to send approve tx due to:", err.Error())
		return "", err
	}
	chain33txLog.Debug("withdrawAsync", "approve with tx hash", approveTxHash)

	withdrawTxHash, err := withdrawViaProxy(ownerPrivateKey, bridgeBankAddr, ethereumReceiver, tokenAddrstr, chainName, rpcURL, amount)
	if err != nil {
		chain33txLog.Error("withdrawAsync", "failed to send withdraw tx due to:", err.Error())
		return "", err
	}
	chain33txLog.Debug("withdrawAsync", "withdraw with tx hash", withdrawTxHash)

	return withdrawTxHash, err
}

func withdrawViaProxy(privateKey chain33Crypto.PrivKey, contractAddr, ethereumReceiver, ethereumTokenAddress, chainName, rpcURL string, amount int64) (string, error) {
	//function withdrawViaProxy(
	//	bytes memory _ethereumReceiver,
	//	address _bridgeTokenAddress,
	//	uint256 _amount
	//)
	parameter := fmt.Sprintf("withdrawViaProxy(%s, %s, %d)", ethereumReceiver, ethereumTokenAddress, amount)
	note := parameter
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeBankABI, false)
	if nil != err {
		chain33txLog.Info("withdraw", "Failed to do abi.Pack due to:", err.Error())
		return "", err
	}

	return sendEvmTx(privateKey, contractAddr, chainName, rpcURL, note, packData, 0)
}

func burnWithIncreaseAsync(ownerPrivateKeyStr, tokenAddrstr, ethereumReceiver string, amount int64, bridgeBankAddr string, chainName, rpcURL string) (string, error) {
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(ownerPrivateKeyStr)
	if nil != err {
		return "", err
	}
	ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return "", err
	}

	approveTxHash, err := increaseApprove(ownerPrivateKey, tokenAddrstr, bridgeBankAddr, chainName, rpcURL, amount)
	if err != nil {
		chain33txLog.Error("burnWithIncreaseAsync", "failed to send approve tx due to:", err.Error())
		return "", err
	}
	chain33txLog.Debug("burnWithIncreaseAsync", "approve with tx hash", approveTxHash)

	//privateKey chain33Crypto.PrivKey, contractAddr, ethereumReceiver, ethereumTokenAddress, chainName, rpcURL string, amount int6
	burnTxHash, err := burn(ownerPrivateKey, bridgeBankAddr, ethereumReceiver, tokenAddrstr, chainName, rpcURL, amount)
	if err != nil {
		chain33txLog.Error("burnWithIncreaseAsync", "failed to send burn tx due to:", err.Error())
		return "", err
	}
	chain33txLog.Debug("burnWithIncreaseAsync", "burn with tx hash", burnTxHash)

	return burnTxHash, err
}

func increaseApprove(privateKey chain33Crypto.PrivKey, contractAddr, spender, chainName, rpcURL string, amount int64) (string, error) {
	note := fmt.Sprintf("increaseAllowance for spender:%s, amount:%d", spender, amount)

	//approve(address spender, uint256 amount)
	parameter := fmt.Sprintf("increaseAllowance(%s, %d)", spender, amount)
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeTokenABI, false)
	if nil != err {
		chain33txLog.Info("increaseAllowance", "Failed to do abi.Pack due to:", err.Error())
		return "", err
	}
	return sendEvmTx(privateKey, contractAddr, chainName, rpcURL, note, packData, 0)
}
