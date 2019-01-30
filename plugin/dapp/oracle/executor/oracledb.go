/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/db/table"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	oty "github.com/33cn/plugin/plugin/dapp/oracle/types"
)

// operation key manage need define
const (
	publishEventKey = "oracle-publish-event"
	//prePublishResultKey = "oracle-prepublish-result"
	//publishResultKey = "oracle-publish-result"
)

// OracleDB struct
type OracleDB struct {
	oty.OracleStatus
}

// NewOracleDB instance
func NewOracleDB(eventID, addr, ty, subTy, content, introduction string, time int64, index int64) *OracleDB {
	oracle := &OracleDB{}
	oracle.EventID = eventID
	oracle.Addr = addr
	oracle.Type = ty
	oracle.SubType = subTy
	oracle.Content = content
	oracle.Introduction = introduction
	oracle.Time = time
	oracle.Status = &oty.EventStatus{OpAddr: addr, Status: oty.EventPublished}
	oracle.PreStatus = &oty.EventStatus{OpAddr: "", Status: oty.NoEvent}
	return oracle
}

// GetKVSet for OracleDB
func (o *OracleDB) GetKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(&o.OracleStatus)
	kvset = append(kvset, &types.KeyValue{Key: Key(o.EventID), Value: value})
	return kvset
}

// Save for OracleDB
func (o *OracleDB) save(db dbm.KV) error {
	set := o.GetKVSet()
	for i := 0; i < len(set); i++ {
		err := db.Set(set[i].GetKey(), set[i].Value)
		if err != nil {
			fmt.Printf("oracledb save failed:[%v]-%v", i, err)
			return err
		}
	}
	return nil
}

// Key for oracle
func Key(id string) (key []byte) {
	key = append(key, []byte("mavl-"+oty.OracleX+"-")...)
	key = append(key, []byte(id)...)
	return key
}

type oracleAction struct {
	db        dbm.KV
	txhash    []byte
	fromaddr  string
	blocktime int64
	height    int64
	index     int
}

func newOracleAction(o *oracle, tx *types.Transaction, index int) *oracleAction {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &oracleAction{o.GetStateDB(), hash, fromaddr,
		o.GetBlockTime(), o.GetHeight(), index}
}

func (action *oracleAction) eventPublish(event *oty.EventPublish) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	eventID := common.ToHex(action.txhash)

	// 事件结果公布事件必须大于区块时间
	if event.Time <= action.blocktime {
		return nil, oty.ErrTimeMustBeFuture
	}
	// 是否是事件发布者
	if !isEventPublisher(action.fromaddr, action.db, false) {
		return nil, oty.ErrNoPrivilege
	}

	_, err := findOracleStatus(action.db, eventID)
	if err != types.ErrNotFound {
		olog.Error("EventPublish", "EventPublish repeated eventID", eventID)
		return nil, oty.ErrOracleRepeatHash
	}

	eventStatus := NewOracleDB(eventID, action.fromaddr, event.Type, event.SubType, event.Content, event.Introduction, event.Time, action.GetIndex())
	olog.Debug("eventPublish", "PublisherAddr", eventStatus.Addr, "EventID", eventStatus.EventID, "Event", eventStatus.Content)

	if err := eventStatus.save(action.db); err != nil {
		return nil, err
	}
	kv = append(kv, eventStatus.GetKVSet()...)

	receiptLog := action.getOracleCommonRecipt(&eventStatus.OracleStatus, oty.TyLogEventPublish)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (action *oracleAction) eventAbort(event *oty.EventAbort) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	//只有发布问题的人能取消问题
	if !isEventPublisher(action.fromaddr, action.db, false) {
		return nil, oty.ErrNoPrivilege
	}

	oracleStatus, err := findOracleStatus(action.db, event.EventID)
	if err == types.ErrNotFound {
		olog.Error("EventAbort", "EventAbort not found eventID", event.EventID)
		return nil, oty.ErrEventIDNotFound
	}

	ora := &OracleDB{*oracleStatus}

	if ora.Status.Status != oty.EventPublished && ora.Status.Status != oty.ResultAborted {
		olog.Error("EventAbort", "EventAbort can not abort for status", ora.Status.Status)
		return nil, oty.ErrEventAbortNotAllowed
	}

	updateStatus(ora, action.GetIndex(), action.fromaddr, oty.EventAborted)

	if err := ora.save(action.db); err != nil {
		return nil, err
	}
	kv = append(kv, ora.GetKVSet()...)

	receiptLog := action.getOracleCommonRecipt(&ora.OracleStatus, oty.TyLogEventAbort)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (action *oracleAction) resultPrePublish(event *oty.ResultPrePublish) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	//只有发布问题的人能取消问题
	if !isEventPublisher(action.fromaddr, action.db, false) {
		return nil, oty.ErrNoPrivilege
	}

	oracleStatus, err := findOracleStatus(action.db, event.EventID)
	if err == types.ErrNotFound {
		olog.Error("ResultPrePublish", "ResultPrePublish not found eventID", event.EventID)
		return nil, oty.ErrEventIDNotFound
	}

	ora := &OracleDB{*oracleStatus}

	if ora.Status.Status != oty.EventPublished && ora.Status.Status != oty.ResultAborted {
		olog.Error("ResultPrePublish", "ResultPrePublish can not pre-publish", ora.Status.Status)
		return nil, oty.ErrResultPrePublishNotAllowed
	}

	updateStatus(ora, action.GetIndex(), action.fromaddr, oty.ResultPrePublished)
	ora.Result = event.Result
	ora.Source = event.Source

	if err := ora.save(action.db); err != nil {
		return nil, err
	}
	kv = append(kv, ora.GetKVSet()...)

	receiptLog := action.getOracleCommonRecipt(&ora.OracleStatus, oty.TyLogResultPrePublish)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (action *oracleAction) resultAbort(event *oty.ResultAbort) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	//只有发布问题的人能取消预发布
	if !isEventPublisher(action.fromaddr, action.db, false) {
		return nil, oty.ErrNoPrivilege
	}

	oracleStatus, err := findOracleStatus(action.db, event.EventID)
	if err == types.ErrNotFound {
		olog.Error("ResultAbort", "ResultAbort not found eventID", event.EventID)
		return nil, oty.ErrEventIDNotFound
	}

	ora := &OracleDB{*oracleStatus}

	if ora.Status.Status != oty.ResultPrePublished {
		olog.Error("ResultAbort", "ResultAbort can not abort", ora.Status.Status)
		return nil, oty.ErrPrePublishAbortNotAllowed
	}

	updateStatus(ora, action.GetIndex(), action.fromaddr, oty.ResultAborted)
	ora.Result = ""
	ora.Source = ""

	if err := ora.save(action.db); err != nil {
		return nil, err
	}
	kv = append(kv, ora.GetKVSet()...)

	receiptLog := action.getOracleCommonRecipt(&ora.OracleStatus, oty.TyLogResultAbort)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (action *oracleAction) resultPublish(event *oty.ResultPublish) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	//只有发布问题的人能取消预发布
	if !isEventPublisher(action.fromaddr, action.db, false) {
		return nil, oty.ErrNoPrivilege
	}

	oracleStatus, err := findOracleStatus(action.db, event.EventID)
	if err == types.ErrNotFound {
		olog.Error("ResultPublish", "ResultPublish not found eventID", event.EventID)
		return nil, oty.ErrEventIDNotFound
	}

	ora := &OracleDB{*oracleStatus}

	if ora.Status.Status != oty.ResultPrePublished {
		olog.Error("ResultPublish", "ResultPublish can not abort", ora.Status.Status)
		return nil, oty.ErrResultPublishNotAllowed
	}

	updateStatus(ora, action.GetIndex(), action.fromaddr, oty.ResultPublished)
	ora.Result = event.Result
	ora.Source = event.Source

	if err := ora.save(action.db); err != nil {
		return nil, err
	}
	kv = append(kv, ora.GetKVSet()...)

	receiptLog := action.getOracleCommonRecipt(&ora.OracleStatus, oty.TyLogResultPublish)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// GetIndex returns index in block
func (action *oracleAction) GetIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

func (action *oracleAction) getOracleCommonRecipt(status *oty.OracleStatus, logTy int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = logTy
	o := &oty.ReceiptOracle{}
	o.EventID = status.EventID
	o.Status = status.GetStatus().Status
	o.Addr = status.Addr
	o.Type = status.Type
	o.PreStatus = status.GetPreStatus().Status
	log.Log = types.Encode(o)
	return log
}

func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		return nil, err
	}
	return value, nil
}

func isEventPublisher(addr string, db dbm.KV, isSolo bool) bool {
	if isSolo {
		return true
	}
	value, err := getManageKey(publishEventKey, db)
	if err != nil {
		olog.Error("OracleEventPublish", "publishEventKey", publishEventKey)
		return false
	}
	if value == nil {
		olog.Error("OracleEventPublish found nil value")
		return false
	}

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		olog.Error("OracleEventPublish", "Decode", value)
		return false
	}

	for _, op := range item.GetArr().Value {
		if op == addr {
			return true
		}
	}
	return false

}

func findOracleStatus(db dbm.KV, eventID string) (*oty.OracleStatus, error) {
	data, err := db.Get(Key(eventID))
	if err != nil {
		olog.Debug("findOracleStatus", "get", err)
		return nil, err
	}
	var oracleStatus oty.OracleStatus
	//decode
	err = types.Decode(data, &oracleStatus)
	if err != nil {
		olog.Debug("findOracleStatus", "decode", err)
		return nil, err
	}
	return &oracleStatus, nil
}

func updateStatus(ora *OracleDB, curIndex int64, addr string, status int32) {
	ora.PreStatus.Status = ora.Status.Status
	ora.PreStatus.OpAddr = ora.Status.OpAddr

	ora.Status.OpAddr = addr
	ora.Status.Status = status
}

// getOracleLisByIDs 获取eventinfo
func getOracleLisByIDs(db dbm.KV, infos *oty.QueryOracleInfos) (types.Message, error) {
	if len(infos.EventID) == 0 {
		return nil, oty.ErrParamNeedIDs
	}
	var status []*oty.OracleStatus
	for i := 0; i < len(infos.EventID); i++ {
		id := infos.EventID[i]
		game, err := findOracleStatus(db, id)
		if err != nil {
			return nil, err
		}
		status = append(status, game)
	}
	return &oty.ReplyOracleStatusList{Status: status}, nil
}

func getEventIDListByStatus(db dbm.KVDB, status int32, eventID string) (types.Message, error) {
	if status <= oty.NoEvent || status > oty.ResultPublished {
		return nil, oty.ErrParamStatusInvalid
	}
	data := &oty.ReceiptOracle{
		EventID: eventID,
		Status:  status,
	}
	return listData(db, data, oty.DefaultCount, oty.ListDESC)
}

func getEventIDListByAddrAndStatus(db dbm.KVDB, addr string, status int32, eventID string) (types.Message, error) {
	if status <= oty.NoEvent || status > oty.ResultPublished {
		return nil, oty.ErrParamStatusInvalid
	}
	if len(addr) == 0 {
		return nil, oty.ErrParamAddressMustnotEmpty
	}

	data := &oty.ReceiptOracle{
		EventID: eventID,
		Status:  status,
		Addr:    addr,
	}
	return listData(db, data, oty.DefaultCount, oty.ListDESC)
}

func getEventIDListByTypeAndStatus(db dbm.KVDB, ty string, status int32, eventID string) (types.Message, error) {
	if status <= oty.NoEvent || status > oty.ResultPublished {
		return nil, oty.ErrParamStatusInvalid
	}
	if len(ty) == 0 {
		return nil, oty.ErrParamTypeMustNotEmpty
	}
	data := &oty.ReceiptOracle{
		EventID: eventID,
		Status:  status,
		Type:    ty,
	}
	return listData(db, data, oty.DefaultCount, oty.ListDESC)
}

func listData(db dbm.KVDB, data *oty.ReceiptOracle, count, direction int32) (types.Message, error) {
	query := oty.NewTable(db).GetQuery(db)
	var primary []byte
	if len(data.EventID) > 0 {
		primary = []byte(data.EventID)
	}
	var rows []*table.Row
	var err error
	if len(data.Addr) > 0 {
		rows, err = query.List("addr_status", data, primary, count, direction)
		if err != nil {
			return nil, err
		}
	} else if len(data.Type) > 0 {
		rows, err = query.List("type_status", data, primary, count, direction)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = query.List("status", data, primary, count, direction)
		if err != nil {
			return nil, err
		}
	}

	var gameIds []string
	for _, row := range rows {
		gameIds = append(gameIds, string(row.Primary))
	}
	if len(gameIds) == 0 {
		return nil, types.ErrNotFound
	}
	return &oty.ReplyEventIDs{EventID: gameIds}, nil
}
