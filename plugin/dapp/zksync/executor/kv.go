package executor

import (
	"fmt"
	"github.com/33cn/chain33/common/address"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func GetAccountIdPrimaryKey(accountId uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"accountId-", accountId))
}

func GetLocalChain33EthPrimaryKey(chain33Addr string, ethAddr string) []byte {
	return []byte(fmt.Sprintf("%s-%s", address.FormatAddrKey(chain33Addr), address.FormatAddrKey(ethAddr)))
}

func GetChain33EthPrimaryKey(chain33Addr string, ethAddr string) []byte {
	return []byte(fmt.Sprintf("%s%s-%s", KeyPrefixStateDB, address.FormatAddrKey(chain33Addr),
		address.FormatAddrKey(ethAddr)))
}

func GetTokenPrimaryKey(accountId uint64, tokenId uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d%s%022d", KeyPrefixStateDB+"token-", accountId, "-", tokenId))
}

func GetNFTIdPrimaryKey(nftTokenId uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"nftTokenId-", nftTokenId))
}

func GetNFTHashPrimaryKey(nftHash string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"nftHash-"+nftHash))
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

func getVerifyKey(chainTitleId string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitleId+"-verifyKey"))
}

func getVerifier(chainTitleId string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitleId+"-"+zt.ZkVerifierKey))
}

func getProofIdKey(chainTitleId string, id uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+chainTitleId+"-ProofId", id))
}

func getLastProofIdKey(chainTitleId string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitleId+"-lastProofId"))
}

func getMaxRecordProofIdKey(chainTitleId string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitleId+"-maxRecordProofId"))
}

func getLastOnChainProofIdKey(chainTitleId string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitleId+"-lastOnChainProofId"))
}

func getEthPriorityQueueKey(chainID uint32) []byte {
	return []byte(fmt.Sprintf("%s-%d", KeyPrefixStateDB+"priorityQueue", chainID))
}

//特意把title放后面，方便按id=1搜索所有的chain
func getProofIdCommitProofKey(chainTitleId string, proofId uint64) []byte {
	return []byte(fmt.Sprintf("%016d-%s", proofId, chainTitleId))
}

func getRootCommitProofKey(chainTitleId, root string) []byte {
	return []byte(fmt.Sprintf("%s-%s", chainTitleId, root))
}

func getHistoryAccountTreeKey(proofId, accountId uint64) []byte {
	return []byte(fmt.Sprintf("%016d.%16d", proofId, accountId))
}

func getZkFeeKey(actionTy int32, tokenId uint64) []byte {
	return []byte(fmt.Sprintf("%016d.%16d", actionTy, tokenId))
}
