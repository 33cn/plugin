zksync合约实现了基于区块链和零知识证明的二层网络


基本功能如下：
1. Deposit, 接收来自L1的存款请求，需指定L2接收者地址，由relayer转发L1的存款交易到L2
2. Withdraw, 从L2提款到L1
3. Transfer, L2内部已知账户id间的转账
4. Transfer2New L2内部向未从L1存款过的账户(没有账户ID)转账,并创建新账户ID
5. ProxyExit L2发起代理其他用户向L1提款，代理者支付交易费
6. SetPubKey 用户向指定L2地址存款后，设置这个地址的公钥，才能使用跟此地址相关的资产
7. FullExit 从L1发起的全额退款，目前L2未支持
8. Swap L2内部不同资产之间的撮合交换
9. Contract2Tree 把从L2的tree转入L3合约的资产转回L2
10. Tree2Contract 把L2的资产转入到L3合约账户
11. Fee 按交易类型设置交易费
12. MintNFT 在L2铸造NFT
13. WithdrawNFT 把在L2铸造的NFT提取到L1
14. TransferNFT L2铸造的NFT转给其他账户

NFT
1. mint, 支持ERC721和ERC1155

分为4个步骤：
1. creator 在指定的DEFAULT_NFT_TOKEN_ID(id=256) 记录mint的个数或次数，对ERC1155也是按一次计数
2. 系统NFT_ACCOUNT_ID(id=2)分配新的NFT id，通过递增其DEFAULT_NFT_TOKEN_ID(id=256)数量
3. 系统NFT_ACCOUNT_ID以设置新NFT id的balance方式记录其指纹信息,hash(creatorId,serialId,protocol,amount,contentHash)
4. receiver 设置新NFT id， amount为铸造数量，对ERC721是1，ERC1155是批量值

用户资产退出机制
1. L1上的存款超过一定期限比如30天未被来自L2的证明确认，则认为L2失效了，任何用户可以激活L1层的exodus模式，准备退出资产。
2. 在L1激活exodus后，未被证明确认的存款，可以通过回滚的方式退回到原来的存款账户(rollbackDepositsForExodusMode)
3. exodus模式激活后，L2的资产由于有一部分在L3合约中,需要先使用contract2tree转到L2，待所有L3合约中资产转到L2后，
   管理员根据最后的tree root 更新到L1合约，再从L2提交exodus证明把资产从L1依次提走,如果管理员超过预设的threshold未设置root,
   在超过2*threshold天数(块数)后,L1合约允许以L1的最后一个proof的treeRoot作为逃生舱的root来提款
4. L2资产退出机制
   1）L2设置exodus mode或exodus的pause mode停止除contract2tree外的所有操作，以准备尽快确定最后的tree root
   2）通知用户把L3的资产都提回到L2，超过一定期限则可以允许管理员把资产都提回到L2(system accountId=3的所有token资产为0为全部提到了L2)
   3）管理员根据L1最后一个success proofId设置exodus final mode,尝试回滚success proofId后的deposit和withdraw
       3.1) 如果此时accountId=3 某token余额非0，则失败
       3.2) 如果回滚deposit,withdraw失败则返回差额的提示，当某account的deposit因为已经transfer而余额不足扣除失败时候会尝试由feeId垫付，
            如果feeId也不够，则返回失败，提示差额，差额部分需要管理员在L1存入以弥补。如果失败后，重新发送此交易并设置knowGap=1则忽视失败而完成回滚
   4）在系统完成final mode后，管理员构建L2的最后的tree root(cli: zksync l2 build_tree)设置到L1后根据此最终root进行逃生舱提款


交易费设置
1. 除了deposit外，其他操作都需要设置交易费，因为操作都需要零知识证明，防止用户大量使用非常小的交易值而不断触发证明来作恶
2. 由于不同erc20 token有各自的精度，除contract2tree外都是按token精度来处理交易amount和fee的。
   比如eth精度是18，那交易的amount和fee都是要以精度18来，1eth=1e18, usdt精度为6， 1U=1000000来设置
   在tree2contract中，存储到L3合约的balance时候，会按系统精度8来转化，比如1U=1e6，在合约中会扩大为1e8，也是代表1token
   而1eth，1e18的amount存储到L3合约时候会去掉1e10的精度，变为1e8的amount存储，仍代表1eth
3. contract2tree的交易的amount和fee是统一按L3系统精度来设置的，因为用户是把在L3的amount提回到二层
   比如提回1eth，amount按系统精度为8设置为100000000，在contract2tree action的实现中会根据eth实际精度18存储为1e18，也就是扩大1e10
   比如提回2USDT,amount按系统精度为8设置为200000000， 在contract2tree的实现中会把值按精度6缩减为2000000存储