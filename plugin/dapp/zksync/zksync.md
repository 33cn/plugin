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
9. Contract2Tree 把从L2的tree转入chain33合约的资产转回L2
10. Tree2Contract 把L2的资产转入到chain33合约账户
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
3. exodus模式激活后，L2的资产由于有一部分在chain33合约中,需要先使用contract2tree转到L2，待所有合约中资产转到L2后，根据最后的tree root 更新到L1合约，
   再从L2提交exodus证明把资产从L1依次提走
4. L2资产退出机制
   1）首先确定最后一个L1的proofId，在L2上通过此ID找到此证明后的第一个失效交易和失效证明(失效交易是L1和L2的跨链交易比如deposit,
   withdraw,proxyexit 这些交易状态需要在L2回滚)，可通过cmd: zksync query proof firstop 获得指定proofId后的第一个失效tx和失效proof root
   2）L2如果是平行链，平行链配置文件设置失效交易和失效证明后，从0开始重新同步(相当于回滚)，如果平行链有自共识，设置关闭自共识，待重新同步后
   管理员在通告上的指定时间后设置"退出清算"模式，在此模式下，只有contract2tree交易可以执行，待chain33合约中的L2资产都转回到L2上，管理员在L1设置清算root
   3）用户根据清算root计算本账户相应token资产的退出证明，自动提交到L1退出资产。或者交易所可以通过批量退出的机制帮助用户退出资产。
   4）L2如果是联盟链也可以通过类似失效交易和失效证明回滚的方式完成状态更新，或者通过提交到公链的证明的pubdata重新计算

交易费设置
1. 除了deposit外，其他操作都需要设置交易费，因为操作都需要零知识证明，防止用户大量使用非常小的交易值而不断触发证明来作恶
2. 由于不同erc20 token有各自的精度，除contract2tree外都是按token精度来处理交易amount和fee的。
   比如eth精度是18，那交易的amount和fee都是要以精度18来，1eth=1e18, usdt精度为6， 1U=1000000来设置
   在tree2contract中，存储到chain33合约的balance时候，会按系统精度8来转化，比如1U=1e6，在合约中会扩大为1e8，也是代表1token
   而1eth，1e18的amount存储到chain33合约时候会去掉1e10的精度，变为1e8的amount存储，仍代表1eth
3. contract2tree的交易的amount和fee是统一按chain33系统精度来设置的，因为用户是把在chain33的amount提回到二层
   比如提回1eth，amount按系统精度为8设置为100000000，在contract2tree action的实现中会根据eth实际精度18存储为1e18，也就是扩大1e10
   比如提回2USDT,amount按系统精度为8设置为200000000， 在contract2tree的实现中会把值按精度6缩减为2000000存储