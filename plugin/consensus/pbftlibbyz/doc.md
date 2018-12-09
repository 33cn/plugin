# pbftlibbyz

## 开发目的和功能说明

开发目的是实现chain33中共识模块中的pbft模块。pbft原作者的主页上给出了一个官方的pbft实现，主要是用C语言实现，这里希望可以用go语言直接调用该pbft库提供的一些接口，实现模块要求。

## 代码说明

代码的实现在plugin/consensus/pbftlibbyz上进行。整体上代码沿用原先的结构，只是把共识部分替换成了调用官方库的接口。bft文件夹就是官方提供的库。

区块链中的节点用Client结构体表示，其中isClient变量表示该节点是client节点或replica节点，这里面的client节点和replica节点沿用的是pbft论文中的概念，client节点负责Propose区块，然后replica节点进行共识返回消息给client节点，client节点再把区块写入。

在CreateBlock()函数里面检查如果不是client节点就进行replica节点相应的初始化接着开始监听来自client的请求。如果是client节点则进行一些client的初始化，接着开始新建一个块。原先是Propose(&newblock)和readReply()两个函数，分别是往channel里写入request和读reply，原先开发的pbft代码会读到channel的request进行共识接着返回reply到channel里。现在把这两个函数合并成了一个ProposeAndReadReply函数，调用官方库的Byz_alloc_request，Byz_invoke函数发送request，这时候直到收到reply函数才会进行下去，接着打印出得到块的日志，pbft的工作就完成了。

最后在cfuns.go里面写了一些replica的helper函数供调用。

## 使用测试

使用docker进行相关测试。原先的pbft单元测试只运行了一个节点，我们这里需要运行5个节点，1个client节点和4个replica节点，在`pbftlibbyz_test.go`中，对于client节点会模拟一些交易，打包区块，而对于replica节点就负责共识部分。

主要的配置工作是节点的IP。在`bft/config`下配置节点的IP，以及在5个`.toml`文件中配置好节点IP和节点编号。

具体使用上，首先在plugin/consensus/pbftlibbyz目录下，运行`./build-docker.sh`编译出docker镜像，然后运行`./run-docker.sh`运行5个docker容器(容器名为replica-1,replica-2,replica-3,replca-4,client)，使用`docker attach client`就可以看到client节点打包区块等日志输出。





