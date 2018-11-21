package abi

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