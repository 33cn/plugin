package test

import (
	"fmt"
	"log"
	"time"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/plugin/plugin/dapp/exchange/executor"
	"github.com/golang/protobuf/proto"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/queue"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

//ExecCli ...
type ExecCli struct {
	ldb        db.KVDB
	sdb        db.DB
	height     int64
	blockTime  int64
	difficulty uint64
	q          queue.Queue
	cfg        *types.Chain33Config
	execAddr   string

	accA  *account.DB //exec account
	accA1 *account.DB //exec token account
	accB  *account.DB
	accB1 *account.DB
	accC  *account.DB
	accC1 *account.DB
	accD  *account.DB
	accD1 *account.DB
	accF  *account.DB
	accF1 *account.DB
}

//Nodes ...
var (
	Nodes = []string{
		"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
		"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR",
		"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k",
		"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs",
		"1PTGVR7TUm1MJUH7M1UNcKBGMvfJ7nCrnN",
	}
)

//NewExecCli ...
func NewExecCli() *ExecCli {
	dir, sdb, ldb := util.CreateTestDB()
	log.Println(dir)

	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")

	executor.Init(et.ExchangeX, cfg, nil)
	total := 100000000 * types.DefaultCoinPrecision
	accountA := &types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[0],
	}
	accountB := &types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[1],
	}
	accountC := &types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[2],
	}
	accountD := &types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[3],
	}
	accountFee := &types.Account{
		Balance: 0,
		Frozen:  0,
		Addr:    Nodes[4],
	}

	execAddr := address.ExecAddress(et.ExchangeX)

	accA, _ := account.NewAccountDB(cfg, "coins", "bty", sdb)
	accA.SaveExecAccount(execAddr, accountA)

	accB, _ := account.NewAccountDB(cfg, "coins", "bty", sdb)
	accB.SaveExecAccount(execAddr, accountB)

	accC, _ := account.NewAccountDB(cfg, "coins", "bty", sdb)
	accC.SaveExecAccount(execAddr, accountC)

	accD, _ := account.NewAccountDB(cfg, "coins", "bty", sdb)
	accD.SaveExecAccount(execAddr, accountD)

	accF, _ := account.NewAccountDB(cfg, "coins", "bty", sdb)
	accF.SaveExecAccount(execAddr, accountFee)

	accA1, _ := account.NewAccountDB(cfg, "token", "CCNY", sdb)
	accA1.SaveExecAccount(execAddr, accountA)

	accB1, _ := account.NewAccountDB(cfg, "token", "CCNY", sdb)
	accB1.SaveExecAccount(execAddr, accountB)

	accC1, _ := account.NewAccountDB(cfg, "token", "CCNY", sdb)
	accC1.SaveExecAccount(execAddr, accountC)

	accD1, _ := account.NewAccountDB(cfg, "token", "CCNY", sdb)
	accD1.SaveExecAccount(execAddr, accountD)

	accF1, _ := account.NewAccountDB(cfg, "token", "CCNY", sdb)
	accF1.SaveExecAccount(execAddr, accountFee)

	q := queue.New("channel")
	q.SetConfig(cfg)

	return &ExecCli{
		ldb:        ldb,
		sdb:        sdb,
		height:     1,
		blockTime:  time.Now().Unix(),
		difficulty: 1539918074,
		q:          q,
		cfg:        cfg,
		execAddr:   execAddr,

		accA:  accA,
		accA1: accA1,
		accB:  accB,
		accB1: accB1,
		accC:  accC,
		accC1: accC1,
		accD:  accD,
		accD1: accD1,
		accF:  accF,
		accF1: accF1,
	}
}

//Send ...
func (c *ExecCli) Send(tx *types.Transaction, hexKey string) ([]*types.ReceiptLog, error) {
	var err error
	tx, err = types.FormatTx(c.cfg, et.ExchangeX, tx)
	if err != nil {
		return nil, err
	}

	tx, err = signTx(tx, hexKey)
	if err != nil {
		return nil, err
	}

	api, _ := client.New(c.q.Client(), nil)
	exec := executor.NewExchange()
	exec.SetAPI(api)
	if err := exec.CheckTx(tx, int(1)); err != nil {
		return nil, err
	}

	c.height++
	c.blockTime += 10
	c.difficulty++

	exec.SetStateDB(c.sdb)
	exec.SetLocalDB(c.ldb)
	exec.SetEnv(c.height, c.blockTime, c.difficulty)

	receipt, err := exec.Exec(tx, int(1))
	if err != nil {
		return nil, err
	}

	for _, kv := range receipt.KV {
		c.sdb.Set(kv.Key, kv.Value)
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		return nil, err
	}

	for _, kv := range set.KV {
		c.ldb.Set(kv.Key, kv.Value)
	}

	//save to database
	util.SaveKVList(c.sdb, set.KV)
	return receipt.Logs, nil
}

//Query ...
func (c *ExecCli) Query(fn string, msg proto.Message) ([]byte, error) {
	api, _ := client.New(c.q.Client(), nil)
	exec := executor.NewExchange()
	exec.SetAPI(api)
	exec.SetStateDB(c.sdb)
	exec.SetLocalDB(c.ldb)
	exec.SetEnv(c.height, c.blockTime, c.difficulty)
	r, err := exec.Query(fn, types.Encode(msg))
	if err != nil {
		return nil, err
	}
	return types.Encode(r), nil
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("", signType), -1)
	if err != nil {
		return tx, err
	}

	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return tx, err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return tx, err
	}

	tx.Sign(int32(signType), privKey)
	return tx, nil
}

//GetExecAccount ...
func (c *ExecCli) GetExecAccount(addr string, exec string, symbol string) (*types.Account, error) {
	//mavl-{coins}-{bty}-exec-{26htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp}:{1JmFaA6unrCFYEWPGRi7uuXY1KthTJxJEP}
	//mavl-{token}-{ccny}-exec-{26htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp}:{1JmFaA6unrCFYEWPGRi7uuXY1KthTJxJEP}
	key := []byte(fmt.Sprintf("mavl-%s-%s-exec-%s:%s", exec, symbol, c.execAddr, addr))
	bytes, err := c.sdb.Get(key)
	if err != nil {
		return nil, err
	}

	var acc types.Account
	err = types.Decode(bytes, &acc)
	if err != nil {
		return nil, err
	}

	return &acc, nil
}
