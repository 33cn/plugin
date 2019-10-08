// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import "fmt"

func calcCollateralizeKey(CollateralizeID string) []byte {
	key := fmt.Sprintf("LODB-Collateralize-create:%s", CollateralizeID)
	return []byte(key)
}

func calcCollateralizeBorrowPrefix(CollateralizeID string, addr string) []byte {
	key := fmt.Sprintf("LODB-Collateralize-borrow:%s:%s", CollateralizeID, addr)
	return []byte(key)
}

func calcCollateralizeBorrowKey(CollateralizeID string, addr string) []byte {
	key := fmt.Sprintf("LODB-Collateralize-buy:%s:%s:%18d", CollateralizeID, addr)
	return []byte(key)
}

func calcCollateralizeRepayPrefix(CollateralizeID string, addr string) []byte {
	key := fmt.Sprintf("LODB-Collateralize-repay:%s:%s", CollateralizeID, addr)
	return []byte(key)
}

func calcCollateralizeRepayKey(CollateralizeID string) []byte {
	key := fmt.Sprintf("LODB-Collateralize-repay:%s:%10d", CollateralizeID)
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
