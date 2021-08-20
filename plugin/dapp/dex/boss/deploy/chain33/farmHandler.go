package chain33

import (
	"errors"
	"fmt"
	"time"

	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/cakeToken"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/masterChef"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/syrupBar"
	"github.com/spf13/cobra"
)

func DeployFarm(cmd *cobra.Command) error {
	caller, _ := cmd.Flags().GetString("caller")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	txhexCakeToken, err := deployContract(cmd, cakeToken.CakeTokenBin, cakeToken.CakeTokenABI, "", "CakeToken")
	if err != nil {
		return errors.New(err.Error())
	}

	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("Deploy CakeToken timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(txhexCakeToken, rpcLaddr)
				if data == "" {
					fmt.Println("No receipt received yet for Deploy CakeToken tx and continue to wait")
					continue
				} else if data != "2" {
					return errors.New("DeployPancakeFactory failed due to" + ", ty = " + data)
				}
				fmt.Println("Succeed to deploy CakeToken with address =", getContractAddr(caller, txhexCakeToken))
				fmt.Println("")
				goto deploySyrupBar
			}
		}
	}

deploySyrupBar:
	txhexSyrupBar, err := deployContract(cmd, syrupBar.SyrupBarBin, syrupBar.SyrupBarABI, getContractAddr(caller, txhexCakeToken), "SyrupBar")
	if err != nil {
		return errors.New(err.Error())
	}

	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("Deploy SyrupBar timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(txhexSyrupBar, rpcLaddr)
				if data == "" {
					fmt.Println("No receipt received yet for Deploy SyrupBar tx and continue to wait")
					continue
				} else if data != "2" {
					return errors.New("Deploy SyrupBar failed due to" + ", ty = " + data)
				}
				fmt.Println("Succeed to deploy SyrupBar with address =", getContractAddr(caller, txhexSyrupBar))
				fmt.Println("")
				goto deployMasterChef
			}
		}
	}

deployMasterChef:
	// constructor(
	//        CakeToken _cake,
	//        SyrupBar _syrup,
	//        address _devaddr,
	//        uint256 _cakePerBlock,
	//        uint256 _startBlock
	//    )
	// masterChef.DeployMasterChef(auth, ethClient, cakeTokenAddr, SyrupBarAddr, deployerAddr, big.NewInt(5*1e18), big.NewInt(100))
	txparam := getContractAddr(caller, txhexCakeToken) + "," + getContractAddr(caller, txhexSyrupBar) + "," + caller + ", 5000000000000000000, 100"
	txhexMasterChef, err := deployContract(cmd, masterChef.MasterChefBin, masterChef.MasterChefABI, txparam, "MasterChef")
	if err != nil {
		return errors.New(err.Error())
	}

	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("Deploy MasterChef timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(txhexMasterChef, rpcLaddr)
				if data == "" {
					fmt.Println("No receipt received yet for Deploy MasterChef tx and continue to wait")
					continue
				} else if data != "2" {
					return errors.New("Deploy MasterChef failed due to" + ", ty = " + data)
				}
				fmt.Println("Succeed to deploy MasterChef with address =", getContractAddr(caller, txhexMasterChef))
				fmt.Println("")
				return nil
			}
		}
	}
}
