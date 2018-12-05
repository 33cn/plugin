// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

// PrivacyAccountResult display privacy account result
type PrivacyAccountResult struct {
	Token         string `json:"Token,omitempty"`
	Txhash        string `json:"Txhash,omitempty"`
	OutIndex      int32  `json:"OutIndex,omitempty"`
	Amount        string `json:"Amount,omitempty"`
	OnetimePubKey string `json:"OnetimePubKey,omitempty"`
}

// PrivacyAccountInfoResult display privacy account information result
type PrivacyAccountInfoResult struct {
	AvailableDetail []*PrivacyAccountResult `json:"AvailableDetail,omitempty"`
	FrozenDetail    []*PrivacyAccountResult `json:"FrozenDetail,omitempty"`
	AvailableAmount string                  `json:"AvailableAmount,omitempty"`
	FrozenAmount    string                  `json:"FrozenAmount,omitempty"`
	TotalAmount     string                  `json:"TotalAmount,omitempty"`
}

// PrivacyAccountSpendResult display privacy account spend result
type PrivacyAccountSpendResult struct {
	Txhash string                  `json:"Txhash,omitempty"`
	Res    []*PrivacyAccountResult `json:"Spend,omitempty"`
}

// ShowRescanResult display rescan utxos result
type ShowRescanResult struct {
	Addr       string `json:"addr"`
	FlagString string `json:"FlagString"`
}

type showRescanResults struct {
	RescanResults []*ShowRescanResult `json:"ShowRescanResults,omitempty"`
}

// ShowEnablePrivacy display enable privacy
type ShowEnablePrivacy struct {
	Results []*ShowPriAddrResult `json:"results"`
}

// ShowPriAddrResult display privacy address result
type ShowPriAddrResult struct {
	Addr string `json:"addr"`
	IsOK bool   `json:"IsOK"`
	Msg  string `json:"msg"`
}
