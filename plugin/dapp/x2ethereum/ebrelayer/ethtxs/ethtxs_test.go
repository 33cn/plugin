package ethtxs

//
//import (
//	"context"
//	"crypto/ecdsa"
//	"math/big"
//	"testing"
//
//	chain33Common "github.com/33cn/chain33/common"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethinterface"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
//	"github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core"
//	"github.com/ethereum/go-ethereum/crypto"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//)
//
//func Test_LoadABI(t *testing.T) {
//	abi1 := LoadABI(Chain33BankABI)
//	abi2 := LoadABI(Chain33BridgeABI)
//	abi3 := LoadABI(EthereumBankABI)
//	assert.NotEmpty(t, abi1, abi2, abi3)
//}
//
//func Test_isWebsocketURL(t *testing.T) {
//	bret := isWebsocketURL("ws://127.0.0.1:7545/")
//	assert.Equal(t, bret, true)
//
//	bret = isWebsocketURL("https://127.0.0.1:7545/")
//	assert.Equal(t, bret, false)
//}
//
//func TestContractRegistry_String(t *testing.T) {
//	assert.Equal(t, Valset.String(), "valset")
//	assert.Equal(t, Oracle.String(), "oracle")
//	assert.Equal(t, BridgeBank.String(), "bridgebank")
//	assert.Equal(t, Chain33Bridge.String(), "chain33bridge")
//}
//
//func Test_GetAddressFromBridgeRegistry(t *testing.T) {
//	genesiskey, _ := crypto.GenerateKey()
//	alloc := make(core.GenesisAlloc)
//	genesisAddr := crypto.PubkeyToAddress(genesiskey.PublicKey)
//	genesisAccount := core.GenesisAccount{
//		Balance:    big.NewInt(10000000000 * 10000),
//		PrivateKey: crypto.FromECDSA(genesiskey),
//	}
//	alloc[genesisAddr] = genesisAccount
//	gasLimit := uint64(100000000)
//	sim := new(ethinterface.SimExtend)
//	sim.SimulatedBackend = backends.NewSimulatedBackend(alloc, gasLimit)
//
//	bridgebankTest := ContractRegistry(5)
//	_, err := GetAddressFromBridgeRegistry(sim, genesisAddr, genesisAddr, bridgebankTest)
//	require.NotNil(t, err)
//}
//
//func Test_RelayOracleClaimToEthereum(t *testing.T) {
//	para, sim, x2EthContracts, _, err := deployContracts()
//	require.NoError(t, err)
//
//	claimType := events.MsgBurn
//	privateKeySlice, err := chain33Common.FromHex("0x3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e")
//	require.Nil(t, err)
//	privateKey, err := crypto.ToECDSA(privateKeySlice)
//	require.Nil(t, err)
//
//	prophecyClaim := ProphecyClaim{
//		ClaimType:            events.MsgBurn,
//		Chain33Sender:        []byte("12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"),
//		EthereumReceiver:     common.HexToAddress("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF"),
//		TokenContractAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
//		Symbol:               "eth",
//		Amount:               big.NewInt(100000000000000000),
//	}
//
//	chain33TxHash := common.Hex2Bytes("fd5747c43d1460bb6f8a7a26c66b4ccab5500d05668278efe5c0fd5951dfd909")
//	txhash, err := RelayOracleClaimToEthereum(x2EthContracts.Oracle, sim, para.InitValidators[0], claimType, prophecyClaim, privateKey, chain33TxHash)
//	require.Nil(t, err)
//	assert.Equal(t, txhash, "0x6fa087c7a2a8a4421f6e269fbc6c0838e99fa59d5760155a71cd7eb1c01aafad")
//
//	//hash := "0xc0c22aa6198fdde0dbe47ddadbe449f736b82ed4a498871de5d5f4ad9ae122a0"
//	//status := GetEthTxStatus(sim, common.HexToHash(hash))
//	//assert.Equal(t, status, EthTxPending.String())
//
//	_, err = revokeNonce(para.Operator)
//	require.Nil(t, err)
//}
//
//func Test_revokeNonce(t *testing.T) {
//
//}
//
//func deployContracts() (*DeployPara, *ethinterface.SimExtend, *X2EthContracts, *X2EthDeployInfo, error) {
//	// 0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a
//	deployerPrivateKey := "8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
//	// 0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f
//	ethValidatorAddrKeyA := "3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
//	ethValidatorAddrKeyB := "a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
//	ethValidatorAddrKeyC := "bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
//	ethValidatorAddrKeyD := "c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"
//
//	ethValidatorAddrKeys := make([]string, 0)
//	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyA)
//	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyB)
//	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyC)
//	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyD)
//
//	ctx := context.Background()
//	//var backend bind.ContractBackend
//	backend, para := PrepareTestEnvironment(deployerPrivateKey, ethValidatorAddrKeys)
//	sim := new(ethinterface.SimExtend)
//	sim.SimulatedBackend = backend.(*backends.SimulatedBackend)
//
//	callMsg := ethereum.CallMsg{
//		From: para.Deployer,
//		Data: common.FromHex(generated.BridgeBankBin),
//	}
//
//	_, err := sim.EstimateGas(ctx, callMsg)
//	if nil != err {
//		panic("failed to estimate gas due to:" + err.Error())
//	}
//	x2EthContracts, x2EthDeployInfo, err := DeployAndInit(sim, para)
//	if nil != err {
//		return nil, nil, nil, nil, err
//	}
//	sim.Commit()
//
//	return para, sim, x2EthContracts, x2EthDeployInfo, nil
//}
//
//func PrepareTestEnvironment(deployerPrivateKey string, ethValidatorAddrKeys []string) (bind.ContractBackend, *DeployPara) {
//	genesiskey, _ := crypto.HexToECDSA(deployerPrivateKey)
//	alloc := make(core.GenesisAlloc)
//	genesisAddr := crypto.PubkeyToAddress(genesiskey.PublicKey)
//	genesisAccount := core.GenesisAccount{
//		Balance:    big.NewInt(10000000000 * 10000),
//		PrivateKey: crypto.FromECDSA(genesiskey),
//	}
//	alloc[genesisAddr] = genesisAccount
//
//	var InitValidators []common.Address
//	var ValidatorPriKey []*ecdsa.PrivateKey
//	for _, v := range ethValidatorAddrKeys {
//		key, _ := crypto.HexToECDSA(v)
//		addr := crypto.PubkeyToAddress(key.PublicKey)
//		InitValidators = append(InitValidators, addr)
//		ValidatorPriKey = append(ValidatorPriKey, key)
//
//		account := core.GenesisAccount{
//			Balance:    big.NewInt(100000000 * 100),
//			PrivateKey: crypto.FromECDSA(key),
//		}
//		alloc[addr] = account
//	}
//
//	gasLimit := uint64(100000000)
//	sim := backends.NewSimulatedBackend(alloc, gasLimit)
//
//	InitPowers := []*big.Int{big.NewInt(80), big.NewInt(10), big.NewInt(10), big.NewInt(10)}
//
//	para := &DeployPara{
//		DeployPrivateKey: genesiskey,
//		Deployer:         genesisAddr,
//		Operator:         genesisAddr,
//		InitValidators:   InitValidators,
//		ValidatorPriKey:  ValidatorPriKey,
//		InitPowers:       InitPowers,
//	}
//
//	return sim, para
//}
