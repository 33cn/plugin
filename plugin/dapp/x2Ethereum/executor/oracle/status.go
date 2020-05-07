package oracle

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

// StatusText is an enum used to represent the status of the prophecy
type StatusText int

var StatusTextToString = [...]string{"pending", "success", "failed", "withdrawed"}
var StringToStatusText = map[string]types.EthBridgeStatus{
	"pending":    types.EthBridgeStatus_PendingStatusText,
	"success":    types.EthBridgeStatus_SuccessStatusText,
	"failed":     types.EthBridgeStatus_FailedStatusText,
	"withdrawed": types.EthBridgeStatus_WithdrawedStatusText,
}

func (text StatusText) String() string {
	return StatusTextToString[text]
}

func (text StatusText) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", text.String())), nil
}

func (text *StatusText) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	stringKey, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	// Note that if the string cannot be found then it will be set to the zero value, 'pending' in this case.
	*text = StatusText(StringToStatusText[stringKey])
	return nil
}
