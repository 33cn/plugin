// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

// retrieve errors
var (
	ErrRetrieveRepeatAddress   = errors.New("ErrRetrieveRepeatAddress")
	ErrRetrieveDefaultAddress  = errors.New("ErrRetrieveDefaultAddress")
	ErrRetrievePeriodLimit     = errors.New("ErrRetrievePeriodLimit")
	ErrRetrieveAmountLimit     = errors.New("ErrRetrieveAmountLimit")
	ErrRetrieveTimeweightLimit = errors.New("ErrRetrieveTimeweightLimit")
	ErrRetrievePrepareAddress  = errors.New("ErrRetrievePrepareAddress")
	ErrRetrievePerformAddress  = errors.New("ErrRetrievePerformAddress")
	ErrRetrieveCancelAddress   = errors.New("ErrRetrieveCancelAddress")
	ErrRetrieveStatus          = errors.New("ErrRetrieveStatus")
	ErrRetrieveRelateLimit     = errors.New("ErrRetrieveRelateLimit")
	ErrRetrieveRelation        = errors.New("ErrRetrieveRelation")
	ErrRetrieveNoBalance       = errors.New("ErrRetrieveNoBalance")
)
