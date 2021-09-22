#在chain33部署操作手册
##步骤一: 离线创建3笔部署router合约的交易
```
交易1: 部署合约: weth9
交易2: 部署合约: factory
交易3: 部署合约: router

./boss offline chain33 router -f 1 -k 0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944 -n "deploy router to chain33" -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt --chainID 33

-f, --fee float: 交易费设置，因为只是少量几笔交易，且部署交易消耗gas较多，直接设置1个代币即可
-k, --key string: 部署人的私钥，用于对交易签名
-n, --note string: 备注信息 
-a, --feeToSetter: 设置交易费收费地址（说明：该地址用来指定收取交易费地址的地址，而不是该地址用来收取交易费）
--chainID 平行链的chainID
生成交易文件：farm.txt：router.txt

```

##步骤二: 离线创建5笔部署farm合约的交易
```
交易1: 部署合约: cakeToken
交易2: 部署合约: SyrupBar
交易3: 部署合约: masterChef
交易4: 转移所有权，将cake token的所有权转移给masterchef
交易5: 转移所有权: 将SyrupBar的所有权转移给masterchef


./boss offline chain33 farm masterChef -f 1 -k 0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944 -n "deploy farm to chain33" -d 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -s 10 -m 5000000000000000000 --paraName user.p.para.
生成交易文件：farm.txt
```

##步骤三: 离线创建多笔增加lp token的交易
```
./boss offline chain33 farm addPool -f 1 -k 0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944 -p 1000 -l 1HEp4BiA54iaKx5LrgN9iihkgmd3YxC2xM -m 13YwvpqTatoFepe31c5TUXvi26SbNpC3Qq --paraName user.p.para.
```

##步骤四: 串行发送交易文件中的交易
```
./boss offline chain33 send -f xxx.txt
```