package events

import (
	"strings"
	"testing"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
)

func Test_UnpackLogLock(t *testing.T) {
	abiJSON := generated.BridgeBankABI
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	require.Nil(t, err)
	eventName := LogLockFromETH.String()
	eventData := []byte("nil")

	_, err = UnpackLogLock(contractABI, eventName, eventData)
	require.NotNil(t, err)

	_, err = UnpackLogBurn(contractABI, eventName, eventData)
	require.NotNil(t, err)
}
