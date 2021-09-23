###  离线部署 ethereum 跨链合约及各操作
*** 

#### 基础步骤
* 在线创建交易 `./boss4x ethereum offline create ...` 需要在线查询 nonce 等信息
* 离线签名交易 `./boss4x ethereum offline sign -f xxx.txt -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230`
* 在线发送签名后文件 `./boss4x ethereum offline send -f deploysigntxs.txt` 默认签名后的文件名称都是 deploysigntxs.txt
*** 

#### 离线部署 ethereum 跨链合约
* 在线创建交易
```
交易1: 部署合约: Valset
交易2: 部署合约: EthereumBridge
交易3: 部署合约: Oracle
交易4: 部署合约: BridgeBank
交易5: 在合约EthereumBridge中设置BridgeBank合约地址
交易6: 在合约EthereumBridge中设置Oracle合约地址
交易7: 部署合约: BridgeRegistry
交易7: 部署合约: MulSign

命令：
./boss4x ethereum offline create -p 25,25,25,25 -o 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a -v 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a,0x0df9a824699bc5878232c9e612fe1a5346a5a368,0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1,0xd9dab021e74ecf475788ed7b61356056b2095830

参数说明：
  -p, --initPowers string        验证者权重, as: '25,25,25,25'
  -o, --owner string             部署者地址
  -v, --validatorsAddrs string   验证者地址, as: 'addr,addr,addr,addr'

  --rpc_laddr_ethereum string    ethereum url 地址 (默认 "http://localhost:7545")

输出:
tx is written to file:  deploytxs.txt

把交易信息写入文件中
```

* 离线签名交易
```
./boss4x ethereum offline sign -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230

参数说明：
  -f, --file string   需要签名的文件, 默认:deploytxs.txt (default "deploytxs.txt")
  -k, --key string    部署者的私钥
```

* 发送签名后文件
```
./boss4x ethereum offline send -f deploysigntxs.txt
```
***
#### 离线部署 ERC20 跨链合约
* 在线创建交易
```
命令：
./boss4x ethereum offline create_erc20 -m 33000000000000000000 -s YCC -o 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a

参数说明：
  -m, --amount string       金额
  -d, --deployAddr string   部署者地址
  -o, --owner string        拥有者地址
  -s, --symbol string       erc20 symbol

输出
tx is written to file:  deployErc20YCC.txt
把交易信息写入 deployErc20XXX.txt 文件中, 其中 XXX 为 erc20 symbol
```

* 离线签名交易
```
./boss4x ethereum offline sign -f deployErc20YCC.txt -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
```

* 发送签名后文件
```
./boss4x ethereum offline send -f deploysigntxs.txt
```

***
#### create_add_lock_list
* 在线创建交易
```
命令：
./boss4x ethereum offline create_add_lock_list -s YCC -t 0x20a32A5680EBf55740B0C98B54cDE8e6FD5a4FB0 -c 0xC65B02a22B714b55D708518E2426a22ffB79113d -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a

参数说明：
  -c, --contract string     bridgebank 合约地址
  -d, --deployAddr string   部署者地址
  -s, --symbol string       token symbol
  -t, --token string        token addr

输出
tx is written to file:  create_add_lock_list.txt
```

***
#### 创建 bridge token
* 在线创建交易
```
命令：
./boss4x ethereum offline create_bridge_token -s BTY -c 0xC65B02a22B714b55D708518E2426a22ffB79113d -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a

参数说明：
  -c, --contract string     bridgebank 合约地址
  -d, --deployAddr string   部署者地址
  -s, --symbol string       token symbol

输出
tx is written to file:  create_bridge_token.txt
```

***
#### 设置离线多签地址信息
* 在线创建交易
```
命令：
./boss4x ethereum offline multisign_setup -m 0xbf271b2B23DA4fA8Dc93Ce86D27dd09796a7Bf54 -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a -o 0x4c85848a7E2985B76f06a7Ed338FCB3aF94a7DCf,0x6F163E6daf0090D897AD7016484f10e0cE844994,0xbc333839E37bc7fAAD0137aBaE2275030555101f,0x495953A743ef169EC5D4aC7b5F786BF2Bd56aFd5

参数说明：
  -d, --deployAddr string   部署者地址
  -m, --multisign string    离线多签合约地址
  -o, --owner string        多签的地址, 用','分隔

输出
tx is written to file:  multisign_setup.txt
```

***
#### 设置离线多签地址
* 在线创建交易
```
命令：
./boss4x ethereum offline set_offline_addr -a 0xbf271b2B23DA4fA8Dc93Ce86D27dd09796a7Bf54 -c 0xC65B02a22B714b55D708518E2426a22ffB79113d -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a

参数说明：
  -a, --address string      离线多签地址
  -c, --contract string     bridgebank 合约地址
  -d, --deployAddr string   部署合约地址

输出
tx is written to file:  set_offline_addr.txt
```

***
#### 离线多签设置
* 在线创建交易
```
命令：
./boss4x ethereum offline set_offline_token -s ETH -m 20 -c 0xC65B02a22B714b55D708518E2426a22ffB79113d -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a

参数说明：
  -c, --contract string     bridgebank 合约地址
  -d, --deployAddr string   deploy 部署者地址
  -p, --percents uint8      百分比 (默认 50),达到阈值后默认转帐 50% 到离线多签的地址
  -s, --symbol string       token 标识
  -m, --threshold float     阈值
  -t, --token string        token 地址

输出
tx is written to file:  set_offline_token.txt
```

***
#### 离线多签转帐
* 转帐预备交易--在线操作
```
命令：
./boss4x ethereum offline multisign_transfer_prepare -a 3 -r 0xC65B02a22B714b55D708518E2426a22ffB79113d -c 0xbf271b2B23DA4fA8Dc93Ce86D27dd09796a7Bf54 -d 0x0df9a824699bc5878232c9e612fe1a5346a5a368

参数说明：
  -a, --amount float      转帐金额
  -c, --contract string   离线多签合约地址
  -r, --receiver string   接收者地址
  -d, --sendAddr string   发送这笔交易的地址, 需要扣除部分手续费
  -t, --token string      erc20 地址,空的话，默认转帐 ETH

输出
tx is written to file:  multisign_transfer_prepare.txt
```

* 离线多签地址签名交易--离线操作
```
命令：
./boss4x ethereum offline sign_multisign_tx -k 0x5e8aadb91eaa0fce4df0bcc8bd1af9e703a1d6db78e7a4ebffd6cf045e053574,0x0504bcb22b21874b85b15f1bfae19ad62fc2ad89caefc5344dc669c57efa60db,0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2,0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697

参数说明：
  -f, --file string   tx file, default: multisign_transfer_prepare.txt (default "multisign_transfer_prepare.txt")
  -k, --keys string   owners' private key, separated by ','

输出
tx is written to file:  sign_multisign_tx.txt
```

* 创建转帐交易--在线操作,需要重新获取 nonce 等信息
```
命令：
./boss4x ethereum offline create_multisign_tx

输出
tx is written to file:  create_multisign_tx.txt
```

* 离线签名交易
```
./boss4x ethereum offline sign -f create_multisign_tx.txt -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
```

* 发送签名后文件
```
./boss4x ethereum offline send -f deploysigntxs.txt
```
