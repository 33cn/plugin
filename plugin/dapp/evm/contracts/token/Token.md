# TOKEN 预编译合约的使用说明


### 需求
```bigquery
  token 是plugin原生的go 合约，为了能方便的在第三方工具中使用这些token,需要对token 进行封装，用一套ERC20协议标准对其 
进行封装调用。
```

### 使用
```bigquery
   用户只需要使用内置的solidity 代码进行发布，就可以把自己发行的token 与自定义的预编译地址绑定在一起
 自定义的预编译地址：preToken.sol 中 PRECOMPILE  所表示的地址就是用户所指定的预编译合约地址。
 合约发布之后，即可通过合约地址对token 合约资产进行操作。
```
---
### 步骤

    1. 用户在链上发行原生 TOKEN：ABC
    2. 用户通过plugin/dapp/evm/contracts/token/Token.sol 发布合约，
       确保合约构造函数constructor(string memory name_, uint256 supply_)
       中币种名称和发行量与原生token 保持一致，得到合约地址
    3. 合约地址导入第三方工具进行操作，比如：metamask 
