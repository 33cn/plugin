package events

import (
	"strings"
	"testing"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UnpackLogLock(t *testing.T) {
	abiJSON := generated.BridgeBankABI
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	require.Nil(t, err)
	eventName := LogLock.String()
	eventData := []byte("nil")

	_, err = UnpackLogLock(contractABI, eventName, eventData)
	require.NotNil(t, err)

	_, err = UnpackLogBurn(contractABI, eventName, eventData)
	require.NotNil(t, err)
}

func Test_Chain33MsgAttributeKey(t *testing.T) {
	assert.Equal(t, UnsupportedAttributeKey.String(), "unsupported")
	assert.Equal(t, Chain33Sender.String(), "chain33_sender")
	assert.Equal(t, EthereumReceiver.String(), "ethereum_receiver")
	assert.Equal(t, Coin.String(), "amount")
	assert.Equal(t, TokenContractAddress.String(), "token_contract_address")
}
