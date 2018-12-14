// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import "fmt"

func calcLotteryBuyPrefix(lotteryID string, addr string) []byte {
	key := fmt.Sprintf("LODB-lottery-buy:%s:%s", lotteryID, addr)
	return []byte(key)
}

func calcLotteryBuyRoundPrefix(lotteryID string, addr string, round int64) []byte {
	key := fmt.Sprintf("LODB-lottery-buy:%s:%s:%10d", lotteryID, addr, round)
	return []byte(key)
}

func calcLotteryBuyKey(lotteryID string, addr string, round int64, index int64) []byte {
	key := fmt.Sprintf("LODB-lottery-buy:%s:%s:%10d:%18d", lotteryID, addr, round, index)
	return []byte(key)
}

func calcLotteryDrawPrefix(lotteryID string) []byte {
	key := fmt.Sprintf("LODB-lottery-draw:%s", lotteryID)
	return []byte(key)
}

func calcLotteryDrawKey(lotteryID string, round int64) []byte {
	key := fmt.Sprintf("LODB-lottery-draw:%s:%10d", lotteryID, round)
	return []byte(key)
}

func calcLotteryKey(lotteryID string, status int32) []byte {
	key := fmt.Sprintf("LODB-lottery-:%d:%s", status, lotteryID)
	return []byte(key)
}

func calcLotteryGainPrefix(lotteryID string, addr string) []byte {
	key := fmt.Sprintf("LODB-lottery-gain:%s:%s", lotteryID, addr)
	return []byte(key)
}

func calcLotteryGainKey(lotteryID string, addr string, round int64) []byte {
	key := fmt.Sprintf("LODB-lottery-gain:%s:%s:%10d", lotteryID, addr, round)
	return []byte(key)
}
