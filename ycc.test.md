# ycc 测试说明


## 配置
以下给出各个小结的配置修改。没有给出的**保持不变**

### [p2p]
当前配置5个seeds 节点，分别是

- 119.3.88.160:13801 // 北京
- 49.4.71.232:13801 // 上海
- 47.74.219.81:13801 // 新加坡
- 47.91.45.20:13801 // 澳大利亚
- 47.108.48.151:13801 //成都

在chain33.toml p2p-seeds, 可以根据修改如上的seeds，默认已经配置好国内的3个

### [consensus.sub.pos33]
共识的配置，请根据自己的情况修改以下3个配置：
- listenAddr="0.0.0.0:10901" // 根据自己的需要修改端口
- advertiseAddr="47.108.48.151:10901" // 如果是在NAT内部，请配置自己外部的NAT地址和端口
- bootPeerAddr="49.4.71.232:10901"    // 如同上述p2p的seeds，选择某个seeds的ip，端口保持10901
- nodeID="node.pos33.chengdu"  // 请改为自己的标示。最好是自己钱包的某个钱包地址

## 启动节点
$ nohup ./chain33 &  或者  nohup ./chain33 >/dev/null 2>&1 &  // 后面参数表示日志输出到null

## 启动钱包并开启自动挖矿
$ bash wallet-init.sh

## 充钱购买ticket
$ bash send.sh 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt // 此地址更改上面启动钱包数出的地址


以上为linux下运行。其他系统可以根据脚本自己修改。




