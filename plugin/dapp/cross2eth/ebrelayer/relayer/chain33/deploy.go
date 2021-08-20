package chain33

import (
	"errors"
	"fmt"
	"time"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4chain33/generated"
	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func deployAndInit2Chain33(rpcLaddr, paraChainName string, para4deploy *DeployPara4Chain33) (*X2EthDeployResult, error) {
	deployer := para4deploy.Deployer.String()

	var err error
	constructorPara := ""
	paraLen := len(para4deploy.InitValidators)

	var valsetAddr string
	var ethereumBridgeAddr string
	var oracleAddr string
	var bridgeBankAddr string

	deployBridgeRegistry := &DeployResult{}
	deployBridgeBank := &DeployResult{}
	deployEthereumBridge := &DeployResult{}
	deployValset := &DeployResult{}
	deployOracle := &DeployResult{}

	deployInfo := &X2EthDeployResult{
		BridgeRegistry: deployBridgeRegistry,
		BridgeBank:     deployBridgeBank,
		EthereumBridge: deployEthereumBridge,
		Valset:         deployValset,
		Oracle:         deployOracle,
	}

	//constructor(
	//	address _operator,
	//	address[] memory _initValidators,
	//	uint256[] memory _initPowers
	//)
	if 1 == paraLen {
		constructorPara = fmt.Sprintf("constructor(%s, %s, %d)", para4deploy.Operator.String(),
			para4deploy.InitValidators[0].String(),
			para4deploy.InitPowers[0].Int64())
	} else if 4 == paraLen {
		constructorPara = fmt.Sprintf("constructor(%s, [%s, %s, %s, %s], [%d, %d, %d, %d])", para4deploy.Operator.String(),
			para4deploy.InitValidators[0].String(), para4deploy.InitValidators[1].String(), para4deploy.InitValidators[2].String(), para4deploy.InitValidators[3].String(),
			para4deploy.InitPowers[0].Int64(), para4deploy.InitPowers[1].Int64(), para4deploy.InitPowers[2].Int64(), para4deploy.InitPowers[3].Int64())
	} else {
		panic(fmt.Sprintf("Not support valset with parameter count=%d", paraLen))
	}

	deployValsetHash, err := deploySingleContract(ethcommon.FromHex(generated.ValsetBin), generated.ValsetABI, constructorPara, "valset", paraChainName, para4deploy.Deployer.String(), rpcLaddr)
	if nil != err {
		chain33txLog.Error("DeployAndInit", "failed to DeployValset due to:", err.Error())
		return nil, err
	}
	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeployValset timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(deployValsetHash, rpcLaddr)
				if data == "" {
					chain33txLog.Info("deployAndInit2Chain33", "No receipt received yet for Deploy valset tx and continue to wait", "continue")
					continue
				} else if data != "2" {
					return nil, errors.New("Deploy valset failed due to" + ", ty = " + data)
				}

				deployValset.Address = getContractAddr(deployer, deployValsetHash)
				deployValset.TxHash = deployValsetHash
				valsetAddr = deployValset.Address.String()
				chain33txLog.Info("deployAndInit2Chain33", "Succeed to deploy valset with address =", valsetAddr)
				goto deployEthereumBridge
			}
		}
	}

deployEthereumBridge:
	//x2EthContracts.Chain33Bridge, deployInfo.Chain33Bridge, err = DeployChain33Bridge(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Valset.Address)
	//constructor(
	//	address _operator,
	//	address _valset
	//)
	constructorPara = fmt.Sprintf("constructor(%s, %s)", para4deploy.Operator.String(), valsetAddr)
	deployEthereumBridgeHash, err := deploySingleContract(ethcommon.FromHex(generated.EthereumBridgeBin), generated.EthereumBridgeABI, constructorPara, "EthereumBridge", paraChainName, para4deploy.Deployer.String(), rpcLaddr)
	if nil != err {
		chain33txLog.Error("DeployAndInit", "failed to deployEthereumBridge due to:", err.Error())
		return nil, err
	}
	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("deployEthereumBridge timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(deployEthereumBridgeHash, rpcLaddr)
				if data == "" {
					chain33txLog.Info("deployAndInit2Chain33", "No receipt received yet for Deploy EthereumBridge tx and continue to wait", "continue")
					continue
				} else if data != "2" {
					return nil, errors.New("Deploy EthereumBridge failed due to" + ", ty = " + data)
				}
				deployEthereumBridge.Address = getContractAddr(deployer, deployEthereumBridgeHash)
				deployValset.TxHash = deployEthereumBridgeHash
				ethereumBridgeAddr = deployEthereumBridge.Address.String()
				chain33txLog.Info("deployAndInit2Chain33", "Succeed to deploy EthereumBridge with address =", ethereumBridgeAddr)
				goto deployOracle
			}
		}
	}

deployOracle:
	//constructor(
	//	address _operator,
	//	address _valset,
	//	address _ethereumBridge
	//)
	constructorPara = fmt.Sprintf("constructor(%s, %s, %s)", para4deploy.Operator.String(), valsetAddr, ethereumBridgeAddr)
	//x2EthContracts.Oracle, deployInfo.Oracle, err = DeployOracle(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Valset.Address, deployInfo.Chain33Bridge.Address)
	deployOracleHash, err := deploySingleContract(ethcommon.FromHex(generated.OracleBin), generated.OracleABI, constructorPara, "Oracle", paraChainName, para4deploy.Deployer.String(), rpcLaddr)
	if nil != err {
		chain33txLog.Error("DeployAndInit", "failed to DeployOracle due to:", err.Error())
		return nil, err
	}
	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("deployOracle timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(deployOracleHash, rpcLaddr)
				if data == "" {
					chain33txLog.Info("deployAndInit2Chain33", "No receipt received yet for Deploy Oracle tx and continue to wait", "continue")
					continue
				} else if data != "2" {
					return nil, errors.New("Deploy Oracle failed due to" + ", ty = " + data)
				}
				deployOracle.Address = getContractAddr(deployer, deployOracleHash)
				deployOracle.TxHash = deployOracleHash
				oracleAddr = deployOracle.Address.String()
				chain33txLog.Info("deployAndInit2Chain33", "Succeed to deploy Oracle with address =", oracleAddr)
				goto deployBridgeBank
			}
		}
	}
	/////////////////////////////////////
deployBridgeBank:
	//constructor (
	//	address _operatorAddress,
	//	address _oracleAddress,
	//	address _ethereumBridgeAddress
	//)
	constructorPara = fmt.Sprintf("constructor(%s, %s, %s)", para4deploy.Operator.String(), oracleAddr, ethereumBridgeAddr)
	deployBridgeBankHash, err := deploySingleContract(ethcommon.FromHex(generated.BridgeBankBin), generated.BridgeBankABI, constructorPara, "BridgeBank", paraChainName, para4deploy.Deployer.String(), rpcLaddr)
	if nil != err {
		chain33txLog.Error("DeployAndInit", "failed to DeployBridgeBank due to:", err.Error())
		return nil, err
	}
	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("deployBridgeBank timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(deployBridgeBankHash, rpcLaddr)
				if data == "" {
					chain33txLog.Info("deployAndInit2Chain33", "No receipt received yet for Deploy BridgeBank tx and continue to wait", "continue")
					continue
				} else if data != "2" {
					return nil, errors.New("Deploy BridgeBank failed due to" + ", ty = " + data)
				}
				deployBridgeBank.Address = getContractAddr(deployer, deployBridgeBankHash)
				deployBridgeBank.TxHash = deployOracleHash
				bridgeBankAddr = deployBridgeBank.Address.String()
				chain33txLog.Info("deployAndInit2Chain33", "Succeed to deploy BridgeBank with address =", bridgeBankAddr)
				goto settingBridgeBank
			}
		}
	}

settingBridgeBank:
	////////////////////////
	//function setBridgeBank(
	//	address payable _bridgeBank
	//)
	callPara := fmt.Sprintf("setBridgeBank(%s)", bridgeBankAddr)
	_, packData, err := evmAbi.Pack(callPara, generated.EthereumBridgeABI, false)
	if nil != err {
		chain33txLog.Info("setBridgeBank", "Failed to do abi.Pack due to:", err.Error())
		return nil, ebrelayerTypes.ErrPack
	}
	settingBridgeBankHash, err := sendTx2Evm(packData, rpcLaddr, ethereumBridgeAddr, paraChainName, deployer)
	if nil != err {
		chain33txLog.Error("DeployAndInit", "failed to settingBridgeBank due to:", err.Error())
		return nil, err
	}
	{
		fmt.Println("setBridgeBank tx hash:", settingBridgeBankHash)
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("setBridgeBank timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(settingBridgeBankHash, rpcLaddr)
				if data == "" {
					chain33txLog.Info("deployAndInit2Chain33", "No receipt received yet for for setBridgeBank tx and continue to wait", "continue")
					continue
				} else if data != "2" {
					return nil, errors.New("setBridgeBank failed due to" + ", ty = " + data)
				}
				chain33txLog.Info("deployAndInit2Chain33", "Succeed to setBridgeBank ", "Oh yea!!!")
				goto setOracle
			}
		}
	}

setOracle:
	//function setOracle(
	//	address _oracle
	//)
	callPara = fmt.Sprintf("setOracle(%s)", oracleAddr)
	_, packData, err = evmAbi.Pack(callPara, generated.EthereumBridgeABI, false)
	if nil != err {
		chain33txLog.Info("setOracle", "Failed to do abi.Pack due to:", err.Error())
		return nil, ebrelayerTypes.ErrPack
	}
	setOracleHash, err := sendTx2Evm(packData, rpcLaddr, ethereumBridgeAddr, paraChainName, deployer)
	if nil != err {
		chain33txLog.Error("DeployAndInit", "failed to setOracle due to:", err.Error())
		return nil, err
	}
	{
		fmt.Println("setOracle tx hash:", setOracleHash)
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("setOracle timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(setOracleHash, rpcLaddr)
				if data == "" {
					chain33txLog.Info("deployAndInit2Chain33", "No receipt received yet for for setOracle tx and continue to wait", "continue")
					continue
				} else if data != "2" {
					return nil, errors.New("setOracle failed due to" + ", ty = " + data)
				}
				chain33txLog.Info("deployAndInit2Chain33", "Succeed to setOracle ", "Oh yea!!!")
				goto deployBridgeRegistry
			}
		}
	}

deployBridgeRegistry:
	//constructor(
	//	address _ethereumBridge,
	//	address _bridgeBank,
	//	address _oracle,
	//	address _valset
	//)
	constructorPara = fmt.Sprintf("constructor(%s, %s, %s, %s)", ethereumBridgeAddr, bridgeBankAddr, oracleAddr, valsetAddr)
	deployBridgeRegistryHash, err := deploySingleContract(ethcommon.FromHex(generated.BridgeRegistryBin), generated.BridgeRegistryABI, constructorPara, "BridgeRegistry", paraChainName, para4deploy.Deployer.String(), rpcLaddr)
	if nil != err {
		chain33txLog.Error("DeployAndInit", "failed to deployBridgeRegistry due to:", err.Error())
		return nil, err
	}
	{
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("deployBridgeRegistry timeout")
			case <-oneSecondtimeout.C:
				data, _ := getTxByHashesRpc(deployBridgeRegistryHash, rpcLaddr)
				if data == "" {
					chain33txLog.Info("deployAndInit2Chain33", "No receipt received yet for for deploy BridgeRegistry tx and continue to wait", "continue")
					continue
				} else if data != "2" {
					return nil, errors.New("deployBridgeRegistry failed due to" + ", ty = " + data)
				}

				deployBridgeRegistry.Address = getContractAddr(deployer, deployBridgeRegistryHash)
				deployBridgeRegistry.TxHash = deployBridgeRegistryHash

				chain33txLog.Info("deployAndInit2Chain33", "Succeed to deployBridgeRegistry with address", deployBridgeRegistry.Address.String())
				goto finished
			}
		}
	}
finished:
	return deployInfo, nil
}

func deployMulSign2Chain33(rpcLaddr, paraChainName, deployer string) (string, error) {
	deployMulSign, err := deploySingleContract(ethcommon.FromHex(generated.GnosisSafeBin), generated.GnosisSafeABI, "", "mul-sign", paraChainName, deployer, rpcLaddr)
	if nil != err {
		chain33txLog.Error("deployMulSign2Chain33", "failed to deployMulSign due to:", err.Error())
		return "", err
	}

	timeout := time.NewTimer(300 * time.Second)
	oneSecondtimeout := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeout.C:
			panic("deployMulSign timeout")
		case <-oneSecondtimeout.C:
			data, _ := getTxByHashesRpc(deployMulSign, rpcLaddr)
			if data == "" {
				chain33txLog.Info("deployMulSign2Chain33", "No receipt received yet for Deploy MulSign tx and continue to wait", "continue")
				continue
			} else if data != "2" {
				return "", errors.New("Deploy MulSign failed due to" + ", ty = " + data)
			}

			address := getContractAddr(deployer, deployMulSign)
			mulSignAddr := address.String()
			chain33txLog.Info("deployMulSign2Chain33", "Succeed to deploy MulSign with address =", mulSignAddr)
			return mulSignAddr, nil
		}
	}
}

func deployERC20ToChain33(rpcLaddr, paraChainName, deployer string, param ebTypes.ERC20Token) (string, error) {
	constructorPara := "constructor(" + param.Symbol + "," + param.Symbol + "," + param.Amount + "," + param.Owner + ")"
	deployErc20, err := deploySingleContract(ethcommon.FromHex(erc20.ERC20Bin), erc20.ERC20ABI, constructorPara, "Erc20:"+param.Symbol, paraChainName, deployer, rpcLaddr)
	if nil != err {
		chain33txLog.Error("deployERC20ToChain33", "failed to deployMulSign due to:", err.Error())
		return "", err
	}

	timeout := time.NewTimer(300 * time.Second)
	oneSecondtimeout := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeout.C:
			panic("deployMulSign timeout")
		case <-oneSecondtimeout.C:
			data, _ := getTxByHashesRpc(deployErc20, rpcLaddr)
			if data == "" {
				chain33txLog.Info("deployERC20ToChain33", "No receipt received yet for Deploy erc20 tx and continue to wait", "continue")
				continue
			} else if data != "2" {
				return "", errors.New("Deploy erc20 failed due to" + ", ty = " + data)
			}

			address := getContractAddr(deployer, deployErc20)
			erc20Addr := address.String()
			chain33txLog.Info("deployERC20ToChain33", "Succeed to deploy erc20 with address =", erc20Addr)
			return erc20Addr, nil
		}
	}
}
