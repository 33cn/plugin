# 平行链授权节点一键式部署

## 需求
 平行链申请超级节点步骤比较多，需要设置的地方也比较多，容易遗漏
 
## 使用
1. 把编译好的chain33,chain33-cli,chain33.toml,和chain33.para.toml和本脚本目录放到一起
1. 修改config文件配置项，每个配置项都有说明
1. make docker-compose 缺省启动
1. make docker-compose op=nodegroup 配置超级账户组
1. make docker-compose op=wallet 配置钱包开启共识
1. 系统缺省会把区块链数据从docker重映射到本地storage目录，以后默认启动历史数据不会丢失
1. make docker-compose down 停止当前平行链

## 系统重启
1. 系统重启使用 make docker-compose即可，　起来后  make docker-compose op=miner开启挖矿 
 
              