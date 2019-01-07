package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

//NewTable 创建一个新的表格, 返回handle
func (u *js) newTable(name, config, defaultvalue string) (id int64, err error) {
	u.globalHanldeID++
	id = u.globalHanldeID
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
		Prefix:  strings.Trim(string(prefix), "-"),
		Name:    row.config["#tablename"],
		Primary: row.config["#primary"],
		Index:   indexs,
	}
	t, err := table.NewTable(row, kvdb, opt)
	if err != nil {
		return 0, err
	}
	u.globalTableHandle.Store(id, t)
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
func (u *js) closeTable(id int64) error {
	_, ok := u.globalTableHandle.Load(id)
	if !ok {
		return types.ErrNotFound
	}
	u.globalTableHandle.Delete(id)
	return nil
}

func (u *js) getTable(id int64) (*table.Table, error) {
	if value, ok := u.globalTableHandle.Load(id); ok {
		return value.(*table.Table), nil
	}
	return nil, types.ErrNotFound
}

func (u *js) getTabler(id int64) (tabler, error) {
	if value, ok := u.globalTableHandle.Load(id); ok {
		return value.(tabler), nil
	}
	return nil, types.ErrNotFound
}

func (u *js) registerTableFunc(vm *otto.Otto, name string) {
	u.newTableFunc(vm, name)
	u.tableAddFunc(vm)
	u.tableReplaceFunc(vm)
	u.tableDelFunc(vm)
	u.tableCloseFunc(vm)
	u.tableSave(vm)
	u.tableJoinFunc(vm)
	u.tableQueryFunc(vm)
	u.tableGetFunc(vm)
	u.tableJoinKeyFunc(vm)
}

func (u *js) tableJoinKeyFunc(vm *otto.Otto) {
	vm.Set("table_joinkey", func(call otto.FunctionCall) otto.Value {
		left, err := call.Argument(0).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		right, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		key := table.JoinKey([]byte(left), []byte(right))
		return okReturn(vm, string(key))
	})
}

func (u *js) tableAddFunc(vm *otto.Otto) {
	vm.Set("table_add", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := u.getTable(id)
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

func (u *js) tableReplaceFunc(vm *otto.Otto) {
	vm.Set("table_replace", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := u.getTable(id)
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

func (u *js) tableDelFunc(vm *otto.Otto) {
	vm.Set("table_del", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := u.getTable(id)
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

func (u *js) tableGetFunc(vm *otto.Otto) {
	vm.Set("table_get", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := u.getTable(id)
		if err != nil {
			return errReturn(vm, err)
		}
		key, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		row, err := call.Argument(2).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		meta := tab.GetMeta()
		meta.SetPayload(&jsproto.JsLog{Data: row})
		result, err := meta.Get(key)
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, string(result))
	})
}

type tabler interface {
	GetMeta() table.RowMeta
	ListIndex(indexName string, prefix []byte, primaryKey []byte, count, direction int32) (rows []*table.Row, err error)
	Save() (kvs []*types.KeyValue, err error)
}

func (u *js) tableSave(vm *otto.Otto) {
	vm.Set("table_save", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := u.getTabler(id)
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

func (u *js) tableCloseFunc(vm *otto.Otto) {
	vm.Set("table_close", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		err = u.closeTable(id)
		if err != nil {
			return errReturn(vm, err)
		}
		return okReturn(vm, "ok")
	})
}

func (u *js) tableQueryFunc(vm *otto.Otto) {
	vm.Set("table_query", func(call otto.FunctionCall) otto.Value {
		id, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		tab, err := u.getTabler(id)
		if err != nil {
			return errReturn(vm, err)
		}
		//参数
		//List(indexName string, data types.Message, primaryKey []byte, count, direction int32) (rows []*Row, err error)
		indexName, err := call.Argument(1).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		prefix, err := call.Argument(2).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		primaryKey, err := call.Argument(3).ToString()
		if err != nil {
			return errReturn(vm, err)
		}
		count, err := call.Argument(4).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		direction, err := call.Argument(5).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		bprefix := []byte(prefix)
		if prefix == "" {
			bprefix = nil
		}
		bprimaryKey := []byte(primaryKey)
		if primaryKey == "" {
			bprimaryKey = nil
		}
		rows, err := tab.ListIndex(indexName, bprefix, bprimaryKey, int32(count), int32(direction))
		if err != nil {
			return errReturn(vm, err)
		}
		_, isjoin := tab.(*table.JoinTable)
		querylist := make([]interface{}, len(rows))
		for i := 0; i < len(rows); i++ {
			if isjoin {
				joindata, ok := rows[i].Data.(*table.JoinData)
				if !ok {
					return errReturn(vm, errors.New("jointable has no joindata"))
				}
				leftdata, ok := joindata.Left.(*jsproto.JsLog)
				if !ok {
					return errReturn(vm, errors.New("leftdata is not JsLog"))
				}
				rightdata, ok := joindata.Right.(*jsproto.JsLog)
				if !ok {
					return errReturn(vm, errors.New("rightdata is not jslog"))
				}
				obj := make(map[string]interface{})
				obj["left"] = leftdata.Data
				obj["right"] = rightdata.Data
				querylist[i] = obj
			} else {
				leftdata, ok := rows[i].Data.(*jsproto.JsLog)
				if !ok {
					return errReturn(vm, errors.New("data is not JsLog"))
				}
				obj := make(map[string]interface{})
				obj["left"] = leftdata.Data
				querylist[i] = obj
			}
		}
		retvalue, err := vm.ToValue(querylist)
		if err != nil {
			return errReturn(vm, err)
		}
		return retvalue
	})
}

func (u *js) tableJoinFunc(vm *otto.Otto) {
	vm.Set("new_join_table", func(call otto.FunctionCall) otto.Value {
		left, err := call.Argument(0).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		lefttab, err := u.getTable(left)
		if err != nil {
			return errReturn(vm, err)
		}
		right, err := call.Argument(1).ToInteger()
		if err != nil {
			return errReturn(vm, err)
		}
		righttab, err := u.getTable(right)
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
		u.globalHanldeID++
		id = u.globalHanldeID
		u.globalTableHandle.Store(id, join)
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
	config       map[string]string
	data         map[string]interface{}
	lastdata     types.Message
	isint        *regexp.Regexp
	isfloat      *regexp.Regexp
	defaultvalue string
}

//NewJSONRow 创建一个row
func NewJSONRow(config string, defaultvalue string) (*JSONRow, error) {
	row := &JSONRow{}
	row.isint = regexp.MustCompile(`%\d*d`)
	row.isfloat = regexp.MustCompile(`%[\.\d]*f`)
	row.config = make(map[string]string)
	err := json.Unmarshal([]byte(config), &row.config)
	if err != nil {
		return nil, err
	}
	row.defaultvalue = defaultvalue
	row.JsLog = &jsproto.JsLog{Data: defaultvalue}
	err = row.parse()
	if err != nil {
		return nil, err
	}
	return row, nil
}

//CreateRow 创建一行
func (row *JSONRow) CreateRow() *table.Row {
	return &table.Row{Data: &jsproto.JsLog{Data: row.defaultvalue}}
}

func (row *JSONRow) parse() error {
	row.data = make(map[string]interface{})
	d := json.NewDecoder(bytes.NewBufferString(row.JsLog.Data))
	d.UseNumber()
	return d.Decode(&row.data)
}

//SetPayload 设置行的内容
func (row *JSONRow) SetPayload(data types.Message) error {
	if row.lastdata == data {
		return nil
	}
	row.lastdata = data
	if rowdata, ok := data.(*jsproto.JsLog); ok {
		row.JsLog = rowdata
		return row.parse()
	}
	return types.ErrTypeAsset
}

//Get value of row
func (row *JSONRow) Get(key string) ([]byte, error) {
	v, err := row.get(key)
	return v, err
}

func (row *JSONRow) get(key string) ([]byte, error) {
	if format, ok := row.config[key]; ok {
		if data, ok := row.data[key]; ok {
			if row.isint.Match([]byte(format)) { //int
				s := fmt.Sprint(data)
				num, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return []byte(fmt.Sprintf(format, num)), nil
			} else if row.isfloat.Match([]byte(format)) { //float
				s := fmt.Sprint(data)
				num, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return nil, err
				}
				return []byte(fmt.Sprintf(format, num)), nil
			} else { //string
				if n, ok := data.(json.Number); ok {
					data = n.String()
				}
			}
			return []byte(fmt.Sprintf(format, data)), nil
		}
	}
	return nil, errors.New("get key " + key + "from data err")
}
