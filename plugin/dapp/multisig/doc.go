// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
多重签名合约实现功能：
创建一个多重签名的账户，并指定owner已经权重
账户属性交易：
	owner的add/del/modify/replace
	资产每日限额的修改
	请求权重的修改
账户属性的修改都是一个交易，需要满足权重的要求才能被执行。

账户交易的确认和撤销：当交易提交者的权重不能满足权重要求时，需要其余的owner来一起确认。owner可以撤销自己对某笔交易的确认，但此交易必须是没有被执行。已执行的不应许撤销

多重签名账户的转入和转出：转入时，to地址必须是多重签名地址，from地址必须是非多重签名地址；
						 转出时，from地址必须是多重签名地址，to地址必须是非多重签名地址； 传出交易需要校验权重



第一：创建一个多重签名的账户地址（创建交易的txhash 生成对应的addr），创建时必须指定两个初始的owner 以及requiredWeight权重，owner的权重之和必须不能小于requiredWeight

此多重签名账户拥有的属性：

owner列表：owner 地址已经weight

assets dailyLimit ： "execer": 和"symbol": 以及对应"dailyLimit" 。
	注释：当从此账户转出的额度小于每日限额时，不需要校验owner的权重是否大于requiredWeight权重。
		当转出额度超过每日限额时，需要多个owner一起来确认此交易，并且确认此交易的owner的权重之和必须大于requiredWeight权重，校验才能被执行。

"requiredWeight":  交易额度超过每日限额只有需要的权重。

//cli命令行：构造交易并签名发送
cli multisig create -d 10 -e coins -s bty -a "1Kkgztjcni3xKw95y2VZHwPpsSHDEH5sXF 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj" -w "20 10" -r 15

cli wallet sign -a “创建者地址” -d
cli wallet send -d


//查看创建的账户个数
cli multisig get_account_count


//通过账户index获取账户地址，此时的index就是上面get_account_count的返回值。index从0开始的。为了以后分区间获取账户信息
cli multisig get_accounts -e 0 -s 0

//通过账户addr获取账户详情，142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc 多重签名账户地址
cli multisig get_acc_info -a "142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc"


第二步：给指定的多重签名账户142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc添加owner：1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo 。使用已存在的owner 提交交易1Kkgztjcni3xKw95y2VZHwPpsSHDEH5sXF
cli multisig owner_add  -a "142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc" -o 1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo -w 5

//del owner
cli multisig  owner_del  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -o "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK"

//replace owner
cli multisig  owner_replace  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -n 1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK -o 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj
// modify owner
cli multisig  owner_modify  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -o "1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo" -w 11

cli wallet sign -a 1Kkgztjcni3xKw95y2VZHwPpsSHDEH5sXF -d
cli wallet send -d


第三步：多重签名账户交易查询：

// 查询多重签名账户上的交易计数，需要指定多重签名账户地址
cli multisig  get_tx_count  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc
{
    "data": 1
}
// 查询多重签名账户交易，需要指定多重签名账户地址以及txid  索引区间，可以指定查询交易的执行状态
cli multisig  get_txids  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -s 0 -e 0

// 查询交易信息
cli multisig  get_tx_info  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -i 0



第四部：请求权重和资产每日限额的修改：
// modify  dailylimit  修改coins：bty的每日限额为11个
cli multisig  dailylimit -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -e coins -s bty -d 11
//增加token：TEST的资产每日限额为11个
cli multisig  dailylimit -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -e token -s TEST -d 11

// modify weight 修改请求权重的值。
cli multisig  weight_modify -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -w 16



第五步：账户交易的确认和撤销
//确认某笔交易，只能确认没有被执行的交易
cli multisig confirm  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -i 9

//撤销对某笔交易的确认，只能撤销没有被执行的交易
cli multisig confirm  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -i 9 -c f




第六步：多重签名账户的资产转入和转出
	//转入：
	//首先1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd账户需要向 multisig合约中转币
	cli exec addr -e "multisig"
	14uBEP6LSHKdFvy97pTYRPVPAqij6bteee

	cli send bty transfer -a 50 -n test  -t 14uBEP6LSHKdFvy97pTYRPVPAqij6bteee -k 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd

	//然后才能在multisig合约中从1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd账户转币到多重签名账户142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc
	cli multisig transfer_in -a 40 -e coins -s bty  -t 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -n test

	cli wallet sign -a 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd -d

	cli wallet send -d

	//此时可以查看多重签名账户142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc的资产，转入的币都被冻结
	cli multisig  get_assets  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc

	//转出：在multisig合约中从多重签名账户142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc转币到1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj账户。由owner 1Kkgztjcni3xKw95y2VZHwPpsSHDEH5sXF提交交易
	cli multisig  transfer_out  -a 11 -e coins -s bty -f 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc -t 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj -n test
	cli wallet sign -a 1Kkgztjcni3xKw95y2VZHwPpsSHDEH5sXF -d
	cli wallet send -d

	//此时可以查看多重签名账户142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc的资产，冻结币有减少
	cli multisig  get_assets  -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc
	cli account balance -a 142YMLZKZr3aBeiQcNqbSGkcj48ctka1tc

	//查看1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj账户在multisig合约中的余额有增加
	cli multisig  get_assets  -a 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj
	cli account balance -a 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj


	//从 multisig合约中取出1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj账户上的币
	cli bty withdraw  -a 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj -e multisig
*/
package multisig
