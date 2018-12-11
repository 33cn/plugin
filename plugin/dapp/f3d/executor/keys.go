/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import "fmt"

func calcF3dBuyRound(round int64, addr, index string) []byte {
	key := fmt.Sprintf("LODB-f3d-buy:%010d:%s:%s", round, addr, index)
	return []byte(key)
}

func calcF3dBuyPrefix(round int64, addr string) []byte {
	key := fmt.Sprintf("LODB-f3d-buy:%010d:%s", round, addr)
	return []byte(key)
}

func calcF3dAddrRound(round int64, addr string) []byte {
	key := fmt.Sprintf("LODB-f3d-AddrInfos:%010d:%s", round, addr)
	return []byte(key)
}

func calcF3dAddrPrefix(round int64) []byte {
	key := fmt.Sprintf("LODB-f3d-AddrInfos:%010d", round)
	return []byte(key)
}
