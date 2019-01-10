// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package js

/*
Java Script VM contract

cli 命令行
Available Commands:
  call        call java script contract
  create      create java script contract
  query       query java script contract

cli jsvm create
Flags:
  -c, --code string   path of js file,it must always be in utf-8.
  -h, --help          help for create
  -n, --name string   contract name

cli jsvm call
Flags:
  -a, --args string       json str of args
  -f, --funcname string   java script contract funcname
  -h, --help              help for call
  -n, --name string       java script contract name

cli jsvm query
Flags:
  -a, --args string       json str of args
  -f, --funcname string   java script contract funcname
  -h, --help              help for query
  -n, --name string       java script contract name

测试步骤：
第一步：创建钱包
cli seed save -p heyubin -s "voice leisure mechanic tape cluster grunt receive joke nurse between monkey lunch save useful cruise"

cli wallet unlock -p heyubin

cli account import_key  -l miner -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944

第二步：创建名为test的js合约，合约代码使用test.js代码文件（必须是utf-8格式）
cli send jsvm create -c "../plugin/dapp/js/executor/test.js"  -n test -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt

第三步：调用test合约的hello函数
cli send jsvm  call -a "{\"hello\": \"world\"}" -f hello -n test -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt

第四步：query test合约hello函数
cli jsvm  query -a "{\"hello\": \"world\"}" -f hello -n test
*/
