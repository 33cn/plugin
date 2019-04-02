// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrInvalidTitle invalid commit msg title
	ErrInvalidTitle = errors.New("ErrInvalidTitle")
	// ErrTitleNotExist commit msg title not exist
	ErrTitleNotExist = errors.New("ErrTitleNotExist")
	// ErrNodeNotForTheTitle the node not match with title
	ErrNodeNotForTheTitle = errors.New("ErrNodeNotForTheTitle")
	// ErrParaBlockHashNoMatch block hash not match with before
	ErrParaBlockHashNoMatch = errors.New("ErrParaBlockHashNoMatch")
	// ErrParaMinerBaseIndex miner base index not 0
	ErrParaMinerBaseIndex = errors.New("ErrParaMinerBaseIndex")
	// ErrParaMinerTxType the 0 tx is not miner tx
	ErrParaMinerTxType = errors.New("ErrParaMinerTxType")
	// ErrParaEmptyMinerTx block no miner tx
	ErrParaEmptyMinerTx = errors.New("ErrParaEmptyMinerTx")
	// ErrParaMinerExecErr miner tx exec error
	ErrParaMinerExecErr = errors.New("ErrParaMinerExecErr")
	// ErrParaWaitingNewSeq para waiting main node new seq coming
	ErrParaWaitingNewSeq = errors.New("ErrParaWaitingNewSeq")
	// ErrParaCurHashNotMatch para curr main hash not match with pre, main node may switched
	ErrParaCurHashNotMatch = errors.New("ErrParaCurHashNotMatch")
	// ErrParaUnSupportNodeOper unsupport node operation
	ErrParaUnSupportNodeOper = errors.New("ErrParaUnSupportNodeOper")
	//ErrParaNodeAddrExisted node addr exist in group
	ErrParaNodeAddrExisted = errors.New("ErrParaNodeAddrExisted")
	//ErrParaNodeAddrNotExisted node addr not exist in group
	ErrParaNodeAddrNotExisted = errors.New("ErrParaNodeAddrNotExisted")
	//ErrParaManageNodesNotSet config manage node not set
	ErrParaManageNodesNotSet = errors.New("ErrParaManageNodesNotSet")
	//ErrParaNodeGroupNotSet para config node group not set by take over
	ErrParaNodeGroupNotSet = errors.New("ErrParaManageNodesNotSet")
	//ErrParaNodeGroupExisted para config group taked over alreay
	ErrParaNodeGroupExisted = errors.New("ErrParaNodesExisted")
	//ErrParaNodeGroupLastAddr last super node not be allow to quite
	ErrParaNodeGroupLastAddr = errors.New("ErrParaNodeGroupLastAddr")
	//ErrParaNodeVoteSelf vote self not allow
	ErrParaNodeVoteSelf = errors.New("ErrParaNodeVoteSelf")
)
