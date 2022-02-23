##  离线部署跨链合约及各操作
[TOC]

### 前期地址及配置文件准备
#### ethereum 地址准备 (数据来源 WL and LBZ)
* ethereum 的 socket 通信地址, eg: ws://182.160.7.143:8546
* ethereum 的 http url 地址, eg: http://182.160.7.143:8545

#### 部署平行链 (数据来源 HZM)
得到 chain33 rcp url, eg: http://35.77.111.58:8901

#### 准备4台以上的服务器部署中继器 (数据来源 HZM)
其中一台部署代理中继器
剩下的部署普通中继器, 根据普通验证人个数配置, 3个以上, 一一对应

#### 编译代码
```shell
git clone git@github.com:33cn/plugin.git
make
scp plugin/build/ci/bridgevmxgo 目标服务器
```

#### ethereum 部署配置文件
把要部署需要的数据写入 deploy_ethereum.toml 配置文件
```toml
# 合约部署人
operatorAddr="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
# 验证人地址, 至少配置3个以上, 即大于等于3个
validatorsAddr=["0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]
# 验证人权重
initPowers=[25, 25, 25, 25]
# 主链 symbol 如果是BSC, 改为 BNB
symbol="ETH"
# 离线多签地址, 至少配置3个以上, 即大于等于3个
multisignAddrs=["0x4c85848a7E2985B76f06a7Ed338FCB3aF94a7DCf", "0x6F163E6daf0090D897AD7016484f10e0cE844994", "0xbc333839E37bc7fAAD0137aBaE2275030555101f", "0x495953A743ef169EC5D4aC7b5F786BF2Bd56aFd5"]
```

#### ethereum 端所需地址及说明 (数据来源管理员)
|地址|说明|
|----|----|
|operatorAddr|合约部署人, 需要比较多的金额, 用于部署合约时需要的手续费|
|validatorsAddr[]|普通验证人地址, 3个以上, 需要少量金额, 用于用户从chain33中提币时手续费, 需要监测, 地址金额不能为空否则提币失败, BNB 建议 0.1 个, 根据需求增加或减少|
|multisignAddrs[]|离线多签地址, 3个以上, 需要少量金额, 用于多签提币时手续费, BNB 建议 0.1 个, 根据需求增加或减少|
|validatorsAddrp|代理验证人地址, 代理打币地址, 需要较多金额, 需要监测, 每天结束后, 查看剩余金额, 金额不足继续打币, BNB 建议 0.1 个, 根据需求增加或减少|

#### chain33 部署配置文件
把要部署需要的数据写入 chain33_ethereum.toml 配置文件
```toml
# 验证人地址, 至少配置3个以上, 即大于等于3个
validatorsAddr=["1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ", "155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6", "13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv", "113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG"]
# 验证人权重
initPowers=[25, 25, 25, 25]
# 离线多签地址
multisignAddrs=["168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9", "13KTf57aCkVVJYNJBXBBveiA5V811SrLcT", "1JQwQWsShTHC4zxHzbUfYQK4kRBriUQdEe", "1NHuKqoKe3hyv52PF8XBAyaTmJWAqA2Jbb"]
```

#### chain33 端所需地址及说明 (数据来源管理员)
|地址|说明|
|----|----|
|operatorAddr|合约部署人, 需要比较多的金额, 用于部署合约时需要的手续费|
|validatorsAddr[]|普通验证人地址, 3个以上, 需要少量金额, 用于用户从chain33中提币时手续费, 需要监测, 地址金额不能为空否则提币失败, BTY 建议 20 个, 根据需求增加或减少|
|multisignAddrs[]|离线多签地址, 3个以上, 需要少量金额, 用于多签提币时手续费, BTY 建议 20 个, 根据需求增加或减少|
|validatorsAddrp|代理验证人地址, BTY 建议 20 个, 根据需求增加或减少|
|validatorsAddrsp|代理验证人地址, 代理收币地址|

###  离线部署 ethereum 跨链合约及各操作
#### 基础步骤
* 在线创建交易 `./boss4x ethereum offline create ...` 需要在线查询 nonce 等信息
* 离线签名交易 `./boss4x ethereum offline sign -f xxx.txt -k ...`
* 在线发送签名后文件 `./boss4x ethereum offline send -f deploysigntxs.txt` 默认签名后的文件名称都是 deploysigntxs.txt

拼凑 boss4x 命令

./boss4x --rpc_laddr_ethereum http://139.9.219.183:8545 --chainEthId 1
```
--chainEthId int              chainId, 如果是Bsc, chainId为56, 如果是ethereum, chainId为1, 查询链接:https://chainlist.org/
--rpc_laddr_ethereum string   ethereum http url (default "http://localhost:7545")
```

#### 离线部署 ethereum 跨链合约
* 在线创建交易
```
交易1: 部署合约: Valset
交易2: 部署合约: EthereumBridge
交易3: 部署合约: Oracle
交易4: 部署合约: BridgeBank
交易5: 在合约EthereumBridge中设置BridgeBank合约地址
交易6: 在合约EthereumBridge中设置Oracle合约地址
交易7: 设置 symbol
交易8: 部署合约: BridgeRegistry
交易9: 部署合约: MulSign
交易10: 设置 bridgebank 合约地址可以转到多签合约地址
交易11: 设置离线多签地址信息
```
文件部署命令:
```shell
./boss4x ethereum offline create_file -c deploy_ethereum.toml
```

* 离线签名交易
```
./boss4x ethereum offline sign -k ...

参数说明：
  -f, --file string   需要签名的文件, 默认:deploytxs.txt (default "deploytxs.txt")
  -k, --key string    部署者的私钥
```

* 发送签名后文件
```
./boss4x ethereum offline send -f deploysigntxs.txt
```

* 输出
```
交易4: 部署合约: BridgeBank
交易8: 部署合约: BridgeRegistry
交易9: 部署合约: MulSign
```
记录如上几个合约地址, 后续会用到

#### 设置 symbol 允许被 lock
根据需求设置, 如果是多个不同的 symbol 则设置多次
* 在线创建交易
```
命令：
./boss4x ethereum offline create_add_lock_list -s USDT -t "${ethereumUSDTERC20TokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"

参数说明：
  -c, --contract string     bridgebank 合约地址
  -d, --deployAddr string   部署者地址
  -s, --symbol string       token symbol
  -t, --token string        token addr

输出
tx is written to file:  create_add_lock_list.txt
```

#### 创建 bridge token
根据需求设置, 如果是多个不同的 symbol 则创建多次
* 在线创建交易
```
命令：
./boss4x ethereum offline create_bridge_token -s BTY -c "${ethBridgeBank}" -d "${ethDeployAddr}"

参数说明：
  -c, --contract string     bridgebank 合约地址
  -d, --deployAddr string   部署者地址
  -s, --symbol string       token symbol

输出
tx is written to file:  create_bridge_token.txt
```

#### 离线部署 ERC20 跨链合约
根据需求创建
* 在线创建交易
```
命令：
./boss4x ethereum offline create_erc20 -m 33000000000000000000 -s YCC -o "${ethTestAddr1}" -d "${ethDeployAddr}"

参数说明：
  -m, --amount string       金额
  -d, --deployAddr string   部署者地址
  -o, --owner string        拥有者地址
  -s, --symbol string       erc20 symbol

输出
tx is written to file:  deployErc20YCC.txt
把交易信息写入 deployErc20XXX.txt 文件中, 其中 XXX 为 erc20 symbol
```

#### 设置 bridgebank 金额到多少后自动转入多签合约地址
根据需求创建
* 在线创建交易
```
命令：
./boss4x ethereum offline set_offline_token -s ETH -m ${threshold} -p ${percents} -c "${ethBridgeBank}" -d "${ethDeployAddr}"

参数说明：
  -c, --contract string     bridgebank 合约地址
  -d, --deployAddr string   deploy 部署者地址
  -p, --percents uint8      百分比 (默认 50),达到阈值后默认转帐 50% 到离线多签的地址
  -s, --symbol string       token 标识
  -m, --threshold float     阈值
  -t, --token string        token 地址, 如果是 ETH(主链币), token 地址为空

输出
tx is written to file:  set_offline_token.txt
```

#### 离线多签转帐
根据需求需要是转帐
* 转帐预备交易--在线操作
```
命令：
./boss4x ethereum offline multisign_transfer_prepare -a 8 -r "${ethBridgeBank}" -c "${multisignEthAddr}" -d "${ethTestAddr1}" -t "${ethereumUSDTERC20TokenAddr}"

参数说明：
  -a, --amount float      转帐金额
  -c, --contract string   离线多签合约地址
  -r, --receiver string   接收者地址
  -d, --sendAddr string   发送这笔交易的地址, 需要扣除部分手续费
  -t, --token string      erc20 地址, 空的话, 默认转帐 ETH

输出
tx is written to file:  multisign_transfer_prepare.txt
```

* 离线多签地址签名交易--离线操作
```
命令：
./boss4x ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"

参数说明：
  -f, --file string   tx file, default: multisign_transfer_prepare.txt (default "multisign_transfer_prepare.txt")
  -k, --keys string   owners' private key, separated by ','

输出
tx is written to file:  sign_multisign_tx.txt
```

* 发送交易--在线操作,需要一个地址扣手续费
```
命令：
./boss4x ethereum offline send_multisign_tx -f sign_multisign_tx.txt -k "${ethTestAddrKey1}"
```

###  离线部署 chain33 跨链合约及各操作
#### 基础步骤
* 离线创建交易并签名 `./boss4x chain33 offline create ...`
* 在线发送签名后文件 `./boss4x chain33 offline send -f XXX.txt`

拼凑 boss4x 命令
```shell
./boss4x --rpc_laddr http://${chain33_ip}:8901 --rpc_laddr_ethereum --paraName user.p.para. --chainID 0

--chainID int32               chain id, default to 0
--expire string               transaction expire time (optional) (default "120m")
--paraName string             para chain name,Eg:user.p.fzm.
--rpc_laddr string            http url (default "https://localhost:8801")
```

#### 离线部署 chain33 跨链合约
* 离线创建交易
```
交易1: 部署合约: Valset
交易2: 部署合约: chain33Bridge
交易3: 部署合约: Oracle
交易4: 部署合约: BridgeBank
交易5: 在合约chain33Bridge中设置BridgeBank合约地址
交易6: 在合约chain33Bridge中设置Oracle合约地址
交易7: 部署合约: BridgeRegistry
交易8: 部署合约: MulSign
交易9: 设置 bridgebank 合约地址可以转到多签合约地址
交易10: 设置离线多签地址信息

命令:
./boss4x chain33 offline create_file -f 1 -k "${chain33DeployKey}" -n "deploy crossx to chain33" -c "./deploy_chain33.toml"

执行之后会将交易写入到文件：
  deployCrossX2Chain33.txt
```

* 发送签名后文件
```
./boss4x chain33 offline send -f deployCrossX2Chain33.txt
```


* 输出
```
交易4: 部署合约: BridgeBank
交易7: 部署合约: BridgeRegistry
交易8: 部署合约: MulSign
```
记录如上几个合约地址, 后续会用到

#### 离线部署 ERC20 跨链合约
##### 部署
根据需求部署
* 离线创建交易
```
命令：
./boss4x chain33 offline create_erc20 -s YCC -k ... -o 1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ

参数说明：
  -a, --amount float    铸币金额,默认 3300*1e8
  -k, --key string      部署人的私钥, 用于对交易签名
  -o, --owner string    拥有者地址
  -s, --symbol string   token 标识

执行之后会将交易写入到文件：
  deployErc20XXXChain33.txt 其中 XXX 为 token 标识
```

##### approve_erc20
根据需求设置
* 离线创建交易
```
命令：
./boss4x chain33 offline approve_erc20 -a 330000000000 -s 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -c 1998HqVnt4JUirhC9KL5V71xYU8cFRn82c -k ... 

参数说明：  
  -a, --amount float      审批金额
  -s, --approve string    审批地址, chain33 BridgeBank 合约地址
  -c, --contract string   Erc20 合约地址
  -k, --key string        部署人的私钥, 用于对交易签名

执行之后会将交易写入到文件：
  approve_erc20.txt
```

##### create_add_lock_list
根据需求设置
* 离线创建交易
```
命令：
./boss4x chain33 offline create_add_lock_list -c 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -k ... -t 1998HqVnt4JUirhC9KL5V71xYU8cFRn82c -s YCC

参数说明：
  -c, --contract string   bridgebank 合约地址
  -k, --key string        部署人的私钥, 用于对交易签名
  -s, --symbol string     token 标识
  -t, --token string      Erc20 合约地址


执行之后会将交易写入到文件：
  create_add_lock_list.txt
```

#### create_bridge_token
根据需求创建
* 离线创建交易
```
命令：
./boss4x chain33 offline create_bridge_token -c 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -k ... -s YCC

参数说明：
  -c, --contract string   bridgebank 合约地址
  -k, --key string        部署人的私钥, 用于对交易签名
  -s, --symbol string     token 标识

执行之后会将交易写入到文件：
  create_bridge_token.txt
```
* 获取 bridge_token 地址
```
命令：
./chain33-cli evm query -a 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -c 1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ -b 'getToken2address(YCC)'

参数说明：
  -a, --address string   evm 合约地址,这里是 chain33 BridgeBank 合约地址
  -c, --caller string    the caller address,这里是部署者地址
  -b, --input string     call params (abi format) like foobar(param1,param2),发送的函数
  -t, --path string      abi path(optional), default to .(current directory) (default "./"),abi文件地址,默认本地"./"

输出:
  15XsGjTbV6SxQtDE1SC5oaHx8HbseQ4Lf9 -- bridge_token 地址
```

#### 离线多签设置
根据需求设置
* 离线创建交易
```
命令：
./boss4x chain33 offline set_offline_token -c 1MaP3rrwiLV1wrxPhDwAfHggtei1ByaKrP -s BTY -m 100000000000 -p 50 -k ...

参数说明：
  -c, --contract string    bridgebank 合约地址
  -f, --fee float          交易费
  -k, --key string         部署者私钥
  -n, --note string        备注
  -p, --percents uint8     百分比 (默认 50),达到阈值后默认转帐 50% 到离线多签的地址
  -s, --symbol string      token 标识
  -m, --threshold string   阈值
  -t, --token string       token 地址


执行之后会将交易写入到文件：
  chain33_set_offline_token.txt
```

#### 离线多签转帐
根据需求操作
* 创建转帐交易--在线操作,需要重新获取 nonce 等信息
```
命令：
./boss4x chain33 offline create_multisign_transfer -a 10 -r 168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9 -m 1NFDfEwne4kjuxAZrtYEh4kfSrnGSE7ap

参数说明：  
  -m, --address string    离线多签合约地址
  -a, --amount float      转帐金额
  -r, --receiver string   接收者地址
  -t, --token string      erc20 地址,空的话, 默认转帐 BTY

执行之后会将交易写入到文件：
  create_multisign_transfer.txt
```

* 离线多签地址签名交易--离线操作
```
命令：
./boss4x chain33 offline multisign_transfer -k ... -s 0xcd284cd17456b73619fa609bb9e3105e8eff5d059c5e0b6eb1effbebd4d64144,0xe892212221b3b58211b90194365f4662764b6d5474ef2961ef77c909e31eeed3,0x9d19a2e9a440187010634f4f08ce36e2bc7b521581436a99f05568be94dc66ea,0x45d4ce009e25e6d5e00d8d3a50565944b2e3604aa473680a656b242d9acbff35

参数说明：
  -f, --fee float     手续费
  -t, --file string   签名交易文件, 默认: create_multisign_transfer.txt
  -k, --key string    部署者私钥
  -s, --keys string   离线多签的多个私钥, 用','分隔
  -n, --note string   备注


执行之后会将交易写入到文件：
  multisign_transfer.txt
```

###  离线部署 xgo 合约及各操作
#### 基础步骤
* 离线创建并签名交易 `./evmxgoboss4x chain33 offline create ... -k ...`
* 在线发送签名后文件 `./evmxgoboss4x chain33 offline send -f ...`

拼凑 evmxgoboss4x 命令 (scp plugin/build/ci/bridgevmxgo/evmxgoboss4x 目标服务器)

./evmxgoboss4x --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para. --chainID 0
```
--chainID int32      chain id, default to 0
--expire string      transaction expire time (optional) (default "120m")
--paraName string    平行链名称
--rpc_laddr string   平行链 url (default "https://localhost:8801")
```

#### 离线部署 chain33 跨链合约
* 离线创建并签名交易
```
交易1: 部署合约 Valset
交易2: 部署合约 EthereumBridge
交易3: 部署合约 Oracle
交易4: 部署合约 BridgeBank
交易5: 设置合约 set BridgeBank to EthBridge 
交易6: 设置合约 set Oracle to EthBridge 
交易7: 部署合约 BridgeRegistry 

命令：
./evmxgoboss4x chain33 offline create -f 1 -k "${chain33DeployKey}" -n "deploy crossx to chain33" -r "${chain33DeployAddr}, [${chain33Validatora}, ${chain33Validatorb}, ${chain33Validatorc}, ${chain33Validatord}], [96, 1, 1, 1]"
    
参数说明：
  -f, --fee float         手续费
  -k, --key string        部署合约的私钥
  -n, --note string       交易备注
  -r, --valset string     valset 合约参数, 格式: 'addr, [addr, addr, addr, addr], [25, 25, 25, 25]','部署地址,[验证者A地址, ...],[验证者A权重, ...]'
输出:
把交易信息写入 deployBridgevmxgo2Chain33.txt 文件中
```

* 发送签名后文件
```
./evmxgoboss4x chain33 offline send -f deployBridgevmxgo2Chain33.txt
```

#### 设置 symbol 允许被 lock
根据需求设置
* 在线创建交易
```
命令：
./evmxgoboss4x chain33 offline create_add_lock_list -s ETH -t "${chain33EthBridgeTokenAddr}" -c "${XgoChain33BridgeBank}" -k "${chain33DeployKey}"

参数说明：
  -c, --contract string   创建的 xgo bridgebank 合约地址
  -f, --fee float         手续费
  -k, --key string        部署合约的私钥
  -n, --note string       交易备注
  -s, --symbol string     token symbol
  -t, --token string      chain33 evm bridge token 地址  

输出
tx is written to file:  create_add_lock_list.txt
```

#### 平行链管理者设置 bridgevmxgo 信息
根据需求设置
##### manage 设置 bridgevmxgo 合约地址
```shell
# 创建交易
# XgoChain33BridgeBank 部署的 xgo BridgeBank 合约地址
curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key":"bridgevmxgo-contract-addr","value":"{\"address\":\"'"${XgoChain33BridgeBank}"'\"}","op":"add","addr":""}}]}' -H 'content-type:text/plain;' "http://${docker_chain33_ip}:8901"

# 用平行链管理者地址签名
./chain33_cli wallet sign -k "$paraMainAddrKey" -d "${tx}"
```

##### manage add symbol
```shell
# 创建交易
# symbol 需要增加的 symbol
# bridgeTokenAddr chain33 对应的 BridgeToken 地址, 例如:chain33EthBridgeTokenAddr
curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key":"evmxgo-mint-'"${symbol}"'","value":"{\"address\":\"'"${bridgeTokenAddr}"'\",\"precision\":8,\"introduction\":\"symbol:'"${symbol}"', bridgeTokenAddr:'"${bridgeTokenAddr}"'\"}","op":"add","addr":""}}]}' -H 'content-type:text/plain;' "http://${docker_chain33_ip}:8901"

# 用平行链管理者地址签名
./chain33_cli wallet sign -k "$paraMainAddrKey" -d "${tx}"
```
