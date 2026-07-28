---
name: liki-qimen
description: 奇门遁甲模块（组件）。八门、九星、八神、用神、决策。
---

# 奇门遁甲

## 知识索引

| 文件 | 功能 |
|------|------|
| knowledge/bamen.md | 八门吉凶 |
| knowledge/jiuxing.md | 九星吉凶 |

## 技术流程

1. **路由判断**：用户问方向/时机时使用奇门。
2. **收集参数**：问题 + 时间 + 方位（可选）。
3. **调用引擎**：排盘。
4. **解读**：调 knowledge/ 匹配八门九星。

📖 搜索 knowledge/bamen.md → 读取八门表
📖 搜索 knowledge/jiuxing.md → 读取九星表
