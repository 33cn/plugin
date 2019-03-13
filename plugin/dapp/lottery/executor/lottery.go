// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"sort"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/lottery/types"
)

var llog = log.New("module", "execs.lottery")
var driverName = pty.LotteryX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Lottery{}))
}

type subConfig struct {
	ParaRemoteGrpcClient string `json:"paraRemoteGrpcClient"`
}

var cfg subConfig

// Init lottery
func Init(name string, sub []byte) {
	driverName := GetName()
	if name != driverName {
		panic("system dapp can't be rename")
	}
	if sub != nil {
		types.MustDecode(sub, &cfg)
	}
	drivers.Register(driverName, newLottery, types.GetDappFork(driverName, "Enable"))
}

// GetName for lottery
func GetName() string {
	return newLottery().GetName()
}

// Lottery driver
type Lottery struct {
	drivers.DriverBase
}

func newLottery() drivers.Driver {
	l := &Lottery{}
	l.SetChild(l)
	l.SetExecutorType(types.LoadExecutorType(driverName))
	return l
}

// GetDriverName for lottery
func (lott *Lottery) GetDriverName() string {
	return pty.LotteryX
}

func (lott *Lottery) findLotteryBuyRecords(prefix []byte) (*pty.LotteryBuyRecords, error) {
	count := 0
	var key []byte
	var records pty.LotteryBuyRecords

	for {
		values, err := lott.GetLocalDB().List(prefix, key, DefultCount, 0)
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			var record pty.LotteryBuyRecord
			err := types.Decode(value, &record)
			if err != nil {
				continue
			}
			records.Records = append(records.Records, &record)
		}
		count += len(values)
		if len(values) < int(DefultCount) {
			break
		}
		key = []byte(fmt.Sprintf("%s:%18d", prefix, records.Records[count-1].Index))
	}
	llog.Info("findLotteryBuyRecords", "count", count)

	return &records, nil
}

func (lott *Lottery) findLotteryBuyRecord(key []byte) (*pty.LotteryBuyRecord, error) {
	value, err := lott.GetLocalDB().Get(key)
	if err != nil && err != types.ErrNotFound {
		llog.Error("findLotteryBuyRecord", "err", err)
		return nil, err
	}
	if err == types.ErrNotFound {
		return nil, nil
	}
	var record pty.LotteryBuyRecord

	err = types.Decode(value, &record)
	if err != nil {
		llog.Error("findLotteryBuyRecord", "err", err)
		return nil, err
	}
	return &record, nil
}

func (lott *Lottery) findLotteryDrawRecord(key []byte) (*pty.LotteryDrawRecord, error) {
	value, err := lott.GetLocalDB().Get(key)
	if err != nil && err != types.ErrNotFound {
		llog.Error("findLotteryDrawRecord", "err", err)
		return nil, err
	}
	if err == types.ErrNotFound {
		return nil, nil
	}
	var record pty.LotteryDrawRecord

	err = types.Decode(value, &record)
	if err != nil {
		llog.Error("findLotteryDrawRecord", "err", err)
		return nil, err
	}
	return &record, nil
}

func (lott *Lottery) findLotteryGainRecord(key []byte) (*pty.LotteryGainRecord, error) {
	value, err := lott.GetLocalDB().Get(key)
	if err != nil && err != types.ErrNotFound {
		llog.Error("findLotteryGainRecord", "err", err)
		return nil, err
	}
	if err == types.ErrNotFound {
		return nil, nil
	}
	var record pty.LotteryGainRecord

	err = types.Decode(value, &record)
	if err != nil {
		llog.Error("findLotteryGainRecord", "err", err)
		return nil, err
	}
	return &record, nil
}

func (lott *Lottery) saveLotteryBuy(lotterylog *pty.ReceiptLottery) (kvs []*types.KeyValue) {
	key := calcLotteryBuyKey(lotterylog.LotteryId, lotterylog.Addr, lotterylog.Round, lotterylog.Index)
	record := &pty.LotteryBuyRecord{Number: lotterylog.Number, Amount: lotterylog.Amount, Round: lotterylog.Round, Type: 0, Way: lotterylog.Way, Index: lotterylog.Index, Time: lotterylog.Time, TxHash: lotterylog.TxHash}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

	kvs = append(kvs, kv)
	return kvs
}

func (lott *Lottery) deleteLotteryBuy(lotterylog *pty.ReceiptLottery) (kvs []*types.KeyValue) {
	key := calcLotteryBuyKey(lotterylog.LotteryId, lotterylog.Addr, lotterylog.Round, lotterylog.Index)

	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (lott *Lottery) updateLotteryBuy(lotterylog *pty.ReceiptLottery, isAdd bool) (kvs []*types.KeyValue) {
	if lotterylog.UpdateInfo != nil {
		llog.Debug("updateLotteryBuy")
		buyInfo := lotterylog.UpdateInfo.BuyInfo
		//sort for map
		addrkeys := make([]string, len(buyInfo))
		i := 0

		for addr := range buyInfo {
			addrkeys[i] = addr
			i++
		}
		sort.Strings(addrkeys)
		//update old record
		for _, addr := range addrkeys {
			for _, updateRec := range buyInfo[addr].Records {
				//find addr, index
				key := calcLotteryBuyKey(lotterylog.LotteryId, addr, lotterylog.Round, updateRec.Index)
				record, err := lott.findLotteryBuyRecord(key)
				if err != nil || record == nil {
					return kvs
				}

				if isAdd {
					llog.Debug("updateLotteryBuy update key")
					record.Type = updateRec.Type
				} else {
					record.Type = 0
				}

				kv := &types.KeyValue{Key: key, Value: types.Encode(record)}
				kvs = append(kvs, kv)
			}
		}
		return kvs
	}
	return kvs
}

func (lott *Lottery) saveLotteryDraw(lotterylog *pty.ReceiptLottery) (kvs []*types.KeyValue) {
	key := calcLotteryDrawKey(lotterylog.LotteryId, lotterylog.Round)
	record := &pty.LotteryDrawRecord{Number: lotterylog.LuckyNumber, Round: lotterylog.Round, Time: lotterylog.Time, TxHash: lotterylog.TxHash, TotalAddrNum: lotterylog.TotalAddrNum, BuyAmount: lotterylog.BuyAmount, LuckyAddrNum: lotterylog.LuckyAddrNum, TotalFund: lotterylog.TotalFund, Factor: lotterylog.Factor}
	kv := &types.KeyValue{Key: key, Value: types.Encode(record)}
	kvs = append(kvs, kv)
	return kvs
}

func (lott *Lottery) deleteLotteryDraw(lotterylog *pty.ReceiptLottery) (kvs []*types.KeyValue) {
	key := calcLotteryDrawKey(lotterylog.LotteryId, lotterylog.Round)
	kv := &types.KeyValue{Key: key, Value: nil}
	kvs = append(kvs, kv)
	return kvs
}

func (lott *Lottery) saveLottery(lotterylog *pty.ReceiptLottery) (kvs []*types.KeyValue) {
	if lotterylog.PrevStatus > 0 {
		kv := dellottery(lotterylog.LotteryId, lotterylog.PrevStatus)
		kvs = append(kvs, kv)
	}
	kvs = append(kvs, addlottery(lotterylog.LotteryId, lotterylog.Status))
	return kvs
}

func (lott *Lottery) deleteLottery(lotterylog *pty.ReceiptLottery) (kvs []*types.KeyValue) {
	if lotterylog.PrevStatus > 0 {
		kv := addlottery(lotterylog.LotteryId, lotterylog.PrevStatus)
		kvs = append(kvs, kv)
	}
	kvs = append(kvs, dellottery(lotterylog.LotteryId, lotterylog.Status))
	return kvs
}

func addlottery(lotteryID string, status int32) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcLotteryKey(lotteryID, status)
	kv.Value = []byte(lotteryID)
	return kv
}

func dellottery(lotteryID string, status int32) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcLotteryKey(lotteryID, status)
	kv.Value = nil
	return kv
}

// GetPayloadValue lotteryAction
func (lott *Lottery) GetPayloadValue() types.Message {
	return &pty.LotteryAction{}
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (lott *Lottery) CheckReceiptExecOk() bool {
	return true
}

func (lott *Lottery) saveLotteryGain(lotterylog *pty.ReceiptLottery) (kvs []*types.KeyValue) {
	for _, gain := range lotterylog.GainInfos.Gains {
		key := calcLotteryGainKey(lotterylog.LotteryId, gain.Addr, lotterylog.Round)
		record := &pty.LotteryGainRecord{Addr: gain.Addr, BuyAmount: gain.BuyAmount, FundAmount: gain.FundAmount, Round: lotterylog.Round}
		kv := &types.KeyValue{Key: key, Value: types.Encode(record)}

		kvs = append(kvs, kv)
	}

	return kvs
}

func (lott *Lottery) deleteLotteryGain(lotterylog *pty.ReceiptLottery) (kvs []*types.KeyValue) {
	for _, gain := range lotterylog.GainInfos.Gains {
		kv := &types.KeyValue{}
		kv.Key = calcLotteryGainKey(lotterylog.LotteryId, gain.Addr, lotterylog.Round)
		kv.Value = nil

		kvs = append(kvs, kv)
	}

	return kvs
}
