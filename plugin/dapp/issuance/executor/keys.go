// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import "fmt"

func calcIssuanceKey(issuanceID string, index int64) []byte {
	key := fmt.Sprintf("LODB-Issuance-ID:%s:%018d", issuanceID, index)
	return []byte(key)
}

func calcIssuanceStatusPrefix(status int32) []byte {
	key := fmt.Sprintf("LODB-Issuance-status-index:%d", status)
	return []byte(key)
}

func calcIssuanceStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-Issuance-status:%d:%018d", status, index)
	return []byte(key)
}

func calcIssuanceRecordAddrPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-Issuance-record-addr:%s", addr)
	return []byte(key)
}

func calcIssuanceRecordAddrKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-Issuance-record-addr:%s:%018d", addr, index)
	return []byte(key)
}

func calcIssuancePriceKey(time string) []byte {
	key := fmt.Sprintf("LODB-Issuance-price:%s", time)
	return []byte(key)
}

func calcIssuanceLatestPriceKey() []byte {
	key := fmt.Sprintf("LODB-Issuance-latest-price")
	return []byte(key)
}

func calcIssuanceRecordStatusPrefix(status string) []byte {
	key := fmt.Sprintf("LODB-Issuance-record-status:%s", status)
	return []byte(key)
}

func calcIssuanceRecordStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-Issuance-record-status:%d:%018d", status, index)
	return []byte(key)
}