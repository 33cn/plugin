package offline

import (
	"fmt"
	"math/big"
	"strings"

	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4chain33/generated"
	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	chain33Relayer "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/chain33"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	relayerutils "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common/math"
	btcecsecp256k1 "github.com/btcsuite/btcd/btcec"
	ethSecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/spf13/cobra"
)

/*
./boss4x chain33 offline create_multisign_transfer -a 10 -r 168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9 -m 1NFDfEwne4kjuxAZrtYEh4kfSrnGSE7ap
./boss4x chain33 offline multisign_transfer -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262a -s 0xcd284cd17456b73619fa609bb9e3105e8eff5d059c5e0b6eb1effbebd4d64144,0xe892212221b3b58211b90194365f4662764b6d5474ef2961ef77c909e31eeed3,0x9d19a2e9a440187010634f4f08ce36e2bc7b521581436a99f05568be94dc66ea,0x45d4ce009e25e6d5e00d8d3a50565944b2e3604aa473680a656b242d9acbff35 --chainID 33
./boss4x chain33 offline send -f multisign_transfer.txt
*/

type transferTxData struct {
	Receiver      string
	Token         string
	MultisignAddr string
	Data          string
	Amount        float64
	name          string
}

func CreateMultisignTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_multisign_transfer",
		Short: "create multisign transfer tx",
		Run:   CreateMultisignTransfer,
	}
	addCreateMultisignTransferFlags(cmd)
	return cmd
}

func addCreateMultisignTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("receiver", "r", "", "receive address")
	_ = cmd.MarkFlagRequired("receiver")

	cmd.Flags().Float64P("amount", "a", 0, "amount to transfer")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("token", "t", "", "erc20 address,not need to set for BTY(optional)")

	//
	cmd.Flags().StringP("address", "m", "", "multisign address")
	_ = cmd.MarkFlagRequired("address")
}

func CreateMultisignTransfer(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	receiver, _ := cmd.Flags().GetString("receiver")
	token, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	multisign, _ := cmd.Flags().GetString("address")

	//对于平台币转账，这个data只是个占位符，没有作用
	dataStr := "0x"
	safeTxGas := int64(10 * 10000)
	baseGas := 0
	gasPrice := 0
	valueStr := relayerutils.ToWei(amount, 8).String()
	//如果是bty转账,则直接将to地址设置为receiver,而如果是ERC20转账，则需要将其设置为token地址
	to := receiver
	//如果是erc20转账，则需要构建data数据
	if "" != token {
		parameter := fmt.Sprintf("transfer(%s, %s)", receiver, relayerutils.ToWei(amount, 8).String())
		_, data, err := evmAbi.Pack(parameter, erc20.ERC20ABI, false)
		if err != nil {
			fmt.Println("Failed to do abi.Pack due to:", err.Error())
			return
		}
		//对于其他erc20资产，直接将其设置为0
		valueStr = "0"
		to = token
		dataStr = chain33Common.ToHex(data)
	}

	//获取nonce
	nonce := getMulSignNonce(multisign, rpcLaddr)
	parameter2getHash := fmt.Sprintf("getTransactionHash(%s, %s, %s, 0, %d, %d, %d, %s, %s, %d)", to, valueStr, dataStr,
		safeTxGas, baseGas, gasPrice, ebrelayerTypes.NilAddrChain33, ebrelayerTypes.NilAddrChain33, nonce)

	result := chain33Relayer.Query(multisign, parameter2getHash, multisign, rpcLaddr, generated.GnosisSafeABI)
	if nil == result {
		fmt.Println("Failed to getTransactionHash :", ebrelayerTypes.ErrGetTransactionHash)
		return
	}
	contentHashArray := result.([32]byte)
	contentHash := contentHashArray[:]

	var txinfo transferTxData
	txinfo.Receiver = receiver
	txinfo.MultisignAddr = multisign
	txinfo.Amount = amount
	txinfo.Data = chain33Common.ToHex(contentHash)
	txinfo.Token = token
	txinfo.name = "create_multisign_transfer"
	writeToFile(txinfo.name+".txt", txinfo)
}

func MultisignTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign_transfer",
		Short: "create multisign transfer tx and sign",
		Run:   MultisignTransfer,
	}
	addMultisignTransferFlags(cmd)
	return cmd
}

func addMultisignTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "t", "create_multisign_transfer.txt", "tx file, default: create_multisign_transfer.txt")
	cmd.Flags().StringP("keys", "s", "", "owners' private key, separated by ','")
	_ = cmd.MarkFlagRequired("keys")
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func MultisignTransfer(cmd *cobra.Command, _ []string) {
	keysStr, _ := cmd.Flags().GetString("keys")
	keys := strings.Split(keysStr, ",")
	txFilePath, _ := cmd.Flags().GetString("file")
	var txinfo transferTxData
	err := paraseFile(txFilePath, &txinfo)
	if err != nil {
		fmt.Println("paraseFile Err:", err)
		return
	}

	//对于平台币转账，这个data只是个占位符，没有作用
	dataStr := "0x"
	contentHash, err := chain33Common.FromHex(txinfo.Data)
	safeTxGas := int64(10 * 10000)
	baseGas := 0
	gasPrice := 0
	valueStr := relayerutils.ToWei(txinfo.Amount, 8).String()
	//如果是bty转账,则直接将to地址设置为receiver,而如果是ERC20转账，则需要将其设置为token地址
	to := txinfo.Receiver
	//如果是erc20转账，则需要构建data数据
	if "" != txinfo.Token {
		parameter := fmt.Sprintf("transfer(%s, %s)", txinfo.Receiver, relayerutils.ToWei(txinfo.Amount, 8).String())
		_, data, err := evmAbi.Pack(parameter, erc20.ERC20ABI, false)
		if err != nil {
			fmt.Println("evmAbi.Pack(parameter, erc20.ERC20ABI, false)", "Failed", err.Error())
			return
		}
		//对于其他erc20资产，直接将其设置为0
		valueStr = "0"
		to = txinfo.Token
		dataStr = chain33Common.ToHex(data)
	}

	var sigs []byte
	for _, privateKey := range keys {
		var driver secp256k1.Driver
		privateKeySli, err := chain33Common.FromHex(privateKey)
		if nil != err {
			fmt.Println("evmAbi.Pack(parameter, erc20.ERC20ABI, false)", "Failed", err.Error())
			return
		}
		ownerPrivateKey, err := driver.PrivKeyFromBytes(privateKeySli)
		if nil != err {
			fmt.Println("evmAbi.Pack(parameter, erc20.ERC20ABI, false)", "Failed", err.Error())
			return
		}
		temp, _ := btcecsecp256k1.PrivKeyFromBytes(btcecsecp256k1.S256(), ownerPrivateKey.Bytes())
		privateKey4chain33Ecdsa := temp.ToECDSA()

		sig, err := ethSecp256k1.Sign(contentHash, math.PaddedBigBytes(privateKey4chain33Ecdsa.D, 32))
		if nil != err {
			fmt.Println("evmAbi.Pack(parameter, erc20.ERC20ABI, false)", "Failed", err.Error())
			return
		}

		sig[64] += 27
		sigs = append(sigs, sig...)
	}

	//构造execTransaction参数
	parameter2Exec := fmt.Sprintf("execTransaction(%s, %s, %s, 0, %d, %d, %d, %s, %s, %s)", to, valueStr, dataStr,
		safeTxGas, baseGas, gasPrice, ebrelayerTypes.NilAddrChain33, ebrelayerTypes.NilAddrChain33, chain33Common.ToHex(sigs))
	_, packData, err := evmAbi.Pack(parameter2Exec, generated.GnosisSafeABI, false)
	if nil != err {
		fmt.Println("execTransaction evmAbi.Pack", "Failed", err.Error())
		return
	}

	callContractAndSignWrite(cmd, packData, txinfo.MultisignAddr, "multisign_transfer")
}

func getMulSignNonce(mulsign, rpcLaddr string) int64 {
	parameter := fmt.Sprintf("nonce()")

	result := chain33Relayer.Query(mulsign, parameter, mulsign, rpcLaddr, generated.GnosisSafeABI)
	if nil == result {
		return 0
	}
	nonce := result.(*big.Int)
	return nonce.Int64()
}
