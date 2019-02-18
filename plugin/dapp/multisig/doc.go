// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multisig

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

cli 命令行主要分三块：account 账户相关的，owner 相关的以及tx交易相关的
cli multisig
Available Commands:
  account     multisig account
  owner       multisig owner
  tx          multisig tx

cli multisig  account
Available Commands:
  address     get multisig account address
  assets      get assets of multisig account
  count       get multisig account count
  create      Create a multisig account transaction
  creator     get all multisig accounts created by the address
  dailylimit  Create a modify assets dailylimit transaction
  info        get multisig account info
  owner       get multisig accounts by the owner
  unspent     get assets unspent today amount
  weight      Create a modify required weight transaction

cli multisig  owner
Available Commands:
  add         Create a add owner  transaction
  del         Create a del owner transaction
  modify      Create a modify owner weight transaction
  replace     Create a replace owner transaction

cli multisig  tx
Available Commands:
  confirm          Create a confirm transaction
  confirmed_weight get the weight of the transaction confirmed.
  count            get multisig tx count
  info             get multisig account tx info
  transfer_in      Create a transfer to multisig account transaction
  transfer_out     Create a transfer from multisig account transaction
  txids            get multisig txids


测试步骤如下：
cli seed save -p heyubin -s "voice leisure mechanic tape cluster grunt receive joke nurse between monkey lunch save useful cruise"

cli wallet unlock -p heyubin

cli account import_key  -l miner -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944

cli account create -l heyubin
cli account create -l heyubin1
cli account create -l heyubin2
cli account create -l heyubin3
cli account create -l heyubin4
cli account create -l heyubin5
cli account create -l heyubin6
cli account create -l heyubin7
cli account create -l heyubin8


cli send bty transfer -a 100 -n test  -t 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
cli send bty transfer -a 100 -n test  -t 1Kkgztjcni3xKw95y2VZHwPpsSHDEH5sXF -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
cli send bty transfer -a 100 -n test  -t 1N8LP5gBufZXCEdf3hyViDhWFqeB7WPGdv -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
cli send bty transfer -a 100 -n test  -t 1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt

cli send bty transfer -a 100 -n test  -t "1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj" -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
cli send bty transfer -a 100 -n test  -t "17a5NQTf9M2Dz9qBS8KiQ8VUg8qhoYeQbA" -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
cli send bty transfer -a 100 -n test  -t "1DeGvSFX8HAFsuHxhaVkLX56Ke3FzFbdct" -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
cli send bty transfer -a 100 -n test  -t "166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf" -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
cli send bty transfer -a 100 -n test  -t "1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo" -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt



第一步：1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd地址创建多重签名的账户，owner：1Kkgztjcni3xKw95y2VZHwPpsSHDEH5sXF  1N8LP5gBufZXCEdf3hyViDhWFqeB7WPGdv
//构建交易
cli send multisig account create -d 10 -e coins -s BTY -a "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK-1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj" -w "20-10" -r 15 -k 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd

//查看创建的账户个数
cli multisig account count

//通过账户index获取账户地址
cli multisig account address -e 0 -s 0

//通过账户addr获取账户详情
cli multisig account info -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47"


第二步，向multisig合约中转账 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd

cli exec addr -e "multisig"
14uBEP6LSHKdFvy97pTYRPVPAqij6bteee

// 转账
cli send bty transfer -a 50 -n test  -t 14uBEP6LSHKdFvy97pTYRPVPAqij6bteee -k 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd

cli account balance -a 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd


第三步：从指定账户转账到多重签名地址 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd   --》 "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47"

cli send multisig tx transfer_in -a 40 -e coins -s BTY  -t "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -n test -k 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd


//查看1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd地上有40被转出了
cli account balance -a 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd


//多重签名地址转入40并冻结
cli account balance -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47"

cli multisig  account assets  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47"


第四步：从多重签名账户传出  "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47"  --》1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj  owner:1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK签名
cli send multisig  tx transfer_out  -a 11 -e coins -s BTY -f "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -t 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj -n test -k "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK"


查询账户信息
cli multisig account info -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47"

cli multisig  account assets  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47"

查询接受地址是否收到币
cli multisig  account assets  -a 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj


// 查询交易计数
cli multisig  tx count  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47"

// 查询交易txid
cli multisig   tx txids  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -s 0 -e 0

// 查询交易信息
cli multisig  tx info  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -i 0


//owner "1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj," 转账5个币到1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj    owner:1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj签名
cli send multisig  tx transfer_out  -a 5 -e coins -s BTY -f "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -t 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj -n test -k "1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj"


第五步：测试add/del owner  使用高权重的owner添加交易 "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK",  add 1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo

cli send multisig owner add  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -o 1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo -w 5 -k  "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK"


查看owner的添加
cli multisig  account info -a 13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47

cli multisig  tx info  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -i 0

//del owner
cli send multisig  owner del  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -o "1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo"  -k 1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK
// modify  dailylimit
cli send multisig  account dailylimit -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -e coins -s BTY -d 12 -k 1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK
// modify weight
cli send multisig  account weight -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -w 16 -k 1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK
//replace owner
cli send multisig  owner replace  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -n 166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf -o 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj -k  1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK
// modify owner
cli send multisig  owner modify  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -o "166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf" -w 11 -k 1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK

//获取指定地址创建的所有多重签名账户
cli multisig account creator -a 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd

// 获取指定账户上指定资产的每日余额
cli multisig  account unspent  -a 13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47 -e coins -s BTY

第五步：测试交易的确认和撤销
//权重低的转账，owner：166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf
cli send multisig  tx transfer_out  -a 10 -e coins -s BTY -f "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -t 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj -n test -k "166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf"


//撤销对某笔交易的确认
cli send   multisig tx confirm  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -i 8 -c f  -k 166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf


//确认某笔交易
cli send multisig tx confirm  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -i 8 -k "166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf"

cli send multisig tx confirm  -a "13q53Ga1kquDCqx7EWF8FU94tLUK18Zd47" -i 8 -k "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK"

// 获取owner拥有的所有多重签名地址，不指定地址时返回的是本钱包拥有的所有多重签名地址
cli  multisig account owner -a 166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf
*/
