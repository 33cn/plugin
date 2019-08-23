[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/33cn/plugin)
[![pipeline status](https://api.travis-ci.org/33cn/plugin.svg?branch=master)](https://travis-ci.org/33cn/plugin/)
[![Go Report Card](https://goreportcard.com/badge/github.com/33cn/plugin?branch=master)](https://goreportcard.com/report/github.com/33cn/plugin)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/github/33cn/plugin?svg=true&branch=master&passingText=Windows%20-%20OK&failingText=Windows%20-%20failed&pendingText=Windows%20-%20pending)](https://ci.appveyor.com/project/33cn/plugin)
[![codecov](https://codecov.io/gh/33cn/plugin/branch/master/graph/badge.svg)](https://codecov.io/gh/33cn/plugin)

# chain33 官方插件系统

* chain33地址: https://github.com/33cn/chain33
* chain33官网: https://chain.33.cn

### 环境

```
需要安装golang1.12 or latest
```

#### 支持make file的平台

```
// 开启mod功能
export GO111MODULE=on

//国内用户需要导入阿里云代理，用于下载依赖包
export GOPROXY=https://mirrors.aliyun.com/goproxy

make
```
就可以完成编译安装

```
注意：国内用户需要导入一下代理，才能获取依赖包，mod功能在Makefile中默认开启
```

## 运行

```
./chain33 -f chain33.toml
```
注意，默认配置会连接chain33 测试网络

## 注意:

使用mod管理依赖包，主要就是翻墙问题
为了解决包依赖翻墙下载问题，我们提供了阿里云代理。


## 贡献代码：

详细的细节步骤可以见 https://github.com/33cn/chain33
这里只是简单的步骤：

#### 准备阶段:

* 首先点击 右上角的 fork 图标， 把chain33 fork 到自己的分支 比如我的是 vipwzw/plugin
* `git clone https://github.com/vipwzw/plugin.git $GOPATH/src/github.com/33cn/plugin`

```
注意：这里要 clone 到 $GOPATH/src/github.com/33cn/plugin, 否则go 包路径会找不到
```

clone 完成后，执行
```
make addupstream
```

#### 创建分支准备开发新功能

```
make branch b=branch_dev_name
```
#### 提交代码

```
make push b=branch_dev_name m="hello world"
```
如果m不设置，那么不会执行 git commit 的命令

#### 测试代码
类似plugin/dapp/relay,在cmd目录下编写自己插件的Makefile和build.sh
在build目录下写testcase和相关的Dockerfile和docker-compose配置文件,
testcase的规则参考plugin/dapp/testcase_compose_rule.md

用户可以在travis自己工程里面设置自己plugin的DAPP变量，如DAPP设置为relay，则travis里面run relay的testcase

