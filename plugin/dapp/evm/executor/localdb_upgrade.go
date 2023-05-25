package executor

import (
	"context"
	"fmt"

	_ "github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/grpcclient"
	_ "github.com/33cn/chain33/system/address/eth"
	"github.com/33cn/chain33/system/crypto/secp256k1eth"
	"github.com/33cn/chain33/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	evmnonceLocaldbVersioin = "LODB-evmnonce-version"
)

var evmlog = log.New("module", "execs.evm")

type keyTy map[string][]*types.Transaction

func (evm *EVMExecutor) Upgrade() (*types.LocalDBSet, error) {
	version, err := getVersion(evm.GetLocalDB())
	if err == nil && version == 2 { //默认版本号是1
		return nil, nil
	}
	evmlog.Info("++++++++++++++Evm Upgrade+++++++++++++++")
	return evm.upgradeLocalDBV2(evm.GetHeight())

}

func (evm *EVMExecutor) upgradeLocalDBV2(endHeight int64) (*types.LocalDBSet, error) {
	var startHeight int64 = 26625000
	if endHeight < startHeight {
		return nil, nil
	}
	var kvset types.LocalDBSet
	kvs, err := evm.fixNonceLocalDBPart1(startHeight, endHeight)
	if err != nil {
		return nil, errors.Wrap(err, "fixevmnonceLocalDBPart1 setVersion")
	}

	if len(kvs) > 0 {
		kvset.KV = append(kvset.KV, kvs...)
	} else {
		evmlog.Info("upgradeLocalDBV2 kv empty")
	}

	kvs, err = setVersion(evm.GetLocalDB(), 2) //进入修正模式
	if err != nil {
		return nil, errors.Wrap(err, "upgradeLocalDBV2 setVersion")
	}

	if len(kvs) > 0 {
		kvset.KV = append(kvset.KV, kvs...)
	}

	return &kvset, nil

}

func (evm *EVMExecutor) fixNonceLocalDBPart1(startHeight, endHeight int64) ([]*types.KeyValue, error) {
	start := startHeight
	var evmNonceStore = make(keyTy)
	var end = start
	for {
		end = start + 10
		if end > endHeight {
			end = endHeight
		}

		details, err := evm.GetAPI().GetBlocks(&types.ReqBlocks{
			IsDetail: true,
			Start:    start,
			End:      end,
		})
		if err != nil {
			evmlog.Error("fixNonceLocalDBPart1", "getblock err:", err)
			continue
		}
		evmlog.Debug("fixNonceLocalDBPart1 down success ", "start:", start, "end:", end)
		paraseBlockDetails(details, evmNonceStore)
		start = end + 1
		if start > endHeight {
			break
		}

	}
	//开始统计nonce信息
	evmlog.Info("fixNonceLocalDBPart1", "need check evm address num:", len(evmNonceStore))
	kv, err := evm.statisticSigleTxNonce(evmNonceStore)
	if err != nil {
		return nil, err
	}

	return kv, nil
}

func (evm *EVMExecutor) statisticSigleTxNonce(keymap keyTy) ([]*types.KeyValue, error) {
	var kvs []*types.KeyValue
	var count int
	for addr, _ := range keymap {
		count++
		//get current nonce
		nonceV, err := evm.GetLocalDB().Get(secp256k1eth.CaculCoinsEvmAccountKey(addr))
		if err != nil {
			return nil, err
		}
		var evmNonce types.EvmAccountNonce
		err = types.Decode(nonceV, &evmNonce)
		if err != nil {
			panic(err)
		}

		nonce := evmNonce.GetNonce()
		processPercent := fmt.Sprintf("Process Percent:%v%", (count*100)/len(keymap))
		evmlog.Info("statisticSigleTxNonce", "current nonce :", evmNonce.GetNonce(), "addr:", addr, "Upgrade ", processPercent)
		//统计所有的
		txs, err := getTxsByAddrV2(addr, evm)
		if err != nil {
			panic(err)
		}
		caculNonce, err := caculEvmNonce(txs)
		if err != nil {
			evmlog.Error("statisticSigleTxNonce", "caculEvmNonce err", err)
			return nil, errors.Wrap(err, "statisticSigleTxNonce checkEvmNonceAdd")
		}

		if caculNonce != nonce {
			//修正nonce
			var evmNonce types.EvmAccountNonce
			evmNonce.Nonce = caculNonce
			nonceLocalKey := secp256k1eth.CaculCoinsEvmAccountKey(addr)
			err = evm.GetLocalDB().Set(nonceLocalKey, types.Encode(&evmNonce))
			if err != nil {
				return nil, errors.Wrap(err, "localdb set nonce")
			}

			evmlog.Warn("+++++++ evm nonce need upgrade +++++++", "addr:", addr, "currentNonce:", nonce, "setNonce:", caculNonce)
			kv := &types.KeyValue{Key: nonceLocalKey, Value: types.Encode(&evmNonce)}
			kvs = append(kvs, kv)
		} else {
			evmlog.Debug("statisticSigleTxNonce  localdb check ok", "addr:", addr, "currentNonce:", nonce, "setNonce:", caculNonce)
		}

	}

	return kvs, nil
}

func getTxsByAddrV2(addr string, evm *EVMExecutor) ([]*types.Transaction, error) {

	prefix := types.CalcTxAddrDirHashKey(addr, 1, "")
	infos, err := evm.GetLocalDB().List(prefix, nil, 1024, 1)
	if err != nil {
		evmlog.Error("getTxsByAddrV2", "db.list err", err)
		return nil, errors.Wrap(err, "statisticSigleTxNonce.GetTxListByAddr ")
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

	gcli, err := grpcclient.NewMainChainClient(evm.GetAPI().GetConfig(), "cloud.bityuan.com:8802")
	if err != nil {
		panic(err)
	}

	txdetails, err := gcli.GetTransactionByHashes(context.Background(), &types.ReqHashes{Hashes: txhashes})
	if err != nil {
		evmlog.Error("getTxsByAddrV2", "GetTransactionByHashes:", err)
		return nil, err
	}
	var txs []*types.Transaction

	for _, tx := range txdetails.GetTxs() {
		txs = append(txs, tx.GetTx())
	}

	return txs, nil

}
func getTxsByAddr(addr string, evm *EVMExecutor) ([]*types.Transaction, error) {
	prefix := types.CalcTxAddrDirHashKey(addr, 1, "")
	infos, err := evm.GetLocalDB().List(prefix, nil, 1024, 1)
	if err != nil {
		evmlog.Error("getTxsByAddr", "db.list err", err)
		return nil, errors.Wrap(err, "statisticSigleTxNonce.GetTxListByAddr ")
	}
	//解析获取哈希list

	var replyTxInfos types.ReplyTxInfos
	replyTxInfos.TxInfos = make([]*types.ReplyTxInfo, len(infos))
	for index, infobytes := range infos {
		var replyTxInfo types.ReplyTxInfo
		err = types.Decode(infobytes, &replyTxInfo)
		if err != nil {
			evmlog.Error("getTxsByAddr", "Decode err", err)
			return nil, err
		}
		replyTxInfos.TxInfos[index] = &replyTxInfo
	}

	var txhashes = make([][]byte, 0)
	for i, info := range replyTxInfos.GetTxInfos() {
		evmlog.Info("getTxsByAddr", "index:", i, "addr:", addr, "hash:", common.Bytes2Hex(info.GetHash()))
		txhashes = append(txhashes, info.GetHash())
	}

	evmlog.Info("getTxsByAddr", "GetTransactionByAddr success from:", addr, "size:", len(replyTxInfos.GetTxInfos()), "txhashes size:", len(txhashes))

	//获取交易
	txdetails, err := evm.GetAPI().GetTransactionByHash(&types.ReqHashes{
		Hashes: txhashes,
	})
	if err != nil {
		panic(err)
	}

	var txs []*types.Transaction

	for _, tx := range txdetails.GetTxs() {
		txs = append(txs, tx.GetTx())
	}
	return txs, nil
	/*var txs []*types.Transaction
	for _, hash := range txhashes {
		evmlog.Info("statisticSigleTxNonce", "txhash:", common.Bytes2Hex(hash))
		rawTx, err := evm.GetLocalDB().Get(evm.GetAPI().GetConfig().CalcTxKey(hash))
		if err != nil {
			evmlog.Error("statisticSigleTxNonce", "db Get tx err", err)
			return nil, errors.Wrap(err, "statisticSigleTxNonce.db Get tx ")
		}
		var txResult types.TxResult
		err = types.Decode(rawTx, &txResult)
		if err != nil {
			panic(err)
		}
		txs = append(txs, txResult.Tx)
		if txResult.GetTx().From() != addr {
			evmlog.Info(fmt.Sprintf("address not match tx.From:%v,addr:%v,hash:%v,tx.hash", txResult.GetTx().From(), addr, common.Bytes2Hex(hash)), common.Bytes2Hex(txResult.GetTx().Hash()))
		}
	}*/

}
func caculEvmNonce(txs []*types.Transaction) (rightNonce int64, err error) {
	evmlog.Debug("caculEvmNonce", "txnum:", len(txs), "tx.From", txs[0].From())
	var nextNonce int64
	var initFlag bool
	for _, tx := range txs {
		if !types.IsEthSignID(tx.GetSignature().GetTy()) {
			continue
		}
		evmlog.Debug("caculEvmNonce", "addr:", tx.From(), "nonce:", tx.Nonce)
		if tx.GetNonce() == 0 && !initFlag { //兼容交易第一笔理论上应该是0
			nextNonce++
			initFlag = true
			continue
		}

		if tx.GetNonce() == nextNonce {
			nextNonce++
		}
	}
	return nextNonce, nil
}

func paraseBlockDetails(details *types.BlockDetails, keymap keyTy) {
	evmlog.Debug("paraseBlockDetails", "items:", len(details.GetItems()))
	for _, detail := range details.GetItems() {
		var evmtxNum int
		for _, tx := range detail.GetBlock().GetTxs() {
			if types.IsEthSignID(tx.GetSignature().GetTy()) {
				evmtxNum++
				if len(keymap[tx.From()]) == 0 {
					var txs []*types.Transaction
					keymap[tx.From()] = txs
				}
				txs := keymap[tx.From()]
				txs = append(txs, tx)
				keymap[tx.From()] = txs
			}
		}

	}

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
