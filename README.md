bityuan 系统

### 安装

##### 1. 安装govendor 工具

```
go get -u -v github.com/kardianos/govendor
```

#### 支持make file的平台

```
go get gitlab.33.cn/chain33/bityuan
make
```
就可以完成编译安装

#### 不支持的平台，可以手工执行下面的命令

```
go get gitlab.33.cn/chain33/bityuan
govendor init
govendor fetch +m
govendor add +e
go build -i -o tool gitlab.33.cn/chain33/chain33/cmd/tools
./tool updateinit --path "plugin" --packname "gitlab.33.cn/chain33/bityuan"
./tool import -path "plugin" --packname "gitlab.33.cn/chain33/bityuan" --conf "plugin/plugin.toml"
go build -i -o bityuan
go build -i -o bityuan-cli gitlab.33.cn/chain33/bityuan/cli
```
