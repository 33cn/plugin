package executor

import (
	"strings"
	"testing"
	"time"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
	"github.com/robertkrimen/otto"
	"github.com/stretchr/testify/assert"
)

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

Exec.prototype.hello = function(args) {
    this.kvc.add("args", args)
    this.kvc.add("action", "exec")
    this.kvc.add("context", this.context)
    this.kvc.addlog('{"key1": "value1"}')
    this.kvc.addlog('{"key2": "value2"}')
	return this.kvc.receipt()
}

ExecLocal.prototype.hello = function(args) {
    this.kvc.add("args", args)
    this.kvc.add("action", "exec")
    this.kvc.add("log", this.logs)
    this.kvc.add("context", this.context)
	return this.kvc.receipt()
}
`

func initExec(ldb db.DB, kvdb db.KVDB, t *testing.T) *js {
	e := newjs().(*js)
	e.SetEnv(1, time.Now().Unix(), 1)
	e.SetLocalDB(kvdb)
	e.SetStateDB(kvdb)

	c, tx := createCodeTx("test", jscode)
	receipt, err := e.Exec_Create(c, tx, 0)
	assert.Nil(t, err)
	util.SaveKVList(ldb, receipt.KV)
	return e
}

func createCodeTx(name, jscode string) (*jsproto.Create, *types.Transaction) {
	data := &jsproto.Create{
		Code: jscode,
		Name: name,
	}
	return data, &types.Transaction{Execer: []byte("js"), Payload: types.Encode(data)}
}

func callCodeTx(name, f, args string) (*jsproto.Call, *types.Transaction) {
	data := &jsproto.Call{
		Funcname: f,
		Name:     name,
		Args:     args,
	}
	return data, &types.Transaction{Execer: []byte("js"), Payload: types.Encode(data)}
}

func TestCallcode(t *testing.T) {
	dir, ldb, kvdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, ldb)
	e := initExec(ldb, kvdb, t)

	call, tx := callCodeTx("test", "hello", `{"hello":"world"}`)
	receipt, err := e.Exec_Call(call, tx, 0)
	assert.Nil(t, err)
	util.SaveKVList(ldb, receipt.KV)
	util.PrintKV(receipt.KV)
}

func TestCallError(t *testing.T) {
	dir, ldb, kvdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, ldb)
	e := initExec(ldb, kvdb, t)
	//test call error(invalid json input)
	call, tx := callCodeTx("test", "hello", `{hello":"world"}`)
	_, err := e.callVM("exec", call, tx, 0, nil)
	_, ok := err.(*otto.Error)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, strings.Contains(err.Error(), "SyntaxError"))

	call, tx = callCodeTx("test", "hello", `{"hello":"world"}`)
	_, err = e.callVM("hello", call, tx, 0, nil)
	_, ok = err.(*otto.Error)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, strings.Contains(err.Error(), errInvalidFuncPrefix.Error()))

	call, tx = callCodeTx("test", "hello2", `{"hello":"world"}`)
	_, err = e.callVM("exec", call, tx, 0, nil)
	_, ok = err.(*otto.Error)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, strings.Contains(err.Error(), errFuncNotFound.Error()))
}
