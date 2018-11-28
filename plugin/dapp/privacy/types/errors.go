// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

// ErrXXX defines error message
var (
	ErrGeFromBytesVartime    = errors.New("ErrGeFromBytesVartime")
	ErrPrivacyNotEnabled     = errors.New("ErrPrivacyNotEnabled")
	ErrPrivacyTxFeeNotEnough = errors.New("ErrPrivacyTxFeeNotEnough")
	ErrRescanFlagScaning     = errors.New("ErrRescanFlagScaning")
	ErrNoUTXORec4Token       = errors.New("ErrNoUTXORec4Token")
	ErrNoUTXORec4Amount      = errors.New("ErrNoUTXORec4Amount")
	ErrNotEnoughUTXOs        = errors.New("ErrNotEnoughUTXOs")
	ErrNoSuchPrivacyTX       = errors.New("ErrNoSuchPrivacyTX")
	ErrDoubleSpendOccur      = errors.New("ErrDoubleSpendOccur")
	ErrOutputIndex           = errors.New("ErrOutputIndex")
	ErrPubkeysOfUTXO         = errors.New("ErrPubkeysOfUTXO")
	ErrRecoverUTXO           = errors.New("ErrRecoverUTXO")
)
