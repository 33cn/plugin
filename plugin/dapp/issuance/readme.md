## 发行合约表结构

### 总发行表issuer表结构
字段名称|类型|说明
---|---|---
issuanceId|string|总发行ID，主键
status|int32|发行状态（1：已发行 2：已下线）

### 总发行表issuer表索引
索引名|说明
---|---
status|根据发行状态查询总发行ID

### 大户发行表debt表结构
字段名称|类型|说明
---|---|---
debtId|string|大户发行ID，主键
issuanceId|string|总发行ID
accountAddr|string|用户地址
status|int32|发行状态（1：已发行 2：价格清算告警 3：价格清算 4：超时清算告警 5：超时清算 6：关闭）

### 大户发行表debt表索引
索引名|说明
---|---
status|根据大户发行状态查询大户发行ID
addr|根据大户地址查询大户发行ID
addr_status|根据发行状态和大户地址查询大户发行ID
