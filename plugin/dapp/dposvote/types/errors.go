// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

// Errors for Dpos
var (
	ErrNoSuchVote               = errors.New("ErrNoSuchVote")
	ErrNotEnoughVotes           = errors.New("ErrNotEnoughVotes")
	ErrCandidatorExist          = errors.New("ErrCandidatorExist")
	ErrCandidatorInvalidStatus  = errors.New("ErrCandidatorInvalidStatus")
	ErrCandidatorNotExist       = errors.New("ErrCandidatorNotExist")
	ErrCandidatorNotEnough      = errors.New("ErrCandidatorNotEnough")
	ErrCandidatorNotLegal       = errors.New("ErrCandidatorNotLegal")
	ErrVrfMNotRegisted          = errors.New("ErrVrfMNotRegisted")
	ErrVrfMAlreadyRegisted      = errors.New("ErrVrfMAlreadyRegisted")
	ErrVrfRPAlreadyRegisted     = errors.New("ErrVrfRPAlreadyRegisted")
	ErrNoPrivilege              = errors.New("ErrNoPrivilege")
	ErrParamStatusInvalid       = errors.New("ErrParamStatusInvalid")
	ErrParamAddressMustnotEmpty = errors.New("ErrParamAddressMustnotEmpty")
	ErrSaveTable                = errors.New("ErrSaveTable")
)
