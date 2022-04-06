package ethtxs

// ------------------------------------------------------------
//    Network: Validates input and initializes a websocket Ethereum client.
// ------------------------------------------------------------

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
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

func SetupEthClient(ethURL *[]string) (*ethclient.Client, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	for i := 0; i < len(*ethURL); i++ {
		urlSelected, err := SelectAndRoundEthURL(ethURL)
		if nil != err {
			txslog.Error("SetupEthClient", "SelectAndRoundEthURL err", err.Error())
			return nil, err
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
		return client, nil
	}
	return nil, errors.New("FailedToSetupEthClient")
}

func SetupEthClients(ethURL *[]string) ([]ethinterface.EthClientSpec, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	var Clients []ethinterface.EthClientSpec
	for i := 0; i < len(*ethURL); i++ {
		urlSelected, err := SelectAndRoundEthURL(ethURL)
		if nil != err {
			txslog.Error("SetupEthClients", "SelectAndRoundEthURL err", err.Error())
			return nil, err
		}
		client, err := Dial2MakeEthClient(urlSelected)
		if nil != err {
			txslog.Error("SetupEthClient", "Dial2MakeEthClient err", err.Error())
			continue
		}
		_, err = client.NetworkID(timeout)
		if err != nil {
			txslog.Error("SetupEthClients", "Failed to get NetworkID due to:%s", err.Error())
			continue
		}
		txslog.Debug("SetupEthClients", "SelectAndRoundEthURL:", urlSelected)
		Clients = append(Clients, client)
	}

	if len(Clients) > 0 {
		return Clients, nil
	}
	return nil, errors.New("FailedToSetupEthClients")
}

func SetupRecommendClients(ethURL *[]string) ([]ethinterface.EthClientSpec, error) {
	var Clients []ethinterface.EthClientSpec
	for i := 0; i < len(*ethURL); i++ {
		urlSelected, err := SelectAndRoundEthURL(ethURL)
		if nil != err {
			txslog.Error("SetupRecommendClients", "SelectAndRoundEthURL err", err.Error())
			return nil, err
		}
		client, err := Dial2MakeEthClient(urlSelected)
		if nil != err {
			txslog.Error("SetupEthClient", "Dial2MakeEthClient err", err.Error())
			continue
		}
		txslog.Debug("SetupRecommendClients", "SelectAndRoundEthURL:", urlSelected)
		Clients = append(Clients, client)
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
