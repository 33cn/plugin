// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/33cn/chain33/util/testnode"
	wcom "github.com/33cn/chain33/wallet/common"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
)

func Test_WalletTicket(t *testing.T) {
	minerAddr := "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
	t.Log("Begin wallet ticket test")

	cfg, sub := testnode.GetDefaultConfig()
	cfg.Consensus.Name = "ticket"
	mock33 := testnode.NewWithConfig(cfg, sub, nil)
	defer mock33.Close()
	err := mock33.WaitHeight(0)
	assert.Nil(t, err)
	msg, err := mock33.GetAPI().Query(ty.TicketX, "TicketList", &ty.TicketList{Addr: minerAddr, Status: 1})
	assert.Nil(t, err)
	ticketList := msg.(*ty.ReplyTicketList)
	assert.NotNil(t, ticketList)
	//return
	FlushTicket(mock33.GetAPI())
	err = mock33.WaitHeight(2)
	assert.Nil(t, err)
	header, err := mock33.GetAPI().GetLastHeader()
	require.Equal(t, err, nil)
	require.Equal(t, header.Height >= 2, true)

	in := &ty.TicketClose{MinerAddress: minerAddr}
	msg, err = mock33.GetAPI().ExecWalletFunc(ty.TicketX, "CloseTickets", in)
	assert.Nil(t, err)
	hashes := msg.(*types.ReplyHashes)
	assert.NotNil(t, hashes)

	in = &ty.TicketClose{}
	msg, err = mock33.GetAPI().ExecWalletFunc(ty.TicketX, "CloseTickets", in)
	assert.Nil(t, err)
	hashes = msg.(*types.ReplyHashes)
	assert.NotNil(t, hashes)
	t.Log("End wallet ticket test")
}

func Test_ForceCloseTicketList(t *testing.T) {

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	ticket.walletOperate = new(walletOperateMock)
	t1 := &ty.Ticket{Status: 1, IsGenesis: false}
	t2 := &ty.Ticket{Status: 2, IsGenesis: false}
	t3 := &ty.Ticket{Status: 3, IsGenesis: false}
	tlist := []*ty.Ticket{t1, t2, t3}

	r1, r2 := ticket.forceCloseTicketList(0, nil, tlist)
	assert.Nil(t, r1)
	assert.Nil(t, r2)

}

func Test_CloseTicketsByAddr(t *testing.T) {
	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	wallet.api = qapi
	ticket.walletOperate = wallet

	t1 := &ty.Ticket{Status: 1, IsGenesis: false}
	t2 := &ty.Ticket{Status: 2, IsGenesis: false}
	t3 := &ty.Ticket{Status: 3, IsGenesis: false}

	tlist := &ty.ReplyTicketList{Tickets: []*ty.Ticket{t1, t2, t3}}
	qapi.On("Query", ty.TicketX, "TicketList", mock.Anything).Return(tlist, nil)

	ticket.closeTicketsByAddr(0, priKey)

}

func Test_BuyTicketOne(t *testing.T) {

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	ticket.walletOperate = new(walletOperateMock)
	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)
	ticket.buyTicketOne(0, priKey)

}

func Test_BuyMinerAddrTicketOne(t *testing.T) {
	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	wallet.api = qapi
	ticket.walletOperate = wallet

	tlist := &types.ReplyStrings{Datas: []string{"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"}}
	qapi.On("Query", ty.TicketX, "MinerSourceList", mock.Anything).Return(tlist, nil)

	ticket.buyMinerAddrTicketOne(0, priKey)

}

type walletOperateMock struct {
	api client.QueueProtocolAPI
}

func (_m *walletOperateMock) AddrInWallet(addr string) bool {
	return false
}

// CheckWalletStatus provides a mock function with given fields:
func (_m *walletOperateMock) CheckWalletStatus() (bool, error) {
	return false, nil
}

// GetAPI provides a mock function with given fields:
func (_m *walletOperateMock) GetAPI() client.QueueProtocolAPI {
	return _m.api
}

// GetAllPrivKeys provides a mock function with given fields:
func (_m *walletOperateMock) GetAllPrivKeys() ([]crypto.PrivKey, error) {
	return nil, nil
}

// GetBalance provides a mock function with given fields: addr, execer
func (_m *walletOperateMock) GetBalance(addr string, execer string) (*types.Account, error) {

	return &types.Account{Balance: 10}, nil
}

// GetBlockHeight provides a mock function with given fields:
func (_m *walletOperateMock) GetBlockHeight() int64 {
	return 0
}

// GetConfig provides a mock function with given fields:
func (_m *walletOperateMock) GetConfig() *types.Wallet {
	return nil
}

// GetDBStore provides a mock function with given fields:
func (_m *walletOperateMock) GetDBStore() db.DB {
	return nil
}

// GetLastHeader provides a mock function with given fields:
func (_m *walletOperateMock) GetLastHeader() *types.Header {
	return nil
}

// GetMutex provides a mock function with given fields:
func (_m *walletOperateMock) GetMutex() *sync.Mutex {
	return nil
}

// GetPassword provides a mock function with given fields:
func (_m *walletOperateMock) GetPassword() string {
	return ""
}

// GetPrivKeyByAddr provides a mock function with given fields: addr
func (_m *walletOperateMock) GetPrivKeyByAddr(addr string) (crypto.PrivKey, error) {
	return nil, nil
}

// GetRandom provides a mock function with given fields:
func (_m *walletOperateMock) GetRandom() *rand.Rand {
	return nil
}

// GetSignType provides a mock function with given fields:
func (_m *walletOperateMock) GetSignType() int {
	return 0
}

// GetTxDetailByHashs provides a mock function with given fields: ReqHashes
func (_m *walletOperateMock) GetTxDetailByHashs(ReqHashes *types.ReqHashes) {
	return
}

// GetWaitGroup provides a mock function with given fields:
func (_m *walletOperateMock) GetWaitGroup() *sync.WaitGroup {
	return nil
}

// GetWalletAccounts provides a mock function with given fields:
func (_m *walletOperateMock) GetWalletAccounts() ([]*types.WalletAccountStore, error) {
	return nil, nil
}

// GetWalletDone provides a mock function with given fields:
func (_m *walletOperateMock) GetWalletDone() chan struct{} {
	return nil
}

// IsCaughtUp provides a mock function with given fields:
func (_m *walletOperateMock) IsCaughtUp() bool {
	return false
}

// IsClose provides a mock function with given fields:
func (_m *walletOperateMock) IsClose() bool {
	return false
}

// IsWalletLocked provides a mock function with given fields:
func (_m *walletOperateMock) IsWalletLocked() bool {
	return true
}

// Nonce provides a mock function with given fields:
func (_m *walletOperateMock) Nonce() int64 {
	return 0
}

// RegisterMineStatusReporter provides a mock function with given fields: reporter
func (_m *walletOperateMock) RegisterMineStatusReporter(reporter wcom.MineStatusReport) error {
	return nil
}

// SendToAddress provides a mock function with given fields: priv, addrto, amount, note, Istoken, tokenSymbol
func (_m *walletOperateMock) SendToAddress(priv crypto.PrivKey, addrto string, amount int64, note string, Istoken bool, tokenSymbol string) (*types.ReplyHash, error) {
	return nil, nil
}

// SendTransaction provides a mock function with given fields: payload, execer, priv, to
func (_m *walletOperateMock) SendTransaction(payload types.Message, execer []byte, priv crypto.PrivKey, to string) ([]byte, error) {
	return nil, nil
}

// WaitTx provides a mock function with given fields: hash
func (_m *walletOperateMock) WaitTx(hash []byte) *types.TransactionDetail {
	return nil
}

// WaitTxs provides a mock function with given fields: hashes
func (_m *walletOperateMock) WaitTxs(hashes [][]byte) []*types.TransactionDetail {
	return nil
}
