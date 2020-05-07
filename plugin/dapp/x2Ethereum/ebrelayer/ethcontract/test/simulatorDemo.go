package test

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/ethcontract/generated"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// Generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()

	alloc := make(core.GenesisAlloc)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	genesisAccount := core.GenesisAccount{
		Balance:    big.NewInt(10000000000),
		PrivateKey: crypto.FromECDSA(key),
	}
	alloc[addr] = genesisAccount

	gasLimit := uint64(100000000)
	sim := backends.NewSimulatedBackend(alloc, gasLimit)
	ctx := context.Background()
	gasPrice, err := sim.SuggestGasPrice(ctx)
	if err != nil {
		panic("Failed to SuggestGasPrice due to:" + err.Error())
	}

	auth := bind.NewKeyedTransactor(key)
	//auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	// Deploy a token contract on the simulated blockchain
	cosmosBridgeAddr := common.HexToAddress("0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a")
	bridgeBankAddr := common.HexToAddress("0x0df9a824699bc5878232c9e612fe1a5346a5a368")
	oracleAddr := common.HexToAddress("0x92c8b16afd6d423652559c6e266cbe1c29bfd84f")
	valsetAddr := common.HexToAddress("0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1")

	_, _, bridgeRegistry, err := generated.DeployBridgeRegistry(auth, sim, cosmosBridgeAddr, bridgeBankAddr, oracleAddr, valsetAddr)
	if err != nil {
		log.Fatalf("Failed to deploy BridgeRegistry contract: %v", err)
	}
	balance, _ := sim.BalanceAt(ctx, addr, nil)
	fmt.Println("balance before:", balance.String())

	// Print the current (non existent) and pending name of the contract
	cosmosBridgeAddrRetrive, _ := bridgeRegistry.Chain33Bridge(nil)
	fmt.Println("Pre-mining cosmosBridgeAddr:", cosmosBridgeAddrRetrive.String())

	cosmosBridgeAddrRetrive, _ = bridgeRegistry.Chain33Bridge(&bind.CallOpts{Pending: true})
	fmt.Println("Pre-mining pending cosmosBridgeAddr:", cosmosBridgeAddrRetrive.String())

	balance, _ = sim.BalanceAt(ctx, addr, nil)
	fmt.Println("balance after:", balance.String())
	// Commit all pending transactions in the simulator and print the names again
	sim.Commit()

	balance, _ = sim.BalanceAt(ctx, addr, nil)
	fmt.Println("balance after:", balance.String())

	cosmosBridgeAddrRetrive, _ = bridgeRegistry.Chain33Bridge(nil)
	fmt.Println("Post-mining cosmosBridgeAddr:", cosmosBridgeAddrRetrive.String())

	cosmosBridgeAddrRetrive, _ = bridgeRegistry.Chain33Bridge(&bind.CallOpts{Pending: true})
	fmt.Println("Post-mining pending cosmosBridgeAddr:", cosmosBridgeAddrRetrive.String())

	cosmosBridgeAddrRetrive, _ = bridgeRegistry.Chain33Bridge(nil)
	fmt.Println("Post-mining cosmosBridgeAddr:", cosmosBridgeAddrRetrive.String())

	balance, _ = sim.BalanceAt(ctx, addr, nil)
	fmt.Println("balance after:", balance.String())
}
