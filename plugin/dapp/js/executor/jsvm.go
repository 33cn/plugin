package executor

import (
	"errors"
	"fmt"

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
}

func parseJsReturn(jsvalue *otto.Object) (kvlist []*types.KeyValue, logs []*types.ReceiptLog, err error) {
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
		kv, err := parseKV(data)
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
		data, err := getString(obj, fmt.Sprint(i))
		if err != nil {
			return nil, nil, err
		}
		l := &types.ReceiptLog{
			Ty: ptypes.TyLogJs, Log: types.Encode(&jsproto.JsLog{Data: data})}
		logs = append(logs, l)
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

func parseKV(data *otto.Object) (kv *types.KeyValue, err error) {
	key, err := getString(data, "key")
	if err != nil {
		return nil, err
	}
	value, err := getString(data, "value")
	if err != nil {
		return nil, err
	}
	return &types.KeyValue{Key: []byte(key), Value: []byte(value)}, nil
}
