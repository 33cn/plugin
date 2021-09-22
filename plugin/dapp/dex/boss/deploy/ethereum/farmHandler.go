package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/cakeToken"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/masterChef"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/syrupBar"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func GetCakeBalance(owner string, pid int64) (string, error) {
	masterChefInt, err := masterChef.NewMasterChef(common.HexToAddress("0xD88654a6aAc42a7192d697a8250a93246De882C6"), ethClient)
	if nil != err {
		return "", err
	}
	ownerAddr := common.HexToAddress(owner)
	opts := &bind.CallOpts{
		From:    ownerAddr,
		Context: context.Background(),
	}
	amount, err := masterChefInt.PendingCake(opts, big.NewInt(pid), ownerAddr)
	if nil != err {
		return "", err
	}
	return amount.String(), nil
}

func DeployFarm(key string) error {
	_ = recoverEthTestNetPrivateKey(key)
	//1st step to deploy factory
	auth, err := PrepareAuth(privateKey, deployerAddr)
	if nil != err {
		return err
	}

	cakeTokenAddr, deploycakeTokenTx, _, err := cakeToken.DeployCakeToken(auth, ethClient)
	if nil != err {
		panic(fmt.Sprintf("Failed to DeployCakeToken with err:%s", err.Error()))
	}

	{
		fmt.Println("\nDeployCakeToken tx hash:", deploycakeTokenTx.Hash().String())
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeployCakeToken timeout")
			case <-oneSecondtimeout.C:
				_, err := ethClient.TransactionReceipt(context.Background(), deploycakeTokenTx.Hash())
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received yet for DeployCakeToken tx and continue to wait")
					continue
				} else if err != nil {
					panic("DeployCakeToken failed due to" + err.Error())
				}
				fmt.Println("\n Succeed to deploy DeployCakeToken with address =", cakeTokenAddr.String())
				goto deploySyrupBar
			}
		}
	}

deploySyrupBar:
	auth, err = PrepareAuth(privateKey, deployerAddr)
	if nil != err {
		return err
	}
	SyrupBarAddr, deploySyrupBarTx, _, err := syrupBar.DeploySyrupBar(auth, ethClient, cakeTokenAddr)
	if err != nil {
		panic(fmt.Sprintf("Failed to DeploySyrupBar with err:%s", err.Error()))
	}

	{
		fmt.Println("\nDeploySyrupBar tx hash:", deploySyrupBarTx.Hash().String())
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeploySyrupBar timeout")
			case <-oneSecondtimeout.C:
				_, err := ethClient.TransactionReceipt(context.Background(), deploySyrupBarTx.Hash())
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received yet for DeploySyrupBar tx and continue to wait")
					continue
				} else if err != nil {
					panic("DeploySyrupBar failed due to" + err.Error())
				}
				fmt.Println("\n Succeed to deploy DeploySyrupBar with address =", SyrupBarAddr.String())
				goto deployMasterchef
			}
		}
	}

deployMasterchef:
	auth, err = PrepareAuth(privateKey, deployerAddr)
	if nil != err {
		return err
	}
	//auth *bind.TransactOpts, backend bind.ContractBackend, _cake common.Address, _syrup common.Address, _devaddr common.Address, _cakePerBlock *big.Int, _startBlock *big.Int
	MasterChefAddr, deployMasterChefTx, _, err := masterChef.DeployMasterChef(auth, ethClient, cakeTokenAddr, SyrupBarAddr, deployerAddr, big.NewInt(5*1e18), big.NewInt(100))
	if err != nil {
		panic(fmt.Sprintf("Failed to DeployMasterChef with err:%s", err.Error()))
	}

	{
		fmt.Println("\nDeployMasterChef tx hash:", deployMasterChefTx.Hash().String())
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeployMasterChef timeout")
			case <-oneSecondtimeout.C:
				_, err := ethClient.TransactionReceipt(context.Background(), deployMasterChefTx.Hash())
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received yet for DeployMasterChef tx and continue to wait")
					continue
				} else if err != nil {
					panic("DeployMasterChef failed due to" + err.Error())
				}
				fmt.Println("\n Succeed to deploy DeployMasterChef with address =", MasterChefAddr.String())
				return nil
			}
		}
	}
}

func AddPool2FarmHandle(masterChefAddrStr, key string, allocPoint int64, lpToken string, withUpdate bool, gasLimit uint64) (err error) {
	masterChefAddr := common.HexToAddress(masterChefAddrStr)
	masterChefInt, err := masterChef.NewMasterChef(masterChefAddr, ethClient)
	if nil != err {
		return err
	}

	_ = recoverEthTestNetPrivateKey(key)
	//1st step to deploy factory
	auth, err := PrepareAuth(privateKey, deployerAddr)
	if nil != err {
		return err
	}
	auth.GasLimit = gasLimit
	AddPool2FarmTx, err := masterChefInt.Add(auth, big.NewInt(int64(allocPoint)), common.HexToAddress(lpToken), withUpdate)
	if err != nil {
		panic(fmt.Sprintf("Failed to AddPool2FarmTx with err:%s", err.Error()))
	}

	fmt.Println("\nAddPool2FarmTx tx hash:", AddPool2FarmTx.Hash().String())
	timeout := time.NewTimer(300 * time.Second)
	oneSecondtimeout := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeout.C:
			panic("AddPool2FarmTx timeout")
		case <-oneSecondtimeout.C:
			_, err := ethClient.TransactionReceipt(context.Background(), AddPool2FarmTx.Hash())
			if err == ethereum.NotFound {
				fmt.Println("\n No receipt received yet for AddPool2FarmTx tx and continue to wait")
				continue
			} else if err != nil {
				panic("AddPool2FarmTx failed due to" + err.Error())
			}
			return nil
		}
	}
}

func UpdateAllocPointHandle(masterChefAddrStr, key string, pid, allocPoint int64, withUpdate bool) (err error) {
	masterChefAddr := common.HexToAddress(masterChefAddrStr)
	masterChefInt, err := masterChef.NewMasterChef(masterChefAddr, ethClient)
	if nil != err {
		return err
	}

	_ = recoverEthTestNetPrivateKey(key)
	auth, err := PrepareAuth(privateKey, deployerAddr)
	if nil != err {
		return err
	}

	//_pid *big.Int, _allocPoint *big.Int, _withUpdate bool
	SetPool2FarmTx, err := masterChefInt.Set(auth, big.NewInt(pid), big.NewInt(allocPoint), withUpdate)
	if err != nil {
		panic(fmt.Sprintf("Failed to SetPool2FarmTx with err:%s", err.Error()))
	}

	{
		fmt.Println("\nSetPool2FarmTx tx hash:", SetPool2FarmTx.Hash().String())
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("SetPool2FarmTx timeout")
			case <-oneSecondtimeout.C:
				_, err := ethClient.TransactionReceipt(context.Background(), SetPool2FarmTx.Hash())
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received yet for SetPool2FarmTx tx and continue to wait")
					continue
				} else if err != nil {
					panic("SetPool2FarmTx failed due to" + err.Error())
				}
				return nil
			}
		}
	}
}

func TransferOwnerShipHandle(newOwner, contract, key string) (err error) {
	contractAddr := common.HexToAddress(contract)
	contractInt, err := syrupBar.NewSyrupBar(contractAddr, ethClient)
	if nil != err {
		return err
	}

	_ = recoverEthTestNetPrivateKey(key)
	auth, err := PrepareAuth(privateKey, deployerAddr)
	if nil != err {
		return err
	}

	//_pid *big.Int, _allocPoint *big.Int, _withUpdate bool

	newOwnerAddr := common.HexToAddress(newOwner)
	TransferOwnershipTx, err := contractInt.TransferOwnership(auth, newOwnerAddr)
	if err != nil {
		panic(fmt.Sprintf("Failed to TransferOwnership with err:%s", err.Error()))
	}

	{
		fmt.Println("\nTransferOwnership tx hash:", TransferOwnershipTx.Hash().String())
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("TransferOwnership timeout")
			case <-oneSecondtimeout.C:
				_, err := ethClient.TransactionReceipt(context.Background(), TransferOwnershipTx.Hash())
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received yet for TransferOwnership tx and continue to wait")
					continue
				} else if err != nil {
					panic("TransferOwnership failed due to" + err.Error())
				}
				return nil
			}
		}
	}
}

func updateCakePerBlockHandle(cakePerBlock *big.Int, startBlock int64, masterchef, key string) (err error) {
	_ = recoverEthTestNetPrivateKey(key)
	masterChefInt, err := masterChef.NewMasterChef(common.HexToAddress(masterchef), ethClient)
	if nil != err {
		return err
	}
	auth, err := PrepareAuth(privateKey, deployerAddr)
	if nil != err {
		return err
	}
	updateCakePerBlockTx, err := masterChefInt.UpdateCakePerBlock(auth, cakePerBlock, big.NewInt(startBlock))
	if nil != err {
		panic(fmt.Sprintf("Failed to UpdateCakePerBlock with err:%s", err.Error()))
	}

	{
		fmt.Println("\nUpdateCakePerBlock tx hash:", updateCakePerBlockTx.Hash().String())
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("UpdateCakePerBlock timeout")
			case <-oneSecondtimeout.C:
				_, err := ethClient.TransactionReceipt(context.Background(), updateCakePerBlockTx.Hash())
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received yet for UpdateCakePerBlock tx and continue to wait")
					continue
				} else if err != nil {
					panic("UpdateCakePerBlock failed due to" + err.Error())
				}
				fmt.Println("\n Succeed to do the UpdateCakePerBlock operation")
				return nil
			}
		}
	}

}
