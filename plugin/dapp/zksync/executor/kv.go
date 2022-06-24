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

func getVerifyKey(chainTitle string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitle+"-verifyKey"))
}

func getVerifier(chainTitle string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitle+"-"+zt.ZkVerifierKey))
}

func getProofIdKey(chainTitle string, id uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+chainTitle+"-ProofId", id))
}

func getLastProofIdKey(chainTitle string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitle+"-lastProofId"))
}

func getMaxRecordProofIdKey(chainTitle string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitle+"-maxRecordProofId"))
}

func getLastOnChainProofIdKey(chainTitle string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+chainTitle+"-lastOnChainProofId"))
}

func getEthPriorityQueueKey(chainID uint32) []byte {
	return []byte(fmt.Sprintf("%s-%d", KeyPrefixStateDB+"priorityQueue", chainID))
}

//特意把title放后面，方便按id=1搜索所有的chain
func getProofIdCommitProofKey(chainTitle string, proofId uint64) []byte {
	return []byte(fmt.Sprintf("%016d-%s", proofId, chainTitle))
}

func getRootCommitProofKey(chainTitle, root string) []byte {
	return []byte(fmt.Sprintf("%s-%s", chainTitle, root))
}

func getHistoryAccountTreeKey(proofId, accountId uint64) []byte {
	return []byte(fmt.Sprintf("%016d.%16d", proofId, accountId))
}

func getZkFeeKey(actionTy int32, tokenId uint64) []byte {
	return []byte(fmt.Sprintf("%016d.%16d", actionTy, tokenId))
}
