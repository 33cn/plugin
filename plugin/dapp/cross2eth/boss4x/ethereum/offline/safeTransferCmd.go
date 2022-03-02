package offline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"

	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	eoff "github.com/33cn/plugin/plugin/dapp/dex/boss/deploy/ethereum/offline"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

type safeTxData struct {
	Content       []byte //json.RawMessage
	To            common.Address
	Value         *big.Int
	TransferData  []byte //json.RawMessage
	GasLimit      int64
	GasPrice      *big.Int
	Nonce         uint64
	TxNonce       uint64
	CrontractAddr string
	SendAddr      string
	name          string
}

/*
./boss4x ethereum offline multisign_transfer_prepare -a 3 -r 0xC65B02a22B714b55D708518E2426a22ffB79113d -c 0xbf271b2B23DA4fA8Dc93Ce86D27dd09796a7Bf54 -d 0xd9dab021e74ecf475788ed7b61356056b2095830
./boss4x ethereum offline sign_multisign_tx -k 0x5e8aadb91eaa0fce4df0bcc8bd1af9e703a1d6db78e7a4ebffd6cf045e053574,0x0504bcb22b21874b85b15f1bfae19ad62fc2ad89caefc5344dc669c57efa60db,0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2,0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697
./boss4x ethereum offline create_multisign_tx
./boss4x ethereum offline sign -f create_multisign_tx.txt -k c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b
./boss4x ethereum offline send -f deploysigntxs.txt
*/

// PrepareCreateMultisignTransferTxCmd 预备创建一个多签转帐交易 在线
func PrepareCreateMultisignTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign_transfer_prepare", //first step
		Short: "prepare create multisign transfer tx",
		Run:   prepareCreateMultisignTransferTx,
	}
	addPrepareCreateMultisignTransferTxFlags(cmd)
	return cmd
}

func addPrepareCreateMultisignTransferTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("sendAddr", "d", "", "send addr")
	_ = cmd.MarkFlagRequired("sendAddr")
	cmd.Flags().StringP("contract", "c", "", "mulSignAddr contract address")
	_ = cmd.MarkFlagRequired("contract")
	cmd.Flags().StringP("receiver", "r", "", "receive address")
	_ = cmd.MarkFlagRequired("receiver")
	cmd.Flags().Float64P("amount", "a", 0, "amount to transfer")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("token", "t", "", "erc20 address,not need to set for ETH(optional)")
}

func prepareCreateMultisignTransferTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	receiver, _ := cmd.Flags().GetString("receiver")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	sendAddr, _ := cmd.Flags().GetString("sendAddr")
	multiSignAddrstr, _ := cmd.Flags().GetString("contract")

	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("Dial Err:", err)
		return
	}

	gnosisSafeAddr := common.HexToAddress(multiSignAddrstr)
	gnosisSafeInt, err := gnosis.NewGnosisSafe(gnosisSafeAddr, client)
	if nil != err {
		fmt.Println("NewGnosisSafe Err:", err)
		return
	}
	AddressZero := common.HexToAddress(ebTypes.EthNilAddr)

	toAddr := common.HexToAddress(receiver)
	sendData := []byte{'0', 'x'}
	baseGas := big.NewInt(0)
	gasPrice := big.NewInt(0)
	value := big.NewInt(0)
	safeTxGas := big.NewInt(10 * 10000)

	opts := &bind.CallOpts{
		From:    common.HexToAddress(sendAddr),
		Context: context.Background(),
	}

	txNonce, err := client.NonceAt(context.Background(), common.HexToAddress(sendAddr), nil)
	if err != nil {
		fmt.Println("NonceAt Err:", err)
		return
	}

	if tokenAddr == "" {
		realAmount := utils.ToWei(amount, 18)
		value, _ = value.SetString(utils.TrimZeroAndDot(realAmount.String()), 10)
	} else {
		toAddr = common.HexToAddress(tokenAddr)
		erc20Abi, err := abi.JSON(strings.NewReader(erc20.ERC20ABI))
		if err != nil {
			fmt.Println("strings.NewReader(erc20.ERC20ABI) Err:", err)
			return
		}

		tokenInstance, err := erc20.NewERC20(toAddr, client)
		if err != nil {
			fmt.Println("NewERC20 Err:", err)
			return
		}
		decimals, err := tokenInstance.Decimals(opts)
		if err != nil {
			fmt.Println("Decimals Err:", err)
			return
		}

		realAmount := utils.ToWei(amount, int64(decimals))
		value, _ = value.SetString(utils.TrimZeroAndDot(realAmount.String()), 10)
		sendData, err = erc20Abi.Pack("transfer", common.HexToAddress(receiver), value)
		if err != nil {
			fmt.Println("Pack Err:", err)
			return
		}
		//对于erc20这种方式 最后需要将其设置为0
		value = big.NewInt(0)
	}

	contracAddr := common.HexToAddress(multiSignAddrstr)
	var msg ethereum.CallMsg
	msg.Data = sendData
	msg.From = common.HexToAddress(sendAddr)
	msg.To = &contracAddr
	msg.Value = big.NewInt(0)
	//估算gas
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		fmt.Println("EstimateGas Err:", err)
		return
	}

	if gasLimit < 100*10000 {
		gasLimit = 100 * 10000
	}

	nonce, err := gnosisSafeInt.Nonce(opts)
	if err != nil {
		fmt.Println("NewGnosisSafe Err:", err)
		return
	}
	signContent, err := gnosisSafeInt.GetTransactionHash(opts, toAddr, value, sendData, 0,
		safeTxGas, baseGas, gasPrice, AddressZero, AddressZero, nonce)
	if err != nil {
		fmt.Println("GetTransactionHash Err:", err)
		return
	}

	var txdata safeTxData
	txdata.Content = signContent[:]
	txdata.TransferData = sendData
	txdata.GasLimit = int64(gasLimit)
	txdata.GasPrice = gasPrice
	txdata.To = toAddr
	txdata.CrontractAddr = multiSignAddrstr
	txdata.SendAddr = sendAddr
	txdata.Value = value
	txdata.TxNonce = txNonce
	txdata.name = "multisign_transfer_prepare"
	writeToFile(txdata.name+".txt", txdata)
}

// PreliminarySignMultisignTransferTxCmd 多签转帐交易 多签多个地址签名 离线
func PreliminarySignMultisignTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign_multisign_tx", //first step
		Short: "sign multisign tx",
		Run:   preliminarySignMultisignTransferTx,
	}
	addPreliminarySignMultisignTransferTxFlag(cmd)
	return cmd
}

func addPreliminarySignMultisignTransferTxFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "multisign_transfer_prepare.txt", "tx file, default: multisign_transfer_prepare.txt")
	cmd.Flags().StringP("keys", "k", "", "owners' private key, separated by ','")
	_ = cmd.MarkFlagRequired("keys")
}

//签名交易
func preliminarySignMultisignTransferTx(cmd *cobra.Command, _ []string) {
	txFilePath, _ := cmd.Flags().GetString("file")
	keys, _ := cmd.Flags().GetString("keys")
	privateKeys := strings.Split(keys, ",")

	var txinfo safeTxData
	err := paraseFile(txFilePath, &txinfo)
	if err != nil {
		fmt.Println("paraseFile Err:", err)
		return
	}

	sigs, err := buildSigs(txinfo.Content, privateKeys)
	if err != nil {
		fmt.Println("buildSigs Err:", err)
		return
	}
	txinfo.Content = sigs
	writeToFile("sign_multisign_tx.txt", txinfo)
}

// CreateMultisignTransferTxCmd 创建多签转帐交易
func CreateMultisignTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_multisign_tx",
		Short: "create multisign transfer tx",
		Run:   CreateMultisignTransferTx,
	}
	addCreateMultisignTransferTxFlags(cmd)
	return cmd
}

func addCreateMultisignTransferTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "sign_multisign_tx.txt", "tx file, default: sign_multisign_tx.txt")
}

func CreateMultisignTransferTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")
	txFilePath, _ := cmd.Flags().GetString("file")

	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("Dial Err:", err)
		return
	}

	var txinfo safeTxData
	err = paraseFile(txFilePath, &txinfo)
	if err != nil {
		fmt.Println("paraseFile Err:", err)
		return
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Println("SuggestGasPrice Err:", err)
		return
	}

	gnoAbi, err := abi.JSON(strings.NewReader(gnosis.GnosisSafeABI))
	if err != nil {
		fmt.Println("JSON Err:", err)
		return
	}
	zeroAddr := common.HexToAddress(ebTypes.EthNilAddr)
	safeTxGas := big.NewInt(10 * 10000)

	gnoData, err := gnoAbi.Pack("execTransaction", txinfo.To, txinfo.Value, txinfo.TransferData, uint8(0),
		safeTxGas, big.NewInt(0), gasPrice, zeroAddr, zeroAddr, txinfo.Content)
	if err != nil {
		fmt.Println("Pack execTransaction Err:", err)
		return
	}

	CreateTxInfoAndWrite(gnoData, txinfo.SendAddr, txinfo.CrontractAddr, "create_multisign_tx", url, chainEthId)
}

func buildSigs(data []byte, privateKeys []string) ([]byte, error) {
	var sigs []byte
	for _, privateKeyStr := range privateKeys {
		privateKey, err := crypto.ToECDSA(common.FromHex(privateKeyStr))
		if nil != err {
			return nil, errors.New("paraseKey err:" + err.Error())
		}

		signature, err := crypto.Sign(data, privateKey)
		if err != nil {
			return nil, err
		}
		signature[64] += 27
		sigs = append(sigs, signature[:]...)
	}

	return sigs, nil
}

// SendMultisignTransferTxCmd 创建多签转帐交易
func SendMultisignTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send_multisign_tx",
		Short: "send multisign transfer tx",
		Run:   SendMultisignTransferTx,
	}
	addSendMultisignTransferTxFlags(cmd)
	return cmd
}

func addSendMultisignTransferTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "sign_multisign_tx.txt", "tx file, default: sign_multisign_tx.txt")
	cmd.Flags().StringP("key", "k", "", "private key ")
	_ = cmd.MarkFlagRequired("key")
}

func SendMultisignTransferTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")
	txFilePath, _ := cmd.Flags().GetString("file")
	privatekey, _ := cmd.Flags().GetString("key")
	deployPrivateKey, err := crypto.ToECDSA(common.FromHex(privatekey))
	if err != nil {
		panic(err)
	}

	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("Dial Err:", err)
		return
	}

	var txinfo safeTxData
	err = paraseFile(txFilePath, &txinfo)
	if err != nil {
		fmt.Println("paraseFile Err:", err)
		return
	}

	gnoAbi, err := abi.JSON(strings.NewReader(gnosis.GnosisSafeABI))
	if err != nil {
		fmt.Println("JSON Err:", err)
		return
	}
	zeroAddr := common.HexToAddress(ebTypes.EthNilAddr)
	safeTxGas := big.NewInt(10 * 10000)

	gnoData, err := gnoAbi.Pack("execTransaction", txinfo.To, txinfo.Value, txinfo.TransferData, uint8(0),
		safeTxGas, big.NewInt(0), big.NewInt(0), zeroAddr, zeroAddr, txinfo.Content)
	if err != nil {
		fmt.Println("Pack execTransaction Err:", err)
		return
	}

	info := CreateTxInfo(gnoData, txinfo.SendAddr, txinfo.CrontractAddr, "create_multisign_tx", url, chainEthId)
	if info == nil {
		return
	}

	// sign
	deployTxInfo := DeployInfo{}
	var tx types.Transaction
	err = tx.UnmarshalBinary(common.FromHex(info.RawTx))
	if err != nil {
		panic(err)
	}
	signedTx, txHash, err := eoff.SignEIP155Tx(deployPrivateKey, &tx, chainEthId)
	if err != nil {
		panic(err)
	}
	deployTxInfo.RawTx = signedTx
	deployTxInfo.TxHash = txHash

	// send
	txSend := new(types.Transaction)
	err = txSend.UnmarshalBinary(common.FromHex(deployTxInfo.RawTx))
	if err != nil {
		panic(err)
	}
	err = client.SendTransaction(context.Background(), txSend)
	if err != nil {
		fmt.Println("err:", err)
		panic(err)
	}
	ret := &DeployContractRet{ContractAddr: deployTxInfo.ContractorAddr.String(), TxHash: txSend.Hash().String(), ContractName: deployTxInfo.Name}
	checkTxStatus(client, txSend.Hash().String(), deployTxInfo.Name)

	data, err := json.MarshalIndent(ret, "", "\t")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(data))
}
