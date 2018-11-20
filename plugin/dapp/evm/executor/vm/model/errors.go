// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package model

import "errors"

var (
	// ErrOutOfGas                 out of gas
	ErrOutOfGas = errors.New("out of gas")
	// ErrCodeStoreOutOfGas        contract creation code storage out of gas
	ErrCodeStoreOutOfGas = errors.New("contract creation code storage out of gas")
	// ErrDepth                    max call depth exceeded
	ErrDepth = errors.New("max call depth exceeded")
	// ErrInsufficientBalance      insufficient balance for transfer
	ErrInsufficientBalance = errors.New("insufficient balance for transfer")
	// ErrContractAddressCollision contract address collision
	ErrContractAddressCollision = errors.New("contract address collision")
	// ErrGasUintOverflow          gas uint64 overflow
	ErrGasUintOverflow = errors.New("gas uint64 overflow")
	// ErrAddrNotExists            address not exists
	ErrAddrNotExists = errors.New("address not exists")
	// ErrTransferBetweenContracts transferring between contracts not supports
	ErrTransferBetweenContracts = errors.New("transferring between contracts not supports")
	// ErrTransferBetweenEOA       transferring between external accounts not supports
	ErrTransferBetweenEOA = errors.New("transferring between external accounts not supports")
	// ErrNoCreator                contract has no creator information
	ErrNoCreator = errors.New("contract has no creator information")
	// ErrDestruct                 contract has been destructed
	ErrDestruct = errors.New("contract has been destructed")

	// ErrWriteProtection       evm: write protection
	ErrWriteProtection = errors.New("evm: write protection")
	// ErrReturnDataOutOfBounds evm: return data out of bounds
	ErrReturnDataOutOfBounds = errors.New("evm: return data out of bounds")
	// ErrExecutionReverted     evm: execution reverted
	ErrExecutionReverted = errors.New("evm: execution reverted")
	// ErrMaxCodeSizeExceeded   evm: max code size exceeded
	ErrMaxCodeSizeExceeded = errors.New("evm: max code size exceeded")

	// ErrNoCoinsAccount no coins account in executor!
	ErrNoCoinsAccount = errors.New("no coins account in executor")
)
