package offline

import (
	"crypto/ecdsa"
	"math/big"
	"strings"

	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/cakeToken"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/masterChef"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/syrupBar"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type SignCakeToken struct {
}

func (s *SignCakeToken) reWriteDeployCakeToken(nonce uint64, gasPrice *big.Int, key *ecdsa.PrivateKey, params ...interface{}) (signedTx, hash string, err error) {
	parsed, err := abi.JSON(strings.NewReader(cakeToken.CakeTokenABI))
	if err != nil {
		return
	}
	input, err := parsed.Pack("", params...)
	if err != nil {
		return
	}
	abiBin := cakeToken.CakeTokenBin
	data := append(common.FromHex(abiBin), input...)

	var amount = new(big.Int)
	ntx := types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data)
	return SignTx(key, ntx)

}

type signsyrupBar struct {
}

func (s *signsyrupBar) reWriteDeploysyrupBar(nonce uint64, gasPrice *big.Int, key *ecdsa.PrivateKey, cakeAddress common.Address) (signedTx, hash string, err error) {
	parsed, err := abi.JSON(strings.NewReader(syrupBar.SyrupBarABI))
	if err != nil {
		return
	}
	input, err := parsed.Pack("", cakeAddress)
	if err != nil {
		return
	}
	abiBin := syrupBar.SyrupBarBin
	data := append(common.FromHex(abiBin), input...)

	var amount = new(big.Int)
	ntx := types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data)
	return SignTx(key, ntx)
}

type signMasterChef struct {
}

func (s *signMasterChef) reWriteDeployMasterChef(nonce uint64, gasPrice *big.Int, key *ecdsa.PrivateKey, cakeAddress, syruBarAddress, fromaddr common.Address, reward, _startBlock *big.Int) (signedTx, hash string, err error) {
	parsed, err := abi.JSON(strings.NewReader(masterChef.MasterChefABI))
	if err != nil {
		return
	}
	input, err := parsed.Pack("", cakeAddress, syruBarAddress, fromaddr, reward, _startBlock)
	if err != nil {
		return
	}
	abiBin := masterChef.MasterChefBin
	data := append(common.FromHex(abiBin), input...)

	var amount = new(big.Int)
	ntx := types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data)
	return SignTx(key, ntx)
}
