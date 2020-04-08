# accountmanager合约

## 前言
为适配央行发布的[金融分布式账户技术安全规范](http://www.cfstc.org/bzgk/gk/view/yulan.jsp?i_id=1855)，满足联盟链中的金融监管，账户公钥重置，黑白名单等要求，特意在chain33上面开发了
accountmanager合约

## 使用
合约按照中心化金融服务的设计，有管理员对合约下面的账户进行监管
提供如下功能：
功能|内容
----|----
账户创建|普通账户，管理员账户，其他有特殊权限的系统账户，在accountmanager合约中，accountID具有唯一性，可作为身份的标识
账户授权|普通权限在注册时即以授权，特殊权限则需要管理员进行授权
账户冻结和解冻|账户冻结由管理员发起，冻结的账户不能交易，账户下的资产将会被冻结
账户锁定和恢复|用于当私钥遗失，重置外部私钥的情况，需要有一定的锁定期限，在锁定期内不能转移账户下的资产
账户注销| 账户应设使用期限，默认五年时间，过期账户将被注销，提供已注销账户的查询接口
账户资产|账户资产可在accountmanager合约下进行正常的流转

合约接口,在线构造交易和查询接口分别复用了框架中的CreateTransaction和Query接口，详情请参考
[CreateTransaction接口](https://github.com/33cn/chain33/blob/master/rpc/jrpchandler.go#L1101)和[Query接口](https://github.com/33cn/chain33/blob/master/rpc/jrpchandler.go#L838)

查询方法名称|功能
-----|----
QueryAccountByID|根据账户ID查询账户信息，可用于检查账户ID是否注册
QueryAccountsByStatus|根据状态查询账户信息
QueryExpiredAccounts|查询过期时间
QueryAccountByAddr|根据用户地址查询账户信息
QueryBalanceByID|根据账户ID查询账户资产余额


可参照account_test.go中得相关测试用例，构建相关交易进行测试

## 注意事项

**表结构说明**

表名|主键|索引|用途|说明
 ---|---|---|---|---
 account|index|accountID,addr,status|记录注册账户信息|index是复合索引由{expiretime*1e5+index(注册交易所在区块中的索引)}构成

**表中相关参数说明**

参数名|说明
----|----
Asset|资产名称
op|操作类型 分为supervisor op 1为冻结，2为解冻，3增加有效期,4为授权 apply op 1 撤销账户公钥重置, 2 锁定期结束后，执行重置公钥操作
status|账户状态，0 正常， 1表示冻结, 2表示锁定 3,过期注销
level|账户权限 0 普通，其他根据业务需求自定义
index|账户逾期的时间戳*1e5+注册交易在区块中的索引，占位15 %015d
