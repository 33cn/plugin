package ethtxs

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"testing"

	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LoadABI(t *testing.T) {
	abi1 := LoadABI(Chain33BankABI)
	abi2 := LoadABI(Chain33BridgeABI)
	abi3 := LoadABI(EthereumBankABI)
	assert.NotEmpty(t, abi1, abi2, abi3)
}

func TestContractRegistry_String(t *testing.T) {
	assert.Equal(t, Valset.String(), "valset")
	assert.Equal(t, Oracle.String(), "oracle")
	assert.Equal(t, BridgeBank.String(), "bridgebank")
	assert.Equal(t, Chain33Bridge.String(), "chain33bridge")
}

func Test_GetAddressFromBridgeRegistry(t *testing.T) {
	genesiskey, _ := crypto.GenerateKey()
	alloc := make(core.GenesisAlloc)
	genesisAddr := crypto.PubkeyToAddress(genesiskey.PublicKey)
	genesisAccount := core.GenesisAccount{
		Balance:    big.NewInt(1000000000000 * 10000),
		PrivateKey: crypto.FromECDSA(genesiskey),
	}
	alloc[genesisAddr] = genesisAccount
	gasLimit := uint64(100000000)
	sim := new(ethinterface.SimExtend)
	sim.SimulatedBackend = backends.NewSimulatedBackend(alloc, gasLimit)

	bridgebankTest := ContractRegistry(5)
	_, err := GetAddressFromBridgeRegistry(sim, genesisAddr, genesisAddr, bridgebankTest)
	require.NotNil(t, err)
}

func Test_RelayOracleClaimToEthereum(t *testing.T) {
	para, sim, _, x2EthDeployInfo, err := deployContracts()
	require.NoError(t, err)

	privateKeySlice, err := chain33Common.FromHex("0x3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e")
	require.Nil(t, err)
	privateKey, err := crypto.ToECDSA(privateKeySlice)
	require.Nil(t, err)

	prophecyClaim := ProphecyClaim{
		ClaimType:            events.ClaimTypeBurn,
		Chain33Sender:        []byte("12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"),
		EthereumReceiver:     common.HexToAddress("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF"),
		TokenContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Symbol:               "eth",
		Amount:               big.NewInt(100000000000000000),
		Chain33TxHash:        common.Hex2Bytes("fd5747c43d1460bb6f8a7a26c66b4ccab5500d05668278efe5c0fd5951dfd909"),
	}

	Addr2TxNonce := make(map[common.Address]*NonceMutex)
	oracleInstance, err := GetOracleInstance(sim, x2EthDeployInfo.BridgeRegistry.Address)
	require.Nil(t, err)

	burnOrLockParameter := &BurnOrLockParameter{
		ClientSpec:   &EthClientWithUrl{Client: sim, OracleInstance: oracleInstance},
		Sender:       para.InitValidators[0],
		TokenOnEth:   common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Claim:        prophecyClaim,
		PrivateKey:   privateKey,
		Addr2TxNonce: Addr2TxNonce,
		ChainId:      big.NewInt(1337),
	}

	_, err = RelayOracleClaimToEthereum(burnOrLockParameter)
	require.Nil(t, err)

	_, err = revokeNonce(para.Operator)
	require.Nil(t, err)
}

func deployContracts() (*DeployPara, *ethinterface.SimExtend, *X2EthContracts, *X2EthDeployInfo, error) {
	// 0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a
	deployerPrivateKey := "8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
	// 0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f
	ethValidatorAddrKeyA := "3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
	ethValidatorAddrKeyB := "a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
	ethValidatorAddrKeyC := "bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
	ethValidatorAddrKeyD := "c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"

	ethValidatorAddrKeys := make([]string, 0)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyA)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyB)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyC)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyD)

	ctx := context.Background()
	//var backend bind.ContractBackend
	backend, para := PrepareTestEnvironment(deployerPrivateKey, ethValidatorAddrKeys)
	sim := new(ethinterface.SimExtend)
	sim.SimulatedBackend = backend.(*backends.SimulatedBackend)

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

func PrepareTestEnvironment(deployerPrivateKey string, ethValidatorAddrKeys []string) (bind.ContractBackend, *DeployPara) {
	genesiskey, _ := crypto.HexToECDSA(deployerPrivateKey)
	alloc := make(core.GenesisAlloc)
	genesisAddr := crypto.PubkeyToAddress(genesiskey.PublicKey)
	genesisAccount := core.GenesisAccount{
		Balance:    big.NewInt(1000000000000 * 10000),
		PrivateKey: crypto.FromECDSA(genesiskey),
	}
	alloc[genesisAddr] = genesisAccount

	var InitValidators []common.Address
	var ValidatorPriKey []*ecdsa.PrivateKey
	for _, v := range ethValidatorAddrKeys {
		key, _ := crypto.HexToECDSA(v)
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
	sim := backends.NewSimulatedBackend(alloc, gasLimit)

	InitPowers := []*big.Int{big.NewInt(80), big.NewInt(10), big.NewInt(10), big.NewInt(10)}

	para := &DeployPara{
		DeployPrivateKey: genesiskey,
		Deployer:         genesisAddr,
		Operator:         genesisAddr,
		InitValidators:   InitValidators,
		ValidatorPriKey:  ValidatorPriKey,
		InitPowers:       InitPowers,
	}

	return sim, para
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
	}

	x2EthContracts.Chain33Bridge, deployInfo.Chain33Bridge, err = DeployChain33Bridge(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Valset.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployChain33Bridge due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	x2EthContracts.Oracle, deployInfo.Oracle, err = DeployOracle(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Valset.Address, deployInfo.Chain33Bridge.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployOracle due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	x2EthContracts.BridgeBank, deployInfo.BridgeBank, err = DeployBridgeBank(client, para.DeployPrivateKey, para.Deployer, para.Operator, deployInfo.Oracle.Address, deployInfo.Chain33Bridge.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	auth, err := PrepareAuth(client, para.DeployPrivateKey, para.Deployer)
	if nil != err {
		return nil, nil, err
	}
	_, err = x2EthContracts.Chain33Bridge.SetBridgeBank(auth, deployInfo.BridgeBank.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to SetBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	auth, err = PrepareAuth(client, para.DeployPrivateKey, para.Deployer)
	if nil != err {
		return nil, nil, err
	}
	_, err = x2EthContracts.Chain33Bridge.SetOracle(auth, deployInfo.Oracle.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to SetOracle due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	x2EthContracts.BridgeRegistry, deployInfo.BridgeRegistry, err = DeployBridgeRegistry(client, para.DeployPrivateKey, para.Deployer, deployInfo.Chain33Bridge.Address, deployInfo.BridgeBank.Address, deployInfo.Oracle.Address, deployInfo.Valset.Address)
	if nil != err {
		deployLog.Error("DeployAndInit", "failed to DeployBridgeBank due to:", err.Error())
		return nil, nil, err
	}
	if isSim {
		sim.Commit()
	}

	return x2EthContracts, deployInfo, nil
}

func Test_SelectAndRoundEthURL(t *testing.T) {
	urls := []string{"url1", "url2"}
	decision, err := SelectAndRoundEthURL(&urls)
	assert.Equal(t, nil, err)
	assert.Equal(t, "url1", decision)
	assert.Equal(t, "url2", urls[0])
	assert.Equal(t, "url1", urls[1])

}
