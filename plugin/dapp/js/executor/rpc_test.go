package executor_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestJsVM(t *testing.T) {
	cfg, sub := testnode.GetDefaultConfig()
	cfg.Consensus.Name = "ticket"
	mocker := testnode.NewWithConfig(cfg, sub, nil)
	defer mocker.Close()
	mocker.Listen()

	configCreator(mocker, t)
	//开始部署合约, 测试阶段任何人都可以部署合约
	//后期需要加上权限控制
	//1. 部署合约
	create := &jsproto.Create{
		Code: jscode,
		Name: "test",
	}
	req := &rpctypes.CreateTxIn{
		Execer:     ptypes.JsX,
		ActionName: "Create",
		Payload:    types.MustPBToJSON(create),
	}
	var txhex string
	err := mocker.GetJSONC().Call("Chain33.CreateTransaction", req, &txhex)
	assert.Nil(t, err)
	hash, err := mocker.SendAndSign(mocker.GetHotKey(), txhex)
	assert.Nil(t, err)
	txinfo, err := mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))

	//2. 调用 hello 函数
	call := &jsproto.Call{
		Funcname: "hello",
		Name:     "test",
		Args:     "{}",
	}
	req = &rpctypes.CreateTxIn{
		Execer:     "user." + ptypes.JsX + ".test",
		ActionName: "Call",
		Payload:    types.MustPBToJSON(call),
	}
	err = mocker.GetJSONC().Call("Chain33.CreateTransaction", req, &txhex)
	assert.Nil(t, err)
	hash, err = mocker.SendAndSign(mocker.GetHotKey(), txhex)
	assert.Nil(t, err)
	txinfo, err = mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))

	//3. query 函数查询
	call = &jsproto.Call{
		Funcname: "hello",
		Name:     "test",
		Args:     "{}",
	}
	query := &rpctypes.Query4Jrpc{
		Execer:   "user." + ptypes.JsX + ".test",
		FuncName: "Query",
		Payload:  types.MustPBToJSON(call),
	}
	var queryresult jsproto.QueryResult
	err = mocker.GetJSONC().Call("Chain33.Query", query, &queryresult)
	assert.Nil(t, err)
	t.Log(queryresult.Data)
}

func TestJsGame(t *testing.T) {
	contractName := "test1"
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	err := mocker.SendHot()
	assert.Nil(t, err)
	// 需要配置
	configCreator(mocker, t)

	//开始部署合约, 测试阶段任何人都可以部署合约
	//后期需要加上权限控制
	//1. 部署合约
	create := &jsproto.Create{
		Code: gamecode,
		Name: contractName,
	}
	req := &rpctypes.CreateTxIn{
		Execer:     ptypes.JsX,
		ActionName: "Create",
		Payload:    types.MustPBToJSON(create),
	}
	var txhex string
	err = mocker.GetJSONC().Call("Chain33.CreateTransaction", req, &txhex)
	assert.Nil(t, err)
	hash, err := mocker.SendAndSign(mocker.GetHotKey(), txhex)
	assert.Nil(t, err)
	txinfo, err := mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))
	block := mocker.GetLastBlock()
	balance := mocker.GetAccount(block.StateHash, mocker.GetHotAddress()).Balance
	assert.Equal(t, balance, 10000*types.Coin)
	//2.1 充值到合约
	reqtx := &rpctypes.CreateTx{
		To:          address.ExecAddress("user.jsvm." + contractName),
		Amount:      100 * types.Coin,
		Note:        "12312",
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    "user.jsvm." + contractName,
	}
	err = mocker.GetJSONC().Call("Chain33.CreateRawTransaction", reqtx, &txhex)
	assert.Nil(t, err)
	hash, err = mocker.SendAndSign(mocker.GetHotKey(), txhex)
	assert.Nil(t, err)
	txinfo, err = mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))
	block = mocker.GetLastBlock()
	balance = mocker.GetExecAccount(block.StateHash, "user.jsvm."+contractName, mocker.GetHotAddress()).Balance
	assert.Equal(t, 100*types.Coin, balance)

	reqtx = &rpctypes.CreateTx{
		To:          address.ExecAddress("user.jsvm." + contractName),
		Amount:      100 * types.Coin,
		Note:        "12312",
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    "user.jsvm." + contractName,
	}
	err = mocker.GetJSONC().Call("Chain33.CreateRawTransaction", reqtx, &txhex)
	assert.Nil(t, err)
	hash, err = mocker.SendAndSign(mocker.GetGenesisKey(), txhex)
	assert.Nil(t, err)
	txinfo, err = mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))
	block = mocker.GetLastBlock()
	balance = mocker.GetExecAccount(block.StateHash, "user.jsvm."+contractName, mocker.GetGenesisAddress()).Balance
	assert.Equal(t, 100*types.Coin, balance)
	t.Log(mocker.GetGenesisAddress())
	//2.2 调用 hello 函数(随机数，用nonce)
	privhash := common.Sha256(mocker.GetHotKey().Bytes())
	nonce := rand.Int63()
	num := rand.Int63() % 10
	realhash := common.ToHex(common.Sha256([]byte(string(privhash) + ":" + fmt.Sprint(nonce))))
	myhash := common.ToHex(common.Sha256([]byte(realhash + fmt.Sprint(num))))

	call := &jsproto.Call{
		Funcname: "NewGame",
		Name:     contractName,
		Args:     fmt.Sprintf(`{"bet": %d, "randhash" : "%s"}`, 100*types.Coin, myhash),
	}
	req = &rpctypes.CreateTxIn{
		Execer:     "user." + ptypes.JsX + "." + contractName,
		ActionName: "Call",
		Payload:    types.MustPBToJSON(call),
	}
	err = mocker.GetJSONC().Call("Chain33.CreateTransaction", req, &txhex)
	assert.Nil(t, err)
	hash, err = mocker.SendAndSignNonce(mocker.GetHotKey(), txhex, nonce)
	assert.Nil(t, err)
	txinfo, err = mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))
	gameid := txinfo.Height*100000 + txinfo.Index
	//2.3 guess a number (win)
	call = &jsproto.Call{
		Funcname: "Guess",
		Name:     contractName,
		Args:     fmt.Sprintf(`{"bet": %d, "gameid" : "%d", "num" : %d}`, 1*types.Coin, gameid, num),
	}
	req = &rpctypes.CreateTxIn{
		Execer:     "user." + ptypes.JsX + "." + contractName,
		ActionName: "Call",
		Payload:    types.MustPBToJSON(call),
	}
	err = mocker.GetJSONC().Call("Chain33.CreateTransaction", req, &txhex)
	assert.Nil(t, err)
	hash, err = mocker.SendAndSignNonce(mocker.GetGenesisKey(), txhex, nonce)
	assert.Nil(t, err)
	txinfo, err = mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))

	//2.4 guess a num (failed)
	call = &jsproto.Call{
		Funcname: "Guess",
		Name:     contractName,
		Args:     fmt.Sprintf(`{"bet": %d, "gameid" : "%d", "num" : %d}`, 1*types.Coin, gameid, num+1),
	}
	req = &rpctypes.CreateTxIn{
		Execer:     "user." + ptypes.JsX + "." + contractName,
		ActionName: "Call",
		Payload:    types.MustPBToJSON(call),
	}
	err = mocker.GetJSONC().Call("Chain33.CreateTransaction", req, &txhex)
	assert.Nil(t, err)
	t.Log(mocker.GetHotAddress())
	hash, err = mocker.SendAndSignNonce(mocker.GetGenesisKey(), txhex, nonce)
	assert.Nil(t, err)
	txinfo, err = mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))

	//2.5 close the game
	call = &jsproto.Call{
		Funcname: "CloseGame",
		Name:     contractName,
		Args:     fmt.Sprintf(`{"gameid":%d, "randstr":"%s"}`, gameid, realhash),
	}
	req = &rpctypes.CreateTxIn{
		Execer:     "user." + ptypes.JsX + "." + contractName,
		ActionName: "Call",
		Payload:    types.MustPBToJSON(call),
	}
	err = mocker.GetJSONC().Call("Chain33.CreateTransaction", req, &txhex)
	assert.Nil(t, err)
	t.Log(mocker.GetHotAddress())
	hash, err = mocker.SendAndSignNonce(mocker.GetHotKey(), txhex, nonce)
	assert.Nil(t, err)
	txinfo, err = mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))
	//3.1 query game 函数查询
	call = &jsproto.Call{
		Funcname: "ListGameByAddr",
		Name:     contractName,
		Args:     fmt.Sprintf(`{"addr":"%s", "count" : 20}`, txinfo.Tx.From),
	}
	query := &rpctypes.Query4Jrpc{
		Execer:   "user." + ptypes.JsX + "." + contractName,
		FuncName: "Query",
		Payload:  types.MustPBToJSON(call),
	}
	var queryresult jsproto.QueryResult
	err = mocker.GetJSONC().Call("Chain33.Query", query, &queryresult)
	assert.Nil(t, err)
	t.Log(queryresult.Data)

	//3.2 query match -> status 函数
	call = &jsproto.Call{
		Funcname: "JoinKey",
		Name:     contractName,
		Args:     fmt.Sprintf(`{"left":"%s", "right" : "%s"}`, mocker.GetGenesisAddress(), "2"),
	}
	query = &rpctypes.Query4Jrpc{
		Execer:   "user." + ptypes.JsX + "." + contractName,
		FuncName: "Query",
		Payload:  types.MustPBToJSON(call),
	}
	err = mocker.GetJSONC().Call("Chain33.Query", query, &queryresult)
	assert.Nil(t, err)
	joinkey := queryresult.Data
	reqjson := make(map[string]interface{})
	reqjson["addr#status"] = joinkey
	reqdata, _ := json.Marshal(reqjson)
	call = &jsproto.Call{
		Funcname: "ListMatchByAddr",
		Name:     contractName,
		Args:     string(reqdata),
	}
	query = &rpctypes.Query4Jrpc{
		Execer:   "user." + ptypes.JsX + "." + contractName,
		FuncName: "Query",
		Payload:  types.MustPBToJSON(call),
	}
	err = mocker.GetJSONC().Call("Chain33.Query", query, &queryresult)
	assert.Nil(t, err)
	t.Log(queryresult.Data)
}

func configCreator(mocker *testnode.Chain33Mock, t *testing.T) {
	// 需要配置
	addr := address.PubKeyToAddress(mocker.GetHotKey().PubKey().Bytes()).String()
	creator := &types.ModifyConfig{
		Key:   "js-creator",
		Op:    "add",
		Value: addr,
		Addr:  addr,
	}
	cfgReq := &rpctypes.CreateTxIn{
		Execer:     "manage",
		ActionName: "Modify",
		Payload:    types.MustPBToJSON(creator),
	}
	var cfgtxhex string
	err := mocker.GetJSONC().Call("Chain33.CreateTransaction", cfgReq, &cfgtxhex)
	assert.Nil(t, err)
	hash1, err := mocker.SendAndSign(mocker.GetHotKey(), cfgtxhex)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(hash1)
	assert.Nil(t, err)
}
