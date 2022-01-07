package ethtxs

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	chain33Addr = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
)

type suiteContracts struct {
	suite.Suite
	para            *DeployPara
	sim             *ethinterface.SimExtend
	x2EthContracts  *X2EthContracts
	x2EthDeployInfo *X2EthDeployInfo
}

func TestRunSuiteX2Ethereum(t *testing.T) {
	log := new(suiteContracts)
	suite.Run(t, log)
}

func (c *suiteContracts) SetupSuite() {
	var err error
	c.para, c.sim, c.x2EthContracts, c.x2EthDeployInfo, err = DeployContracts()
	require.Nil(c.T(), err)
}

func (c *suiteContracts) Test_GetOperator() {
	operator, err := GetOperator(c.sim, c.para.InitValidators[0], c.x2EthDeployInfo.BridgeBank.Address)
	require.Nil(c.T(), err)
	assert.Equal(c.T(), operator.String(), c.para.Operator.String())
}

func (c *suiteContracts) Test_IsActiveValidator() {
	bret, err := IsActiveValidator(c.para.InitValidators[0], c.x2EthContracts.Valset)
	require.Nil(c.T(), err)
	assert.Equal(c.T(), bret, true)

	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	bret, err = IsActiveValidator(addr, c.x2EthContracts.Valset)
	require.Nil(c.T(), err) // ???
	assert.Equal(c.T(), bret, false)
}

func (c *suiteContracts) Test_IsProphecyPending() {
	claimID := crypto.Keccak256Hash(big.NewInt(50).Bytes())
	bret, err := IsProphecyPending(claimID, c.para.InitValidators[0], c.x2EthContracts.Chain33Bridge)
	require.Nil(c.T(), err)
	assert.Equal(c.T(), bret, false)
}

func (c *suiteContracts) Test_LogLockToEthBridgeClaim() {
	to := common.FromHex(chain33Addr)
	event := &events.LockEvent{
		From:   c.para.InitValidators[0],
		To:     to,
		Token:  common.HexToAddress(EthNullAddr),
		Symbol: "eth",
		Value:  big.NewInt(10000 * 10000 * 10000),
		Nonce:  big.NewInt(1),
	}
	witnessClaim, err := LogLockToEthBridgeClaim(event, 1, c.x2EthDeployInfo.BridgeBank.Address.String(), "", 18)
	require.Nil(c.T(), err)
	assert.NotEmpty(c.T(), witnessClaim)
	assert.Equal(c.T(), witnessClaim.EthereumChainID, int64(1))
	assert.Equal(c.T(), witnessClaim.BridgeBrankAddr, c.x2EthDeployInfo.BridgeBank.Address.String())
	assert.Equal(c.T(), witnessClaim.TokenAddr, EthNullAddr)
	assert.Equal(c.T(), witnessClaim.Symbol, event.Symbol)
	assert.Equal(c.T(), witnessClaim.EthereumSender, event.From.String())
	//assert.Equal(c.T(), witnessClaim.Chain33Receiver, string(event.To))
	assert.Equal(c.T(), witnessClaim.Amount, "1000000000000")
	assert.Equal(c.T(), witnessClaim.Nonce, event.Nonce.Int64())
	assert.Equal(c.T(), witnessClaim.Decimal, int64(18))

	event.To = nil
	_, err = LogLockToEthBridgeClaim(event, 1, c.x2EthDeployInfo.BridgeBank.Address.String(), "", 18)
	require.NotNil(c.T(), err)
	assert.Equal(c.T(), err, ebrelayerTypes.ErrEmptyAddress)
}

func (c *suiteContracts) Test_LogBurnToEthBridgeClaim() {
	to := common.FromHex(chain33Addr)
	event := &events.BurnEvent{
		OwnerFrom:       c.para.InitValidators[0],
		Chain33Receiver: to,
		Token:           common.HexToAddress(EthNullAddr),
		Symbol:          "bty",
		Amount:          big.NewInt(100),
		Nonce:           big.NewInt(2),
	}
	witnessClaim, err := LogBurnToEthBridgeClaim(event, 1, c.x2EthDeployInfo.BridgeBank.Address.String(), "", 8)
	require.Nil(c.T(), err)
	assert.NotEmpty(c.T(), witnessClaim)
	assert.Equal(c.T(), witnessClaim.EthereumChainID, int64(1))
	assert.Equal(c.T(), witnessClaim.BridgeBrankAddr, c.x2EthDeployInfo.BridgeBank.Address.String())
	assert.Equal(c.T(), witnessClaim.TokenAddr, EthNullAddr)
	assert.Equal(c.T(), witnessClaim.Symbol, event.Symbol)
	assert.Equal(c.T(), witnessClaim.EthereumSender, event.OwnerFrom.String())
	//assert.Equal(c.T(), witnessClaim.Chain33Receiver, string(event.Chain33Receiver))
	assert.Equal(c.T(), witnessClaim.Amount, "100")
	assert.Equal(c.T(), witnessClaim.Nonce, event.Nonce.Int64())
	assert.Equal(c.T(), witnessClaim.Decimal, int64(8))
}

func (c *suiteContracts) Test_RecoverContractHandler() {
	_, _, err := RecoverContractHandler(c.sim, c.x2EthDeployInfo.BridgeRegistry.Address, c.x2EthDeployInfo.BridgeRegistry.Address)
	require.Nil(c.T(), err)
}

func (c *suiteContracts) Test_RecoverOracleInstance() {
	oracleInstance, err := RecoverOracleInstance(c.sim, c.x2EthDeployInfo.BridgeRegistry.Address, c.x2EthDeployInfo.BridgeRegistry.Address)
	require.Nil(c.T(), err)
	require.NotNil(c.T(), oracleInstance)
}

func (c *suiteContracts) Test_GetDeployHeight() {
	height, err := GetDeployHeight(c.sim, c.x2EthDeployInfo.BridgeRegistry.Address, c.x2EthDeployInfo.BridgeRegistry.Address)
	require.Nil(c.T(), err)
	assert.True(c.T(), height > 0)
}

func (c *suiteContracts) Test_GetLockedFunds() {
	balance, err := GetLockedFunds(c.x2EthContracts.BridgeBank, "")
	require.Nil(c.T(), err)
	assert.Equal(c.T(), balance, "0")
}

func PrepareTestEnv() (*ethinterface.SimExtend, *DeployPara) {
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

func DeployContracts() (*DeployPara, *ethinterface.SimExtend, *X2EthContracts, *X2EthDeployInfo, error) {
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
