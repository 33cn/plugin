// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

var (
	idPrefix = "mavl-" + auty.AutonomyX + "-"
	votesRecordPrefix = idPrefix + "vote" + "-"

	localPrefix = "LOCDB" + auty.AutonomyX + "-"
)

var (
	// board
	boardPrefix = idPrefix + "board" + "-"

)

func propBoardID(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", boardPrefix, txHash))
}

func VotesRecord(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", votesRecordPrefix, txHash))
}

func calcBoardKey4StatusHeight(status int32, heightindex string) []byte {
	key := fmt.Sprintf(localPrefix + "%d-" +"%s", status, heightindex)
	return []byte(key)
}
