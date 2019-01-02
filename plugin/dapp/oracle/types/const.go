/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package types

import "errors"

var (
	// OracleX oracle name
	OracleX = "oracle"
)

// oracle action type
const (
	ActionEventPublish = iota + 1 //事件发布
	ActionResultPrePublish
	ActionResultPublish
	ActionEventAbort
	ActionResultAbort
)

// oracle status
const (
	NoEvent = iota
	EventPublished
	EventAborted
	ResultPrePublished
	ResultAborted
	ResultPublished
)

// log type define
const (
	TyLogEventPublish     = 810
	TyLogEventAbort       = 811
	TyLogResultPrePublish = 812
	TyLogResultAbort      = 813
	TyLogResultPublish    = 814
)

// executor action and function define
const (
	// FuncNameQueryOracleListByIDs 根据ids查询OracleStatus
	FuncNameQueryOracleListByIDs = "QueryOraclesByIDs"
	// FuncNameQueryEventIDByStatus 根据状态查询eventID
	FuncNameQueryEventIDByStatus = "QueryEventIDsByStatus"
	// FuncNameQueryEventIDByAddrAndStatus 根据创建者地址和状态查询eventID
	FuncNameQueryEventIDByAddrAndStatus = "QueryEventIDsByAddrAndStatus"
	// FuncNameQueryEventIDByTypeAndStatus 根据事件类型和状态查询eventID
	FuncNameQueryEventIDByTypeAndStatus = "QueryEventIDsByTypeAndStatus"
	// CreateEventPublishTx 创建发布事件交易
	CreateEventPublishTx = "EventPublish"
	// CreateAbortEventPublishTx 创建取消发布事件交易
	CreateAbortEventPublishTx = "EventAbort"
	// CreatePrePublishResultTx 创建预发布事件结果交易
	CreatePrePublishResultTx = "ResultPrePublish"
	// CreateAbortResultPrePublishTx 创建取消预发布的事件结果交易
	CreateAbortResultPrePublishTx = "ResultAbort"
	// CreateResultPublishTx 创建预发布事件结果交易
	CreateResultPublishTx = "ResultPublish"
)

// query param define
const (
	// ListDESC 降序
	ListDESC = int32(0)
	// DefaultCount 默认一次取多少条记录
	DefaultCount = int32(20)
)

// Errors for oracle
var (
	ErrTimeMustBeFuture           = errors.New("ErrTimeMustBeFuture")
	ErrNoPrivilege                = errors.New("ErrNoPrivilege")
	ErrOracleRepeatHash           = errors.New("ErrOracleRepeatHash")
	ErrEventIDNotFound            = errors.New("ErrEventIDNotFound")
	ErrEventAbortNotAllowed       = errors.New("ErrEventAbortNotAllowed")
	ErrResultPrePublishNotAllowed = errors.New("ErrResultPrePublishNotAllowed")
	ErrPrePublishAbortNotAllowed  = errors.New("ErrPrePublishAbortNotAllowed")
	ErrResultPublishNotAllowed    = errors.New("ErrResultPublishNotAllowed")
	ErrParamNeedIDs               = errors.New("ErrParamNeedIDs")
	ErrParamStatusInvalid         = errors.New("ErrParamStatusInvalid")
	ErrParamAddressMustnotEmpty   = errors.New("ErrParamAddressMustnotEmpty")
	ErrParamTypeMustNotEmpty      = errors.New("ErrParamTypeMustNotEmpty")
)
