# exchange合约

## 前言
这是一个基于chain33开发的去中心化交易所合约，不收任何手续费，用于满足一小部分人群或者其他特定业务场景中，虚拟资产之间得交换。

## 使用
合约提供了类似中心化交易所健全的查询接口，所有得接口设计都基于用户的角度去出发
合约接口,在线构造交易和查询接口分别复用了框架中的CreateTransaction和Query接口，详情请参考
[CreateTransaction接口](https://github.com/33cn/chain33/blob/master/rpc/jrpchandler.go#L1101)和[Query接口](https://github.com/33cn/chain33/blob/master/rpc/jrpchandler.go#L838)

查询方法名称|功能
-----|----
QueryMarketDepth|获取指定交易资产的市场深度
QueryHistoryOrderList|实时获取指定交易对已经成交的订单信息
QueryOrder|根据orderID订单号查询具体的订单信息
QueryOrderList|根据用户地址和订单状态（ordered,completed,revoked)，实时地获取相应相应的订单详情

可参照exchange_test.go中得相关测试用例，构建limitOrder或者revokeOrder交易进行相关测试

## 注意事项
合约撮合规则如下：

序号|规则
---|----
1|买家获利得原则
2|买单高于市场价，按价格由低往高撮合
3|卖单低于市场价，按价格由高往低进行撮合
4|价格相同按先进先出的原则进行撮合
5|出于系统安全考虑，最大撮合深度为100单，单笔挂单最小为1e8,就是一个bty

