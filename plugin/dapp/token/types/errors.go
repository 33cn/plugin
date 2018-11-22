// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrTokenNameLen error token name length
	ErrTokenNameLen = errors.New("ErrTokenNameLength")
	// ErrTokenSymbolLen error token symbol length
	ErrTokenSymbolLen = errors.New("ErrTokenSymbolLength")
	// ErrTokenTotalOverflow error token total Overflow
	ErrTokenTotalOverflow = errors.New("ErrTokenTotalOverflow")
	// ErrTokenSymbolUpper error token total Overflow
	ErrTokenSymbolUpper = errors.New("ErrTokenSymbolUpper")
	// ErrTokenIntroLen error token introduction length
	ErrTokenIntroLen = errors.New("ErrTokenIntroductionLen")
	// ErrTokenExist error token symbol exist already
	ErrTokenExist = errors.New("ErrTokenSymbolExistAlready")
	// ErrTokenNotPrecreated error token not pre created
	ErrTokenNotPrecreated = errors.New("ErrTokenNotPrecreated")
	// ErrTokenCreatedApprover error token created approver
	ErrTokenCreatedApprover = errors.New("ErrTokenCreatedApprover")
	// ErrTokenRevoker error token revoker
	ErrTokenRevoker = errors.New("ErrTokenRevoker")
	// ErrTokenCanotRevoked error token canot revoked with wrong status
	ErrTokenCanotRevoked = errors.New("ErrTokenCanotRevokedWithWrongStatus")
	// ErrTokenOwner error token symbol owner not match
	ErrTokenOwner = errors.New("ErrTokenSymbolOwnerNotMatch")
	// ErrTokenHavePrecreated error owner have token pre create yet
	ErrTokenHavePrecreated = errors.New("ErrOwnerHaveTokenPrecreateYet")
	// ErrTokenBlacklist error token blacklist
	ErrTokenBlacklist = errors.New("ErrTokenBlacklist")
	// ErrTokenNotExist error token symbol not exist
	ErrTokenNotExist = errors.New("ErrTokenSymbolNotExist")
)
