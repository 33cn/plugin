---
cip: XXX
title: <一句话标题>
author: <姓名或组织> <邮箱，可选>
status: Draft
type: Standards Track | Process | Informational | Meta
category: <Consensus / Cross-chain / Interface / Executor，Informational 可省略>
created: YYYY-MM-DD
requires: <CIP-XXX 或 PR #NNN，可选，可多个>
replaces: <CIP-XXX，可选>
superseded-by: <CIP-XXX，可选>
discussions-to: <讨论链接>
---

## 摘要

用 3–5 句话说明：改什么、为什么改、影响谁。

## 动机

现状是什么、存在什么问题、造成过什么实际后果（如有链上事件请给出高度与交易哈希）。

## 规范

本节是提案的规范部分，使用 必须/不得/应当/可以（RFC 2119）。

### 数据结构

新增/修改的链上状态、KV key、消息字段、错误码，逐一给出精确名称与含义。

### 规则

状态机的状态、转移条件、边界条件、幂等性要求。

### 分叉与配置

分叉开关名称、默认高度、平行链侧配置映射、参数默认值、未升级节点的后果。

## 理由

解释关键设计取舍：为什么选择该方案而不是被否掉的替代方案。

## 向后兼容

对现有链、现有平行链、现有查询/CLI、现有状态的影响；迁移与升级要求。

## 安全考虑

威胁模型、本提案缓解什么、明确不缓解什么、残留风险与运维缓解措施。

## 测试与参考实现

单元/集成/回归测试要覆盖的用例清单；参考实现所在的分支或 PR。

## 实施路线

阶段划分（本期实现 / 后续阶段）、每阶段的交付物与依赖项。

## 版权

CIP 文本按 CC0-1.0 放弃版权；参考实现代码遵循本仓库 LICENSE。
