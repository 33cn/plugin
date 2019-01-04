package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/33cn/chain33/types"
	ptypes "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
	"github.com/robertkrimen/otto"
)

type blockContext struct {
	Height     int64  `json:"height"`
	Blocktime  int64  `json:"blocktime"`
	DriverName string `json:"driverName"`
	Name       string `json:"name"`
	Curname    string `json:"curname"`
	Difficulty uint64 `json:"difficulty"`
	TxHash     string `json:"txhash"`
	Index      int64  `json:"index"`
	From       string `json:"from"`
}

func parseJsReturn(prefix []byte, jsvalue *otto.Object) (kvlist []*types.KeyValue, logs []*types.ReceiptLog, err error) {
	//kvs
	obj, err := getObject(jsvalue, "kvs")
	if err != nil {
		return nil, nil, ptypes.ErrJsReturnKVSFormat
	}
	if obj.Class() != "Array" {
		return nil, nil, ptypes.ErrJsReturnKVSFormat
	}
	size, err := getInt(obj, "length")
	if err != nil {
		return nil, nil, err
	}
	for i := 0; i < int(size); i++ {
		data, err := getObject(obj, fmt.Sprint(i))
		if err != nil {
			return nil, nil, err
		}
		kv, err := parseKV(prefix, data)
		if err != nil {
			return nil, nil, err
		}
		kvlist = append(kvlist, kv)
	}
	//logs
	obj, err = getObject(jsvalue, "logs")
	if err != nil {
		return nil, nil, ptypes.ErrJsReturnLogsFormat
	}
	if obj.Class() != "Array" {
		return nil, nil, ptypes.ErrJsReturnLogsFormat
	}
	size, err = getInt(obj, "length")
	if err != nil {
		return nil, nil, err
	}
	for i := 0; i < int(size); i++ {
		data, err := getObject(obj, fmt.Sprint(i))
		if err != nil {
			return nil, nil, err
		}
		//
		logdata, err := getString(data, "log")
		if err != nil {
			return nil, nil, err
		}
		format, err := getString(data, "format")
		if err != nil {
			return nil, nil, err
		}
		ty, err := getInt(data, "ty")
		if err != nil {
			return nil, nil, err
		}
		if format == "json" {
			l := &types.ReceiptLog{
				Ty: ptypes.TyLogJs, Log: types.Encode(&jsproto.JsLog{Data: logdata})}
			logs = append(logs, l)
		} else {
			l := &types.ReceiptLog{
				Ty:  int32(ty),
				Log: []byte(logdata),
			}
			logs = append(logs, l)
		}
	}
	return kvlist, logs, nil
}

func getString(data *otto.Object, key string) (string, error) {
	v, err := data.Get(key)
	if err != nil {
		return "", err
	}
	return v.ToString()
}

func getBool(data *otto.Object, key string) (bool, error) {
	v, err := data.Get(key)
	if err != nil {
		return false, err
	}
	return v.ToBoolean()
}

func getInt(data *otto.Object, key string) (int64, error) {
	v, err := data.Get(key)
	if err != nil {
		return 0, err
	}
	return v.ToInteger()
}

func getObject(data *otto.Object, key string) (*otto.Object, error) {
	v, err := data.Get(key)
	if err != nil {
		return nil, err
	}
	if !v.IsObject() {
		return nil, errors.New("chain33.js object get key " + key + " is not object")
	}
	return v.Object(), nil
}

func parseKV(prefix []byte, data *otto.Object) (kv *types.KeyValue, err error) {
	key, err := getString(data, "key")
	if err != nil {
		return nil, err
	}
	value, err := getString(data, "value")
	if err != nil {
		return nil, err
	}
	hasprefix, err := getBool(data, "prefix")
	if err != nil {
		return nil, err
	}
	if !hasprefix {
		key = string(prefix) + key
	}
	return &types.KeyValue{Key: []byte(key), Value: []byte(value)}, nil
}

func rewriteJSON(data []byte) ([]byte, error) {
	dat := make(map[string]interface{})
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	if err := d.Decode(&dat); err != nil {
		return nil, err
	}
	dat = rewriteString(dat)
	return json.Marshal(dat)
}

func rewriteString(dat map[string]interface{}) map[string]interface{} {
	for k, v := range dat {
		if n, ok := v.(json.Number); ok {
			dat[k] = jssafe(n)
		} else if arr, ok := v.([]interface{}); ok {
			for i := 0; i < len(arr); i++ {
				v := arr[i]
				if n, ok := v.(json.Number); ok {
					arr[i] = jssafe(n)
				}
			}
			dat[k] = arr
		} else if d, ok := v.(map[string]interface{}); ok {
			dat[k] = rewriteString(d)
		} else {
			dat[k] = v
		}
	}
	return dat
}

const maxjsint int64 = 9007199254740991

func jssafe(n json.Number) interface{} {
	if strings.Contains(string(n), ".") { //float
		return n
	}
	i, err := n.Int64()
	if err != nil {
		return n
	}
	//javascript can not parse
	if i >= maxjsint || i <= -maxjsint {
		return string(n)
	}
	return n
}
