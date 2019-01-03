package executor

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
	"github.com/robertkrimen/otto"
)

//json 表格相关的实现
//用户可以存入一个json字符串
//json 定义index 以及 format
/* table ->
{
	"#tablename" : "table1",
	"#primary" : "abc",
	"abc" :  "%18d",
	"index1" : "%s",
	"index2" : "%s",
}

默认值配置
{
	"abc" : 0,
	"index1" : "",
	"index2" : "",
}
*/
var globalTableHandle sync.Map
var globalHanldeID int64

//NewTable 创建一个新的表格, 返回handle
func (u *js) newTable(name, config, defaultvalue string) (id int64, err error) {
	for {
		id = atomic.AddInt64(&globalHanldeID, 1) % maxjsint
		if _, ok := globalTableHandle.Load(id); ok {
			continue
		}
		if id < 0 {
			atomic.StoreInt64(&globalHanldeID, 0)
			continue
		}
		break
	}
	row, err := NewJSONRow(config, defaultvalue)
	if err != nil {
		return 0, err
	}
	var kvdb db.KV
	var prefix []byte
	if row.config["#db"] == "localdb" {
		kvdb = u.GetLocalDB()
		_, prefix = calcAllPrefix(name)
	} else if row.config["#db"] == "statedb" {
		kvdb = u.GetStateDB()
		prefix, _ = calcAllPrefix(name)
	} else {
		return 0, ptypes.ErrDBType
	}
	var indexs []string
	for k := range row.config {
		if k[0] == '#' {
			continue
		}
		indexs = append(indexs, k)
	}
	opt := &table.Option{
		Prefix:  string(prefix),
		Name:    row.config["#tablename"],
		Primary: row.config["#primary"],
		Index:   indexs,
	}
	t, err := table.NewTable(row, kvdb, opt)
	if err != nil {
		return 0, err
	}
	globalTableHandle.Store(id, t)
	return id, nil
}

func (u *js) newTableFunc(vm *otto.Otto, name string) {
	vm.Set("table_new", func(call otto.FunctionCall) otto.Value {
		config, err := call.Argument(0).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		defaultvalue, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		id, err := u.newTable(name, config, defaultvalue)
		if err != nil {
			return errReturn(vm, err)
		}
		return newObject(vm).setValue("id", id).value()
	})
}

//CloseTable 关闭表格释放内存
func closeTable(id int64) error {
	_, ok := globalTableHandle.Load(id)
	if !ok {
		return types.ErrNotFound
	}
	globalTableHandle.Delete(id)
	return nil
}

func getTable(id int64) (*table.Table, error) {
	if value, ok := globalTableHandle.Load(id); ok {
		return value.(*table.Table), nil
	}
	return nil, types.ErrNotFound
}

func getSaver(id int64) (saver, error) {
	if value, ok := globalTableHandle.Load(id); ok {
		return value.(saver), nil
	}
	return nil, types.ErrNotFound
}

func registerTableFunc(vm *otto.Otto) {
	tableAddFunc(vm)
	tableReplaceFunc(vm)
	tableDelFunc(vm)
	tableCloseFunc(vm)
	tableSave(vm)
	tableJoinFunc(vm)
}

func tableAddFunc(vm *otto.Otto) {
	vm.Set("table_add", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := getTable(id)
		if err != nil {
			return errReturn(vm, err)
		}
		json, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		err = tab.Add(&jsproto.JsLog{Data: json})
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, "ok")
	})
}

func tableReplaceFunc(vm *otto.Otto) {
	vm.Set("table_replace", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := getTable(id)
		if err != nil {
			return errReturn(vm, err)
		}
		json, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		err = tab.Replace(&jsproto.JsLog{Data: json})
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, "ok")
	})
}

func tableDelFunc(vm *otto.Otto) {
	vm.Set("table_del", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := getTable(id)
		if err != nil {
			return errReturn(vm, err)
		}
		row, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		err = tab.DelRow(&jsproto.JsLog{Data: row})
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, "ok")
	})
}

type saver interface {
	Save() (kvs []*types.KeyValue, err error)
}

func tableSave(vm *otto.Otto) {
	vm.Set("table_save", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := getSaver(id)
		if err != nil {
			return errReturn(vm, err)
		}
		kvs, err := tab.Save()
		if err != nil {
			return errReturn(vm, err)
		}
		return newObject(vm).setValue("kvs", createKVObject(vm, kvs)).value()
	})
}

func tableCloseFunc(vm *otto.Otto) {
	vm.Set("table_close", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		err = closeTable(id)
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, "ok")
	})
}

func tableJoinFunc(vm *otto.Otto) {
	vm.Set("new_join_table", func(call otto.FunctionCall) otto.Value {
		left, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		lefttab, err := getTable(left)
		if err != nil {
			return errReturn(vm, err)
		}
		right, err := call.Argument(1).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		righttab, err := getTable(right)
		if err != nil {
			return errReturn(vm, err)
		}
		index, err := call.Argument(2).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		join, err := table.NewJoinTable(lefttab, righttab, strings.Split(index, ","))
		if err != nil {
			return errReturn(vm, err)
		}
		var id int64
		for {
			id = atomic.AddInt64(&globalHanldeID, 1) % maxjsint
			if _, ok := globalTableHandle.Load(id); ok {
				continue
			}
			if id < 0 {
				atomic.StoreInt64(&globalHanldeID, 0)
				continue
			}
			break
		}
		globalTableHandle.Store(id, join)
		return newObject(vm).setValue("id", id).value()
	})
}

/*
table
要开发一个适合json的table, row 就是一个 js object
handle := new_table(config, defaultvalue)
table_add(handle, row)
table_replace(handle, row)
table_del(handle, row)
table_save(handle)
table_close(handle)
handle := new_join_table(left, right, listofjoinindex)
*/
//join table 的操作(接口完全相同)
//handle3 := new_table(newcofifg{config1, config2})

//JSONRow meta
type JSONRow struct {
	*jsproto.JsLog
	config map[string]string
	data   map[string]interface{}
}

//NewJSONRow 创建一个row
func NewJSONRow(config string, defaultvalue string) (*JSONRow, error) {
	row := &JSONRow{}
	row.config = make(map[string]string)
	err := json.Unmarshal([]byte(config), row.config)
	if err != nil {
		return nil, err
	}
	row.JsLog = &jsproto.JsLog{Data: defaultvalue}
	err = row.parse()
	if err != nil {
		return nil, err
	}
	return row, nil
}

//CreateRow 创建一行
func (row *JSONRow) CreateRow() *table.Row {
	return &table.Row{Data: &jsproto.JsLog{}}
}

func (row *JSONRow) parse() error {
	row.data = make(map[string]interface{})
	return json.Unmarshal([]byte(row.JsLog.Data), row.data)
}

//SetPayload 设置行的内容
func (row *JSONRow) SetPayload(data types.Message) error {
	if rowdata, ok := data.(*jsproto.JsLog); ok {
		row.JsLog = rowdata
		return row.parse()
	}
	return types.ErrTypeAsset
}

//Get value of row
func (row *JSONRow) Get(key string) ([]byte, error) {
	if format, ok := row.config[key]; ok {
		if data, ok := row.data[key]; ok {
			return []byte(fmt.Sprintf(format, data)), nil
		}
	}
	return nil, types.ErrNotFound
}
