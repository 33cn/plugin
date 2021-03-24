# 平行链共识节点挖矿奖励规则
>平行链共识节点是参与平行链共识安全的节点，发送共识交易，同时享受平行链挖矿奖励

## 共识节点资格
1. 共识节点需要加入共识节点账户组才具有发送共识交易的资格，不然即使发送也不会被接受
1. 新的共识节点需要共识节点账户组成员超过2/3投票通过才可以加入共识节点账户组

## 共识节点挖矿奖励
1. 共识节点根据本地计算的区块哈希发送共识交易参与共识，共识达成后促使达成共识的节点享受挖矿奖励
1. 比如共识账户组有4个挖矿账户A,B,C,D，其中A,B,C，D都依次发送共识交易，基于超过2/3的规则，在C的共识交易收到后即达成共识，那么奖励将分给A,B,C，而D是在达成共识后发送的，不享受挖矿奖励

## 绑定共识节点挖矿（矿池的概念)
1. 如果账户不是共识节点，但是想参与挖矿奖励，可以锁定平行链基础coins到paracross合约，并绑定到共识节点参与挖矿
1. 绑定账户可以绑定到一个或多个共识节点参与挖矿
1. 挖矿奖励按锁定coins数量的权重来分配挖矿奖励
1. 如果绑定的共识节点在某高度没有得到挖矿奖励，对应的绑定账户也得不到相应奖励
1. 绑定账户可以通过bind命令的coins 数量的修改来增加或减少锁定coins数量，可以通过unbind命令解除对某个共识节点的绑定

## 奖励规则和金额
>奖励规则和金额可配置
```
    [mver.consensus.paracross]
    #超级节点挖矿奖励
    coinReward=18
    #发展基金奖励
    coinDevFund=12
    #如果超级节点上绑定了委托账户，则奖励超级节点coinBaseReward，其余部分(coinReward-coinBaseReward)按权重分给委托账户
    coinBaseReward=3
    #委托账户最少解绑定时间(按小时)
    unBindTime=24
```
1. 每个区块产生的挖矿总奖励如配置项是coinDevFund+coinReward=30，共识达成后，发展基金账户分走12，剩余的18个coin平均分给达成共识的节点
1. 如果有绑定挖矿的账户绑定了共识节点进行挖矿，则共识节点平分基础的coinBaseReward，剩余部分(coinReward-coinBaseReward)按绑定挖矿锁定币的权重数量分给绑定挖矿的节点。

## 绑定挖矿命令
1. 生成 绑定/解绑定 挖矿 的交易（未签名）

    ```
    {
        "method" : "Chain33.CreateTransaction",
        "params" : [
            {
            "execer" : "{user.p.para}.paracross",
            "actionName" : "ParaBindMiner",
            "payload" : {
    　　　　　　　"bindAction":"1"
                "bindCoins" : 5,
                "targetNode" : "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
            }
            }
        ],
    }
    ```

    **参数说明：**

    |参数|类型|说明|
    |----|----|----|
    |method|string|Chain33.CreateTransaction|
    |execer|string|必须是平行链的执行器user.p.para.paracross,title:user.p.para.按需调整|
    |actionName|string|ParaBindMiner|
    |bindAction|string|绑定:1，解绑定:2|
    |bindCoins|int|绑定挖矿冻结币的份额，需冻结平行链原生代币，解绑定不需要此配置|
    |targetNode|string|绑定目标共识节点，需要是共识账户组的成员|

## 挖矿奖励的转出
1.查询挖矿奖励
>挖矿产生的奖励在平行链的paracross 执行器中
    ```
    {
        "method": "Chain33.GetBalance",
        "params": [{
            "addresses": ["{共识账户地址}"],
            "execer": "user.p.para.paracross"
        }]
    }

    1. cli命令方法
    ./chain33-cli --rpc_laddr http://localhost:8901 account balance -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    {
        "addr": "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
        "execAccount": [
            {
                "execer": "user.p.para.paracross",
                "account": {
                    "balance": "2227.0000",
                    "frozen": "0.0000"
                }
            }
        ]
    }
    
    2. rpc方法:
    curl -ksd '{"method":"Chain33.GetBalance","params":[{"addresses":["1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"],"execer":"user.p.para.paracross"}]}' http://172.28.0.2:8901
    响应：
    {
        "result": [{
            "currency": 0,
            "balance": 227500000000,
            "frozen": 0,
            "addr": "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
        }],
    }
    

    
    ```
2.转出挖矿奖励
>需要从平行链执行器paracross下把奖励 withdraw出到平行链coins合约的签名地址下
```
1. cli命令方式:
./chain33-cli --rpc_laddr http://localhost:8801 --paraName {平行链title} send coins withdraw -a {数量} -e user.p.para.paracross -k ${私钥}

例:
./chain33-cli --rpc_laddr http://localhost:8801 --paraName user.p.para. send coins withdraw -a 2000000000 -e user.p.para.paracross -k ${私钥}

响应：
./chain33-cli --rpc_laddr http://localhost:8901 account balance -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
{
    "addr": "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
    "execAccount": [
        {
            "execer": "user.p.para.paracross",
            "account": {
                "balance": "1032.0000",
                "frozen": "0.0000"
            }
        },
        {
            "execer": "user.p.para.coins",
            "account": {
                "balance": "2020.0000",
                "frozen": "0.0000"
            }
        }
    ]
}
注:user.p.para.coins下就是自己的余额

rpc方法
1.创建交易:
{
	"method": "Chain33.CreateRawTransaction",
	"params": [{
		"to": "19WJJv96nKAU4sHFWqGmsqfjxd37jazqii",
		"amount": 2000000000,
		"fee": 2000000,
		"isWithdraw": true,
		"execName": "user.p.para.paracross",
		"execer": "user.p.para.coins"
	}]
}
注释：
    1) "to": "19WJJv96nKAU4sHFWqGmsqfjxd37jazqii", 平行链paracross执行器地址，不需要修改
    2) amount,fee需要自己设置

2.签名
{
	"method": "Chain33.SignRawTx",
	"params": [{
		"privkey": "{私钥}",
		"txHex": "{交易数据}",
		"expire": "120s"
	}]
}
3.发送交易
{
	"method": "Chain33.SendTransaction",
	"params": [{
		"data": "{签名数据}"
	}]
}


```

