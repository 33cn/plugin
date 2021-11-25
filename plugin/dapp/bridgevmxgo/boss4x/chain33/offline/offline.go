package offline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/33cn/plugin/plugin/dapp/dex/utils"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var crossXfileName = "deployBridgevmxgo2Chain33.txt"

func Boss4xOfflineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "offline",
		Short: "create and sign offline tx to deploy and set cross contracts to chain33",
	}
	cmd.AddCommand(
		CreateBridgevmxgoCmd(),
		SendSignTxs2Chain33Cmd(),
		CreateERC20Cmd(),
		ApproveErc20Cmd(),
		AddToken2LockListCmd(),
		CreateNewBridgeTokenCmd(),
		SetupCmd(),
		ConfigOfflineSaveAccountCmd(),
		ConfigLockedTokenOfflineSaveCmd(),
		CreateMultisignTransferCmd(),
		MultisignTransferCmd(),
	)
	return cmd
}

func SendSignTxs2Chain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "send all the txs to chain33 in serial",
		Run:   sendSignTxs2Chain33,
	}
	addSendSignTxs2Chain33Flags(cmd)
	return cmd
}

func addSendSignTxs2Chain33Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "signed tx file")
	_ = cmd.MarkFlagRequired("file")
}

func sendSignTxs2Chain33(cmd *cobra.Command, _ []string) {
	filePath, _ := cmd.Flags().GetString("file")
	url, _ := cmd.Flags().GetString("rpc_laddr")
	utils.SendSignTxs2Chain33(filePath, url)
}

func getTxInfo(cmd *cobra.Command) *utils.TxCreateInfo {
	privateKey, _ := cmd.Flags().GetString("key")
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	info := &utils.TxCreateInfo{
		PrivateKey: privateKey,
		Expire:     expire,
		Note:       note,
		Fee:        feeInt64,
		ParaName:   paraName,
		ChainID:    chainID,
	}

	return info
}

func writeToFile(fileName string, content interface{}) {
	jbytes, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(fileName, jbytes, 0666)
	if err != nil {
		fmt.Println("Failed to write to file:", fileName)
	}
	fmt.Println("tx is written to file: ", fileName)
}

func paraseFile(file string, result interface{}) error {
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

func callContractAndSignWrite(cmd *cobra.Command, para []byte, contractAddr, name string) {
	action := &evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: name, Para: para, ContractAddr: contractAddr}
	content, txHash, err := utils.CallContractAndSign(getTxInfo(cmd), action, contractAddr)
	if nil != err {
		fmt.Println("CallContractAndSign", "Failed", err.Error())
		return
	}

	Tx := &utils.Chain33OfflineTx{
		ContractAddr:  contractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: name,
	}

	_, err = json.MarshalIndent(Tx, "", "    ")
	if err != nil {
		fmt.Println("MarshalIndent error", err.Error())
		return
	}

	var txs []*utils.Chain33OfflineTx
	txs = append(txs, Tx)

	fileName := fmt.Sprintf(Tx.OperationName + ".txt")
	fmt.Printf("Write all the txs to file:   %s \n", fileName)
	utils.WriteToFileInJson(fileName, txs)
}
