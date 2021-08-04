package executor_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	"github.com/33cn/plugin/plugin/dapp/ticket/executor"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin/consensus/init"
	_ "github.com/33cn/plugin/plugin/dapp/ticket"
)

var mock33 *testnode.Chain33Mock

func TestMain(m *testing.M) {
	mock33 = testnode.New("testdata/chain33.cfg.toml", nil)
	mock33.Listen()
	code := m.Run()
	mock33.Close()
	os.Exit(code)
}

func TestTicketPrice(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
	//test price
	ti := &executor.DB{}
	assert.Equal(t, ti.GetRealPrice(cfg), 10000*types.DefaultCoinPrecision)

	ti = &executor.DB{}
	ti.Price = 10
	assert.Equal(t, ti.GetRealPrice(cfg), int64(10))
}

func TestCheckFork(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
	assert.Equal(t, int64(1), cfg.GetFork("ForkChainParamV2"))
	p1 := ty.GetTicketMinerParam(cfg, 0)
	assert.Equal(t, 10000*types.DefaultCoinPrecision, p1.TicketPrice)
	p1 = ty.GetTicketMinerParam(cfg, 1)
	assert.Equal(t, 3000*types.DefaultCoinPrecision, p1.TicketPrice)
	p1 = ty.GetTicketMinerParam(cfg, 2)
	assert.Equal(t, 3000*types.DefaultCoinPrecision, p1.TicketPrice)
	p1 = ty.GetTicketMinerParam(cfg, 3)
	assert.Equal(t, 3000*types.DefaultCoinPrecision, p1.TicketPrice)
}

func TestTicket(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
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
	tx = createBindMiner(t, cfg, hotaddr, addr, mock33.GetGenesisKey())
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
	for i := mock33.GetLastBlock().Height; i < 100; i++ {
		err = mock33.WaitHeight(i)
		assert.Nil(t, err)
		//查询票是否自动close，并且购买了新的票
		req := &types.ReqWalletTransactionList{Count: 1000}
		resp, err := mock33.GetAPI().ExecWalletFunc("wallet", "WalletTransactionList", req)
		assert.Nil(t, err)
		list := resp.(*types.WalletTxDetails)
		hastclose := false
		hastopen := false
		for _, tx := range list.TxDetails {
			if tx.Height < 1 {
				continue
			}
			if tx.ActionName == "tclose" && tx.Receipt.Ty == 2 {
				hastclose = true
			}
			if tx.ActionName == "topen" && tx.Receipt.Ty == 2 {
				hastopen = true
				fmt.Println(tx)
				list := ticketList(t, mock33, &ty.TicketList{Addr: tx.Fromaddr, Status: 1})
				for _, ti := range list.GetTickets() {
					if strings.Contains(ti.TicketId, hex.EncodeToString(tx.Txhash)) {
						assert.Equal(t, 3000*types.DefaultCoinPrecision, ti.Price)
					}
				}
			}
		}
		if hastclose && hastopen {
			return
		}
	}
	t.Error("wait 100 , open and close not happened")
}

func createBindMiner(t *testing.T, cfg *types.Chain33Config, m, r string, priv crypto.PrivKey) *types.Transaction {
	ety := types.LoadExecutorType("ticket")
	tx, err := ety.Create("Tbind", &ty.TicketBind{MinerAddress: m, ReturnAddress: r})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, "ticket", tx)
	assert.Nil(t, err)
	tx.Sign(types.SECP256K1, priv)
	return tx
}

func ticketList(t *testing.T, mock33 *testnode.Chain33Mock, req proto.Message) *ty.ReplyTicketList {
	data, err := mock33.GetAPI().Query("ticket", "TicketList", req)
	assert.Nil(t, err)
	return data.(*ty.ReplyTicketList)
}
