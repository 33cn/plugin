// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticket

import (
	"fmt"
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)

func TestTicket(t *testing.T) {
	testTicket(t)
}

func testTicket(t *testing.T) {
	mock33 := testnode.New("testdata/chain33.cfg.toml", nil)
	defer mock33.Close()
	mock33.Listen()
	reply, err := mock33.GetAPI().ExecWalletFunc("ticket", "WalletAutoMiner", &ty.MinerFlag{Flag: 1})
	assert.Nil(t, err)
	assert.Equal(t, true, reply.(*types.Reply).IsOk)
	acc := account.NewCoinsAccount()
	addr := mock33.GetGenesisAddress()
	accounts, err := acc.GetBalance(mock33.GetAPI(), &types.ReqBalance{Execer: "ticket", Addresses: []string{addr}})
	assert.Nil(t, err)
	assert.Equal(t, accounts[0].Balance, int64(0))
	hotaddr := mock33.GetHotAddress()
	_, err = acc.GetBalance(mock33.GetAPI(), &types.ReqBalance{Execer: "coins", Addresses: []string{hotaddr}})
	assert.Nil(t, err)
	//assert.Equal(t, accounts[0].Balance, int64(1000000000000))
	//send to address
	tx := util.CreateCoinsTx(mock33.GetHotKey(), mock33.GetGenesisAddress(), types.Coin/100)
	mock33.SendTx(tx)
	mock33.Wait()
	//bind miner
	tx = createBindMiner(t, hotaddr, addr, mock33.GetGenesisKey())
	hash := mock33.SendTx(tx)
	detail, err := mock33.WaitTx(hash)
	assert.Nil(t, err)
	//debug:
	//js, _ := json.MarshalIndent(detail, "", " ")
	//fmt.Println(string(js))
	_, err = mock33.GetAPI().ExecWalletFunc("ticket", "WalletAutoMiner", &ty.MinerFlag{Flag: 0})
	assert.Nil(t, err)
	status, err := mock33.GetAPI().GetWalletStatus()
	assert.Nil(t, err)
	assert.Equal(t, false, status.IsAutoMining)
	assert.Equal(t, int32(2), detail.Receipt.Ty)
	_, err = mock33.GetAPI().ExecWalletFunc("ticket", "WalletAutoMiner", &ty.MinerFlag{Flag: 1})
	assert.Nil(t, err)
	status, err = mock33.GetAPI().GetWalletStatus()
	assert.Nil(t, err)
	assert.Equal(t, true, status.IsAutoMining)
	err = mock33.WaitHeight(50)
	assert.Nil(t, err)
	//查询票是否自动close，并且购买了新的票
	req := &types.ReqWalletTransactionList{Count: 1000}
	list, err := mock33.GetAPI().WalletTransactionList(req)
	assert.Nil(t, err)
	hastclose := false
	hastopen := false
	for _, tx := range list.TxDetails {
		if tx.ActionName == "tclose" && tx.Receipt.Ty == 2 {
			hastclose = true
		}
		if tx.ActionName == "topen" && tx.Receipt.Ty == 2 {
			hastopen = true
		}
	}
	assert.Equal(t, true, hastclose)
	assert.Equal(t, true, hastopen)
	//查询合约中的余额
	accounts, err = acc.GetBalance(mock33.GetAPI(), &types.ReqBalance{Execer: "ticket", Addresses: []string{addr}})
	assert.Nil(t, err)
	fmt.Println(accounts[0])
}

func createBindMiner(t *testing.T, m, r string, priv crypto.PrivKey) *types.Transaction {
	ety := types.LoadExecutorType("ticket")
	tx, err := ety.Create("Tbind", &ty.TicketBind{MinerAddress: m, ReturnAddress: r})
	assert.Nil(t, err)
	tx, err = types.FormatTx("ticket", tx)
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
	cr, _ := crypto.New("secp256k1")
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
}

func TestProcEvent(t *testing.T) {
	c := Client{}
	ret := c.ProcEvent(&queue.Message{})
	assert.Equal(t, ret, true)
}

func Test_genPrivHash(t *testing.T) {
	c, err := crypto.New(types.GetSignName("", types.SECP256K1))
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
	c := &Client{}

	bits, bt, err := c.getNextRequiredDifficulty(nil, 1)
	assert.NoError(t, err)
	assert.Equal(t, bt, defaultModify)
	assert.Equal(t, bits, types.GetP(0).PowLimitBits)
}
