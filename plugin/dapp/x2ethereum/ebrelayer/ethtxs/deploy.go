package ethtxs

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethinterface"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	deployLog = log15.New("contract deployer", "deployer")
)

//DeployResult ...
type DeployResult struct {
	Address common.Address
	TxHash  string
}

//X2EthContracts ...
type X2EthContracts struct {
	BridgeRegistry *generated.BridgeRegistry
	BridgeBank     *generated.BridgeBank
	Chain33Bridge  *generated.Chain33Bridge
	Valset         *generated.Valset
	Oracle         *generated.Oracle
}

//X2EthDeployInfo ...
type X2EthDeployInfo struct {
	BridgeRegistry *DeployResult
	BridgeBank     *DeployResult
	Chain33Bridge  *DeployResult
	Valset         *DeployResult
	Oracle         *DeployResult
}

//DeployPara ...
type DeployPara struct {
	DeployPrivateKey *ecdsa.PrivateKey
	Deployer         common.Address
	Operator         common.Address
	InitValidators   []common.Address
	ValidatorPriKey  []*ecdsa.PrivateKey
	InitPowers       []*big.Int
}

//OperatorInfo ...
type OperatorInfo struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

//DeployValset : 部署Valset
func DeployValset(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer common.Address, operator common.Address, initValidators []common.Address, initPowers []*big.Int) (*generated.Valset, *DeployResult, error) {
	auth, err := PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, valset, err := generated.DeployValset(auth, client, operator, initValidators, initPowers)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}

	return valset, deployResult, nil
}

//DeployChain33Bridge : 部署Chain33Bridge
func DeployChain33Bridge(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer common.Address, operator, valset common.Address) (*generated.Chain33Bridge, *DeployResult, error) {
	auth, err := PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, chain33Bridge, err := generated.DeployChain33Bridge(auth, client, operator, valset)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}
	return chain33Bridge, deployResult, nil
}

//DeployOracle : 部署Oracle
func DeployOracle(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer, operator, valset, chain33Bridge common.Address) (*generated.Oracle, *DeployResult, error) {
	auth, err := PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, oracle, err := generated.DeployOracle(auth, client, operator, valset, chain33Bridge)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}
	return oracle, deployResult, nil
}

//DeployBridgeBank : 部署BridgeBank
func DeployBridgeBank(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer, operator, oracle, chain33Bridge common.Address) (*generated.BridgeBank, *DeployResult, error) {
	auth, err := PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, bridgeBank, err := generated.DeployBridgeBank(auth, client, operator, oracle, chain33Bridge)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}
	return bridgeBank, deployResult, nil
}

//DeployBridgeRegistry : 部署BridgeRegistry
func DeployBridgeRegistry(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer, chain33BridgeAddr, bridgeBankAddr, oracleAddr, valsetAddr common.Address) (*generated.BridgeRegistry, *DeployResult, error) {
	auth, err := PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, bridgeRegistry, err := generated.DeployBridgeRegistry(auth, client, chain33BridgeAddr, bridgeBankAddr, oracleAddr, valsetAddr)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}
	return bridgeRegistry, deployResult, nil
}

//DeployAndInit ...
func DeployAndInit(client ethinterface.EthClientSpec, para *DeployPara) (*X2EthContracts, *X2EthDeployInfo, error) {
	x2EthContracts := &X2EthContracts{}
	deployInfo := &X2EthDeployInfo{}
	var err error

	/////////////////////////////////////
	sim, isSim := client.(*ethinterface.SimExtend)
	if isSim {
		fmt.Println("Use the simulator")
	} else {
		fmt.Println("Use the actual Ethereum")
	}

	x2EthContracts.Valset, deployInfo.Valset, err = DeployValset(client, para.DeployPrivateKey, para.Deployer, para.Operator, para.InitValidators, para.InitPowers)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployValset due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	} else {
		fmt.Println("\nDeployValset tx hash:", deployInfo.Valset.TxHash)
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeployValset timeout")
			case <-oneSecondtimeout.C:
				_, err := client.TransactionReceipt(context.Background(), common.HexToHash(deployInfo.Valset.TxHash))
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received yet for DeployValset tx and continue to wait")
					continue
				} else if err != nil {
					panic("DeployValset failed due to" + err.Error())
				}

				callopts := &bind.CallOpts{
					Pending: true,
					From:    para.Deployer,
					Context: context.Background(),
				}
				operator, err := x2EthContracts.Valset.Operator(callopts)
				if nil != err {
					panic(err.Error())
				}

				if operator.String() != para.Operator.String() {
					fmt.Printf("operator queried from valset is:%s, and setted is:%s", operator.String(), para.Operator.String())
					panic("operator query is not same as setted ")
				}
				goto deployChain33Bridge
			}
		}
	}

deployChain33Bridge:
	x2EthContracts.Chain33Bridge, deployInfo.Chain33Bridge, err = DeployChain33Bridge(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Valset.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployChain33Bridge due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	} else {
		fmt.Printf("\n\nGoing to DeployChain33Bridge with valset address:%s", deployInfo.Valset.Address.String())
		fmt.Println("\nDeployChain33Bridge tx hash:", deployInfo.Chain33Bridge.TxHash)
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeployChain33Bridge timeout")
			case <-oneSecondtimeout.C:
				_, err := client.TransactionReceipt(context.Background(), common.HexToHash(deployInfo.Chain33Bridge.TxHash))
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received for DeployChain33Bridge tx and continue to wait")
					continue
				} else if err != nil {
					panic("DeployChain33Bridge failed due to" + err.Error())
				}

				callopts := &bind.CallOpts{
					Pending: true,
					From:    para.Deployer,
					Context: context.Background(),
				}
				operator, err := x2EthContracts.Chain33Bridge.Operator(callopts)
				if nil != err {
					panic(err.Error())
				}

				if operator.String() != para.Operator.String() {
					fmt.Printf("operator queried from valset is:%s, and setted is:%s", operator.String(), para.Operator.String())
					panic("operator query is not same as setted ")
				}
				goto deployOracle
			}
		}
	}

deployOracle:
	x2EthContracts.Oracle, deployInfo.Oracle, err = DeployOracle(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Valset.Address, deployInfo.Chain33Bridge.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployOracle due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	} else {
		fmt.Println("DeployOracle tx hash:", deployInfo.Oracle.TxHash)
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeployOracle timeout")
			case <-oneSecondtimeout.C:
				_, err := client.TransactionReceipt(context.Background(), common.HexToHash(deployInfo.Oracle.TxHash))
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received for DeployOracle tx and continue to wait")
					continue
				} else if err != nil {
					panic("DeployOracle failed due to" + err.Error())
				}

				callopts := &bind.CallOpts{
					Pending: true,
					From:    para.Deployer,
					Context: context.Background(),
				}
				operator, err := x2EthContracts.Oracle.Operator(callopts)
				if nil != err {
					panic(err.Error())
				}

				if operator.String() != para.Operator.String() {
					fmt.Printf("operator queried from valset is:%s, and setted is:%s", operator.String(), para.Operator.String())
					panic("operator query is not same as setted ")
				}
				goto deployBridgeBank
			}
		}
	}
	/////////////////////////////////////
deployBridgeBank:
	x2EthContracts.BridgeBank, deployInfo.BridgeBank, err = DeployBridgeBank(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Oracle.Address, deployInfo.Chain33Bridge.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	} else {
		fmt.Println("DeployBridgeBank tx hash:", deployInfo.BridgeBank.TxHash)
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeployBridgeBank timeout")
			case <-oneSecondtimeout.C:
				_, err := client.TransactionReceipt(context.Background(), common.HexToHash(deployInfo.BridgeBank.TxHash))
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received for DeployOracle tx and continue to wait")
					continue
				} else if err != nil {
					panic("DeployBridgeBank failed due to" + err.Error())
				}

				callopts := &bind.CallOpts{
					Pending: true,
					From:    para.Deployer,
					Context: context.Background(),
				}
				operator, err := x2EthContracts.BridgeBank.Operator(callopts)
				if nil != err {
					panic(err.Error())
				}

				if operator.String() != para.Operator.String() {
					fmt.Printf("operator queried from valset is:%s, and setted is:%s", operator.String(), para.Operator.String())
					panic("operator query is not same as setted ")
				}
				goto settingBridgeBank
			}
		}
	}

settingBridgeBank:
	////////////////////////
	auth, err := PrepareAuth(client, para.DeployPrivateKey, para.Deployer)
	if nil != err {
		return nil, nil, err
	}
	setBridgeBankTx, err := x2EthContracts.Chain33Bridge.SetBridgeBank(auth, deployInfo.BridgeBank.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to SetBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	} else {
		fmt.Println("setBridgeBankTx tx hash:", setBridgeBankTx.Hash().String())
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("setBridgeBankTx timeout")
			case <-oneSecondtimeout.C:
				_, err := client.TransactionReceipt(context.Background(), setBridgeBankTx.Hash())
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received for setBridgeBankTx tx and continue to wait")
					continue
				} else if err != nil {
					panic("setBridgeBankTx failed due to" + err.Error())
				}

				callopts := &bind.CallOpts{
					Pending: true,
					From:    para.Deployer,
					Context: context.Background(),
				}
				yes, err := x2EthContracts.Chain33Bridge.HasBridgeBank(callopts)
				if nil != err {
					panic(err.Error())
				}

				if !yes {
					fmt.Printf("BridgeBank doesn't exist")
					panic("BridgeBank doesn't exist")
				}
				goto setOracle
			}
		}
	}

setOracle:
	auth, err = PrepareAuth(client, para.DeployPrivateKey, para.Deployer)
	if nil != err {
		return nil, nil, err
	}
	setOracleTx, err := x2EthContracts.Chain33Bridge.SetOracle(auth, deployInfo.Oracle.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to SetOracle due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	} else {
		fmt.Println("setOracleTx tx hash:", setOracleTx.Hash().String())
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("setBridgeBankTx timeout")
			case <-oneSecondtimeout.C:
				_, err := client.TransactionReceipt(context.Background(), setOracleTx.Hash())
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received for setOracleTx tx and continue to wait")
					continue
				} else if err != nil {
					panic("setOracleTx failed due to" + err.Error())
				}

				callopts := &bind.CallOpts{
					Pending: true,
					From:    para.Deployer,
					Context: context.Background(),
				}
				yes, err := x2EthContracts.Chain33Bridge.HasOracle(callopts)
				if nil != err {
					panic(err.Error())
				}

				if !yes {
					fmt.Printf("oracle doesn't exist")
					panic("oracle doesn't exist")
				}
				goto deployBridgeRegistry
			}
		}
	}

deployBridgeRegistry:
	x2EthContracts.BridgeRegistry, deployInfo.BridgeRegistry, err = DeployBridgeRegistry(client, para.DeployPrivateKey, para.Deployer, deployInfo.Chain33Bridge.Address, deployInfo.BridgeBank.Address, deployInfo.Oracle.Address, deployInfo.Valset.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	} else {
		fmt.Println("DeployBridgeRegistry tx hash:", deployInfo.BridgeRegistry.TxHash)
		timeout := time.NewTimer(300 * time.Second)
		oneSecondtimeout := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeout.C:
				panic("DeployBridgeRegistry timeout")
			case <-oneSecondtimeout.C:
				_, err := client.TransactionReceipt(context.Background(), common.HexToHash(deployInfo.BridgeRegistry.TxHash))
				if err == ethereum.NotFound {
					fmt.Println("\n No receipt received for DeployOracle tx and continue to wait")
					continue
				} else if err != nil {
					panic("DeployBridgeRegistry failed due to" + err.Error())
				}

				callopts := &bind.CallOpts{
					Pending: true,
					From:    para.Deployer,
					Context: context.Background(),
				}
				oracleAddr, err := x2EthContracts.BridgeRegistry.Oracle(callopts)
				if nil != err {
					panic(err.Error())
				}

				if oracleAddr.String() != deployInfo.Oracle.Address.String() {
					fmt.Printf("oracleAddr queried from BridgeRegistry is:%s, and setted is:%s", oracleAddr.String(), deployInfo.Oracle.Address.String())
					panic("oracleAddr query is not same as setted ")
				}
				goto finished
			}
		}
	}
finished:

	return x2EthContracts, deployInfo, nil
}
