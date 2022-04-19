package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	chain33Types "github.com/33cn/chain33/types"
	ebrelayerChain33 "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/chain33"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/golang/protobuf/proto"
)

type TxCreateInfo struct {
	PrivateKey string
	Expire     string
	Note       string
	Fee        int64
	ParaName   string
	ChainID    int32
}

type Chain33OfflineTx struct {
	ContractAddr  string
	TxHash        string
	SignedRawTx   string
	OperationName string
	Interval      time.Duration
}

func CreateContractAndSign(txCreateInfo *TxCreateInfo, code, abi, parameter, contractName string) (string, []byte, error) {
	var action evmtypes.EVMContractAction
	bCode, err := common.FromHex(code)

	exector := types.GetExecName("evm", txCreateInfo.ParaName)
	to := address.ExecAddress(exector)

	if err != nil {
		return "", nil, errors.New(contractName + " parse evm code error " + err.Error())
	}
	action = evmtypes.EVMContractAction{Amount: 0, Code: bCode, GasLimit: 0, GasPrice: 0, Note: txCreateInfo.Note, Alias: contractName, ContractAddr: to}
	if parameter != "" {
		constructorPara := "constructor(" + parameter + ")"
		packData, err := evmAbi.PackContructorPara(constructorPara, abi)
		if err != nil {
			return "", nil, errors.New(contractName + " " + constructorPara + " Pack Contructor Para error:" + err.Error())
		}
		action.Code = append(action.Code, packData...)
	}

	data, txHash, err := CreateAndSignEvmTx(txCreateInfo.ChainID, &action, exector, txCreateInfo.PrivateKey, txCreateInfo.Expire, txCreateInfo.Fee)
	if err != nil {
		return "", nil, errors.New(contractName + " create contract error:" + err.Error())
	}

	return data, txHash, nil
}

func CreateAndSignEvmTx(chainID int32, action proto.Message, execer, privateKeyStr, expire string, fee int64) (string, []byte, error) {
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(action), Fee: fee, To: address.ExecAddress(execer), ChainID: chainID}

	expireInt64, err := types.ParseExpire(expire)
	if nil != err {
		return "", nil, err
	}

	if expireInt64 > types.ExpireBound {
		if expireInt64 < int64(time.Second*120) {
			expireInt64 = int64(time.Second * 120)
		}
		//用秒数来表示的时间
		tx.Expire = types.Now().Unix() + expireInt64/int64(time.Second)
	} else {
		tx.Expire = expireInt64
	}

	tx.Fee = int64(1e7)
	if tx.Fee < fee {
		tx.Fee += fee
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	tx.ChainID = chainID

	var driver secp256k1.Driver
	privateKeySli, err := common.FromHex(privateKeyStr)
	if nil != err {
		return "", nil, err
	}
	privateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		return "", nil, err
	}

	tx.Sign(types.SECP256K1, privateKey)
	txData := types.Encode(tx)
	dataStr := common.ToHex(txData)

	return dataStr, tx.Hash(), nil
}

func WriteContractFile(fileName string, content string) {
	err := ioutil.WriteFile(fileName, []byte(content), 0666)
	if err != nil {
		fmt.Println("Failed to write to file:", fileName)
	}
	fmt.Println("tx is written to file: ", fileName)
}

func CallContractAndSign(txCreateInfo *TxCreateInfo, action *evmtypes.EVMContractAction, contractAddr string) (string, []byte, error) {
	exector := types.GetExecName("evm", txCreateInfo.ParaName)
	data, txHash, err := CreateAndSignEvmTx(txCreateInfo.ChainID, action, exector, txCreateInfo.PrivateKey, txCreateInfo.Expire, txCreateInfo.Fee)
	if err != nil {
		return "", nil, errors.New(contractAddr + " call contract error:" + err.Error())
	}
	return data, txHash, nil
}

func ParseFileInJson(file string, result interface{}) error {
	_, err := os.Stat(file)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return json.Unmarshal(b, result)
}

func WriteToFileInJson(fileName string, content interface{}) {
	jbytes, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(fileName, jbytes, 0666)
	if err != nil {
		fmt.Println("Failed to write to file:", fileName)
	}
	fmt.Println("Txs are written to file", fileName)
}

func SendSignTxs2Chain33(filePath, rpcUrl string) {
	var rdata []*Chain33OfflineTx
	var retData []*Chain33OfflineTx
	err := ParseFileInJson(filePath, &rdata)
	if err != nil {
		fmt.Printf("parse file with error:%s, make sure file:%s exist \n", err.Error(), filePath)
		return
	}

	for _, deployInfo := range rdata {
		txhash, err := sendTransactionRpc(deployInfo.SignedRawTx, rpcUrl)
		if nil != err {
			fmt.Printf("Failed to send %s to chain33 due to error:%s \n", deployInfo.OperationName, err.Error())
			return
		}
		if deployInfo.Interval != 0 {
			time.Sleep(deployInfo.Interval)
		}

		countTime := 0
		for {
			result := ebrelayerChain33.GetTxStatusByHashesRpc(txhash, rpcUrl)
			//等待交易执行
			if ebTypes.Invalid_Chain33Tx_Status == result {
				time.Sleep(time.Second)

				countTime++
				// 超过2分钟 超时退出
				if countTime > 120 {
					fmt.Println("time out txhash: ", txhash)
					return
				}
				continue
			}
			if result != chain33Types.ExecOk {
				fmt.Println("Failed to send txhash: ", txhash, "result: ", result)
				return
			} else {
				break
			}
		}

		retData = append(retData, &Chain33OfflineTx{TxHash: txhash, ContractAddr: deployInfo.ContractAddr, OperationName: deployInfo.OperationName})
	}

	data, err := json.MarshalIndent(retData, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(data))
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
