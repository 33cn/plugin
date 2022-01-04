package ethtxs

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	chain33Address "github.com/33cn/chain33/common/address"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

//NewProphecyClaimPara ...
type NewProphecyClaimPara struct {
	ClaimType     uint8
	Chain33Sender []byte
	TokenAddr     common.Address
	EthReceiver   common.Address
	Symbol        string
	Amount        *big.Int
	Txhash        []byte
}

//Burn ...
func Burn(ownerPrivateKeyStr, tokenAddrstr, chain33Receiver string, bridgeBank common.Address, amount *big.Int,
	bridgeBankIns *generated.BridgeBank, client ethinterface.EthClientSpec, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	ownerPrivateKey, err := crypto.ToECDSA(common.FromHex(ownerPrivateKeyStr))
	if nil != err {
		return "", err
	}
	ownerAddr := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)
	var prepareDone bool

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(ownerAddr)
		}
	}()
	auth, err := PrepareAuth4MultiEthereum(client, ownerPrivateKey, ownerAddr, addr2TxNonce)
	if nil != err {
		return "", err
	}

	prepareDone = true

	tokenAddr := common.HexToAddress(tokenAddrstr)
	tokenInstance, err := generated.NewBridgeToken(tokenAddr, client)
	if nil != err {
		return "", err
	}
	//chain33bank 是bridgeBank的基类，所以使用bridgeBank的地址
	tx, err := tokenInstance.Approve(auth, bridgeBank, amount)
	if nil != err {
		return "", err
	}

	err = waitEthTxFinished(client, tx.Hash(), "Approve")
	if nil != err {
		return "", err
	}
	txslog.Info("Burn", "Approve tx with hash", tx.Hash().String())

	prepareDone = false

	auth, err = PrepareAuth4MultiEthereum(client, ownerPrivateKey, ownerAddr, addr2TxNonce)
	if nil != err {
		return "", err
	}

	prepareDone = true

	receAddr, err := chain33Address.NewAddrFromString(chain33Receiver)
	if nil != err {
		txslog.Info("Burn", "Failed to decode chain33 address due to", err.Error())
		return "", err
	}

	tx, err = bridgeBankIns.BurnBridgeTokens(auth, receAddr.Hash160[:], tokenAddr, amount)
	if nil != err {
		return "", err
	}
	err = waitEthTxFinished(client, tx.Hash(), "Burn")
	if nil != err {
		return "", err
	}

	return tx.Hash().String(), nil
}

//BurnAsync ...
func BurnAsync(ownerPrivateKeyStr, tokenAddrstr, chain33Receiver string, amount *big.Int, bridgeBankIns *generated.BridgeBank,
	client ethinterface.EthClientSpec, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	ownerPrivateKey, err := crypto.ToECDSA(common.FromHex(ownerPrivateKeyStr))
	if nil != err {
		return "", err
	}
	ownerAddr := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)

	var prepareDone bool

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(ownerAddr)
		}
	}()
	auth, err := PrepareAuth4MultiEthereum(client, ownerPrivateKey, ownerAddr, addr2TxNonce)
	if nil != err {
		return "", err
	}

	prepareDone = true

	tokenAddr := common.HexToAddress(tokenAddrstr)
	receAddr, err := chain33Address.NewAddrFromString(chain33Receiver)
	if nil != err {
		txslog.Info("BurnAsync", "Failed to decode chain33 address due to", err.Error())
		return "", err
	}

	tx, err := bridgeBankIns.BurnBridgeTokens(auth, receAddr.Hash160[:], tokenAddr, amount)
	if nil != err {
		return "", err
	}

	return tx.Hash().String(), nil
}

//TransferToken ...
func TransferToken(tokenAddr, fromPrivateKeyStr, toAddr string, amount *big.Int, client ethinterface.EthClientSpec, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	tokenInstance, err := generated.NewBridgeToken(common.HexToAddress(tokenAddr), client)
	if nil != err {
		return "", err
	}

	var prepareDone bool

	fromPrivateKey, err := crypto.ToECDSA(common.FromHex(fromPrivateKeyStr))
	if nil != err {
		return "", err
	}
	fromAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(fromAddr)
		}
	}()

	auth, err := PrepareAuth4MultiEthereum(client, fromPrivateKey, fromAddr, addr2TxNonce)
	if nil != err {
		return "", err
	}

	prepareDone = true

	tx, err := tokenInstance.Transfer(auth, common.HexToAddress(toAddr), amount)
	if nil != err {
		return "", err
	}

	err = waitEthTxFinished(client, tx.Hash(), "TransferFromToken")
	if nil != err {
		return "", err
	}
	return tx.Hash().String(), nil
}

func TransferEth(fromPrivateKeyStr, toAddr string, amount *big.Int, client ethinterface.EthClientSpec, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	txslog.Info("TransferEth", "toAddr", toAddr, "amount", amount)
	var prepareDone bool

	fromPrivateKey, err := crypto.ToECDSA(common.FromHex(fromPrivateKeyStr))
	if nil != err {
		return "", err
	}
	fromAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(fromAddr)
		}
	}()

	auth, err := PrepareAuth4MultiEthereum(client, fromPrivateKey, fromAddr, addr2TxNonce)
	if nil != err {
		return "", err
	}

	prepareDone = true

	gasLimit := uint64(21100)

	toAddress := common.HexToAddress(toAddr)
	//var data []byte

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    uint64(auth.Nonce.Int64()),
		To:       &toAddress,
		Value:    amount,
		Gas:      gasLimit, //auth.GasLimit,
		GasPrice: auth.GasPrice,
		//Data:     data,
	})

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		txslog.Error("TransferEth NetworkID", "tx", tx)
		return "", err
	}

	txslog.Info("TransferEth", "chainID", chainID, "amount", amount, "tx", tx)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivateKey)
	if err != nil {
		txslog.Error("TransferEth SignTx", "err", err)
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		txslog.Error("TransferEth SendTransaction", "err", err)
		return "", err
	}

	err = waitEthTxFinished(client, signedTx.Hash(), "TransferEth")
	if nil != err {
		return "", err
	}
	return signedTx.Hash().String(), nil
}

//LockEthErc20Asset ...
func LockEthErc20Asset(ownerPrivateKeyStr, tokenAddrStr, chain33Receiver string, amount *big.Int, client ethinterface.EthClientSpec, bridgeBank *generated.BridgeBank, bridgeBankAddr common.Address, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	var prepareDone bool
	txslog.Info("LockEthErc20Asset", "ownerPrivateKeyStr", ownerPrivateKeyStr, "tokenAddrStr", tokenAddrStr, "chain33Receiver", chain33Receiver, "amount", amount.String())
	ownerPrivateKey, err := crypto.ToECDSA(common.FromHex(ownerPrivateKeyStr))
	if nil != err {
		return "", err
	}
	ownerAddr := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(ownerAddr)
		}
	}()

	//ETH转账，空地址，且设置value
	var tokenAddr common.Address
	if "" != tokenAddrStr {
		//如果是eth以外的erc20，则需要先进行approve操作
		tokenAddr = common.HexToAddress(tokenAddrStr)
		tokenInstance, err := generated.NewBridgeToken(tokenAddr, client)
		if nil != err {
			return "", err
		}
		auth, err := PrepareAuth4MultiEthereum(client, ownerPrivateKey, ownerAddr, addr2TxNonce)
		if nil != err {
			txslog.Error("LockEthErc20Asset", "PrepareAuth err", err.Error())
			return "", err
		}

		prepareDone = true

		//chain33bank 是bridgeBank的基类，所以使用bridgeBank的地址
		tx, err := tokenInstance.Approve(auth, bridgeBankAddr, amount)
		if nil != err {
			return "", err
		}
		err = waitEthTxFinished(client, tx.Hash(), "Approve")
		if nil != err {
			return "", err
		}
		txslog.Info("LockEthErc20Asset", "Approve tx with hash", tx.Hash().String())
	}

	prepareDone = false

	auth, err := PrepareAuth4MultiEthereum(client, ownerPrivateKey, ownerAddr, addr2TxNonce)
	if nil != err {
		txslog.Error("LockEthErc20Asset", "PrepareAuth err", err.Error())
		return "", err
	}

	prepareDone = true

	if "" == tokenAddrStr {
		auth.Value = amount
	}

	recvAddr, err := chain33Address.NewAddrFromString(chain33Receiver)
	if nil != err {
		txslog.Info("LockEthErc20Asset", "Failed to decode chain33 address due to", err.Error())
		return "", err
	}

	tx, err := bridgeBank.Lock(auth, recvAddr.Hash160[:], tokenAddr, amount)
	if nil != err {
		txslog.Error("LockEthErc20Asset", "lock err", err.Error())
		return "", err
	}
	err = waitEthTxFinished(client, tx.Hash(), "LockEthErc20Asset")
	if nil != err {
		txslog.Error("LockEthErc20Asset", "waitEthTxFinished err", err.Error())
		return "", err
	}

	return tx.Hash().String(), nil
}

//LockEthErc20AssetAsync ...
func LockEthErc20AssetAsync(ownerPrivateKeyStr, tokenAddrStr, chain33Receiver string, amount *big.Int, client ethinterface.EthClientSpec, bridgeBank *generated.BridgeBank, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	txslog.Info("LockEthErc20AssetAsync", "ownerPrivateKeyStr", ownerPrivateKeyStr, "tokenAddrStr", tokenAddrStr, "chain33Receiver", chain33Receiver, "amount", amount.String())
	ownerPrivateKey, err := crypto.ToECDSA(common.FromHex(ownerPrivateKeyStr))
	if nil != err {
		return "", err
	}
	ownerAddr := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)

	var prepareDone bool

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(ownerAddr)
		}
	}()

	auth, err := PrepareAuth4MultiEthereum(client, ownerPrivateKey, ownerAddr, addr2TxNonce)
	if nil != err {
		txslog.Error("LockEthErc20AssetAsync", "PrepareAuth err", err.Error())
		return "", err
	}
	prepareDone = true

	//ETH转账，空地址，且设置value
	var tokenAddr common.Address
	if "" == tokenAddrStr {
		auth.Value = amount
	}

	if "" != tokenAddrStr {
		tokenAddr = common.HexToAddress(tokenAddrStr)
	}
	recvAddr, err := chain33Address.NewAddrFromString(chain33Receiver)
	if nil != err {
		txslog.Info("LockEthErc20AssetAsync", "Failed to decode chain33 address due to", err.Error())
		return "", err
	}

	tx, err := bridgeBank.Lock(auth, recvAddr.Hash160[:], tokenAddr, amount)
	if nil != err {
		txslog.Error("LockEthErc20AssetAsync", "lock err", err.Error())
		_, err = revokeNonce(ownerAddr)
		if err != nil {
			return "", err
		}
		return "", err
	}
	return tx.Hash().String(), nil
}

func SetupMultiSign(ownerPrivateKeyStr, multiSignAddrstr string, owners []string, client ethinterface.EthClientSpec, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	txslog.Info("SetupMultiSign", "owners", owners, "multiSignAddrstr", multiSignAddrstr)

	ownerPrivateKey, err := crypto.ToECDSA(common.FromHex(ownerPrivateKeyStr))
	if nil != err {
		return "", err
	}
	ownerAddr := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)

	var prepareDone bool

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(ownerAddr)
		}
	}()

	auth, err := PrepareAuth4MultiEthereum(client, ownerPrivateKey, ownerAddr, addr2TxNonce)
	if nil != err {
		txslog.Error("SetupMultiSign", "PrepareAuth err", err.Error())
		return "", err
	}
	prepareDone = true

	gnosisSafeAddr := common.HexToAddress(multiSignAddrstr)
	gnosisSafeInt, err := gnosis.NewGnosisSafe(gnosisSafeAddr, client)
	if nil != err {
		return "", err
	}

	var _owners []common.Address
	for _, onwer := range owners {
		_owners = append(_owners, common.HexToAddress(onwer))
	}
	AddressZero := common.HexToAddress(ebTypes.EthNilAddr)

	//safe.setup([user1.address, user2.address], 1, AddressZero, "0x", handler.address, AddressZero, 0, AddressZero)
	setupTx, err := gnosisSafeInt.Setup(auth, _owners, big.NewInt(int64(len(_owners))), AddressZero, []byte{'0', 'x'}, AddressZero, AddressZero, big.NewInt(int64(0)), AddressZero)
	if nil != err {
		txslog.Error("SetupMultiSign", "Failed to setupTx with err:", err.Error())
		return "", err
	}

	txslog.Info("SetupMultiSign", "SetupMultiSign tx hash:", setupTx.Hash().String())
	err = waitEthTxFinished(client, setupTx.Hash(), "SetupMultiSign")
	if nil != err {
		return "", err
	}

	return setupTx.Hash().String(), nil
}

func SafeTransfer(ownerPrivateKeyStr, multiSignAddrstr, receiver, token string, privateKeys []string, amount float64, client ethinterface.EthClientSpec, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	txslog.Info("SafeTransfer", "multiSignAddrstr", multiSignAddrstr, "receiver", receiver, "token", token, "amount", amount)

	ownerPrivateKey, err := crypto.ToECDSA(common.FromHex(ownerPrivateKeyStr))
	if nil != err {
		return "", err
	}
	ownerAddr := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)

	var prepareDone bool
	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(ownerAddr)
		}
	}()

	auth, err := PrepareAuth4MultiEthereum(client, ownerPrivateKey, ownerAddr, addr2TxNonce)
	if nil != err {
		txslog.Error("SafeTransfer", "PrepareAuth err", err.Error())
		return "", err
	}
	prepareDone = true
	auth.GasLimit = GasLimit

	gnosisSafeAddr := common.HexToAddress(multiSignAddrstr)
	gnosisSafeInt, err := gnosis.NewGnosisSafe(gnosisSafeAddr, client)
	if nil != err {
		return "", err
	}
	AddressZero := common.HexToAddress(ebTypes.EthNilAddr)

	_to := common.HexToAddress(receiver)
	_data := []byte{'0', 'x'}
	safeTxGas := big.NewInt(10 * 10000)
	baseGas := big.NewInt(0)
	gasPrice := big.NewInt(0)
	value := big.NewInt(0)
	opts := &bind.CallOpts{
		From:    ownerAddr,
		Context: context.Background(),
	}
	//Eth transfer
	if token == "" {
		value.Mul(big.NewInt(int64(amount)), big.NewInt(int64(1e18)))
	} else {
		_to = common.HexToAddress(token)

		erc20Abi, err := abi.JSON(strings.NewReader(erc20.ERC20ABI))
		if err != nil {
			return "", err
		}

		tokenInstance, err := erc20.NewERC20(_to, client)
		if err != nil {
			return "", err
		}
		decimals, err := tokenInstance.Decimals(opts)
		if err != nil {
			return "", err
		}

		dec, ok := ebTypes.DecimalsPrefix[decimals]
		if !ok {
			txslog.Error("SafeTransfer", "not support the decimals =", decimals)
			return "", errors.New("not support the decimals")
		}
		value.Mul(big.NewInt(int64(amount)), big.NewInt(dec))
		//value = utils.ToWei(amount, int64(decimals))

		_data, err = erc20Abi.Pack("transfer", common.HexToAddress(receiver), value)
		if err != nil {
			return "", err
		}
		//对于erc20这种方式 最后需要将其设置为0
		value = big.NewInt(0)
	}

	nonce, err := gnosisSafeInt.Nonce(opts)
	if err != nil {
		txslog.Error("SafeTransfer", "Failed to get Nonce", err.Error())
		return "", err
	}

	signContent, err := gnosisSafeInt.GetTransactionHash(opts, _to, value, _data, 0,
		safeTxGas, baseGas, gasPrice, AddressZero, AddressZero, nonce)
	if err != nil {
		txslog.Error("SafeTransfer", "Failed to GetTransactionHash", err.Error())
		return "", err
	}

	sigs, err := buildSigs(signContent[:], privateKeys)
	if err != nil {
		txslog.Error("SafeTransfer", "Failed to buildSigs", err.Error())
		return "", err
	}

	txslog.Info("SafeTransfer", "value str", value.String(), "value int64", value.Int64())
	execTx, err := gnosisSafeInt.ExecTransaction(auth, _to, value, _data, 0,
		safeTxGas, baseGas, gasPrice, AddressZero, AddressZero, sigs)
	if nil != err {
		txslog.Error("SafeTransfer", "Failed to ExecTransaction", err.Error())
		return "", err
	}

	txslog.Info("SetupMultiSign", "SetupMultiSign tx hash:", execTx.Hash().String())
	err = waitEthTxFinished(client, execTx.Hash(), "SetupMultiSign")
	if nil != err {
		return "", err
	}

	return execTx.Hash().String(), nil
}

func buildSigs(data []byte, privateKeys []string) ([]byte, error) {
	txslog.Info("buildSigs", "data:", common.Bytes2Hex(data))

	var sigs []byte
	for _, privateKeyStr := range privateKeys {
		privateKey, err := crypto.ToECDSA(common.FromHex(privateKeyStr))
		if nil != err {
			return nil, errors.New("failed to recover private key")
		}

		signature, err := crypto.Sign(data, privateKey)
		if err != nil {
			txslog.Error("buildSigs", "Failed to sign data due to:"+err.Error())
			return nil, err
		}
		signature[64] += 27
		sigs = append(sigs, signature[:]...)
	}

	return sigs, nil
}

func ConfigOfflineSaveAccount(addr string, client ethinterface.EthClientSpec, para *OperatorInfo, x2EthContracts *X2EthContracts, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	txslog.Info("ConfigOfflineSaveAccount", "addr", addr)
	if nil == para {
		return "", errors.New("no operator private key configured")
	}

	var prepareDone bool
	var err error

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(para.Address)
		}
	}()

	auth, err := PrepareAuth4MultiEthereum(client, para.PrivateKey, para.Address, addr2TxNonce)
	if nil != err {
		return "", err
	}

	prepareDone = true

	OfflineAddr := common.HexToAddress(addr)
	tx, err := x2EthContracts.BridgeBank.BridgeBankTransactor.ConfigOfflineSaveAccount(auth, OfflineAddr)
	if nil != err {
		return "", err
	}

	sim, isSim := client.(*ethinterface.SimExtend)
	if isSim {
		fmt.Println("Use the simulator")
		sim.Commit()
	}

	txslog.Info("ConfigOfflineSaveAccount", "tx.Hash()", tx.Hash().String())
	err = waitEthTxFinished(client, tx.Hash(), "ConfigOfflineSaveAccount")
	if nil != err {
		return "", err
	}

	return tx.Hash().String(), nil
}

func ConfigplatformTokenSymbol(symbol string, client ethinterface.EthClientSpec, para *OperatorInfo, x2EthContracts *X2EthContracts, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	txslog.Info("ConfigplatformTokenSymbol", "symbol", symbol)
	if nil == para {
		return "", errors.New("no operator private key configured")
	}

	var prepareDone bool
	var err error

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(para.Address)
		}
	}()

	auth, err := PrepareAuth4MultiEthereum(client, para.PrivateKey, para.Address, addr2TxNonce)
	if nil != err {
		return "", err
	}

	prepareDone = true

	tx, err := x2EthContracts.BridgeBank.BridgeBankTransactor.ConfigplatformTokenSymbol(auth, symbol)
	if nil != err {
		return "", err
	}

	sim, isSim := client.(*ethinterface.SimExtend)
	if isSim {
		fmt.Println("Use the simulator")
		sim.Commit()
	}

	txslog.Info("ConfigplatformTokenSymbol", "tx.Hash()", tx.Hash().String())
	err = waitEthTxFinished(client, tx.Hash(), "ConfigplatformTokenSymbol")
	if nil != err {
		return "", err
	}

	return tx.Hash().String(), nil
}

func ConfigLockedTokenOfflineSave(addr, symbol string, threshold *big.Int, percents uint8, client ethinterface.EthClientSpec, para *OperatorInfo, x2EthContracts *X2EthContracts, addr2TxNonce map[common.Address]*NonceMutex) (string, error) {
	txslog.Info("ConfigLockedTokenOfflineSave", "addr", addr, "symbol", symbol, "threshold", threshold, "percents", percents)
	if nil == para {
		return "", errors.New("no operator private key configured")
	}

	var prepareDone bool
	var err error

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(para.Address)
		}
	}()

	auth, err := PrepareAuth4MultiEthereum(client, para.PrivateKey, para.Address, addr2TxNonce)
	if nil != err {
		return "", err
	}

	prepareDone = true

	if addr == "" || symbol == "ETH" {
		addr = ebTypes.EthNilAddr
	}
	ToeknAddr := common.HexToAddress(addr)
	tx, err := x2EthContracts.BridgeBank.BridgeBankTransactor.ConfigLockedTokenOfflineSave(auth, ToeknAddr, symbol, threshold, percents)
	if nil != err {
		return "", err
	}

	sim, isSim := client.(*ethinterface.SimExtend)
	if isSim {
		fmt.Println("Use the simulator")
		sim.Commit()
	}

	txslog.Info("ConfigLockedTokenOfflineSave", "tx.Hash()", tx.Hash().String())
	err = waitEthTxFinished(client, tx.Hash(), "ConfigLockedTokenOfflineSave")
	if nil != err {
		return "", err
	}

	return tx.Hash().String(), nil
}
