// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

const (
	// ActionTransfer for transfer
	ActionTransfer = 4
	// ActionGenesis for genesis
	ActionGenesis = 5
	// ActionWithdraw for Withdraw
	ActionWithdraw = 6
	// EvmxgoActionTransferToExec for evmxgo transfer to exec
	EvmxgoActionTransferToExec = 11
	// EvmxgoActionMint for evmxgo mint
	EvmxgoActionMint = 12
	// EvmxgoActionBurn for evmxgo burn
	EvmxgoActionBurn = 13
	// EvmxgoActionMint for evmxgo mint map
	EvmxgoActionMintMap = 14
	EvmxgoActionBurnMap = 15
)

const (
	// TyLogEvmxgoTransfer log for token tranfer
	TyLogEvmxgoTransfer = 313
	// TyLogEvmxgoDeposit log for token deposit
	TyLogEvmxgoDeposit = 315
	// TyLogEvmxgoExecTransfer log for token exec transfer
	TyLogEvmxgoExecTransfer = 316
	// TyLogEvmxgoExecWithdraw log for token exec withdraw
	TyLogEvmxgoExecWithdraw = 317
	// TyLogEvmxgoExecDeposit log for token exec deposit
	TyLogEvmxgoExecDeposit = 318
	// TyLogEvmxgoExecFrozen log for token exec frozen
	TyLogEvmxgoExecFrozen = 319
	// TyLogEvmxgoExecActive log for token exec active
	TyLogEvmxgoExecActive = 320
	// TyLogEvmxgoGenesisTransfer log for token genesis rransfer
	TyLogEvmxgoGenesisTransfer = 321
	// TyLogEvmxgoGenesisDeposit log for token genesis deposit
	TyLogEvmxgoGenesisDeposit = 322
	// TyLogEvmxgoMint log for evmxgo mint
	TyLogEvmxgoMint = 323
	// TyLogEvmxgoBurn log for evmxgo burn
	TyLogEvmxgoBurn = 324
	// TyLogEvmxgoMint log for evmxgo mint
	TyLogEvmxgoMintMap = 325
	// TyLogEvmxgoBurn log for evmxgo burn
	TyLogEvmxgoBurnMap = 326
)
