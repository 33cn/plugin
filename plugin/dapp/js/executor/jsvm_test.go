package executor

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
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
	var obj = getlocaldb("context")
	return tojson(obj)
}
`

func initExec(ldb db.DB, kvdb db.KVDB, t assert.TestingT) *js {
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
	assert.Equal(t, string(receipt.KV[0].Value), `{"hello":"world"}`)
	assert.Equal(t, string(receipt.KV[1].Value), "exec")
	var data blockContext
	err = json.Unmarshal(receipt.KV[2].Value, &data)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), data.Difficulty)
	assert.Equal(t, "js", data.DriverName)
	assert.Equal(t, int64(1), data.Height)
	assert.Equal(t, int64(0), data.Index)

	kvset, err := e.ExecLocal_Call(call, tx, &types.ReceiptData{Logs: receipt.Logs}, 0)
	assert.Nil(t, err)
	util.SaveKVList(ldb, kvset.KV)
	assert.Equal(t, string(kvset.KV[0].Value), `{"hello":"world"}`)
	assert.Equal(t, string(kvset.KV[1].Value), "execlocal")
	//test log is ok
	assert.Equal(t, string(kvset.KV[2].Value), `[{"key1":"value1"},{"key2":"value2"}]`)
	//test context
	err = json.Unmarshal(kvset.KV[3].Value, &data)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), data.Difficulty)
	assert.Equal(t, "js", data.DriverName)
	assert.Equal(t, int64(1), data.Height)
	assert.Equal(t, int64(0), data.Index)

	//call query
	jsondata, err := e.Query_Query(call)
	assert.Nil(t, err)
	err = json.Unmarshal([]byte(jsondata.(*jsproto.QueryResult).Data), &data)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), data.Difficulty)
	assert.Equal(t, "js", data.DriverName)
	assert.Equal(t, int64(1), data.Height)
	assert.Equal(t, int64(0), data.Index)
	//call rollback
	kvset, err = e.ExecDelLocal_Call(call, tx, &types.ReceiptData{Logs: receipt.Logs}, 0)
	assert.Nil(t, err)
	util.SaveKVList(ldb, kvset.KV)
	assert.Equal(t, 5, len(kvset.KV))
	for i := 0; i < len(kvset.KV); i++ {
		assert.Equal(t, kvset.KV[i].Value, []byte(nil))
	}
}

func TestCallError(t *testing.T) {
	dir, ldb, kvdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, ldb)
	e := initExec(ldb, kvdb, t)
	//test call error(invalid json input)
	call, tx := callCodeTx("test", "hello", `{hello":"world"}`)
	_, err := e.callVM("exec", call, tx, 0, nil)
	_, ok := err.(*otto.Error)
	assert.Equal(t, false, ok)
	assert.Equal(t, true, strings.Contains(err.Error(), "invalid character 'h'"))

	call, tx = callCodeTx("test", "hello", `{"hello":"world"}`)
	_, err = e.callVM("hello", call, tx, 0, nil)
	_, ok = err.(*otto.Error)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, strings.Contains(err.Error(), ptypes.ErrInvalidFuncPrefix.Error()))

	call, tx = callCodeTx("test", "hello2", `{"hello":"world"}`)
	_, err = e.callVM("exec", call, tx, 0, nil)
	_, ok = err.(*otto.Error)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, strings.Contains(err.Error(), ptypes.ErrFuncNotFound.Error()))
}

//数字非常大的数字的处理
func TestBigInt(t *testing.T) {
	dir, ldb, kvdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, ldb)
	e := initExec(ldb, kvdb, t)
	//test call error(invalid json input)
	s := fmt.Sprintf(`{"balance":%d,"balance1":%d,"balance2":%d,"balance3":%d}`, math.MaxInt64, math.MinInt64, 9007199254740990, -9007199254740990)
	call, tx := callCodeTx("test", "hello", s)
	data, err := e.callVM("exec", call, tx, 0, nil)
	kvs, _, err := parseJsReturn(data)
	assert.Nil(t, err)
	assert.Equal(t, `{"balance":"9223372036854775807","balance1":"-9223372036854775808","balance2":9007199254740990,"balance3":-9007199254740990}`, string(kvs[0].Value))
}

func BenchmarkBigInt(b *testing.B) {
	dir, ldb, kvdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, ldb)
	e := initExec(ldb, kvdb, b)
	//test call error(invalid json input)
	s := fmt.Sprintf(`{"balance":%d,"balance1":%d,"balance2":%d,"balance3":%d}`, math.MaxInt64, math.MinInt64, 9007199254740990, -9007199254740990)
	call, tx := callCodeTx("test", "hello", s)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.callVM("exec", call, tx, 0, nil)
		assert.Nil(b, err)
	}
}

func TestRewriteJSON(t *testing.T) {
	s := fmt.Sprintf(`{"balance":%d,"balance1":%d,"balance2":%d,"balance3":%d}`, math.MaxInt64, math.MinInt64, 9007199254740990, -9007199254740990)
	quota := fmt.Sprintf(`{"balance":"%d","balance1":"%d","balance2":%d,"balance3":%d}`, math.MaxInt64, math.MinInt64, 9007199254740990, -9007199254740990)
	data, err := rewriteJSON([]byte(s))
	assert.Nil(t, err)
	assert.Equal(t, quota, string(data))
	data2 := make(map[string]interface{})
	data2["ints"] = []int64{math.MaxInt64, math.MinInt64, 9007199254740990, -9007199254740990, 1, 0}
	data2["float"] = []float64{1.1, 1000000000000000000000000000, 10000000000000000}
	json1, err := json.Marshal(data2)
	assert.Nil(t, err)
	//assert.Equal(t, `{"float":[1.1,1100000000000000000000,-1100000000000000000000],"ints":[9223372036854775807,-9223372036854775808,9007199254740990,-9007199254740990,1,0]}`, string(json1))
	json2, err := rewriteJSON(json1)
	assert.Nil(t, err)
	assert.Equal(t, string(json2), `{"float":[1.1,1e+27,"10000000000000000"],"ints":["9223372036854775807","-9223372036854775808",9007199254740990,-9007199254740990,1,0]}`)
}

func TestCalcLocalPrefix(t *testing.T) {
	assert.Equal(t, calcLocalPrefix([]byte("a")), []byte("LODB-a-"))
	assert.Equal(t, calcStatePrefix([]byte("a")), []byte("mavl-a-"))
	assert.Equal(t, calcCodeKey("a"), []byte("mavl-js-code-a"))
	assert.Equal(t, calcRollbackKey([]byte("a")), []byte("LODB-js-rollback-a"))
}
