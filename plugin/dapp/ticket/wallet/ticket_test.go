// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
	"github.com/33cn/chain33/common/address"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	wcom "github.com/33cn/chain33/wallet/common"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
)

const (
	sendhash = "sendhash"
)

func TestForceCloseTicketList(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.GetModuleConfig().Consensus.Name = "ticket"

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	qapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	wallet.api = qapi

	ticket.walletOperate = wallet
	t1 := &ty.Ticket{Status: ty.TicketOpened, IsGenesis: false}
	t2 := &ty.Ticket{Status: ty.TicketMined, IsGenesis: false}
	t3 := &ty.Ticket{Status: ty.TicketClosed, IsGenesis: false}

	now := types.Now().Unix()
	t4 := &ty.Ticket{Status: ty.TicketOpened, IsGenesis: false, CreateTime: now}
	t5 := &ty.Ticket{Status: ty.TicketMined, IsGenesis: false, CreateTime: now}
	t6 := &ty.Ticket{Status: ty.TicketMined, IsGenesis: false, MinerTime: now}

	tlist := []*ty.Ticket{t1, t2, t3, t4, t5, t6}

	r1, r2 := ticket.forceCloseTicketList(0, nil, address.DefaultID, tlist)
	assert.Equal(t, []byte(sendhash), r1)
	assert.Nil(t, r2)

}

func TestCloseTicketsByAddr(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.GetModuleConfig().Consensus.Name = "ticket"

	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	qapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	wallet.api = qapi
	ticket.walletOperate = wallet

	t1 := &ty.Ticket{Status: ty.TicketOpened, IsGenesis: false}
	t2 := &ty.Ticket{Status: ty.TicketMined, IsGenesis: false}
	t3 := &ty.Ticket{Status: ty.TicketClosed, IsGenesis: false}

	tlist := &ty.ReplyTicketList{Tickets: []*ty.Ticket{t1, t2, t3}}
	qapi.On("Query", ty.TicketX, "TicketList", mock.Anything).Return(tlist, nil)

	r1, r2 := ticket.closeTicketsByAddr(0, priKey, address.DefaultID)
	assert.Equal(t, []byte(sendhash), r1)
	assert.Nil(t, r2)

}

func TestBuyTicketOne(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.GetModuleConfig().Consensus.Name = "ticket"

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	qapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	wallet.api = qapi
	ticket.walletOperate = wallet
	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)
	hash, r1, r2 := ticket.buyTicketOne(0, priKey, address.DefaultID)
	assert.Equal(t, []byte(sendhash), hash)
	assert.Equal(t, 10, r1)
	assert.Nil(t, r2)

}

func TestBuyMinerAddrTicketOne(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.GetModuleConfig().Consensus.Name = "ticket"

	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	ticket.cfg = &subConfig{}
	ticket.initMinerWhiteList(nil)
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	qapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	wallet.api = qapi
	ticket.walletOperate = wallet

	tlist := &types.ReplyStrings{Datas: []string{"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"}}
	qapi.On("Query", ty.TicketX, "MinerSourceList", mock.Anything).Return(tlist, nil)

	hashs, r2, r3 := ticket.buyMinerAddrTicketOne(0, priKey, address.DefaultID)
	assert.Equal(t, [][]byte{[]byte(sendhash)}, hashs)
	assert.Equal(t, 10, r2)
	assert.Nil(t, r3)

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

	return &types.Account{Balance: 10000000000000, Addr: "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"}, nil
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
func (_m *walletOperateMock) SendToAddress(priv crypto.PrivKey, addressID int32, addrto string, amount int64, note string, Istoken bool, tokenSymbol string) (*types.ReplyHash, error) {
	return &types.ReplyHash{Hash: []byte(sendhash)}, nil
}

// SendTransaction provides a mock function with given fields: payload, execer, priv, to
func (_m *walletOperateMock) SendTransaction(payload types.Message, execer []byte, priv crypto.PrivKey, addressID int32, to string) (hash []byte, err error) {
	return []byte(sendhash), nil
}

// WaitTx provides a mock function with given fields: hash
func (_m *walletOperateMock) WaitTx(hash []byte) *types.TransactionDetail {
	return nil
}

// WaitTxs provides a mock function with given fields: hashes
func (_m *walletOperateMock) WaitTxs(hashes [][]byte) []*types.TransactionDetail {
	return nil
}

// GetCoinType provides a mock function with given fields:
func (_m *walletOperateMock) GetCoinType() uint32 {
	return 0
}
