###  离线部署 ethereum 跨链合约及各操作
*** 

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
交易7: 设置 symbol
交易8: 部署合约: BridgeRegistry
交易9: 部署合约: MulSign

命令：
./boss4x ethereum offline create -s "ETH" -p "25,25,25,25" -o "${ethDeployAddr}" -v "${ethValidatorAddra},${ethValidatorAddrb},${ethValidatorAddrc},${ethValidatorAddrd}"
    
参数说明：
  -p, --initPowers string        验证者权重, as: '25,25,25,25'
  -o, --owner string             部署者地址
  -s, --symbol string            symbol
  -v, --validatorsAddrs string   验证者地址, as: 'addr,addr,addr,addr'

  --rpc_laddr_ethereum string    ethereum url 地址 (默认 "http://localhost:7545")

输出:
tx is written to file:  deploytxs.txt

把交易信息写入文件中
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
***

* 输出
```
交易4: 部署合约: BridgeBank
交易8: 部署合约: BridgeRegistry
交易9: 部署合约: MulSign
```

#### 设置离线多签地址信息
* 在线创建交易
```
命令：
./boss4x ethereum offline multisign_setup -m "${multisignEthAddr}" -d "${ethDeployAddr}" -o "${ethMultisignA},${ethMultisignB},${ethMultisignC},${ethMultisignD}"

参数说明：
  -d, --deployAddr string   部署者地址
  -m, --multisign string    离线多签合约地址
  -o, --owner string        多签的地址, 用','分隔

输出
tx is written to file:  multisign_setup.txt
```

***

#### 设置 bridgebank 合约地址可以转到多签合约地址
* 在线创建交易
```
命令：
./boss4x ethereum offline set_offline_addr -a "${multisignEthAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"

参数说明：
  -a, --address string      离线多签地址
  -c, --contract string     bridgebank 合约地址
  -d, --deployAddr string   部署合约地址

输出
tx is written to file:  set_offline_addr.txt
```

***

#### 设置 symbol 允许被 lock
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

***

#### 创建 bridge token
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
***

#### 设置 bridgebank 金额到多少后自动转入多签合约地址
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
***

#### 测试离线部署 USDT ERC20 合约
* 在线创建交易
```
命令：
./boss4x ethereum offline create_tether_usdt -m 33000000000000000000 -s USDT -d "${ethTestAddr1}"

参数说明：
  -m, --amount string       金额
  -d, --owner string        拥有者地址
  -s, --symbol string       erc20 symbol

输出
tx is written to file:  deployTetherUSDT.txt
```
***

#### 离线部署 ERC20 跨链合约
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
***

#### 离线多签转帐
* 转帐预备交易--在线操作
```
命令：
./boss4x ethereum offline multisign_transfer_prepare -a 8 -r "${ethBridgeBank}" -c "${multisignEthAddr}" -d "${ethTestAddr1}" -t "${ethereumUSDTERC20TokenAddr}"

参数说明：
  -a, --amount float      转帐金额
  -c, --contract string   离线多签合约地址
  -r, --receiver string   接收者地址
  -d, --sendAddr string   发送这笔交易的地址, 需要扣除部分手续费
  -t, --token string      erc20 地址, 空的话，默认转帐 ETH

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
