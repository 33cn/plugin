package utils

// --------------------------------------------------------
//      Utils
//
//      Utils contains utility functionality for the ebrelayer.
// --------------------------------------------------------

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	simplejson "github.com/bitly/go-simplejson"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	nullAddress = "0x0000000000000000000000000000000000000000"
)

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

//GetDecimalsFromNode ...
func GetDecimalsFromNode(addr string, nodeAddr string) (int64, error) {
	if addr == "0x0000000000000000000000000000000000000000" || addr == "" {
		return 18, nil
	}
	Hashprefix := "0x313ce567"
	postData := fmt.Sprintf(`{"id":1,"jsonrpc":"2.0","method":"eth_call","params":[{"to":"%s", "data":"%s"},"latest"]}`, addr, Hashprefix)

	retryTimes := 0
RETRY:
	res, err := sendToServer(nodeAddr, strings.NewReader(postData))
	if err != nil {
		log.Error("GetDecimals", "error:", err.Error())
		if retryTimes > 3 {
			return 0, err
		}
		retryTimes++
		goto RETRY
	}
	js, err := simplejson.NewJson(res)
	if err != nil {
		log.Error("GetDecimals", "NewJson error:", err.Error())
		if retryTimes > 3 {
			return 0, err
		}
		retryTimes++
		goto RETRY
	}
	result := js.Get("result").MustString()

	decimals, err := strconv.ParseInt(result, 0, 64)
	if err != nil {
		if retryTimes > 3 {
			return 0, err
		}
		retryTimes++
		goto RETRY
	}
	return decimals, nil
}

func sendToServer(url string, req io.Reader) ([]byte, error) {
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
