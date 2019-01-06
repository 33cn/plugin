package executor_test

import (
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
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
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
	//2.2 调用 hello 函数(随机数，用nonce)
	privhash := common.Sha256(mocker.GetHotKey().Bytes())
	nonce := rand.Int63()
	num := rand.Int63() % 10
	realhash := common.Sha256([]byte(string(privhash) + ":" + fmt.Sprint(nonce)))
	myhash := common.ToHex(common.Sha256([]byte(string(realhash) + fmt.Sprint(num))))

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
	t.Log(mocker.GetHotAddress())
	hash, err = mocker.SendAndSignNonce(mocker.GetHotKey(), txhex, nonce)
	assert.Nil(t, err)
	txinfo, err = mocker.WaitTx(hash)
	assert.Nil(t, err)
	assert.Equal(t, txinfo.Receipt.Ty, int32(2))

	//3. query 函数查询
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
}
