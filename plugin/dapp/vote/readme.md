
## 投票合约
基于区块链，公开透明的投票系统


### 交易功能
以下功能需要创建交易，并在链上执行

#### 创建投票组(CreateGroup)

- 设定投票组名称，管理员和组成员
- 默认创建者为管理员
- 可设定组成员的投票权重，默认都为1
- 投票组ID在交易执行后自动生成

##### 交易请求
```proto

//创建投票组
message CreateGroup {
    string   name                    = 1; //投票组名称
    repeated string admins           = 2; //管理员地址列表, 创建者默认为管理员
    repeated GroupMember members     = 3; //组员
    string               description = 4; //描述
}

message GroupMember {
    string addr       = 1; //用户地址
    uint32 voteWeight = 2; //投票权重， 不填时默认为1
    string nickName   = 3; //群昵称
}

```

##### 交易回执
```proto

// 投票组信息
message GroupInfo {

    string   ID                      = 1; //投票组ID
    string   name                    = 2; //投票组名称
    uint32   memberNum               = 3; //组员数量
    string   creator                 = 4; //创建者
    repeated string admins           = 5; //管理员列表
    repeated GroupMember members     = 6; //成员列表
    string               description = 7; //描述信息
    uint32 voteNum = 8; //投票数量
}

```
##### 创建交易示例

- 创建交易通用json rpc接口，Chain33.CreateTransaction
- actionName: CreateGroup

```bash
curl -kd  '{"method":"Chain33.CreateTransaction","params":[{"execer":"vote","actionName":"CreateGroup","payload":{"name":"group30","admins":[],"members":[{"addr":"1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5","voteWeight":0}],"description":""}}],"id":0}' http://localhost:8801
```

#### 更新投票组(UpdateGroup)
指定投票组ID，并由管理员添加或删除组成员、管理员

##### 交易请求
```proto
message UpdateGroup {
    string   groupID                = 1; //投票组ID
    repeated GroupMember addMembers = 2; //需要增加的组成员
    repeated string removeMembers   = 3; //删除组成员的地址列表
    repeated string addAdmins       = 4; //增加管理员
    repeated string removeAdmins    = 5; //删除管理员
}

message GroupMember {
    string addr       = 1; //用户地址
    uint32 voteWeight = 2; //投票权重， 不填时默认为1
    string nickName   = 3; //群昵称
}
```
##### 交易回执
```proto

// 投票组信息
message GroupInfo {

    string   ID                      = 1; //投票组ID
    string   name                    = 2; //投票组名称
    uint32   memberNum               = 3; //组员数量
    string   creator                 = 4; //创建者
    repeated string admins           = 5; //管理员列表
    repeated GroupMember members     = 6; //成员列表
    string               description = 7; //描述信息
    uint32 voteNum = 8; //投票数量
}

```

##### 创建交易示例

- 创建交易通用json rpc接口，Chain33.CreateTransaction
- actionName: UpdateGroup

```bash
curl -kd  '{"method":"Chain33.CreateTransaction","params":[{"execer":"vote","actionName":"UpdateGroup","payload":{"groupID":"g000000000000700000","addMembers":[{"addr":"member1","voteWeight":0},{"addr":"member2","voteWeight":0}],"removeMembers":["member3"],"addAdmins":["admin1"],"removeAdmins":["admin2"]}}],"id":0}' http://localhost:8801
```

#### 创建投票(CreateVote)
- 由管理员发起
- 设定投票名称，投票选项，关联投票组ID
- 关联投票组，即只有这些投票组的成员进行投票
- 投票ID在交易执行后生成

##### 交易请求
```proto

// 创建投票交易，请求结构
message CreateVote {
    string   name                  = 1; //投票名称
    string   groupID               = 2; //投票关联组
    repeated string voteOptions    = 3; //投票选项列表
    int64           beginTimestamp = 4; //投票开始时间戳
    int64           endTimestamp   = 5; //投票结束时间戳
    string          description    = 6; //描述信息
}

```

##### 交易回执
```proto

//投票信息
message VoteInfo {

    string   ID                        = 1;  //投票ID
    string   name                      = 2;  //投票名称
    string   creator                   = 3;  //创建者
    string   groupID                   = 4;  //投票关联的投票组
    repeated VoteOption voteOptions    = 5;  //投票的选项
    int64               beginTimestamp = 6;  //投票开始时间戳
    int64               endTimestamp   = 7;  //投票结束时间戳
    repeated CommitInfo commitInfos    = 8;  //已投票的提交信息
    string              description    = 9;  //描述信息
    uint32              status         = 10; //状态，1即将开始，2正在进行，3已经结束，4已关闭
}

//投票选项
message VoteOption {
    string option = 1; //投票选项
    uint32 score  = 2; //投票得分
}
```

##### 创建交易示例

- 创建交易通用json rpc接口，Chain33.CreateTransaction
- actionName: CreateVote

```bash
curl -kd  '{"method":"Chain33.CreateTransaction","params":[{"execer":"vote","actionName":"CreateVote","payload":{"name":"vote1","groupID":"g000000000000600000","voteOptions":["A","B","C"],"beginTimestamp":"1611562096","endTimestamp":"1611648496","description":""}}],"id":0}' http://localhost:8801
```

#### 提交投票(CommitVote)
- 投票组成员发起投票交易
- 指定投票ID，投票选项
- 投票选项使用数组下标标识，而不是选项内容

##### 交易请求
```proto

// 创建提交投票交易，请求结构
message CommitVote {
    string voteID      = 1; //投票ID
    uint32 optionIndex = 2; //投票选项数组下标，下标对应投票内容
}

```

##### 交易回执
```proto

//投票信息
message CommitInfo {
    string addr       = 1; //提交地址
    string txHash     = 2; //提交交易哈希
    uint32 voteWeight = 3; //投票权重
}
```

##### 创建交易示例

- 创建交易通用json rpc接口，Chain33.CreateTransaction
- actionName: CommitVote

```bash
curl -kd  '{"method":"Chain33.CreateTransaction","params":[{"execer":"vote","actionName":"CommitVote","payload":{"voteID":"v000000000001300000","optionIndex":0}}],"id":0}' http://localhost:8801

```

#### 关闭投票(CloseVote)
- 由管理员发起，将指定投票关闭

##### 交易请求
```proto

message CloseVote {
    string voteID = 1; // 投票ID
}

```

##### 交易回执
```proto

message VoteInfo {

    string   ID                        = 1;  //投票ID
    string   name                      = 2;  //投票名称
    string   creator                   = 3;  //创建者
    string   groupID                   = 4;  //投票关联的投票组
    repeated VoteOption voteOptions    = 5;  //投票的选项
    int64               beginTimestamp = 6;  //投票开始时间戳
    int64               endTimestamp   = 7;  //投票结束时间戳
    repeated CommitInfo commitInfos    = 8;  //已投票的提交信息
    string              description    = 9;  //描述信息
    uint32              status         = 10; //状态，1即将开始，2正在进行，3已经结束，4已关闭
    string              groupName      = 11; //所属投票组名称
}
```

##### 创建交易示例

- 创建交易通用json rpc接口，Chain33.CreateTransaction
- actionName: CloseVote

```bash
curl -kd  '{"method":"Chain33.CreateTransaction","params":[{"execer":"vote","actionName":"CloseVote","payload":{"voteID":"v000000000001300000"}}],"id":0}' http://localhost:8801
```

#### 更新用户信息(UpdateMember)
- 目前仅支持用户更新名称信息

##### 交易请求
```proto

message UpdateMember {
    string name = 1; //用户名称
}

```

##### 交易回执
```proto

message MemberInfo {
    string   addr            = 1; //地址
    string   name            = 2; //用户名称
    repeated string groupIDs = 3; //所属投票组的ID列表
}
```

##### 创建交易示例

- 创建交易通用json rpc接口，Chain33.CreateTransaction
- actionName: UpdateMember

```bash
 curl -kd  '{"method":"Chain33.CreateTransaction","params":[{"execer":"vote","actionName":"UpdateMember","payload":{"name":"name1"}}],"id":0}' http://localhost:8801
```


### 查询功能
以下功能为本地查询，无需创建交易

#### 获取组信息(GetGroups)
根据投票组ID查询组信息，支持多个同时查询

##### 请求结构
```proto
message ReqStrings {
    repeated string items = 1; //投票组ID列表
}
```

##### 响应结构

```proto
message GroupInfos {
    repeated GroupInfo groupList = 1; //投票组信息列表
}
```
##### 示例

- 通用查询json rpc接口，Chain33.Query
- funcName: GetGroups

```bash
curl -ksd '{"method":"Chain33.Query","params":[{"execer":"vote","funcName":"GetGroups","payload":{"items":["g000000000001700000","g000000000001800000"]}}],"id":0}' http://localhost:8801
```

#### 获取投票信息(GetVotes)
根据投票ID查询投票信息，支持多个同时查询

##### 请求结构
```proto
message ReqStrings {
    repeated string items = 1; //投票ID列表
}
```

##### 响应结构
```proto
message ReplyVoteList {
    repeated VoteInfo voteList         = 1; //投票列表
    int64             currentTimestamp = 2; //当前系统时间
}
```

##### 示例

- 通用查询json rpc接口，Chain33.Query
- funcName: GetVotes

```bash
curl -kd  '{"method":"Chain33.Query","params":[{"execer":"vote","funcName":"GetVotes","payload":{"items":["v000000000001300000","v000000000001400000"]}}],"id":0}' http://localhost:8801
```

#### 获取成员信息(GetMembers)
根据用户地址，获取用户信息，支持多个同时查询

##### 请求结构
```proto
message ReqStrings {
    repeated string items = 1; //用户地址列表
}
```

##### 响应结构
```proto

message MemberInfos {
    repeated MemberInfo memberList = 1; //投票组成员信息列表
}

message MemberInfo {
    string   addr            = 1; //地址
    string   name            = 2; //用户名称
    repeated string groupIDs = 3; //所属投票组的ID列表
}
```

##### 示例

- 通用查询json rpc接口，Chain33.Query
- funcName: GetMembers
-
```bash
curl -kd  '{"method":"Chain33.Query","params":[{"execer":"vote","funcName":"GetMembers","payload":{"items":["1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5"]}}],"id":0}' http://localhost:8801
```

#### 获取投票组列表(ListGroup)
全局投票组有序列表

##### 请求结构
```proto
//列表请求结构
message ReqListItem {
    string startItemID = 1; //列表开始的ID，如请求组列表即groupID，不包含在结果中
    int32  count       = 2; //请求列表项数量, 0表示请求所有
    int32  direction   = 3; // 0表示根据ID降序，1表示升序，目前ID和区块高度正相关
}
```

##### 响应结构
```proto
message GroupInfos {
    repeated GroupInfo groupList = 1; //投票组信息列表
}
```

##### 示例

- 通用查询json rpc接口，Chain33.Query
- funcName: ListGroup

```bash
curl -kd  '{"method":"Chain33.Query","params":[{"execer":"vote","funcName":"ListGroup","payload":{"startItemID":"","count":2,"direction":0}}],"id":0}' http://localhost:8801
```

#### 获取投票列表(ListVote)
- 获取全局投票列表
- 可指定groupID，获取指定组的投票列表
- 可指定投票状态进行分页查找, status 0表示不做状态区分

##### 请求结构
```proto
//列表请求结构
message ReqListVote {
    string      groupID = 1; //指定所属组ID
    ReqListItem listReq = 2; //列表请求
    uint32      status  = 3; //指定投票状态, 1即将开始，2正在进行，3已经结束，4已关闭
}

message ReqListItem {
    string startItemID = 1; //列表开始的ID，如请求组列表即groupID，不包含在结果中
    int32  count       = 2; //请求列表项数量, 0表示请求所有
    int32  direction   = 3; // 0表示根据ID降序，1表示升序，目前ID和区块高度正相关
}
```


##### 响应结构
```proto
message ReplyVoteList {
    repeated VoteInfo voteList         = 1; //投票列表
    int64             currentTimestamp = 2; //当前系统时间
}
```

##### 示例

- 通用查询json rpc接口，Chain33.Query
- funcName: ListVote

```bash
curl -kd  '{"method":"Chain33.Query","params":[{"execer":"vote","funcName":"ListVote","payload":{"groupID":"","listReq":{"startItemID":"","count":2,"direction":0}}}],"id":0}' http://localhost:8801
```

#### 获取用户列表(ListMember)
全局用户有序列表

##### 请求结构

```proto
//列表请求结构
message ReqListItem {
    string startItemID = 1; //列表开始的ID，如请求组列表即groupID，不包含在结果中
    int32  count       = 2; //请求列表项数量, 0表示请求所有
    int32  direction   = 3; // 0表示根据ID降序，1表示升序，目前ID和区块高度正相关
}
```

##### 响应结构
```proto
message MemberInfos {
    repeated MemberInfo memberList = 1; //投票组成员信息列表
}
```

##### 示例

- 通用查询json rpc接口，Chain33.Query
- funcName: ListMember

```bash
curl -kd  '{"method":"Chain33.Query","params":[{"execer":"vote","funcName":"ListMember","payload":{"startItemID":"","count":1,"direction":1}}],"id":0}' http://localhost:8801
```

#### 错误码表
发送交易和查询等接口可能返回的错误

|名称 |含义
|---|---|
errEmptyName|       名称为空
errInvalidMemberWeights | 非法投票权重
errDuplicateMember      | 投票组成员重复
errDuplicateAdmin       | 投票组管理员重复
errInvalidVoteTime      | 非法投票时间
errInvalidVoteOption    | 非法投票选项
errVoteNotExist         | 投票不存在
errGroupNotExist        | 投票组不存在
errStateDBGet           | 状态数据获取错误
errInvalidVoteID        | 非法投票ID
errInvalidGroupID       | 非法投票组ID
errInvalidOptionIndex   | 非法投票索引
errAddrAlreadyVoted     | 已完成投票
errVoteAlreadyFinished  | 投票已结束
errVoteNotStarted       | 投票未开始
errVoteAlreadyClosed    | 投票已关闭
errAddrPermissionDenied | 地址没有权限




#### 其他

[投票合约proto源文件](proto/vote.proto)
