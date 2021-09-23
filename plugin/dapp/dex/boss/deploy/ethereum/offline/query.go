package offline

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

type queryCmd struct {
}

func (q *queryCmd) queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query address", //first step
		Short: " query gasPrice,nonce from the spec address",
		Run:   q.query, //对要部署的factory合约进行签名
	}
	q.addQueryFlags(cmd)
	return cmd
}

func (q *queryCmd) addQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "account address")
	_ = cmd.MarkFlagRequired("address")
}

func (q *queryCmd) query(cmd *cobra.Command, args []string) {
	_ = args
	url, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("address")

	client, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	price, err := client.SuggestGasPrice(ctx)
	if err != nil {
		panic(err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(addr))
	if nil != err {
		fmt.Println("err:", err)
	}
	var info SignCmd
	info.From = addr
	info.GasPrice = price.Uint64()
	info.Nonce = nonce
	info.Timestamp = time.Now().Format(time.RFC3339)
	writeToFile("accountinfo.txt", &info)
	return

}

//deploay Factory contractor

type DeployContract struct {
	ContractAddr string
	TxHash       string
	Nonce        uint64
	RawTx        string
	ContractName string
	Interval     time.Duration
}

func (d *DeployContract) DeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send", //first step
		Short: " send signed raw tx",
		Run:   d.send, //对要部署的factory合约进行签名
	}
	d.addSendFlags(cmd)
	return cmd
}

func (d *DeployContract) addSendFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "*.txt signed tx")
	_ = cmd.MarkFlagRequired("file")
}

func (d *DeployContract) send(cmd *cobra.Command, args []string) {
	_ = args
	filePath, _ := cmd.Flags().GetString("file")
	url, _ := cmd.Flags().GetString("rpc_laddr")
	//解析文件数据
	fmt.Println("file", filePath)
	var rdata = make([]*DeployContract, 0)
	err := paraseFile(filePath, &rdata)
	if err != nil {
		fmt.Println("paraseFile,err", err.Error())
		return
	}
	fmt.Println("parase ready total tx num.", len(rdata))
	for i, deployInfo := range rdata {
		if deployInfo.Interval != 0 {
			time.Sleep(deployInfo.Interval)
		}
		tx := new(types.Transaction)
		err = tx.UnmarshalBinary(common.FromHex(deployInfo.RawTx))
		if err != nil {
			panic(err)
		}
		client, err := ethclient.Dial(url)
		if err != nil {
			panic(err)
		}
		err = client.SendTransaction(context.Background(), tx)
		if err != nil {
			fmt.Println("err:", err)
			panic(err)
		}
		fmt.Println("deplay contract Index Tx", i+1, "TxHash", tx.Hash().String(), "contractName", deployInfo.ContractName, "contractAddr", deployInfo.ContractAddr)
		time.Sleep(time.Second)
	}

	fmt.Println("All tx send ...")

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

func writeToFile(fileName string, content interface{}) {
	jbytes, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(fileName, jbytes, 0666)
	if err != nil {
		fmt.Println("Failed to write to file:", fileName)
	}
	fmt.Println("tx is written to file: ", fileName, "writeContent:", string(jbytes))
}
