// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticket

import (
	"crypto/ecdsa"
	"fmt"
	"testing"
	"time"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/crypto"
	vrf "github.com/33cn/chain33/common/vrf/secp256k1"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"

	apimocks "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/merkle"
	_ "github.com/33cn/chain33/system"
	drivers "github.com/33cn/chain33/system/consensus"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
	"github.com/stretchr/testify/mock"
)

func TestTicket(t *testing.T) {
	testTicket(t)
}

func testTicket(t *testing.T) {
	mock33 := testnode.New("testdata/chain33.cfg.toml", nil)
	defer mock33.Close()
	cfg := mock33.GetClient().GetConfig()
	mock33.Listen()
	reply, err := mock33.GetAPI().ExecWalletFunc("ticket", "WalletAutoMiner", &ty.MinerFlag{Flag: 1})
	assert.Nil(t, err)
	assert.Equal(t, true, reply.(*types.Reply).IsOk)
	acc := account.NewCoinsAccount(cfg)
	addr := mock33.GetGenesisAddress()
	accounts, err := acc.GetBalance(mock33.GetAPI(), &types.ReqBalance{Execer: "ticket", Addresses: []string{addr}})
	assert.Nil(t, err)
	assert.Equal(t, accounts[0].Balance, int64(0))
	hotaddr := mock33.GetHotAddress()
	_, err = acc.GetBalance(mock33.GetAPI(), &types.ReqBalance{Execer: "coins", Addresses: []string{hotaddr}})
	assert.Nil(t, err)
	//assert.Equal(t, accounts[0].Balance, int64(1000000000000))
	//send to address
	tx := util.CreateCoinsTx(cfg, mock33.GetHotKey(), mock33.GetGenesisAddress(), types.DefaultCoinPrecision/100)
	mock33.SendTx(tx)
	mock33.Wait()
	//bind miner
	tx = createBindMiner(cfg, t, hotaddr, addr, mock33.GetGenesisKey())
	hash := mock33.SendTx(tx)
	detail, err := mock33.WaitTx(hash)
	assert.Nil(t, err)
	//debug:
	//js, _ := json.MarshalIndent(detail, "", " ")
	//fmt.Println(string(js))
	_, err = mock33.GetAPI().ExecWalletFunc("ticket", "WalletAutoMiner", &ty.MinerFlag{Flag: 0})
	assert.Nil(t, err)
	status, err := mock33.GetAPI().ExecWalletFunc("wallet", "GetWalletStatus", &types.ReqNil{})
	assert.Nil(t, err)
	assert.Equal(t, false, status.(*types.WalletStatus).IsAutoMining)
	assert.Equal(t, int32(2), detail.Receipt.Ty)
	_, err = mock33.GetAPI().ExecWalletFunc("ticket", "WalletAutoMiner", &ty.MinerFlag{Flag: 1})
	assert.Nil(t, err)
	status, err = mock33.GetAPI().ExecWalletFunc("wallet", "GetWalletStatus", &types.ReqNil{})
	assert.Nil(t, err)
	assert.Equal(t, true, status.(*types.WalletStatus).IsAutoMining)
	start := time.Now()
	height := int64(0)
	hastclose := false
	hastopen := false
	for {
		height += 20
		err = mock33.WaitHeight(height)
		assert.Nil(t, err)
		//查询票是否自动close，并且购买了新的票
		req := &types.ReqWalletTransactionList{Count: 1000}
		resp, err := mock33.GetAPI().ExecWalletFunc("wallet", "WalletTransactionList", req)
		list := resp.(*types.WalletTxDetails)
		assert.Nil(t, err)
		for _, tx := range list.TxDetails {
			if tx.ActionName == "tclose" && tx.Receipt.Ty == 2 {
				hastclose = true
			}
			if tx.ActionName == "topen" && tx.Receipt.Ty == 2 {
				hastopen = true
			}
		}
		if hastopen == true && hastclose == true || time.Since(start) > 100*time.Second {
			break
		}
	}
	assert.Equal(t, true, hastclose)
	assert.Equal(t, true, hastopen)
	//查询合约中的余额
	accounts, err = acc.GetBalance(mock33.GetAPI(), &types.ReqBalance{Execer: "ticket", Addresses: []string{addr}})
	assert.Nil(t, err)
	fmt.Println(accounts[0])

	//测试最优节点的选择,难度相同
	lastBlock := mock33.GetLastBlock()
	temblock := types.Clone(lastBlock)
	newblock := temblock.(*types.Block)
	newblock.GetTxs()[0].Nonce = newblock.GetTxs()[0].Nonce + 1
	newblock.TxHash = merkle.CalcMerkleRoot(cfg, newblock.GetHeight(), newblock.GetTxs())

	isbestBlock := util.CmpBestBlock(mock33.GetClient(), newblock, lastBlock.Hash(cfg))
	assert.Equal(t, isbestBlock, false)
}

func createBindMiner(cfg *types.Chain33Config, t *testing.T, m, r string, priv crypto.PrivKey) *types.Transaction {
	ety := types.LoadExecutorType("ticket")
	tx, err := ety.Create("Tbind", &ty.TicketBind{MinerAddress: m, ReturnAddress: r})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, "ticket", tx)
	assert.Nil(t, err)
	tx.Sign(types.SECP256K1, priv)
	return tx
}

func TestTicketMap(t *testing.T) {
	c := Client{}
	ticketList := &ty.ReplyTicketList{}
	ticketList.Tickets = []*ty.Ticket{
		{TicketId: "1111"},
		{TicketId: "2222"},
		{TicketId: "3333"},
		{TicketId: "4444"},
	}
	privmap := make(map[string]crypto.PrivKey)
	//通过privkey生成一个pubkey然后换算成对应的addr
	cr, _ := crypto.Load("secp256k1", -1)
	priv, _ := cr.PrivKeyFromBytes([]byte("2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989"))
	privmap["1111"] = priv
	privmap["2222"] = priv
	privmap["3333"] = priv
	privmap["4444"] = priv

	assert.Equal(t, c.getTicketCount(), int64(0))
	c.setTicket(ticketList, privmap)
	assert.Equal(t, c.getTicketCount(), int64(4))
	c.delTicket("3333")
	assert.Equal(t, c.getTicketCount(), int64(3))

	c.setTicket(ticketList, nil)
	assert.Equal(t, c.getTicketCount(), int64(0))

	c.setTicket(nil, privmap)
	assert.Equal(t, c.getTicketCount(), int64(0))

	c.setTicket(nil, nil)
	assert.Equal(t, c.getTicketCount(), int64(0))
	_, err := c.Query_GetTicketCount(&types.ReqNil{})
	assert.Nil(t, err)
}

func TestProcEvent(t *testing.T) {
	c := Client{}
	ret := c.ProcEvent(&queue.Message{})
	assert.Equal(t, ret, true)
}

func Test_genPrivHash(t *testing.T) {
	c, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	assert.NoError(t, err)
	priv, _ := c.GenKey()

	bt, err := genPrivHash(priv, "AA:BB:CC:DD")
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(bt))

	bt, err = genPrivHash(priv, "111:222:333:444")
	assert.NoError(t, err)
	assert.Equal(t, 32, len(bt))
}

func Test_getNextRequiredDifficulty(t *testing.T) {
	cfg := types.NewChain33Config(types.ReadFile("testdata/chain33.cfg.toml"))

	api := new(apimocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	c := &Client{BaseClient: &drivers.BaseClient{}}
	c.SetAPI(api)

	bits, bt, err := c.getNextRequiredDifficulty(nil, 1)
	assert.NoError(t, err)
	assert.Equal(t, bt, defaultModify)
	assert.Equal(t, bits, cfg.GetP(0).PowLimitBits)
}

func Test_vrfVerify(t *testing.T) {
	c, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	assert.NoError(t, err)
	priv, err := c.GenKey()
	assert.NoError(t, err)
	pub := priv.PubKey().Bytes()

	privKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), priv.Bytes())
	vpriv := &vrf.PrivateKey{PrivateKey: (*ecdsa.PrivateKey)(privKey)}

	m1 := []byte("data1")
	m2 := []byte("data2")
	m3 := []byte("data2")
	hash1, proof1 := vpriv.Evaluate(m1)
	hash2, proof2 := vpriv.Evaluate(m2)
	hash3, proof3 := vpriv.Evaluate(m3)
	for _, tc := range []struct {
		m     []byte
		hash  [32]byte
		proof []byte
		err   error
	}{
		{m1, hash1, proof1, nil},
		{m2, hash2, proof2, nil},
		{m3, hash3, proof3, nil},
		{m3, hash3, proof2, nil},
		{m3, hash3, proof1, ty.ErrVrfVerify},
		{m3, hash1, proof3, ty.ErrVrfVerify},
	} {
		err := vrfVerify(pub, tc.m, tc.proof, tc.hash[:])
		if got, want := err, tc.err; got != want {
			t.Errorf("vrfVerify(%s, %x): %v, want %v", tc.m, tc.proof, got, want)
		}
	}
}
