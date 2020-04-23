// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/hex"
	"testing"

	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	InitExecutor(nil)
}

func TestDecodeLogNewPos33Ticket(t *testing.T) {
	var logTmp = &ReceiptPos33Ticket{}

	dec := types.Encode(logTmp)

	strdec := hex.EncodeToString(dec)
	rlog := &rpctypes.ReceiptLog{
		Ty:  TyLogNewPos33Ticket,
		Log: "0x" + strdec,
	}

	logs := []*rpctypes.ReceiptLog{}
	logs = append(logs, rlog)

	var data = &rpctypes.ReceiptData{
		Ty:   5,
		Logs: logs,
	}
	result, err := rpctypes.DecodeLog([]byte("pos33"), data)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LogNewPos33Ticket", result.Logs[0].TyName)
}

func TestDecodeLogClosePos33Ticket(t *testing.T) {
	var logTmp = &ReceiptPos33Ticket{}

	dec := types.Encode(logTmp)

	strdec := hex.EncodeToString(dec)
	rlog := &rpctypes.ReceiptLog{
		Ty:  TyLogClosePos33Ticket,
		Log: "0x" + strdec,
	}

	logs := []*rpctypes.ReceiptLog{}
	logs = append(logs, rlog)

	var data = &rpctypes.ReceiptData{
		Ty:   5,
		Logs: logs,
	}
	result, err := rpctypes.DecodeLog([]byte("pos33"), data)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LogClosePos33Ticket", result.Logs[0].TyName)
}

func TestDecodeLogMinerPos33Ticket(t *testing.T) {
	var logTmp = &ReceiptPos33Ticket{}

	dec := types.Encode(logTmp)

	strdec := hex.EncodeToString(dec)
	rlog := &rpctypes.ReceiptLog{
		Ty:  TyLogMinerPos33Ticket,
		Log: "0x" + strdec,
	}

	logs := []*rpctypes.ReceiptLog{}
	logs = append(logs, rlog)

	var data = &rpctypes.ReceiptData{
		Ty:   5,
		Logs: logs,
	}
	result, err := rpctypes.DecodeLog([]byte("pos33"), data)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LogMinerPos33Ticket", result.Logs[0].TyName)
}

func TestDecodeLogPos33TicketBind(t *testing.T) {
	var logTmp = &ReceiptPos33TicketBind{}

	dec := types.Encode(logTmp)

	strdec := hex.EncodeToString(dec)
	rlog := &rpctypes.ReceiptLog{
		Ty:  TyLogPos33TicketBind,
		Log: "0x" + strdec,
	}

	logs := []*rpctypes.ReceiptLog{}
	logs = append(logs, rlog)

	var data = &rpctypes.ReceiptData{
		Ty:   5,
		Logs: logs,
	}
	result, err := rpctypes.DecodeLog([]byte("pos33"), data)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LogPos33TicketBind", result.Logs[0].TyName)
}

func TestProtoNewEncodeOldDecode(t *testing.T) {
	tnew := &Pos33TicketMiner{
		Bits:     1,
		Reward:   1,
		TicketId: "id",
		Modify:   []byte("modify"),
		PrivHash: []byte("hash"),
	}
	data := types.Encode(tnew)
	told := &Pos33TicketMinerOld{}
	err := types.Decode(data, told)
	assert.Nil(t, err)
	assert.Equal(t, told.Bits, uint32(1))
	assert.Equal(t, told.Reward, int64(1))
	assert.Equal(t, told.TicketId, "id")
	assert.Equal(t, told.Modify, []byte("modify"))
}
