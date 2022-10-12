// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/address"

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

	paraBindMinderNode string
	paraBindMinderAddr string

	//监督节点
	paraSupervisionNodes            string
	paraSupervisionNodeIDPrefix     string
	localSupervisionNodeStatusTitle string
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

	//bind miner,node和miner角色要区分开，不然如果miner也是node，就会混淆
	paraBindMinderNode = "mavl-paracross-bindminernode-"
	paraBindMinderAddr = "mavl-paracross-bindmineraddr-"

	localTx = "LODB-paracross-titleHeightAddr-"
	localTitle = "LODB-paracross-title-"
	localTitleHeight = "LODB-paracross-titleHeight-"
	localAssetKey = "LODB-paracross-asset-"

	localNodeTitleStatus = "LODB-paracross-nodesTitleStatus-"
	localNodeTitleDone = "LODB-paracross-nodesTitleDone-"

	localNodeGroupStatusTitle = "LODB-paracross-nodegroupStatusTitle-"

	paraSupervisionNodes = "mavl-paracross-supervision-nodes-title-"
	paraSupervisionNodeIDPrefix = "mavl-paracross-title-nodeid-supervision-"
	localSupervisionNodeStatusTitle = "LODB-paracross-supervision-nodeStatusTitle-"
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
	return []byte(fmt.Sprintf(paraConfigNodeAddr+"%s-%s", title, address.FormatAddrKey(addr)))
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
	return []byte(fmt.Sprintf(localTx+"%s-%012-%s", title, height, address.FormatAddrKey(addr)))
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
	return []byte(fmt.Sprintf(localNodeTitleDone+"%s-%s", title, address.FormatAddrKey(addr)))
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

/////bind miner

//统计共识节点绑定挖矿地址总数量
//key: prefix-nodeAddr  val: bindTotalCount
func calcParaNodeBindMinerCount(node string) []byte {
	return []byte(fmt.Sprintf(paraBindMinderNode+"%s", node))
}

//记录共识节点某一索引绑定的挖矿地址，一一对应，以此地址获取更详细信息
//key: prefix-nodeAddr-index   val:bindMinerAddr
func calcParaNodeBindMinerIndex(node string, index int64) []byte {
	return []byte(fmt.Sprintf(paraBindMinderNode+"%s-%d", node, index))
}

//记录node和miner bind详细信息
//key: prefix-nodeAddr-miner  val:miner detail info
func calcParaBindMinerAddr(node, miner string) []byte {
	return []byte(fmt.Sprintf(paraBindMinderNode+"%s-%s", node, address.FormatAddrKey(miner)))
}

//key: prefix-minerAddr  val: node list
func calcParaMinerBindNodeList(miner string) []byte {
	return []byte(fmt.Sprintf(paraBindMinderAddr+"%s", address.FormatAddrKey(miner)))
}

/////supervision
func calcParaSupervisionNodeGroupAddrsKey(title string) []byte {
	return []byte(fmt.Sprintf(paraSupervisionNodes+"%s", title))
}

func calcParaSupervisionNodeIDKey(title, hash string) string {
	return fmt.Sprintf(paraSupervisionNodeIDPrefix+"%s-%s", title, hash)
}

func calcLocalSupervisionNodeStatusTitle(title string, status int32, addr, id string) []byte {
	return []byte(fmt.Sprintf(localSupervisionNodeStatusTitle+"%s-%02d-%s-%s-%s", title, status, address.FormatAddrKey(addr), id))
}

func calcLocalSupervisionNodeStatusTitlePrefix(title string, status int32) []byte {
	return []byte(fmt.Sprintf(localSupervisionNodeStatusTitle+"%s-%02d", title, status))
}

func calcLocalSupervisionNodeStatusTitleAllPrefix(title string) []byte {
	return []byte(fmt.Sprintf(localSupervisionNodeStatusTitle+"%s-", title))
}
