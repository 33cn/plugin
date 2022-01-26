package setup

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethtxs"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

//PrepareTestEnv ...
func PrepareTestEnv() (*ethinterface.SimExtend, *ethtxs.DeployPara) {
	genesiskey, _ := crypto.GenerateKey()
	alloc := make(core.GenesisAlloc)
	genesisAddr := crypto.PubkeyToAddress(genesiskey.PublicKey)
	genesisAccount := core.GenesisAccount{
		Balance:    big.NewInt(1000000000000 * 10000),
		PrivateKey: crypto.FromECDSA(genesiskey),
	}
	alloc[genesisAddr] = genesisAccount

	var InitValidators []common.Address
	var ValidatorPriKey []*ecdsa.PrivateKey
	for i := 0; i < 4; i++ {
		key, _ := crypto.GenerateKey()
		addr := crypto.PubkeyToAddress(key.PublicKey)
		InitValidators = append(InitValidators, addr)
		ValidatorPriKey = append(ValidatorPriKey, key)

		account := core.GenesisAccount{
			Balance:    big.NewInt(1000000000000 * 10000),
			PrivateKey: crypto.FromECDSA(key),
		}
		alloc[addr] = account
	}
	gasLimit := uint64(100000000)
	sim := new(ethinterface.SimExtend)
	sim.SimulatedBackend = backends.NewSimulatedBackend(alloc, gasLimit)

	InitPowers := []*big.Int{big.NewInt(80), big.NewInt(10), big.NewInt(10), big.NewInt(10)}
	para := &ethtxs.DeployPara{
		DeployPrivateKey: genesiskey,
		Deployer:         genesisAddr,
		Operator:         genesisAddr,
		InitValidators:   InitValidators,
		ValidatorPriKey:  ValidatorPriKey,
		InitPowers:       InitPowers,
	}

	return sim, para
}

//DeployContracts ...
func DeployContracts() (*ethtxs.DeployPara, *ethinterface.SimExtend, *ethtxs.X2EthContracts, *ethtxs.X2EthDeployInfo, error) {
	ctx := context.Background()
	sim, para := PrepareTestEnv()

	opts, _ := bind.NewKeyedTransactorWithChainID(para.DeployPrivateKey, big.NewInt(1337))
	parsed, _ := abi.JSON(strings.NewReader(generated.BridgeBankBin))
	contractAddr, _, _, _ := bind.DeployContract(opts, parsed, common.FromHex(generated.BridgeBankBin), sim)
	sim.Commit()

	callMsg := ethereum.CallMsg{
		From: para.Deployer,
		To:   &contractAddr,
		Data: common.FromHex(generated.BridgeBankBin),
	}

	_, err := sim.EstimateGas(ctx, callMsg)
	if nil != err {
		panic("failed to estimate gas due to:" + err.Error())
	}
	x2EthContracts, x2EthDeployInfo, err := DeployAndInit(sim, para)
	if nil != err {
		return nil, nil, nil, nil, err
	}
	sim.Commit()

	return para, sim, x2EthContracts, x2EthDeployInfo, nil
}

//DeployValset : 部署Valset
func DeployValset(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer common.Address, operator common.Address, initValidators []common.Address, initPowers []*big.Int) (*generated.Valset, *ethtxs.DeployResult, error) {
	auth, err := ethtxs.PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, valset, err := generated.DeployValset(auth, client, operator, initValidators, initPowers)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &ethtxs.DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}

	return valset, deployResult, nil
}

//DeployChain33Bridge : 部署Chain33Bridge
func DeployChain33Bridge(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer common.Address, operator, valset common.Address) (*generated.Chain33Bridge, *ethtxs.DeployResult, error) {
	auth, err := ethtxs.PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, chain33Bridge, err := generated.DeployChain33Bridge(auth, client, operator, valset)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &ethtxs.DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}
	return chain33Bridge, deployResult, nil
}

//DeployOracle : 部署Oracle
func DeployOracle(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer, operator, valset, chain33Bridge common.Address) (*generated.Oracle, *ethtxs.DeployResult, error) {
	auth, err := ethtxs.PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, oracle, err := generated.DeployOracle(auth, client, operator, valset, chain33Bridge)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &ethtxs.DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}
	return oracle, deployResult, nil
}

//DeployBridgeBank : 部署BridgeBank
func DeployBridgeBank(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer, operator, oracle, chain33Bridge common.Address) (*generated.BridgeBank, *ethtxs.DeployResult, error) {
	auth, err := ethtxs.PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, bridgeBank, err := generated.DeployBridgeBank(auth, client, operator, oracle, chain33Bridge)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &ethtxs.DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}
	return bridgeBank, deployResult, nil
}

//DeployBridgeRegistry : 部署BridgeRegistry
func DeployBridgeRegistry(client ethinterface.EthClientSpec, privateKey *ecdsa.PrivateKey, deployer, chain33BridgeAddr, bridgeBankAddr, oracleAddr, valsetAddr common.Address) (*generated.BridgeRegistry, *ethtxs.DeployResult, error) {
	auth, err := ethtxs.PrepareAuth(client, privateKey, deployer)
	if nil != err {
		return nil, nil, err
	}

	//部署合约
	addr, tx, bridgeRegistry, err := generated.DeployBridgeRegistry(auth, client, chain33BridgeAddr, bridgeBankAddr, oracleAddr, valsetAddr)
	if err != nil {
		return nil, nil, err
	}

	deployResult := &ethtxs.DeployResult{
		Address: addr,
		TxHash:  tx.Hash().String(),
	}
	return bridgeRegistry, deployResult, nil
}

//DeployAndInit ...
func DeployAndInit(client ethinterface.EthClientSpec, para *ethtxs.DeployPara) (*ethtxs.X2EthContracts, *ethtxs.X2EthDeployInfo, error) {
	x2EthContracts := &ethtxs.X2EthContracts{}
	deployInfo := &ethtxs.X2EthDeployInfo{}
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
		fmt.Println("DeployAndInit", "failed to DeployValset due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	x2EthContracts.Chain33Bridge, deployInfo.Chain33Bridge, err = DeployChain33Bridge(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Valset.Address)
	if nil != err {
		fmt.Println("DeployAndInit", "failed to DeployChain33Bridge due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	x2EthContracts.Oracle, deployInfo.Oracle, err = DeployOracle(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Valset.Address, deployInfo.Chain33Bridge.Address)
	if nil != err {
		fmt.Println("DeployAndInit", "failed to DeployOracle due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	x2EthContracts.BridgeBank, deployInfo.BridgeBank, err = DeployBridgeBank(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Oracle.Address, deployInfo.Chain33Bridge.Address)
	if nil != err {
		fmt.Println("DeployAndInit", "failed to DeployBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	auth, err := ethtxs.PrepareAuth(client, para.DeployPrivateKey, para.Deployer)
	if nil != err {
		return nil, nil, err
	}
	_, err = x2EthContracts.Chain33Bridge.SetBridgeBank(auth, deployInfo.BridgeBank.Address)
	if nil != err {
		fmt.Println("DeployAndInit", "failed to SetBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	auth, err = ethtxs.PrepareAuth(client, para.DeployPrivateKey, para.Deployer)
	if nil != err {
		return nil, nil, err
	}
	_, err = x2EthContracts.Chain33Bridge.SetOracle(auth, deployInfo.Oracle.Address)
	if nil != err {
		fmt.Println("DeployAndInit", "failed to SetOracle due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	x2EthContracts.BridgeRegistry, deployInfo.BridgeRegistry, err = DeployBridgeRegistry(client, para.DeployPrivateKey, para.Deployer, deployInfo.Chain33Bridge.Address, deployInfo.BridgeBank.Address, deployInfo.Oracle.Address, deployInfo.Valset.Address)
	if nil != err {
		fmt.Println("DeployAndInit", "failed to DeployBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	auth, err = ethtxs.PrepareAuth(client, para.DeployPrivateKey, para.Deployer)
	if nil != err {
		return nil, nil, err
	}
	_, err = x2EthContracts.BridgeBank.ConfigplatformTokenSymbol(auth, "ETH")
	if nil != err {
		fmt.Println("DeployAndInit", "failed to ConfigplatformTokenSymbol due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	return x2EthContracts, deployInfo, nil
}
