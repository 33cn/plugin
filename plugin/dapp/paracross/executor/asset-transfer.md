# paracross 执行器 跨链交易之 资产转移方案

## 逻辑介绍

交易功能， 把资产从主链移动到平行链， 或从平行链将资产移动回主链。

账户模型
 1. 合约资产移动总帐号: 记录所有从主链转移出去的资产
 1. 平行链资产总帐号 :  记录从主链转移到某个平行链的资产
 1. 用户主链的paracross合约帐号： 同普通的约合帐号
 1. 用户平行链paracross合约帐号： 同普通的约合帐号

资产流转
 1. A: conis/token -> 主链paracross合约帐号
 1. A: 主链paracross合约帐号 -> 平行链paracross合约帐号
 1. A: 平行链paracross合约帐号 -> 使用
 1. B(可能也是A): 平行链paracross合约帐号 -> 主链paracross合约帐号

## 实现上的说明

这个交易有点特殊， 需要同时在主链和某平行链上执行
 1. 为了容易认出， 暂定方案 执行器名字带上另行链title：如 user.p.guodun.paracross， 也就是说需要在主链上能执行这样的交易

关于交易序号的说明
 1. 由于平行链上加了挖矿交易， 交易列表不再是全部从主链区块中过滤
 1. 挖矿交易记录的交易结果的bitmap， 再主链根据主链区块高度，重新获取交易，然后找到对应bit　获取执行结果

## 交易

asset-transfer 分两种， 主链转出， 主链转入


主链转出执行 transfer
 * 主链
   1. 用户主链paracross合约帐号， balance -
   1. 某平行链paracross合约帐号， balance +
 * 平行链(如果上面步骤失败，， 平行链会过滤掉这个交易)
   1. 平行链中 用户paracross合约帐号  balance +

主链转入 withdraw
 * 平行链
   1. 平行链中 用户paracross合约帐号  balance -
 * 主链(上面步骤失败， 不进行操作)
   1. commit 交易共识时执行
   1. 某平行链paracross合约帐号， balance -
   1. 用户主链paracross合约帐号， balance +

cross-transfer 把transfer和withdraw统一为transfer，　通过统一地址符号，内部判断是transfer和withdraw
统一地址符号：title+执行器+符号
1. 主链title缺省为空，类似：coins.bty, token.FZM
1. 平行链符号为: user.p.test.coins.para 或user.p.game.token.FZM

主链转移资产场景： type=0, tx.exec=user.p.test.
1. 主链本币转移：	symbol:{coins/token}.bty/cny or bty/cny,
   平行链资产：	paracross-coins.bty
2. 主链外币转移: 	symbol: user.p.para.coins.ccny,
   平行链资产:	paracross-paracross.user.p.para.coins.ccny
3. 平行链本币提回: symbol: user.p.test.coins.ccny
   平行链资产：	paracross账户coins.ccny资产释放
平行链转移资产场景：type=1,tx.exec=user.p.test.
1.　平行链本币转移:	symbol:user.p.test.{coins/token}.ccny
    主链产生资产:	paracross-user.p.test.{coins}.ccny
2.  主链外币提取:	symbol: user.p.para.coins.ccny
    主链恢复外币资产:user.p.test.paracross地址释放user.p.para.coins.ccny
3.  主链本币提取:	symbol: coins.bty
    主链恢复本币资产:user.p.test.paracross地址释放coin.bty


交易执行代码分为 三个部分
 1. 主链
 1. 平行链
 1. 共识时

### kv，日志， 查询

 1. kv 帐号变化
    1. 用户合约帐号
    1.  合约资产帐号 (主链有， 平行链无)
    1. 平行链资产帐号
 1. 日志
    1. 对应帐号变化的日志
 1. 查询
    1. 查询： 合约资产帐号/平行链资产帐号

```
帐号      主链部分                                                     平行链部分
account     主链合约    平行链帐号在主链    用户A在主链       用户B在主链     平行链合约   用户A    用户B
1 A转账5bty    5           0                5              0               0       0           0
到para合约
2 B转账5bty    10          0                5              5               0       0           0
到para合约
3 A跨链4bty    10          4                1              5               0       0           0      主链执行完
到平行链        10          4                1              5               4       4           0      平行链执行完
4 A用3bty      10          4                1              5               1       1           0
在平行链
5  B赚到2bty   10          4                1              5               3       1           2
在平行链， 并把2bty转到合约
6 B从平行链     10          4                1              5               3       1           2       主链打包
提币 1bty      10          4                1              5               2       1           1       平行链执行完
              10           3                1              6               2       1           1       主链共识完
```

