package executor

import "fmt"

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

func getVerifyKey() []byte {
	return []byte(fmt.Sprintf("%s", KeyPrefixStateDB+"verifyKey"))
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
