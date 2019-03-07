// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/types"
)

var (
	title                string
	titleHeight          string
	managerConfigNodes   string //manager 合约配置的nodes
	paraConfigNodes      string //平行链自组织配置的nodes，最初是从manager同步过来
	paraConfigNodeAddr   string //平行链配置节点账户
	localTx              string
	localTitle           string
	localTitleHeight     string
	localAssetKey        string
	localNodeTitleStatus string
	localNodeTitleDone   string
)

func setPrefix() {
	title = "mavl-paracross-title-"
	titleHeight = "mavl-paracross-titleHeight-"
	managerConfigNodes = "paracross-nodes-"
	paraConfigNodes = "mavl-paracross-nodes-title-"
	paraConfigNodeAddr = "mavl-paracross-nodes-titleAddr-"
	localTx = "LODB-paracross-titleHeightAddr-"
	localTitle = "LODB-paracross-title-"
	localTitleHeight = "LODB-paracross-titleHeight-"
	localAssetKey = "LODB-paracross-asset-"

	localNodeTitleStatus = "LODB-paracross-nodesTitleStatus-"
	localNodeTitleDone = "LODB-paracross-nodesTitleDone-"

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

func calcParaNodeGroupKey(title string) []byte {
	return []byte(fmt.Sprintf(paraConfigNodes+"%s", title))
}

func calcParaNodeAddrKey(title string, addr string) []byte {
	return []byte(fmt.Sprintf(paraConfigNodeAddr+"%s-%s", title, addr))
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

func calcLocalNodeTitleStatus(title, addr string, status int32) []byte {
	return []byte(fmt.Sprintf(localNodeTitleStatus+"%s-%02d-%s", title, status, addr))
}

func calcLocalNodeStatusPrefix(title string, status int32) []byte {
	return []byte(fmt.Sprintf(localNodeTitleStatus+"%s-%02d", title, status))
}

func calcLocalNodeTitleDone(title, addr string) []byte {
	return []byte(fmt.Sprintf(localNodeTitleDone+"%s-%s", title, addr))
}
