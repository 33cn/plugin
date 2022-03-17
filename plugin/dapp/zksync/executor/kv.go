package executor

import (
	"fmt"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func GetAccountIdPrimaryKey(accountId uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"accountId-", accountId))
}

func GetLocalChain33EthPrimaryKey(chain33Addr string, ethAddr string) []byte {
	return []byte(fmt.Sprintf("%s", chain33Addr+"-"+ethAddr))
}

func GetChain33EthPrimaryKey(chain33Addr string, ethAddr string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chain33Addr+"-"+ethAddr))
}

func GetTokenPrimaryKey(accountId uint64, tokenId uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d%s%022d", KeyPrefixStateDB+"token-", accountId, "-", tokenId))
}

func GetRootIndexPrimaryKey(rootIndex uint64) []byte {
	return []byte(fmt.Sprintf("%s%016d", KeyPrefixStateDB+"rootIndex-", rootIndex))
}

func GetAccountTreeKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"accountTree"))
}

func getHeightKey(height int64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"treeHeightRoot", height))
}

func getVerifyKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"verifyKey"))
}

func getVerifier() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+zt.ZkVerifierKey))
}

func getLastCommitProofKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"commitProof"))
}

func getHeightCommitProofKey(blockHeight uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"proofHeight", blockHeight))
}

func getValidatorsKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"validators"))
}

func getEthPriorityQueueKey(chainID uint32) []byte {
	return []byte(fmt.Sprintf("%s-%d", KeyPrefixStateDB+"priorityQueue", chainID))
}

func getProofIdCommitProofKey(proofId uint64) []byte {
	return []byte(fmt.Sprintf("%016d", proofId))
}

func getRootCommitProofKey(root string) []byte {
	return []byte(fmt.Sprintf("%s", root))
}

func getHistoryAccountTreeKey(proofId, accountId uint64) []byte {
	return []byte(fmt.Sprintf("%016d.%16d", proofId, accountId))
}
