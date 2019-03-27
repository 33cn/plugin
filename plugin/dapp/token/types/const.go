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
	// TokenActionPreCreate for token pre create
	TokenActionPreCreate = 7
	// TokenActionFinishCreate for token finish create
	TokenActionFinishCreate = 8
	// TokenActionRevokeCreate for token revoke create
	TokenActionRevokeCreate = 9
	// TokenActionTransferToExec for token transfer to exec
	TokenActionTransferToExec = 11
	// TokenActionMint for token mint
	TokenActionMint = 12
	// TokenActionBurn for token burn
	TokenActionBurn = 13
)

// token status
const (
	// TokenStatusPreCreated token pre create status
	TokenStatusPreCreated = iota
	// TokenStatusCreated token create status
	TokenStatusCreated
	// TokenStatusCreateRevoked token revoke status
	TokenStatusCreateRevoked
)

var (
	// TokenX token name
	TokenX = "token"

	// ForkTokenBlackListX fork const
	ForkTokenBlackListX = "ForkTokenBlackList"
	// ForkBadTokenSymbolX fork const
	ForkBadTokenSymbolX = "ForkBadTokenSymbol"
	// ForkTokenPriceX fork const
	ForkTokenPriceX = "ForkTokenPrice"
	// ForkTokenSymbolWithNumberX fork const
	ForkTokenSymbolWithNumberX = "ForkTokenSymbolWithNumber"
	// ForkTokenCheckX  fork check impl bug
	ForkTokenCheckX = "ForkTokenCheck"
)

const (
	// TyLogPreCreateToken log for pre create token
	TyLogPreCreateToken = 211
	// TyLogFinishCreateToken log for finish create token
	TyLogFinishCreateToken = 212
	// TyLogRevokeCreateToken log for revoke create token
	TyLogRevokeCreateToken = 213
	// TyLogTokenTransfer log for token tranfer
	TyLogTokenTransfer = 313
	// TyLogTokenGenesis log for token genesis
	TyLogTokenGenesis = 314
	// TyLogTokenDeposit log for token deposit
	TyLogTokenDeposit = 315
	// TyLogTokenExecTransfer log for token exec transfer
	TyLogTokenExecTransfer = 316
	// TyLogTokenExecWithdraw log for token exec withdraw
	TyLogTokenExecWithdraw = 317
	// TyLogTokenExecDeposit log for token exec deposit
	TyLogTokenExecDeposit = 318
	// TyLogTokenExecFrozen log for token exec frozen
	TyLogTokenExecFrozen = 319
	// TyLogTokenExecActive log for token exec active
	TyLogTokenExecActive = 320
	// TyLogTokenGenesisTransfer log for token genesis rransfer
	TyLogTokenGenesisTransfer = 321
	// TyLogTokenGenesisDeposit log for token genesis deposit
	TyLogTokenGenesisDeposit = 322
	// TyLogTokenMint log for token mint
	TyLogTokenMint = 323
	// TyLogTokenBurn log for token burn
	TyLogTokenBurn = 324
)

const (
	// TokenNameLenLimit token name length limit
	TokenNameLenLimit = 128
	// TokenSymbolLenLimit token symbol length limit
	TokenSymbolLenLimit = 16
	// TokenIntroLenLimit token introduction length limit
	TokenIntroLenLimit = 1024
)

const (
	// CategoryMintBurnSupport support mint & burn
	CategoryMintBurnSupport = 1 << iota
)
