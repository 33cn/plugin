// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import "fmt"

func calcCollateralizeKey(CollateralizeID string) []byte {
	key := fmt.Sprintf("LODB-Collateralize-ID:%s", CollateralizeID)
	return []byte(key)
}

func calcCollateralizeStatusPrefix(status int32) []byte {
	key := fmt.Sprintf("LODB-Collateralize-status-index:%d", status)
	return []byte(key)
}

func calcCollateralizeStatusKey(status int32, index int64) []byte {
	key := fmt.Sprintf("LODB-Collateralize-status:%d:%018d", status, index)
	return []byte(key)
}

func calcCollateralizeAddrPrefix(addr string) []byte {
	key := fmt.Sprintf("LODB-Collateralize-addr:%s", addr)
	return []byte(key)
}

func calcCollateralizeAddrKey(addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-Collateralize-addr:%s:%018d", addr, index)
	return []byte(key)
}

func calcCollateralizePriceKey(time string) []byte {
	key := fmt.Sprintf("LODB-Collateralize-price:%s", time)
	return []byte(key)
}

func calcCollateralizeLatestPriceKey() []byte {
	key := fmt.Sprintf("LODB-Collateralize-latest-price")
	return []byte(key)
}
