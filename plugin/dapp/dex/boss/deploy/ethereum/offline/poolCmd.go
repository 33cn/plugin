package offline

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/masterChef"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/syrupBar"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

type AddPool struct {
	allocPoint int64
	lpToken    string
	withUpdate bool
}

func (a *AddPool) AddPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add pool",
		Short: "add pool to farm ",
		Run:   a.AddPool2Farm,
	}

	a.addAddPoolCmdFlags(cmd)
	return cmd
}
func (a *AddPool) addAddPoolCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("masterchef", "m", "", "master Chef Addr ")
	_ = cmd.MarkFlagRequired("masterchef")

	cmd.Flags().StringP("lptoken", "l", "", "lp Addr ")
	_ = cmd.MarkFlagRequired("lptoken")

	cmd.Flags().Int64P("alloc", "a", 0, "allocation point ")
	_ = cmd.MarkFlagRequired("alloc")

	cmd.Flags().BoolP("update", "u", true, "with update")
	_ = cmd.MarkFlagRequired("update")

	cmd.Flags().StringP("priv", "p", "", "private key")
	_ = cmd.MarkFlagRequired("priv")

	cmd.Flags().StringP("file", "f", "accountinfo.txt", "account info")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().Uint64P("nonce", "n", 0, "transaction count")
	cmd.MarkFlagRequired("nonce")
	cmd.Flags().Uint64P("gasprice", "g", 1000000000, "gas price")
	cmd.MarkFlagRequired("gasprice")

}

func (a *AddPool) AddPool2Farm(cmd *cobra.Command, args []string) {
	masterChefAddrStr, _ := cmd.Flags().GetString("masterchef")
	allocPoint, _ := cmd.Flags().GetInt64("alloc")
	lpToken, _ := cmd.Flags().GetString("lptoken")
	update, _ := cmd.Flags().GetBool("update")
	key, _ := cmd.Flags().GetString("priv")
	nonce, _ := cmd.Flags().GetUint64("nonce")
	price, _ := cmd.Flags().GetUint64("gasprice")
	priv, from, err := recoverBinancePrivateKey(key)
	if err != nil {
		panic(err)
	}

	a.allocPoint = allocPoint
	a.lpToken = lpToken
	a.withUpdate = update
	var signData = make([]*DeployContract, 0)
	var signInfo SignCmd
	signInfo.Nonce = nonce
	signInfo.GasPrice = price
	signInfo.From = from.String()

	//--------------------
	//sign addpool
	//--------------------
	signedtx, hash, err := a.reWriteAddPool2Farm(signInfo.Nonce, masterChefAddrStr, big.NewInt(int64(signInfo.GasPrice)), priv)
	if nil != err {
		fmt.Println("Failed to AddPool2Farm due to:", err.Error())
		return
	}
	var addPoolData = new(DeployContract)
	addPoolData.Nonce = signInfo.Nonce
	addPoolData.RawTx = signedtx
	addPoolData.TxHash = hash
	addPoolData.ContractName = "addpool"
	signData = append(signData, addPoolData)
	writeToFile("addPool.txt", signData)
	fmt.Println("Succeed to sign AddPool")
}

func (a *AddPool) reWriteAddPool2Farm(nonce uint64, masterChefAddrStr string, gasPrice *big.Int, key *ecdsa.PrivateKey) (signedTx, hash string, err error) {
	masterChefAddr := common.HexToAddress(masterChefAddrStr)
	parsed, err := abi.JSON(strings.NewReader(masterChef.MasterChefABI))
	input, err := parsed.Pack("add", big.NewInt(a.allocPoint), common.HexToAddress(a.lpToken), a.withUpdate)
	if err != nil {
		panic(err)
	}
	ntx := types.NewTransaction(nonce, masterChefAddr, new(big.Int), gasLimit, gasPrice, input)
	return SignTx(key, ntx)
}

//------------
//update

type updateAllocPoint struct {
	pid, allocPoint int64
	withUpdate      bool
}

func (u *updateAllocPoint) UpdateAllocPointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update alloc point",
		Short: "Update the given pool's CAKE allocation point",
		Run:   u.UpdateAllocPoint,
	}

	u.updateAllocPointCmdFlags(cmd)
	return cmd
}

func (u *updateAllocPoint) UpdateAllocPoint(cmd *cobra.Command, args []string) {
	masterChefAddrStr, _ := cmd.Flags().GetString("masterchef")
	pid, _ := cmd.Flags().GetInt64("pid")
	allocPoint, _ := cmd.Flags().GetInt64("alloc")
	update, _ := cmd.Flags().GetBool("update")
	key, _ := cmd.Flags().GetString("priv")
	nonce, _ := cmd.Flags().GetUint64("nonce")
	price, _ := cmd.Flags().GetUint64("gasprice")
	priv, from, err := recoverBinancePrivateKey(key)
	if err != nil {
		panic(err)
	}
	u.pid = pid
	u.allocPoint = allocPoint
	u.withUpdate = update
	var signInfo SignCmd
	var signData = make([]*DeployContract, 0)
	signInfo.Nonce = nonce
	signInfo.GasPrice = price
	signInfo.From = from.String()
	signedtx, hash, err := u.rewriteUpdateAllocPoint(masterChefAddrStr, signInfo.Nonce, big.NewInt(int64(signInfo.GasPrice)), priv)
	if err != nil {
		panic(err)
	}

	var updateAllocData = new(DeployContract)
	updateAllocData.Nonce = signInfo.Nonce
	updateAllocData.RawTx = signedtx
	updateAllocData.TxHash = hash
	updateAllocData.ContractName = "updateAllocPoint"
	signData = append(signData, updateAllocData)
	writeToFile("updateAllocPoint.txt", signData)
	fmt.Println("Succeed to sign updateAllocPoint")

}
func (u *updateAllocPoint) updateAllocPointCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("masterchef", "m", "", "master Chef Addr ")
	_ = cmd.MarkFlagRequired("masterchef")

	cmd.Flags().Int64P("pid", "d", 0, "id of pool")
	_ = cmd.MarkFlagRequired("pid")

	cmd.Flags().Int64P("alloc", "a", 0, "allocation point ")
	_ = cmd.MarkFlagRequired("alloc")

	cmd.Flags().BoolP("update", "u", true, "with update")
	_ = cmd.MarkFlagRequired("update")
	cmd.Flags().StringP("priv", "p", "", "private key")
	_ = cmd.MarkFlagRequired("priv")

	cmd.Flags().Uint64P("nonce", "n", 0, "transaction count")
	cmd.MarkFlagRequired("nonce")
	cmd.Flags().Uint64P("gasprice", "g", 1000000000, "gas price")
	cmd.MarkFlagRequired("gasprice")

}

func (u *updateAllocPoint) rewriteUpdateAllocPoint(masterChefAddrStr string, nonce uint64, gasPrice *big.Int, key *ecdsa.PrivateKey) (signedTx, hash string, err error) {
	masterChefAddr := common.HexToAddress(masterChefAddrStr)
	parsed, err := abi.JSON(strings.NewReader(masterChef.MasterChefABI))
	input, err := parsed.Pack("set", big.NewInt(u.pid), big.NewInt(u.allocPoint), u.withUpdate)
	if err != nil {
		panic(err)
	}
	ntx := types.NewTransaction(nonce, masterChefAddr, new(big.Int), gasLimit, gasPrice, input)
	return SignTx(key, ntx)

}

type transferOwnerShip struct {
}

func (t *transferOwnerShip) TransferOwnerShipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer OwnerShip",
		Short: "transfer OwnerShip",
		Run:   t.TransferOwnerShip,
	}

	t.TransferOwnerShipFlags(cmd)
	return cmd
}

func (t *transferOwnerShip) TransferOwnerShipFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("new", "n", "", "new owner")
	_ = cmd.MarkFlagRequired("new")

	cmd.Flags().StringP("contract", "c", "", "contract address")
	_ = cmd.MarkFlagRequired("contract")

	cmd.Flags().StringP("file", "f", "accountinfo.txt", "account info")
	_ = cmd.MarkFlagRequired("file")

	cmd.Flags().StringP("priv", "p", "", "private key")
	_ = cmd.MarkFlagRequired("priv")

}

func (t *transferOwnerShip) TransferOwnerShip(cmd *cobra.Command, args []string) {
	newOwner, _ := cmd.Flags().GetString("new")
	contract, _ := cmd.Flags().GetString("contract")
	key, _ := cmd.Flags().GetString("priv")
	filePath, _ := cmd.Flags().GetString("file")
	priv, from, err := recoverBinancePrivateKey(key)
	if err != nil {
		panic(err)
	}

	var signInfo SignCmd
	paraseFile(filePath, &signInfo)
	checkFile(signInfo.From, from.String(), signInfo.Timestamp)

	signedtx, hash, err := TransferOwnerShipHandle(signInfo.Nonce, big.NewInt(int64(signInfo.GasPrice)), newOwner, contract, priv)
	if nil != err {
		fmt.Println("Failed to TransferOwnerShip due to:", err.Error())
		return
	}

	var transferOwner = new(DeployContract)
	var signData = make([]*DeployContract, 0)
	transferOwner.Nonce = signInfo.Nonce
	transferOwner.RawTx = signedtx
	transferOwner.TxHash = hash
	transferOwner.ContractName = "transferOwnership"
	signData = append(signData, transferOwner)
	writeToFile("transferOwner.txt", signData)
	fmt.Println("Succeed to sign TransferOwnerShip")
}

func TransferOwnerShipHandle(nonce uint64, gasPrice *big.Int, newOwner, contract string, key *ecdsa.PrivateKey) (signedtx, hash string, err error) {
	contractAddr := common.HexToAddress(contract)
	newOwnerAddr := common.HexToAddress(newOwner)
	parsed, err := abi.JSON(strings.NewReader(syrupBar.SyrupBarABI))
	input, err := parsed.Pack("transferOwnership", newOwnerAddr)
	if err != nil {
		return
	}
	ntx := types.NewTransaction(nonce, contractAddr, big.NewInt(0), gasLimit, gasPrice, input)
	return SignTx(key, ntx)

}

func checkFile(from, keyaddr, timestamp string) {
	//check is timeout
	tim, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		panic(err)
	}
	if time.Now().After(tim.Add(time.Hour)) {
		panic("after 60 minutes timeout,the accountinfo.txt invalid,please reQuery")
	}
	if !strings.EqualFold(from, keyaddr) {
		panic("deployed address mismatch!!!")
	}

}
