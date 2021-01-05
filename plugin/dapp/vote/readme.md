
## 投票合约
基于区块链，公开透明的投票系统


### 交易功能
以下功能需要创建交易，并在链上执行

#### 创建投票组

- 设定投票组名称，管理员和组成员
- 默认创建者为管理员
- 可设定组成员的投票权重，默认都为1
- 投票组ID在交易执行后自动生成

##### 交易请求
```proto

//创建投票组
message CreateGroup {
    string   name                = 1; //投票组名称
    repeated string admins       = 2; //管理员地址列表
    repeated GroupMember members = 3; //组员
}

message GroupMember {
    string addr       = 1; //用户地址
    uint32 voteWeight = 2; //投票权重， 不填时默认为1
}

``` 

##### 交易回执
```proto
 
// 投票组信息
message GroupInfo {

    string   ID                  = 1; //投票组ID
    string   name                = 2; //投票组名称
    uint32   memberNum           = 3; //组员数量
    string   creator             = 4; //创建者
    repeated string admins       = 5; //管理员列表
    repeated GroupMember members = 6; //成员列表
}

```

#### 更新投票组成员
指定投票组ID，并由管理员添加或删除组成员

##### 交易请求
```proto
message UpdateMember {
    string   groupID                  = 1; //投票组ID
    repeated GroupMember addMembers   = 2; //需要增加的组成员
    repeated string removeMemberAddrs = 3; //删除组成员的地址列表
}
```
##### 交易回执
```proto
 
// 投票组信息
message GroupInfo {

    string   ID                  = 1; //投票组ID
    string   name                = 2; //投票组名称
    uint32   memberNum           = 3; //组员数量
    string   creator             = 4; //创建者
    repeated string admins       = 5; //管理员列表
    repeated GroupMember members = 6; //成员列表
}

```


#### 创建投票
- 设定投票名称，投票选项，关联投票组ID
- 关联投票组，即只有这些投票组的成员进行投票
- 投票ID在交易执行后生成

##### 交易请求
```proto

// 创建投票交易，请求结构
message CreateVote {
    string   name                  = 1; //投票名称
    repeated string voteGroups     = 2; //投票关联组列表
    repeated string voteOptions    = 3; //投票选项列表
    int64           beginTimestamp = 4; //投票开始时间戳
    int64           endTimestamp   = 5; //投票结束时间戳
}

``` 

##### 交易回执
```proto

//投票信息
message VoteInfo {

    string   ID                        = 1; //投票ID
    string   name                      = 2; //投票名称
    string   creator                   = 3; //投票创建者
    repeated string voteGroups         = 4; //投票关联的投票组
    repeated VoteOption voteOptions    = 5; //投票的选项
    int64               beginTimestamp = 6; //投票开始时间戳
    int64               endTimestamp   = 7; //投票结束时间戳
    repeated string votedMembers       = 8; //已投票的成员
}

//投票选项
message VoteOption {
    string option = 1; //投票选项
    uint32 score  = 2; //投票得分
}
```

#### 提交投票
- 投票组成员发起投票交易
- 指定所在投票组ID，投票ID，投票选项
- 投票选项使用数组下标标识，而不是选项内容

##### 交易请求
```proto

// 创建提交投票交易，请求结构
message CommitVote {
    string voteID      = 1; //投票ID
    string groupID     = 2; //所属投票组ID
    uint32 optionIndex = 3; //投票选项数组下标，下标对应投票内容
}

```

##### 交易回执
```proto

//投票信息
message VoteInfo {

    string   ID                        = 1; //投票ID
    string   name                      = 2; //投票名称
    string   creator                   = 3; //投票创建者
    repeated string voteGroups         = 4; //投票关联的投票组
    repeated VoteOption voteOptions    = 5; //投票的选项
    int64               beginTimestamp = 6; //投票开始时间戳
    int64               endTimestamp   = 7; //投票结束时间戳
    repeated string votedMembers       = 8; //已投票的成员
}
```


### 查询功能
以下功能为本地查询，无需创建交易

#### 获取组信息
根据投票组ID查询组信息

##### 请求结构
```proto
message ReqString {
    string Data = 1;
}
```

##### 响应结构

```proto
message GroupVoteInfo {

    GroupInfo groupInfo     = 1; // 投票组信息
    repeated string voteIDs = 2; //投票组关联的投票ID信息
}
```

#### 获取投票信息
根据投票ID查询投票信息

##### 请求结构
```proto
message ReqString {
    string Data = 1;
}
```

##### 响应结构
```proto
message VoteInfo {

    string   ID                        = 1; //投票ID
    string   name                      = 2; //投票名称
    string   creator                   = 3; //投票创建者
    repeated string voteGroups         = 4; //投票关联的投票组
    repeated VoteOption voteOptions    = 5; //投票的选项
    int64               beginTimestamp = 6; //投票开始时间戳
    int64               endTimestamp   = 7; //投票结束时间戳
    repeated string votedMembers       = 8; //已投票的成员
}
```

#### 获取成员信息
根据成员地址，查询成员信息

##### 请求结构
```proto
message ReqString {
    string Data = 1;
}
```

##### 响应结构
```proto
 
message MemberInfo {
    string   addr            = 1; //地址
    repeated string groupIDs = 2; //所属投票组的ID列表
}
```

#### 获取投票组列表

##### 请求结构
```proto
//列表请求结构
message ReqListItem {
    string startItemID = 1; //列表开始的ID，如请求组列表即groupID，不包含在结果中
    int32  count       = 2; //请求列表项数量
    int32  direction   = 3; // 0表示根据ID降序，1表示升序
}
```

##### 响应结构
```proto
message GroupVoteInfos {
    repeated GroupVoteInfo groupList = 1; //投票组信息列表
}
```

#### 获取投票列表

##### 请求结构
```proto
//列表请求结构
message ReqListItem {
    string startItemID = 1; //列表开始的ID，如请求组列表即groupID，不包含在结果中
    int32  count       = 2; //请求列表项数量
    int32  direction   = 3; // 0表示根据ID降序，1表示升序
}
```

##### 响应结构
```proto
message VoteInfos {
    repeated VoteInfo voteList = 1; //投票信息列表
}
```

#### 获取投票组成员列表

##### 请求结构

```proto
//列表请求结构
message ReqListItem {
    string startItemID = 1; //列表开始的ID，如请求组列表即groupID，不包含在结果中
    int32  count       = 2; //请求列表项数量
    int32  direction   = 3; // 0表示根据ID降序，1表示升序
}
```

##### 响应结构
```proto
message MemberInfos {
    repeated MemberInfo memberList = 1; //投票组成员信息列表
}
```

#### 其他 
 
[投票合约proto源文件](proto/vote.proto)
