package ethtxs

// ------------------------------------------------------------
//    Network: Validates input and initializes a websocket Ethereum client.
// ------------------------------------------------------------

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
)

// SelectAndRoundEthURL: 获取配置列表中的第一个配置项，同时将其调整配置配置项
func SelectAndRoundEthURL(ethURL *[]string) (string, error) {
	if 0 == len(*ethURL) {
		return "", errors.New("NullEthURlCofigured")
	}

	result := (*ethURL)[0]

	if len(*ethURL) > 0 {
		*ethURL = append((*ethURL)[1:], result)
	}
	return result, nil
}

func SetupEthClient(ethURL *[]string) (*ethclient.Client, error) {
	for i := 0; i < len(*ethURL); i++ {
		urlSelected, err := SelectAndRoundEthURL(ethURL)
		if nil != err {
			txslog.Error("SetupEthClient", "SelectAndRoundEthURL err", err.Error())
			return nil, err
		}
		client, err := Dial2MakeEthClient(urlSelected)
		if nil != err {
			continue
		}
		return client, nil
	}
	return nil, errors.New("FailedToSetupEthClient")
}

// Dial2MakeEthClient : returns boolean indicating if a URL is valid websocket ethclient
func Dial2MakeEthClient(ethURL string) (*ethclient.Client, error) {
	if strings.TrimSpace(ethURL) == "" {
		return nil, nil
	}

	client, err := ethclient.Dial(ethURL)
	if err != nil {
		return nil, fmt.Errorf("error dialing websocket client %w", err)
	}

	return client, nil
}
