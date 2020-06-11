package events

import (
	"math/big"
	"strings"
	"testing"

	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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

func Test_NewEventWrite(t *testing.T) {
	event := LockEvent{
		Symbol: "bty",
	}

	NewEventWrite("1", event)
	assert.Equal(t, EventRecords["1"].Symbol, "bty")
}

func Test_NewChain33Msg(t *testing.T) {
	_ = NewChain33Msg(MsgBurn, []byte("12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"), common.HexToAddress("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF"),
		"eth", big.NewInt(100000000000000000), common.HexToAddress("0x0000000000000000000000000000000000000000"))
}

func Test_Chain33MsgAttributeKey(t *testing.T) {
	assert.Equal(t, UnsupportedAttributeKey.String(), "unsupported")
	assert.Equal(t, Chain33Sender.String(), "chain33_sender")
	assert.Equal(t, EthereumReceiver.String(), "ethereum_receiver")
	assert.Equal(t, Coin.String(), "amount")
	assert.Equal(t, TokenContractAddress.String(), "token_contract_address")
}
