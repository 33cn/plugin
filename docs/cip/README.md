# CIP — Chain33 Improvement Proposal

CIP 是 Chain33 生态的改进提案机制，参照 Bitcoin 的 [BIP](https://github.com/bitcoin/bips) 与 Ethereum 的 [EIP](https://github.com/ethereum/EIPs)。
一份 CIP 是一份**设计文档**，用来描述：共识/跨链规则变更、执行器与接口标准、流程与治理、或信息性说明。代码实现与 CIP 分离：CIP 先评审定稿，再由实施计划落地并随分叉激活。

## 目录与命名

```
docs/cip/
├── README.md          # 本文档：流程与规范
├── cip-template.md    # 新建 CIP 使用的模板
└── cip-001-<slug>.md  # 具体提案，按序号排列
```

- 文件名：`cip-<三位序号>-<slug>.md`（序号从 `001` 开始，按合并顺序分配）。
- 文档内引用写作 `CIP-001`。
- 流程文档（`README.md`、`cip-template.md`）不占用序号。
- `slug` 用小写英文与连字符，简短描述主题。

## 类型（Type）

| 类型 | 用途 |
|---|---|
| `Standards Track` | 共识、跨链、执行器、RPC/接口、配置与分叉规则等需要实现者遵循的变更 |
| `Process` | 流程、治理、发布与协作方式的变更 |
| `Informational` | 设计说明、最佳实践、背景资料，不要求实现 |
| `Meta` | 关于 CIP 流程本身的变更 |

`Standards Track` 提案应当注明 `category`（如 `Consensus`、`Cross-chain`、`Interface`、`Executor`）。

## 状态（Status）

```
Draft ──> Review ──> Last Call ──> Accepted ──> Final ──> Active
   │          │           │            │
   └──────────┴───────────┴────────────┴──> Withdrawn / Rejected / Deferred / Superseded
```

| 状态 | 含义 | 进入条件 |
|---|---|---|
| `Draft` | 草案，内容可自由修改 | PR 提交即进入 |
| `Review` | 请求社区与维护者评审 | 作者认为内容已完整 |
| `Last Call` | 最后征集意见 | 评审无重大异议，维护者设定截止时间（默认 14 天） |
| `Accepted` | 已接受，进入实施 | Last Call 期满且无阻塞性异议 |
| `Final` | 实现已完成并随分叉高度发布 | 代码合入 + 分叉高度确定 |
| `Active` | 已在链上分叉激活 | 主链达到分叉高度并稳定运行 |
| `Withdrawn` | 作者主动撤回 | 作者决定 |
| `Rejected` | 维护者否决 | 存在阻塞性技术或治理问题 |
| `Deferred` | 暂缓，等待前置条件 | 依赖项未就绪 |
| `Superseded` | 被新 CIP 取代 | 头部填写 `superseded-by: CIP-XXX` |

状态变更通过修改 CIP 头部的 `status` 字段并提交 PR 完成，评审记录保留在 PR 中。

## 头部元数据

每份 CIP 必须以 YAML 风格头部开始：

```yaml
---
cip: 001
title: 一句话标题
author: 姓名或组织 <邮箱，可选>
status: Draft
type: Standards Track
category: Consensus / Cross-chain
created: YYYY-MM-DD
requires: CIP-XXX 或 PR #NNN（可选，可多个）
replaces: CIP-XXX（可选）
superseded-by: CIP-XXX（可选）
discussions-to: 讨论链接（PR/issue）
---
```

## 必需章节

`摘要`、`动机`、`规范`、`理由`、`向后兼容`、`安全考虑`、`测试与参考实现`、`实施路线`、`版权`。
规范章节使用 RFC 2119 关键词：**必须（MUST）**、**不得（MUST NOT）**、**应当（SHOULD）**、**可以（MAY）**。

## 提交流程

1. 复制 `cip-template.md`，按 `cip-<序号>-<slug>.md` 命名（新提案先向维护者申请序号，或按 PR 顺序由维护者分配）。
2. 提交 PR 到本仓库，PR 标题格式 `CIP-XXX: <title>`，初始 `status: Draft`。
3. 评审期间在 PR 中讨论并更新文档；状态推进由作者发起、维护者确认。
4. `Accepted` 后，实施计划写入 `docs/superpowers/plans/` 或对应的设计文档，代码实现另行 PR。
5. 分叉激活后把状态更新为 `Active`。

## 与代码/分叉的关系

- CIP 描述**规则与理由**；具体的参数取值、KV 布局、错误码属于规范的一部分，必须写明。
- 任何共识相关的 CIP 都必须说明：分叉开关名称、默认高度、平行链侧配置映射、以及未升级节点的后果。
- CIP 不替代审计与测试：实施必须包含单元/集成测试与升级检查清单。

## 版权

CIP 文本按 [CC0-1.0](https://creativecommons.org/publicdomain/zero/1.0/) 放弃版权（与 EIP 一致）；参考实现代码遵循本仓库的 BSD 3-Clause License。维护者可以调整此约定。
