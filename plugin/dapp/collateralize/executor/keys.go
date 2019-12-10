// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import "fmt"

func calcCollateralizeKey(collateralizeID string, index int64) []byte {
	key := fmt.Sprintf("LODB-collateralize-ID:%s:%018d", collateralizeID, index)
	return []byte(key)
}

func calcCollateralizeStatusPrefix(status int32) []byte {
	key := fmt.Sprintf("LODB-collateralize-status:%d", status)
	return []byte(key)
}

func calcCollateralizeStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-collateralize-status:%d:%018d", status, index)
	return []byte(key)
}

func calcCollateralizeAddrPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-collateralize-addr:%s", addr)
	return []byte(key)
}

func calcCollateralizeAddrKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-collateralize-addr:%s:%018d", addr, index)
	return []byte(key)
}

func calcCollateralizeRecordAddrPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-collateralize-record-addr:%s", addr)
	return []byte(key)
}

func calcCollateralizeRecordAddrKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-collateralize-record-addr:%s:%018d", addr, index)
	return []byte(key)
}

func calcCollateralizeRecordStatusPrefix(status string) []byte {
	key := fmt.Sprintf("LODB-collateralize-record-status:%s", status)
	return []byte(key)
}

func calcCollateralizeRecordStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-collateralize-record-status:%d:%018d", status, index)
	return []byte(key)
}
