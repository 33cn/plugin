package ethtxs

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
	"github.com/ethereum/go-ethereum"
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

//CreateBridgeToken ...
func CreateBridgeToken(symbol string, client ethinterface.EthClientSpec, para *OperatorInfo, x2EthDeployInfo *X2EthDeployInfo, x2EthContracts *X2EthContracts) (string, error) {
	if nil == para {
		return "", errors.New("no operator private key configured")
	}
	//订阅事件
	eventName := "LogNewBridgeToken"
	bridgeBankABI := LoadABI(BridgeBankABI)
	logNewBridgeTokenSig := bridgeBankABI.Events[eventName].ID.Hex()
	query := ethereum.FilterQuery{
		Addresses: []common.Address{x2EthDeployInfo.BridgeBank.Address},
	}
	// We will check logs for new events
	logs := make(chan types.Log)
	// Filter by contract and event, write results to logs
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if nil != err {
		txslog.Error("CreateBrigeToken", "failed to SubscribeFilterLogs", err.Error())
		return "", err
	}

	var prepareDone bool

	defer func() {
		if err != nil && prepareDone {
			_, _ = revokeNonce(para.Address)
		}
	}()

	//创建token
	auth, err := PrepareAuth(client, para.PrivateKey, para.Address)
	if nil != err {
		return "", err
	}

	prepareDone = true

	tx, err := x2EthContracts.BridgeBank.BridgeBankTransactor.CreateNewBridgeToken(auth, symbol)
	if nil != err {
		return "", err
	}

	sim, isSim := client.(*ethinterface.SimExtend)
	if isSim {
		fmt.Println("Use the simulator")
		sim.Commit()
	}

	err = waitEthTxFinished(client, tx.Hash(), "CreateBridgeToken")
	if nil != err {
		return "", err
	}

	logEvent := &events.LogNewBridgeToken{}
	select {
	// Handle any errors
	case err := <-sub.Err():
		return "", err
	// vLog is raw event data
	case vLog := <-logs:
		// Check if the event is a 'LogLock' event
		if vLog.Topics[0].Hex() == logNewBridgeTokenSig {
			txslog.Debug("CreateBrigeToken", "Witnessed new event", eventName, "Block number", vLog.BlockNumber)

			err = bridgeBankABI.UnpackIntoInterface(logEvent, eventName, vLog.Data)
			if nil != err {
				return "", err
			}
			if symbol != logEvent.Symbol {
				txslog.Error("CreateBrigeToken", "symbol", symbol, "logEvent.Symbol", logEvent.Symbol)
			}
			txslog.Info("CreateBrigeToken", "Witnessed new event", eventName, "Block number", vLog.BlockNumber, "token address", logEvent.Token.String())
			break
		}
	}
	return logEvent.Token.String(), nil
}

//CreateERC20Token ...
func CreateERC20Token(symbol string, client ethinterface.EthClientSpec, para *OperatorInfo) (string, error) {
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

	auth, err := PrepareAuth(client, para.PrivateKey, para.Address)
	if nil != err {
		return "", err
	}

	prepareDone = true

	tokenAddr, tx, _, err := generated.DeployBridgeToken(auth, client, symbol)
	if nil != err {
		return "", err
	}

	err = waitEthTxFinished(client, tx.Hash(), "CreateERC20Token")
	if nil != err {
		return "", err
	}

	return tokenAddr.String(), nil
}

//MintERC20Token ...
func MintERC20Token(tokenAddr, ownerAddr string, amount *big.Int, client ethinterface.EthClientSpec, para *OperatorInfo) (string, error) {
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

	operatorAuth, err := PrepareAuth(client, para.PrivateKey, para.Address)
	if nil != err {
		return "", err
	}

	prepareDone = true

	erc20TokenInstance, err := generated.NewBridgeToken(common.HexToAddress(tokenAddr), client)
	if nil != err {
		return "", err
	}
	tx, err := erc20TokenInstance.Mint(operatorAuth, common.HexToAddress(ownerAddr), amount)
	if nil != err {
		return "", err
	}

	err = waitEthTxFinished(client, tx.Hash(), "MintERC20Token")
	if nil != err {
		return "", err
	}

	return tx.Hash().String(), nil
}

//ApproveAllowance ...
func ApproveAllowance(ownerPrivateKeyStr, tokenAddr string, bridgeBank common.Address, amount *big.Int, client ethinterface.EthClientSpec) (string, error) {
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

	auth, err := PrepareAuth(client, ownerPrivateKey, ownerAddr)
	if nil != err {
		return "", err
	}

	prepareDone = true

	erc20TokenInstance, err := generated.NewBridgeToken(common.HexToAddress(tokenAddr), client)
	if nil != err {
		return "", err
	}

	tx, err := erc20TokenInstance.Approve(auth, bridgeBank, amount)
	if nil != err {
		return "", err
	}

	err = waitEthTxFinished(client, tx.Hash(), "ApproveAllowance")
	if nil != err {
		return "", err
	}

	return tx.Hash().String(), nil
}

//Burn ...
func Burn(ownerPrivateKeyStr, tokenAddrstr, chain33Receiver string, bridgeBank common.Address, amount *big.Int, bridgeBankIns *generated.BridgeBank, client ethinterface.EthClientSpec) (string, error) {
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
	auth, err := PrepareAuth(client, ownerPrivateKey, ownerAddr)
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

	auth, err = PrepareAuth(client, ownerPrivateKey, ownerAddr)
	if nil != err {
		return "", err
	}

	prepareDone = true

	tx, err = bridgeBankIns.BurnBridgeTokens(auth, []byte(chain33Receiver), tokenAddr, amount)
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
func BurnAsync(ownerPrivateKeyStr, tokenAddrstr, chain33Receiver string, amount *big.Int, bridgeBankIns *generated.BridgeBank, client ethinterface.EthClientSpec) (string, error) {
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
	auth, err := PrepareAuth(client, ownerPrivateKey, ownerAddr)
	if nil != err {
		return "", err
	}

	prepareDone = true

	tokenAddr := common.HexToAddress(tokenAddrstr)
	tx, err := bridgeBankIns.BurnBridgeTokens(auth, []byte(chain33Receiver), tokenAddr, amount)
	if nil != err {
		return "", err
	}

	return tx.Hash().String(), nil
}

//TransferToken ...
func TransferToken(tokenAddr, fromPrivateKeyStr, toAddr string, amount *big.Int, client ethinterface.EthClientSpec) (string, error) {
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

	auth, err := PrepareAuth(client, fromPrivateKey, fromAddr)
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

//LockEthErc20Asset ...
func LockEthErc20Asset(ownerPrivateKeyStr, tokenAddrStr, chain33Receiver string, amount *big.Int, client ethinterface.EthClientSpec, bridgeBank *generated.BridgeBank, bridgeBankAddr common.Address) (string, error) {
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
		auth, err := PrepareAuth(client, ownerPrivateKey, ownerAddr)
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

	auth, err := PrepareAuth(client, ownerPrivateKey, ownerAddr)
	if nil != err {
		txslog.Error("LockEthErc20Asset", "PrepareAuth err", err.Error())
		return "", err
	}

	prepareDone = true

	if "" == tokenAddrStr {
		auth.Value = amount
	}

	tx, err := bridgeBank.Lock(auth, []byte(chain33Receiver), tokenAddr, amount)
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
func LockEthErc20AssetAsync(ownerPrivateKeyStr, tokenAddrStr, chain33Receiver string, amount *big.Int, client ethinterface.EthClientSpec, bridgeBank *generated.BridgeBank) (string, error) {
	txslog.Info("LockEthErc20AssetAsync", "ownerPrivateKeyStr", ownerPrivateKeyStr, "tokenAddrStr", tokenAddrStr, "chain33Receiver", chain33Receiver, "amount", amount.String())
	ownerPrivateKey, err := crypto.ToECDSA(common.FromHex(ownerPrivateKeyStr))
	if nil != err {
		return "", err
	}
	ownerAddr := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)

	auth, err := PrepareAuth(client, ownerPrivateKey, ownerAddr)
	if nil != err {
		txslog.Error("LockEthErc20AssetAsync", "PrepareAuth err", err.Error())
		return "", err
	}
	//ETH转账，空地址，且设置value
	var tokenAddr common.Address
	if "" == tokenAddrStr {
		auth.Value = amount
	}

	if "" != tokenAddrStr {
		tokenAddr = common.HexToAddress(tokenAddrStr)
	}
	tx, err := bridgeBank.Lock(auth, []byte(chain33Receiver), tokenAddr, amount)
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
