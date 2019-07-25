// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

var (
	idPrefix          = "mavl-" + auty.AutonomyX + "-"
	votesRecordPrefix = idPrefix + "vote" + "-"

	localPrefix = "LODB-" + auty.AutonomyX + "-"
)

func votesRecord(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", votesRecordPrefix, txHash))
}

var (
	// board
	boardPrefix            = idPrefix + "board" + "-"
	localBoardPrefix       = localPrefix + "board" + "-"
	boardVotesRecordPrefix = boardPrefix + "vote" + "-"
)

func activeBoardID() []byte {
	return []byte(boardPrefix)
}

func propBoardID(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", boardPrefix, txHash))
}

func boardVotesRecord(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", boardVotesRecordPrefix, txHash))
}

func calcBoardKey4StatusHeight(status int32, heightindex string) []byte {
	key := fmt.Sprintf(localBoardPrefix+"%d-"+"%s", status, heightindex)
	return []byte(key)
}

var (
	// project
	projectPrefix      = idPrefix + "project" + "-"
	localProjectPrefix = localPrefix + "project" + "-"
)

func propProjectID(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", projectPrefix, txHash))
}

func calcProjectKey4StatusHeight(status int32, heightindex string) []byte {
	key := fmt.Sprintf(localProjectPrefix+"%d-"+"%s", status, heightindex)
	return []byte(key)
}

var (
	// rule
	rulePrefix      = idPrefix + "rule" + "-"
	localRulePrefix = localPrefix + "rule" + "-"
)

func activeRuleID() []byte {
	return []byte(rulePrefix)
}

func propRuleID(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", rulePrefix, txHash))
}

func calcRuleKey4StatusHeight(status int32, heightindex string) []byte {
	key := fmt.Sprintf(localRulePrefix+"%d-"+"%s", status, heightindex)
	return []byte(key)
}

var (
	// comment
	localCommentPrefix = localPrefix + "cmt" + "-"
)

func calcCommentHeight(ID, heightindex string) []byte {
	key := fmt.Sprintf(localCommentPrefix+"%s-"+"%s", ID, heightindex)
	return []byte(key)
}
