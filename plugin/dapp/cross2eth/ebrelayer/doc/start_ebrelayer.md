##  启动 relayer
[TOC]

### 启动 relayer A
#### 完成 ethererum 和 chain33 相关合约的部署
得到 BridgeRegistryOnChain33, BridgeRegistryOnEth, multisignChain33Addr, multisignEthAddr 4个合约地址。

#### 修改 relayer.toml 配置文件
|字段|说明|
|----|----|
|pushName|4 个 relayer 不同相同, `sed -i 's/^pushName=.*/pushName="XXX"/g' relayer.toml`|
|ethProvider|ethereum 的 socket 通信地址, 例如: wss://rinkeby.infura.io/ws/v3/404eb4acc421426ebeb6e92c7ce9a270, 如果有多个就根据 EthChainName 分别配置|
|EthProviderCli|ethereum 的 http url 地址, 例如: https://rinkeby.infura.io/ws/v3/404eb4acc421426ebeb6e92c7ce9a270, 如果有多个就根据 EthChainName 分别配置|
|BridgeRegistry|部署在 ethereum 的 BridgeRegistry 地址, 如果有多个就根据 EthChainName 分别配置|
|chain33BridgeRegistry|部署在 chain33 的 BridgeRegistry 地址|
|ChainID4Chain33|chain33 链的 ID, 默认为 0|
|ChainName|链的名称, 用来区分主链和平行链, 如user.p.xxx., 必须包含最后一个点|
|chain33Host|平行链的 host 地址, 默认: http://localhost:8801|
|pushHost|relayer 的 host 地址, 默认: http://localhost:20000|
|pushBind|relayer 的 bind 端口, 默认: 0.0.0.0:20000|

#### 首次启动 relayer 进行设置
```shell
# 设置密码
./ebcli_A set_pwd -p 密码
# 解锁
./ebcli_A unlock -p 密码

# 设置 chain33 验证私钥
./ebcli_A chain33 import_privatekey -k "${chain33ValidatorKeya}"
# 设置 ethereum 验证私钥
./ebcli_A ethereum import_privatekey -k "${ethValidatorAddrKeya}"

# 设置 chain33 多签合约地址
./ebcli_A chain33 multisign set_multiSign -a "${chain33MultisignAddr}"
# 设置 ethereum 多签合约地址
./ebcli_A ethereum multisign set_multiSign -a "${ethereumMultisignAddr}"
```

#### 运行持续启动 relayer
编写脚步 startRelyer.sh
```shell
#!/usr/bin/env bash
# shellcheck disable=SC2050
# shellcheck source=/dev/null
set -x
set +e

while [ 1 == 1 ]; do
    pid=$(ps -ef | grep "./ebrelayer" | grep -v 'grep' | awk '{print $2}' | xargs)
    while [ "${pid}" == "" ]; do
        time=$(date "+%m-%d-%H:%M:%S")
        nohup "./ebrelayer" >"./ebrelayer${time}.log" 2>&1 &
        sleep 2

        ./ebcli_A unlock -p 密码
        sleep 2

        pid=$(ps -ef | grep "./ebrelayer" | grep -v 'grep' | awk '{print $2}' | xargs)
    done
    sleep 2
done
```
启动脚本 `nohup ./startRelyer.sh 2>&1 &`

为了安全起见建议手动调用 `./ebcli_A unlock -p 密码`

### 启动 relayer B C D
#### 修改 relayer.toml 配置文件
先 cp relayerA 的配置文件, 然后修改以下字段:

|字段|说明|
|----|----|
|pushName|4 个 relayer 不同相同, `sed -i 's/^pushName=.*/pushName="XXX"/g' relayer.toml`|
|chain33Host|平行链的 host 地址, 默认: http://localhost:8801, 4 个 relayer 对应 4 个不同 chain33 平行链地址|

#### 首次启动 relayer 进行设置
```shell
# 设置密码
./ebcli_A set_pwd -p 密码
# 解锁
./ebcli_A unlock -p 密码

# 设置 chain33 验证私钥
./ebcli_A chain33 import_privatekey -k "${chain33ValidatorKeya}"
# 设置 ethereum 验证私钥
./ebcli_A ethereum import_privatekey -k "${ethValidatorAddrKeya}"
```

***

### 启动代理 relayer proxy
#### 修改 relayer.toml 配置文件
先 cp relayerA 的配置文件, 然后修改以下字段:

|字段|说明|
|----|----|
|pushName|4 个 relayer 不同相同, `sed -i 's/^pushName=.*/pushName="XXX"/g' relayer.toml`|
|ProcessWithDraw|改为 true|
|chain33Host|平行链的 host 地址, 默认: http://localhost:8801, 选任意一个 chain33 平行链地址就可以|
|RemindUrl|代理打币地址不够时, 提醒打币发送短信的URL|

#### 首次启动 relayer 进行设置
同上...

#### 设置 chain33 代理地址, 及手续费设置
```shell
# 设置 withdraw 的手续费及每日转帐最大值, 实时变动, 价格波动大的时候重新设置
./ebcli_A ethereum cfgWithdraw -f 0.2 -s ETH -a 100 -d 18

Flags:
  -a, --amount float    每日最大值
  -d, --decimal int8    token 精度
  -f, --fee float       手续费
  -s, --symbol string   symbol
  
# 设置 chain33 代理地址
./boss4x chain33 offline set_withdraw_proxy -c "${chain33BridgeBank}" -a "${chain33Validatorsp}" -k "${chain33DeployKey}" -n "set_withdraw_proxy:${chain33Validatorsp}"
Flags:
  -a, --address string    withdraw address
  -c, --contract string   bridgebank contract address
  -f, --fee float         contract gas fee (optional)
  -k, --key string        the deployer private key
  -n, --note string       transaction note info (optional)
```

### 停止或升级 relayer
直接 kill
```shell
ps -ef | grep ebrelayer
root      3661 21631  0 17:58 pts/0    00:00:00 grep --color=auto ebrelayer
root      4066  4057  0 15:47 pts/0    00:00:05 ./ebrelayer
kill 4066
```
如果是用持续启动方式`nohup ./startRelyer.sh 2>&1 &`, 要先 kill startRelyer.sh, 再 kill ebrelayer