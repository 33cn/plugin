```bash
# 发布合约
./chain33-cli send wasm create -n evidence -p evidence合约路径 -k 用户私钥

# 调用合约
# -p 0 表示传给set方法的参数
# -v "string1","string2","string3" 表示设置三个环境变量，可以在合约内部分别调用getENV(0), getENV(1), getENV(2)读取

./chain33-cli send wasm call -n evidence -m set -p 0 -v "string1","string2","string3"  -k 用户私钥
```

