// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"
	"github.com/stretchr/testify/assert"
)

const (
	testStateCheck = iota + 1
	testStateExec
	testStateExecLocal
	testStateExecDelLocal
)

func testExec(mock *testExecMock, tcArr []*testcase, priv string, t *testing.T) {

	exec := mock.exec
	for i, tc := range tcArr {

		signPriv := priv
		if tc.priv != "" {
			signPriv = tc.priv
		}
		tx, err := createTx(mock, tc.payload, signPriv, tc.systemCreate)
		assert.NoErrorf(t, err, "createTxErr, testIndex=%d", tc.index)
		if err != nil {
			continue
		}
		if len(tc.testSign) > 0 {
			tx.Signature.Signature = append([]byte(""), tc.testSign...)
		}
		if tc.testFee > 0 {
			tx.Fee = tc.testFee
		}
		err = exec.CheckTx(tx, i)
		assert.Equalf(t, tc.expectCheckErr, err, "checkTx err index %d", tc.index)
		if tc.testState == testStateCheck {
			continue
		}

		recp, err := exec.Exec(tx, i)
		recpData := &types.ReceiptData{
			Ty:   recp.GetTy(),
			Logs: recp.GetLogs(),
		}
		if err == nil && len(recp.GetKV()) > 0 {
			util.SaveKVList(mock.stateDB, recp.KV)
			mock.addBlockTx(tx, recpData)
		}
		assert.Equalf(t, tc.expectExecErr, err, "execTx err index %d", tc.index)
		if tc.testState == testStateExec {
			continue
		}
		kvSet, err := exec.ExecLocal(tx, recpData, i)
		for _, kv := range kvSet.GetKV() {
			err := mock.localDB.Set(kv.Key, kv.Value)
			assert.Nil(t, err)
		}
		assert.Equalf(t, tc.expectExecLocalErr, err, "execLocalTx err index %d", tc.index)

		if tc.testState == testStateExecLocal {
			continue
		}

		kvSet, err = exec.ExecDelLocal(tx, recpData, i)
		for _, kv := range kvSet.GetKV() {
			err := mock.localDB.Set(kv.Key, kv.Value)
			assert.Nil(t, err)
		}
		assert.Equalf(t, tc.expectExecDelErr, err, "execDelLocalTx err index %d", tc.index)
	}
}

func TestPrivacy_CheckTx(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	//用于测试双花
	testKeyImage := []byte("testKeyImage")
	mock.stateDB.Set(calcPrivacyKeyImageKey("coins", "bty", testKeyImage), []byte("testval"))
	tcArr := []*testcase{
		{
			index:          1,
			payload:        &pty.Public2Privacy{},
			expectCheckErr: types.ErrInvalidParam,
		},
		{
			index:   2,
			payload: &pty.Public2Privacy{Tokenname: "bty"},
		},
		{
			index:          4,
			payload:        &pty.Privacy2Public{Tokenname: "bty", Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{}}},
			expectCheckErr: pty.ErrNilUtxoInput,
		},
		{
			index:          5,
			payload:        &pty.Privacy2Privacy{Tokenname: "bty", Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{{}}}},
			expectCheckErr: pty.ErrNilUtxoOutput,
		},
		{
			index:          6,
			payload:        &pty.Privacy2Public{Tokenname: "bty", Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{{}}}},
			expectCheckErr: pty.ErrRingSign,
		},
		{
			index:          7,
			payload:        &pty.Privacy2Public{Tokenname: "bty", Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{{KeyImage: testKeyImage}}}},
			expectCheckErr: pty.ErrDoubleSpendOccur,
			testSign:       types.Encode(&types.RingSignature{Items: []*types.RingSignatureItem{{Pubkey: [][]byte{[]byte("test")}}}}),
		},
		{
			index:          8,
			payload:        &pty.Privacy2Public{Tokenname: "bty", Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{{UtxoGlobalIndex: []*pty.UTXOGlobalIndex{{}}}}}},
			expectCheckErr: pty.ErrPubkeysOfUTXO,
			testSign:       types.Encode(&types.RingSignature{Items: []*types.RingSignatureItem{{Pubkey: [][]byte{[]byte("test")}}}}),
		},
		{
			index:          9,
			payload:        &pty.Privacy2Public{Tokenname: "bty", Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{{}}}},
			expectCheckErr: pty.ErrPrivacyTxFeeNotEnough,
			testSign:       types.Encode(&types.RingSignature{Items: []*types.RingSignatureItem{{Pubkey: [][]byte{[]byte("test")}}}}),
		},
		{
			index:          10,
			payload:        &pty.Privacy2Public{Tokenname: "bty", Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{{}}}},
			expectCheckErr: pty.ErrPrivacyTxFeeNotEnough,
			testSign:       types.Encode(&types.RingSignature{Items: []*types.RingSignatureItem{{Pubkey: [][]byte{[]byte("test")}}}}),
			testFee:        pty.PrivacyTxFee,
		},
	}

	for _, tc := range tcArr {
		tc.systemCreate = true
		tc.testState = testStateCheck
	}
	testExec(mock, tcArr, testPrivateKeys[0], t)
}

func TestPrivacy_Exec_Public2Privacy(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	tcArr := []*testcase{
		{
			index: 1,
			payload: &pty.ReqCreatePrivacyTx{
				AssetExec:  "btc-coins",
				Tokenname:  "btc",
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[0],
			},
			expectExecErr: types.ErrExecNameNotAllow,
		},
		{
			index: 2,
			payload: &pty.ReqCreatePrivacyTx{
				Amount:     types.DefaultCoinPrecision * 10001,
				Pubkeypair: testPubkeyPairs[0],
			},
			expectExecErr: types.ErrNoBalance,
		},
		{
			index: 2,
			payload: &pty.ReqCreatePrivacyTx{
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[0],
			},
		},
	}

	for _, tc := range tcArr {
		req := tc.payload.(*pty.ReqCreatePrivacyTx)
		req.ActionType = pty.ActionPublic2Privacy
		tc.testState = testStateExec
	}
	testExec(mock, tcArr, testPrivateKeys[0], t)
}

func TestPrivacy_Exec_Privacy2Privacy(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	tcArr := []*testcase{
		{
			index: 1,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPublic2Privacy,
				Amount:     types.DefaultCoinPrecision * 9,
				Pubkeypair: testPubkeyPairs[0],
			},
		},
		{
			index: 2,
			payload: &pty.ReqCreatePrivacyTx{
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[1],
				From:       testAddrs[0],
			},
		},
	}

	for _, tc := range tcArr {
		req := tc.payload.(*pty.ReqCreatePrivacyTx)
		if req.ActionType == 0 {
			req.ActionType = pty.ActionPrivacy2Privacy
		}
		tc.testState = testStateExec
	}
	testExec(mock, tcArr, testPrivateKeys[0], t)
}

func TestPrivacy_Exec_Privacy2Public(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	tcArr := []*testcase{
		{
			index: 1,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPublic2Privacy,
				Amount:     types.DefaultCoinPrecision * 9,
				Pubkeypair: testPubkeyPairs[0],
			},
		},
		{
			index: 2,
			payload: &pty.ReqCreatePrivacyTx{
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[1],
				From:       testAddrs[0],
				To:         testAddrs[1],
			},
		},
	}

	for _, tc := range tcArr {
		req := tc.payload.(*pty.ReqCreatePrivacyTx)
		if req.ActionType == 0 {
			req.ActionType = pty.ActionPrivacy2Public
		}
		tc.testState = testStateExec
	}
	testExec(mock, tcArr, testPrivateKeys[0], t)
}

func TestPrivacy_ExecLocal(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	tcArr := []*testcase{
		{
			index: 1,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPublic2Privacy,
				Amount:     types.DefaultCoinPrecision * 9,
				Pubkeypair: testPubkeyPairs[0],
			},
		},
		{
			index: 2,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPrivacy2Privacy,
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[1],
				From:       testAddrs[0],
			},
		},
		{
			index: 3,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPrivacy2Public,
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[1],
				From:       testAddrs[0],
				To:         testAddrs[1],
			},
		},
	}

	for _, tc := range tcArr {
		tc.testState = testStateExecLocal
	}
	testExec(mock, tcArr, testPrivateKeys[0], t)
}

func TestPrivacy_ExecDelLocal(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()
	tcArr := []*testcase{
		{
			index: 1,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPublic2Privacy,
				Amount:     types.DefaultCoinPrecision * 9,
				Pubkeypair: testPubkeyPairs[0],
			},
		},
		{
			index: 2,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPrivacy2Privacy,
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[1],
				From:       testAddrs[0],
			},
		},
		{
			index: 3,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPrivacy2Public,
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[1],
				From:       testAddrs[0],
				To:         testAddrs[1],
			},
		},
	}

	for _, tc := range tcArr {
		tc.testState = testStateExecDelLocal
	}
	testExec(mock, tcArr, testPrivateKeys[0], t)
}
