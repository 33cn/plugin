package ethtxs

// ------------------------------------------------------------
//    Network: Validates input and initializes a websocket Ethereum client.
// ------------------------------------------------------------

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	cross2ethErrors "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SelectAndRoundEthURL: 获取配置列表中的第一个配置项，同时将其调整配置配置项
func SelectAndRoundEthURL(ethURL *[]string) (string, error) {
	if 0 == len(*ethURL) {
		return "", errors.New("NullEthURlCofigured")
	}

	result := (*ethURL)[0]

	if len(*ethURL) > 1 {
		*ethURL = append((*ethURL)[1:], result)
	}
	return result, nil
}

func SetupEthClient(ethURL *[]string) (*ethclient.Client, string, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for i := 0; i < len(*ethURL); i++ {
		urlSelected, err := SelectAndRoundEthURL(ethURL)
		if nil != err {
			txslog.Error("SetupEthClient", "SelectAndRoundEthURL err", err.Error())
			return nil, "", err
		}
		client, err := Dial2MakeEthClient(urlSelected)
		if nil != err {
			txslog.Error("SetupEthClient", "Dial2MakeEthClient err", err.Error())
			continue
		}
		_, err = client.NetworkID(timeout)
		if err != nil {
			txslog.Error("SetupEthClient", "Failed to get NetworkID due to:%s", err.Error())
			continue
		}
		txslog.Debug("SetupEthClient", "SelectAndRoundEthURL:", urlSelected)
		return client, urlSelected, nil
	}
	return nil, "", errors.New("FailedToSetupEthClient")
}

func GetOracleInstance(client ethinterface.EthClientSpec, registry common.Address) (*generated.Oracle, error) {
	nilAddr := common.Address{}
	if nilAddr == registry {
		return nil, cross2ethErrors.ErrContractNotRegistered
	}

	oracleAddr, err := GetAddressFromBridgeRegistry(client, registry, registry, Oracle)
	if nil != err {
		return nil, errors.New("failed to get addr for bridgeBank from registry " + err.Error())
	}
	oracleInstance, err := generated.NewOracle(*oracleAddr, client)
	if nil != err {
		return nil, errors.New("failed to NewOracle " + err.Error())
	}
	return oracleInstance, nil
}

func SetupEthClients(ethURL *[]string, registry common.Address) ([]*EthClientWithUrl, *big.Int, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var Clients []*EthClientWithUrl
	var clientChainID *big.Int
	for i := 0; i < len(*ethURL); i++ {
		urlSelected, err := SelectAndRoundEthURL(ethURL)
		if nil != err {
			txslog.Error("SetupEthClients", "SelectAndRoundEthURL err", err.Error())
			return nil, nil, err
		}
		client, err := Dial2MakeEthClient(urlSelected)
		if nil != err {
			txslog.Error("SetupEthClient", "Dial2MakeEthClient err", err.Error())
			continue
		}
		chainID, err := client.NetworkID(timeout)
		if err != nil {
			txslog.Error("SetupEthClients", "Failed to get NetworkID due to", err.Error())
			continue
		}
		txslog.Debug("SetupEthClients", "SelectAndRoundEthURL:", urlSelected)

		oracleInstance, err := GetOracleInstance(client, registry)
		if nil != err {
			txslog.Error("SetupEthClient", "GetOracleInstance err", err.Error())
		}

		Clients = append(Clients, &EthClientWithUrl{Client: client, ClientUrl: urlSelected, OracleInstance: oracleInstance})
		clientChainID = chainID
	}

	if len(Clients) > 0 {
		return Clients, clientChainID, nil
	}
	return nil, nil, errors.New("FailedToSetupEthClients")
}

func SetupRecommendClients(ethURL *[]string, registry common.Address) ([]*EthClientWithUrl, error) {
	var Clients []*EthClientWithUrl
	for i := 0; i < len(*ethURL); i++ {
		urlSelected, err := SelectAndRoundEthURL(ethURL)
		if nil != err {
			txslog.Error("SetupRecommendClients", "SelectAndRoundEthURL err", err.Error())
			return nil, err
		}
		client, err := Dial2MakeEthClient(urlSelected)
		if nil != err {
			txslog.Error("SetupRecommendClients", "Dial2MakeEthClient err", err.Error())
			continue
		}
		txslog.Debug("SetupRecommendClients", "SelectAndRoundEthURL:", urlSelected)

		oracleInstance, err := GetOracleInstance(client, registry)
		if nil != err {
			txslog.Error("SetupRecommendClients", "GetOracleInstance err", err.Error())
		}
		Clients = append(Clients, &EthClientWithUrl{Client: client, ClientUrl: urlSelected, OracleInstance: oracleInstance})
	}

	if len(Clients) > 0 {
		return Clients, nil
	}
	return nil, errors.New("FailedToSetupRecommendClients")
}

// Dial2MakeEthClient : returns boolean indicating if a URL is valid websocket ethclient
func Dial2MakeEthClient(ethURL string) (*ethclient.Client, error) {
	if strings.TrimSpace(ethURL) == "" {
		return nil, nil
	}

	client, err := ethclient.Dial(ethURL)
	if err != nil {
		return nil, fmt.Errorf("url %s error dialing websocket client %w", ethURL, err)
	}

	return client, nil
}
