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

func TestDecodeLogNewTicket(t *testing.T) {
	var logTmp = &ReceiptTicket{}

	dec := types.Encode(logTmp)

	strdec := hex.EncodeToString(dec)
	rlog := &rpctypes.ReceiptLog{
		Ty:  TyLogNewTicket,
		Log: "0x" + strdec,
	}

	logs := []*rpctypes.ReceiptLog{}
	logs = append(logs, rlog)

	var data = &rpctypes.ReceiptData{
		Ty:   5,
		Logs: logs,
	}
	result, err := rpctypes.DecodeLog([]byte("ticket"), data)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LogNewTicket", result.Logs[0].TyName)
}

func TestDecodeLogCloseTicket(t *testing.T) {
	var logTmp = &ReceiptTicket{}

	dec := types.Encode(logTmp)

	strdec := hex.EncodeToString(dec)
	rlog := &rpctypes.ReceiptLog{
		Ty:  TyLogCloseTicket,
		Log: "0x" + strdec,
	}

	logs := []*rpctypes.ReceiptLog{}
	logs = append(logs, rlog)

	var data = &rpctypes.ReceiptData{
		Ty:   5,
		Logs: logs,
	}
	result, err := rpctypes.DecodeLog([]byte("ticket"), data)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LogCloseTicket", result.Logs[0].TyName)
}

func TestDecodeLogMinerTicket(t *testing.T) {
	var logTmp = &ReceiptTicket{}

	dec := types.Encode(logTmp)

	strdec := hex.EncodeToString(dec)
	rlog := &rpctypes.ReceiptLog{
		Ty:  TyLogMinerTicket,
		Log: "0x" + strdec,
	}

	logs := []*rpctypes.ReceiptLog{}
	logs = append(logs, rlog)

	var data = &rpctypes.ReceiptData{
		Ty:   5,
		Logs: logs,
	}
	result, err := rpctypes.DecodeLog([]byte("ticket"), data)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LogMinerTicket", result.Logs[0].TyName)
}

func TestDecodeLogTicketBind(t *testing.T) {
	var logTmp = &ReceiptTicketBind{}

	dec := types.Encode(logTmp)

	strdec := hex.EncodeToString(dec)
	rlog := &rpctypes.ReceiptLog{
		Ty:  TyLogTicketBind,
		Log: "0x" + strdec,
	}

	logs := []*rpctypes.ReceiptLog{}
	logs = append(logs, rlog)

	var data = &rpctypes.ReceiptData{
		Ty:   5,
		Logs: logs,
	}
	result, err := rpctypes.DecodeLog([]byte("ticket"), data)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "LogTicketBind", result.Logs[0].TyName)
}

func TestProtoNewEncodeOldDecode(t *testing.T) {
	tnew := &TicketMiner{
		Bits:     1,
		Reward:   1,
		TicketId: "id",
		Modify:   []byte("modify"),
		PrivHash: []byte("hash"),
	}
	data := types.Encode(tnew)
	told := &TicketMinerOld{}
	err := types.Decode(data, told)
	assert.Nil(t, err)
	assert.Equal(t, told.Bits, uint32(1))
	assert.Equal(t, told.Reward, int64(1))
	assert.Equal(t, told.TicketId, "id")
	assert.Equal(t, told.Modify, []byte("modify"))
}
