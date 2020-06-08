package ethtxs

import (
	"math/big"
	"testing"

	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethinterface"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
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

func Test_isWebsocketURL(t *testing.T) {
	bret := isWebsocketURL("ws://127.0.0.1:7545/")
	assert.Equal(t, bret, true)

	bret = isWebsocketURL("https://127.0.0.1:7545/")
	assert.Equal(t, bret, false)
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
		Balance:    big.NewInt(10000000000 * 10000),
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
