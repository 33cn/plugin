// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"reflect"

	//log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

// RelayX name for executor
var RelayX = "relay"

//var tlog = log.New("module", name)
//log for relay
const (
	TyLogRelayCreate       = 350
	TyLogRelayRevokeCreate = 351
	TyLogRelayAccept       = 352
	TyLogRelayRevokeAccept = 353
	TyLogRelayConfirmTx    = 354
	TyLogRelayFinishTx     = 355
	TyLogRelayRcvBTCHead   = 356
)

// relay
const (
	// RelayRevokeCreate revoke created order
	RelayRevokeCreate = iota
	// RelayRevokeAccept revoke accept order
	RelayRevokeAccept
)

const (
	// RelayOrderBuy define relay buy order
	RelayOrderBuy = iota
	// RelayOrderSell define relay sell order
	RelayOrderSell
)

// RelayOrderOperation buy or sell operation
var RelayOrderOperation = map[uint32]string{
	RelayOrderBuy:  "buy",
	RelayOrderSell: "sell",
}

const (
	// RelayUnlock revoke order
	RelayUnlock = iota
	// RelayCancel order owner cancel order
	RelayCancel
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(RelayX))
	types.RegistorExecutor(RelayX, NewType())
	types.RegisterDappFork(RelayX, "Enable", 570000)
}

// NewType new relay type
func NewType() *RelayType {
	c := &RelayType{}
	c.SetChild(c)
	return c
}

// GetPayload return relay action msg
func (r *RelayType) GetPayload() types.Message {
	return &RelayAction{}
}

// RelayType relay exec type
type RelayType struct {
	types.ExecTypeBase
}

// GetName return relay name
func (r *RelayType) GetName() string {
	return RelayX
}

// GetLogMap return receipt log map function
func (r *RelayType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogRelayCreate:       {Ty: reflect.TypeOf(ReceiptRelayLog{}), Name: "LogRelayCreate"},
		TyLogRelayRevokeCreate: {Ty: reflect.TypeOf(ReceiptRelayLog{}), Name: "LogRelayRevokeCreate"},
		TyLogRelayAccept:       {Ty: reflect.TypeOf(ReceiptRelayLog{}), Name: "LogRelayAccept"},
		TyLogRelayRevokeAccept: {Ty: reflect.TypeOf(ReceiptRelayLog{}), Name: "LogRelayRevokeAccept"},
		TyLogRelayConfirmTx:    {Ty: reflect.TypeOf(ReceiptRelayLog{}), Name: "LogRelayConfirmTx"},
		TyLogRelayFinishTx:     {Ty: reflect.TypeOf(ReceiptRelayLog{}), Name: "LogRelayFinishTx"},
		TyLogRelayRcvBTCHead:   {Ty: reflect.TypeOf(ReceiptRelayRcvBTCHeaders{}), Name: "LogRelayRcvBTCHead"},
	}
}

const (
	// RelayActionCreate relay create order action
	RelayActionCreate = iota
	// RelayActionAccept accept order action
	RelayActionAccept
	// RelayActionRevoke revoke order action
	RelayActionRevoke
	// RelayActionConfirmTx confirm tx action
	RelayActionConfirmTx
	// RelayActionVerifyTx relayd send this tx to verify btc tx
	RelayActionVerifyTx
	// RelayActionVerifyCmdTx verify tx by cli action
	RelayActionVerifyCmdTx
	// RelayActionRcvBTCHeaders relay rcv BTC headers by this
	RelayActionRcvBTCHeaders
)

// GetTypeMap get relay action type map
func (r *RelayType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Create":     RelayActionCreate,
		"Accept":     RelayActionAccept,
		"Revoke":     RelayActionRevoke,
		"ConfirmTx":  RelayActionConfirmTx,
		"Verify":     RelayActionVerifyTx,
		"VerifyCli":  RelayActionVerifyCmdTx,
		"BtcHeaders": RelayActionRcvBTCHeaders,
	}
}

// ActionName return action name
func (r RelayType) ActionName(tx *types.Transaction) string {
	var relay RelayAction
	err := types.Decode(tx.Payload, &relay)
	if err != nil {
		return "unknown-relay-action-err"
	}
	if relay.Ty == RelayActionCreate && relay.GetCreate() != nil {
		return "relayCreateTx"
	}
	if relay.Ty == RelayActionRevoke && relay.GetRevoke() != nil {
		return "relayRevokeTx"
	}
	if relay.Ty == RelayActionAccept && relay.GetAccept() != nil {
		return "relayAcceptTx"
	}
	if relay.Ty == RelayActionConfirmTx && relay.GetConfirmTx() != nil {
		return "relayConfirmTx"
	}
	if relay.Ty == RelayActionVerifyTx && relay.GetVerify() != nil {
		return "relayVerifyTx"
	}
	if relay.Ty == RelayActionRcvBTCHeaders && relay.GetBtcHeaders() != nil {
		return "relay-receive-btc-heads"
	}
	return "unknown"
}

// Amount return relay create bty amount
func (r *RelayType) Amount(tx *types.Transaction) (int64, error) {
	data, err := r.DecodePayload(tx)
	if err != nil {
		return 0, err
	}
	relay := data.(*RelayAction)
	if RelayActionCreate == relay.Ty && relay.GetCreate() != nil {
		return int64(relay.GetCreate().BtyAmount), nil
	}
	return 0, nil
}

// CreateTx relay create tx TODO 暂时不修改实现， 先完成结构的重构
func (r *RelayType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	var tx *types.Transaction
	return tx, nil
}
