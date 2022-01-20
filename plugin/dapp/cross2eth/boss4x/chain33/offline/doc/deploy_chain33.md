###  离线部署 chain33 跨链合约及各操作
*** 

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
*** 

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

命令：
./boss4x chain33 offline create -f 1 -k "${chain33DeployKey}" -n "deploy crossx to chain33" -r "${chain33DeployAddr}, [${chain33Validatora}, ${chain33Validatorb}, ${chain33Validatorc}, ${chain33Validatord}], [25, 25, 25, 25]" -m "${chain33MultisignA},${chain33MultisignB},${chain33MultisignC},${chain33MultisignD}"

参数说明：
  -f, --fee float               交易费设置，因为只是少量几笔交易，且部署交易消耗gas较多，直接设置1个代币即可
  -k, --key string              部署人的私钥，用于对交易签名
  -m, --multisignAddrs string   离线多签地址, as: 'addr,addr,addr,addr'
  -n, --note string             备注信息 
  -r, --valset string           构造函数参数,严格按照该格式输入'addr, [addr, addr, addr, addr], [25, 25, 25, 25]',其中第一个地址为部署人私钥对应地址，后面4个地址为不同验证人的地址，4个数字为不同验证人的权重

  --rpc_laddr string    chain33 url 地址 (默认 "https://localhost:8801")
  --chainID int32       平行链的chainID, 默认: 0(代表主链)

执行之后会将交易写入到文件：
  deployCrossX2Chain33.txt
```

* 发送签名后文件
```
./boss4x chain33 offline send -f deployCrossX2Chain33.txt
```
***

#### 文件部署
把要部署需要的数据写入 chain33_ethereum.toml 配置文件
```toml
# 验证人地址，至少配置３个以上，即大于等于３个
validatorsAddr=["1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ", "155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6", "13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv", "113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG"]
# 验证人权重
initPowers=[25, 25, 25, 25]
# 离线多签地址
multisignAddrs=["168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9", "13KTf57aCkVVJYNJBXBBveiA5V811SrLcT", "1JQwQWsShTHC4zxHzbUfYQK4kRBriUQdEe", "1NHuKqoKe3hyv52PF8XBAyaTmJWAqA2Jbb"]
```
命令:
```shell
./boss4x chain33 offline create_file -f 1 -k "${chain33DeployKey}" -n "deploy crossx to chain33" -c "./deploy_chain33.toml"
```

#### 离线部署 ERC20 跨链合约
* 离线创建交易
```
命令：
./boss4x chain33 offline create_erc20 -s YCC -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -o 1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ --chainID 33

参数说明：
  -a, --amount float    铸币金额,默认 3300*1e8
  -k, --key string      部署人的私钥，用于对交易签名
  -o, --owner string    拥有者地址
  -s, --symbol string   token 标识

执行之后会将交易写入到文件：
  deployErc20XXXChain33.txt 其中 XXX 为 token 标识
```

#### approve_erc20
* 离线创建交易
```
命令：
./boss4x chain33 offline approve_erc20 -a 330000000000 -s 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -c 1998HqVnt4JUirhC9KL5V71xYU8cFRn82c -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae  --chainID 33

参数说明：  
  -a, --amount float      审批金额
  -s, --approve string    审批地址, chain33 BridgeBank 合约地址
  -c, --contract string   Erc20 合约地址
  -k, --key string        部署人的私钥，用于对交易签名

执行之后会将交易写入到文件：
  approve_erc20.txt
```

#### create_add_lock_list
* 离线创建交易
```
命令：
./boss4x chain33 offline create_add_lock_list -c 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -t 1998HqVnt4JUirhC9KL5V71xYU8cFRn82c --chainID 33 -s YCC

参数说明：
  -c, --contract string   bridgebank 合约地址
  -k, --key string        部署人的私钥，用于对交易签名
  -s, --symbol string     token 标识
  -t, --token string      Erc20 合约地址


执行之后会将交易写入到文件：
  create_add_lock_list.txt
```

#### create_bridge_token
* 离线创建交易
```
命令：
./boss4x chain33 offline create_bridge_token -c 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -s YCC --chainID 33

参数说明：
  -c, --contract string   bridgebank 合约地址
  -k, --key string        部署人的私钥，用于对交易签名
  -s, --symbol string     token 标识

执行之后会将交易写入到文件：
  create_bridge_token.txt
```
* 获取 bridge_token 地址
```
命令：
./chain33-cli evm abi call -a 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -c 1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ -b 'getToken2address(YCC)'

参数说明：
  -a, --address string   evm 合约地址,这里是 chain33 BridgeBank 合约地址
  -c, --caller string    the caller address,这里是部署者地址
  -b, --input string     call params (abi format) like foobar(param1,param2),发送的函数
  -t, --path string      abi path(optional), default to .(current directory) (default "./"),abi文件地址,默认本地"./"

输出:
  15XsGjTbV6SxQtDE1SC5oaHx8HbseQ4Lf9 -- bridge_token 地址
```

***
#### 离线多签设置
* 离线创建交易
```
命令：
./boss4x chain33 offline set_offline_token -c 1MaP3rrwiLV1wrxPhDwAfHggtei1ByaKrP -s BTY -m 100000000000 -p 50 -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae --chainID 33

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

***
#### 离线多签转帐
* 创建转帐交易--在线操作,需要重新获取 nonce 等信息
```
命令：
./boss4x chain33 offline create_multisign_transfer -a 10 -r 168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9 -m 1NFDfEwne4kjuxAZrtYEh4kfSrnGSE7ap

参数说明：  
  -m, --address string    离线多签合约地址
  -a, --amount float      转帐金额
  -r, --receiver string   接收者地址
  -t, --token string      erc20 地址,空的话，默认转帐 BTY


执行之后会将交易写入到文件：
  create_multisign_transfer.txt
```

* 离线多签地址签名交易--离线操作
```
命令：
./boss4x chain33 offline multisign_transfer -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262a -s 0xcd284cd17456b73619fa609bb9e3105e8eff5d059c5e0b6eb1effbebd4d64144,0xe892212221b3b58211b90194365f4662764b6d5474ef2961ef77c909e31eeed3,0x9d19a2e9a440187010634f4f08ce36e2bc7b521581436a99f05568be94dc66ea,0x45d4ce009e25e6d5e00d8d3a50565944b2e3604aa473680a656b242d9acbff35 --chainID 33

参数说明：
  -f, --fee float     手续费
  -t, --file string   签名交易文件, 默认: create_multisign_transfer.txt
  -k, --key string    部署者私钥
  -s, --keys string   离线多签的多个私钥, 用','分隔
  -n, --note string   备注


执行之后会将交易写入到文件：
  multisign_transfer.txt
```
