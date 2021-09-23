// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"
	"github.com/stretchr/testify/assert"
)

var (
	execTestCases = []*testcase{
		{
			index: 1,
			payload: &pty.ReqCreatePrivacyTx{
				ActionType: pty.ActionPublic2Privacy,
				Amount:     types.DefaultCoinPrecision,
				Pubkeypair: testPubkeyPairs[0],
			},
		},
	}
)

type queryTestCase struct {
	index             int
	funcName          string
	params            types.Message
	expectErr         error
	expectReply       types.Message
	disableReplyCheck bool
}

func testQuery(mock *testExecMock, tcArr []*queryTestCase, t *testing.T) {

	for _, tc := range tcArr {

		reply, err := mock.exec.Query(tc.funcName, types.Encode(tc.params))
		assert.Equalf(t, tc.expectErr, err, "queryTest index=%d", tc.index)
		if err == nil && !tc.disableReplyCheck {
			assert.Equalf(t, tc.expectReply, reply, "queryTest index=%d", tc.index)
		}
	}
}

func TestPrivacy_Query_ShowAmountsOfUTXO(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()

	for _, tc := range execTestCases {
		tc.testState = testStateExecLocal
	}
	testExec(mock, execTestCases, testPrivateKeys[0], t)

	queryCases := []*queryTestCase{
		{
			index: 1,
			params: &pty.ReqPrivacyToken{
				AssetSymbol: "btc",
			},
			expectErr: types.ErrNotFound,
		},
		{
			index: 2,
			params: &pty.ReqPrivacyToken{
				AssetSymbol: "bty",
			},
			expectReply: &pty.ReplyPrivacyAmounts{
				AmountDetail: []*pty.AmountDetail{
					{Amount: types.DefaultCoinPrecision, Count: 1},
				},
			},
		},
	}

	for _, tc := range queryCases {
		req := tc.params.(*pty.ReqPrivacyToken)
		req.AssetExec = "coins"
		tc.funcName = "ShowAmountsOfUTXO"
	}
	testQuery(mock, queryCases, t)
}

func TestPrivacy_Query_ShowUTXOs4SpecifiedAmount(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()

	for _, tc := range execTestCases {
		tc.testState = testStateExecLocal
	}
	testExec(mock, execTestCases, testPrivateKeys[0], t)

	queryCases := []*queryTestCase{
		{
			index: 1,
			params: &pty.ReqPrivacyToken{
				AssetSymbol: "bty",
			},
			expectErr: types.ErrNotFound,
		},
		{
			index: 2,
			params: &pty.ReqPrivacyToken{
				AssetSymbol: "bty",
				Amount:      types.DefaultCoinPrecision,
			},
			disableReplyCheck: true,
		},
	}

	for _, tc := range queryCases {
		req := tc.params.(*pty.ReqPrivacyToken)
		req.AssetExec = "coins"
		tc.funcName = "ShowUTXOs4SpecifiedAmount"
	}
	testQuery(mock, queryCases, t)
}

func TestPrivacy_Query_GetUTXOGlobalIndex(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()

	for _, tc := range execTestCases {
		tc.testState = testStateExecLocal
	}
	testExec(mock, execTestCases, testPrivateKeys[0], t)

	queryCases := []*queryTestCase{
		{
			index:             1,
			params:            &pty.ReqUTXOGlobalIndex{},
			disableReplyCheck: true,
		},
		{
			index: 2,
			params: &pty.ReqUTXOGlobalIndex{
				AssetSymbol: "btc",
				MixCount:    1,
				Amount:      []int64{types.DefaultCoinPrecision},
			},
			disableReplyCheck: true,
			expectErr:         types.ErrNotFound,
		},
		{
			index: 3,
			params: &pty.ReqUTXOGlobalIndex{
				AssetSymbol: "bty",
				MixCount:    1,
				Amount:      []int64{types.DefaultCoinPrecision, types.DefaultCoinPrecision * 2},
			},
			disableReplyCheck: true,
			expectErr:         types.ErrNotFound,
		},
		{
			index: 4,
			params: &pty.ReqUTXOGlobalIndex{
				AssetSymbol: "bty",
				MixCount:    1,
				Amount:      []int64{types.DefaultCoinPrecision},
			},
			disableReplyCheck: true,
		},
	}

	for _, tc := range queryCases {
		req := tc.params.(*pty.ReqUTXOGlobalIndex)
		if req.AssetExec == "" {
			req.AssetExec = "coins"
		}
		tc.funcName = "GetUTXOGlobalIndex"
	}
	testQuery(mock, queryCases, t)
}

func TestPrivacy_Query_GetTxsByAddr(t *testing.T) {

	mock := &testExecMock{}
	mock.InitEnv()
	defer mock.FreeEnv()

	for _, tc := range execTestCases {
		tc.testState = testStateExecLocal
	}
	testExec(mock, execTestCases, testPrivateKeys[0], t)

	queryCases := []*queryTestCase{
		{
			index: 1,
			params: &types.ReqAddr{
				Addr: testAddrs[0],
			},
			expectErr: types.ErrNotFound,
		},
	}

	for _, tc := range queryCases {
		tc.funcName = "GetTxsByAddr"
	}
	testQuery(mock, queryCases, t)
}
