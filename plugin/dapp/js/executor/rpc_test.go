package executor_test

import (
	"testing"

	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

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

var jscode = `
//数据结构设计
//kvlist [{key:"key1", value:"value1"},{key:"key2", value:"value2"}]
//log 设计 {json data}
function Init(context) {
    this.kvc = new kvcreator("init")
    this.context = context
    this.kvc.add("action", "init")
    this.kvc.add("context", this.context)
    return this.kvc.receipt()
}

function Exec(context) {
    this.kvc = new kvcreator("exec")
	this.context = context
}

function ExecLocal(context, logs) {
    this.kvc = new kvcreator("local")
	this.context = context
    this.logs = logs
}

function Query(context) {
	this.kvc = new kvcreator("query")
	this.context = context
}

Exec.prototype.hello = function(args) {
    this.kvc.add("args", args)
    this.kvc.add("action", "exec")
    this.kvc.add("context", this.context)
    this.kvc.addlog({"key1": "value1"})
    this.kvc.addlog({"key2": "value2"})
	return this.kvc.receipt()
}

ExecLocal.prototype.hello = function(args) {
    this.kvc.add("args", args)
    this.kvc.add("action", "execlocal")
    this.kvc.add("log", this.logs)
    this.kvc.add("context", this.context)
	return this.kvc.receipt()
}

//return a json string
Query.prototype.hello = function(args) {
	return tojson({hello:"wzw"})
}
`
