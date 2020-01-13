## 借贷合约表结构

### 放贷表coller表结构
字段名称|类型|说明
---|---|---
collateralizeId|string|放贷ID，主键
accountAddr|string|大户地址
recordId|string|借贷ID
status|int32|放贷状态（1：已放贷 2：已收回）

### 放贷表coller表索引
索引名|说明
---|---
status|根据放贷状态查询放贷ID
addr|根据大户地址查询放贷ID
addr_status|根据放贷状态和大户地址查询放贷ID

### 借贷表borrow表结构
字段名称|类型|说明
---|---|---
recordId|string|借贷ID，主键
collateralizeId|string|放贷ID
accountAddr|string|用户地址
status|int32|借贷状态（1：已发行 2：价格清算告警 3：价格清算 4：超时清算告警 5：超时清算 6：已清算）

### 放贷表borrow表索引
索引名|说明
---|---
status|根据借贷状态查询借贷ID
addr|根据大户地址查询借贷ID
addr_status|根据借贷状态和用户地址查询借贷ID
id_status|根据放贷ID和借贷状态查询借贷ID
id_addr|根据放贷ID和用户地址查询借贷ID