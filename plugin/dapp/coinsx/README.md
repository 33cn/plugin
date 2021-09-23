# coinsx合约

## 前言
1. 为配合BSN联盟链对p2p原生token转账限制规则，为了不影响既有的coins合约，增加了coinsx合约，
2. coinsx合约转账，创世等功能和coins完全一致，增加了对转账的管理功能


## 使用
1. coinsx执行器和coins执行器不同，通过toml配置文件coinExec=“”配置，缺省是coins执行器
1. 如果配置为coinExec="coinsx",则原生代币为coinsx合约，创世到coinsx合约
1. coins cli 会根据配置文件修改相应执行器，transfer，withdraw等命令和coins一致， 
1. json-rpc 构建交易需要明确采用配置的coinExec，缺省是coins
1. 平行链资产转移
   1. 旧的接口assetTransfer/assetWithdraw 仍只接受coins，如果主链是coinsx，则会失败，需要使用新接口
   1. 新接口crossTransfer，通过必填的交易执行器参数保证了平行链铸造的资产和主链保持一致，而不是和平行链配置一致
        1. 比如主链coinsx，平行链缺省代币是coins, 主链转移到平行链资产仍为coinsx.bty
        1. 比如主链coins, 平行链缺省代币是coinsx, 平行链转移到主链资产为 coinsx.symbol

## coinsx 管理功能
1. 只有节点超级管理员可以配置
1. 配置转账使能和受限标志
1. 配置链上管理员组，增删管理员

## p2p转账限制规则
1. 系统缺省转账受限
1. 节点超级管理员from或to 转账都不受限
1. 如果配置了转账使能，则任何用户转账不受限
1. 如果配置了转账受限功能，则和超级管理员转账不受限，个人之间转账受限