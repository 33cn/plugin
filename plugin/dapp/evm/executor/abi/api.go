package abi

import (
	"fmt"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/golang-collections/collections/stack"
	"reflect"
	"strconv"
)

// Pack 使用ABI方式调用时，将调用方式转换为EVM底层处理的十六进制编码
// abiData 完整的ABI定义
// param 调用方法及参数
func Pack(param, abiData string) ([]byte, error) {
	return nil, nil
}

// Unpack 将调用返回结果按照ABI的格式序列化为json
// data 合约方法返回值
// abiData 完整的ABI定义
func Unpack(data []byte, abiData string) (string, error) {
	return "", nil
}

// 从字符串格式的输入参数取值（单个），获取Go类型的
func goValue(typ Type, val string) (res interface{}, err error) {
	switch typ.T {
	case IntTy:
		bitSize := 0
		pos := uint(typ.Kind - reflect.Int)
		if pos > 0 {
			bitSize = (2 << pos) * 2
		}
		x, err := strconv.ParseInt(val, 10, bitSize)
		if err != nil {
			return res, err
		}
		return x, nil
	case UintTy:
		bitSize := 0
		pos := uint(typ.Kind - reflect.Uint)
		if pos > 0 {
			bitSize = (2 << pos) * 2
		}
		x, err := strconv.ParseUint(val, 10, bitSize)
		if err != nil {
			return res, err
		}
		return x, nil
	case BoolTy:
		x, err := strconv.ParseBool(val)
		if err != nil {
			return res, err
		}
		return x, nil
	case StringTy:
		return val, nil
	//case SliceTy:
	//	var data []interface{}
	//	subs, err := getSubArrayStr(val)
	//	if err != nil {
	//		return res, err
	//	}
	//	for idx, sub := range subs {
	//		subVal, er := goValue(*typ.Elem, sub)
	//		if er != nil {
	//			return res, er
	//		}
	//		data[idx] = subVal
	//	}
	//	return data, nil
	//case ArrayTy:
	//	var data [typ.Size]interface{}
	//	subs, err := getSubArrayStr(val)
	//	if err != nil {
	//		return res, err
	//	}
	//	for idx, sub := range subs {
	//		subVal, er := goValue(*typ.Elem, sub)
	//		if er != nil {
	//			return res, er
	//		}
	//		data[idx] = subVal
	//	}
	//	return data, nil
	case AddressTy:
		addr := common.StringToAddress(val)
		if addr == nil {
			return res, fmt.Errorf("invalid  address: %v", val)
		}
		return addr.ToHash160(), nil
	case FixedBytesTy:
		//rtype := reflect.ArrayOf(typ.Size, reflect.TypeOf(byte(0)))
		//value := reflect.New(rtype).Elem()
		//value.SetBytes(x)

		// 固定长度多字节，输入时以十六进制方式表示，如 0xabcd00ff
		//x, err := common.HexToBytes(val)
		//if err != nil {
		//	return res, err
		//}
		//var data [typ.Size]byte
		//copy(data[:], x)
		//return data, nil
	case BytesTy:
		// 单个字节，输入时以十六进制方式表示，如 0xab
		x, err := common.HexToBytes(val)
		if err != nil {
			return res, err
		}
		return x[0], nil
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
	return res, nil
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
			if ss.Len() > 0 && stackPeek(ss) == '"' {
				data = append(data, b)
			}
		case ',':
			// 逗号有可能是多级数组里面的分隔符，我们只处理最外层数组的分隔，
			// 因此，需要判断当前栈中是否只有一个'['，否则就当做普通内容对待
			if ss.Len() == 1 && stackPeek(ss) == '[' {
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
			if ss.Len() == 1 && stackPeek(ss) == '[' {
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

func stackPeek(ss *stack.Stack) rune {
	return ss.Peek().(rune)
}