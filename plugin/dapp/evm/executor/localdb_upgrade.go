package executor

import (
	"context"
	"fmt"
	"time"

	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/system/crypto/secp256k1eth"
	"github.com/33cn/chain33/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	evmnonceLocaldbVersioin = "LODB-evmnonce-version"
)

var evmlog = log.New("module", "execs.evm")

func (evm *EVMExecutor) Upgrade() (*types.LocalDBSet, error) {
	//nonceUpGrade
	conf := types.Conf(evm.GetAPI().GetConfig(), "config.exec.sub.evm")
	if !conf.IsEnable("nonceUpGrade") {
		return nil, nil
	}

	version, err := getVersion(evm.GetLocalDB())
	if err == nil && version == 2 { //默认版本号是1
		return nil, nil
	}
	evmlog.Info("++++++++++++++Evm Upgrade+++++++++++++++")
	return evm.upgradeLocalDBV2()
}

func (evm *EVMExecutor) upgradeLocalDBV2() (*types.LocalDBSet, error) {
	var kvset types.LocalDBSet
	kvs, err := evm.upgradeNonceLocalDBV2()
	if err != nil {
		return nil, errors.Wrap(err, "upgradeLocalDBV2 setVersion")
	}

	if len(kvs) > 0 {
		kvset.KV = append(kvset.KV, kvs...)
	} else {
		evmlog.Info("---------- upgradeLocalDBV2 kv empty ----------")
	}

	kvs, err = setVersion(evm.GetLocalDB(), 2) //设定新版本
	if err != nil {
		return nil, errors.Wrap(err, "upgradeLocalDBV2 setVersion")
	}

	if len(kvs) > 0 {
		kvset.KV = append(kvset.KV, kvs...)
	}

	return &kvset, nil
}

func (evm *EVMExecutor) upgradeNonceLocalDBV2() ([]*types.KeyValue, error) {

	var kvs []*types.KeyValue
	prefix := "LODB-" + "evm" + "-noncestate:"
	allEvmAccountKey, err := evm.GetLocalDB().List([]byte(prefix), nil, 0, 0)
	if err != nil {
		panic(err)
	}
	if len(allEvmAccountKey) == 0 {
		return kvs, nil
	}
	evmlog.Info("upgradeNonceLocalDBV2", "getAccoutEvmKey total num:", len(allEvmAccountKey), "currentHeight:", evm.GetHeight())
	conf := types.Conf(evm.GetAPI().GetConfig(), "config.exec.sub.evm")
	seedUrl := conf.GStr("upgradeUrl")
	gcli, err := grpcclient.NewMainChainClient(evm.GetAPI().GetConfig(), seedUrl)
	if err != nil {
		panic(err)
	}
	//check tx list
	var index int
	var upgradeNonceLog []string
	var emptyAddrNum int
	for {
		processPercent := fmt.Sprintf("Index:%v,Total:%v,EmptyAddrNum:%v,Process Percent:%v %v", index, len(allEvmAccountKey), emptyAddrNum, ((index+1)*100)/len(allEvmAccountKey), "%")
		evmlog.Info("upgradeNonceLocalDBV2", "Upgrade ", processPercent)

		if index == len(allEvmAccountKey)-1 {
			break
		}

		var evmNonce types.EvmAccountNonce
		err = types.Decode(allEvmAccountKey[index], &evmNonce)
		if err != nil {
			panic(err)
		}
		if evmNonce.GetAddr() == "" {
			index++
			emptyAddrNum++
			continue
		}
		evmlog.Debug("upgradeNonceLocalDBV2", "addr", evmNonce.Addr, "nonce", evmNonce.GetNonce())
		checkAddr := evmNonce.Addr

		txs, err := getAllTxByAddr(checkAddr, gcli, evm)
		if err != nil {
			if err == types.ErrNotFound {
				//当前数据库不存在此交易 nonce 应当是0
				if evmNonce.GetNonce() != 0 {
					setkvs, printLog := evm.updateEvmNonce(checkAddr, evmNonce.GetNonce(), 0)
					kvs = append(kvs, setkvs...)
					evmlog.Warn("upgradeNonceLocalDBV2", "ErrNotFound Upgrade addr:", checkAddr, "set new nonce: ", "0")
					upgradeNonceLog = append(upgradeNonceLog, printLog...)

				}
				//已经为0
				index++
			} else {
				evmlog.Error("upgradeNonceLocalDBV2", "getTxsByAddrV2,err:", err.Error())
			}

			continue
		}

		caculNonce := cacuEvmNonce(txs)
		setkvs, printLog := evm.updateEvmNonce(checkAddr, evmNonce.GetNonce(), caculNonce)
		kvs = append(kvs, setkvs...)
		upgradeNonceLog = append(upgradeNonceLog, printLog...)
		index++
	}

	for _, logstr := range upgradeNonceLog {
		evmlog.Warn(logstr)
	}

	return kvs, nil

}

func (evm *EVMExecutor) updateEvmNonce(addr string, evmLocalNonce, cacuNonce int64) ([]*types.KeyValue, []string) {
	var kvs []*types.KeyValue
	var printLog []string
	if cacuNonce != evmLocalNonce {
		//修正nonce
		var evmNonce types.EvmAccountNonce
		evmNonce.Nonce = cacuNonce
		evmNonce.Addr = addr
		nonceLocalKey := secp256k1eth.CaculCoinsEvmAccountKey(addr)
		err := evm.GetLocalDB().Set(nonceLocalKey, types.Encode(&evmNonce))
		if err != nil {
			panic(errors.Wrap(err, "localdb set nonce"))
		}

		evmlog.Warn("+++++++ evm nonce need upgrade +++++++", "addr:", addr, "currentNonce:", evmLocalNonce, "setNonce:", cacuNonce)
		printLog = append(printLog, fmt.Sprintf("Upgrade addr:%v,current nonce: %v,set nonce:%v\n", addr, evmLocalNonce, cacuNonce))
		kv := &types.KeyValue{Key: nonceLocalKey, Value: types.Encode(&evmNonce)}
		kvs = append(kvs, kv)
	} else {
		evmlog.Debug("fixNonceLocalDBPart1V2  localdb check ok", "addr:", addr, "currentNonce:", evmLocalNonce, "setNonce:", cacuNonce)
	}

	return kvs, printLog
}

func getAllTxByAddr(addr string, gcli types.Chain33Client, evm *EVMExecutor) ([]*types.Transaction, error) {
	prefix := types.CalcTxAddrDirHashKey(addr, 1, "")
	infos, err := evm.GetLocalDB().List(prefix, nil, 0, 1)
	if err != nil {
		//evmlog.Error("getTxsByAddrV2", "db.list err", err)
		return nil, err
	}
	//解析获取哈希list

	var replyTxInfos types.ReplyTxInfos
	replyTxInfos.TxInfos = make([]*types.ReplyTxInfo, len(infos))
	for index, infobytes := range infos {
		var replyTxInfo types.ReplyTxInfo
		err = types.Decode(infobytes, &replyTxInfo)
		if err != nil {
			evmlog.Error("getTxsByAddrV2", "Decode err", err)
			return nil, err
		}
		replyTxInfos.TxInfos[index] = &replyTxInfo
	}

	var txhashes = make([][]byte, 0)
	for i, info := range replyTxInfos.GetTxInfos() {
		evmlog.Debug("getTxsByAddrV2", "index:", i, "addr:", addr, "hash:", common.Bytes2Hex(info.GetHash()))
		txhashes = append(txhashes, info.GetHash())
	}

	var pageNum = 1000
	var hashNum = len(txhashes)
	var startIndex = 0
	var txs []*types.Transaction
	for {
		endIndex := startIndex + pageNum
		if endIndex > hashNum {
			endIndex = hashNum
		}
		txdetails, err := gcli.GetTransactionByHashes(context.Background(), &types.ReqHashes{Hashes: txhashes[startIndex:endIndex]})
		if err != nil {
			evmlog.Error("getTxsByAddrV2", "GetTransactionByHashes:", err)
			time.Sleep(time.Second)
			continue
		}

		for _, tx := range txdetails.GetTxs() {
			txs = append(txs, tx.GetTx())
		}
		startIndex = endIndex
		if startIndex >= hashNum {
			break
		}
	}
	if len(txs) != hashNum {
		evmlog.Error("getTxsByAddrV2", "GetTransactionByHashes get tx num wrong,expect:", hashNum, "actually:", len(txs))
		return txs, errors.Errorf("get tx wrong num")
	}

	return txs, nil

}

func cacuEvmNonce(txs []*types.Transaction) (rightNonce int64) {
	evmlog.Debug("caculEvmNonce", "txnum:", len(txs), "tx.From", txs[0].From())
	var nextNonce int64
	var initFlag bool
	for _, tx := range txs {
		if !types.IsEthSignID(tx.GetSignature().GetTy()) {
			continue
		}
		evmlog.Debug("caculEvmNonce", "addr:", tx.From(), "nonce:", tx.Nonce)
		if !initFlag { //兼容交易第一笔理论上应该是0
			nextNonce++
			initFlag = true
			continue
		}

		if tx.GetNonce() == nextNonce {
			nextNonce++
		}
	}
	return nextNonce
}

// localdb Version
func getVersion(kvdb dbm.KV) (int, error) {
	value, err := kvdb.Get([]byte(evmnonceLocaldbVersioin))
	if err != nil && err != types.ErrNotFound {
		return 1, err
	}
	if err == types.ErrNotFound {
		return 1, nil
	}
	var v types.Int32
	err = types.Decode(value, &v)
	if err != nil {
		return 1, err
	}
	return int(v.Data), nil
}

func setVersion(kvdb dbm.KV, version int) ([]*types.KeyValue, error) {
	v := types.Int32{Data: int32(version)}
	x := types.Encode(&v)
	err := kvdb.Set([]byte(evmnonceLocaldbVersioin), x)
	return []*types.KeyValue{{Key: []byte(evmnonceLocaldbVersioin), Value: x}}, err
}
