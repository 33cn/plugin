package offline

import (
	"fmt"

	eoff "github.com/33cn/plugin/plugin/dapp/dex/boss/deploy/ethereum/offline"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

func SignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign", //first step
		Short: "sign tx",
		Run:   signTx,
	}
	addSignFlag(cmd)
	return cmd
}

func addSignFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "private key ")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("file", "f", "deploytxs.txt", "tx file, default:deploytxs.txt")
}

func signTx(cmd *cobra.Command, _ []string) {
	privatekey, _ := cmd.Flags().GetString("key")
	txFilePath, _ := cmd.Flags().GetString("file")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")
	deployPrivateKey, err := crypto.ToECDSA(common.FromHex(privatekey))
	if err != nil {
		panic(err)
	}

	var deployTxInfos = make([]DeployInfo, 0)
	err = paraseFile(txFilePath, &deployTxInfos)
	if err != nil {
		fmt.Println("paraseFile,err", err.Error())
		return
	}
	fmt.Println("deployTxInfos size:", len(deployTxInfos))
	for i, info := range deployTxInfos {
		var tx types.Transaction
		err = tx.UnmarshalBinary(common.FromHex(info.RawTx))
		if err != nil {
			panic(err)
		}
		signedTx, txHash, err := eoff.SignEIP155Tx(deployPrivateKey, &tx, chainEthId)
		if err != nil {
			panic(err)
		}
		deployTxInfos[i].RawTx = signedTx
		deployTxInfos[i].TxHash = txHash
	}

	//finsh write to file
	writeToFile("deploysigntxs.txt", deployTxInfos)
}
