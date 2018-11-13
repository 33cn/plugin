[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/33cn/plugin)
[![pipeline status](https://api.travis-ci.org/33cn/plugin.svg?branch=master)](https://travis-ci.org/33cn/plugin/)
[![Go Report Card](https://goreportcard.com/badge/github.com/33cn/plugin?branch=master)](https://goreportcard.com/report/github.com/33cn/plugin)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/github/33cn/plugin?svg=true&branch=master&passingText=Windows%20-%20OK&failingText=Windows%20-%20failed&pendingText=Windows%20-%20pending)](https://ci.appveyor.com/project/33cn/plugin)

# chain33 官方插件系统

* chain33地址: https://github.com/33cn/chain33
* chain33官网: https://chain.33.cn

## 安装

##### 1. 安装govendor 工具

```
go get -u -v github.com/kardianos/govendor
```

#### 支持make file的平台

```
make
```
就可以完成编译安装

## 运行

```
./chain33 -f chain33.toml
```
注意，默认配置会连接chain33 测试网络

## 注意:

从头开始安装vendor 有非常大的难度，主要问题是带宽 和 翻墙问题
为了解决包依赖等问题，我们直接提供了vendor目录。


