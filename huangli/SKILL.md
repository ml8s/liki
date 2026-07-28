---
name: liki-huangli
description: 黄历择日模块（组件）。宜忌查询、吉日选择。
---

# 黄历择日

## 知识索引

| 文件 | 功能 | 输入 | 输出 |
|------|------|------|------|
| knowledge/yiji.md | 事项宜忌 | 事项类型+神煞 | 宜/忌/平 |
| knowledge/jiri.md | 吉日选择 | 多日候选+优先级 | 1-3推荐日 |

## 技术流程

1. **收集参数**：日期范围、事件类型、命主信息（可选）。
2. **调用引擎**：查指定日期的黄历宜忌。
3. **筛选推荐**：调 knowledge/jiri.md → 按优先级选出 1-3 日。
4. **输出** → `app/auspicious.md` 按模板输出。

📖 搜索 knowledge/yiji.md → 读取宜忌规则
📖 搜索 knowledge/jiri.md → 读取吉日选择规则
