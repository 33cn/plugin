# 测试流程

1. 起eth节点

   ```
   cd /opt/src/github.com/cosmos/peggy/testnet-contracts
   
   truffle develop
   ```

   

2. 起chain33节点

   ```
   make docker-compose
   ```

   

3. 修改配置项，起relayer

   修改 chain33Host,BridgeRegistry,pushHost,operatorAddr,deployerPrivateKey,validatorsAddr

   注意：

   BridgeRegistry 需要先起relayer然后部署完成后才有，然后停掉relayer，重新跑

4. 修改脚本中的私钥，跑部署脚本

   ```
   ./bridgeBankTest.sh
   ```

   

5. 在ethereum上发行bty

   ```
   ./ebcli_A relayer ethereum token4chain33 -s bty
   ```

6. 跑测试用例

   ```
   ./test.sh
   ```

7. 查询ispending的prophecy

   ```
   ./ebcli_A relayer ethereum ispending -i 1
   ```

8. 处理这个prophecy

   ```
   ./ebcli_A relayer ethereum process -i 1
   ```

9. 查询余额

   ```
   ./ebcli_A relayer ethereum balance -o 0x7B95B6EC7EbD73572298cEf32Bb54FA408207359 -a 0xbAf2646b8DaD8776fc74Bf4C8d59E6fB3720eddf
   ```

   

