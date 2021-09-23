```bash
# 发布合约
./chain33-cli send wasm create -n evidence -p evidence合约路径 -k 用户私钥

# 调用合约
# -v "string1","string2" 表示设置2个环境变量，可以在合约内部分别调用getENV(0), getENV(1)读取

./chain33-cli send wasm call -n evidence -m AddStateTx -v "string1","string2"  -k 用户私钥
```

