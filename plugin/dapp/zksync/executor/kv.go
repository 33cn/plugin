package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/address"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func GetAccountIdPrimaryKeyPrefix() string {
	return fmt.Sprintf("%s", KeyPrefixStateDB+"accountId-")
}

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

func GetTokenPrimaryKeyPrefix() string {
	return fmt.Sprintf("%s", KeyPrefixStateDB+"token-")
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

func getVerifyKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"-verifyKey"))
}

func getVerifier() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"-"+zt.ZkVerifierKey))
}

func getLastOnChainProofIdKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"-lastOnChainProofId"))
}

//last eth priority id key
func getEthPriorityQueueKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"priorityQueue"))
}

//特意把title放后面，方便按id=1搜索所有的chain
func getProofIdCommitProofKey(proofId uint64) []byte {
	return []byte(fmt.Sprintf("%016d", proofId))
}

func getRootCommitProofKey(root string) []byte {
	return []byte(fmt.Sprintf("%s", root))
}

func getHistoryAccountTreeKey(proofId, accountId uint64) []byte {
	return []byte(fmt.Sprintf("%016d.%16d", proofId, accountId))
}

func getZkFeeKey(actionTy int32, tokenId uint64) []byte {
	return []byte(fmt.Sprintf("%s%02d-%03d", KeyPrefixStateDB+"fee-", actionTy, tokenId))
}

func CalcLatestAccountIDKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"latestAccountID"))
}

func getExodusModeKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"exodusMode"))
}

//GetTokenSymbolKey tokenId 对应symbol
func GetTokenSymbolKey(tokenId string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"tokenId-"+tokenId))
}

//GetTokenSymbolIdKey token symbol 对应id
func GetTokenSymbolIdKey(symbol string) []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"tokenSym-"+symbol))
}

func getLastProofIdKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"-lastProof"))
}

func getMaxRecordProofIdKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"-maxRecordProofId"))
}

func getProofIdKey(id uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"-ProofId", id))
}

//the first L2 op that not be verified by proof
func getL2FirstQueueIdKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"-L2FirstQueId"))
}

//the last op that queued to L2, 从0开始
func getL2LastQueueIdKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"-L2LastQueId"))
}

//the specific L2 op queue id data key
func getL2QueueIdKey(id int64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"-L2QueueId", id))
}

//the proof id to the end first queue id key, the end first queue id == last pubdata's operation
func getProofId2QueueIdKey(proofID uint64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"-proof2queueId", proofID))
}

//the proof id to the end first queue id key, the end first queue id == last pubdata's operation
func getL1PriorityId2QueueIdKey(priorityId int64) []byte {
	return []byte(fmt.Sprintf("%s%022d", KeyPrefixStateDB+"-priority2QueId", priorityId))
}
