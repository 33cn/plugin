[![pipeline status](https://api.travis-ci.org/33cn/plugin.svg?branch=master)](https://travis-ci.org/33cn/plugin/)
[![Go Report Card](https://goreportcard.com/badge/github.com/33cn/plugin?branch=master)](https://goreportcard.com/report/github.com/33cn/plugin)

# chain33 官方插件系统

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


