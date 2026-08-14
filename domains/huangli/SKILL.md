---
name: liki-huangli
description: 黄历择日模块（组件）。宜忌查询、吉日选择。
---

# 黄历择日

## 路由

- **何时使用**：择日/当日宜忌/黄道吉日（「今天适合做什么」「哪天搬家/结婚/开业好」）
- 占卜问事（该不该做/吉凶成败）走六爻或奇门，黄历只管择日与宜忌

## 边界规则

- 宜忌基于神煞（建除十二神/黄黑道/二十八宿），断语查表（duanyu/yiji.md）不臆造
- 输出「宜/忌/平」，不产出吉凶评级档位
- 吉日选择按多日候选 + 事项优先级（duanyu/jiri.md）推荐 1-3 日

## 知识索引

| 文件 | 功能 | 输入 | 输出 |
|------|------|------|------|
| domains/huangli/duanyu/yiji.md | 事项宜忌 | 事项类型+神煞 | 宜/忌/平 |
| domains/huangli/duanyu/jiri.md | 吉日选择 | 多日候选+优先级 | 1-3推荐日 |

## 技术流程

1. **收集参数**：日期范围、事件类型、命主信息（可选）。
2. **调用引擎**：`huangli.days`（参数 `start_date` 起始日 YYYY-MM-DD + `count` 天数，默认 3 最多 30）查连续 N 天黄历宜忌。
3. **筛选推荐**：调 domains/huangli/duanyu/jiri.md → 按优先级选出 1-3 日。
4. **输出**：结构化数据，由执行主干分发到 app/auspicious 按模板输出。

📖 搜索 domains/huangli/duanyu/yiji.md → 读取宜忌规则
📖 搜索 domains/huangli/duanyu/jiri.md → 读取吉日选择规则
