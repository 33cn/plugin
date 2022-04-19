package utils

// --------------------------------------------------------
//      Utils
//
//      Utils contains utility functionality for the ebrelayer.
// --------------------------------------------------------

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	chain33Abi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	nullAddress = "0x0000000000000000000000000000000000000000"
)

var Decimal2value = map[int]int64{
	1:  1e1,
	2:  1e2,
	3:  1e3,
	4:  1e4,
	5:  1e5,
	6:  1e6,
	7:  1e7,
	8:  1e8,
	9:  1e9,
	10: 1e10,
	11: 1e11,
}

var log = log15.New("module", "utils")

// IsZeroAddress : checks an Ethereum address and returns a bool which indicates if it is the null address
func IsZeroAddress(address common.Address) bool {
	return address == common.HexToAddress(nullAddress)
}

//IsValidPassWord 密码合法性校验,密码长度在8-30位之间。必须是数字+字母的组合
func IsValidPassWord(password string) bool {
	pwLen := len(password)
	if pwLen < 8 || pwLen > 30 {
		return false
	}

	var char bool
	var digit bool
	for _, s := range password {
		if unicode.IsLetter(s) {
			char = true
		} else if unicode.IsDigit(s) {
			digit = true
		} else {
			return false
		}
	}
	return char && digit
}

func decodeInt64(int64bytes []byte) (int64, error) {
	var value types.Int64
	err := types.Decode(int64bytes, &value)
	if err != nil {
		//may be old database format json...
		err = json.Unmarshal(int64bytes, &value.Data)
		if err != nil {
			return -1, types.ErrUnmarshal
		}
	}
	return value.Data, nil
}

//LoadInt64FromDB ...
func LoadInt64FromDB(key []byte, db dbm.DB) (int64, error) {
	bytes, err := db.Get(key)
	if bytes == nil || err != nil {
		//if err != dbm.ErrNotFoundInDb {
		//	log.Error("LoadInt64FromDB", "error", err)
		//}
		return 0, types.ErrHeightNotExist
	}
	return decodeInt64(bytes)
}

//QueryTxhashes ...
func QueryTxhashes(prefix []byte, db dbm.DB) []string {
	kvdb := dbm.NewKVDB(db)
	hashes, err := kvdb.List(prefix, nil, 10, 1)
	if nil != err {
		return nil
	}

	var hashStrs []string
	for _, hash := range hashes {
		hashStrs = append(hashStrs, string(hash))
	}
	return hashStrs
}

//Addr2DecimalsKey ...
var (
	Addr2DecimalsKey = []byte("prefix_for_Addr2Decimals")
)

//CalAddr2DecimalsPrefix ...
func CalAddr2DecimalsPrefix(tokenAddr string) []byte {
	return []byte(fmt.Sprintf("%s-%s", Addr2DecimalsKey, tokenAddr))
}

//GetDecimalsFromDB ...
func GetDecimalsFromDB(addr string, db dbm.DB) (int64, error) {
	res, err := db.Get(CalAddr2DecimalsPrefix(addr))
	if err != nil {
		return 0, err
	}
	var addr2Decimals map[string]int64
	err = json.Unmarshal(res, &addr2Decimals)
	if err != nil {
		return 0, err
	}
	if d, ok := addr2Decimals[addr]; ok {
		return d, nil
	}
	return 0, types.ErrNotFound
}

func SimpleGetDecimals(addr string) (int64, error) {
	if addr == "0x0000000000000000000000000000000000000000" || addr == "" {
		return 18, nil
	}
	return 8, nil
}

//GetDecimalsFromNode ...
func GetDecimalsFromNode(addr, rpcLaddr string) (int64, error) {
	if addr == "0x0000000000000000000000000000000000000000" || addr == "" {
		return 18, nil
	}

	client, err := ethclient.Dial(rpcLaddr)
	if err != nil {
		log.Error("GetDecimals", "SetupEthClient error:", err.Error())
		return 0, err
	}

	msg, err := QueryResult("decimals()", generated.ERC20ABI, addr, addr, client)
	decimals, err := strconv.ParseInt(msg, 0, 64)
	if err != nil {
		log.Error("GetDecimals", "ParseInt error:", err.Error())
		return 0, err
	}
	return decimals, nil
}

func QueryResult(param, abiData, contract, owner string, client ethinterface.EthClientSpec) (string, error) {
	log.Info("QueryResult", "param", param, "contract", contract, "owner", owner)
	// 首先解析参数字符串，分析出方法名以及个参数取值
	methodName, params, err := chain33Abi.ProcFuncCall(param)
	if err != nil {
		return methodName + " ProcFuncCall fail", err
	}

	// 解析ABI数据结构，获取本次调用的方法对象
	abi_, err := chain33Abi.JSON(strings.NewReader(abiData))
	if err != nil {
		log.Error("QueryResult", "JSON fail", err)
		return methodName + " JSON fail", err
	}

	var method chain33Abi.Method
	var ok bool
	if method, ok = abi_.Methods[methodName]; !ok {
		err = fmt.Errorf("function %v not exists", methodName)
		return methodName, err
	}

	if !method.IsConstant() {
		return methodName, errors.New("method is not readonly")
	}
	if len(params) != method.Inputs.LengthNonIndexed() {
		err = fmt.Errorf("function params error:%v", params)
		return methodName, err
	}

	// 获取方法参数对象，遍历解析各参数，获得参数的Go取值
	paramVals := []interface{}{}
	if len(params) != 0 {
		// 首先检查参数个数和ABI中定义的是否一致
		if method.Inputs.LengthNonIndexed() != len(params) {
			err = fmt.Errorf("function Params count error: %v", param)
			return methodName, err
		}

		for i, v := range method.Inputs.NonIndexed() {
			paramVal, err := chain33Abi.Str2GoValue(v.Type, params[i])
			if err != nil {
				log.Error("QueryResult", "Str2GoValue fail", err)
				return methodName + " Str2GoValue fail", err
			}
			paramVals = append(paramVals, paramVal)
		}
	}

	ownerAddr := common.HexToAddress(owner)
	opts := &bind.CallOpts{
		Pending: true,
		From:    ownerAddr,
		Context: context.Background(),
	}
	var out []interface{}
	// Convert the raw abi into a usable format
	contractABI, err := abi.JSON(strings.NewReader(abiData))
	if err != nil {
		return "JSON err", err
	}
	boundContract := bind.NewBoundContract(common.HexToAddress(contract), contractABI, client, nil, nil)
	err = boundContract.Call(opts, &out, methodName, paramVals...)
	if err != nil {
		log.Error("QueryResult", "call fail", err)
		return "call err", err
	}
	return fmt.Sprint(out[0]), err
}

func SendToServer(url string, req io.Reader) ([]byte, error) {
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(10 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
	var request *http.Request
	var err error

	request, err = http.NewRequest("POST", url, req)

	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}

//MultiplySpecifyTimes ...
func MultiplySpecifyTimes(start float64, time int64) float64 {
	for i := 0; i < int(time); i++ {
		start *= 10
	}
	return start
}

//Toeth ...
func Toeth(amount string, decimal int64) float64 {

	bf := big.NewFloat(0)
	var ok bool
	bf, ok = bf.SetString(TrimZeroAndDot(amount))
	if !ok {
		return 0
	}
	bf = bf.Quo(bf, big.NewFloat(MultiplySpecifyTimes(1, decimal)))
	f, _ := bf.Float64()
	return f
}

//ToWei 将eth单位的金额转为wei单位
func ToWei(amount float64, decimal int64) *big.Int {
	var ok bool
	bn := big.NewInt(1)
	if decimal > 4 {
		bn, ok = bn.SetString(TrimZeroAndDot(fmt.Sprintf("%.0f", MultiplySpecifyTimes(math.Trunc(amount*1e4+0.0000000000001), decimal-4))), 10)
	} else {
		bn, ok = bn.SetString(TrimZeroAndDot(fmt.Sprintf("%.0f", MultiplySpecifyTimes(amount, decimal))), 10)
	}
	if ok {
		return bn
	}

	return nil
}

func SmalToBig(amount float64, decimals uint8) *big.Int {
	bfa := big.NewFloat(amount)
	bfa = bfa.Mul(bfa, big.NewFloat(1).SetInt64(ebTypes.DecimalsPrefix[decimals]))
	bn, _ := bfa.Int(nil)
	return bn
}

//TrimZeroAndDot ...
func TrimZeroAndDot(s string) string {
	if strings.Contains(s, ".") {
		var trimDotStr string
		trimZeroStr := strings.TrimRight(s, "0")
		trimDotStr = strings.TrimRight(trimZeroStr, ".")
		return trimDotStr
	}

	return s
}

//CheckPower ...
func CheckPower(power int64) bool {
	if power <= 0 || power > 100 {
		return false
	}
	return true
}

//DivideDot ...
func DivideDot(in string) (left, right string, err error) {
	if strings.Contains(in, ".") {
		ss := strings.Split(in, ".")
		return ss[0], ss[1], nil
	}
	return "", "", errors.New("Divide error")
}

//IsExecAddrMatch ...
func IsExecAddrMatch(name string, to string) bool {
	toaddr := address.ExecAddress(name)
	return toaddr == to
}
