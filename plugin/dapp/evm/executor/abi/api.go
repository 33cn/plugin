package abi

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/golang-collections/collections/stack"
)

// Pack 使用ABI方式调用时，将调用方式转换为EVM底层处理的十六进制编码
// abiData 完整的ABI定义
// param 调用方法及参数
// readOnly 是否只读，如果调用的方法不为只读，则报错
// 调用方式： foo(param1,param2)
func Pack(param, abiData string, readOnly bool) (methodName string, packData []byte, err error) {
	// 首先解析参数字符串，分析出方法名以及个参数取值
	methodName, params, err := ProcFuncCall(param)

	if err != nil {
		return methodName, packData, err
	}

	// 解析ABI数据结构，获取本次调用的方法对象
	abi, err := JSON(strings.NewReader(abiData))
	if err != nil {
		return methodName, packData, err
	}

	var method Method
	var ok bool
	if method, ok = abi.Methods[methodName]; !ok {
		err = fmt.Errorf("function %v not exists", methodName)
		return methodName, packData, err
	}

	if readOnly && !method.IsConstant() {
		return methodName, packData, errors.New("method is not readonly")
	}
	if len(params) != method.Inputs.LengthNonIndexed() {
		err = fmt.Errorf("function params error:%v", params)
		return methodName, packData, err
	}
	// 获取方法参数对象，遍历解析各参数，获得参数的Go取值
	paramVals := []interface{}{}
	if len(params) != 0 {
		// 首先检查参数个数和ABI中定义的是否一致
		if method.Inputs.LengthNonIndexed() != len(params) {
			err = fmt.Errorf("function Params count error: %v", param)
			return methodName, packData, err
		}

		for i, v := range method.Inputs.NonIndexed() {
			paramVal, err := Str2GoValue(v.Type, params[i])
			if err != nil {
				return methodName, packData, err
			}
			paramVals = append(paramVals, paramVal)
		}
	}

	// 使用Abi对象将方法和参数进行打包
	packData, err = abi.Pack(methodName, paramVals...)
	return methodName, packData, err
}

func PackContructorPara(param, abiStr string) (packData []byte, err error) {
	_, params, err := ProcFuncCall(param)
	if err != nil {
		return nil, err
	}

	parsedAbi, err := JSON(strings.NewReader(abiStr))
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse evm code error", err)
		return
	}

	method := parsedAbi.Constructor

	paramVals := []interface{}{}
	if len(params) != 0 {
		// 首先检查参数个数和ABI中定义的是否一致
		if method.Inputs.LengthNonIndexed() != len(params) {
			err = fmt.Errorf("function Params count error: %v", param)
			return nil, err
		}

		for i, v := range method.Inputs.NonIndexed() {
			paramVal, err := Str2GoValue(v.Type, params[i])
			if err != nil {
				return nil, err
			}
			paramVals = append(paramVals, paramVal)
		}
	}
	packData, err = parsedAbi.Constructor.Inputs.Pack(paramVals...)
	if err != nil {
		return nil, err
	}
	return packData, nil

}

// Unpack 将调用返回结果按照ABI的格式序列化为json
// data 合约方法返回值
// abiData 完整的ABI定义
func Unpack(data []byte, methodName, abiData string) (output []*Param, err error) {
	if len(data) == 0 {
		log.Info("Unpack", "Data len", 0, "methodName", methodName)
		return output, err
	}

	// 解析ABI数据结构，获取本次调用的方法对象
	abi, err := JSON(strings.NewReader(abiData))
	if err != nil {
		return output, err
	}

	var method Method
	var ok bool
	if method, ok = abi.Methods[methodName]; !ok {
		return output, fmt.Errorf("function %v not exists", methodName)
	}

	if method.Outputs.LengthNonIndexed() == 0 {
		return output, err
	}

	values, err := method.Outputs.UnpackValues(data)
	if err != nil {
		return output, err
	}
	output = []*Param{}

	for i, v := range values {
		arg := method.Outputs[i]
		pval := &Param{Name: arg.Name, Type: arg.Type.String(), Value: v}
		if arg.Type.String() == "address" {
			pval.Value = v.(common.Hash160Address).ToAddress().String()
			log.Info("Unpack address", "address", pval.Value)
		}
		output = append(output, pval)
	}

	return
}

//同时支持input,output和event三种数据的unpack
func UnpackAllTypes(data []byte, name, abiData string) (output []*Param, err error) {
	if len(data) == 0 {
		log.Info("Unpack", "Data len", 0, "name", name)
		return output, err
	}
	// 解析ABI数据结构，获取本次调用的方法对象
	abi, err := JSON(strings.NewReader(abiData))
	if err != nil {
		return output, err
	}

	var values []interface{}
	var arguments Arguments
	if method, ok := abi.Methods[name]; ok {
		if len(data)%32 == 0 {
			values, err = method.Outputs.UnpackValues(data)
			arguments = method.Outputs
		} else if len(data[4:])%32 == 0 {
			values, err = method.Inputs.UnpackValues(data[4:])
			arguments = method.Inputs
		} else {
			return output, errors.New("UnpackAllTypes: improperly formatted data")
		}
		if arguments.LengthNonIndexed() == 0 {
			return output, nil
		}
	} else if event, ok := abi.Events[name]; ok {
		values, err = event.Inputs.UnpackValues(data)
		arguments = event.Inputs
	} else {
		return output, errors.New("UnpackAllTypes: could not locate named method or event")
	}

	if err != nil {
		return output, err
	}

	if len(arguments) == 0 || len(values) != len(arguments) {
		return output, errors.New("wrong data to unpack")
	}

	output = []*Param{}
	for i, v := range values {
		arg := arguments[i]
		pval := &Param{Name: arg.Name, Type: arg.Type.String(), Value: v}
		if arg.Type.String() == "address" {
			pval.Value = v.(common.Hash160Address).ToAddress().String()
			log.Info("Unpack address", "address", pval.Value)
		}

		output = append(output, pval)
	}

	return
}

// Param 返回值参数结构定义
type Param struct {
	// Name 参数名称
	Name string `json:"name"`
	// Type 参数类型
	Type string `json:"type"`
	// Value 参数取值
	Value interface{} `json:"value"`
}

func convertUint(val uint64, kind reflect.Kind) interface{} {
	switch kind {
	case reflect.Uint:
		return uint(val)
	case reflect.Uint8:
		return uint8(val)
	case reflect.Uint16:
		return uint16(val)
	case reflect.Uint32:
		return uint32(val)
	case reflect.Uint64:
		return val
	}
	return val
}

func convertInt(val int64, kind reflect.Kind) interface{} {
	switch kind {
	case reflect.Int:
		return int(val)
	case reflect.Int8:
		return int8(val)
	case reflect.Int16:
		return int16(val)
	case reflect.Int32:
		return int32(val)
	case reflect.Int64:
		return val
	}
	return val
}

// 从字符串格式的输入参数取值（单个），获取Go类型的
func Str2GoValue(typ Type, val string) (res interface{}, err error) {
	switch typ.T {
	case IntTy:
		if typ.Size < 256 {
			x, err := strconv.ParseInt(val, 10, typ.Size)
			if err != nil {
				return res, err
			}
			return convertInt(x, typ.GetType().Kind()), nil
		}
		b := new(big.Int)
		b.SetString(val, 10)
		return b, err
	case UintTy:
		if typ.Size < 256 {
			x, err := strconv.ParseUint(val, 10, typ.Size)
			if err != nil {
				return res, err
			}
			return convertUint(x, typ.GetType().Kind()), nil
		}
		b := new(big.Int)
		b.SetString(val, 10)
		return b, err
	case BoolTy:
		x, err := strconv.ParseBool(val)
		if err != nil {
			return res, err
		}
		return x, nil
	case StringTy:
		return val, nil
	case SliceTy:
		subs, err := procArrayItem(val)
		if err != nil {
			return res, err
		}
		rval := reflect.MakeSlice(typ.GetType(), len(subs), len(subs))
		for idx, sub := range subs {
			subVal, er := Str2GoValue(*typ.Elem, sub)
			if er != nil {
				//return res, er
				subparams, err := procArrayItem(sub) //解析复合类型中的多个元素
				if err != nil {
					return res, er
				}
				fmt.Println("subparams len", len(subparams), "subparams", subparams)
				//获取符合类型对应的参数类型(address,bytes)
				abityps := strings.Split(typ.Elem.stringKind[1:len(typ.Elem.stringKind)-1], ",") //解析为[address, bytes]
				fmt.Println("abityps", abityps)
				//复合类型，继续解析参数

				for i := 0; i < len(subparams); i++ {
					tp, err := NewType(abityps[i], "", nil)
					if err != nil {
						fmt.Println("NewType", err.Error())
						continue
					}
					fmt.Println("tp", tp.stringKind, "types", tp.T)
					subVal, err = Str2GoValue(tp, subparams[i]) //处理对应的元素
					if err != nil {
						fmt.Println("str2GoValue err ", err.Error())
						continue
					}

					fmt.Println("subVal", subVal)
					rval.Index(idx).Field(i).Set(reflect.ValueOf(subVal))

				}

				return rval.Interface(), nil

			}
			rval.Index(idx).Set(reflect.ValueOf(subVal))
		}
		return rval.Interface(), nil
	case ArrayTy:
		rval := reflect.New(typ.GetType()).Elem()
		subs, err := procArrayItem(val)
		if err != nil {
			return res, err
		}
		for idx, sub := range subs {
			subVal, er := Str2GoValue(*typ.Elem, sub)
			if er != nil {
				return res, er
			}
			rval.Index(idx).Set(reflect.ValueOf(subVal))
		}
		return rval.Interface(), nil
	case AddressTy:
		addr := common.StringToAddress(val)
		if addr == nil {
			return res, fmt.Errorf("invalid  address: %v", val)
		}
		return addr.ToHash160(), nil
	case FixedBytesTy:
		// 固定长度多字节，输入时以十六进制方式表示，如 0xabcd00ff
		x, err := common.HexToBytes(val)
		if err != nil {
			return res, err
		}
		rval := reflect.New(typ.GetType()).Elem()
		for i, b := range x {
			rval.Index(i).Set(reflect.ValueOf(b))
		}
		return rval.Interface(), nil
	case BytesTy:
		// 单个字节，输入时以十六进制方式表示，如 0xab
		x, err := common.HexToBytes(val)
		if err != nil {
			return res, err
		}
		return x, nil
	case HashTy:
		// 哈希类型，也是以十六进制为输入，如：0xabcdef
		x, err := common.HexToBytes(val)
		if err != nil {
			return res, err
		}
		return common.BytesToHash(x), nil
	default:
		return res, fmt.Errorf("not support type: %v", typ.stringKind)
	}
}

// 本方法可以将一个表示数组的字符串，经过处理后，返回数组内的字面元素；
// 如果数组为多层，则只返回第一级
// 例如："[a,b,c]" -> "a","b","c"
// 例如："[[a,b],[c,d]]" -> "[a,b]", "[c,d]"
// 因为格式比较复杂，正则表达式不适合处理，所以使用栈的方式来处理
func procArrayItem(val string) (res []string, err error) {
	ss := stack.New()
	data := []rune{}
	for _, b := range val {
		switch b {
		case ' ':
			// 只有字符串元素中间的空格才是有效的
			if ss.Len() > 0 && peekRune(ss) == '"' {
				data = append(data, b)
			}
		case ',':
			// 逗号有可能是多级数组里面的分隔符，我们只处理最外层数组的分隔，
			// 因此，需要判断当前栈中是否只有一个'['，否则就当做普通内容对待
			if ss.Len() == 1 && peekRune(ss) == '[' {
				// 当前元素结束
				res = append(res, string(data))
				data = []rune{}

			} else {
				data = append(data, b)
			}
		case '"':
			// 双引号首次出现时需要入栈，下次出现时需要将两者之间的内容进行拼接
			if ss.Peek() == b {
				ss.Pop()
			} else {
				ss.Push(b)
			}
			//data = append(data, b)
		case '[':
			// 只有当栈为空时，'['才会当做数组的开始，否则全部视作普通内容
			if ss.Len() == 0 {
				data = []rune{}
			} else {
				data = append(data, b)
			}
			ss.Push(b)
		case ']':
			// 只有当栈中只有一个']'时，才会被当做数组结束，否则就当做普通内容对待
			if ss.Len() == 1 && peekRune(ss) == '[' {
				// 整个数组结束
				res = append(res, string(data))
			} else {
				data = append(data, b)
			}
			ss.Pop()
		default:
			// 其它情况全部视作普通内容
			data = append(data, b)
		}
	}

	if ss.Len() != 0 {
		return nil, fmt.Errorf("invalid array format:%v", val)
	}
	return res, err
}

func peekRune(ss *stack.Stack) rune {
	return ss.Peek().(rune)
}

// 解析方法调用字符串，返回方法名以及方法参数
// 例如：foo(param1,param2) -> [foo,param1,param2]
func ProcFuncCall(param string) (funcName string, res []string, err error) {
	lidx := strings.Index(param, "(")
	ridx := strings.LastIndex(param, ")")

	if lidx == -1 || ridx == -1 {
		return funcName, res, fmt.Errorf("invalid function signature:%v", param)
	}

	funcName = strings.TrimSpace(param[:lidx])
	params := strings.TrimSpace(param[lidx+1 : ridx])

	// 将方法参数转换为数组形式，重用数组内容解析逻辑，获得各个具体的参数
	if len(params) > 0 {
		res, err = procArrayItem(fmt.Sprintf("[%v]", params))
	}

	return funcName, res, err
}
