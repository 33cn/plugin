// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"strings"

	"github.com/33cn/chain33/types"
)

const (
	paraNodeIDUnifyPrefix = "mavl-paracross-title-node"
)

var (
	title                     string
	titleHeight               string
	managerConfigNodes        string //manager 合约配置的nodes
	paraConfigNodes           string //平行链自组织配置的nodes，最初是从manager同步过来
	paraConfigNodeAddr        string //平行链配置节点账户
	paraNodeGroupStatusAddrs  string //正在申请的addrs
	paraNodeIDPrefix          string
	paraNodeGroupIDPrefix     string
	localTx                   string
	localTitle                string
	localTitleHeight          string
	localAssetKey             string
	localNodeTitleStatus      string
	localNodeTitleDone        string
	localNodeGroupStatusTitle string

	paraSelfConsensStages        string
	paraSelfConsensStageIDPrefix string
)

func setPrefix() {
	title = "mavl-paracross-title-"
	titleHeight = "mavl-paracross-titleHeight-"
	managerConfigNodes = "paracross-nodes-"
	paraConfigNodes = "mavl-paracross-nodes-title-"
	paraConfigNodeAddr = "mavl-paracross-nodes-titleAddr-"
	paraNodeGroupStatusAddrs = "mavl-paracross-nodegroup-apply-title-"
	paraNodeIDPrefix = "mavl-paracross-title-nodeid-"
	paraNodeGroupIDPrefix = "mavl-paracross-title-nodegroupid-"

	paraSelfConsensStages = "mavl-paracross-selfconsens-stages-"
	paraSelfConsensStageIDPrefix = "mavl-paracross-selfconsens-id-"

	localTx = "LODB-paracross-titleHeightAddr-"
	localTitle = "LODB-paracross-title-"
	localTitleHeight = "LODB-paracross-titleHeight-"
	localAssetKey = "LODB-paracross-asset-"

	localNodeTitleStatus = "LODB-paracross-nodesTitleStatus-"
	localNodeTitleDone = "LODB-paracross-nodesTitleDone-"

	localNodeGroupStatusTitle = "LODB-paracross-nodegroupStatusTitle-"

}

func calcTitleKey(t string) []byte {
	return []byte(fmt.Sprintf(title+"%s", t))
}

func calcTitleHeightKey(title string, height int64) []byte {
	return []byte(fmt.Sprintf(titleHeight+"%s-%d", title, height))
}

func calcLocalHeightKey(title string, height int64) []byte {
	return []byte(fmt.Sprintf(localTitleHeight+"%s-%d", title, height))
}

func calcManageConfigNodesKey(title string) []byte {
	key := managerConfigNodes + title
	return []byte(types.ManageKey(key))
}

func calcParaNodeGroupAddrsKey(title string) []byte {
	return []byte(fmt.Sprintf(paraConfigNodes+"%s", title))
}

func calcParaNodeAddrKey(title string, addr string) []byte {
	return []byte(fmt.Sprintf(paraConfigNodeAddr+"%s-%s", title, addr))
}

func calcParaNodeGroupStatusKey(title string) []byte {
	return []byte(fmt.Sprintf(paraNodeGroupStatusAddrs+"%s", title))
}

func calcParaNodeIDKey(title, hash string) string {
	return fmt.Sprintf(paraNodeIDPrefix+"%s-%s", title, hash)
}

func calcParaNodeGroupIDKey(title, hash string) string {
	return fmt.Sprintf(paraNodeGroupIDPrefix+"%s-%s", title, hash)
}

func calcParaSelfConsStagesKey() []byte {
	return []byte(fmt.Sprintf(paraSelfConsensStages))
}

func calcParaSelfConsensStageIDKey(hash string) string {
	return fmt.Sprintf(paraSelfConsensStageIDPrefix+"%s", hash)
}

func getParaNodeIDSuffix(id string) string {
	if !strings.HasPrefix(id, paraNodeIDUnifyPrefix) {
		return id
	}

	ok, txID, ids := getRealTxHashID(id)
	if ok {
		return txID
	}

	//对于nodegroup 创建的"mavl-paracross-title-nodegroupid-user.p.para.-0xb6cd0274587...a61e444e9f848a4c02d7b-1"特殊场景
	if len(ids) > 1 {
		txID = ids[len(ids)-2] + "-" + txID
		if strings.HasPrefix(txID, "0x") {
			return txID
		}
	}
	return id
}

func getRealTxHashID(id string) (bool, string, []string) {
	ids := strings.Split(id, "-")
	txID := ids[len(ids)-1]
	if strings.HasPrefix(txID, "0x") {
		return true, txID, ids
	}
	return false, txID, ids
}

func calcLocalTxKey(title string, height int64, addr string) []byte {
	return []byte(fmt.Sprintf(localTx+"%s-%012-%s", title, height, addr))
}

func calcLocalTitleKey(title string) []byte {
	return []byte(fmt.Sprintf(localTitle+"%s", title))
}

func calcLocalTitlePrefix() []byte {
	return []byte(localTitle)
}

func calcLocalAssetKey(hash []byte) []byte {
	return []byte(fmt.Sprintf(localAssetKey+"%s", hash))
}

func calcLocalNodeTitleStatus(title string, status int32, id string) []byte {
	return []byte(fmt.Sprintf(localNodeTitleStatus+"%s-%02d-%s", title, status, id))
}

func calcLocalNodeStatusPrefix(title string, status int32) []byte {
	return []byte(fmt.Sprintf(localNodeTitleStatus+"%s-%02d-", title, status))
}

func calcLocalNodeTitlePrefix(title string) []byte {
	return []byte(fmt.Sprintf(localNodeTitleStatus+"%s-", title))
}

func calcLocalNodeTitleDone(title, addr string) []byte {
	return []byte(fmt.Sprintf(localNodeTitleDone+"%s-%s", title, addr))
}

func calcLocalNodeGroupStatusTitle(status int32, title, id string) []byte {
	return []byte(fmt.Sprintf(localNodeGroupStatusTitle+"%02d-%s-%s", status, title, id))
}

func calcLocalNodeGroupStatusPrefix(status int32) []byte {
	return []byte(fmt.Sprintf(localNodeGroupStatusTitle+"%02d-", status))
}

func calcLocalNodeGroupAllPrefix() []byte {
	return []byte(fmt.Sprintf(localNodeGroupStatusTitle))
}
