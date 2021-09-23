### 合约开发
新建 cpp 和 hpp 文件，并导入 common.h 头文件，其中 common.h 中声明了 chain33 中的回调函数，是合约调用 chain33 系统方法的接口。   

合约中的导出方法的所有参数都只能是数字类型，且必须有一个数字类型的返回值，其中非负值表示执行成功，负值表示执行失败。

### 合约编译

#### Emscripten 环境安装
```bash
git clone https://github.com/juj/emsdk.git
cd emsdk
git checkout 6adb624e04b0c6a0f4c5c06d3685f4ca2be7691d # 用旧版本
./emsdk install latest
./emsdk activate latest

# 安装完之后需要配置临时环境变量，新建终端时需要重新执行以下命令

# on Linux or Mac OS X
source ./emsdk_env.sh

# on Windows
emsdk_env.bat
```

#### 使用 Emscripten 编译合约

```bash
em++ -o dice.wasm dice.cpp -s WASM=1 -O3 -s EXPORTED_FUNCTIONS="[_startgame, _deposit, _play, _draw, _stopgame]" -s ERROR_ON_UNDEFINED_SYMBOLS=0
```

- em++ 是 wasm 编译器，用来编译 c++ 代码，编译c代码可以用 emcc  
- -o 指定输出合约文件名，格式为 .wasm   
- dice.cpp 是源文件   
- -s WASM=1 指定生成 wasm 文件   
- -O3 优化等级，可以优化生成的 wasm 文件大小，可指定为 1～3   
- -s EXPORTED_FUNCTIONS 指定导出方法列表，以逗号分隔，合约外部只可以调用导出方法，不能调用非导出方法，导出方法的函数名需要额外加一个 "_"
- -s ERROR_ON_UNDEFINED_SYMBOLS=0 忽略未定义错误，因为 common.h 中声明的方法没有具体的 c/c++ 实现，而是在 chain33 中用 go 实现，因此需要忽略该错误，否则将编译失败

参考文档：https://developer.mozilla.org/en-US/docs/WebAssembly

#### 安装 wabt（the WebAssembly Binary Toolkit）

```bash
git clone --recursive https://github.com/WebAssembly/wabt
cd wabt
mkdir build
cd build
cmake ..
cmake --build .
```

```bash
# 通过 wasm 文件生成接口abi，abi中可以通过import关键字找到外部导入方法，以及export关键字找到编译时指定的导出方法。
wabt/bin/wasm2wat dice.wasm
```

### 发布合约
```bash
# 若合约已存在则会创建失败，可以换一个合约名发布
./chain33-cli send wasm create -n 指定合约名 -p wasm合约路径 -k 用户私钥
```

### 检查合约发布结果
```bash
# 检查链上是否存在该合约
./chain33-cli wasm check -n 合约名
```

### 更新合约
```bash
# 更新合约要求合约已存在，且只有合约创建者有更新权限
./chain33-cli send wasm update -n 指定合约名 -p wasm合约路径 -k 用户私钥
```

### 调用合约
```bash
#其中参数为用逗号分隔的数字列表，字符串参数为逗号分隔的字符串列表
./chain33-cli send wasm call -n 发布合约时指定的合约 -m 调用合约方法名 -p 参数 -v 字符串参数 -k 用户私钥  
```

### 查询合约数据
```bash
# 查询statedb
./chain33-cli wasm query state -n 合约名 -k 数据库key  

# 查询localdb
./chain33-cli wasm query local -n 合约名 -k 数据库key  
```

### 转账及提款
```bash
#部分合约调用可能需要在合约中有余额，需要先转账到 wasm 合约
./chain33-cli send coins send_exec -e wasm -a 数量 -k 用户私钥

#提款
./chain33-cli send coins withdraw -e wasm -a 数量 -k 用户私钥
```
### RPC接口调用
```bash
#构造调用合约的交易
curl http://localhost:8801 -ksd '{"method":"wasm.CallContract", "params":[{"contract":"dice","method":"play","parameters":[1000000000, 10]}]}'

#签名
curl http://localhost:8801 -ksd '{"method":"Chain33.SignRawTx","params":[{"privkey":"0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01","txhex":"0x0a047761736d1218180212140a04646963651204706c61791a068094ebdc030a20a08d0630ec91c19ede9ef4d1693a22314b3732554137393845775a66427855546b4265686864766b656f3277377446344c","expire":"300s"}]}'

#发送
curl http://localhost:8801 -ksd '{"method":"Chain33.SendTransaction","params":[{"data":"0a047761736d1218180212140a04646963651204706c61791a068094ebdc030a1a6d080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a46304402201dc04e89da9220e42b2768a23cd2e6a7c452b2bfd30e0799f5c6f1b035151d1402201160929f74feb26be4205cf4432bdf377eb775f189db2883556cedc31c4fb01920a08d0628b2cb90fb0530ec91c19ede9ef4d1693a22314b3732554137393845775a66427855546b4265686864766b656f3277377446344c"}]}'
```