zksync 实现了基于区块链和零知识证明的二层网络

基本功能如下：
1. Deposit, 接收来自L1的存款请求。
2. Withdraw, 从L2提款到L1

NFT
1. mint, 支持ERC721和ERC1155

分为4个步骤：
1. creator 在指定的DEFAULT_NFT_TOKEN_ID(id=256) 记录mint的个数或次数，对ERC1155也是按一次计数
2. 系统NFT_ACCOUNT_ID(id=2)分配新的NFT id，通过递增其DEFAULT_NFT_TOKEN_ID(id=256)数量
3. 系统NFT_ACCOUNT_ID以设置新NFT id的balance方式记录其指纹信息,hash(creatorId,serialId,protocol,amount,contentHash)
4. receiver 设置新NFT id， amount为铸造数量，对ERC721是1，ERC1155是批量值


