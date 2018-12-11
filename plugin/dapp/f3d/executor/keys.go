/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import "fmt"

func calcF3dBuyRound(round int64, addr string, index int64) []byte {
	key := fmt.Sprintf("LODB-f3d-buy:%010d:%s:%018d", round, addr, index)
	return []byte(key)
}

func calcF3dBuyPrefix(round int64, addr string) []byte {
	key := fmt.Sprintf("LODB-f3d-buy:%010d:%s:", round, addr)
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

func calcF3dStartRound(round int64) []byte {
	key := fmt.Sprintf("LODB-f3d-start:%010d", round)
	return []byte(key)
}

func calcF3dStartPrefix() []byte {
	key := fmt.Sprintf("LODB-f3d-start:")
	return []byte(key)
}

func calcF3dDrawRound(round int64) []byte {
	key := fmt.Sprintf("LODB-f3d-draw:%010d", round)
	return []byte(key)
}

func calcF3dDrawPrefix() []byte {
	key := fmt.Sprintf("LODB-f3d-draw:")
	return []byte(key)
}